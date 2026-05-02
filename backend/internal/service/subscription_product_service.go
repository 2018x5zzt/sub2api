package service

import (
	"context"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type SubscriptionProductService struct {
	repo ProductSubscriptionRepository
}

func NewSubscriptionProductService(repo ProductSubscriptionRepository) *SubscriptionProductService {
	return &SubscriptionProductService{repo: repo}
}

func (s *SubscriptionProductService) GetActiveProductSubscription(ctx context.Context, userID, groupID int64) (*ProductSettlementContext, error) {
	if s == nil || s.repo == nil {
		return nil, ErrSubscriptionNotFound
	}
	binding, sub, err := s.repo.GetActiveProductSubscriptionByUserAndGroupID(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if binding == nil || sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	if binding.ProductStatus != SubscriptionProductStatusActive || binding.BindingStatus != SubscriptionProductBindingStatusActive {
		return nil, ErrSubscriptionSuspended
	}
	if !sub.IsActive() {
		return nil, ErrSubscriptionExpired
	}
	return &ProductSettlementContext{Binding: binding, Subscription: sub}, nil
}

func (s *SubscriptionProductService) ListActiveUserProducts(ctx context.Context, userID int64) ([]ActiveSubscriptionProduct, error) {
	if s == nil || s.repo == nil {
		return []ActiveSubscriptionProduct{}, nil
	}
	return s.repo.ListActiveProductsByUserID(ctx, userID)
}

func (s *SubscriptionProductService) ListVisibleGroups(ctx context.Context, userID int64) ([]Group, error) {
	if s == nil || s.repo == nil {
		return []Group{}, nil
	}
	return s.repo.ListVisibleGroupsByUserID(ctx, userID)
}

func (s *SubscriptionProductService) ListProducts(ctx context.Context) ([]SubscriptionProduct, error) {
	if s == nil || s.repo == nil {
		return []SubscriptionProduct{}, nil
	}
	return s.repo.ListProducts(ctx)
}

func (s *SubscriptionProductService) ResolveActiveProductByGroupID(ctx context.Context, groupID int64) (*SubscriptionProduct, error) {
	if s == nil || s.repo == nil {
		return nil, ErrSubscriptionNotFound
	}
	if groupID <= 0 {
		return nil, ErrSubscriptionNotFound
	}
	return s.repo.ResolveActiveProductByGroupID(ctx, groupID)
}

func (s *SubscriptionProductService) CreateProduct(ctx context.Context, input *CreateSubscriptionProductInput) (*SubscriptionProduct, error) {
	if s == nil || s.repo == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if input == nil {
		return nil, ErrSubscriptionNilInput
	}
	return s.repo.CreateProduct(ctx, input)
}

func (s *SubscriptionProductService) UpdateProduct(ctx context.Context, productID int64, input *UpdateSubscriptionProductInput) (*SubscriptionProduct, error) {
	if s == nil || s.repo == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if input == nil {
		return nil, ErrSubscriptionNilInput
	}
	return s.repo.UpdateProduct(ctx, productID, input)
}

func (s *SubscriptionProductService) SyncProductBindings(ctx context.Context, productID int64, inputs []SubscriptionProductBindingInput) ([]SubscriptionProductBindingDetail, error) {
	if s == nil || s.repo == nil {
		return []SubscriptionProductBindingDetail{}, nil
	}
	return s.repo.SyncProductBindings(ctx, productID, inputs)
}

func (s *SubscriptionProductService) ListProductBindings(ctx context.Context, productID int64) ([]SubscriptionProductBindingDetail, error) {
	if s == nil || s.repo == nil {
		return []SubscriptionProductBindingDetail{}, nil
	}
	return s.repo.ListProductBindings(ctx, productID)
}

func (s *SubscriptionProductService) ListProductSubscriptions(ctx context.Context, productID int64) ([]UserProductSubscription, error) {
	if s == nil || s.repo == nil {
		return []UserProductSubscription{}, nil
	}
	return s.repo.ListProductSubscriptions(ctx, productID)
}

func (s *SubscriptionProductService) ListUserProductSubscriptionsForAdmin(ctx context.Context, params AdminProductSubscriptionListParams) ([]AdminProductSubscriptionListItem, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return []AdminProductSubscriptionListItem{}, &pagination.PaginationResult{Total: 0, Page: 1, PageSize: 20, Pages: 1}, nil
	}
	return s.repo.ListUserProductSubscriptionsForAdmin(ctx, params)
}

func (s *SubscriptionProductService) GetUserProductSummary(ctx context.Context, userID int64) (*SubscriptionProductSummary, error) {
	products, err := s.ListActiveUserProducts(ctx, userID)
	if err != nil {
		return nil, err
	}
	summary := &SubscriptionProductSummary{
		ActiveCount: len(products),
		Products:    products,
	}
	for _, item := range products {
		summary.TotalMonthlyUsageUSD += item.Subscription.MonthlyUsageUSD
		summary.TotalMonthlyLimitUSD += item.Product.MonthlyLimitUSD
	}
	return summary, nil
}

func (s *SubscriptionProductService) GetUserProductProgress(ctx context.Context, userID int64) (*SubscriptionProductSummary, error) {
	return s.GetUserProductSummary(ctx, userID)
}

func (s *SubscriptionProductService) AssignOrExtendProductSubscription(ctx context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error) {
	if s == nil || s.repo == nil {
		return nil, false, ErrProductSubscriptionAssignerUnavailable
	}
	if input == nil {
		return nil, false, ErrSubscriptionNilInput
	}
	return s.repo.AssignOrExtendProductSubscription(ctx, input)
}

func (s *SubscriptionProductService) CheckProductLimits(settlement *ProductSettlementContext, additionalDebitCost float64) error {
	if settlement == nil || settlement.Binding == nil || settlement.Subscription == nil {
		return ErrSubscriptionNotFound
	}
	product := settlement.Binding.Product()
	sub := settlement.Subscription
	normalizeExpiredProductSubscriptionWindow(sub, product, time.Now())
	if !sub.CheckDailyLimit(product, additionalDebitCost) {
		return ErrDailyLimitExceeded
	}
	if !sub.CheckWeeklyLimit(product, additionalDebitCost) {
		return ErrWeeklyLimitExceeded
	}
	if !sub.CheckMonthlyLimit(product, additionalDebitCost) {
		return ErrMonthlyLimitExceeded
	}
	return nil
}

func normalizeExpiredProductSubscriptionWindow(sub *UserProductSubscription, product *SubscriptionProduct, now time.Time) {
	if sub == nil {
		return
	}
	if sub.NeedsDailyReset() {
		windowStart := startOfDay(now)
		if product != nil && product.HasDailyLimit() {
			carryoverConsumed := maxFloat64(sub.DailyCarryoverInUSD-sub.DailyCarryoverRemainingUSD, 0)
			freshUsed := maxFloat64(sub.DailyUsageUSD-carryoverConsumed, 0)
			freshRemaining := maxFloat64(product.DailyLimitUSD-freshUsed, 0)
			if sub.DailyWindowStart != nil &&
				elapsedDailyWindows(*sub.DailyWindowStart, windowStart) == 1 &&
				sub.Status == SubscriptionStatusActive &&
				sub.ExpiresAt.After(windowStart) {
				sub.DailyCarryoverInUSD = freshRemaining
				sub.DailyCarryoverRemainingUSD = freshRemaining
			} else {
				sub.DailyCarryoverInUSD = 0
				sub.DailyCarryoverRemainingUSD = 0
			}
			sub.DailyWindowStart = &windowStart
			sub.DailyUsageUSD = 0
		} else {
			sub.DailyWindowStart = nil
			sub.DailyUsageUSD = 0
			sub.DailyCarryoverInUSD = 0
			sub.DailyCarryoverRemainingUSD = 0
		}
	}
	if sub.NeedsWeeklyReset() {
		windowStart := startOfDay(now)
		sub.WeeklyWindowStart = &windowStart
		sub.WeeklyUsageUSD = 0
	}
	if productSubscriptionMonthlyWindowExpired(sub.MonthlyWindowStart, now) {
		windowStart := startOfMonth(now)
		sub.MonthlyWindowStart = &windowStart
		sub.MonthlyUsageUSD = 0
	}
}

func productSubscriptionMonthlyWindowExpired(windowStart *time.Time, now time.Time) bool {
	if windowStart == nil {
		return false
	}
	return startOfMonth(now).After(startOfMonth(*windowStart))
}

func NormalizeExpiredProductSubscriptionWindowForRepository(sub *UserProductSubscription, product *SubscriptionProduct, now time.Time) {
	normalizeExpiredProductSubscriptionWindow(sub, product, now)
}

func IsProductSubscriptionNotFound(err error) bool {
	return errors.Is(err, ErrSubscriptionNotFound)
}
