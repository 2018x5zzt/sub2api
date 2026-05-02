package service

import (
	"context"
	"errors"
)

type ProductAwareSubscriptionAssigner struct {
	legacy  DefaultSubscriptionAssigner
	product *SubscriptionProductService
}

func NewProductAwareSubscriptionAssigner(legacy DefaultSubscriptionAssigner, product *SubscriptionProductService) *ProductAwareSubscriptionAssigner {
	return &ProductAwareSubscriptionAssigner{legacy: legacy, product: product}
}

func (a *ProductAwareSubscriptionAssigner) AssignOrExtendSubscription(ctx context.Context, input *AssignSubscriptionInput) (*UserSubscription, bool, error) {
	if input == nil {
		return nil, false, ErrSubscriptionNilInput
	}
	if a != nil && a.product != nil {
		productID, ok, err := a.ResolveActiveProductIDByGroupID(ctx, input.GroupID)
		if err == nil && ok {
			productSub, reused, assignErr := a.product.AssignOrExtendProductSubscription(ctx, &AssignProductSubscriptionInput{
				UserID:       input.UserID,
				ProductID:    productID,
				ValidityDays: input.ValidityDays,
				AssignedBy:   input.AssignedBy,
				Notes:        input.Notes,
			})
			if assignErr != nil {
				return nil, false, assignErr
			}
			return userSubscriptionFromProductAssignment(productSub, input.GroupID), reused, nil
		}
		if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
			return nil, false, err
		}
	}
	if a == nil || a.legacy == nil {
		return nil, false, ErrProductSubscriptionAssignerUnavailable
	}
	return a.legacy.AssignOrExtendSubscription(ctx, input)
}

func (a *ProductAwareSubscriptionAssigner) ResolveActiveProductIDByGroupID(ctx context.Context, groupID int64) (int64, bool, error) {
	if a == nil || a.product == nil {
		return 0, false, nil
	}
	product, err := a.product.ResolveActiveProductByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	if product == nil {
		return 0, false, nil
	}
	return product.ID, true, nil
}

func (a *ProductAwareSubscriptionAssigner) ListProducts(ctx context.Context) ([]SubscriptionProduct, error) {
	if a == nil || a.product == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	return a.product.ListProducts(ctx)
}

func userSubscriptionFromProductAssignment(productSub *UserProductSubscription, groupID int64) *UserSubscription {
	if productSub == nil {
		return nil
	}
	return &UserSubscription{
		ID:                 productSub.ID,
		UserID:             productSub.UserID,
		GroupID:            groupID,
		StartsAt:           productSub.StartsAt,
		ExpiresAt:          productSub.ExpiresAt,
		Status:             productSub.Status,
		DailyWindowStart:   productSub.DailyWindowStart,
		WeeklyWindowStart:  productSub.WeeklyWindowStart,
		MonthlyWindowStart: productSub.MonthlyWindowStart,
		DailyUsageUSD:      productSub.DailyUsageUSD,
		WeeklyUsageUSD:     productSub.WeeklyUsageUSD,
		MonthlyUsageUSD:    productSub.MonthlyUsageUSD,
		AssignedBy:         productSub.AssignedBy,
		AssignedAt:         productSub.AssignedAt,
		Notes:              productSub.Notes,
		CreatedAt:          productSub.CreatedAt,
		UpdatedAt:          productSub.UpdatedAt,
	}
}
