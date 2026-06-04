//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type compactKeepaliveRepoStub struct {
	values map[string]string
}

func (s *compactKeepaliveRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *compactKeepaliveRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.values == nil {
		return "", ErrSettingNotFound
	}
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}

func (s *compactKeepaliveRepoStub) Set(ctx context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *compactKeepaliveRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *compactKeepaliveRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *compactKeepaliveRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *compactKeepaliveRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_CompactHeartbeatKeepaliveSettings_Defaults(t *testing.T) {
	repo := &compactKeepaliveRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	got, err := svc.GetOpenAICompactHeartbeatKeepaliveSettings(context.Background())
	require.NoError(t, err)
	require.NotNil(t, got)
	require.False(t, got.Enabled)
	require.Equal(t, 85, got.StartAfterSeconds)
	require.Equal(t, 25, got.IntervalSeconds)
}

func TestSettingService_CompactHeartbeatKeepaliveSettings_SetAndGet(t *testing.T) {
	repo := &compactKeepaliveRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.SetOpenAICompactHeartbeatKeepaliveSettings(context.Background(), &OpenAICompactHeartbeatKeepaliveSettings{
		Enabled:           true,
		StartAfterSeconds: 85,
		IntervalSeconds:   25,
	})
	require.NoError(t, err)

	got, err := svc.GetOpenAICompactHeartbeatKeepaliveSettings(context.Background())
	require.NoError(t, err)
	require.True(t, got.Enabled)
	require.Equal(t, 85, got.StartAfterSeconds)
	require.Equal(t, 25, got.IntervalSeconds)
}

func TestSettingService_CompactHeartbeatKeepaliveSettings_SetRejectsOutOfRange(t *testing.T) {
	repo := &compactKeepaliveRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.SetOpenAICompactHeartbeatKeepaliveSettings(context.Background(), &OpenAICompactHeartbeatKeepaliveSettings{
		Enabled:           true,
		StartAfterSeconds: 10,
		IntervalSeconds:   25,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "start_after_seconds")

	err = svc.SetOpenAICompactHeartbeatKeepaliveSettings(context.Background(), &OpenAICompactHeartbeatKeepaliveSettings{
		Enabled:           true,
		StartAfterSeconds: 85,
		IntervalSeconds:   2,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "interval_seconds")
}
