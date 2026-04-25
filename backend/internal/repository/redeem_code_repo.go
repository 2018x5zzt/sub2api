package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbredeemcode "github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type redeemCodeRepository struct {
	client *dbent.Client
}

func NewRedeemCodeRepository(client *dbent.Client) service.RedeemCodeRepository {
	return &redeemCodeRepository{client: client}
}

func (r *redeemCodeRepository) Create(ctx context.Context, code *service.RedeemCode) error {
	sourceType := normalizedRedeemSourceType(code.SourceType)
	created, err := r.client.RedeemCode.Create().
		SetCode(code.Code).
		SetType(code.Type).
		SetValue(code.Value).
		SetStatus(code.Status).
		SetSourceType(sourceType).
		SetNotes(code.Notes).
		SetValidityDays(code.ValidityDays).
		SetNillableUsedBy(code.UsedBy).
		SetNillableUsedAt(code.UsedAt).
		SetNillableGroupID(code.GroupID).
		SetNillableProductID(code.ProductID).
		Save(ctx)
	if err == nil {
		code.ID = created.ID
		code.CreatedAt = created.CreatedAt
	}
	return err
}

func (r *redeemCodeRepository) CreateBatch(ctx context.Context, codes []service.RedeemCode) error {
	if len(codes) == 0 {
		return nil
	}

	builders := make([]*dbent.RedeemCodeCreate, 0, len(codes))
	for i := range codes {
		c := &codes[i]
		b := r.client.RedeemCode.Create().
			SetCode(c.Code).
			SetType(c.Type).
			SetValue(c.Value).
			SetStatus(c.Status).
			SetSourceType(normalizedRedeemSourceType(c.SourceType)).
			SetNotes(c.Notes).
			SetValidityDays(c.ValidityDays).
			SetNillableUsedBy(c.UsedBy).
			SetNillableUsedAt(c.UsedAt).
			SetNillableGroupID(c.GroupID).
			SetNillableProductID(c.ProductID)
		builders = append(builders, b)
	}

	return r.client.RedeemCode.CreateBulk(builders...).Exec(ctx)
}

func (r *redeemCodeRepository) GetByID(ctx context.Context, id int64) (*service.RedeemCode, error) {
	m, err := r.client.RedeemCode.Query().
		Where(dbredeemcode.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrRedeemCodeNotFound
		}
		return nil, err
	}
	return redeemCodeEntityToService(m), nil
}

func (r *redeemCodeRepository) GetByCode(ctx context.Context, code string) (*service.RedeemCode, error) {
	m, err := r.client.RedeemCode.Query().
		Where(dbredeemcode.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrRedeemCodeNotFound
		}
		return nil, err
	}
	return redeemCodeEntityToService(m), nil
}

func (r *redeemCodeRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.client.RedeemCode.Delete().Where(dbredeemcode.IDEQ(id)).Exec(ctx)
	return err
}

func (r *redeemCodeRepository) List(ctx context.Context, params pagination.PaginationParams) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	return r.ListWithFilters(ctx, params, "", "", "")
}

func (r *redeemCodeRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, codeType, status, search string) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	q := r.client.RedeemCode.Query()

	if codeType != "" {
		q = q.Where(dbredeemcode.TypeEQ(codeType))
	}
	if status != "" {
		q = q.Where(dbredeemcode.StatusEQ(status))
	}
	if search != "" {
		q = q.Where(
			dbredeemcode.Or(
				dbredeemcode.CodeContainsFold(search),
				dbredeemcode.HasUserWith(user.EmailContainsFold(search)),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	codes, err := q.
		WithUser().
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(dbredeemcode.FieldID)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	outCodes := redeemCodeEntitiesToService(codes)

	return outCodes, paginationResultFromTotal(int64(total), params), nil
}

func (r *redeemCodeRepository) Update(ctx context.Context, code *service.RedeemCode) error {
	up := r.client.RedeemCode.UpdateOneID(code.ID).
		SetCode(code.Code).
		SetType(code.Type).
		SetValue(code.Value).
		SetStatus(code.Status).
		SetSourceType(normalizedRedeemSourceType(code.SourceType)).
		SetNotes(code.Notes).
		SetValidityDays(code.ValidityDays)

	if code.UsedBy != nil {
		up.SetUsedBy(*code.UsedBy)
	} else {
		up.ClearUsedBy()
	}
	if code.UsedAt != nil {
		up.SetUsedAt(*code.UsedAt)
	} else {
		up.ClearUsedAt()
	}
	if code.GroupID != nil {
		up.SetGroupID(*code.GroupID)
	} else {
		up.ClearGroupID()
	}
	if code.ProductID != nil {
		up.SetProductID(*code.ProductID)
	} else {
		up.ClearProductID()
	}

	updated, err := up.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrRedeemCodeNotFound
		}
		return err
	}
	code.CreatedAt = updated.CreatedAt
	return nil
}

