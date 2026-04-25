package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildBackfillReport_SkipsDuplicateActiveProductSubscriptions(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	input := BackfillInput{
		ProductID:      101,
		SourceGroupIDs: []int64{88},
		LegacySubscriptions: []LegacySubscriptionRow{
			{ID: 1, UserID: 10, GroupID: 88, Status: "active", StartsAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour)},
		},
		ExistingProductSubscriptions: []ExistingProductSubscriptionRow{
			{ID: 501, UserID: 10, ProductID: 101, Status: "active"},
			{ID: 502, UserID: 10, ProductID: 101, Status: "active"},
		},
	}

	report := BuildBackfillReport(input)

	require.Len(t, report.Skipped, 1)
	require.Equal(t, SkipReasonDuplicateActiveProductSubscription, report.Skipped[0].Reason)
	require.Empty(t, report.Planned)
}

func TestBuildBackfillReport_RequiresDeterministicLegacySource(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	input := BackfillInput{
		ProductID:      101,
		SourceGroupIDs: []int64{88},
		LegacySubscriptions: []LegacySubscriptionRow{
			{ID: 1, UserID: 10, GroupID: 88, Status: "active", StartsAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour)},
			{ID: 2, UserID: 10, GroupID: 88, Status: "active", StartsAt: now, ExpiresAt: now.Add(20 * 24 * time.Hour)},
		},
	}

	report := BuildBackfillReport(input)

	require.Len(t, report.Skipped, 2)
	require.Equal(t, SkipReasonAmbiguousLegacySource, report.Skipped[0].Reason)
	require.Equal(t, SkipReasonAmbiguousLegacySource, report.Skipped[1].Reason)
	require.Empty(t, report.Planned)
}

func TestApplyBackfill_IdempotentWhenMigrationBatchReruns(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	state := &InMemoryBackfillState{
		NextProductSubscriptionID: 1000,
		ProductID:                 101,
		SourceGroupIDs:            []int64{88},
		LegacySubscriptions: []LegacySubscriptionRow{
			{ID: 1, UserID: 10, GroupID: 88, Status: "active", StartsAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour), MonthlyUsageUSD: 12.5},
		},
	}
	opts := BackfillOptions{MigrationBatch: "2026-04-25-gpt-monthly"}

	first, err := ApplyBackfill(state, opts)
	require.NoError(t, err)
	require.Len(t, first.Applied, 1)
	require.Equal(t, int64(1000), first.Applied[0].ProductSubscriptionID)

	second, err := ApplyBackfill(state, opts)
	require.NoError(t, err)
	require.Empty(t, second.Applied)
	require.Len(t, second.Skipped, 1)
	require.Equal(t, SkipReasonAlreadyMigrated, second.Skipped[0].Reason)
	require.Len(t, state.ProductSubscriptions, 1)
	require.Len(t, state.MigrationSources, 1)
}
