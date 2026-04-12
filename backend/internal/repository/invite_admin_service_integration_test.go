//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type failingBalanceUserRepo struct {
	*userRepository
	err error
}

func (r *failingBalanceUserRepo) UpdateBalance(context.Context, int64, float64) error {
	return r.err
}

type failingRelationshipEventAdminRepo struct {
	err error
}

func (r *failingRelationshipEventAdminRepo) Create(context.Context, *service.InviteRelationshipEvent) error {
	return r.err
}

func (r *failingRelationshipEventAdminRepo) ListByInvitee(context.Context, int64) ([]service.InviteRelationshipEvent, error) {
	return nil, nil
}

func (r *failingRelationshipEventAdminRepo) GetEffectiveInviterAt(context.Context, int64, time.Time) (*int64, error) {
	return nil, nil
}

type staticQualifyingRechargeRepo struct {
	events []service.InviteQualifyingRecharge
}

func (r *staticQualifyingRechargeRepo) ListInviteQualifyingRecharges(context.Context, service.InviteRecomputeScope) ([]service.InviteQualifyingRecharge, error) {
	return r.events, nil
}

type staticEffectiveInviterRepo struct {
	inviters map[int64]int64
}

func (r *staticEffectiveInviterRepo) Create(context.Context, *service.InviteRelationshipEvent) error {
	return nil
}

func (r *staticEffectiveInviterRepo) ListByInvitee(context.Context, int64) ([]service.InviteRelationshipEvent, error) {
	return nil, nil
}

func (r *staticEffectiveInviterRepo) GetEffectiveInviterAt(_ context.Context, inviteeUserID int64, _ time.Time) (*int64, error) {
	inviterID, ok := r.inviters[inviteeUserID]
	if !ok {
		return nil, nil
	}
	return &inviterID, nil
}

func TestAdminService_RebindInviter_RollsBackWhenEventWriteFails(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	userRepo := NewUserRepository(client, nil).(*userRepository)
	actionRepo := NewInviteAdminActionRepository(client)

	operator := createInviteAdminServiceUser(t, userRepo, "invite-admin-rebind-operator@example.com", nil, nil)
	oldInviter := createInviteAdminServiceUser(t, userRepo, "invite-admin-rebind-old@example.com", nil, nil)
	newInviter := createInviteAdminServiceUser(t, userRepo, "invite-admin-rebind-new@example.com", nil, nil)
	invitee := createInviteAdminServiceUser(
		t,
		userRepo,
		"invite-admin-rebind-invitee@example.com",
		int64Ptr(oldInviter.ID),
		timePtr(time.Date(2026, 4, 12, 8, 0, 0, 0, time.UTC)),
	)
	cleanupInviteAdminServiceUsers(t, operator.ID, oldInviter.ID, newInviter.ID, invitee.ID)

	svc := newInviteAdminServiceForIntegration(
		userRepo,
		nil,
		&failingRelationshipEventAdminRepo{err: errors.New("event write failed")},
		actionRepo,
		nil,
		client,
	)

	err := svc.RebindInviter(ctx, service.RebindInviterInput{
		OperatorUserID:   operator.ID,
		InviteeUserID:    invitee.ID,
		NewInviterUserID: newInviter.ID,
		Reason:           "integration rebind rollback",
	})
	require.ErrorContains(t, err, "event write failed")

	reloadedInvitee, getErr := userRepo.GetByID(ctx, invitee.ID)
	require.NoError(t, getErr)
	require.NotNil(t, reloadedInvitee.InvitedByUserID)
	require.Equal(t, oldInviter.ID, *reloadedInvitee.InvitedByUserID)
	require.Zero(t, countInviteAdminActionsByReason(t, "integration rebind rollback"))
}

