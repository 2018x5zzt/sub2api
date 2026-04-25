package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type subscriptionReadCacheInvalidatorStub struct {
	calls       int
	lastUserID  int64
	lastGroupID int64
}

func (s *subscriptionReadCacheInvalidatorStub) InvalidateSubCache(userID, groupID int64) {
	s.calls++
	s.lastUserID = userID
	s.lastGroupID = groupID
}

func newGatewayRecordUsageServiceWithBillingRepoForCacheTest(usageRepo UsageLogRepository, billingRepo UsageBillingRepository, userRepo UserRepository, subRepo UserSubscriptionRepository) *GatewayService {
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1.1
	return NewGatewayService(
		nil,
		nil,
		usageRepo,
		billingRepo,
		userRepo,
		subRepo,
		nil,
		nil,
		cfg,
		nil,
		nil,
		NewBillingService(cfg, nil),
		nil,
		&BillingCacheService{},
		nil,
		nil,
		&DeferredService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func TestOpenAIRecordUsage_InvalidatesSubscriptionCachesAfterBilling(t *testing.T) {
	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newOpenAIRecordUsageServiceForTest(usageRepo, userRepo, subRepo, nil)

	cache := &billingEligibilityCacheStub{}
	svc.billingCacheService = NewBillingCacheService(cache, nil, nil, nil, &config.Config{})
	t.Cleanup(svc.billingCacheService.Stop)

	invalidator := &subscriptionReadCacheInvalidatorStub{}
	svc.SetSubscriptionReadCacheInvalidator(invalidator)

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID: "resp_subscription_cache_invalidate",
			Usage:     OpenAIUsage{InputTokens: 10, OutputTokens: 5},
			Model:     "gpt-5.1",
			Duration:  time.Second,
		},
		APIKey:       &APIKey{ID: 100, GroupID: i64p(88), Group: &Group{ID: 88, SubscriptionType: SubscriptionTypeSubscription}},
		User:         &User{ID: 200},
		Account:      &Account{ID: 300},
		Subscription: &UserSubscription{ID: 99},
	})

	require.NoError(t, err)
	require.Equal(t, 1, subRepo.incrementCalls)
	require.Equal(t, 0, cache.subscriptionUpdates)
	require.Equal(t, 1, cache.subscriptionInvalidates)
	require.Equal(t, 1, invalidator.calls)
	require.Equal(t, int64(200), invalidator.lastUserID)
	require.Equal(t, int64(88), invalidator.lastGroupID)
}

func TestGatewayServiceRecordUsage_InvalidatesSubscriptionCachesAfterBilling(t *testing.T) {
	usageRepo := &openAIRecordUsageLogRepoStub{}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newGatewayRecordUsageServiceWithBillingRepoForCacheTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, subRepo)

	cache := &billingEligibilityCacheStub{}
	svc.billingCacheService = NewBillingCacheService(cache, nil, nil, nil, &config.Config{})
	t.Cleanup(svc.billingCacheService.Stop)

	invalidator := &subscriptionReadCacheInvalidatorStub{}
	svc.SetSubscriptionReadCacheInvalidator(invalidator)

	err := svc.RecordUsage(context.Background(), &RecordUsageInput{
		Result: &ForwardResult{
			RequestID: "gateway_subscription_cache_invalidate",
			Usage: ClaudeUsage{
				InputTokens:  10,
				OutputTokens: 6,
			},
			Model:    "claude-sonnet-4",
			Duration: time.Second,
		},
		APIKey:       &APIKey{ID: 501, GroupID: i64p(88), Group: &Group{ID: 88, SubscriptionType: SubscriptionTypeSubscription}},
		User:         &User{ID: 601},
		Account:      &Account{ID: 701},
		Subscription: &UserSubscription{ID: 801},
	})

	require.NoError(t, err)
	require.Equal(t, 1, billingRepo.calls)
	require.Equal(t, 0, subRepo.incrementCalls)
	require.Equal(t, 0, cache.subscriptionUpdates)
	require.Equal(t, 1, cache.subscriptionInvalidates)
	require.Equal(t, 1, invalidator.calls)
	require.Equal(t, int64(601), invalidator.lastUserID)
	require.Equal(t, int64(88), invalidator.lastGroupID)
}
