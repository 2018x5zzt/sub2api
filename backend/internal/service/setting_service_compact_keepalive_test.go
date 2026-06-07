//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

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
