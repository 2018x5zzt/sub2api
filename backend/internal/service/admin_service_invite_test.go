//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type inviteAdminQueryRepoStub struct {
	stats         *AdminInviteStats
	relationships []AdminInviteRelationship
	rewards       []AdminInviteRewardRow
	actions       []InviteAdminAction
}

func inviteAdminInt64Ptr(v int64) *int64 { return &v }

func (s *inviteAdminQueryRepoStub) GetStats(context.Context) (*AdminInviteStats, error) {
	return s.stats, nil
}

func (s *inviteAdminQueryRepoStub) ListRelationships(context.Context, pagination.PaginationParams, AdminInviteRelationshipFilters) ([]AdminInviteRelationship, *pagination.PaginationResult, error) {
	return s.relationships, &pagination.PaginationResult{Total: int64(len(s.relationships)), Page: 1, PageSize: 20}, nil
}

func (s *inviteAdminQueryRepoStub) ListRewards(context.Context, pagination.PaginationParams, AdminInviteRewardFilters) ([]AdminInviteRewardRow, *pagination.PaginationResult, error) {
	return s.rewards, &pagination.PaginationResult{Total: int64(len(s.rewards)), Page: 1, PageSize: 20}, nil
}

func (s *inviteAdminQueryRepoStub) ListActions(context.Context, pagination.PaginationParams, InviteAdminActionFilters) ([]InviteAdminAction, *pagination.PaginationResult, error) {
	return s.actions, &pagination.PaginationResult{Total: int64(len(s.actions)), Page: 1, PageSize: 20}, nil
}

func TestAdminInviteReads(t *testing.T) {
	svc := &adminServiceImpl{
		inviteAdminQueryRepo: &inviteAdminQueryRepoStub{
			stats: &AdminInviteStats{
				TotalInvitedUsers:         3,
				QualifiedRewardUsersTotal: 2,
				BaseRewardsTotal:          15,
				ManualGrantsTotal:         5,
				RecomputeAdjustmentsTotal: -2,
			},
			relationships: []AdminInviteRelationship{
				{
					InviteeUserID:        8,
					InviteeEmail:         "invitee@example.com",
					CurrentInviterUserID: inviteAdminInt64Ptr(7),
					CurrentInviterEmail:  "inviter@example.com",
				},
			},
			rewards: []AdminInviteRewardRow{
				{RewardType: InviteRewardTypeBase, RewardAmount: 5},
			},
			actions: []InviteAdminAction{
				{ActionType: InviteAdminActionTypeManualGrant, Reason: "operator repair"},
			},
		},
	}

	stats, err := svc.GetInviteStats(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(3), stats.TotalInvitedUsers)

	relationships, total, err := svc.ListInviteRelationships(context.Background(), 1, 20, AdminInviteRelationshipFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "invitee@example.com", relationships[0].InviteeEmail)

	rewards, total, err := svc.ListInviteRewards(context.Background(), 1, 20, AdminInviteRewardFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, InviteRewardTypeBase, rewards[0].RewardType)

	actions, total, err := svc.ListInviteActions(context.Background(), 1, 20, InviteAdminActionFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, InviteAdminActionTypeManualGrant, actions[0].ActionType)
}

type inviteWriteUserRepoStub struct {
	users              map[int64]*User
	updateBindingCalls []struct {
		inviteeID int64
		inviterID *int64
	}
	balanceUpdates []struct {
		userID int64
		amount float64
	}
}

func (s *inviteWriteUserRepoStub) Create(context.Context, *User) error { return nil }

func (s *inviteWriteUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	copy := *user
	return &copy, nil
}

func (s *inviteWriteUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	return nil, ErrUserNotFound
}

func (s *inviteWriteUserRepoStub) GetByInviteCode(context.Context, string) (*User, error) {
	return nil, ErrUserNotFound
}

func (s *inviteWriteUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	return nil, ErrUserNotFound
}

func (s *inviteWriteUserRepoStub) Update(context.Context, *User) error { return nil }

func (s *inviteWriteUserRepoStub) Delete(context.Context, int64) error { return nil }

