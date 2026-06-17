package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type openAIImageJobWatchdogStoreStub struct {
	calls       int
	lastTimeout time.Duration
	lastLimit   int
	lastTTL     time.Duration
}

func (s *openAIImageJobWatchdogStoreStub) Create(context.Context, *OpenAIImageJobRecord, time.Duration) error {
	return nil
}

func (s *openAIImageJobWatchdogStoreStub) Get(context.Context, string, OpenAIImageJobOwner) (*OpenAIImageJobRecord, bool, error) {
	return nil, false, nil
}

func (s *openAIImageJobWatchdogStoreStub) SetRunning(context.Context, string, time.Time, time.Duration) (bool, error) {
	return true, nil
}

func (s *openAIImageJobWatchdogStoreStub) Complete(context.Context, string, OpenAIImageJobStatus, int, map[string][]string, []byte, time.Time, time.Duration) error {
	return nil
}

func (s *openAIImageJobWatchdogStoreStub) MarkStaleTimeouts(_ context.Context, _ time.Time, timeout time.Duration, limit int, ttl time.Duration) (int64, error) {
	s.calls++
	s.lastTimeout = timeout
	s.lastLimit = limit
	s.lastTTL = ttl
	return 2, nil
}

func TestNewOpenAIImageJobWatchdogUsesGatewayConfig(t *testing.T) {
	store := &openAIImageJobWatchdogStoreStub{}
	svc := NewOpenAIImageJobWatchdog(store, nil, nil, &config.Config{
		Gateway: config.GatewayConfig{
			OpenAIImageJobs: config.OpenAIImageJobConfig{
				TimeoutSeconds:          1200,
				TTLSeconds:              86400,
				WatchdogIntervalSeconds: 7,
				WatchdogBatchSize:       123,
				WatchdogLockTTLSeconds:  45,
			},
		},
	})

	require.Equal(t, 7*time.Second, svc.interval)
	require.Equal(t, 123, svc.batch)
	require.Equal(t, 20*time.Minute, svc.timeout)
	require.Equal(t, 24*time.Hour, svc.ttl)
	require.Equal(t, 45*time.Second, svc.lockTTL)
}

func TestOpenAIImageJobWatchdogRunOnceMarksStaleTimeouts(t *testing.T) {
	store := &openAIImageJobWatchdogStoreStub{}
	svc := NewOpenAIImageJobWatchdog(store, nil, nil, &config.Config{
		Gateway: config.GatewayConfig{
			OpenAIImageJobs: config.OpenAIImageJobConfig{
				TimeoutSeconds:          60,
				TTLSeconds:              3600,
				WatchdogIntervalSeconds: 10,
				WatchdogBatchSize:       77,
			},
		},
	})

	svc.runOnce()

	require.Equal(t, 1, store.calls)
	require.Equal(t, time.Minute, store.lastTimeout)
	require.Equal(t, 77, store.lastLimit)
	require.Equal(t, time.Hour, store.lastTTL)
}
