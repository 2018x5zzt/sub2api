package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type inviteAdminActionRepository struct {
	client *dbent.Client
}

func NewInviteAdminActionRepository(client *dbent.Client) service.InviteAdminActionRepository {
	return &inviteAdminActionRepository{client: client}
}

func (r *inviteAdminActionRepository) Create(ctx context.Context, action *service.InviteAdminAction) error {
	if action == nil {
		return nil
	}

	client := clientFromContext(ctx, r.client)
	row, err := client.InviteAdminAction.Create().
		SetActionType(action.ActionType).
		SetOperatorUserID(action.OperatorUserID).
		SetTargetUserID(action.TargetUserID).
		SetReason(action.Reason).
		SetRequestSnapshotJSON(normalizeJSONMap(copyJSONMap(action.RequestSnapshotJSON))).
		SetResultSnapshotJSON(normalizeJSONMap(copyJSONMap(action.ResultSnapshotJSON))).
		Save(ctx)
	if err != nil {
		return err
	}
	applyInviteAdminActionEntityToService(action, row)
	return nil
}

func applyInviteAdminActionEntityToService(dst *service.InviteAdminAction, src *dbent.InviteAdminAction) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.ActionType = src.ActionType
	dst.OperatorUserID = src.OperatorUserID
	dst.TargetUserID = src.TargetUserID
	dst.Reason = src.Reason
	dst.RequestSnapshotJSON = copyJSONMap(src.RequestSnapshotJSON)
	dst.ResultSnapshotJSON = copyJSONMap(src.ResultSnapshotJSON)
	dst.CreatedAt = src.CreatedAt
}
