package service

import (
	"context"
	"time"
)

type GroupHealthSnapshot struct {
	GroupID       int64
	BucketTime    time.Time
	HealthPercent int
	HealthState   string
}

type GroupHealthSnapshotRepository interface {
	UpsertBatch(ctx context.Context, snapshots []GroupHealthSnapshot) error
	ListRecentByGroupIDs(ctx context.Context, groupIDs []int64, since time.Time) (map[int64][]GroupHealthSnapshot, error)
	DeleteBefore(ctx context.Context, cutoff time.Time) (int, error)
}
