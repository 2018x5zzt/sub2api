package service

import (
	"context"
	"time"
)

const (
	SubscriptionProductStatusActive        = "active"
	SubscriptionProductBindingStatusActive = "active"
)

type SubscriptionProduct struct {
	ID                  int64
	Code                string
	Name                string
	Status              string
	DefaultValidityDays int
	DailyLimitUSD       float64
	WeeklyLimitUSD      float64
	MonthlyLimitUSD     float64
}

func (p *SubscriptionProduct) HasDailyLimit() bool {
	return p != nil && p.DailyLimitUSD > 0
}

func (p *SubscriptionProduct) HasWeeklyLimit() bool {
	return p != nil && p.WeeklyLimitUSD > 0
}

func (p *SubscriptionProduct) HasMonthlyLimit() bool {
	return p != nil && p.MonthlyLimitUSD > 0
}

type SubscriptionProductBinding struct {
	ProductID   int64
	ProductCode string
	ProductName string

	DefaultValidityDays int
	DailyLimitUSD       float64
	WeeklyLimitUSD      float64
	MonthlyLimitUSD     float64

	GroupID           int64
	GroupName         string
	GroupPlatform     string
	GroupStatus       string
	GroupSubscription string
	DebitMultiplier   float64
	ProductStatus     string
	BindingStatus     string
}

func (b *SubscriptionProductBinding) Product() *SubscriptionProduct {
	if b == nil {
		return nil
	}
	return &SubscriptionProduct{
		ID:                  b.ProductID,
		Code:                b.ProductCode,
		Name:                b.ProductName,
		Status:              b.ProductStatus,
		DefaultValidityDays: b.DefaultValidityDays,
		DailyLimitUSD:       b.DailyLimitUSD,
		WeeklyLimitUSD:      b.WeeklyLimitUSD,
		MonthlyLimitUSD:     b.MonthlyLimitUSD,
	}
}

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
}

func (s *UserProductSubscription) IsActive() bool {
	return s != nil && s.Status == SubscriptionStatusActive && time.Now().Before(s.ExpiresAt)
}

func (s *UserProductSubscription) NeedsDailyReset() bool {
	if s == nil || s.DailyWindowStart == nil {
		return false
	}
	return time.Since(*s.DailyWindowStart) >= 24*time.Hour
}

func (s *UserProductSubscription) NeedsWeeklyReset() bool {
	if s == nil || s.WeeklyWindowStart == nil {
		return false
	}
	return time.Since(*s.WeeklyWindowStart) >= 7*24*time.Hour
}

func (s *UserProductSubscription) NeedsMonthlyReset() bool {
	if s == nil || s.MonthlyWindowStart == nil {
		return false
	}
	return time.Since(*s.MonthlyWindowStart) >= 30*24*time.Hour
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

type ProductSettlementContext struct {
	Binding      *SubscriptionProductBinding
	Subscription *UserProductSubscription
}

type ProductSubscriptionRepository interface {
	GetActiveProductSubscriptionByUserAndGroupID(ctx context.Context, userID, groupID int64) (*SubscriptionProductBinding, *UserProductSubscription, error)
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func elapsedDailyWindows(windowStart, newWindowStart time.Time) int {
	if newWindowStart.Before(windowStart) {
		return 0
	}
	return int(newWindowStart.Sub(windowStart) / (24 * time.Hour))
}
