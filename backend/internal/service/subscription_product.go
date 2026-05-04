package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	SubscriptionProductStatusDraft    = "draft"
	SubscriptionProductStatusActive   = "active"
	SubscriptionProductStatusDisabled = "disabled"

	SubscriptionProductBindingStatusActive   = "active"
	SubscriptionProductBindingStatusInactive = "inactive"
)

type SubscriptionProduct struct {
	ID                  int64
	Code                string
	Name                string
	Description         string
	Status              string
	ProductFamily       string
	DefaultValidityDays int
	DailyLimitUSD       float64
	WeeklyLimitUSD      float64
	MonthlyLimitUSD     float64
	SortOrder           int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CreateSubscriptionProductInput struct {
	Code          string
	Name          string
	Description   string
	Status        string
	ProductFamily string

	DefaultValidityDays int
	DailyLimitUSD       float64
	WeeklyLimitUSD      float64
	MonthlyLimitUSD     float64
	SortOrder           int
}

type UpdateSubscriptionProductInput struct {
	Code          *string
	Name          *string
	Description   *string
	Status        *string
	ProductFamily *string

	DefaultValidityDays *int
	DailyLimitUSD       *float64
	WeeklyLimitUSD      *float64
	MonthlyLimitUSD     *float64
	SortOrder           *int
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
	ProductID     int64
	ProductCode   string
	ProductName   string
	ProductFamily string

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
		ProductFamily:       b.ProductFamily,
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

	AssignedBy *int64
	AssignedAt time.Time
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AdminProductSubscriptionListParams struct {
	Page      int
	PageSize  int
	Search    string
	ProductID int64
	UserID    int64
	Status    string
	SortBy    string
	SortOrder string
}

type AdminProductSubscriptionListItem struct {
	UserProductSubscription

	UserEmail    string
	UserUsername string

	ProductCode   string
	ProductName   string
	DailyLimitUSD float64

	CarryoverUsedUSD   float64
	FreshDailyUsageUSD float64
}

type SubscriptionProductBindingInput struct {
	GroupID         int64
	DebitMultiplier float64
	Status          string
	SortOrder       int
}

type SubscriptionProductBindingDetail struct {
	ProductID       int64
	GroupID         int64
	GroupName       string
	DebitMultiplier float64
	Status          string
	SortOrder       int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type AssignProductSubscriptionInput struct {
	UserID       int64
	ProductID    int64
	ValidityDays int
	AssignedBy   int64
	Notes        string
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

type SubscriptionProductGroupSummary struct {
	GroupID         int64
	GroupName       string
	GroupPlatform   string
	DebitMultiplier float64
	Status          string
	SortOrder       int
}

type ActiveSubscriptionProduct struct {
	Product      SubscriptionProduct
	Subscription UserProductSubscription
	Groups       []SubscriptionProductGroupSummary
}

type SubscriptionProductSummary struct {
	ActiveCount          int
	TotalMonthlyUsageUSD float64
	TotalMonthlyLimitUSD float64
	Products             []ActiveSubscriptionProduct
}

type ProductSubscriptionRepository interface {
	GetActiveProductSubscriptionByUserAndGroupID(ctx context.Context, userID, groupID int64) (*SubscriptionProductBinding, *UserProductSubscription, error)
	ListActiveProductsByUserID(ctx context.Context, userID int64) ([]ActiveSubscriptionProduct, error)
	ListVisibleGroupsByUserID(ctx context.Context, userID int64) ([]Group, error)
	ListProducts(ctx context.Context) ([]SubscriptionProduct, error)
	ResolveActiveProductByGroupID(ctx context.Context, groupID int64) (*SubscriptionProduct, error)
	CreateProduct(ctx context.Context, input *CreateSubscriptionProductInput) (*SubscriptionProduct, error)
	UpdateProduct(ctx context.Context, productID int64, input *UpdateSubscriptionProductInput) (*SubscriptionProduct, error)
	ListProductBindings(ctx context.Context, productID int64) ([]SubscriptionProductBindingDetail, error)
	SyncProductBindings(ctx context.Context, productID int64, inputs []SubscriptionProductBindingInput) ([]SubscriptionProductBindingDetail, error)
	ListProductSubscriptions(ctx context.Context, productID int64) ([]UserProductSubscription, error)
	ListUserProductSubscriptionsForAdmin(ctx context.Context, params AdminProductSubscriptionListParams) ([]AdminProductSubscriptionListItem, *pagination.PaginationResult, error)
	AssignOrExtendProductSubscription(ctx context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error)
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
