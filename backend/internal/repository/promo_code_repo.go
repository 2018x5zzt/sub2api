package repository

import (
	"context"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/promocode"
	"github.com/Wei-Shaw/sub2api/ent/promocodeusage"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type promoCodeRepository struct {
	client *dbent.Client
}

func NewPromoCodeRepository(client *dbent.Client) service.PromoCodeRepository {
	return &promoCodeRepository{client: client}
}

func (r *promoCodeRepository) Create(ctx context.Context, code *service.PromoCode) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PromoCode.Create().
		SetCode(code.Code).
		SetScene(code.Scene).
		SetBonusAmount(code.BonusAmount).
		SetRandomBonusPoolAmount(code.RandomBonusPoolAmount).
		SetRandomBonusRemaining(code.RandomBonusRemaining).
		SetMaxUses(code.MaxUses).
		SetUsedCount(code.UsedCount).
		SetLeaderboardEnabled(code.LeaderboardEnabled).
		SetStatus(code.Status).
		SetSuccessMessage(code.SuccessMessage).
		SetNotes(code.Notes)

	if code.ExpiresAt != nil {
		builder.SetExpiresAt(*code.ExpiresAt)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	code.ID = created.ID
	code.CreatedAt = created.CreatedAt
	code.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *promoCodeRepository) GetByID(ctx context.Context, id int64) (*service.PromoCode, error) {
	m, err := r.client.PromoCode.Query().
		Where(promocode.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPromoCodeNotFound
		}
		return nil, err
	}
	return promoCodeEntityToService(m), nil
}

func (r *promoCodeRepository) GetByCode(ctx context.Context, code string) (*service.PromoCode, error) {
	m, err := r.client.PromoCode.Query().
		Where(promocode.CodeEqualFold(code)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPromoCodeNotFound
		}
		return nil, err
	}
	return promoCodeEntityToService(m), nil
}

func (r *promoCodeRepository) GetByCodeForUpdate(ctx context.Context, code string) (*service.PromoCode, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PromoCode.Query().
		Where(promocode.CodeEqualFold(code)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrPromoCodeNotFound
		}
		return nil, err
	}
	return promoCodeEntityToService(m), nil
}

func (r *promoCodeRepository) Update(ctx context.Context, code *service.PromoCode) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PromoCode.UpdateOneID(code.ID).
		SetCode(code.Code).
		SetScene(code.Scene).
		SetBonusAmount(code.BonusAmount).
		SetRandomBonusPoolAmount(code.RandomBonusPoolAmount).
		SetRandomBonusRemaining(code.RandomBonusRemaining).
		SetMaxUses(code.MaxUses).
		SetUsedCount(code.UsedCount).
		SetLeaderboardEnabled(code.LeaderboardEnabled).
		SetStatus(code.Status).
		SetSuccessMessage(code.SuccessMessage).
		SetNotes(code.Notes)

	if code.ExpiresAt != nil {
		builder.SetExpiresAt(*code.ExpiresAt)
	} else {
		builder.ClearExpiresAt()
	}
	if strings.TrimSpace(code.SuccessMessage) == "" {
		builder.ClearSuccessMessage()
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrPromoCodeNotFound
		}
		return err
	}

	code.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *promoCodeRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.PromoCode.Delete().Where(promocode.IDEQ(id)).Exec(ctx)
	return err
}

func (r *promoCodeRepository) List(ctx context.Context, params pagination.PaginationParams) ([]service.PromoCode, *pagination.PaginationResult, error) {
	return r.ListWithFilters(ctx, params, service.PromoCodeSceneRegister, "", "")
}

