package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type billingEligibilityCacheStub struct {
	subscriptionGets        int
	userBalanceGets         int
	subscriptionInvalidates int
	subscriptionUpdates     int

	subscriptionData *SubscriptionCacheData
	subscriptionErr  error
	userBalance      float64
	userBalanceErr   error
}

func (s *billingEligibilityCacheStub) GetUserBalance(ctx context.Context, userID int64) (float64, error) {
	s.userBalanceGets++
	return s.userBalance, s.userBalanceErr
}

func (s *billingEligibilityCacheStub) SetUserBalance(ctx context.Context, userID int64, balance float64) error {
	return nil
}

func (s *billingEligibilityCacheStub) DeductUserBalance(ctx context.Context, userID int64, amount float64) error {
	return nil
}

func (s *billingEligibilityCacheStub) InvalidateUserBalance(ctx context.Context, userID int64) error {
	return nil
}

func (s *billingEligibilityCacheStub) GetSubscriptionCache(ctx context.Context, userID, groupID int64) (*SubscriptionCacheData, error) {
	s.subscriptionGets++
	if s.subscriptionData != nil || s.subscriptionErr == nil {
		return s.subscriptionData, s.subscriptionErr
	}
	return nil, s.subscriptionErr
}

func (s *billingEligibilityCacheStub) SetSubscriptionCache(ctx context.Context, userID, groupID int64, data *SubscriptionCacheData) error {
	return nil
}

func (s *billingEligibilityCacheStub) UpdateSubscriptionUsage(ctx context.Context, userID, groupID int64, cost float64) error {
	s.subscriptionUpdates++
	return nil
}

func (s *billingEligibilityCacheStub) InvalidateSubscriptionCache(ctx context.Context, userID, groupID int64) error {
	s.subscriptionInvalidates++
	return nil
}

func (s *billingEligibilityCacheStub) GetAPIKeyRateLimit(ctx context.Context, keyID int64) (*APIKeyRateLimitCacheData, error) {
	return nil, errors.New("not implemented")
}

func (s *billingEligibilityCacheStub) SetAPIKeyRateLimit(ctx context.Context, keyID int64, data *APIKeyRateLimitCacheData) error {
	return nil
}

func (s *billingEligibilityCacheStub) UpdateAPIKeyRateLimitUsage(ctx context.Context, keyID int64, cost float64) error {
	return nil
}

func (s *billingEligibilityCacheStub) InvalidateAPIKeyRateLimit(ctx context.Context, keyID int64) error {
	return nil
}

type billingEligibilityUserRepoStub struct {
	balance float64
}

func (s billingEligibilityUserRepoStub) Create(context.Context, *User) error { panic("unexpected Create call") }
func (s billingEligibilityUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return &User{Balance: s.balance}, nil
}
func (s billingEligibilityUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}
func (s billingEligibilityUserRepoStub) GetByInviteCode(context.Context, string) (*User, error) {
	panic("unexpected GetByInviteCode call")
}
func (s billingEligibilityUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}
func (s billingEligibilityUserRepoStub) Update(context.Context, *User) error { panic("unexpected Update call") }
func (s billingEligibilityUserRepoStub) Delete(context.Context, int64) error { panic("unexpected Delete call") }
func (s billingEligibilityUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s billingEligibilityUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s billingEligibilityUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected UpdateBalance call")
}
func (s billingEligibilityUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}
func (s billingEligibilityUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}
func (s billingEligibilityUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}
func (s billingEligibilityUserRepoStub) ExistsByInviteCode(context.Context, string) (bool, error) {
	panic("unexpected ExistsByInviteCode call")
}
func (s billingEligibilityUserRepoStub) CountInviteesByInviter(context.Context, int64) (int64, error) {
	panic("unexpected CountInviteesByInviter call")
}
func (s billingEligibilityUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}
func (s billingEligibilityUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}
func (s billingEligibilityUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}
func (s billingEligibilityUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}
func (s billingEligibilityUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp call")
}
func (s billingEligibilityUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

type billingEligibilitySubRepoStub struct {
	UserSubscriptionRepository
	sub *UserSubscription
	err error
}

func (s billingEligibilitySubRepoStub) GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	cp := *s.sub
	return &cp, nil
}

func TestBillingCacheServiceCheckSubscriptionEligibilityFromSnapshot_UsesCarryoverAwareLimit(t *testing.T) {
	cache := &billingEligibilityCacheStub{
		subscriptionErr: errors.New("subscription cache should not be read"),
	}
	svc := NewBillingCacheService(cache, billingEligibilityUserRepoStub{balance: 0}, billingEligibilitySubRepoStub{}, nil, &config.Config{})
	t.Cleanup(svc.Stop)

	group := &Group{
		ID:               88,
		SubscriptionType: SubscriptionTypeSubscription,
		DailyLimitUSD:    ptrFloat64(45),
	}
	sub := &UserSubscription{
		Status:                    SubscriptionStatusActive,
		ExpiresAt:                 time.Now().Add(24 * time.Hour),
		DailyWindowStart:          ptrTime(time.Now().UTC().Truncate(24 * time.Hour)),
		DailyUsageUSD:             50,
		DailyCarryoverInUSD:       15,
		DailyCarryoverRemainingUSD: 10,
	}

	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, group, sub)

	require.NoError(t, err)
	require.Equal(t, 0, cache.subscriptionGets)
	require.Equal(t, 0, cache.userBalanceGets)
}

func TestBillingCacheServiceCheckBillingEligibility_FallsBackToRepoWhenSnapshotMissing(t *testing.T) {
	subRepo := billingEligibilitySubRepoStub{
		sub: &UserSubscription{
			Status:        SubscriptionStatusActive,
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			DailyUsageUSD: 10,
		},
	}
	svc := NewBillingCacheService(nil, billingEligibilityUserRepoStub{balance: 0}, subRepo, nil, &config.Config{})
	t.Cleanup(svc.Stop)

	group := &Group{
		ID:               99,
		SubscriptionType: SubscriptionTypeSubscription,
		DailyLimitUSD:    ptrFloat64(45),
	}

	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, group, nil)

	require.NoError(t, err)
}
