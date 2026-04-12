package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

func (s *adminServiceImpl) GetInviteStats(ctx context.Context) (*AdminInviteStats, error) {
	return s.inviteAdminQueryRepo.GetStats(ctx)
}

func (s *adminServiceImpl) ListInviteRelationships(ctx context.Context, page, pageSize int, filters AdminInviteRelationshipFilters) ([]AdminInviteRelationship, int64, error) {
	rows, result, err := s.inviteAdminQueryRepo.ListRelationships(ctx, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}, filters)
	if err != nil {
		return nil, 0, err
	}
	return rows, result.Total, nil
}

func (s *adminServiceImpl) ListInviteRewards(ctx context.Context, page, pageSize int, filters AdminInviteRewardFilters) ([]AdminInviteRewardRow, int64, error) {
	rows, result, err := s.inviteAdminQueryRepo.ListRewards(ctx, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}, filters)
	if err != nil {
		return nil, 0, err
	}
	return rows, result.Total, nil
}

func (s *adminServiceImpl) ListInviteActions(ctx context.Context, page, pageSize int, filters InviteAdminActionFilters) ([]InviteAdminAction, int64, error) {
	rows, result, err := s.inviteAdminQueryRepo.ListActions(ctx, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}, filters)
	if err != nil {
		return nil, 0, err
	}
	return rows, result.Total, nil
}

func (s *adminServiceImpl) RebindInviter(ctx context.Context, input RebindInviterInput) error {
	if strings.TrimSpace(input.Reason) == "" {
		return infraerrors.BadRequest("INVITE_REASON_REQUIRED", "reason is required")
	}
	if input.InviteeUserID <= 0 || input.NewInviterUserID <= 0 {
		return infraerrors.BadRequest("INVITE_USER_ID_REQUIRED", "invitee_user_id and new_inviter_user_id are required")
	}
	if input.InviteeUserID == input.NewInviterUserID {
		return infraerrors.BadRequest("INVITE_SELF_BIND", "invitee cannot invite self")
	}
	if s.inviteBindingUserRepo == nil {
		return infraerrors.ServiceUnavailable("INVITE_REBIND_UNAVAILABLE", "invite rebinding is not available")
	}
	if s.inviteAdminActionRepo == nil || s.inviteRelationshipEventRepo == nil {
		return infraerrors.ServiceUnavailable("INVITE_ADMIN_WRITE_UNAVAILABLE", "invite admin write path is not available")
	}

	invitee, err := s.userRepo.GetByID(ctx, input.InviteeUserID)
	if err != nil {
		return err
	}
	if _, err := s.userRepo.GetByID(ctx, input.NewInviterUserID); err != nil {
		return err
	}
	if invitee.InvitedByUserID != nil && *invitee.InvitedByUserID == input.NewInviterUserID {
		return infraerrors.BadRequest("INVITE_REBIND_NOOP", "new inviter must differ from current inviter")
	}

	return s.withInviteAdminWriteTx(ctx, func(txCtx context.Context) error {
		action := &InviteAdminAction{
			ActionType:     InviteAdminActionTypeRebind,
			OperatorUserID: input.OperatorUserID,
			TargetUserID:   input.InviteeUserID,
			Reason:         strings.TrimSpace(input.Reason),
			RequestSnapshotJSON: map[string]any{
				"invitee_user_id":     input.InviteeUserID,
				"new_inviter_user_id": input.NewInviterUserID,
			},
			ResultSnapshotJSON: map[string]any{
				"historical_rewards_changed": false,
			},
		}
		if err := s.inviteAdminActionRepo.Create(txCtx, action); err != nil {
			return err
		}

		if err := s.inviteBindingUserRepo.UpdateInviterBinding(txCtx, input.InviteeUserID, &input.NewInviterUserID); err != nil {
			return err
		}

		return s.inviteRelationshipEventRepo.Create(txCtx, &InviteRelationshipEvent{
			InviteeUserID:         input.InviteeUserID,
			PreviousInviterUserID: invitee.InvitedByUserID,
			NewInviterUserID:      inviteInt64Ptr(input.NewInviterUserID),
			EventType:             InviteRelationshipEventTypeAdminRebind,
			EffectiveAt:           time.Now().UTC(),
			OperatorUserID:        inviteInt64Ptr(input.OperatorUserID),
			Reason:                strings.TrimSpace(input.Reason),
		})
	})
}

