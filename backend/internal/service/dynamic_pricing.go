package service

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

const (
	GroupPricingModeFixed   = "fixed"
	GroupPricingModeDynamic = "dynamic"

	BudgetMultiplierMin     = 3.0
	BudgetMultiplierMax     = 50.0
	DefaultBudgetMultiplier = 8.0
	dynamicBudgetEpsilon    = 1e-9
	dynamicBudgetWindow     = 7 * 24 * time.Hour
)

var (
	ErrGroupPricingModeInvalid      = infraerrors.BadRequest("GROUP_PRICING_MODE_INVALID", "pricing_mode must be fixed or dynamic")
	ErrGroupDefaultBudgetRequired   = infraerrors.BadRequest("GROUP_DEFAULT_BUDGET_REQUIRED", "default budget multiplier is required for dynamic pricing groups")
	ErrAPIKeyBudgetRequired         = infraerrors.BadRequest("API_KEY_BUDGET_REQUIRED", "budget multiplier is required for dynamic pricing groups")
	ErrBudgetMultiplierOutOfRange   = infraerrors.BadRequest("BUDGET_MULTIPLIER_OUT_OF_RANGE", "budget multiplier must be between 3 and 50")
	ErrAPIKeyGroupImmutable         = infraerrors.BadRequest("API_KEY_GROUP_IMMUTABLE", "api key group cannot be changed after creation")
	ErrDynamicPricingBudgetExceeded = infraerrors.TooManyRequests(
		"DYNAMIC_PRICING_BUDGET_EXCEEDED",
		"当前预算倍率下没有可用账号，可调高预算倍率后重试",
	)
)

func normalizeGroupPricingMode(mode string) string {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return GroupPricingModeFixed
	}
	return normalized
}

func validateGroupPricingMode(mode string) (string, error) {
	normalized := normalizeGroupPricingMode(mode)
	switch normalized {
	case GroupPricingModeFixed, GroupPricingModeDynamic:
		return normalized, nil
	default:
		return "", ErrGroupPricingModeInvalid
	}
}

func validateBudgetMultiplier(value *float64, requiredErr error) (*float64, error) {
	if value == nil {
		if requiredErr != nil {
			return nil, requiredErr
		}
		return nil, nil
	}
	if *value < BudgetMultiplierMin || *value > BudgetMultiplierMax {
		return nil, ErrBudgetMultiplierOutOfRange
	}
	validated := *value
	return &validated, nil
}

type userGroupRateResolverFunc func(ctx context.Context, userID, groupID int64, groupDefaultMultiplier float64) float64

type dynamicPricingBudgetStateKey struct{}

type dynamicPricingBudgetState struct {
	enabled                   bool
	budgetMultiplier          float64
	windowStart               time.Time
	windowEnd                 time.Time
	currentStandardCost       float64
	currentActualCost         float64
	currentAverageMultiplier  float64
	estimatedNextStandardCost float64
}

func apiKeyFromContext(ctx context.Context) *APIKey {
	if ctx == nil {
		return nil
	}
	apiKey, _ := ctx.Value(ctxkey.APIKey).(*APIKey)
	return apiKey
}

func dynamicPricingGroupFromContext(ctx context.Context, groupID *int64) *Group {
	if ctx == nil || groupID == nil || *groupID <= 0 {
		return nil
	}
	if apiKey := apiKeyFromContext(ctx); apiKey != nil && apiKey.GroupID != nil && *apiKey.GroupID == *groupID && apiKey.Group != nil {
		return apiKey.Group
	}
	if group, ok := ctx.Value(ctxkey.Group).(*Group); ok && IsGroupContextValid(group) && group.ID == *groupID {
		return group
	}
	return nil
}

func resolveDynamicBudgetMultiplier(ctx context.Context, group *Group) (float64, bool) {
	if group == nil || !group.IsDynamicPricing() {
		return 0, false
	}
	if apiKey := apiKeyFromContext(ctx); apiKey != nil && apiKey.BudgetMultiplier != nil {
		return *apiKey.BudgetMultiplier, true
	}
	if group.DefaultBudgetMultiplier != nil {
		return *group.DefaultBudgetMultiplier, true
	}
	return DefaultBudgetMultiplier, true
}

func withDynamicPricingBudgetState(ctx context.Context, groupID *int64, usageRepo UsageLogRepository) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Value(dynamicPricingBudgetStateKey{}).(*dynamicPricingBudgetState); ok {
		return ctx
	}
	state := buildDynamicPricingBudgetState(ctx, groupID, usageRepo, time.Now())
	return context.WithValue(ctx, dynamicPricingBudgetStateKey{}, state)
}

func dynamicPricingBudgetStateFromContext(ctx context.Context) *dynamicPricingBudgetState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(dynamicPricingBudgetStateKey{}).(*dynamicPricingBudgetState)
	return state
}

