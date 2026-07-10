package service

import (
	"context"
	"time"
)

// XlabOAuthCodeStore persists single-use xlab OAuth authorization codes.
type XlabOAuthCodeStore interface {
	StoreCode(ctx context.Context, tokenID string, ttl time.Duration) error
	ConsumeCode(ctx context.Context, tokenID string) (bool, error)
}