func (s *adminServiceImpl) CreateManualInviteGrant(ctx context.Context, input ManualInviteGrantInput) error {
	if strings.TrimSpace(input.Reason) == "" {
		return infraerrors.BadRequest("INVITE_REASON_REQUIRED", "reason is required")
	}
	if len(input.Lines) == 0 {
		return infraerrors.BadRequest("INVITE_GRANT_LINES_REQUIRED", "at least one line is required")
	}
	if s.inviteRewardRecordRepo == nil || s.inviteAdminActionRepo == nil {
		return infraerrors.ServiceUnavailable("INVITE_ADMIN_WRITE_UNAVAILABLE", "invite admin write path is not available")
	}
	if _, err := s.userRepo.GetByID(ctx, input.TargetUserID); err != nil {
		return err
	}

	return s.withInviteAdminWriteTx(ctx, func(txCtx context.Context) error {
		action := &InviteAdminAction{
			ActionType:     InviteAdminActionTypeManualGrant,
			OperatorUserID: input.OperatorUserID,
			TargetUserID:   input.TargetUserID,
			Reason:         strings.TrimSpace(input.Reason),
		}
		if err := s.inviteAdminActionRepo.Create(txCtx, action); err != nil {
			return err
		}

		rows := make([]InviteRewardRecord, 0, len(input.Lines))
		for i := range input.Lines {
			line := input.Lines[i]
			rows = append(rows, InviteRewardRecord{
				InviterUserID:      line.InviterUserID,
				InviteeUserID:      line.InviteeUserID,
				RewardTargetUserID: line.RewardTargetUserID,
				RewardRole:         line.RewardRole,
				RewardType:         InviteRewardTypeManualGrant,
				RewardAmount:       line.RewardAmount,
				Status:             "applied",
				Notes:              line.Notes,
				AdminActionID:      &action.ID,
			})
		}
		if err := s.inviteRewardRecordRepo.CreateBatch(txCtx, rows); err != nil {
			return err
		}

		for i := range input.Lines {
			line := input.Lines[i]
			if err := s.userRepo.UpdateBalance(txCtx, line.RewardTargetUserID, line.RewardAmount); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *adminServiceImpl) PreviewInviteRecompute(ctx context.Context, input InviteRecomputeInput) (*InviteRecomputePreview, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return nil, infraerrors.BadRequest("INVITE_REASON_REQUIRED", "reason is required")
	}
	if s.inviteQualifyingRechargeRepo == nil || s.inviteRelationshipEventRepo == nil || s.inviteRewardRecordRepo == nil {
		return nil, infraerrors.ServiceUnavailable("INVITE_RECOMPUTE_UNAVAILABLE", "invite recompute is not available")
	}

	events, err := s.inviteQualifyingRechargeRepo.ListInviteQualifyingRecharges(ctx, input.Scope)
	if err != nil {
		return nil, err
	}

	expectedTotals := map[string]float64{}
	for i := range events {
		event := events[i]
		inviterUserID, err := s.inviteRelationshipEventRepo.GetEffectiveInviterAt(ctx, event.InviteeUserID, event.UsedAt)
		if err != nil || inviterUserID == nil {
			continue
		}
		if input.Scope.InviterUserID != nil && *input.Scope.InviterUserID != *inviterUserID {
			continue
		}

		expectedTotals[inviteRecomputeKey(*inviterUserID, event.InviteeUserID, *inviterUserID, InviteRewardRoleInviter)] += event.TriggerRedeemCodeValue * InviteBaseRewardRate
		expectedTotals[inviteRecomputeKey(*inviterUserID, event.InviteeUserID, event.InviteeUserID, InviteRewardRoleInvitee)] += event.TriggerRedeemCodeValue * InviteBaseRewardRate
	}

	currentTotals, err := s.inviteRewardRecordRepo.SumRewardTotalsForScope(ctx, input.Scope)
	if err != nil {
		return nil, err
	}

	allKeys := map[string]struct{}{}
	for key := range expectedTotals {
		allKeys[key] = struct{}{}
	}
	for key := range currentTotals {
		allKeys[key] = struct{}{}
	}

	deltas := make([]InviteRecomputeDelta, 0, len(allKeys))
	for key := range allKeys {
		expected := expectedTotals[key]
		current := currentTotals[key]
		delta := expected - current
		if delta == 0 {
			continue
		}
		deltas = append(deltas, InviteRecomputeDeltaFromKey(key, current, expected, delta))
	}
	sortInviteRecomputeDeltas(deltas)

	return &InviteRecomputePreview{
		ScopeHash:            computeInviteRecomputeScopeHash(input.Scope),
		QualifyingEventCount: len(events),
		Deltas:               deltas,
	}, nil
}

func (s *adminServiceImpl) ExecuteInviteRecompute(ctx context.Context, input InviteRecomputeExecuteInput) error {
	if strings.TrimSpace(input.Reason) == "" {
		return infraerrors.BadRequest("INVITE_REASON_REQUIRED", "reason is required")
	}
	if input.ScopeHash != computeInviteRecomputeScopeHash(input.Scope) {
		return infraerrors.BadRequest("INVITE_RECOMPUTE_SCOPE_MISMATCH", "preview scope no longer matches")
	}
	if s.inviteAdminActionRepo == nil || s.inviteRewardRecordRepo == nil {
		return infraerrors.ServiceUnavailable("INVITE_RECOMPUTE_UNAVAILABLE", "invite recompute is not available")
	}

	preview, err := s.PreviewInviteRecompute(ctx, InviteRecomputeInput{
		OperatorUserID: input.OperatorUserID,
		Reason:         input.Reason,
		Scope:          input.Scope,
	})
	if err != nil {
		return err
	}
	if len(preview.Deltas) == 0 {
		return nil
	}

	return s.withInviteAdminWriteTx(ctx, func(txCtx context.Context) error {
		action := &InviteAdminAction{
			ActionType:     InviteAdminActionTypeRecompute,
			OperatorUserID: input.OperatorUserID,
			TargetUserID:   resolveInviteRecomputeActionTargetUserID(input.Scope, preview.Deltas),
			Reason:         strings.TrimSpace(input.Reason),
			RequestSnapshotJSON: map[string]any{
				"scope":      input.Scope,
				"scope_hash": input.ScopeHash,
			},
			ResultSnapshotJSON: map[string]any{
				"deltas": preview.Deltas,
			},
		}
		if err := s.inviteAdminActionRepo.Create(txCtx, action); err != nil {
			return err
		}

		rows := make([]InviteRewardRecord, 0, len(preview.Deltas))
		for i := range preview.Deltas {
			delta := preview.Deltas[i]
			rows = append(rows, InviteRewardRecord{
				InviterUserID:      delta.InviterUserID,
				InviteeUserID:      delta.InviteeUserID,
				RewardTargetUserID: delta.RewardTargetUserID,
				RewardRole:         delta.RewardRole,
				RewardType:         InviteRewardTypeRecomputeDelta,
				RewardAmount:       delta.DeltaAmount,
				Status:             "applied",
				Notes:              fmt.Sprintf("recompute delta: current=%0.8f expected=%0.8f", delta.CurrentAmount, delta.ExpectedAmount),
				AdminActionID:      &action.ID,
			})
		}
		if err := s.inviteRewardRecordRepo.CreateBatch(txCtx, rows); err != nil {
			return err
		}
		for i := range preview.Deltas {
			delta := preview.Deltas[i]
			if err := s.userRepo.UpdateBalance(txCtx, delta.RewardTargetUserID, delta.DeltaAmount); err != nil {
				return err
			}
		}
		return nil
	})
}

func inviteInt64Ptr(v int64) *int64 {
	return &v
}

func (s *adminServiceImpl) withInviteAdminWriteTx(ctx context.Context, fn func(context.Context) error) error {
	if dbent.TxFromContext(ctx) != nil || s.entClient == nil {
		return fn(ctx)
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func sortInviteRecomputeDeltas(deltas []InviteRecomputeDelta) {
	sort.Slice(deltas, func(i, j int) bool {
		return inviteRecomputeKey(
			deltas[i].InviterUserID,
			deltas[i].InviteeUserID,
			deltas[i].RewardTargetUserID,
			deltas[i].RewardRole,
		) < inviteRecomputeKey(
			deltas[j].InviterUserID,
			deltas[j].InviteeUserID,
			deltas[j].RewardTargetUserID,
			deltas[j].RewardRole,
		)
	})
}

func resolveInviteRecomputeActionTargetUserID(scope InviteRecomputeScope, deltas []InviteRecomputeDelta) int64 {
	if scope.InviteeUserID != nil {
		return *scope.InviteeUserID
	}
	if scope.InviterUserID != nil {
		return *scope.InviterUserID
	}
	if len(deltas) == 0 {
		return 0
	}
	if deltas[0].InviteeUserID > 0 {
		return deltas[0].InviteeUserID
	}
	return deltas[0].RewardTargetUserID
}
