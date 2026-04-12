package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"go.uber.org/zap"
)

const (
	// openAI429SilentFailoverBudget limits how long a single request may keep
	// silently switching accounts after upstream 429s before the handler gives
	// up and returns a client-visible 429.
	openAI429SilentFailoverBudget = 5 * time.Second
	// Keep the poll interval short so the handler can pick up newly available
	// accounts without blocking too long on a single retry step.
	openAI429SilentFailoverPollInterval = 200 * time.Millisecond
)

type openAI429SilentFailoverState struct {
	startedAt time.Time
	switches  int
}

func isOpenAI429Failover(failoverErr *service.UpstreamFailoverError) bool {
	return failoverErr != nil && failoverErr.StatusCode == http.StatusTooManyRequests
}

func shouldRetrySameOpenAIAccount(failoverErr *service.UpstreamFailoverError) bool {
	return failoverErr != nil && failoverErr.RetryableOnSameAccount && !isOpenAI429Failover(failoverErr)
}

func (s *openAI429SilentFailoverState) noteSwitch(failoverErr *service.UpstreamFailoverError, now time.Time) bool {
	if !isOpenAI429Failover(failoverErr) {
		return false
	}
	if s.startedAt.IsZero() {
		s.startedAt = now
	}
	s.switches++
	return true
}

func (s openAI429SilentFailoverState) switchCount() int {
	return s.switches
}

func clearOpenAI429ExcludedAccounts(failedAccountIDs map[int64]struct{}) int {
	if len(failedAccountIDs) == 0 {
		return 0
	}
	cleared := 0
	for accountID := range failedAccountIDs {
		delete(failedAccountIDs, accountID)
		cleared++
	}
	return cleared
}

func (s openAI429SilentFailoverState) remaining(now time.Time) time.Duration {
	if s.startedAt.IsZero() {
		return 0
	}
	remaining := openAI429SilentFailoverBudget - now.Sub(s.startedAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s openAI429SilentFailoverState) nextWaitDuration(now time.Time) time.Duration {
	remaining := s.remaining(now)
	if remaining <= 0 {
		return 0
	}
	if remaining < openAI429SilentFailoverPollInterval {
		return remaining
	}
	return openAI429SilentFailoverPollInterval
}

func (h *OpenAIGatewayHandler) waitForOpenAI429SilentFailover(
	ctx context.Context,
	reqLog *zap.Logger,
	state *openAI429SilentFailoverState,
	failedAccountIDs map[int64]struct{},
	stage string,
) bool {
	if state == nil {
		return false
	}
	now := time.Now()
	waitFor := state.nextWaitDuration(now)
	if waitFor <= 0 {
		return false
	}
	if reqLog != nil {
		reqLog.Info("openai.rate_limit_silent_failover_wait",
			zap.String("stage", stage),
			zap.Int64("wait_ms", waitFor.Milliseconds()),
			zap.Int64("budget_remaining_ms", state.remaining(now).Milliseconds()),
		)
	}
	if !sleepWithContext(ctx, waitFor) {
		return false
	}
	cleared := clearOpenAI429ExcludedAccounts(failedAccountIDs)
	if reqLog != nil && cleared > 0 {
		reqLog.Debug("openai.rate_limit_silent_failover_retry_sweep_reset",
			zap.String("stage", stage),
			zap.Int("cleared_excluded_account_count", cleared),
			zap.Int("rate_limit_switch_count", state.switchCount()),
		)
	}
	return true
}
