//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type inviteUserRepoStub struct {
	usersByID         map[int64]*User
	usersByInviteCode map[string]*User
	existingCodes     map[string]bool
}

type inviteRewardRepoStub struct {
	created          []InviteRewardRecord
	totalBase        float64
	createErr        error
	createBatchCalls int
}

func (s *inviteUserRepoStub) Create(context.Context, *User) error {
	panic("unexpected Create call")
}

func (s *inviteUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	u, ok := s.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *inviteUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}

func (s *inviteUserRepoStub) GetByInviteCode(_ context.Context, code string) (*User, error) {
	u, ok := s.usersByInviteCode[code]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *inviteUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (s *inviteUserRepoStub) Update(context.Context, *User) error {
	panic("unexpected Update call")
}

func (s *inviteUserRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *inviteUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *inviteUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *inviteUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected UpdateBalance call")
}

func (s *inviteUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}

func (s *inviteUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}

func (s *inviteUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}

func (s *inviteUserRepoStub) ExistsByInviteCode(_ context.Context, code string) (bool, error) {
	return s.existingCodes[code], nil
}

func (s *inviteUserRepoStub) CountInviteesByInviter(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *inviteUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}

func (s *inviteUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}

func (s *inviteUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}

func (s *inviteUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}

func (s *inviteUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp call")
}

func (s *inviteUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

func (s *inviteRewardRepoStub) CreateBatch(_ context.Context, records []InviteRewardRecord) error {
	s.createBatchCalls++
	if s.createErr != nil {
		return s.createErr
	}
	s.created = append(s.created, records...)
	return nil
}

func (s *inviteRewardRepoStub) ListByRewardTarget(_ context.Context, _ int64, _ pagination.PaginationParams) ([]InviteRewardRecord, *pagination.PaginationResult, error) {
	return s.created, &pagination.PaginationResult{Total: int64(len(s.created))}, nil
}

func (s *inviteRewardRepoStub) ListByAdminActionID(_ context.Context, _ int64) ([]InviteRewardRecord, error) {
	return s.created, nil
}

func (s *inviteRewardRepoStub) SumBaseRewardsByTargetAndRole(_ context.Context, _ int64, _ string) (float64, error) {
	return s.totalBase, nil
}

func TestInviteService_GenerateUniqueInviteCodeRetriesOnCollision(t *testing.T) {
	first := true
	svc := &InviteService{
		userRepo: &inviteUserRepoStub{
			existingCodes: map[string]bool{
				"ABCDEFGH": true,
				"ZXCVBN12": false,
			},
		},
		codeGenerator: func() (string, error) {
			if first {
				first = false
				return "ABCDEFGH", nil
			}
			return "ZXCVBN12", nil
		},
	}

	code, err := svc.GenerateUniqueInviteCode(context.Background())
	require.NoError(t, err)
	require.Equal(t, "ZXCVBN12", code)
}

func TestInviteService_ResolveInviterByCodeReturnsUser(t *testing.T) {
	inviter := &User{ID: 7, InviteCode: "INVITER07", Status: StatusActive}
	svc := &InviteService{
		userRepo: &inviteUserRepoStub{
			usersByInviteCode: map[string]*User{"INVITER07": inviter},
		},
	}

	got, err := svc.ResolveInviterByCode(context.Background(), "inviter07")
	require.NoError(t, err)
	require.Equal(t, inviter.ID, got.ID)
}

func TestInviteService_ResolveInviterByCodeRejectsInactiveUser(t *testing.T) {
	svc := &InviteService{
		userRepo: &inviteUserRepoStub{
			usersByInviteCode: map[string]*User{
				"INVITER07": {ID: 7, InviteCode: "INVITER07", Status: StatusDisabled},
			},
		},
	}

	got, err := svc.ResolveInviterByCode(context.Background(), "INVITER07")
	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, got)
}

func TestInviteService_BuildSummaryUsesConfiguredFrontendURL(t *testing.T) {
	now := time.Now()
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyFrontendURL: "https://portal.example.com/app",
	}}, &config.Config{})
	svc := &InviteService{
		settingService: settingService,
		rewardRepo:     &inviteRewardRepoStub{totalBase: 12.5},
	}

	summary := svc.buildSummaryForUser(context.Background(), &User{ID: 3, InviteCode: "HELLO123", CreatedAt: now}, 4, 180.0, 12.5)
	require.Equal(t, "https://portal.example.com/app/register?invite=HELLO123", summary.InviteLink)
	require.Equal(t, int64(4), summary.InvitedUsersTotal)
	require.Equal(t, 180.0, summary.InviteesRechargeTotal)
}

