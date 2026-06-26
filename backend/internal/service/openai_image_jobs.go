package service

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
)

const (
	OpenAIImageJobStatusQueued    OpenAIImageJobStatus = "queued"
	OpenAIImageJobStatusRunning   OpenAIImageJobStatus = "running"
	OpenAIImageJobStatusSucceeded OpenAIImageJobStatus = "succeeded"
	OpenAIImageJobStatusFailed    OpenAIImageJobStatus = "failed"
	OpenAIImageJobStatusTimeout   OpenAIImageJobStatus = "timeout"

	openAIImageJobWatchdogLeaderLockKey = "openai:image_jobs:watchdog:leader"
)

type OpenAIImageJobStatus string

type OpenAIImageJobOwner struct {
	UserID   int64 `json:"user_id"`
	APIKeyID int64 `json:"api_key_id"`
}

type OpenAIImageJobRecord struct {
	ID          string               `json:"id"`
	Owner       OpenAIImageJobOwner  `json:"owner"`
	Status      OpenAIImageJobStatus `json:"status"`
	StatusCode  int                  `json:"status_code,omitempty"`
	Headers     map[string][]string  `json:"headers,omitempty"`
	Body        []byte               `json:"body,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	CompletedAt time.Time            `json:"completed_at,omitempty"`
}

type OpenAIImageJobStore interface {
	Create(ctx context.Context, record *OpenAIImageJobRecord, ttl time.Duration) error
	Get(ctx context.Context, id string, owner OpenAIImageJobOwner) (*OpenAIImageJobRecord, bool, error)
	SetRunning(ctx context.Context, id string, updatedAt time.Time, ttl time.Duration) (bool, error)
	Complete(ctx context.Context, id string, status OpenAIImageJobStatus, statusCode int, headers map[string][]string, body []byte, completedAt time.Time, ttl time.Duration) error
	MarkStaleTimeouts(ctx context.Context, now time.Time, timeout time.Duration, limit int, ttl time.Duration) (int64, error)
}

type OpenAIImageJobWatchdog struct {
	store    OpenAIImageJobStore
	interval time.Duration
	timeout  time.Duration
	ttl      time.Duration
	batch    int
	lockTTL  time.Duration

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string

	startOnce sync.Once
	stopOnce  sync.Once
	stopCh    chan struct{}
}

func NewOpenAIImageJobWatchdog(store OpenAIImageJobStore, lockCache LeaderLockCache, db *sql.DB, cfg *config.Config) *OpenAIImageJobWatchdog {
	jobCfg := config.DefaultOpenAIImageJobConfig()
	if cfg != nil {
		jobCfg = cfg.Gateway.OpenAIImageJobs.WithDefaults()
	}
	return &OpenAIImageJobWatchdog{
		store:      store,
		interval:   time.Duration(jobCfg.WatchdogIntervalSeconds) * time.Second,
		timeout:    time.Duration(jobCfg.TimeoutSeconds) * time.Second,
		ttl:        time.Duration(jobCfg.TTLSeconds) * time.Second,
		batch:      jobCfg.WatchdogBatchSize,
		lockTTL:    time.Duration(jobCfg.WatchdogLockTTLSeconds) * time.Second,
		lockCache:  lockCache,
		db:         db,
		instanceID: uuid.NewString(),
		stopCh:     make(chan struct{}),
	}
}

func (s *OpenAIImageJobWatchdog) Start() {
	if s == nil || s.store == nil || s.interval <= 0 {
		return
	}
	s.startOnce.Do(func() {
		logger.LegacyPrintf("service.openai_image_job_watchdog", "[OpenAIImageJobWatchdog] started interval=%s timeout=%s batch=%d", s.interval, s.timeout, s.batch)
		go s.runLoop()
	})
}

func (s *OpenAIImageJobWatchdog) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.stopCh != nil {
			close(s.stopCh)
		}
		logger.LegacyPrintf("service.openai_image_job_watchdog", "[OpenAIImageJobWatchdog] stopped")
	})
}

func (s *OpenAIImageJobWatchdog) runLoop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.runOnce()
	for {
		select {
		case <-ticker.C:
			s.runOnce()
		case <-s.stopCh:
			return
		}
	}
}

func (s *OpenAIImageJobWatchdog) runOnce() {
	if s == nil || s.store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, openAIImageJobWatchdogLeaderLockKey, s.instanceID, s.lockTTL)
	if !ok {
		return
	}
	defer release()

	changed, err := s.store.MarkStaleTimeouts(ctx, time.Now(), s.timeout, s.batch, s.ttl)
	if err != nil {
		logger.LegacyPrintf("service.openai_image_job_watchdog", "[OpenAIImageJobWatchdog] scan failed err=%v", err)
		return
	}
	if changed > 0 {
		logger.LegacyPrintf("service.openai_image_job_watchdog", "[OpenAIImageJobWatchdog] marked stale jobs timeout count=%d", changed)
	}
}

func OpenAIImageJobTimeoutHeaders() map[string][]string {
	return http.Header{"Content-Type": []string{"application/json"}}
}
