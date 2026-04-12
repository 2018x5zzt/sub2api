package repository

import (
	"context"
	"math"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbinviteadminaction "github.com/Wei-Shaw/sub2api/ent/inviteadminaction"
	dbinviterelationshipevent "github.com/Wei-Shaw/sub2api/ent/inviterelationshipevent"
	dbinviterewardrecord "github.com/Wei-Shaw/sub2api/ent/inviterewardrecord"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type inviteAdminQueryRepository struct {
	client *dbent.Client
}

func NewInviteAdminQueryRepository(client *dbent.Client) service.InviteAdminQueryRepository {
	return &inviteAdminQueryRepository{client: client}
}

func (r *inviteAdminQueryRepository) GetStats(ctx context.Context) (*service.AdminInviteStats, error) {
	client := clientFromContext(ctx, r.client)

	totalInvitedUsers, err := client.User.Query().
		Where(dbuser.InvitedByUserIDNotNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	rewardRows, err := client.InviteRewardRecord.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	stats := &service.AdminInviteStats{
		TotalInvitedUsers: int64(totalInvitedUsers),
	}
	qualifiedInvitees := make(map[int64]struct{})
	for _, row := range rewardRows {
		switch row.RewardType {
		case service.InviteRewardTypeBase:
			stats.BaseRewardsTotal += row.RewardAmount
			qualifiedInvitees[row.InviteeUserID] = struct{}{}
		case service.InviteRewardTypeManualGrant:
			stats.ManualGrantsTotal += row.RewardAmount
		case service.InviteRewardTypeRecomputeDelta:
			stats.RecomputeAdjustmentsTotal += row.RewardAmount
		}
	}
	stats.QualifiedRewardUsersTotal = int64(len(qualifiedInvitees))
	return stats, nil
}

func (r *inviteAdminQueryRepository) ListRelationships(ctx context.Context, params pagination.PaginationParams, filters service.AdminInviteRelationshipFilters) ([]service.AdminInviteRelationship, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := client.User.Query().Where(dbuser.InvitedByUserIDNotNil())
	if filters.InviterUserID != nil {
		query = query.Where(dbuser.InvitedByUserIDEQ(*filters.InviterUserID))
	}
	if filters.InviteeUserID != nil {
		query = query.Where(dbuser.IDEQ(*filters.InviteeUserID))
	}
	if filters.StartAt != nil {
		query = query.Where(dbuser.InviteBoundAtGTE(*filters.StartAt))
	}
	if filters.EndAt != nil {
		query = query.Where(dbuser.InviteBoundAtLTE(*filters.EndAt))
	}

	users, err := query.Order(dbuser.ByInviteBoundAt(sql.OrderDesc()), dbuser.ByID(sql.OrderDesc())).All(ctx)
	if err != nil {
		return nil, nil, err
	}

	emailByUserID, err := r.userEmailMap(ctx, collectRelationshipUserIDs(users))
	if err != nil {
		return nil, nil, err
	}
	lastEventsByInvitee, err := r.latestRelationshipEvents(ctx, collectInviteeIDs(users))
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.AdminInviteRelationship, 0, len(users))
	search := strings.ToLower(strings.TrimSpace(filters.Search))
	for _, user := range users {
		item := service.AdminInviteRelationship{
			InviteeUserID:        user.ID,
			InviteeEmail:         user.Email,
			InviteCode:           derefString(user.InviteCode),
			CurrentInviterUserID: user.InvitedByUserID,
			InviteBoundAt:        user.InviteBoundAt,
		}
		if user.InvitedByUserID != nil {
			item.CurrentInviterEmail = emailByUserID[*user.InvitedByUserID]
		}
		if event, ok := lastEventsByInvitee[user.ID]; ok {
			item.LastEventType = event.EventType
			lastEventAt := event.EffectiveAt
			item.LastEventAt = &lastEventAt
		}
		if search != "" && !relationshipMatchesSearch(item, search) {
			continue
		}
		items = append(items, item)
	}

	start, end, pageResult := paginateBounds(len(items), params)
	return items[start:end], pageResult, nil
}

func (r *inviteAdminQueryRepository) ListRewards(ctx context.Context, params pagination.PaginationParams, filters service.AdminInviteRewardFilters) ([]service.AdminInviteRewardRow, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := client.InviteRewardRecord.Query()
	if filters.RewardType != "" {
		query = query.Where(dbinviterewardrecord.RewardTypeEQ(filters.RewardType))
	}
	if filters.TargetUserID != nil {
		query = query.Where(dbinviterewardrecord.RewardTargetUserIDEQ(*filters.TargetUserID))
	}
	if filters.StartAt != nil {
		query = query.Where(dbinviterewardrecord.CreatedAtGTE(*filters.StartAt))
	}
	if filters.EndAt != nil {
		query = query.Where(dbinviterewardrecord.CreatedAtLTE(*filters.EndAt))
	}

	rows, err := query.Order(dbinviterewardrecord.ByCreatedAt(sql.OrderDesc()), dbinviterewardrecord.ByID(sql.OrderDesc())).All(ctx)
	if err != nil {
		return nil, nil, err
	}

	emailByUserID, err := r.userEmailMap(ctx, collectRewardUserIDs(rows))
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.AdminInviteRewardRow, 0, len(rows))
	search := strings.ToLower(strings.TrimSpace(filters.Search))
	for _, row := range rows {
		item := service.AdminInviteRewardRow{
			RewardTargetUserID:  row.RewardTargetUserID,
			RewardTargetEmail:   emailByUserID[row.RewardTargetUserID],
			InviterUserID:       row.InviterUserID,
			InviterEmail:        emailByUserID[row.InviterUserID],
			InviteeUserID:       row.InviteeUserID,
			InviteeEmail:        emailByUserID[row.InviteeUserID],
			RewardRole:          row.RewardRole,
			RewardType:          row.RewardType,
			RewardAmount:        row.RewardAmount,
			CreatedAt:           row.CreatedAt,
			AdminActionID:       row.AdminActionID,
			TriggerRedeemCodeID: row.TriggerRedeemCodeID,
		}
		if search != "" && !rewardMatchesSearch(item, search) {
			continue
		}
		items = append(items, item)
	}

	start, end, pageResult := paginateBounds(len(items), params)
	return items[start:end], pageResult, nil
}

