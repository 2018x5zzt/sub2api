package repository

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const xlabOAuthProviderCodeKeyPrefix = "xlab:oauth:code:"

type xlabOAuthCodeStore struct {
	rdb *redis.Client
}

func NewXlabOAuthCodeStore(rdb *redis.Client) service.XlabOAuthCodeStore {
	return &xlabOAuthCodeStore{rdb: rdb}
}

func (s *xlabOAuthCodeStore) StoreCode(ctx context.Context, tokenID string, ttl time.Duration) error {
	return s.rdb.Set(ctx, xlabOAuthProviderCodeKeyPrefix+strings.TrimSpace(tokenID), "1", ttl).Err()
}

func (s *xlabOAuthCodeStore) ConsumeCode(ctx context.Context, tokenID string) (bool, error) {
	_, err := s.rdb.GetDel(ctx, xlabOAuthProviderCodeKeyPrefix+strings.TrimSpace(tokenID)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
