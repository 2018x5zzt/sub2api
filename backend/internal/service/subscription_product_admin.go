package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionproduct"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionproductgroup"
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

func (s *SubscriptionProductAdminService) ListProducts(ctx context.Context) ([]SubscriptionProduct, error) {
	if s == nil || s.entClient == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	client := productAdminClientFromContext(ctx, s.entClient)
	products, err := client.SubscriptionProduct.Query().
		Order(dbent.Asc(subscriptionproduct.FieldSortOrder), dbent.Asc(subscriptionproduct.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subscription products: %w", err)
	}
	out := make([]SubscriptionProduct, 0, len(products))
	for _, product := range products {
		out = append(out, *subscriptionProductEntityToService(product))
	}
	return out, nil
}

func (s *SubscriptionProductAdminService) CreateProduct(ctx context.Context, input *CreateSubscriptionProductInput) (*SubscriptionProduct, error) {
	if s == nil || s.entClient == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if input == nil {
		return nil, ErrSubscriptionNilInput
	}
	code := strings.TrimSpace(input.Code)
	name := strings.TrimSpace(input.Name)
	if code == "" || name == "" {
		return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "code and name are required")
	}
	if err := validateProductLimits(input.DailyLimitUSD, input.WeeklyLimitUSD, input.MonthlyLimitUSD); err != nil {
		return nil, err
	}
	status, err := normalizeSubscriptionProductStatus(input.Status)
	if err != nil {
		return nil, err
	}

	client := productAdminClientFromContext(ctx, s.entClient)
	created, err := client.SubscriptionProduct.Create().
		SetCode(code).
		SetName(name).
		SetDescription(strings.TrimSpace(input.Description)).
		SetStatus(status).
		SetDefaultValidityDays(normalizeProductValidityDays(input.DefaultValidityDays, 30)).
		SetDailyLimitUsd(input.DailyLimitUSD).
		SetWeeklyLimitUsd(input.WeeklyLimitUSD).
		SetMonthlyLimitUsd(input.MonthlyLimitUSD).
		SetSortOrder(input.SortOrder).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create subscription product: %w", err)
	}
	return subscriptionProductEntityToService(created), nil
}

func (s *SubscriptionProductAdminService) UpdateProduct(ctx context.Context, productID int64, input *UpdateSubscriptionProductInput) (*SubscriptionProduct, error) {
	if s == nil || s.entClient == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if productID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "product_id is required")
	}
	if input == nil {
		return nil, ErrSubscriptionNilInput
	}
	client := productAdminClientFromContext(ctx, s.entClient)
	update := client.SubscriptionProduct.UpdateOneID(productID)
	if input.Code != nil {
		code := strings.TrimSpace(*input.Code)
		if code == "" {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "code cannot be empty")
		}
		update.SetCode(code)
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "name cannot be empty")
		}
		update.SetName(name)
	}
	if input.Description != nil {
		update.SetDescription(strings.TrimSpace(*input.Description))
	}
	if input.Status != nil {
		status, err := normalizeSubscriptionProductStatus(*input.Status)
		if err != nil {
			return nil, err
		}
		update.SetStatus(status)
	}
	if input.DefaultValidityDays != nil {
		if *input.DefaultValidityDays <= 0 || *input.DefaultValidityDays > MaxValidityDays {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "default_validity_days is out of range")
		}
		update.SetDefaultValidityDays(*input.DefaultValidityDays)
	}
	if input.DailyLimitUSD != nil {
		if *input.DailyLimitUSD < 0 {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "daily_limit_usd cannot be negative")
		}
		update.SetDailyLimitUsd(*input.DailyLimitUSD)
	}
	if input.WeeklyLimitUSD != nil {
		if *input.WeeklyLimitUSD < 0 {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "weekly_limit_usd cannot be negative")
		}
		update.SetWeeklyLimitUsd(*input.WeeklyLimitUSD)
	}
	if input.MonthlyLimitUSD != nil {
		if *input.MonthlyLimitUSD < 0 {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "monthly_limit_usd cannot be negative")
		}
		update.SetMonthlyLimitUsd(*input.MonthlyLimitUSD)
	}
	if input.SortOrder != nil {
		update.SetSortOrder(*input.SortOrder)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, NewSubscriptionProductNotFoundError(0).WithMetadata(map[string]string{
				"product_id": fmt.Sprintf("%d", productID),
			})
		}
		return nil, fmt.Errorf("update subscription product: %w", err)
	}
	return subscriptionProductEntityToService(updated), nil
}