func (r *inviteAdminQueryRepository) ListActions(ctx context.Context, params pagination.PaginationParams, filters service.InviteAdminActionFilters) ([]service.InviteAdminAction, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := client.InviteAdminAction.Query()
	if filters.ActionType != "" {
		query = query.Where(dbinviteadminaction.ActionTypeEQ(filters.ActionType))
	}
	if filters.TargetUserID != nil {
		query = query.Where(dbinviteadminaction.TargetUserIDEQ(*filters.TargetUserID))
	}
	if filters.OperatorUserID != nil {
		query = query.Where(dbinviteadminaction.OperatorUserIDEQ(*filters.OperatorUserID))
	}
	if filters.StartAt != nil {
		query = query.Where(dbinviteadminaction.CreatedAtGTE(*filters.StartAt))
	}
	if filters.EndAt != nil {
		query = query.Where(dbinviteadminaction.CreatedAtLTE(*filters.EndAt))
	}

	rows, err := query.Order(dbinviteadminaction.ByCreatedAt(sql.OrderDesc()), dbinviteadminaction.ByID(sql.OrderDesc())).All(ctx)
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.InviteAdminAction, 0, len(rows))
	for _, row := range rows {
		item := service.InviteAdminAction{}
		applyInviteAdminActionEntityToService(&item, row)
		items = append(items, item)
	}

	start, end, pageResult := paginateBounds(len(items), params)
	return items[start:end], pageResult, nil
}

func (r *inviteAdminQueryRepository) userEmailMap(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	if len(userIDs) == 0 {
		return map[int64]string{}, nil
	}

	client := clientFromContext(ctx, r.client)
	users, err := client.User.Query().Where(dbuser.IDIn(userIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}

	emailByUserID := make(map[int64]string, len(users))
	for _, user := range users {
		emailByUserID[user.ID] = user.Email
	}
	return emailByUserID, nil
}

func (r *inviteAdminQueryRepository) latestRelationshipEvents(ctx context.Context, inviteeIDs []int64) (map[int64]*dbent.InviteRelationshipEvent, error) {
	if len(inviteeIDs) == 0 {
		return map[int64]*dbent.InviteRelationshipEvent{}, nil
	}

	client := clientFromContext(ctx, r.client)
	rows, err := client.InviteRelationshipEvent.Query().
		Where(dbinviterelationshipevent.InviteeUserIDIn(inviteeIDs...)).
		Order(
			dbinviterelationshipevent.ByEffectiveAt(sql.OrderDesc()),
			dbinviterelationshipevent.ByID(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[int64]*dbent.InviteRelationshipEvent, len(inviteeIDs))
	for _, row := range rows {
		if _, exists := out[row.InviteeUserID]; exists {
			continue
		}
		out[row.InviteeUserID] = row
	}
	return out, nil
}

func collectRelationshipUserIDs(users []*dbent.User) []int64 {
	ids := make(map[int64]struct{}, len(users)*2)
	for _, user := range users {
		ids[user.ID] = struct{}{}
		if user.InvitedByUserID != nil {
			ids[*user.InvitedByUserID] = struct{}{}
		}
	}
	return mapKeys(ids)
}

func collectInviteeIDs(users []*dbent.User) []int64 {
	ids := make([]int64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func collectRewardUserIDs(rows []*dbent.InviteRewardRecord) []int64 {
	ids := make(map[int64]struct{}, len(rows)*3)
	for _, row := range rows {
		ids[row.RewardTargetUserID] = struct{}{}
		ids[row.InviterUserID] = struct{}{}
		ids[row.InviteeUserID] = struct{}{}
	}
	return mapKeys(ids)
}

func mapKeys(values map[int64]struct{}) []int64 {
	ids := make([]int64, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	return ids
}

func relationshipMatchesSearch(item service.AdminInviteRelationship, search string) bool {
	return strings.Contains(strings.ToLower(item.InviteeEmail), search) ||
		strings.Contains(strings.ToLower(item.CurrentInviterEmail), search) ||
		strings.Contains(strings.ToLower(item.InviteCode), search)
}

func rewardMatchesSearch(item service.AdminInviteRewardRow, search string) bool {
	return strings.Contains(strings.ToLower(item.RewardTargetEmail), search) ||
		strings.Contains(strings.ToLower(item.InviterEmail), search) ||
		strings.Contains(strings.ToLower(item.InviteeEmail), search) ||
		strings.Contains(strings.ToLower(item.RewardType), search)
}

func paginateBounds(total int, params pagination.PaginationParams) (int, int, *pagination.PaginationResult) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.Limit()
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	if pages < 1 {
		pages = 1
	}
	return start, end, &pagination.PaginationResult{
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}
}

func withinRange(value time.Time, startAt, endAt *time.Time) bool {
	if startAt != nil && value.Before(*startAt) {
		return false
	}
	if endAt != nil && value.After(*endAt) {
		return false
	}
	return true
}
