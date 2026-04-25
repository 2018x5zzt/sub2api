//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type maintenanceUserSubRepoStub struct {
	userSubRepoNoop

	resetDailyCalled   bool
	resetWeeklyCalled  bool
	resetMonthlyCalled bool
}

func (r *maintenanceUserSubRepoStub) ResetDailyUsage(_ context.Context, _ int64, _ time.Time) error {
	r.resetDailyCalled = true
	return nil
}

func (r *maintenanceUserSubRepoStub) ResetWeeklyUsage(_ context.Context, _ int64, _ time.Time) error {
	r.resetWeeklyCalled = true
	return nil
}

func (r *maintenanceUserSubRepoStub) ResetMonthlyUsage(_ context.Context, _ int64, _ time.Time) error {
	r.resetMonthlyCalled = true
	return nil
}

func newMaintenanceSvc(stub *maintenanceUserSubRepoStub) *SubscriptionService {
	return NewSubscriptionService(groupRepoNoop{}, stub, nil, nil, nil)
}

func TestCheckAndResetWindows_SkipsDailyResetPersistence(t *testing.T) {
	stub := &maintenanceUserSubRepoStub{}
	svc := newMaintenanceSvc(stub)
	windowStart := startOfDay(time.Now()).Add(-24 * time.Hour)
	sub := &UserSubscription{
		ID:                         1,
		UserID:                     10,
		GroupID:                    20,
		DailyWindowStart:           &windowStart,
		DailyUsageUSD:              18,
		DailyCarryoverInUSD:        9,
		DailyCarryoverRemainingUSD: 4,
	}

	err := svc.CheckAndResetWindows(context.Background(), sub)

	require.NoError(t, err)
	require.False(t, stub.resetDailyCalled, "daily rollover should not persist carryover state in maintenance")
	require.False(t, stub.resetWeeklyCalled)
	require.False(t, stub.resetMonthlyCalled)
	require.InDelta(t, 18, sub.DailyUsageUSD, 1e-6)
	require.InDelta(t, 9, sub.DailyCarryoverInUSD, 1e-6)
	require.InDelta(t, 4, sub.DailyCarryoverRemainingUSD, 1e-6)
}

func TestCheckAndResetWindows_StillResetsWeeklyAndMonthly(t *testing.T) {
	stub := &maintenanceUserSubRepoStub{}
	svc := newMaintenanceSvc(stub)
	windowStart := startOfDay(time.Now()).Add(-31 * 24 * time.Hour)
	sub := &UserSubscription{
		ID:                2,
		UserID:            10,
		GroupID:           20,
		DailyWindowStart:  &windowStart,
		WeeklyWindowStart: &windowStart,
		MonthlyWindowStart: &windowStart,
		WeeklyUsageUSD:    16,
		MonthlyUsageUSD:   42,
	}

	err := svc.CheckAndResetWindows(context.Background(), sub)

	require.NoError(t, err)
	require.False(t, stub.resetDailyCalled)
	require.True(t, stub.resetWeeklyCalled, "weekly maintenance should remain active")
	require.True(t, stub.resetMonthlyCalled, "monthly maintenance should remain active")
	require.InDelta(t, 0, sub.WeeklyUsageUSD, 1e-6)
	require.InDelta(t, 0, sub.MonthlyUsageUSD, 1e-6)
}
