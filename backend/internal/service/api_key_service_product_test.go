//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type productAPIKeyUserRepoStub struct {
	userRepoNoop
	user *User
}

func (s productAPIKeyUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	cp := *s.user
	return &cp, nil
}

type productAPIKeyGroupRepoStub struct {
	groupRepoNoop
	group *Group
}

func (s productAPIKeyGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	cp := *s.group
	return &cp, nil
}

type productAPIKeyRepoStub struct {
	apiKeyRepoNoop
	created *APIKey
}

func (s *productAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	cp := *key
	s.created = &cp
	return nil
}

func (s *productAPIKeyRepoStub) ExistsByKey(context.Context, string) (bool, error) {
	return false, nil
}

func TestAPIKeyService_Create_ProductSettledGroupRequiresProductSubscription(t *testing.T) {
	group := &Group{ID: 88, Name: "Product Pro", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}
	user := &User{ID: 42, Status: StatusActive}
	apiKeyRepo := &productAPIKeyRepoStub{}
	productRepo := &subscriptionProductRepoStub{
		binding: &SubscriptionProductBinding{
			ProductID:       101,
			ProductCode:     "shared",
			ProductName:     "Shared",
			ProductStatus:   SubscriptionProductStatusActive,
			BindingStatus:   SubscriptionProductBindingStatusActive,
			GroupID:         group.ID,
			DebitMultiplier: 1,
		},
		subscription: &UserProductSubscription{
			ID:        202,
			UserID:    user.ID,
			ProductID: 101,
			Status:    SubscriptionStatusActive,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	svc := &APIKeyService{
		apiKeyRepo:                 apiKeyRepo,
		userRepo:                   productAPIKeyUserRepoStub{user: user},
		groupRepo:                  productAPIKeyGroupRepoStub{group: group},
		userSubRepo:                &userSubRepoNoop{},
		subscriptionProductService: NewSubscriptionProductService(productRepo),
		cfg:                        &config.Config{},
	}

	created, err := svc.Create(context.Background(), user.ID, CreateAPIKeyRequest{Name: "product", GroupID: &group.ID})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, apiKeyRepo.created)
	require.NotNil(t, apiKeyRepo.created.GroupID)
	require.Equal(t, group.ID, *apiKeyRepo.created.GroupID)
	require.Equal(t, 1, productRepo.bindingCalls)
	require.Equal(t, 1, productRepo.subscriptionCalls)
}
