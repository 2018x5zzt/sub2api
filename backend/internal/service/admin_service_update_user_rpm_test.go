//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// rpmUserRepoStub 复用 admin_service_update_balance_test.go 的基础 stub 结构，
// 只在 Update 时把入参克隆一份，便于断言修改后的 RPMLimit。
type rpmUserRepoStub struct {
	*userRepoStub
	lastUpdated *User
}

func (s *rpmUserRepoStub) Update(_ context.Context, user *User) error {
	if user == nil {
		return nil
	}
	clone := *user
	s.lastUpdated = &clone
	if s.userRepoStub != nil {
		s.userRepoStub.user = &clone
	}
	return nil
}

func TestAdminService_UpdateUser_InvalidatesAuthCacheOnRPMLimitChange(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", RPMLimit: 10}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: invalidator,
	}

	newRPM := 60
	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{
		RPMLimit: &newRPM,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, 60, updated.RPMLimit)
	require.Equal(t, []int64{42}, invalidator.userIDs, "仅修改 RPMLimit 也应失效 API Key 认证缓存")
}

func TestAdminService_UpdateUser_NoInvalidateWhenRPMLimitUnchanged(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", RPMLimit: 10, Username: "old"}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: invalidator,
	}

	newName := "new"
	sameRPM := 10
	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{
		Username: &newName,
		RPMLimit: &sameRPM,
	})
	require.NoError(t, err)
	require.Empty(t, invalidator.userIDs, "只改 username 不应触发认证缓存失效")
}

func TestAdminService_UpdateUser_ValidatesFallbackGroupAuthorization(t *testing.T) {
	groupID := int64(88)
	enabled := true
	limit := 5.0
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com"}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	svc := &adminServiceImpl{
		userRepo:       repo,
		groupRepo:      &groupRepoStubForAdmin{getByID: &Group{ID: groupID, Name: "exclusive", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, IsExclusive: true}},
		redeemCodeRepo: &redeemRepoStub{},
	}

	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{
		SubscriptionBalanceFallbackEnabled:  &enabled,
		SubscriptionBalanceFallbackLimitUSD: &limit,
		SubscriptionBalanceFallbackGroupID:  &groupID,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authorized")
	require.Nil(t, repo.lastUpdated)
}

func TestAdminService_UpdateUser_UpdatesFallbackSettings(t *testing.T) {
	groupID := int64(89)
	enabled := true
	limit := 5.0
	used := 1.25
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", AllowedGroups: []int64{groupID}}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		groupRepo:            &groupRepoStubForAdmin{getByID: &Group{ID: groupID, Name: "exclusive", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, IsExclusive: true}},
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: invalidator,
	}

	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{
		SubscriptionBalanceFallbackEnabled:  &enabled,
		SubscriptionBalanceFallbackLimitUSD: &limit,
		SubscriptionBalanceFallbackUsedUSD:  &used,
		SubscriptionBalanceFallbackGroupID:  &groupID,
	})
	require.NoError(t, err)
	require.True(t, updated.SubscriptionBalanceFallbackEnabled)
	require.Equal(t, 5.0, updated.SubscriptionBalanceFallbackLimitUSD)
	require.Equal(t, 1.25, updated.SubscriptionBalanceFallbackUsedUSD)
	require.NotNil(t, updated.SubscriptionBalanceFallbackGroupID)
	require.Equal(t, groupID, *updated.SubscriptionBalanceFallbackGroupID)
	require.Equal(t, []int64{42}, invalidator.userIDs)
}
