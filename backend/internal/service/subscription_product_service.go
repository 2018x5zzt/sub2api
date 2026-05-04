package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var beijingLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type SubscriptionProductService struct {
	repo ProductSubscriptionRepository
}

func NewSubscriptionProductService(repo ProductSubscriptionRepository) *SubscriptionProductService {
	return &SubscriptionProductService{repo: repo}
}

func (s *SubscriptionProductService) GetActiveProductSubscription(ctx context.Context, userID, groupID int64) (*ProductSettlementContext, error) {
	return s.GetActiveProductSubscriptionForFamily(ctx, userID, groupID, nil)
}

func (s *SubscriptionProductService) GetActiveProductSubscriptionForFamily(ctx context.Context, userID, groupID int64, productFamily *string) (*ProductSettlementContext, error) {
	if s == nil || s.repo == nil {
		return nil, ErrSubscriptionNotFound
	}
	binding, sub, err := s.repo.GetActiveProductSubscriptionByUserAndGroupID(ctx, userID, groupID, productFamily)
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

func (s *SubscriptionProductService) ListVisibleProductFamilies(ctx context.Context, userID, groupID int64) ([]string, error) {
	products, err := s.ListActiveUserProducts(ctx, userID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	families := make([]string, 0)
	for _, item := range products {
		matchesGroup := false
		for _, group := range item.Groups {
			if group.GroupID == groupID {
				matchesGroup = true
				break
			}
		}
		if !matchesGroup {
			continue
		}
		family := strings.TrimSpace(item.Product.ProductFamily)
		if family == "" {
			family = "gpt"
		}
		if _, ok := seen[family]; ok {
			continue
		}
		seen[family] = struct{}{}
		families = append(families, family)
	}
	return families, nil
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
	input.ProductFamily = "gpt"
	return s.repo.CreateProduct(ctx, input)
}

func (s *SubscriptionProductService) UpdateProduct(ctx context.Context, productID int64, input *UpdateSubscriptionProductInput) (*SubscriptionProduct, error) {
	if s == nil || s.repo == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if input == nil {
		return nil, ErrSubscriptionNilInput
	}
	family := "gpt"
	input.ProductFamily = &family
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

func (s *SubscriptionProductService) AdjustProductSubscription(ctx context.Context, subscriptionID int64, input *AdjustProductSubscriptionInput) (*UserProductSubscription, error) {
	if s == nil || s.repo == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if subscriptionID <= 0 || input == nil {
		return nil, ErrSubscriptionNilInput
	}
	if input.Status != nil {
		switch *input.Status {
		case SubscriptionStatusActive, SubscriptionStatusExpired, ProductSubscriptionStatusRevoked:
		default:
			return nil, ErrSubscriptionInvalid
		}
	}
	return s.repo.AdjustProductSubscription(ctx, subscriptionID, input)
}

func (s *SubscriptionProductService) ResetProductSubscriptionQuota(ctx context.Context, subscriptionID int64, input *ResetProductSubscriptionQuotaInput) (*UserProductSubscription, error) {
	if s == nil || s.repo == nil {
		return nil, ErrProductSubscriptionAssignerUnavailable
	}
	if subscriptionID <= 0 || input == nil || (!input.Daily && !input.Weekly && !input.Monthly) {
		return nil, ErrSubscriptionNilInput
	}
	return s.repo.ResetProductSubscriptionQuota(ctx, subscriptionID, input)
}

func (s *SubscriptionProductService) RevokeProductSubscription(ctx context.Context, subscriptionID int64) error {
	if s == nil || s.repo == nil {
		return ErrProductSubscriptionAssignerUnavailable
	}
	if subscriptionID <= 0 {
		return ErrSubscriptionNilInput
	}
	return s.repo.RevokeProductSubscription(ctx, subscriptionID)
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
	windowStart := beijingStartOfDay(now)
	if productSubscriptionDailyWindowExpired(sub.DailyWindowStart, now) {
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
	if productSubscriptionRollingWindowExpired(sub.WeeklyWindowStart, now, 7*24*time.Hour) {
		weeklyStart := productSubscriptionNextRollingWindowStart(sub.WeeklyWindowStart, now, 7*24*time.Hour)
		sub.WeeklyWindowStart = &weeklyStart
		sub.WeeklyUsageUSD = 0
	}
	if productSubscriptionMonthlyWindowExpired(sub.MonthlyWindowStart, now) {
		monthlyStart := productSubscriptionNextRollingWindowStart(sub.MonthlyWindowStart, now, 30*24*time.Hour)
		sub.MonthlyWindowStart = &monthlyStart
		sub.MonthlyUsageUSD = 0
	}
}

func productSubscriptionDailyWindowExpired(windowStart *time.Time, now time.Time) bool {
	if windowStart == nil {
		return false
	}
	return !beijingStartOfDay(now).Equal(beijingStartOfDay(*windowStart))
}

func productSubscriptionMonthlyWindowExpired(windowStart *time.Time, now time.Time) bool {
	return productSubscriptionRollingWindowExpired(windowStart, now, 30*24*time.Hour)
}

func productSubscriptionRollingWindowExpired(windowStart *time.Time, now time.Time, window time.Duration) bool {
	return windowStart != nil && !now.Before(windowStart.Add(window))
}

func productSubscriptionNextRollingWindowStart(windowStart *time.Time, now time.Time, window time.Duration) time.Time {
	if windowStart == nil {
		return now
	}
	if now.Before(*windowStart) {
		return *windowStart
	}
	next := *windowStart
	for !now.Before(next.Add(window)) {
		next = next.Add(window)
	}
	return next
}

func beijingStartOfDay(t time.Time) time.Time {
	inBeijing := t.In(beijingLocation)
	return time.Date(inBeijing.Year(), inBeijing.Month(), inBeijing.Day(), 0, 0, 0, 0, beijingLocation)
}

func NormalizeExpiredProductSubscriptionWindowForRepository(sub *UserProductSubscription, product *SubscriptionProduct, now time.Time) {
	normalizeExpiredProductSubscriptionWindow(sub, product, now)
}

func IsProductSubscriptionNotFound(err error) bool {
	return errors.Is(err, ErrSubscriptionNotFound)
}