func (r *promoCodeRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, scene, status, search string) ([]service.PromoCode, *pagination.PaginationResult, error) {
	q := r.client.PromoCode.Query().
		Where(promocode.SceneEQ(scene))

	if status != "" {
		q = q.Where(promocode.StatusEQ(status))
	}
	if search != "" {
		q = q.Where(promocode.CodeContainsFold(search))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	codes, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(promocode.FieldID)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	outCodes := promoCodeEntitiesToService(codes)

	return outCodes, paginationResultFromTotal(int64(total), params), nil
}

func (r *promoCodeRepository) CreateUsage(ctx context.Context, usage *service.PromoCodeUsage) error {
	client := clientFromContext(ctx, r.client)
	created, err := client.PromoCodeUsage.Create().
		SetPromoCodeID(usage.PromoCodeID).
		SetUserID(usage.UserID).
		SetBonusAmount(usage.BonusAmount).
		SetFixedBonusAmount(usage.FixedBonusAmount).
		SetRandomBonusAmount(usage.RandomBonusAmount).
		SetUsedAt(usage.UsedAt).
		Save(ctx)
	if err != nil {
		return err
	}

	usage.ID = created.ID
	return nil
}

func (r *promoCodeRepository) GetUsageByPromoCodeAndUser(ctx context.Context, promoCodeID, userID int64) (*service.PromoCodeUsage, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PromoCodeUsage.Query().
		Where(
			promocodeusage.PromoCodeIDEQ(promoCodeID),
			promocodeusage.UserIDEQ(userID),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return promoCodeUsageEntityToService(m), nil
}

func (r *promoCodeRepository) ListAllUsagesByPromoCode(ctx context.Context, promoCodeID int64) ([]service.PromoCodeUsage, error) {
	client := clientFromContext(ctx, r.client)
	usages, err := client.PromoCodeUsage.Query().
		Where(promocodeusage.PromoCodeIDEQ(promoCodeID)).
		WithUser().
		Order(dbent.Desc(promocodeusage.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return promoCodeUsageEntitiesToService(usages), nil
}

func (r *promoCodeRepository) ListUsagesByPromoCode(ctx context.Context, promoCodeID int64, params pagination.PaginationParams) ([]service.PromoCodeUsage, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.PromoCodeUsage.Query().
		Where(promocodeusage.PromoCodeIDEQ(promoCodeID))

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	usages, err := q.
		WithUser().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(promocodeusage.FieldID)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	outUsages := promoCodeUsageEntitiesToService(usages)

	return outUsages, paginationResultFromTotal(int64(total), params), nil
}

func (r *promoCodeRepository) IncrementUsedCount(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.PromoCode.UpdateOneID(id).
		AddUsedCount(1).
		Save(ctx)
	return err
}

// Entity to Service conversions

func promoCodeEntityToService(m *dbent.PromoCode) *service.PromoCode {
	if m == nil {
		return nil
	}
	return &service.PromoCode{
		ID:                    m.ID,
		Code:                  m.Code,
		Scene:                 m.Scene,
		BonusAmount:           m.BonusAmount,
		RandomBonusPoolAmount: m.RandomBonusPoolAmount,
		RandomBonusRemaining:  m.RandomBonusRemaining,
		MaxUses:               m.MaxUses,
		UsedCount:             m.UsedCount,
		LeaderboardEnabled:    m.LeaderboardEnabled,
		Status:                m.Status,
		ExpiresAt:             m.ExpiresAt,
		SuccessMessage:        derefString(m.SuccessMessage),
		Notes:                 derefString(m.Notes),
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}

func promoCodeEntitiesToService(models []*dbent.PromoCode) []service.PromoCode {
	out := make([]service.PromoCode, 0, len(models))
	for i := range models {
		if s := promoCodeEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}

func promoCodeUsageEntityToService(m *dbent.PromoCodeUsage) *service.PromoCodeUsage {
	if m == nil {
		return nil
	}
	out := &service.PromoCodeUsage{
		ID:                m.ID,
		PromoCodeID:       m.PromoCodeID,
		UserID:            m.UserID,
		BonusAmount:       m.BonusAmount,
		FixedBonusAmount:  m.FixedBonusAmount,
		RandomBonusAmount: m.RandomBonusAmount,
		UsedAt:            m.UsedAt,
	}
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	return out
}

func promoCodeUsageEntitiesToService(models []*dbent.PromoCodeUsage) []service.PromoCodeUsage {
	out := make([]service.PromoCodeUsage, 0, len(models))
	for i := range models {
		if s := promoCodeUsageEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
