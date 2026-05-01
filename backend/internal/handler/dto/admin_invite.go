package dto

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminInviteStats struct {
	TotalInvitedUsers         int64   `json:"total_invited_users"`
	QualifiedRewardUsersTotal int64   `json:"qualified_reward_users_total"`
	BaseRewardsTotal          float64 `json:"base_rewards_total"`
	ManualGrantsTotal         float64 `json:"manual_grants_total"`
	RecomputeAdjustmentsTotal float64 `json:"recompute_adjustments_total"`
}

type AdminInviteRelationship struct {
	InviteeUserID        int64      `json:"invitee_user_id"`
	InviteeEmail         string     `json:"invitee_email"`
	InviteCode           string     `json:"invite_code"`
	CurrentInviterUserID *int64     `json:"current_inviter_user_id,omitempty"`
	CurrentInviterEmail  string     `json:"current_inviter_email,omitempty"`
	InviteBoundAt        *time.Time `json:"invite_bound_at,omitempty"`
	LastEventType        string     `json:"last_event_type,omitempty"`
	LastEventAt          *time.Time `json:"last_event_at,omitempty"`
}

type AdminInviteRewardRow struct {
	RewardTargetUserID  int64     `json:"reward_target_user_id"`
	RewardTargetEmail   string    `json:"reward_target_email"`
	InviterUserID       int64     `json:"inviter_user_id"`
	InviterEmail        string    `json:"inviter_email"`
	InviteeUserID       int64     `json:"invitee_user_id"`
	InviteeEmail        string    `json:"invitee_email"`
	RewardRole          string    `json:"reward_role"`
	RewardType          string    `json:"reward_type"`
	RewardAmount        float64   `json:"reward_amount"`
	CreatedAt           time.Time `json:"created_at"`
	AdminActionID       *int64    `json:"admin_action_id,omitempty"`
	TriggerRedeemCodeID *int64    `json:"trigger_redeem_code_id,omitempty"`
}

type AdminInviteAction struct {
	ID                  int64          `json:"id"`
	ActionType          string         `json:"action_type"`
	OperatorUserID      int64          `json:"operator_user_id"`
	TargetUserID        int64          `json:"target_user_id"`
	Reason              string         `json:"reason"`
	RequestSnapshotJSON map[string]any `json:"request_snapshot_json"`
	ResultSnapshotJSON  map[string]any `json:"result_snapshot_json"`
	CreatedAt           time.Time      `json:"created_at"`
}

type AdminInviteRecomputeDelta struct {
	InviterUserID      int64   `json:"inviter_user_id"`
	InviteeUserID      int64   `json:"invitee_user_id"`
	RewardTargetUserID int64   `json:"reward_target_user_id"`
	RewardRole         string  `json:"reward_role"`
	CurrentAmount      float64 `json:"current_amount"`
	ExpectedAmount     float64 `json:"expected_amount"`
	DeltaAmount        float64 `json:"delta_amount"`
}

type AdminInviteRecomputePreview struct {
	ScopeHash            string                      `json:"scope_hash"`
	QualifyingEventCount int                         `json:"qualifying_event_count"`
	Deltas               []AdminInviteRecomputeDelta `json:"deltas"`
}

func AdminInviteStatsFromService(in *service.AdminInviteStats) *AdminInviteStats {
	if in == nil {
		return nil
	}
	return &AdminInviteStats{
		TotalInvitedUsers:         in.TotalInvitedUsers,
		QualifiedRewardUsersTotal: in.QualifiedRewardUsersTotal,
		BaseRewardsTotal:          in.BaseRewardsTotal,
		ManualGrantsTotal:         in.ManualGrantsTotal,
		RecomputeAdjustmentsTotal: in.RecomputeAdjustmentsTotal,
	}
}

func AdminInviteRelationshipsFromService(in []service.AdminInviteRelationship) []AdminInviteRelationship {
	out := make([]AdminInviteRelationship, 0, len(in))
	for i := range in {
		out = append(out, AdminInviteRelationship{
			InviteeUserID:        in[i].InviteeUserID,
			InviteeEmail:         in[i].InviteeEmail,
			InviteCode:           in[i].InviteCode,
			CurrentInviterUserID: in[i].CurrentInviterUserID,
			CurrentInviterEmail:  in[i].CurrentInviterEmail,
			InviteBoundAt:        in[i].InviteBoundAt,
			LastEventType:        in[i].LastEventType,
			LastEventAt:          in[i].LastEventAt,
		})
	}
	return out
}

func AdminInviteRewardsFromService(in []service.AdminInviteRewardRow) []AdminInviteRewardRow {
	out := make([]AdminInviteRewardRow, 0, len(in))
	for i := range in {
		out = append(out, AdminInviteRewardRow{
			RewardTargetUserID:  in[i].RewardTargetUserID,
			RewardTargetEmail:   in[i].RewardTargetEmail,
			InviterUserID:       in[i].InviterUserID,
			InviterEmail:        in[i].InviterEmail,
			InviteeUserID:       in[i].InviteeUserID,
			InviteeEmail:        in[i].InviteeEmail,
			RewardRole:          in[i].RewardRole,
			RewardType:          in[i].RewardType,
			RewardAmount:        in[i].RewardAmount,
			CreatedAt:           in[i].CreatedAt,
			AdminActionID:       in[i].AdminActionID,
			TriggerRedeemCodeID: in[i].TriggerRedeemCodeID,
		})
	}
	return out
}

