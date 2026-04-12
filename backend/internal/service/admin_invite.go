package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type AdminInviteStats struct {
	TotalInvitedUsers         int64
	QualifiedRewardUsersTotal int64
	BaseRewardsTotal          float64
	ManualGrantsTotal         float64
	RecomputeAdjustmentsTotal float64
}

type AdminInviteRelationship struct {
	InviteeUserID        int64
	InviteeEmail         string
	InviteCode           string
	CurrentInviterUserID *int64
	CurrentInviterEmail  string
	InviteBoundAt        *time.Time
	LastEventType        string
	LastEventAt          *time.Time
}

type AdminInviteRelationshipFilters struct {
	Search        string
	InviterUserID *int64
	InviteeUserID *int64
	StartAt       *time.Time
	EndAt         *time.Time
}

type AdminInviteRewardFilters struct {
	Search       string
	RewardType   string
	TargetUserID *int64
	StartAt      *time.Time
	EndAt        *time.Time
}

type AdminInviteRewardRow struct {
	RewardTargetUserID  int64
	RewardTargetEmail   string
	InviterUserID       int64
	InviterEmail        string
	InviteeUserID       int64
	InviteeEmail        string
	RewardRole          string
	RewardType          string
	RewardAmount        float64
	CreatedAt           time.Time
	AdminActionID       *int64
	TriggerRedeemCodeID *int64
}

type InviteAdminActionFilters struct {
	ActionType     string
	TargetUserID   *int64
	OperatorUserID *int64
	StartAt        *time.Time
	EndAt          *time.Time
}

type RebindInviterInput struct {
	OperatorUserID   int64
	InviteeUserID    int64
	NewInviterUserID int64
	Reason           string
}

type ManualInviteGrantLine struct {
	InviterUserID      int64
	InviteeUserID      int64
	RewardTargetUserID int64
	RewardRole         string
	RewardAmount       float64
	Notes              string
}

type ManualInviteGrantInput struct {
	OperatorUserID int64
	TargetUserID   int64
	Reason         string
	Lines          []ManualInviteGrantLine
}

type InviteRecomputeScope struct {
	InviteeUserID *int64
	InviterUserID *int64
	StartAt       *time.Time
	EndAt         *time.Time
}

type InviteRecomputeInput struct {
	OperatorUserID int64
	Reason         string
	Scope          InviteRecomputeScope
}

type InviteRecomputeExecuteInput struct {
	OperatorUserID int64
	Reason         string
	Scope          InviteRecomputeScope
	ScopeHash      string
}

type InviteRecomputePreview struct {
	ScopeHash            string
	QualifyingEventCount int
	Deltas               []InviteRecomputeDelta
}

type InviteRecomputeDelta struct {
	InviterUserID      int64
	InviteeUserID      int64
	RewardTargetUserID int64
	RewardRole         string
	CurrentAmount      float64
	ExpectedAmount     float64
	DeltaAmount        float64
}

type InviteQualifyingRecharge struct {
	InviteeUserID          int64
	TriggerRedeemCodeID    int64
	TriggerRedeemCodeValue float64
	UsedAt                 time.Time
}

type InviteAdminQueryRepository interface {
	GetStats(ctx context.Context) (*AdminInviteStats, error)
	ListRelationships(ctx context.Context, params pagination.PaginationParams, filters AdminInviteRelationshipFilters) ([]AdminInviteRelationship, *pagination.PaginationResult, error)
	ListRewards(ctx context.Context, params pagination.PaginationParams, filters AdminInviteRewardFilters) ([]AdminInviteRewardRow, *pagination.PaginationResult, error)
	ListActions(ctx context.Context, params pagination.PaginationParams, filters InviteAdminActionFilters) ([]InviteAdminAction, *pagination.PaginationResult, error)
}

type InviteRewardAdminRepository interface {
	CreateBatch(ctx context.Context, records []InviteRewardRecord) error
	SumRewardTotalsForScope(ctx context.Context, scope InviteRecomputeScope) (map[string]float64, error)
}

type InviteRelationshipEventAdminRepository interface {
	Create(ctx context.Context, event *InviteRelationshipEvent) error
	ListByInvitee(ctx context.Context, inviteeUserID int64) ([]InviteRelationshipEvent, error)
	GetEffectiveInviterAt(ctx context.Context, inviteeUserID int64, at time.Time) (*int64, error)
}

type InviteQualifyingRechargeRepository interface {
	ListInviteQualifyingRecharges(ctx context.Context, scope InviteRecomputeScope) ([]InviteQualifyingRecharge, error)
}

func computeInviteRecomputeScopeHash(scope InviteRecomputeScope) string {
	payload, _ := json.Marshal(scope)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func inviteRecomputeKey(inviterUserID, inviteeUserID, rewardTargetUserID int64, rewardRole string) string {
	return strconv.FormatInt(inviterUserID, 10) +
		":" + strconv.FormatInt(inviteeUserID, 10) +
		":" + strconv.FormatInt(rewardTargetUserID, 10) +
		":" + rewardRole
}

func InviteRecomputeDeltaFromKey(key string, current, expected, delta float64) InviteRecomputeDelta {
	parts := strings.Split(key, ":")
	if len(parts) != 4 {
		return InviteRecomputeDelta{
			CurrentAmount:  current,
			ExpectedAmount: expected,
			DeltaAmount:    delta,
		}
	}

	inviterUserID, _ := strconv.ParseInt(parts[0], 10, 64)
	inviteeUserID, _ := strconv.ParseInt(parts[1], 10, 64)
	rewardTargetUserID, _ := strconv.ParseInt(parts[2], 10, 64)

	return InviteRecomputeDelta{
		InviterUserID:      inviterUserID,
		InviteeUserID:      inviteeUserID,
		RewardTargetUserID: rewardTargetUserID,
		RewardRole:         parts[3],
		CurrentAmount:      current,
		ExpectedAmount:     expected,
		DeltaAmount:        delta,
	}
}
