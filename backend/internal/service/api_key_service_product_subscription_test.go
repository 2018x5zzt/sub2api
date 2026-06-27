//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type apiKeyProductUserRepoStub struct {
	userRepoStubForGroupUpdate
	user *User
}

func (s *apiKeyProductUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	if s.user == nil || s.user.ID != id {
		return nil, ErrUserNotFound
	}
	clone := *s.user
	return &clone, nil
}

type apiKeyProductGroupRepoStub struct {
	groupRepoStubForGroupUpdate
	groups map[int64]Group
	active []Group
}

func (s *apiKeyProductGroupRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	group, ok := s.groups[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	clone := group
	return &clone, nil
}

func (s *apiKeyProductGroupRepoStub) ListActive(context.Context) ([]Group, error) {
	return append([]Group(nil), s.active...), nil
}

type apiKeyProductAPIKeyRepoStub struct {
	apiKeyRepoStubForGroupUpdate
	created *APIKey
}

func (s *apiKeyProductAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	clone := *key
	clone.ID = 101
	s.created = &clone
	key.ID = clone.ID
	return nil
}

type apiKeyProductLegacySubRepoStub struct {
	userSubRepoNoop
	getActiveCalled  bool
	listActiveCalled bool
}

func (s *apiKeyProductLegacySubRepoStub) GetActiveByUserIDAndGroupID(context.Context, int64, int64) (*UserSubscription, error) {
	s.getActiveCalled = true
	return nil, ErrSubscriptionNotFound
}

func (s *apiKeyProductLegacySubRepoStub) ListActiveByUserID(context.Context, int64) ([]UserSubscription, error) {
	s.listActiveCalled = true
	return []UserSubscription{}, nil
}

func TestAPIKeyServiceGetAvailableGroupsIncludesProductSubscriptionGroups(t *testing.T) {
	user := &User{ID: 42, Status: StatusActive}
	standardGroup := Group{ID: 1, Name: "Standard", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard}
	subscriptionGroup := Group{ID: 7, Name: "Product Sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}
	legacySubs := &apiKeyProductLegacySubRepoStub{}
	productSvc := NewSubscriptionProductService(&productAwareProductRepoStub{
		listVisibleGroupsByUserID: func(_ context.Context, userID int64) ([]Group, error) {
			require.Equal(t, user.ID, userID)
			return []Group{subscriptionGroup}, nil
		},
	})
	svc := NewAPIKeyService(
		nil,
		&apiKeyProductUserRepoStub{user: user},
		&apiKeyProductGroupRepoStub{
			groups: map[int64]Group{standardGroup.ID: standardGroup, subscriptionGroup.ID: subscriptionGroup},
			active: []Group{standardGroup, subscriptionGroup},
		},
		legacySubs,
		nil,
		nil,
		&config.Config{},
	)
	svc.SetSubscriptionProductService(productSvc)

	groups, err := svc.GetAvailableGroups(context.Background(), user.ID)

	require.NoError(t, err)
	require.Equal(t, []int64{standardGroup.ID, subscriptionGroup.ID}, groupIDs(groups))
	require.False(t, legacySubs.listActiveCalled, "legacy user_subscriptions must not drive key group visibility")
}

func TestAPIKeyServiceCreateAllowsProductSubscriptionGroup(t *testing.T) {
	user := &User{ID: 42, Status: StatusActive}
	groupID := int64(7)
	subscriptionGroup := Group{ID: groupID, Name: "Product Sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}
	legacySubs := &apiKeyProductLegacySubRepoStub{}
	apiKeyRepo := &apiKeyProductAPIKeyRepoStub{}
	productSvc := NewSubscriptionProductService(&productAwareProductRepoStub{
		getActiveProductSubscriptionByUserAndGroupID: func(_ context.Context, userID, requestedGroupID int64, productFamily *string) (*SubscriptionProductBinding, *UserProductSubscription, error) {
			require.Equal(t, user.ID, userID)
			require.Equal(t, groupID, requestedGroupID)
			require.Nil(t, productFamily)
			return &SubscriptionProductBinding{
					ProductID:     99,
					ProductStatus: SubscriptionProductStatusActive,
					BindingStatus: SubscriptionProductBindingStatusActive,
				}, &UserProductSubscription{
					ID:        1001,
					UserID:    user.ID,
					ProductID: 99,
					Status:    SubscriptionStatusActive,
					StartsAt:  time.Now().Add(-time.Hour),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
		},
	})
	svc := NewAPIKeyService(
		apiKeyRepo,
		&apiKeyProductUserRepoStub{user: user},
		&apiKeyProductGroupRepoStub{
			groups: map[int64]Group{groupID: subscriptionGroup},
			active: []Group{subscriptionGroup},
		},
		legacySubs,
		nil,
		nil,
		&config.Config{},
	)
	svc.SetSubscriptionProductService(productSvc)

	key, err := svc.Create(context.Background(), user.ID, CreateAPIKeyRequest{
		Name:    "product subscription key",
		GroupID: &groupID,
	})

	require.NoError(t, err)
	require.NotNil(t, key.GroupID)
	require.Equal(t, groupID, *key.GroupID)
	require.NotNil(t, apiKeyRepo.created)
	require.NotNil(t, apiKeyRepo.created.GroupID)
	require.Equal(t, groupID, *apiKeyRepo.created.GroupID)
	require.False(t, legacySubs.getActiveCalled, "legacy user_subscriptions must not authorize product subscription groups")
}

func groupIDs(groups []Group) []int64 {
	ids := make([]int64, 0, len(groups))
	for _, group := range groups {
		ids = append(ids, group.ID)
	}
	return ids
}
