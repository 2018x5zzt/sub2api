package service

import (
	"context"
	"errors"
	"time"
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
	if sub.NeedsMonthlyReset() {
		windowStart := startOfDay(now)
		sub.MonthlyWindowStart = &windowStart
		sub.MonthlyUsageUSD = 0
	}
}

func IsProductSubscriptionNotFound(err error) bool {
	return errors.Is(err, ErrSubscriptionNotFound)
}