func TestAdminService_CreateManualInviteGrant_RollsBackWhenBalanceUpdateFails(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	baseUserRepo := NewUserRepository(client, nil).(*userRepository)
	userRepo := &failingBalanceUserRepo{
		userRepository: baseUserRepo,
		err:            errors.New("balance update failed"),
	}
	rewardRepo := NewInviteRewardRecordRepository(client).(*inviteRewardRecordRepository)
	actionRepo := NewInviteAdminActionRepository(client)

	operator := createInviteAdminServiceUser(t, baseUserRepo, "invite-admin-manual-operator@example.com", nil, nil)
	target := createInviteAdminServiceUser(t, baseUserRepo, "invite-admin-manual-target@example.com", nil, nil)
	inviter := createInviteAdminServiceUser(t, baseUserRepo, "invite-admin-manual-inviter@example.com", nil, nil)
	cleanupInviteAdminServiceUsers(t, operator.ID, target.ID, inviter.ID)

	originalTarget, err := baseUserRepo.GetByID(ctx, target.ID)
	require.NoError(t, err)

	svc := newInviteAdminServiceForIntegration(
		userRepo,
		rewardRepo,
		nil,
		actionRepo,
		nil,
		client,
	)

	err = svc.CreateManualInviteGrant(ctx, service.ManualInviteGrantInput{
		OperatorUserID: operator.ID,
		TargetUserID:   target.ID,
		Reason:         "integration manual rollback",
		Lines: []service.ManualInviteGrantLine{{
			InviterUserID:      inviter.ID,
			InviteeUserID:      target.ID,
			RewardTargetUserID: target.ID,
			RewardRole:         service.InviteRewardRoleInvitee,
			RewardAmount:       18.5,
			Notes:              "integration-manual-rollback",
		}},
	})
	require.ErrorContains(t, err, "balance update failed")

	reloadedTarget, getErr := baseUserRepo.GetByID(ctx, target.ID)
	require.NoError(t, getErr)
	require.Equal(t, originalTarget.Balance, reloadedTarget.Balance)
	require.Zero(t, countInviteAdminActionsByReason(t, "integration manual rollback"))
	require.Zero(t, countInviteRewardRecordsByTypeAndTarget(t, service.InviteRewardTypeManualGrant, target.ID))
}

func TestAdminService_ExecuteInviteRecompute_RollsBackWhenBalanceUpdateFails(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	baseUserRepo := NewUserRepository(client, nil).(*userRepository)
	userRepo := &failingBalanceUserRepo{
		userRepository: baseUserRepo,
		err:            errors.New("balance update failed"),
	}
	rewardRepo := NewInviteRewardRecordRepository(client).(*inviteRewardRecordRepository)
	actionRepo := NewInviteAdminActionRepository(client)

	operator := createInviteAdminServiceUser(t, baseUserRepo, "invite-admin-recompute-operator@example.com", nil, nil)
	inviter := createInviteAdminServiceUser(t, baseUserRepo, "invite-admin-recompute-inviter@example.com", nil, nil)
	invitee := createInviteAdminServiceUser(t, baseUserRepo, "invite-admin-recompute-invitee@example.com", nil, nil)
	cleanupInviteAdminServiceUsers(t, operator.ID, inviter.ID, invitee.ID)

	svc := newInviteAdminServiceForIntegration(
		userRepo,
		rewardRepo,
		&staticEffectiveInviterRepo{inviters: map[int64]int64{invitee.ID: inviter.ID}},
		actionRepo,
		&staticQualifyingRechargeRepo{
			events: []service.InviteQualifyingRecharge{
				{
					InviteeUserID:          invitee.ID,
					TriggerRedeemCodeID:    101,
					TriggerRedeemCodeValue: 100,
					UsedAt:                 time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC),
				},
			},
		},
		client,
	)

	preview, err := svc.PreviewInviteRecompute(ctx, service.InviteRecomputeInput{
		OperatorUserID: operator.ID,
		Reason:         "integration recompute rollback",
		Scope:          service.InviteRecomputeScope{InviteeUserID: int64Ptr(invitee.ID)},
	})
	require.NoError(t, err)

	err = svc.ExecuteInviteRecompute(ctx, service.InviteRecomputeExecuteInput{
		OperatorUserID: operator.ID,
		Reason:         "integration recompute rollback",
		Scope:          service.InviteRecomputeScope{InviteeUserID: int64Ptr(invitee.ID)},
		ScopeHash:      preview.ScopeHash,
	})
	require.ErrorContains(t, err, "balance update failed")

	require.Zero(t, countInviteAdminActionsByReason(t, "integration recompute rollback"))
	require.Zero(t, countInviteRewardRecordsByTypeAndTarget(t, service.InviteRewardTypeRecomputeDelta, inviter.ID))
	require.Zero(t, countInviteRewardRecordsByTypeAndTarget(t, service.InviteRewardTypeRecomputeDelta, invitee.ID))
}

