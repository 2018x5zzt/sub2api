package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionProductService_GetActiveProductSubscription(t *testing.T) {
	repo := &subscriptionProductRepoStub{
		binding: &SubscriptionProductBinding{
			ProductID:       10,
			ProductCode:     "gpt_monthly",
			ProductName:     "GPT Monthly",
			ProductStatus:   "active",
			GroupID:         20,
			GroupName:       "pro",
			BindingStatus:   "active",
			DebitMultiplier: 1.5,
		},
		subscription: &UserProductSubscription{
			ID:        30,
			UserID:    40,
			ProductID: 10,
			Status:    SubscriptionStatusActive,
			StartsAt:  time.Now().Add(-time.Hour),
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	svc := NewSubscriptionProductService(repo)

	got, err := svc.GetActiveProductSubscription(context.Background(), 40, 20)

	require.NoError(t, err)
	require.Equal(t, int64(10), got.Binding.ProductID)
	require.Equal(t, int64(30), got.Subscription.ID)
	require.Equal(t, 1.5, got.Binding.DebitMultiplier)
	require.Equal(t, 1, repo.bindingCalls)
	require.Equal(t, 1, repo.subscriptionCalls)
}

func TestSubscriptionProductService_GetActiveProductSubscriptionSharedGroupUsesUserProduct(t *testing.T) {
	repo := &subscriptionProductRepoStub{
		binding: &SubscriptionProductBinding{
			ProductID:       10,
			ProductCode:     "gpt_daily_45",
			ProductName:     "GPT Daily 45",
			ProductStatus:   SubscriptionProductStatusActive,
			GroupID:         20,
			GroupName:       "pro",
			BindingStatus:   SubscriptionProductBindingStatusActive,
			DebitMultiplier: 1,
		},
		bindingByUserGroup: &SubscriptionProductBinding{
			ProductID:       11,
			ProductCode:     "gpt_daily_150",
			ProductName:     "GPT Daily 150",
			ProductStatus:   SubscriptionProductStatusActive,
			GroupID:         20,
			GroupName:       "pro",
			BindingStatus:   SubscriptionProductBindingStatusActive,
			DebitMultiplier: 1.5,
		},
		subscriptionByUserGroup: &UserProductSubscription{
			ID:        31,
			UserID:    40,
			ProductID: 11,
			Status:    SubscriptionStatusActive,
			StartsAt:  time.Now().Add(-time.Hour),
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	repo.getSubscription = func(_ context.Context, _ int64, productID int64) (*UserProductSubscription, error) {
		require.Equal(t, int64(10), productID)
		return nil, ErrSubscriptionNotFound
	}
	svc := NewSubscriptionProductService(repo)

	got, err := svc.GetActiveProductSubscription(context.Background(), 40, 20)

	require.NoError(t, err)
	require.Equal(t, int64(11), got.Binding.ProductID)
	require.Equal(t, "gpt_daily_150", got.Binding.ProductCode)
	require.Equal(t, int64(31), got.Subscription.ID)
	require.Equal(t, 1.5, got.Binding.DebitMultiplier)
	require.Equal(t, 1, repo.bindingByUserGroupCalls)
	require.Equal(t, 0, repo.bindingCalls)
	require.Equal(t, 0, repo.subscriptionCalls)
}

func TestSubscriptionProductService_ListVisibleGroups(t *testing.T) {
	repo := &subscriptionProductRepoStub{
		groups: []Group{
			{ID: 10, Name: "plus", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription},
			{ID: 11, Name: "pro", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription},
		},
	}
	svc := NewSubscriptionProductService(repo)

	groups, err := svc.ListVisibleGroups(context.Background(), 40)

	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, "plus", groups[0].Name)
	require.Equal(t, "pro", groups[1].Name)
}

type subscriptionProductRepoStub struct {
	binding                 *SubscriptionProductBinding
	subscription            *UserProductSubscription
	bindingByUserGroup      *SubscriptionProductBinding
	subscriptionByUserGroup *UserProductSubscription
	groups                  []Group
	getSubscription         func(ctx context.Context, userID, productID int64) (*UserProductSubscription, error)

	bindingCalls            int
	subscriptionCalls       int
	bindingByUserGroupCalls int
	visibleCalls            int
}

func (s *subscriptionProductRepoStub) GetActiveProductBindingByGroupID(context.Context, int64) (*SubscriptionProductBinding, error) {
	s.bindingCalls++
	return s.binding, nil
}

func (s *subscriptionProductRepoStub) GetActiveUserProductSubscription(ctx context.Context, userID, productID int64) (*UserProductSubscription, error) {
	s.subscriptionCalls++
	if s.getSubscription != nil {
		return s.getSubscription(ctx, userID, productID)
	}
	return s.subscription, nil
}

func (s *subscriptionProductRepoStub) GetActiveProductSubscriptionByUserAndGroupID(context.Context, int64, int64) (*SubscriptionProductBinding, *UserProductSubscription, error) {
	s.bindingByUserGroupCalls++
	return s.bindingByUserGroup, s.subscriptionByUserGroup, nil
}

func (s *subscriptionProductRepoStub) ListVisibleGroupsByUserID(context.Context, int64) ([]Group, error) {
	s.visibleCalls++
	return s.groups, nil
}

func (s *subscriptionProductRepoStub) ListActiveProductsByUserID(context.Context, int64) ([]ActiveSubscriptionProduct, error) {
	return nil, nil
}
