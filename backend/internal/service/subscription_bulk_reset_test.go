//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type bulkResetUserSubRepoStub struct {
	userSubRepoNoop

	listCalls        int
	lastStatus       string
	lastGroupID      *int64
	resetCalls       []int64
	resetWindowStart []time.Time
	pages            map[int][]UserSubscription
}

func (r *bulkResetUserSubRepoStub) List(_ context.Context, params pagination.PaginationParams, _ *int64, groupID *int64, status, _, _, _ string) ([]UserSubscription, *pagination.PaginationResult, error) {
	r.listCalls++
	r.lastStatus = status
	r.lastGroupID = groupID

	subs := r.pages[params.Page]
	total := 0
	for _, pageSubs := range r.pages {
		total += len(pageSubs)
	}

	return subs, &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    len(r.pages),
	}, nil
}

func (r *bulkResetUserSubRepoStub) ResetDailyUsage(_ context.Context, id int64, windowStart time.Time) error {
	r.resetCalls = append(r.resetCalls, id)
	r.resetWindowStart = append(r.resetWindowStart, windowStart)
	return nil
}

func newBulkResetSvc(repo *bulkResetUserSubRepoStub, cache *billingEligibilityCacheStub) *SubscriptionService {
	cfg := &config.Config{}
	cfg.SubscriptionCache.L1Size = 16
	cfg.SubscriptionCache.L1TTLSeconds = 60
	svc := NewSubscriptionService(groupRepoNoop{}, repo, NewBillingCacheService(cache, nil, nil, nil, &config.Config{}), nil, cfg)
	return svc
}

func TestAdminBulkResetDailyQuota_ResetsActiveSubscriptionsAndInvalidatesCaches(t *testing.T) {
	repo := &bulkResetUserSubRepoStub{
		pages: map[int][]UserSubscription{
			1: {
				{ID: 11, UserID: 101, GroupID: 201},
				{ID: 12, UserID: 102, GroupID: 202},
			},
			2: {
				{ID: 13, UserID: 103, GroupID: 203},
			},
		},
	}
	cache := &billingEligibilityCacheStub{}
	svc := newBulkResetSvc(repo, cache)
	t.Cleanup(svc.billingCacheService.Stop)

	// Seed L1 cache to verify compensation clears the local read cache.
	_ = svc.subCacheL1.SetWithTTL(subCacheKey(101, 201), &UserSubscription{ID: 11}, 1, time.Minute)
	_ = svc.subCacheL1.SetWithTTL(subCacheKey(102, 202), &UserSubscription{ID: 12}, 1, time.Minute)
	svc.subCacheL1.Wait()

	result, err := svc.AdminBulkResetDailyQuota(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(3), result.ResetCount)
	require.Equal(t, SubscriptionStatusActive, repo.lastStatus)
	require.Nil(t, repo.lastGroupID)
	require.Equal(t, []int64{11, 12, 13}, repo.resetCalls)
	require.Len(t, repo.resetWindowStart, 3)
	require.Equal(t, 3, cache.subscriptionInvalidates)

	_, ok := svc.subCacheL1.Get(subCacheKey(101, 201))
	require.False(t, ok, "bulk daily reset should evict local L1 cache entries")
	_, ok = svc.subCacheL1.Get(subCacheKey(102, 202))
	require.False(t, ok, "bulk daily reset should evict local L1 cache entries")
}

func TestAdminBulkResetDailyQuota_PropagatesGroupFilter(t *testing.T) {
	groupID := int64(88)
	repo := &bulkResetUserSubRepoStub{
		pages: map[int][]UserSubscription{
			1: {
				{ID: 21, UserID: 301, GroupID: groupID},
			},
		},
	}
	svc := newBulkResetSvc(repo, &billingEligibilityCacheStub{})
	t.Cleanup(svc.billingCacheService.Stop)

	result, err := svc.AdminBulkResetDailyQuota(context.Background(), &groupID)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(1), result.ResetCount)
	require.NotNil(t, repo.lastGroupID)
	require.Equal(t, groupID, *repo.lastGroupID)
}