func AdminInviteActionsFromService(in []service.InviteAdminAction) []AdminInviteAction {
	out := make([]AdminInviteAction, 0, len(in))
	for i := range in {
		out = append(out, AdminInviteAction{
			ID:                  in[i].ID,
			ActionType:          in[i].ActionType,
			OperatorUserID:      in[i].OperatorUserID,
			TargetUserID:        in[i].TargetUserID,
			Reason:              in[i].Reason,
			RequestSnapshotJSON: in[i].RequestSnapshotJSON,
			ResultSnapshotJSON:  in[i].ResultSnapshotJSON,
			CreatedAt:           in[i].CreatedAt,
		})
	}
	return out
}

func AdminInviteRecomputePreviewFromService(in *service.InviteRecomputePreview) *AdminInviteRecomputePreview {
	if in == nil {
		return nil
	}
	out := &AdminInviteRecomputePreview{
		ScopeHash:            in.ScopeHash,
		QualifyingEventCount: in.QualifyingEventCount,
		Deltas:               make([]AdminInviteRecomputeDelta, 0, len(in.Deltas)),
	}
	for i := range in.Deltas {
		delta := in.Deltas[i]
		out.Deltas = append(out.Deltas, AdminInviteRecomputeDelta{
			InviterUserID:      delta.InviterUserID,
			InviteeUserID:      delta.InviteeUserID,
			RewardTargetUserID: delta.RewardTargetUserID,
			RewardRole:         delta.RewardRole,
			CurrentAmount:      delta.CurrentAmount,
			ExpectedAmount:     delta.ExpectedAmount,
			DeltaAmount:        delta.DeltaAmount,
		})
	}
	return out
}

func AdminInviteRelationshipFiltersFromRequest(c *gin.Context) (service.AdminInviteRelationshipFilters, error) {
	inviterUserID, err := parseOptionalInt64Query(c.Query("inviter_user_id"))
	if err != nil {
		return service.AdminInviteRelationshipFilters{}, err
	}
	inviteeUserID, err := parseOptionalInt64Query(c.Query("invitee_user_id"))
	if err != nil {
		return service.AdminInviteRelationshipFilters{}, err
	}
	startAt, err := parseOptionalTimeQuery(c.Query("start_at"))
	if err != nil {
		return service.AdminInviteRelationshipFilters{}, err
	}
	endAt, err := parseOptionalTimeQuery(c.Query("end_at"))
	if err != nil {
		return service.AdminInviteRelationshipFilters{}, err
	}
	return service.AdminInviteRelationshipFilters{
		Search:        strings.TrimSpace(c.Query("search")),
		InviterUserID: inviterUserID,
		InviteeUserID: inviteeUserID,
		StartAt:       startAt,
		EndAt:         endAt,
	}, nil
}

func AdminInviteRewardFiltersFromRequest(c *gin.Context) (service.AdminInviteRewardFilters, error) {
	targetUserID, err := parseOptionalInt64Query(c.Query("target_user_id"))
	if err != nil {
		return service.AdminInviteRewardFilters{}, err
	}
	startAt, err := parseOptionalTimeQuery(c.Query("start_at"))
	if err != nil {
		return service.AdminInviteRewardFilters{}, err
	}
	endAt, err := parseOptionalTimeQuery(c.Query("end_at"))
	if err != nil {
		return service.AdminInviteRewardFilters{}, err
	}
	return service.AdminInviteRewardFilters{
		Search:       strings.TrimSpace(c.Query("search")),
		RewardType:   strings.TrimSpace(c.Query("reward_type")),
		TargetUserID: targetUserID,
		StartAt:      startAt,
		EndAt:        endAt,
	}, nil
}

func InviteAdminActionFiltersFromRequest(c *gin.Context) (service.InviteAdminActionFilters, error) {
	targetUserID, err := parseOptionalInt64Query(c.Query("target_user_id"))
	if err != nil {
		return service.InviteAdminActionFilters{}, err
	}
	operatorUserID, err := parseOptionalInt64Query(c.Query("operator_user_id"))
	if err != nil {
		return service.InviteAdminActionFilters{}, err
	}
	startAt, err := parseOptionalTimeQuery(c.Query("start_at"))
	if err != nil {
		return service.InviteAdminActionFilters{}, err
	}
	endAt, err := parseOptionalTimeQuery(c.Query("end_at"))
	if err != nil {
		return service.InviteAdminActionFilters{}, err
	}
	return service.InviteAdminActionFilters{
		ActionType:     strings.TrimSpace(c.Query("action_type")),
		TargetUserID:   targetUserID,
		OperatorUserID: operatorUserID,
		StartAt:        startAt,
		EndAt:          endAt,
	}, nil
}

func parseOptionalInt64Query(raw string) (*int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseOptionalTimeQuery(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return &parsed, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