func (r *redeemCodeRepository) Use(ctx context.Context, id, userID int64) error {
	now := time.Now()
	client := clientFromContext(ctx, r.client)
	affected, err := client.RedeemCode.Update().
		Where(dbredeemcode.IDEQ(id), dbredeemcode.StatusEQ(service.StatusUnused)).
		SetStatus(service.StatusUsed).
		SetUsedBy(userID).
		SetUsedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrRedeemCodeUsed
	}
	return nil
}

func (r *redeemCodeRepository) ListByUser(ctx context.Context, userID int64, limit int) ([]service.RedeemCode, error) {
	if limit <= 0 {
		limit = 10
	}

	codes, err := r.client.RedeemCode.Query().
		Where(dbredeemcode.UsedByEQ(userID)).
		WithGroup().
		Order(dbent.Desc(dbredeemcode.FieldUsedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return redeemCodeEntitiesToService(codes), nil
}

// ListByUserPaginated returns paginated balance/concurrency history for a user.
// Supports optional type filter (e.g. "balance", "admin_balance", "concurrency", "admin_concurrency", "subscription").
func (r *redeemCodeRepository) ListByUserPaginated(ctx context.Context, userID int64, params pagination.PaginationParams, codeType string) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	q := r.client.RedeemCode.Query().
		Where(dbredeemcode.UsedByEQ(userID))

	// Optional type filter
	if codeType != "" {
		q = q.Where(dbredeemcode.TypeEQ(codeType))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	codes, err := q.
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(dbredeemcode.FieldUsedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return redeemCodeEntitiesToService(codes), paginationResultFromTotal(int64(total), params), nil
}

// SumPositiveBalanceByUser returns total recharged amount (sum of value > 0 where type is balance/admin_balance).
func (r *redeemCodeRepository) SumPositiveBalanceByUser(ctx context.Context, userID int64) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}
	err := r.client.RedeemCode.Query().
		Where(
			dbredeemcode.UsedByEQ(userID),
			dbredeemcode.ValueGT(0),
			dbredeemcode.TypeIn("balance", "admin_balance"),
		).
		Aggregate(dbent.As(dbent.Sum(dbredeemcode.FieldValue), "sum")).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Sum, nil
}

func redeemCodeEntityToService(m *dbent.RedeemCode) *service.RedeemCode {
	if m == nil {
		return nil
	}
	out := &service.RedeemCode{
		ID:           m.ID,
		Code:         m.Code,
		Type:         m.Type,
		Value:        m.Value,
		Status:       m.Status,
		SourceType:   normalizedRedeemSourceType(m.SourceType),
		UsedBy:       m.UsedBy,
		UsedAt:       m.UsedAt,
		Notes:        derefString(m.Notes),
		CreatedAt:    m.CreatedAt,
		GroupID:      m.GroupID,
		ProductID:    m.ProductID,
		ValidityDays: m.ValidityDays,
	}
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	if m.Edges.Group != nil {
		out.Group = groupEntityToService(m.Edges.Group)
	}
	return out
}

func (r *redeemCodeRepository) ListInviteQualifyingRecharges(ctx context.Context, scope service.InviteRecomputeScope) ([]service.InviteQualifyingRecharge, error) {
	client := clientFromContext(ctx, r.client)
	query := client.RedeemCode.Query().
		Where(
			dbredeemcode.TypeEQ(service.RedeemTypeBalance),
			dbredeemcode.SourceTypeEQ(service.RedeemSourceCommercial),
			dbredeemcode.StatusEQ(service.StatusUsed),
			dbredeemcode.UsedByNotNil(),
			dbredeemcode.UsedAtNotNil(),
		)

	if scope.InviteeUserID != nil {
		query = query.Where(dbredeemcode.UsedByEQ(*scope.InviteeUserID))
	}
	if scope.StartAt != nil {
		query = query.Where(dbredeemcode.UsedAtGTE(*scope.StartAt))
	}
	if scope.EndAt != nil {
		query = query.Where(dbredeemcode.UsedAtLTE(*scope.EndAt))
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	return mapQualifyingRecharges(rows), nil
}

func normalizedRedeemSourceType(sourceType string) string {
	return service.NormalizeRedeemSourceType(sourceType, service.RedeemSourceSystemGrant)
}

func redeemCodeEntitiesToService(models []*dbent.RedeemCode) []service.RedeemCode {
	out := make([]service.RedeemCode, 0, len(models))
	for i := range models {
		if s := redeemCodeEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}

func mapQualifyingRecharges(rows []*dbent.RedeemCode) []service.InviteQualifyingRecharge {
	out := make([]service.InviteQualifyingRecharge, 0, len(rows))
	for i := range rows {
		row := rows[i]
		if row.UsedBy == nil || row.UsedAt == nil {
			continue
		}
		out = append(out, service.InviteQualifyingRecharge{
			InviteeUserID:          *row.UsedBy,
			TriggerRedeemCodeID:    row.ID,
			TriggerRedeemCodeValue: row.Value,
			UsedAt:                 *row.UsedAt,
		})
	}
	return out
}
