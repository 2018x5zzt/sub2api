package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionproduct"
	"github.com/Wei-Shaw/sub2api/ent/userproductsubscription"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrProductSubscriptionAssignerUnavailable = infraerrors.ServiceUnavailable(
	"PRODUCT_SUBSCRIPTION_ASSIGNER_UNAVAILABLE",
	"product subscription assigner is not configured",
)

type ProductSubscriptionAssigner interface {
	AssignOrExtendProductSubscription(ctx context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error)
}

type AssignProductSubscriptionInput struct {
	UserID       int64
	ProductID    int64
	ValidityDays int
	AssignedBy   int64
	Notes        string
}

type SubscriptionProductAdminService struct {
	entClient *dbent.Client
}

func NewSubscriptionProductAdminService(entClient *dbent.Client) *SubscriptionProductAdminService {
	return &SubscriptionProductAdminService{entClient: entClient}
}

func (s *SubscriptionProductAdminService) AssignOrExtendProductSubscription(ctx context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error) {
	if s == nil || s.entClient == nil {
		return nil, false, ErrProductSubscriptionAssignerUnavailable
	}
	if input == nil {
		return nil, false, ErrSubscriptionNilInput
	}
	if input.UserID <= 0 || input.ProductID <= 0 {
		return nil, false, infraerrors.BadRequest("INVALID_PRODUCT_SUBSCRIPTION_ASSIGNMENT", "user_id and product_id are required")
	}

	client := productAdminClientFromContext(ctx, s.entClient)
	product, err := client.SubscriptionProduct.Query().
		Where(subscriptionproduct.IDEQ(input.ProductID)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, false, NewSubscriptionProductNotFoundError(0).WithMetadata(map[string]string{
				"product_id": fmt.Sprintf("%d", input.ProductID),
			})
		}
		return nil, false, fmt.Errorf("get subscription product: %w", err)
	}

	validityDays := normalizeProductValidityDays(input.ValidityDays, product.DefaultValidityDays)
	now := time.Now()

	existing, err := client.UserProductSubscription.Query().
		Where(
			userproductsubscription.UserIDEQ(input.UserID),
			userproductsubscription.ProductIDEQ(input.ProductID),
		).
		Order(dbent.Desc(userproductsubscription.FieldExpiresAt), dbent.Desc(userproductsubscription.FieldID)).
		First(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return nil, false, fmt.Errorf("get user product subscription: %w", err)
	}

	if existing != nil {
		newExpiresAt := now.AddDate(0, 0, validityDays)
		if existing.ExpiresAt.After(now) {
			newExpiresAt = existing.ExpiresAt.AddDate(0, 0, validityDays)
		}
		if newExpiresAt.After(MaxExpiresAt) {
			newExpiresAt = MaxExpiresAt
		}

		update := client.UserProductSubscription.UpdateOneID(existing.ID).
			SetExpiresAt(newExpiresAt).
			SetStatus(SubscriptionStatusActive).
			SetAssignedAt(now).
			SetNotes(appendProductSubscriptionNote(derefProductString(existing.Notes), input.Notes))
		if input.AssignedBy > 0 {
			update.SetAssignedBy(input.AssignedBy)
		}
		updated, err := update.Save(ctx)
		if err != nil {
			return nil, false, fmt.Errorf("extend user product subscription: %w", err)
		}
		return userProductSubscriptionEntityToService(updated), true, nil
	}

	expiresAt := now.AddDate(0, 0, validityDays)
	if expiresAt.After(MaxExpiresAt) {
		expiresAt = MaxExpiresAt
	}
	create := client.UserProductSubscription.Create().
		SetUserID(input.UserID).
		SetProductID(input.ProductID).
		SetStartsAt(now).
		SetExpiresAt(expiresAt).
		SetStatus(SubscriptionStatusActive).
		SetAssignedAt(now).
		SetNotes(input.Notes)
	if input.AssignedBy > 0 {
		create.SetAssignedBy(input.AssignedBy)
	}
	created, err := create.Save(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("create user product subscription: %w", err)
	}
	return userProductSubscriptionEntityToService(created), false, nil
}

func productAdminClientFromContext(ctx context.Context, fallback *dbent.Client) *dbent.Client {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	if client := dbent.FromContext(ctx); client != nil {
		return client
	}
	return fallback
}

func normalizeProductValidityDays(days, fallback int) int {
	if days <= 0 {
		days = fallback
	}
	if days <= 0 {
		days = 30
	}
	if days > MaxValidityDays {
		return MaxValidityDays
	}
	return days
}

func appendProductSubscriptionNote(existing, next string) string {
	existing = strings.TrimSpace(existing)
	next = strings.TrimSpace(next)
	if existing == "" {
		return next
	}
	if next == "" {
		return existing
	}
	return existing + "\n" + next
}

func userProductSubscriptionEntityToService(m *dbent.UserProductSubscription) *UserProductSubscription {
	if m == nil {
		return nil
	}
	sub := &UserProductSubscription{
		ID:                         m.ID,
		UserID:                     m.UserID,
		ProductID:                  m.ProductID,
		StartsAt:                   m.StartsAt,
		ExpiresAt:                  m.ExpiresAt,
		Status:                     m.Status,
		DailyWindowStart:           m.DailyWindowStart,
		WeeklyWindowStart:          m.WeeklyWindowStart,
		MonthlyWindowStart:         m.MonthlyWindowStart,
		DailyUsageUSD:              m.DailyUsageUsd,
		WeeklyUsageUSD:             m.WeeklyUsageUsd,
		MonthlyUsageUSD:            m.MonthlyUsageUsd,
		DailyCarryoverInUSD:        m.DailyCarryoverInUsd,
		DailyCarryoverRemainingUSD: m.DailyCarryoverRemainingUsd,
		AssignedBy:                 m.AssignedBy,
		AssignedAt:                 m.AssignedAt,
		Notes:                      derefProductString(m.Notes),
		CreatedAt:                  m.CreatedAt,
		UpdatedAt:                  m.UpdatedAt,
	}
	if m.Edges.Product != nil {
		sub.Product = subscriptionProductEntityToService(m.Edges.Product)
	}
	return sub
}

func subscriptionProductEntityToService(m *dbent.SubscriptionProduct) *SubscriptionProduct {
	if m == nil {
		return nil
	}
	return &SubscriptionProduct{
		ID:                  m.ID,
		Code:                m.Code,
		Name:                m.Name,
		Description:         derefProductString(m.Description),
		Status:              m.Status,
		DefaultValidityDays: m.DefaultValidityDays,
		DailyLimitUSD:       m.DailyLimitUsd,
		WeeklyLimitUSD:      m.WeeklyLimitUsd,
		MonthlyLimitUSD:     m.MonthlyLimitUsd,
		SortOrder:           m.SortOrder,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func derefProductString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
