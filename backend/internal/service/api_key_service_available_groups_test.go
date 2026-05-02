package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type apiKeyAvailableGroupsUserRepo struct {
	user *User
}

func (r *apiKeyAvailableGroupsUserRepo) Create(context.Context, *User) error { return nil }
func (r *apiKeyAvailableGroupsUserRepo) GetByID(context.Context, int64) (*User, error) {
	if r.user == nil {
		return &User{}, nil
	}
	out := *r.user
	return &out, nil
}
func (r *apiKeyAvailableGroupsUserRepo) GetByEmail(context.Context, string) (*User, error) {
	return &User{}, nil
}
func (r *apiKeyAvailableGroupsUserRepo) GetByInviteCode(context.Context, string) (*User, error) {
	return nil, ErrUserNotFound
}
func (r *apiKeyAvailableGroupsUserRepo) ExistsByInviteCode(context.Context, string) (bool, error) {
	return false, nil
}
func (r *apiKeyAvailableGroupsUserRepo) CountInviteesByInviter(context.Context, int64) (int64, error) {
	return 0, nil
}
func (r *apiKeyAvailableGroupsUserRepo) GetFirstAdmin(context.Context) (*User, error) {
	return &User{}, nil
}
func (r *apiKeyAvailableGroupsUserRepo) Update(context.Context, *User) error { return nil }
func (r *apiKeyAvailableGroupsUserRepo) UpdateInviterBinding(context.Context, int64, *int64) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) Delete(context.Context, int64) error { return nil }
func (r *apiKeyAvailableGroupsUserRepo) GetUserAvatar(context.Context, int64) (*UserAvatar, error) {
	return nil, nil
}
func (r *apiKeyAvailableGroupsUserRepo) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	return &UserAvatar{}, nil
}
func (r *apiKeyAvailableGroupsUserRepo) DeleteUserAvatar(context.Context, int64) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *apiKeyAvailableGroupsUserRepo) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *apiKeyAvailableGroupsUserRepo) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	return nil, nil
}
func (r *apiKeyAvailableGroupsUserRepo) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	return nil, nil
}
func (r *apiKeyAvailableGroupsUserRepo) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) UpdateBalance(context.Context, int64, float64) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) DeductBalance(context.Context, int64, float64) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) UpdateConcurrency(context.Context, int64, int) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) ExistsByEmail(context.Context, string) (bool, error) {
	return false, nil
}
func (r *apiKeyAvailableGroupsUserRepo) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	return 0, nil
}
func (r *apiKeyAvailableGroupsUserRepo) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	return nil, nil
}
func (r *apiKeyAvailableGroupsUserRepo) UnbindUserAuthProvider(context.Context, int64, string) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) UpdateTotpSecret(context.Context, int64, *string) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) EnableTotp(context.Context, int64) error {
	return nil
}
func (r *apiKeyAvailableGroupsUserRepo) DisableTotp(context.Context, int64) error {
	return nil
}

type apiKeyAvailableGroupsRepo struct {
	groupRepoNoop
	groups []Group
}

func (r *apiKeyAvailableGroupsRepo) ListActive(context.Context) ([]Group, error) {
	out := make([]Group, len(r.groups))
	copy(out, r.groups)
	return out, nil
}

type apiKeyAvailableGroupsSubscriptionRepo struct {
	userSubRepoNoop
	subs []UserSubscription
}

func (r *apiKeyAvailableGroupsSubscriptionRepo) ListActiveByUserID(context.Context, int64) ([]UserSubscription, error) {
	out := make([]UserSubscription, len(r.subs))
	copy(out, r.subs)
	return out, nil
}

type apiKeyProductVisibleGroupsStub struct {
	groups []Group
}

func (s *apiKeyProductVisibleGroupsStub) ListVisibleGroups(ctx context.Context, userID int64) ([]Group, error) {
	out := make([]Group, len(s.groups))
	copy(out, s.groups)
	return out, nil
}

type apiKeyProductBindRepo struct {
	apiKey  *APIKey
	created []*APIKey
	updated []*APIKey
}

func (r *apiKeyProductBindRepo) Create(ctx context.Context, key *APIKey) error {
	cp := *key
	cp.ID = int64(len(r.created) + 1)
	key.ID = cp.ID
	r.created = append(r.created, &cp)
	return nil
}

func (r *apiKeyProductBindRepo) GetByID(ctx context.Context, id int64) (*APIKey, error) {
	if r.apiKey != nil {
		cp := *r.apiKey
		return &cp, nil
	}
	return nil, ErrAPIKeyNotFound
}

func (r *apiKeyProductBindRepo) Update(ctx context.Context, key *APIKey) error {
	cp := *key
	r.updated = append(r.updated, &cp)
	return nil
}

