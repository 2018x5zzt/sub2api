package service

import "context"

type SubscriptionProductService struct {
	repo ProductSubscriptionRepository
}

func NewSubscriptionProductService(repo ProductSubscriptionRepository) *SubscriptionProductService {
	return &SubscriptionProductService{repo: repo}
}

func (s *SubscriptionProductService) GetActiveProductSubscription(ctx context.Context, userID, groupID int64) (*ProductSettlementContext, error) {
	if repo, ok := s.repo.(ProductSubscriptionByUserGroupRepository); ok {
		binding, sub, err := repo.GetActiveProductSubscriptionByUserAndGroupID(ctx, userID, groupID)
		if err != nil {
			return nil, err
		}
		if binding != nil || sub != nil {
			return buildProductSettlementContext(binding, sub, groupID)
		}
	}

	binding, err := s.repo.GetActiveProductBindingByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	sub, err := s.repo.GetActiveUserProductSubscription(ctx, userID, binding.ProductID)
	if err != nil {
		return nil, NewProductSubscriptionInvalidError(binding, nil).WithCause(err)
	}

	return buildProductSettlementContext(binding, sub, groupID)
}

func buildProductSettlementContext(binding *SubscriptionProductBinding, sub *UserProductSubscription, groupID int64) (*ProductSettlementContext, error) {
	if binding == nil {
		return nil, NewSubscriptionProductNotFoundError(groupID)
	}
	if binding.ProductStatus != SubscriptionProductStatusActive {
		return nil, NewSubscriptionProductInactiveError(binding)
	}
	if binding.BindingStatus != SubscriptionProductBindingStatusActive {
		return nil, NewProductBindingInactiveError(binding)
	}

	if sub == nil || !sub.IsActive() {
		return nil, NewProductSubscriptionInvalidError(binding, sub)
	}

	return &ProductSettlementContext{Binding: binding, Subscription: sub}, nil
}

func (s *SubscriptionProductService) ListVisibleGroups(ctx context.Context, userID int64) ([]Group, error) {
	return s.repo.ListVisibleGroupsByUserID(ctx, userID)
}

func (s *SubscriptionProductService) ListActiveUserProducts(ctx context.Context, userID int64) ([]ActiveSubscriptionProduct, error) {
	return s.repo.ListActiveProductsByUserID(ctx, userID)
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

func (s *SubscriptionProductService) CheckProductLimits(binding *SubscriptionProductBinding, sub *UserProductSubscription, additionalDebitCost float64) error {
	product := binding.Product()
	daily, weekly, monthly := sub.CheckAllLimits(product, additionalDebitCost)
	if daily && weekly && monthly {
		return nil
	}

	remaining := 0.0
	if product != nil && product.HasDailyLimit() {
		remaining = sub.DailyRemainingTotal(product)
	}
	return NewProductLimitExceededError(binding, remaining)
}
