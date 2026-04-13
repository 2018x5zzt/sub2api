//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type settingPromoRepoStub struct {
	values  map[string]string
	getErr  error
	updates map[string]string
}

func (s *settingPromoRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingPromoRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.getErr != nil {
		return "", s.getErr
	}
	value, ok := s.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (s *settingPromoRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingPromoRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingPromoRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.updates = make(map[string]string, len(settings))
	for key, value := range settings {
		s.updates[key] = value
	}
	return nil
}

func (s *settingPromoRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingPromoRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_IsPromoCodeEnabled_DefaultsDisabledOnReadError(t *testing.T) {
	repo := &settingPromoRepoStub{getErr: errors.New("boom")}
	svc := NewSettingService(repo, &config.Config{})

	require.False(t, svc.IsPromoCodeEnabled(context.Background()))
}

func TestSettingService_InitializeDefaultSettings_DisablesPromoCodeByDefault(t *testing.T) {
	repo := &settingPromoRepoStub{values: map[string]string{}}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.InitializeDefaultSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "false", repo.updates[SettingKeyPromoCodeEnabled])
}
