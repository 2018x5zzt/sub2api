package service

import "time"

const (
	SubscriptionProductStatusDraft    = "draft"
	SubscriptionProductStatusActive   = "active"
	SubscriptionProductStatusDisabled = "disabled"

	SubscriptionProductBindingStatusActive   = "active"
	SubscriptionProductBindingStatusInactive = "inactive"
)

type SubscriptionProduct struct {
	ID          int64
	Code        string
	Name        string
	Description string
	Status      string

	DefaultValidityDays int
	DailyLimitUSD       float64
	WeeklyLimitUSD      float64
	MonthlyLimitUSD     float64
	SortOrder           int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *SubscriptionProduct) IsActive() bool {
	return p != nil && p.Status == SubscriptionProductStatusActive
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

	DebitMultiplier float64
	ProductStatus   string
	BindingStatus   string
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

type ProductSettlementContext struct {
	Binding      *SubscriptionProductBinding
	Subscription *UserProductSubscription
}