func TestInviteRewardRecordRepository_SumRewardTotalsForScope_ExcludesManualGrant(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	userRepo := NewUserRepository(client, nil).(*userRepository)
	rewardRepo := NewInviteRewardRecordRepository(client).(*inviteRewardRecordRepository)

	inviter := createInviteAdminServiceUser(t, userRepo, "invite-admin-sum-inviter@example.com", nil, nil)
	invitee := createInviteAdminServiceUser(t, userRepo, "invite-admin-sum-invitee@example.com", nil, nil)

	err := rewardRepo.CreateBatch(ctx, []service.InviteRewardRecord{
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      invitee.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeBase,
			RewardAmount:       5,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      invitee.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeManualGrant,
			RewardAmount:       11,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      invitee.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeRecomputeDelta,
			RewardAmount:       2,
			Status:             "applied",
		},
	})
	require.NoError(t, err)

	totals, err := rewardRepo.SumRewardTotalsForScope(ctx, service.InviteRecomputeScope{
		InviteeUserID: int64Ptr(invitee.ID),
	})
	require.NoError(t, err)
	require.Equal(t, 7.0, totals[inviteRecomputeScopeKey(inviter.ID, invitee.ID, inviter.ID, service.InviteRewardRoleInviter)])
}

func createInviteAdminServiceUser(t *testing.T, repo *userRepository, email string, invitedByUserID *int64, inviteBoundAt *time.Time) *service.User {
	t.Helper()

	user := &service.User{
		Email:           email,
		PasswordHash:    "hash",
		Role:            service.RoleUser,
		Status:          service.StatusActive,
		InviteCode:      inviteCodeForEmail(email),
		InvitedByUserID: invitedByUserID,
		InviteBoundAt:   inviteBoundAt,
	}
	require.NoError(t, repo.Create(context.Background(), user))
	return user
}

func cleanupInviteAdminServiceUsers(t *testing.T, userIDs ...int64) {
	t.Helper()
	if len(userIDs) == 0 {
		return
	}

	t.Cleanup(func() {
		ctx := context.Background()
		_, err := integrationDB.ExecContext(ctx, `
			DELETE FROM invite_reward_records
			WHERE inviter_user_id = ANY($1)
			   OR invitee_user_id = ANY($1)
			   OR reward_target_user_id = ANY($1)
		`, pq.Array(userIDs))
		require.NoError(t, err)

		_, err = integrationDB.ExecContext(ctx, `
			DELETE FROM invite_relationship_events
			WHERE invitee_user_id = ANY($1)
			   OR previous_inviter_user_id = ANY($1)
			   OR new_inviter_user_id = ANY($1)
			   OR operator_user_id = ANY($1)
		`, pq.Array(userIDs))
		require.NoError(t, err)

		_, err = integrationDB.ExecContext(ctx, `
			DELETE FROM invite_admin_actions
			WHERE operator_user_id = ANY($1)
			   OR target_user_id = ANY($1)
		`, pq.Array(userIDs))
		require.NoError(t, err)

		_, err = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id = ANY($1)`, pq.Array(userIDs))
		require.NoError(t, err)
	})
}

func countInviteAdminActionsByReason(t *testing.T, reason string) int {
	t.Helper()

	var total int
	err := integrationDB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM invite_admin_actions WHERE reason = $1`, reason).Scan(&total)
	require.NoError(t, err)
	return total
}

func countInviteRewardRecordsByTypeAndTarget(t *testing.T, rewardType string, rewardTargetUserID int64) int {
	t.Helper()

	var total int
	err := integrationDB.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM invite_reward_records WHERE reward_type = $1 AND reward_target_user_id = $2`,
		rewardType,
		rewardTargetUserID,
	).Scan(&total)
	require.NoError(t, err)
	return total
}

func newInviteAdminServiceForIntegration(
	userRepo service.UserRepository,
	rewardRepo service.InviteRewardAdminRepository,
	relationshipRepo service.InviteRelationshipEventAdminRepository,
	actionRepo service.InviteAdminActionRepository,
	qualifyingRepo service.InviteQualifyingRechargeRepository,
	entClient *dbent.Client,
) service.AdminService {
	return service.NewAdminService(
		userRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		rewardRepo,
		relationshipRepo,
		actionRepo,
		qualifyingRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		entClient,
		nil,
		nil,
		nil,
		nil,
	)
}
