//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type failOnSecondInviteBalanceUserRepo struct {
	base  service.InviteUserRepository
	calls int
	err   error
}

func (r *failOnSecondInviteBalanceUserRepo) GetByID(ctx context.Context, id int64) (*service.User, error) {
	return r.base.GetByID(ctx, id)
}

func (r *failOnSecondInviteBalanceUserRepo) GetByInviteCode(ctx context.Context, code string) (*service.User, error) {
	return r.base.GetByInviteCode(ctx, code)
}

func (r *failOnSecondInviteBalanceUserRepo) ExistsByInviteCode(ctx context.Context, code string) (bool, error) {
	return r.base.ExistsByInviteCode(ctx, code)
}

func (r *failOnSecondInviteBalanceUserRepo) CountInviteesByInviter(ctx context.Context, inviterID int64) (int64, error) {
	return r.base.CountInviteesByInviter(ctx, inviterID)
}

func (r *failOnSecondInviteBalanceUserRepo) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	r.calls++
	if r.calls == 2 {
		return r.err
	}
	return r.base.UpdateBalance(ctx, id, amount)
}

func TestInviteService_ApplyBaseRechargeRewards_RollsBackOnBalanceFailure(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	baseUserRepo := NewUserRepository(client, nil).(*userRepository)
	redeemRepo := NewRedeemCodeRepository(client).(*redeemCodeRepository)
	rewardRepo := NewInviteRewardRecordRepository(client)

	inviter := createInviteAdminServiceUser(t, baseUserRepo, "invite-growth-atomic-inviter@example.com", nil, nil)
	invitee := createInviteAdminServiceUser(t, baseUserRepo, "invite-growth-atomic-invitee@example.com", int64Ptr(inviter.ID), nil)
	cleanupInviteAdminServiceUsers(t, inviter.ID, invitee.ID)

	redeemCode := &service.RedeemCode{
		Code:       "INVITE-GROWTH-ATOMIC-909001",
		Type:       service.RedeemTypeBalance,
		Value:      100,
		Status:     service.StatusUnused,
		SourceType: service.RedeemSourceCommercial,
	}
	require.NoError(t, redeemRepo.Create(ctx, redeemCode))

	svc := service.NewInviteService(
		&failOnSecondInviteBalanceUserRepo{
			base: baseUserRepo,
			err:  errors.New("forced failure on second balance update"),
		},
		rewardRepo,
		client,
	)

	err := svc.ApplyBaseRechargeRewards(ctx, invitee.ID, &service.RedeemCode{
		ID:         redeemCode.ID,
		Type:       service.RedeemTypeBalance,
		SourceType: service.RedeemSourceCommercial,
		Value:      100,
	})
	require.ErrorContains(t, err, "forced failure on second balance update")

	reloadedInviter, getInviterErr := baseUserRepo.GetByID(ctx, inviter.ID)
	require.NoError(t, getInviterErr)
	require.Equal(t, 0.0, reloadedInviter.Balance)

	reloadedInvitee, getInviteeErr := baseUserRepo.GetByID(ctx, invitee.ID)
	require.NoError(t, getInviteeErr)
	require.Equal(t, 0.0, reloadedInvitee.Balance)

	require.Zero(t, countInviteRewardRecordsByTarget(t, inviter.ID))
	require.Zero(t, countInviteRewardRecordsByTarget(t, invitee.ID))
}

func TestInviteService_ApplyBaseRechargeRewards_DuplicateSettlementNoOp(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	userRepo := NewUserRepository(client, nil).(*userRepository)
	redeemRepo := NewRedeemCodeRepository(client).(*redeemCodeRepository)
	rewardRepo := NewInviteRewardRecordRepository(client)

	inviter := createInviteAdminServiceUser(t, userRepo, "invite-growth-duplicate-inviter@example.com", nil, nil)
	invitee := createInviteAdminServiceUser(t, userRepo, "invite-growth-duplicate-invitee@example.com", int64Ptr(inviter.ID), nil)
	cleanupInviteAdminServiceUsers(t, inviter.ID, invitee.ID)

	redeemCode := &service.RedeemCode{
		Code:       fmt.Sprintf("INVITE-GROWTH-DUPLICATE-%d", invitee.ID),
		Type:       service.RedeemTypeBalance,
		Value:      100,
		Status:     service.StatusUnused,
		SourceType: service.RedeemSourceCommercial,
	}
	require.NoError(t, redeemRepo.Create(ctx, redeemCode))

	svc := service.NewInviteService(userRepo, rewardRepo, client)
	inputCode := &service.RedeemCode{
		ID:         redeemCode.ID,
		Type:       service.RedeemTypeBalance,
		SourceType: service.RedeemSourceCommercial,
		Value:      100,
	}

	require.NoError(t, svc.ApplyBaseRechargeRewards(ctx, invitee.ID, inputCode))
	require.NoError(t, svc.ApplyBaseRechargeRewards(ctx, invitee.ID, inputCode))

	reloadedInviter, err := userRepo.GetByID(ctx, inviter.ID)
	require.NoError(t, err)
	require.Equal(t, 3.0, reloadedInviter.Balance)

	reloadedInvitee, err := userRepo.GetByID(ctx, invitee.ID)
	require.NoError(t, err)
	require.Equal(t, 3.0, reloadedInvitee.Balance)

	require.Equal(t, 1, countInviteRewardRecordsByTarget(t, inviter.ID))
	require.Equal(t, 1, countInviteRewardRecordsByTarget(t, invitee.ID))
}

func countInviteRewardRecordsByTarget(t *testing.T, rewardTargetUserID int64) int {
	t.Helper()

	var total int
	err := integrationDB.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM invite_reward_records WHERE reward_target_user_id = $1`,
		rewardTargetUserID,
	).Scan(&total)
	require.NoError(t, err)
	return total
}