func (r *apiKeyProductBindRepo) ExistsByKey(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (r *apiKeyProductBindRepo) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}
func (r *apiKeyProductBindRepo) GetByKey(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKey call")
}
func (r *apiKeyProductBindRepo) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}
func (r *apiKeyProductBindRepo) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}
func (r *apiKeyProductBindRepo) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}
func (r *apiKeyProductBindRepo) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}
func (r *apiKeyProductBindRepo) CountByUserID(context.Context, int64) (int64, error) {
	panic("unexpected CountByUserID call")
}
func (r *apiKeyProductBindRepo) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}
func (r *apiKeyProductBindRepo) SearchAPIKeys(context.Context, int64, string, int) ([]APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}
func (r *apiKeyProductBindRepo) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}
func (r *apiKeyProductBindRepo) UpdateGroupIDByUserAndGroup(context.Context, int64, int64, int64) (int64, error) {
	panic("unexpected UpdateGroupIDByUserAndGroup call")
}
func (r *apiKeyProductBindRepo) CountByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}
func (r *apiKeyProductBindRepo) ListKeysByUserID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByUserID call")
}
func (r *apiKeyProductBindRepo) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByGroupID call")
}
func (r *apiKeyProductBindRepo) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}
func (r *apiKeyProductBindRepo) UpdateLastUsed(context.Context, int64, time.Time) error {
	panic("unexpected UpdateLastUsed call")
}
func (r *apiKeyProductBindRepo) IncrementRateLimitUsage(context.Context, int64, float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}
func (r *apiKeyProductBindRepo) ResetRateLimitWindows(context.Context, int64) error {
	panic("unexpected ResetRateLimitWindows call")
}
func (r *apiKeyProductBindRepo) GetRateLimitData(context.Context, int64) (*APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}

type apiKeyProductBindGroupRepo struct {
	groupRepoNoop

	group *Group
}

func (r *apiKeyProductBindGroupRepo) GetByID(ctx context.Context, id int64) (*Group, error) {
	if r.group == nil {
		return nil, ErrGroupNotFound
	}
	cp := *r.group
	return &cp, nil
}

func TestAPIKeyServiceGetAvailableGroupsIncludesProductSubscriptionGroups(t *testing.T) {
	legacyGroupID := int64(20)
	productGroupID := int64(30)

	svc := NewAPIKeyService(
		nil,
		&apiKeyAvailableGroupsUserRepo{user: &User{ID: 7, AllowedGroups: []int64{legacyGroupID}}},
		&apiKeyAvailableGroupsRepo{groups: []Group{
			{ID: 1, Name: "public", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
			{ID: legacyGroupID, Name: "legacy-sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
			{ID: productGroupID, Name: "product-sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
		}},
		&apiKeyAvailableGroupsSubscriptionRepo{subs: []UserSubscription{{GroupID: legacyGroupID}}},
		nil,
		nil,
		nil,
	)
	svc.productVisibleGroups = &apiKeyProductVisibleGroupsStub{
		groups: []Group{{ID: productGroupID, Name: "product-sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}},
	}

	groups, err := svc.GetAvailableGroups(context.Background(), 7)
	require.NoError(t, err)

	require.Equal(t, []int64{1, legacyGroupID, productGroupID}, groupIDs(groups))
}

func TestAPIKeyServiceCreateAllowsProductSubscriptionGroup(t *testing.T) {
	productGroupID := int64(30)
	repo := &apiKeyProductBindRepo{}
	svc := NewAPIKeyService(
		repo,
		&apiKeyAvailableGroupsUserRepo{user: &User{ID: 7}},
		&apiKeyProductBindGroupRepo{group: &Group{ID: productGroupID, Name: "product-sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}},
		&apiKeyAvailableGroupsSubscriptionRepo{},
		nil,
		nil,
		&config.Config{},
	)
	svc.productVisibleGroups = &apiKeyProductVisibleGroupsStub{
		groups: []Group{{ID: productGroupID, Name: "product-sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}},
	}

	key, err := svc.Create(context.Background(), 7, CreateAPIKeyRequest{
		Name:      "product key",
		GroupID:   &productGroupID,
		CustomKey: stringPtrForProductBind("custom-product-key"),
	})
	require.NoError(t, err)
	require.NotNil(t, key.GroupID)
	require.Equal(t, productGroupID, *key.GroupID)
	require.Len(t, repo.created, 1)
	require.NotNil(t, repo.created[0].GroupID)
	require.Equal(t, productGroupID, *repo.created[0].GroupID)
}

func TestAPIKeyServiceUpdateAllowsProductSubscriptionGroup(t *testing.T) {
	productGroupID := int64(30)
	repo := &apiKeyProductBindRepo{
		apiKey: &APIKey{ID: 99, UserID: 7, Key: "existing-key", Name: "old", Status: StatusActive},
	}
	svc := NewAPIKeyService(
		repo,
		&apiKeyAvailableGroupsUserRepo{user: &User{ID: 7}},
		&apiKeyProductBindGroupRepo{group: &Group{ID: productGroupID, Name: "product-sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}},
		&apiKeyAvailableGroupsSubscriptionRepo{},
		nil,
		nil,
		&config.Config{},
	)
	svc.productVisibleGroups = &apiKeyProductVisibleGroupsStub{
		groups: []Group{{ID: productGroupID, Name: "product-sub", Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}},
	}

	key, err := svc.Update(context.Background(), 99, 7, UpdateAPIKeyRequest{GroupID: &productGroupID})
	require.NoError(t, err)
	require.NotNil(t, key.GroupID)
	require.Equal(t, productGroupID, *key.GroupID)
	require.Len(t, repo.updated, 1)
	require.NotNil(t, repo.updated[0].GroupID)
	require.Equal(t, productGroupID, *repo.updated[0].GroupID)
}

func groupIDs(groups []Group) []int64 {
	out := make([]int64, 0, len(groups))
	for _, group := range groups {
		out = append(out, group.ID)
	}
	return out
}

func stringPtrForProductBind(v string) *string {
	return &v
}
