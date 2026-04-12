package repository

import (
	"context"
	stdsql "database/sql"
	"strconv"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbinviterewardrecord "github.com/Wei-Shaw/sub2api/ent/inviterewardrecord"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type inviteRewardRecordRepository struct {
	client *dbent.Client
}

func NewInviteRewardRecordRepository(client *dbent.Client) service.InviteRewardRecordRepository {
	return &inviteRewardRecordRepository{client: client}
}

func (r *inviteRewardRecordRepository) CreateBatch(ctx context.Context, records []service.InviteRewardRecord) error {
	if len(records) == 0 {
		return nil
	}

	client := clientFromContext(ctx, r.client)
	creates := make([]*dbent.InviteRewardRecordCreate, 0, len(records))
	for i := range records {
		record := records[i]
		create := client.InviteRewardRecord.Create().
			SetInviterUserID(record.InviterUserID).
			SetInviteeUserID(record.InviteeUserID).
			SetTriggerRedeemCodeValue(record.TriggerRedeemCodeValue).
			SetRewardTargetUserID(record.RewardTargetUserID).
			SetRewardRole(record.RewardRole).
			SetRewardType(record.RewardType).
			SetRewardAmount(record.RewardAmount).
			SetStatus(record.Status)
		if record.TriggerRedeemCodeID != nil {
			create.SetTriggerRedeemCodeID(*record.TriggerRedeemCodeID)
		}
		if record.RewardRate != nil {
			create.SetRewardRate(*record.RewardRate)
		}
		if record.Notes != "" {
			create.SetNotes(record.Notes)
		}
		if record.AdminActionID != nil {
			create.SetAdminActionID(*record.AdminActionID)
		}
		creates = append(creates, create)
	}
	if err := client.InviteRewardRecord.CreateBulk(creates...).Exec(ctx); err != nil {
		return translatePersistenceError(err, nil, service.ErrInviteRewardAlreadyRecorded)
	}
	return nil
}

func (r *inviteRewardRecordRepository) ListByRewardTarget(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.InviteRewardRecord, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := client.InviteRewardRecord.Query().
		Where(dbinviterewardrecord.RewardTargetUserIDEQ(userID)).
		Order(dbinviterewardrecord.ByCreatedAt(sql.OrderDesc()))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	rows, err := query.Offset(params.Offset()).Limit(params.Limit()).All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return mapInviteRewardRecordEntities(rows), &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.Limit(),
	}, nil
}

func (r *inviteRewardRecordRepository) ListByAdminActionID(ctx context.Context, adminActionID int64) ([]service.InviteRewardRecord, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.InviteRewardRecord.Query().
		Where(dbinviterewardrecord.AdminActionIDEQ(adminActionID)).
		Order(dbinviterewardrecord.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapInviteRewardRecordEntities(rows), nil
}

func (r *inviteRewardRecordRepository) SumBaseRewardsByTargetAndRole(ctx context.Context, userID int64, rewardRole string) (float64, error) {
	client := clientFromContext(ctx, r.client)

	var rows []struct {
		Sum stdsql.NullFloat64 `json:"sum"`
	}

	err := client.InviteRewardRecord.Query().
		Where(
			dbinviterewardrecord.RewardTargetUserIDEQ(userID),
			dbinviterewardrecord.RewardTypeEQ(service.InviteRewardTypeBase),
			dbinviterewardrecord.RewardRoleEQ(rewardRole),
		).
		Aggregate(dbent.As(dbent.Sum(dbinviterewardrecord.FieldRewardAmount), "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 || !rows[0].Sum.Valid {
		return 0, nil
	}
	return rows[0].Sum.Float64, nil
}

func (r *inviteRewardRecordRepository) ExistsBaseRewardByRedeemCodeID(ctx context.Context, redeemCodeID int64) (bool, error) {
	if redeemCodeID <= 0 {
		return false, nil
	}

	client := clientFromContext(ctx, r.client)
	count, err := client.InviteRewardRecord.Query().
		Where(
			dbinviterewardrecord.TriggerRedeemCodeIDEQ(redeemCodeID),
			dbinviterewardrecord.RewardTypeEQ(service.InviteRewardTypeBase),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *inviteRewardRecordRepository) SumRewardTotalsForScope(ctx context.Context, scope service.InviteRecomputeScope) (map[string]float64, error) {
	client := clientFromContext(ctx, r.client)

	query := client.InviteRewardRecord.Query().Where(
		dbinviterewardrecord.RewardTypeIn(
			service.InviteRewardTypeBase,
			service.InviteRewardTypeRecomputeDelta,
		),
	)
	if scope.InviteeUserID != nil {
		query = query.Where(dbinviterewardrecord.InviteeUserIDEQ(*scope.InviteeUserID))
	}
	if scope.InviterUserID != nil {
		query = query.Where(dbinviterewardrecord.InviterUserIDEQ(*scope.InviterUserID))
	}
	if scope.StartAt != nil {
		query = query.Where(dbinviterewardrecord.CreatedAtGTE(*scope.StartAt))
	}
	if scope.EndAt != nil {
		query = query.Where(dbinviterewardrecord.CreatedAtLTE(*scope.EndAt))
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	totals := make(map[string]float64, len(rows))
	for i := range rows {
		row := rows[i]
		key := inviteRecomputeScopeKey(row.InviterUserID, row.InviteeUserID, row.RewardTargetUserID, row.RewardRole)
		totals[key] += row.RewardAmount
	}
	return totals, nil
}

func inviteRecomputeScopeKey(inviterUserID, inviteeUserID, rewardTargetUserID int64, rewardRole string) string {
	return formatInt64(inviterUserID) + ":" + formatInt64(inviteeUserID) + ":" + formatInt64(rewardTargetUserID) + ":" + rewardRole
}

func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func mapInviteRewardRecordEntities(rows []*dbent.InviteRewardRecord) []service.InviteRewardRecord {
	items := make([]service.InviteRewardRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, service.InviteRewardRecord{
			ID:                     row.ID,
			InviterUserID:          row.InviterUserID,
			InviteeUserID:          row.InviteeUserID,
			TriggerRedeemCodeID:    row.TriggerRedeemCodeID,
			TriggerRedeemCodeValue: row.TriggerRedeemCodeValue,
			RewardTargetUserID:     row.RewardTargetUserID,
			RewardRole:             row.RewardRole,
			RewardType:             row.RewardType,
			RewardRate:             row.RewardRate,
			RewardAmount:           row.RewardAmount,
			Status:                 row.Status,
			Notes:                  derefString(row.Notes),
			CreatedAt:              row.CreatedAt,
			AdminActionID:          row.AdminActionID,
		})
	}
	return items
}