func (s *inviteWriteUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (s *inviteWriteUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (s *inviteWriteUserRepoStub) UpdateBalance(_ context.Context, userID int64, amount float64) error {
	s.balanceUpdates = append(s.balanceUpdates, struct {
		userID int64
		amount float64
	}{userID: userID, amount: amount})
	return nil
}

func (s *inviteWriteUserRepoStub) DeductBalance(context.Context, int64, float64) error { return nil }

func (s *inviteWriteUserRepoStub) UpdateConcurrency(context.Context, int64, int) error { return nil }

func (s *inviteWriteUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	return false, nil
}

func (s *inviteWriteUserRepoStub) ExistsByInviteCode(context.Context, string) (bool, error) {
	return false, nil
}

func (s *inviteWriteUserRepoStub) CountInviteesByInviter(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *inviteWriteUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *inviteWriteUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	return nil
}

func (s *inviteWriteUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	return nil
}

func (s *inviteWriteUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error { return nil }

func (s *inviteWriteUserRepoStub) EnableTotp(context.Context, int64) error { return nil }

func (s *inviteWriteUserRepoStub) DisableTotp(context.Context, int64) error { return nil }

func (s *inviteWriteUserRepoStub) UpdateInviterBinding(_ context.Context, inviteeUserID int64, inviterUserID *int64) error {
	s.updateBindingCalls = append(s.updateBindingCalls, struct {
		inviteeID int64
		inviterID *int64
	}{inviteeID: inviteeUserID, inviterID: inviterUserID})
	if user, ok := s.users[inviteeUserID]; ok {
		user.InvitedByUserID = inviterUserID
	}
	return nil
}

type inviteAdminActionRepoStub struct {
	created []InviteAdminAction
}

func (s *inviteAdminActionRepoStub) Create(_ context.Context, action *InviteAdminAction) error {
	action.ID = int64(len(s.created) + 1)
	s.created = append(s.created, *action)
	return nil
}

type inviteRelationshipEventRepoStub struct {
	created           []InviteRelationshipEvent
	effectiveInviters map[int64]int64
}

func (s *inviteRelationshipEventRepoStub) Create(_ context.Context, event *InviteRelationshipEvent) error {
	if event == nil {
		return nil
	}
	event.ID = int64(len(s.created) + 1)
	if event.EffectiveAt.IsZero() {
		event.EffectiveAt = time.Now().UTC()
	}
	s.created = append(s.created, *event)
	return nil
}

func (s *inviteRelationshipEventRepoStub) ListByInvitee(context.Context, int64) ([]InviteRelationshipEvent, error) {
	return nil, nil
}

func (s *inviteRelationshipEventRepoStub) GetEffectiveInviterAt(_ context.Context, inviteeUserID int64, _ time.Time) (*int64, error) {
	if s.effectiveInviters == nil {
		return nil, nil
	}
	inviterID, ok := s.effectiveInviters[inviteeUserID]
	if !ok {
		return nil, nil
	}
	return &inviterID, nil
}

type inviteWriteRewardRepoStub struct {
	created     []InviteRewardRecord
	scopeTotals map[string]float64
}

func (s *inviteWriteRewardRepoStub) CreateBatch(_ context.Context, records []InviteRewardRecord) error {
	s.created = append(s.created, records...)
	return nil
}

func (s *inviteWriteRewardRepoStub) ListByRewardTarget(context.Context, int64, pagination.PaginationParams) ([]InviteRewardRecord, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (s *inviteWriteRewardRepoStub) ListByAdminActionID(context.Context, int64) ([]InviteRewardRecord, error) {
	return nil, nil
}

func (s *inviteWriteRewardRepoStub) SumBaseRewardsByTargetAndRole(context.Context, int64, string) (float64, error) {
	return 0, nil
}

func (s *inviteWriteRewardRepoStub) SumRewardTotalsForScope(_ context.Context, _ InviteRecomputeScope) (map[string]float64, error) {
	if s.scopeTotals == nil {
		return map[string]float64{}, nil
	}
	return s.scopeTotals, nil
}

type inviteQualifyingRechargeRepoStub struct {
	events []InviteQualifyingRecharge
}

func (s *inviteQualifyingRechargeRepoStub) ListInviteQualifyingRecharges(_ context.Context, _ InviteRecomputeScope) ([]InviteQualifyingRecharge, error) {
	return s.events, nil
}

func TestRebindInviter_OnlyChangesFutureBinding(t *testing.T) {
	inviteeCurrentInviter := int64(7)
	users := &inviteWriteUserRepoStub{
		users: map[int64]*User{
			7: {ID: 7, Email: "old@example.com", Status: StatusActive},
			9: {ID: 9, Email: "new@example.com", Status: StatusActive},
			8: {ID: 8, Email: "invitee@example.com", Status: StatusActive, InvitedByUserID: &inviteeCurrentInviter},
		},
	}

	svc := &adminServiceImpl{
		userRepo:                    users,
		inviteBindingUserRepo:       users,
		inviteAdminActionRepo:       &inviteAdminActionRepoStub{},
		inviteRelationshipEventRepo: &inviteRelationshipEventRepoStub{},
	}

	err := svc.RebindInviter(context.Background(), RebindInviterInput{
		OperatorUserID:   1,
		InviteeUserID:    8,
		NewInviterUserID: 9,
		Reason:           "fraud cleanup",
	})
	require.NoError(t, err)
	require.Len(t, users.updateBindingCalls, 1)
	require.Equal(t, int64(8), users.updateBindingCalls[0].inviteeID)
	require.NotNil(t, users.updateBindingCalls[0].inviterID)
	require.Equal(t, int64(9), *users.updateBindingCalls[0].inviterID)
}

func TestManualInviteGrant_AppendsLedgerAndCreditsBalance(t *testing.T) {
	users := &inviteWriteUserRepoStub{users: map[int64]*User{
		7: {ID: 7, Email: "inviter@example.com", Status: StatusActive},
	}}
	rewards := &inviteWriteRewardRepoStub{}

	svc := &adminServiceImpl{
		userRepo:               users,
		inviteRewardRecordRepo: rewards,
		inviteAdminActionRepo:  &inviteAdminActionRepoStub{},
	}

	err := svc.CreateManualInviteGrant(context.Background(), ManualInviteGrantInput{
		OperatorUserID: 1,
		TargetUserID:   7,
		Reason:         "reissue missing reward",
		Lines: []ManualInviteGrantLine{{
			InviterUserID:      7,
			InviteeUserID:      8,
			RewardTargetUserID: 7,
			RewardRole:         InviteRewardRoleInviter,
			RewardAmount:       6,
			Notes:              "missing historical reward",
		}},
	})
	require.NoError(t, err)
	require.Len(t, rewards.created, 1)
	require.Equal(t, InviteRewardTypeManualGrant, rewards.created[0].RewardType)
	require.Len(t, users.balanceUpdates, 1)
	require.Equal(t, 6.0, users.balanceUpdates[0].amount)
}

func TestPreviewInviteRecompute_ReturnsPositiveDelta(t *testing.T) {
	inviterID := int64(7)
	svc := &adminServiceImpl{
		inviteQualifyingRechargeRepo: &inviteQualifyingRechargeRepoStub{
			events: []InviteQualifyingRecharge{
				{InviteeUserID: 8, TriggerRedeemCodeID: 101, TriggerRedeemCodeValue: 100, UsedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
			},
		},
		inviteRelationshipEventRepo: &inviteRelationshipEventRepoStub{
			effectiveInviters: map[int64]int64{8: inviterID},
		},
		inviteRewardRecordRepo: &inviteWriteRewardRepoStub{
			scopeTotals: map[string]float64{
				"7:8:7:inviter": 0,
				"7:8:8:invitee": 0,
			},
		},
	}

	preview, err := svc.PreviewInviteRecompute(context.Background(), InviteRecomputeInput{
		OperatorUserID: 1,
		Reason:         "rebuild base rewards",
		Scope:          InviteRecomputeScope{InviteeUserID: inviteAdminInt64Ptr(8)},
	})
	require.NoError(t, err)
	require.Equal(t, 1, preview.QualifyingEventCount)
	require.Len(t, preview.Deltas, 2)
	got := []float64{preview.Deltas[0].DeltaAmount, preview.Deltas[1].DeltaAmount}
	require.ElementsMatch(t, []float64{3.0, 3.0}, got)
}

func TestPreviewInviteRecompute_ReturnsStableSortedDeltas(t *testing.T) {
	inviterID := int64(7)
	svc := &adminServiceImpl{
		inviteQualifyingRechargeRepo: &inviteQualifyingRechargeRepoStub{
			events: []InviteQualifyingRecharge{
				{InviteeUserID: 8, TriggerRedeemCodeID: 101, TriggerRedeemCodeValue: 100, UsedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
				{InviteeUserID: 9, TriggerRedeemCodeID: 102, TriggerRedeemCodeValue: 50, UsedAt: time.Date(2026, 4, 1, 11, 0, 0, 0, time.UTC)},
			},
		},
		inviteRelationshipEventRepo: &inviteRelationshipEventRepoStub{
			effectiveInviters: map[int64]int64{
				8: inviterID,
				9: inviterID,
			},
		},
		inviteRewardRecordRepo: &inviteWriteRewardRepoStub{scopeTotals: map[string]float64{}},
	}

	expected := []string{
		"7:8:7:inviter",
		"7:8:8:invitee",
		"7:9:7:inviter",
		"7:9:9:invitee",
	}

	for i := 0; i < 8; i++ {
		preview, err := svc.PreviewInviteRecompute(context.Background(), InviteRecomputeInput{
			OperatorUserID: 1,
			Reason:         "sorted output",
			Scope:          InviteRecomputeScope{InviterUserID: inviteAdminInt64Ptr(inviterID)},
		})
		require.NoError(t, err)
		require.Equal(t, expected, recomputeDeltaKeys(preview.Deltas))
	}
}

func TestExecuteInviteRecompute_WritesDeltaRowsOnly(t *testing.T) {
	users := &inviteWriteUserRepoStub{users: map[int64]*User{
		7: {ID: 7, Email: "inviter@example.com", Status: StatusActive},
		8: {ID: 8, Email: "invitee@example.com", Status: StatusActive},
	}}
	rewards := &inviteWriteRewardRepoStub{scopeTotals: map[string]float64{
		"7:8:7:inviter": 0,
		"7:8:8:invitee": 0,
	}}

	svc := &adminServiceImpl{
		userRepo: users,
		inviteQualifyingRechargeRepo: &inviteQualifyingRechargeRepoStub{
			events: []InviteQualifyingRecharge{
				{InviteeUserID: 8, TriggerRedeemCodeID: 101, TriggerRedeemCodeValue: 100, UsedAt: time.Now().UTC()},
			},
		},
		inviteRelationshipEventRepo: &inviteRelationshipEventRepoStub{effectiveInviters: map[int64]int64{8: 7}},
		inviteRewardRecordRepo:      rewards,
		inviteAdminActionRepo:       &inviteAdminActionRepoStub{},
	}

	preview, err := svc.PreviewInviteRecompute(context.Background(), InviteRecomputeInput{
		OperatorUserID: 1,
		Reason:         "rebuild base rewards",
		Scope:          InviteRecomputeScope{InviteeUserID: inviteAdminInt64Ptr(8)},
	})
	require.NoError(t, err)

	err = svc.ExecuteInviteRecompute(context.Background(), InviteRecomputeExecuteInput{
		OperatorUserID: 1,
		Reason:         "rebuild base rewards",
		Scope:          InviteRecomputeScope{InviteeUserID: inviteAdminInt64Ptr(8)},
		ScopeHash:      preview.ScopeHash,
	})
	require.NoError(t, err)
	require.Len(t, rewards.created, 2)
	require.Equal(t, InviteRewardTypeRecomputeDelta, rewards.created[0].RewardType)
}

func TestExecuteInviteRecompute_UsesScopeInviteeAsActionTarget(t *testing.T) {
	users := &inviteWriteUserRepoStub{users: map[int64]*User{
		7: {ID: 7, Email: "inviter@example.com", Status: StatusActive},
		8: {ID: 8, Email: "invitee@example.com", Status: StatusActive},
	}}
	rewards := &inviteWriteRewardRepoStub{scopeTotals: map[string]float64{
		"7:8:7:inviter": 0,
		"7:8:8:invitee": 0,
	}}
	actionRepo := &inviteAdminActionRepoStub{}

	svc := &adminServiceImpl{
		userRepo: users,
		inviteQualifyingRechargeRepo: &inviteQualifyingRechargeRepoStub{
			events: []InviteQualifyingRecharge{
				{InviteeUserID: 8, TriggerRedeemCodeID: 101, TriggerRedeemCodeValue: 100, UsedAt: time.Now().UTC()},
			},
		},
		inviteRelationshipEventRepo: &inviteRelationshipEventRepoStub{effectiveInviters: map[int64]int64{8: 7}},
		inviteRewardRecordRepo:      rewards,
		inviteAdminActionRepo:       actionRepo,
	}

	preview, err := svc.PreviewInviteRecompute(context.Background(), InviteRecomputeInput{
		OperatorUserID: 1,
		Reason:         "rebuild base rewards",
		Scope:          InviteRecomputeScope{InviteeUserID: inviteAdminInt64Ptr(8)},
	})
	require.NoError(t, err)

	err = svc.ExecuteInviteRecompute(context.Background(), InviteRecomputeExecuteInput{
		OperatorUserID: 1,
		Reason:         "rebuild base rewards",
		Scope:          InviteRecomputeScope{InviteeUserID: inviteAdminInt64Ptr(8)},
		ScopeHash:      preview.ScopeHash,
	})
	require.NoError(t, err)
	require.Len(t, actionRepo.created, 1)
	require.Equal(t, int64(8), actionRepo.created[0].TargetUserID)
}

func recomputeDeltaKeys(deltas []InviteRecomputeDelta) []string {
	keys := make([]string, 0, len(deltas))
	for i := range deltas {
		delta := deltas[i]
		keys = append(keys, inviteRecomputeKey(delta.InviterUserID, delta.InviteeUserID, delta.RewardTargetUserID, delta.RewardRole))
	}
	return keys
}
