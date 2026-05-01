package repository

import (
	"context"
	stdsql "database/sql"
	"strings"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbinviteadminaction "github.com/Wei-Shaw/sub2api/ent/inviteadminaction"
	dbinviterelationshipevent "github.com/Wei-Shaw/sub2api/ent/inviterelationshipevent"
	dbinviterewardrecord "github.com/Wei-Shaw/sub2api/ent/inviterewardrecord"
	dbpredicate "github.com/Wei-Shaw/sub2api/ent/predicate"
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

	qualifiedRewardUsersTotal, err := countDistinctInviteesByRewardType(ctx, client, service.InviteRewardTypeBase)
	if err != nil {
		return nil, err
	}

	baseRewardsTotal, err := sumInviteRewardAmountsByType(ctx, client, service.InviteRewardTypeBase)
	if err != nil {
		return nil, err
	}

	manualGrantsTotal, err := sumInviteRewardAmountsByType(ctx, client, service.InviteRewardTypeManualGrant)
	if err != nil {
		return nil, err
	}

	recomputeAdjustmentsTotal, err := sumInviteRewardAmountsByType(ctx, client, service.InviteRewardTypeRecomputeDelta)
	if err != nil {
		return nil, err
	}

	return &service.AdminInviteStats{
		TotalInvitedUsers:         int64(totalInvitedUsers),
		QualifiedRewardUsersTotal: qualifiedRewardUsersTotal,
		BaseRewardsTotal:          baseRewardsTotal,
		ManualGrantsTotal:         manualGrantsTotal,
		RecomputeAdjustmentsTotal: recomputeAdjustmentsTotal,
	}, nil
}

func (r *inviteAdminQueryRepository) ListRelationships(ctx context.Context, params pagination.PaginationParams, filters service.AdminInviteRelationshipFilters) ([]service.AdminInviteRelationship, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := applyInviteRelationshipFilters(client.User.Query().Where(dbuser.InvitedByUserIDNotNil()), filters)

	total, err := query.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	users, err := query.
		Order(dbuser.ByInviteBoundAt(sql.OrderDesc()), dbuser.ByID(sql.OrderDesc())).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
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
		items = append(items, item)
	}

	return items, paginationResultFromTotal(int64(total), params), nil
}

func (r *inviteAdminQueryRepository) ListRewards(ctx context.Context, params pagination.PaginationParams, filters service.AdminInviteRewardFilters) ([]service.AdminInviteRewardRow, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := applyInviteRewardFilters(client.InviteRewardRecord.Query(), filters)

	total, err := query.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	rows, err := query.
		Order(dbinviterewardrecord.ByCreatedAt(sql.OrderDesc()), dbinviterewardrecord.ByID(sql.OrderDesc())).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	emailByUserID, err := r.userEmailMap(ctx, collectRewardUserIDs(rows))
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.AdminInviteRewardRow, 0, len(rows))
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
		items = append(items, item)
	}

	return items, paginationResultFromTotal(int64(total), params), nil
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

	total, err := query.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	rows, err := query.
		Order(dbinviteadminaction.ByCreatedAt(sql.OrderDesc()), dbinviteadminaction.ByID(sql.OrderDesc())).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.InviteAdminAction, 0, len(rows))
	for _, row := range rows {
		item := service.InviteAdminAction{}
		applyInviteAdminActionEntityToService(&item, row)
		items = append(items, item)
	}

	return items, paginationResultFromTotal(int64(total), params), nil
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

func applyInviteRelationshipFilters(query *dbent.UserQuery, filters service.AdminInviteRelationshipFilters) *dbent.UserQuery {
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

	search := strings.TrimSpace(filters.Search)
	if search == "" {
		return query
	}

	return query.Where(dbpredicate.User(func(s *sql.Selector) {
		inviter := sql.Table(dbuser.Table).As("invite_relationship_inviter")
		s.LeftJoin(inviter).On(s.C(dbuser.FieldInvitedByUserID), inviter.C(dbuser.FieldID))
		s.Where(sql.Or(
			sql.ContainsFold(s.C(dbuser.FieldEmail), search),
			sql.ContainsFold(s.C(dbuser.FieldInviteCode), search),
			sql.ContainsFold(inviter.C(dbuser.FieldEmail), search),
		))
	}))
}

func applyInviteRewardFilters(query *dbent.InviteRewardRecordQuery, filters service.AdminInviteRewardFilters) *dbent.InviteRewardRecordQuery {
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

	search := strings.TrimSpace(filters.Search)
	if search == "" {
		return query
	}

	return query.Where(dbpredicate.InviteRewardRecord(func(s *sql.Selector) {
		inviter := sql.Table(dbuser.Table).As("invite_reward_inviter")
		invitee := sql.Table(dbuser.Table).As("invite_reward_invitee")
		rewardTarget := sql.Table(dbuser.Table).As("invite_reward_target")
		s.LeftJoin(inviter).On(s.C(dbinviterewardrecord.FieldInviterUserID), inviter.C(dbuser.FieldID))
		s.LeftJoin(invitee).On(s.C(dbinviterewardrecord.FieldInviteeUserID), invitee.C(dbuser.FieldID))
		s.LeftJoin(rewardTarget).On(s.C(dbinviterewardrecord.FieldRewardTargetUserID), rewardTarget.C(dbuser.FieldID))
		s.Where(sql.Or(
			sql.ContainsFold(s.C(dbinviterewardrecord.FieldRewardType), search),
			sql.ContainsFold(rewardTarget.C(dbuser.FieldEmail), search),
			sql.ContainsFold(inviter.C(dbuser.FieldEmail), search),
			sql.ContainsFold(invitee.C(dbuser.FieldEmail), search),
		))
	}))
}

func countDistinctInviteesByRewardType(ctx context.Context, client *dbent.Client, rewardType string) (int64, error) {
	var rows []struct {
		Count int64 `json:"count"`
	}

	err := client.InviteRewardRecord.Query().
		Where(dbinviterewardrecord.RewardTypeEQ(rewardType)).
		Aggregate(dbent.As(func(s *sql.Selector) string {
			return sql.Count(sql.Distinct(s.C(dbinviterewardrecord.FieldInviteeUserID)))
		}, "count")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Count, nil
}

func sumInviteRewardAmountsByType(ctx context.Context, client *dbent.Client, rewardType string) (float64, error) {
	var rows []struct {
		Sum stdsql.NullFloat64 `json:"sum"`
	}

	err := client.InviteRewardRecord.Query().
		Where(dbinviterewardrecord.RewardTypeEQ(rewardType)).
		Aggregate(dbent.As(func(s *sql.Selector) string {
			return sql.Sum(s.C(dbinviterewardrecord.FieldRewardAmount))
		}, "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 || !rows[0].Sum.Valid {
		return 0, nil
	}
	return rows[0].Sum.Float64, nil
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