func buildDynamicPricingBudgetState(ctx context.Context, groupID *int64, usageRepo UsageLogRepository, now time.Time) *dynamicPricingBudgetState {
	group := dynamicPricingGroupFromContext(ctx, groupID)
	if group == nil || !group.IsDynamicPricing() {
		return &dynamicPricingBudgetState{}
	}

	budgetMultiplier, ok := resolveDynamicBudgetMultiplier(ctx, group)
	if !ok {
		return &dynamicPricingBudgetState{}
	}

	state := &dynamicPricingBudgetState{
		enabled:          true,
		budgetMultiplier: budgetMultiplier,
		windowEnd:        now,
		windowStart:      now.Add(-dynamicBudgetWindow),
	}

	apiKey := apiKeyFromContext(ctx)
	if apiKey == nil {
		return state
	}
	if !apiKey.CreatedAt.IsZero() && apiKey.CreatedAt.After(state.windowStart) {
		state.windowStart = apiKey.CreatedAt
	}
	if usageRepo == nil || apiKey.ID <= 0 {
		return state
	}

	start := state.windowStart
	end := state.windowEnd
	stats, err := usageRepo.GetStatsWithFilters(ctx, usagestats.UsageLogFilters{
		APIKeyID:  apiKey.ID,
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil || stats == nil {
		return state
	}

	state.currentStandardCost = stats.TotalCost
	state.currentActualCost = stats.TotalActualCost
	if stats.TotalCost > 0 {
		state.currentAverageMultiplier = stats.TotalActualCost / stats.TotalCost
	}
	if stats.TotalRequests > 0 && stats.TotalCost > 0 {
		state.estimatedNextStandardCost = stats.TotalCost / float64(stats.TotalRequests)
	}
	return state
}

func dynamicPricingBudgetExceeded(ctx context.Context, groupID *int64) bool {
	if groupID == nil || *groupID <= 0 {
		return false
	}
	state := dynamicPricingBudgetStateFromContext(ctx)
	if state == nil || !state.enabled || state.currentStandardCost <= 0 {
		return false
	}
	return state.currentAverageMultiplier > state.budgetMultiplier+dynamicBudgetEpsilon
}

func validateDynamicPricingBudgetAvailable(ctx context.Context, groupID *int64) error {
	if dynamicPricingBudgetExceeded(ctx, groupID) {
		return ErrDynamicPricingBudgetExceeded
	}
	return nil
}

func resolveDynamicPricingBaseMultiplier(ctx context.Context, groupID *int64, resolveUserGroupRate userGroupRateResolverFunc) (float64, bool) {
	group := dynamicPricingGroupFromContext(ctx, groupID)
	if group == nil || !group.IsDynamicPricing() {
		return 0, false
	}
	baseMultiplier := group.RateMultiplier
	if apiKey := apiKeyFromContext(ctx); apiKey != nil && apiKey.UserID > 0 && resolveUserGroupRate != nil {
		baseMultiplier = resolveUserGroupRate(ctx, apiKey.UserID, *groupID, group.RateMultiplier)
	}
	return baseMultiplier, true
}

func resolveDynamicPricingEffectiveMultiplier(ctx context.Context, groupID *int64, account *Account, resolveUserGroupRate userGroupRateResolverFunc) (float64, bool) {
	if account == nil || groupID == nil || *groupID <= 0 {
		return 0, false
	}
	baseMultiplier, ok := resolveDynamicPricingBaseMultiplier(ctx, groupID, resolveUserGroupRate)
	if !ok {
		return 0, false
	}
	return baseMultiplier * account.GroupBillingMultiplier(groupID), true
}

func compareDynamicPricingAccountPreference(ctx context.Context, groupID *int64, a, b *Account, resolveUserGroupRate userGroupRateResolverFunc) int {
	if a == nil || b == nil {
		return 0
	}
	group := dynamicPricingGroupFromContext(ctx, groupID)
	if group == nil || !group.IsDynamicPricing() {
		return 0
	}
	budgetMultiplier, ok := resolveDynamicBudgetMultiplier(ctx, group)
	if !ok {
		return 0
	}
	aMultiplier, okA := resolveDynamicPricingEffectiveMultiplier(ctx, groupID, a, resolveUserGroupRate)
	bMultiplier, okB := resolveDynamicPricingEffectiveMultiplier(ctx, groupID, b, resolveUserGroupRate)
	if !okA || !okB {
		return 0
	}

	aWithinBudget := aMultiplier <= budgetMultiplier+dynamicBudgetEpsilon
	bWithinBudget := bMultiplier <= budgetMultiplier+dynamicBudgetEpsilon
	switch {
	case aWithinBudget && !bWithinBudget:
		return -1
	case !aWithinBudget && bWithinBudget:
		return 1
	}

	diff := aMultiplier - bMultiplier
	if diff < 0 {
		diff = -diff
	}
	if diff <= dynamicBudgetEpsilon {
		return 0
	}

	if aWithinBudget {
		if aMultiplier > bMultiplier {
			return -1
		}
		return 1
	}

	if aMultiplier < bMultiplier {
		return -1
	}
	return 1
}

func isAccountWithinDynamicBudget(ctx context.Context, groupID *int64, account *Account, resolveUserGroupRate userGroupRateResolverFunc) bool {
	if account == nil || groupID == nil || *groupID <= 0 {
		return true
	}
	group := dynamicPricingGroupFromContext(ctx, groupID)
	if group == nil || !group.IsDynamicPricing() {
		return true
	}
	budgetMultiplier, ok := resolveDynamicBudgetMultiplier(ctx, group)
	if !ok {
		return true
	}

	effectiveMultiplier, ok := resolveDynamicPricingEffectiveMultiplier(ctx, groupID, account, resolveUserGroupRate)
	if !ok {
		return true
	}
	state := dynamicPricingBudgetStateFromContext(ctx)
	if state == nil || !state.enabled || state.estimatedNextStandardCost <= 0 || state.currentStandardCost <= 0 {
		return effectiveMultiplier <= budgetMultiplier+dynamicBudgetEpsilon
	}

	predictedActualCost := state.currentActualCost + state.estimatedNextStandardCost*effectiveMultiplier
	predictedStandardCost := state.currentStandardCost + state.estimatedNextStandardCost
	if predictedStandardCost <= 0 {
		return effectiveMultiplier <= budgetMultiplier+dynamicBudgetEpsilon
	}
	predictedAverageMultiplier := predictedActualCost / predictedStandardCost
	return predictedAverageMultiplier <= budgetMultiplier+dynamicBudgetEpsilon
}
