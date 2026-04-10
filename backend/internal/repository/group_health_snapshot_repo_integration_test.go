//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupHealthSnapshotRepository_UpsertAndListRecentByGroupIDs(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	repo := newGroupHealthSnapshotRepositoryWithSQL(tx.Client(), tx)

	groupAID := mustCreateSnapshotGroup(t, ctx, tx.Client(), "snapshot-group-a")
	groupBID := mustCreateSnapshotGroup(t, ctx, tx.Client(), "snapshot-group-b")

	firstBucket := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	secondBucket := firstBucket.Add(time.Minute)

	err := repo.UpsertBatch(ctx, []service.GroupHealthSnapshot{
		{
			GroupID:       groupAID,
			BucketTime:    firstBucket,
			HealthPercent: 40,
			HealthState:   "degraded",
		},
		{
			GroupID:       groupAID,
			BucketTime:    secondBucket,
			HealthPercent: 75,
			HealthState:   "healthy",
		},
		{
			GroupID:       groupBID,
			BucketTime:    secondBucket,
			HealthPercent: 0,
			HealthState:   "down",
		},
	})
	require.NoError(t, err)

	err = repo.UpsertBatch(ctx, []service.GroupHealthSnapshot{
		{
			GroupID:       groupAID,
			BucketTime:    secondBucket,
			HealthPercent: 80,
			HealthState:   "healthy",
		},
	})
	require.NoError(t, err)

	got, err := repo.ListRecentByGroupIDs(ctx, []int64{groupAID, groupBID}, firstBucket.Add(-time.Second))
	require.NoError(t, err)
	require.Len(t, got, 2)

	require.Equal(t, []service.GroupHealthSnapshot{
		{
			GroupID:       groupAID,
			BucketTime:    firstBucket,
			HealthPercent: 40,
			HealthState:   "degraded",
		},
		{
			GroupID:       groupAID,
			BucketTime:    secondBucket,
			HealthPercent: 80,
			HealthState:   "healthy",
		},
	}, got[groupAID])

	require.Equal(t, []service.GroupHealthSnapshot{
		{
			GroupID:       groupBID,
			BucketTime:    secondBucket,
			HealthPercent: 0,
			HealthState:   "down",
		},
	}, got[groupBID])
}

func TestGroupHealthSnapshotRepository_DeleteBefore(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	repo := newGroupHealthSnapshotRepositoryWithSQL(tx.Client(), tx)

	groupID := mustCreateSnapshotGroup(t, ctx, tx.Client(), "snapshot-group-delete")

	firstBucket := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	cutoff := firstBucket.Add(time.Minute)
	lastBucket := cutoff.Add(time.Minute)

	err := repo.UpsertBatch(ctx, []service.GroupHealthSnapshot{
		{
			GroupID:       groupID,
			BucketTime:    firstBucket,
			HealthPercent: 30,
			HealthState:   "degraded",
		},
		{
			GroupID:       groupID,
			BucketTime:    cutoff,
			HealthPercent: 60,
			HealthState:   "degraded",
		},
		{
			GroupID:       groupID,
			BucketTime:    lastBucket,
			HealthPercent: 100,
			HealthState:   "healthy",
		},
	})
	require.NoError(t, err)

	deleted, err := repo.DeleteBefore(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, 1, deleted)

	got, err := repo.ListRecentByGroupIDs(ctx, []int64{groupID}, firstBucket.Add(-time.Second))
	require.NoError(t, err)
	require.Equal(t, []service.GroupHealthSnapshot{
		{
			GroupID:       groupID,
			BucketTime:    cutoff,
			HealthPercent: 60,
			HealthState:   "degraded",
		},
		{
			GroupID:       groupID,
			BucketTime:    lastBucket,
			HealthPercent: 100,
			HealthState:   "healthy",
		},
	}, got[groupID])
}

func mustCreateSnapshotGroup(t *testing.T, ctx context.Context, client *dbent.Client, name string) int64 {
	t.Helper()

	group, err := client.Group.Create().
		SetName(name).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	return group.ID
}
