package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGroupHealthHistoryService_RunOnceStoresCurrentSnapshots(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 10, 12, 34, 56, 0, time.UTC)
	groupRepo := &groupHealthHistoryGroupReaderStub{
		groups: []Group{
			{
				ID:                      11,
				ActiveAccountCount:      5,
				RateLimitedAccountCount: 2,
			},
			{
				ID:                      22,
				ActiveAccountCount:      0,
				RateLimitedAccountCount: 0,
			},
		},
	}
	snapshotRepo := &groupHealthHistorySnapshotRepoStub{}
	svc := NewGroupHealthHistoryService(groupRepo, snapshotRepo, time.Minute, 30*24*time.Hour)
	svc.now = func() time.Time { return now }

	svc.runOnce()

	require.Len(t, snapshotRepo.upsertedBatches, 1)
	require.Equal(t, []GroupHealthSnapshot{
		{
			GroupID:       11,
			BucketTime:    now.Truncate(time.Minute),
			HealthPercent: 60,
			HealthState:   "degraded",
		},
		{
			GroupID:       22,
			BucketTime:    now.Truncate(time.Minute),
			HealthPercent: 0,
			HealthState:   "down",
		},
	}, snapshotRepo.upsertedBatches[0])
}

func TestGroupHealthHistoryService_RunOnceDeletesSnapshotsOlderThanThirtyDays(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 10, 12, 34, 56, 0, time.UTC)
	groupRepo := &groupHealthHistoryGroupReaderStub{
		groups: []Group{
			{
				ID:                      33,
				ActiveAccountCount:      3,
				RateLimitedAccountCount: 0,
			},
		},
	}
	snapshotRepo := &groupHealthHistorySnapshotRepoStub{}
	svc := NewGroupHealthHistoryService(groupRepo, snapshotRepo, time.Minute, 30*24*time.Hour)
	svc.now = func() time.Time { return now }

	svc.runOnce()

	require.Len(t, snapshotRepo.deleteBeforeInputs, 1)
	require.Equal(t, now.Truncate(time.Minute).Add(-30*24*time.Hour), snapshotRepo.deleteBeforeInputs[0])
}

type groupHealthHistoryGroupReaderStub struct {
	groups []Group
	err    error
}

func (s *groupHealthHistoryGroupReaderStub) ListActive(context.Context) ([]Group, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]Group(nil), s.groups...), nil
}

type groupHealthHistorySnapshotRepoStub struct {
	upsertedBatches    [][]GroupHealthSnapshot
	deleteBeforeInputs []time.Time
	upsertErr          error
	deleteErr          error
}

func (s *groupHealthHistorySnapshotRepoStub) UpsertBatch(_ context.Context, snapshots []GroupHealthSnapshot) error {
	s.upsertedBatches = append(s.upsertedBatches, append([]GroupHealthSnapshot(nil), snapshots...))
	return s.upsertErr
}

func (s *groupHealthHistorySnapshotRepoStub) ListRecentByGroupIDs(context.Context, []int64, time.Time) (map[int64][]GroupHealthSnapshot, error) {
	return nil, nil
}

func (s *groupHealthHistorySnapshotRepoStub) DeleteBefore(_ context.Context, cutoff time.Time) (int, error) {
	s.deleteBeforeInputs = append(s.deleteBeforeInputs, cutoff)
	if s.deleteErr != nil {
		return 0, s.deleteErr
	}
	return 0, nil
}