type inviteSettlementUserRepoStub struct {
	users       map[int64]*User
	updateCalls int
}

func (s *inviteSettlementUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *inviteSettlementUserRepoStub) GetByInviteCode(_ context.Context, code string) (*User, error) {
	for _, user := range s.users {
		if user.InviteCode == code {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *inviteSettlementUserRepoStub) ExistsByInviteCode(_ context.Context, code string) (bool, error) {
	for _, user := range s.users {
		if user.InviteCode == code {
			return true, nil
		}
	}
	return false, nil
}

func (s *inviteSettlementUserRepoStub) CountInviteesByInviter(_ context.Context, inviterID int64) (int64, error) {
	var total int64
	for _, user := range s.users {
		if user.InvitedByUserID != nil && *user.InvitedByUserID == inviterID {
			total++
		}
	}
	return total, nil
}

func (s *inviteSettlementUserRepoStub) UpdateBalance(_ context.Context, id int64, amount float64) error {
	s.updateCalls++
	s.users[id].Balance += amount
	return nil
}

func TestInviteService_ApplyBaseRechargeRewardsCreditsInviterAndInvitee(t *testing.T) {
	inviterID := int64(7)
	userRepo := &inviteSettlementUserRepoStub{
		users: map[int64]*User{
			7: {ID: 7, Email: "inviter@test.com", Balance: 0, Status: StatusActive},
			8: {ID: 8, Email: "invitee@test.com", Balance: 10, Status: StatusActive, InvitedByUserID: &inviterID},
		},
	}
	rewardRepo := &inviteRewardRepoStub{}
	svc := &InviteService{userRepo: userRepo, rewardRepo: rewardRepo}

	err := svc.ApplyBaseRechargeRewards(context.Background(), 8, &RedeemCode{
		ID: 101, Type: RedeemTypeBalance, SourceType: RedeemSourceCommercial, Value: 100,
	})
	require.NoError(t, err)
	require.Equal(t, 3.0, userRepo.users[7].Balance)
	require.Equal(t, 13.0, userRepo.users[8].Balance)
	require.Len(t, rewardRepo.created, 2)
	require.NotNil(t, rewardRepo.created[0].RewardRate)
	require.NotNil(t, rewardRepo.created[1].RewardRate)
	require.Equal(t, 3.0, rewardRepo.created[0].RewardAmount)
	require.Equal(t, 3.0, rewardRepo.created[1].RewardAmount)
	require.Equal(t, 0.03, *rewardRepo.created[0].RewardRate)
	require.Equal(t, 0.03, *rewardRepo.created[1].RewardRate)
}

func TestInviteService_ApplyBaseRechargeRewardsSkipsNonCommercialRecharge(t *testing.T) {
	inviterID := int64(7)
	userRepo := &inviteSettlementUserRepoStub{
		users: map[int64]*User{
			7: {ID: 7, Balance: 0, Status: StatusActive},
			8: {ID: 8, Balance: 10, Status: StatusActive, InvitedByUserID: &inviterID},
		},
	}
	rewardRepo := &inviteRewardRepoStub{}
	svc := &InviteService{userRepo: userRepo, rewardRepo: rewardRepo}

	err := svc.ApplyBaseRechargeRewards(context.Background(), 8, &RedeemCode{
		ID: 101, Type: RedeemTypeBalance, SourceType: RedeemSourceBenefit, Value: 100,
	})
	require.NoError(t, err)
	require.Equal(t, 0.0, userRepo.users[7].Balance)
	require.Equal(t, 10.0, userRepo.users[8].Balance)
	require.Empty(t, rewardRepo.created)
}

func TestInviteService_ApplyBaseRechargeRewardsSkipsEmptySourceType(t *testing.T) {
	inviterID := int64(7)
	userRepo := &inviteSettlementUserRepoStub{
		users: map[int64]*User{
			7: {ID: 7, Balance: 0, Status: StatusActive},
			8: {ID: 8, Balance: 10, Status: StatusActive, InvitedByUserID: &inviterID},
		},
	}
	rewardRepo := &inviteRewardRepoStub{}
	svc := &InviteService{userRepo: userRepo, rewardRepo: rewardRepo}

	err := svc.ApplyBaseRechargeRewards(context.Background(), 8, &RedeemCode{
		ID: 101, Type: RedeemTypeBalance, SourceType: "", Value: 100,
	})
	require.NoError(t, err)
	require.Equal(t, 0.0, userRepo.users[7].Balance)
	require.Equal(t, 10.0, userRepo.users[8].Balance)
	require.Empty(t, rewardRepo.created)
}

func TestNewInviteService_WithEntClientSetsTransactionalClient(t *testing.T) {
	entClient := &dbent.Client{}
	svc := NewInviteService(&inviteSettlementUserRepoStub{}, &inviteRewardRepoStub{}, entClient)

	require.Equal(t, entClient, svc.entClient)
}

func TestInviteService_ApplyBaseRechargeRewardsSkipsSubscriptionRedeemCode(t *testing.T) {
	inviterID := int64(7)
	userRepo := &inviteSettlementUserRepoStub{
		users: map[int64]*User{
			7: {ID: 7, Balance: 0, Status: StatusActive},
			8: {ID: 8, Balance: 10, Status: StatusActive, InvitedByUserID: &inviterID},
		},
	}
	rewardRepo := &inviteRewardRepoStub{}
	svc := &InviteService{userRepo: userRepo, rewardRepo: rewardRepo}

	err := svc.ApplyBaseRechargeRewards(context.Background(), 8, &RedeemCode{
		ID: 101, Type: RedeemTypeSubscription, SourceType: RedeemSourceCommercial, Value: 100,
	})
	require.NoError(t, err)
	require.Equal(t, 0.0, userRepo.users[7].Balance)
	require.Equal(t, 10.0, userRepo.users[8].Balance)
	require.Empty(t, rewardRepo.created)
}

func TestInviteService_ApplyBaseRechargeRewardsSkipsDuplicateRewardRecord(t *testing.T) {
	inviterID := int64(7)
	userRepo := &inviteSettlementUserRepoStub{
		users: map[int64]*User{
			7: {ID: 7, Balance: 0, Status: StatusActive},
			8: {ID: 8, Balance: 10, Status: StatusActive, InvitedByUserID: &inviterID},
		},
	}
	rewardRepo := &inviteRewardRepoStub{
		createErr: nil,
	}
	svc := &InviteService{
		userRepo:   userRepo,
		rewardRepo: rewardRepo,
		baseRewardExistsFn: func(_ context.Context, redeemCodeID int64) (bool, error) {
			return redeemCodeID == 101, nil
		},
	}

	err := svc.ApplyBaseRechargeRewards(context.Background(), 8, &RedeemCode{
		ID: 101, Type: RedeemTypeBalance, SourceType: RedeemSourceCommercial, Value: 100,
	})
	require.NoError(t, err)
	require.Equal(t, 0, rewardRepo.createBatchCalls)
	require.Equal(t, 0, userRepo.updateCalls)
	require.Equal(t, 0.0, userRepo.users[7].Balance)
	require.Equal(t, 10.0, userRepo.users[8].Balance)
	require.Empty(t, rewardRepo.created)
}
