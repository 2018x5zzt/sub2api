//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type userRepoNoop struct{}

func (userRepoNoop) Create(context.Context, *User) error { panic("unexpected Create call") }
func (userRepoNoop) GetByID(context.Context, int64) (*User, error) {
	panic("unexpected GetByID call")
}
func (userRepoNoop) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}
func (userRepoNoop) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}
func (userRepoNoop) Update(context.Context, *User) error { panic("unexpected Update call") }
func (userRepoNoop) Delete(context.Context, int64) error { panic("unexpected Delete call") }
func (userRepoNoop) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (userRepoNoop) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (userRepoNoop) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected UpdateBalance call")
}
func (userRepoNoop) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}
func (userRepoNoop) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}
func (userRepoNoop) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}
func (userRepoNoop) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}
func (userRepoNoop) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}
func (userRepoNoop) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}
func (userRepoNoop) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}
func (userRepoNoop) EnableTotp(context.Context, int64) error { panic("unexpected EnableTotp call") }
func (userRepoNoop) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

type apiKeyRepoNoop struct{}

func (apiKeyRepoNoop) Create(context.Context, *APIKey) error { panic("unexpected Create call") }
func (apiKeyRepoNoop) GetByID(context.Context, int64) (*APIKey, error) {
	panic("unexpected GetByID call")
}
func (apiKeyRepoNoop) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}
func (apiKeyRepoNoop) GetByKey(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKey call")
}
func (apiKeyRepoNoop) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}
func (apiKeyRepoNoop) Update(context.Context, *APIKey) error { panic("unexpected Update call") }
func (apiKeyRepoNoop) Delete(context.Context, int64) error   { panic("unexpected Delete call") }
func (apiKeyRepoNoop) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}
func (apiKeyRepoNoop) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}
func (apiKeyRepoNoop) CountByUserID(context.Context, int64) (int64, error) {
	panic("unexpected CountByUserID call")
}
func (apiKeyRepoNoop) ExistsByKey(context.Context, string) (bool, error) {
	panic("unexpected ExistsByKey call")
}
func (apiKeyRepoNoop) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}
func (apiKeyRepoNoop) SearchAPIKeys(context.Context, int64, string, int) ([]APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}
func (apiKeyRepoNoop) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}
func (apiKeyRepoNoop) UpdateGroupIDByUserAndGroup(context.Context, int64, int64, int64) (int64, error) {
	panic("unexpected UpdateGroupIDByUserAndGroup call")
}
func (apiKeyRepoNoop) CountByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}
func (apiKeyRepoNoop) ListKeysByUserID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByUserID call")
}
func (apiKeyRepoNoop) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByGroupID call")
}
func (apiKeyRepoNoop) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}
func (apiKeyRepoNoop) UpdateLastUsed(context.Context, int64, time.Time) error {
	panic("unexpected UpdateLastUsed call")
}
func (apiKeyRepoNoop) IncrementRateLimitUsage(context.Context, int64, float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}
func (apiKeyRepoNoop) ResetRateLimitWindows(context.Context, int64) error {
	panic("unexpected ResetRateLimitWindows call")
}
func (apiKeyRepoNoop) GetRateLimitData(context.Context, int64) (*APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}

type apiKeyRepoStubForBudgetUpdate struct {
	apiKeyRepoNoop
	key *APIKey
}

func (s *apiKeyRepoStubForBudgetUpdate) GetByID(_ context.Context, _ int64) (*APIKey, error) {
	clone := *s.key
	return &clone, nil
}

func TestAPIKeyService_Update_RejectsGroupIDChange(t *testing.T) {
	svc := &APIKeyService{
		apiKeyRepo: &apiKeyRepoStubForBudgetUpdate{
			key: &APIKey{
				ID:      1,
				UserID:  7,
				Key:     "sk-group-immutable",
				Name:    "immutable",
				GroupID: int64Ptr(11),
				Status:  StatusActive,
			},
		},
		userRepo: userRepoNoop{},
	}

	_, err := svc.Update(context.Background(), 1, 7, UpdateAPIKeyRequest{
		GroupID: int64Ptr(22),
	})
	require.ErrorIs(t, err, ErrAPIKeyGroupImmutable)
}
