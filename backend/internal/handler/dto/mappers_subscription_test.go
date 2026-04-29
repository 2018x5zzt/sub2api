package dto

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserSubscriptionFromService_LegacySubscriptionDoesNotExposeCarryover(t *testing.T) {
	dailyLimit := 45.0
	sub := &service.UserSubscription{
		ID:                         1,
		UserID:                     2,
		GroupID:                    3,
		Status:                     service.SubscriptionStatusActive,
		ExpiresAt:                  time.Now().Add(24 * time.Hour),
		DailyUsageUSD:              20,
		DailyCarryoverInUSD:        15,
		DailyCarryoverRemainingUSD: 10,
		Group: &service.Group{
			DailyLimitUSD: &dailyLimit,
		},
	}

	out := UserSubscriptionFromService(sub)

	require.NotNil(t, out)
	require.InDelta(t, 0, out.DailyCarryoverInUSD, 1e-6)
	require.InDelta(t, 45, out.DailyEffectiveLimitUSD, 1e-6)
	require.InDelta(t, 25, out.DailyRemainingTotalUSD, 1e-6)
	require.InDelta(t, 0, out.DailyRemainingCarryoverUSD, 1e-6)
}
