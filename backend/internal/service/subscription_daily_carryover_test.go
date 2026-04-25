package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewDailyWindowAdvance_CarriesForwardOnlyFreshUnusedQuota(t *testing.T) {
	windowStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	group := &Group{DailyLimitUSD: ptrFloat64(45)}
	sub := &UserSubscription{
		Status:                    SubscriptionStatusActive,
		ExpiresAt:                 windowStart.Add(7 * 24 * time.Hour),
		DailyWindowStart:          ptrTime(windowStart),
		DailyUsageUSD:             30,
		DailyCarryoverInUSD:       10,
		DailyCarryoverRemainingUSD: 0,
	}

	preview := previewDailyWindowAdvance(sub, group, windowStart.Add(24*time.Hour))

	assert.Equal(t, 1, preview.ElapsedDays)
	assert.Equal(t, windowStart.Add(24*time.Hour), preview.WindowStart)
	assert.InDelta(t, 25, preview.CarryoverInUSD, 1e-6)
	assert.InDelta(t, 25, preview.CarryoverRemainingUSD, 1e-6)
	assert.InDelta(t, 70, preview.EffectiveLimitUSD, 1e-6)
	assert.InDelta(t, 70, preview.RemainingTotalUSD, 1e-6)
}

func TestPreviewDailyWindowAdvance_DropsCarryoverAfterGap(t *testing.T) {
	windowStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	group := &Group{DailyLimitUSD: ptrFloat64(45)}
	sub := &UserSubscription{
		Status:           SubscriptionStatusActive,
		ExpiresAt:        windowStart.Add(7 * 24 * time.Hour),
		DailyWindowStart: ptrTime(windowStart),
		DailyUsageUSD:    20,
	}

	preview := previewDailyWindowAdvance(sub, group, windowStart.Add(48*time.Hour))

	assert.Equal(t, 2, preview.ElapsedDays)
	assert.InDelta(t, 0, preview.CarryoverInUSD, 1e-6)
	assert.InDelta(t, 0, preview.CarryoverRemainingUSD, 1e-6)
	assert.InDelta(t, 45, preview.EffectiveLimitUSD, 1e-6)
	assert.InDelta(t, 45, preview.RemainingTotalUSD, 1e-6)
}

func TestPreviewDailyWindowAdvance_DropsCarryoverWhenSubscriptionExpiredNextDay(t *testing.T) {
	windowStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	group := &Group{DailyLimitUSD: ptrFloat64(45)}
	sub := &UserSubscription{
		Status:           SubscriptionStatusActive,
		ExpiresAt:        windowStart.Add(24 * time.Hour),
		DailyWindowStart: ptrTime(windowStart),
		DailyUsageUSD:    10,
	}

	preview := previewDailyWindowAdvance(sub, group, windowStart.Add(24*time.Hour))

	assert.Equal(t, 1, preview.ElapsedDays)
	assert.InDelta(t, 0, preview.CarryoverInUSD, 1e-6)
	assert.InDelta(t, 0, preview.CarryoverRemainingUSD, 1e-6)
	assert.InDelta(t, 45, preview.EffectiveLimitUSD, 1e-6)
	assert.InDelta(t, 45, preview.RemainingTotalUSD, 1e-6)
}

func TestValidateAndCheckLimits_DailyCarryoverRaisesEffectiveLimit(t *testing.T) {
	svc := newTestSubscriptionService()
	windowStart := startOfDay(time.Now()).Add(-24 * time.Hour)
	expectedWindowStart := startOfDay(time.Now())
	group := &Group{DailyLimitUSD: ptrFloat64(45)}
	sub := &UserSubscription{
		Status:                    SubscriptionStatusActive,
		ExpiresAt:                 time.Now().Add(72 * time.Hour),
		DailyWindowStart:          ptrTime(windowStart),
		DailyUsageUSD:             30,
		DailyCarryoverInUSD:       10,
		DailyCarryoverRemainingUSD: 0,
	}

	needsMaintenance, err := svc.ValidateAndCheckLimits(sub, group)

	require.NoError(t, err)
	assert.True(t, needsMaintenance)
	require.NotNil(t, sub.DailyWindowStart)
	assert.Equal(t, expectedWindowStart, *sub.DailyWindowStart)
	assert.InDelta(t, 0, sub.DailyUsageUSD, 1e-6)
	assert.InDelta(t, 25, sub.DailyCarryoverInUSD, 1e-6)
	assert.InDelta(t, 25, sub.DailyCarryoverRemainingUSD, 1e-6)
	assert.InDelta(t, 70, sub.DailyEffectiveLimit(group), 1e-6)
	assert.InDelta(t, 70, sub.DailyRemainingTotal(group), 1e-6)
	assert.True(t, sub.CheckDailyLimit(group, 50))
	assert.False(t, sub.CheckDailyLimit(group, 71))
}

func TestNormalizeExpiredWindows_DailyCarryoverForDisplay(t *testing.T) {
	windowStart := startOfDay(time.Now()).Add(-24 * time.Hour)
	expectedWindowStart := startOfDay(time.Now())
	subs := []UserSubscription{
		{
			Status:                    SubscriptionStatusActive,
			ExpiresAt:                 time.Now().Add(72 * time.Hour),
			DailyWindowStart:          ptrTime(windowStart),
			DailyUsageUSD:             30,
			DailyCarryoverInUSD:       10,
			DailyCarryoverRemainingUSD: 0,
			Group: &Group{
				DailyLimitUSD: ptrFloat64(45),
			},
		},
	}

	normalizeExpiredWindows(subs)

	require.NotNil(t, subs[0].DailyWindowStart)
	assert.Equal(t, expectedWindowStart, *subs[0].DailyWindowStart)
	assert.InDelta(t, 0, subs[0].DailyUsageUSD, 1e-6)
	assert.InDelta(t, 25, subs[0].DailyCarryoverInUSD, 1e-6)
	assert.InDelta(t, 25, subs[0].DailyCarryoverRemainingUSD, 1e-6)
}
