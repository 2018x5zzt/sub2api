package service

import (
	"context"
	"log"
	"sync"
	"time"
)

type groupHealthHistoryGroupReader interface {
	ListActive(ctx context.Context) ([]Group, error)
}

type GroupHealthHistoryService struct {
	groupRepo    groupHealthHistoryGroupReader
	snapshotRepo GroupHealthSnapshotRepository
	interval     time.Duration
	retention    time.Duration
	stopCh       chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup
	now          func() time.Time
}

func NewGroupHealthHistoryService(
	groupRepo groupHealthHistoryGroupReader,
	snapshotRepo GroupHealthSnapshotRepository,
	interval time.Duration,
	retention time.Duration,
) *GroupHealthHistoryService {
	return &GroupHealthHistoryService{
		groupRepo:    groupRepo,
		snapshotRepo: snapshotRepo,
		interval:     interval,
		retention:    retention,
		stopCh:       make(chan struct{}),
		now:          time.Now,
	}
}

func (s *GroupHealthHistoryService) Start() {
	if s == nil || s.groupRepo == nil || s.snapshotRepo == nil || s.interval <= 0 {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runLoop()
	}()
}

func (s *GroupHealthHistoryService) Stop() {
	if s == nil {
		return
	}

	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *GroupHealthHistoryService) runLoop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runOnce()
		case <-s.stopCh:
			return
		}
	}
}

func (s *GroupHealthHistoryService) runOnce() {
	if s == nil || s.groupRepo == nil || s.snapshotRepo == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		log.Printf("[GroupHealthHistory] list active groups failed: %v", err)
		return
	}

	bucket := s.now().UTC().Truncate(time.Minute)
	snapshots := make([]GroupHealthSnapshot, 0, len(groups))
	for i := range groups {
		health := ComputeGroupPoolHealth(&groups[i])
		snapshots = append(snapshots, GroupHealthSnapshot{
			GroupID:       groups[i].ID,
			BucketTime:    bucket,
			HealthPercent: health.HealthPercent,
			HealthState:   health.HealthState,
		})
	}

	if err := s.snapshotRepo.UpsertBatch(ctx, snapshots); err != nil {
		log.Printf("[GroupHealthHistory] upsert failed: %v", err)
	}
	if _, err := s.snapshotRepo.DeleteBefore(ctx, bucket.Add(-s.retention)); err != nil {
		log.Printf("[GroupHealthHistory] retention cleanup failed: %v", err)
	}
}