func (s *SubscriptionProductAdminService) SyncProductBindings(ctx context.Context, productID int64, inputs []SubscriptionProductBindingInput) ([]SubscriptionProductBindingDetail, error) {
	if s == nil || s.entClient == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if productID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDINGS", "product_id is required")
	}

	client := productAdminClientFromContext(ctx, s.entClient)
	if _, err := client.SubscriptionProduct.Get(ctx, productID); err != nil {
		if dbent.IsNotFound(err) {
			return nil, NewSubscriptionProductNotFoundError(0).WithMetadata(map[string]string{
				"product_id": fmt.Sprintf("%d", productID),
			})
		}
		return nil, fmt.Errorf("get subscription product: %w", err)
	}

	existing, err := client.SubscriptionProductGroup.Query().
		Where(subscriptionproductgroup.ProductIDEQ(productID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subscription product bindings: %w", err)
	}
	existingByGroupID := make(map[int64]*dbent.SubscriptionProductGroup, len(existing))
	for _, binding := range existing {
		existingByGroupID[binding.GroupID] = binding
	}

	seen := make(map[int64]struct{}, len(inputs))
	for _, input := range inputs {
		normalized, err := normalizeProductBindingInput(input)
		if err != nil {
			return nil, err
		}
		seen[normalized.GroupID] = struct{}{}
		if binding := existingByGroupID[normalized.GroupID]; binding != nil {
			if _, err := client.SubscriptionProductGroup.UpdateOneID(binding.ID).
				SetDebitMultiplier(normalized.DebitMultiplier).
				SetStatus(normalized.Status).
				SetSortOrder(normalized.SortOrder).
				Save(ctx); err != nil {
				return nil, fmt.Errorf("update subscription product binding: %w", err)
			}
			continue
		}
		if _, err := client.SubscriptionProductGroup.Create().
			SetProductID(productID).
			SetGroupID(normalized.GroupID).
			SetDebitMultiplier(normalized.DebitMultiplier).
			SetStatus(normalized.Status).
			SetSortOrder(normalized.SortOrder).
			Save(ctx); err != nil {
			return nil, fmt.Errorf("create subscription product binding: %w", err)
		}
	}

	for _, binding := range existing {
		if _, ok := seen[binding.GroupID]; ok {
			continue
		}
		if binding.Status == SubscriptionProductBindingStatusInactive {
			continue
		}
		if _, err := client.SubscriptionProductGroup.UpdateOneID(binding.ID).
			SetStatus(SubscriptionProductBindingStatusInactive).
			Save(ctx); err != nil {
			return nil, fmt.Errorf("deactivate subscription product binding: %w", err)
		}
	}

	return s.listProductBindingDetails(ctx, client, productID)
}

func (s *SubscriptionProductAdminService) ListProductSubscriptions(ctx context.Context, productID int64) ([]UserProductSubscription, error) {
	if s == nil || s.entClient == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if productID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_PRODUCT_SUBSCRIPTION_LIST", "product_id is required")
	}
	client := productAdminClientFromContext(ctx, s.entClient)
	subs, err := client.UserProductSubscription.Query().
		Where(userproductsubscription.ProductIDEQ(productID)).
		WithProduct().
		Order(dbent.Desc(userproductsubscription.FieldExpiresAt), dbent.Desc(userproductsubscription.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list user product subscriptions: %w", err)
	}
	out := make([]UserProductSubscription, 0, len(subs))
	for _, sub := range subs {
		out = append(out, *userProductSubscriptionEntityToService(sub))
	}
	return out, nil
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

func (s *SubscriptionProductAdminService) listProductBindingDetails(ctx context.Context, client *dbent.Client, productID int64) ([]SubscriptionProductBindingDetail, error) {
	bindings, err := client.SubscriptionProductGroup.Query().
		Where(subscriptionproductgroup.ProductIDEQ(productID)).
		Order(dbent.Asc(subscriptionproductgroup.FieldSortOrder), dbent.Asc(subscriptionproductgroup.FieldGroupID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subscription product binding details: %w", err)
	}
	out := make([]SubscriptionProductBindingDetail, 0, len(bindings))
	for _, binding := range bindings {
		detail := SubscriptionProductBindingDetail{
			ProductID:       binding.ProductID,
			GroupID:         binding.GroupID,
			DebitMultiplier: binding.DebitMultiplier,
			Status:          binding.Status,
			SortOrder:       binding.SortOrder,
			CreatedAt:       binding.CreatedAt,
			UpdatedAt:       binding.UpdatedAt,
		}
		if group, err := client.Group.Get(ctx, binding.GroupID); err == nil {
			detail.GroupName = group.Name
		} else if !dbent.IsNotFound(err) {
			return nil, fmt.Errorf("get group for product binding: %w", err)
		}
		out = append(out, detail)
	}
	return out, nil
}

func normalizeSubscriptionProductStatus(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return SubscriptionProductStatusDraft, nil
	}
	switch status {
	case SubscriptionProductStatusDraft, SubscriptionProductStatusActive, SubscriptionProductStatusDisabled:
		return status, nil
	default:
		return "", infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_STATUS", "invalid subscription product status")
	}
}

func normalizeProductBindingInput(input SubscriptionProductBindingInput) (SubscriptionProductBindingInput, error) {
	if input.GroupID <= 0 {
		return input, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDINGS", "group_id is required")
	}
	if input.DebitMultiplier <= 0 {
		input.DebitMultiplier = 1
	}
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = SubscriptionProductBindingStatusActive
	}
	switch input.Status {
	case SubscriptionProductBindingStatusActive, SubscriptionProductBindingStatusInactive:
		return input, nil
	default:
		return input, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDING_STATUS", "invalid subscription product binding status")
	}
}

func validateProductLimits(daily, weekly, monthly float64) error {
	if daily < 0 {
		return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "daily_limit_usd cannot be negative")
	}
	if weekly < 0 {
		return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "weekly_limit_usd cannot be negative")
	}
	if monthly < 0 {
		return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "monthly_limit_usd cannot be negative")
	}
	return nil
}

func derefProductString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
