package service

import "time"

type UserProductSubscription struct {
	ID        int64
	UserID    int64
	ProductID int64

	StartsAt  time.Time
	ExpiresAt time.Time
	Status    string

	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time

	DailyUsageUSD   float64
	WeeklyUsageUSD  float64
	MonthlyUsageUSD float64

	DailyCarryoverInUSD        float64
	DailyCarryoverRemainingUSD float64

	AssignedBy *int64
	AssignedAt time.Time
	Notes      string

	CreatedAt time.Time
	UpdatedAt time.Time

	Product *SubscriptionProduct
}

func (s *UserProductSubscription) IsActive() bool {
	return s != nil && s.Status == SubscriptionStatusActive && time.Now().Before(s.ExpiresAt)
}

func (s *UserProductSubscription) IsExpired() bool {
	return s == nil || time.Now().After(s.ExpiresAt)
}

func (s *UserProductSubscription) DailyEffectiveLimit(product *SubscriptionProduct) float64 {
	if product == nil || !product.HasDailyLimit() {
		return 0
	}
	return product.DailyLimitUSD + maxFloat64(s.DailyCarryoverInUSD, 0)
}

func (s *UserProductSubscription) DailyRemainingTotal(product *SubscriptionProduct) float64 {
	if s == nil {
		return 0
	}
	remaining := s.DailyEffectiveLimit(product) - s.DailyUsageUSD
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *UserProductSubscription) DailyRemainingCarryover() float64 {
	if s == nil || s.DailyCarryoverRemainingUSD < 0 {
		return 0
	}
	return s.DailyCarryoverRemainingUSD
}

func (s *UserProductSubscription) CheckDailyLimit(product *SubscriptionProduct, additionalCost float64) bool {
	if product == nil || !product.HasDailyLimit() {
		return true
	}
	return s != nil && s.DailyUsageUSD+additionalCost <= s.DailyEffectiveLimit(product)
}

func (s *UserProductSubscription) CheckWeeklyLimit(product *SubscriptionProduct, additionalCost float64) bool {
	if product == nil || !product.HasWeeklyLimit() {
		return true
	}
	return s != nil && s.WeeklyUsageUSD+additionalCost <= product.WeeklyLimitUSD
}

func (s *UserProductSubscription) CheckMonthlyLimit(product *SubscriptionProduct, additionalCost float64) bool {
	if product == nil || !product.HasMonthlyLimit() {
		return true
	}
	return s != nil && s.MonthlyUsageUSD+additionalCost <= product.MonthlyLimitUSD
}

func (s *UserProductSubscription) CheckAllLimits(product *SubscriptionProduct, additionalCost float64) (daily, weekly, monthly bool) {
	daily = s.CheckDailyLimit(product, additionalCost)
	weekly = s.CheckWeeklyLimit(product, additionalCost)
	monthly = s.CheckMonthlyLimit(product, additionalCost)
	return
}
