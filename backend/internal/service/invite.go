package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	InviteRewardRoleInviter = "inviter"
	InviteRewardRoleInvitee = "invitee"

	InviteRewardTypeBase        = "base_invite_reward"
	InviteRewardTypeManualGrant = "manual_invite_grant"
	InviteRewardTypeRecomputeDelta = "recompute_delta"

	InviteRelationshipEventTypeRegisterBind = "register_bind"
	InviteRelationshipEventTypeAdminRebind  = "admin_rebind"

	InviteAdminActionTypeRebind      = "rebind_inviter"
	InviteAdminActionTypeManualGrant = "manual_reward_grant"
	InviteAdminActionTypeRecompute   = "recompute_rewards"

	InviteBaseRewardRate = 0.03
)

var ErrInviteRewardAlreadyRecorded = infraerrors.Conflict("INVITE_REWARD_ALREADY_RECORDED", "invite reward already recorded")

type InviteRewardRecord struct {
	ID                     int64
	InviterUserID          int64
	InviteeUserID          int64
	TriggerRedeemCodeID    *int64
	TriggerRedeemCodeValue float64
	RewardTargetUserID     int64
	RewardRole             string
	RewardType             string
	RewardRate             *float64
	RewardAmount           float64
	Status                 string
	Notes                  string
	CreatedAt              time.Time
	AdminActionID          *int64
}

type InviteRelationshipEvent struct {
	ID                    int64
	InviteeUserID         int64
	PreviousInviterUserID *int64
	NewInviterUserID      *int64
	EventType             string
	EffectiveAt           time.Time
	OperatorUserID        *int64
	Reason                string
	CreatedAt             time.Time
}

type InviteAdminAction struct {
	ID                  int64
	ActionType          string
	OperatorUserID      int64
	TargetUserID        int64
	Reason              string
	RequestSnapshotJSON map[string]any
	ResultSnapshotJSON  map[string]any
	CreatedAt           time.Time
}

type InviteRelationshipEventRepository interface {
	Create(ctx context.Context, event *InviteRelationshipEvent) error
	ListByInvitee(ctx context.Context, inviteeUserID int64) ([]InviteRelationshipEvent, error)
}

type InviteAdminActionRepository interface {
	Create(ctx context.Context, action *InviteAdminAction) error
}

type InviteSummary struct {
	InviteCode            string
	InviteLink            string
	InvitedUsersTotal     int64
	InviteesRechargeTotal float64
	BaseRewardsTotal      float64
}
