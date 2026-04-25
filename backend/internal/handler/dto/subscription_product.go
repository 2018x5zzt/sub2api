package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ActiveSubscriptionProduct struct {
	ProductID      int64  `json:"product_id"`
	SubscriptionID int64  `json:"subscription_id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Description    string `json:"description"`

	ExpiresAt *time.Time `json:"expires_at"`
	Status    string     `json:"status"`

	DailyUsageUSD   float64 `json:"daily_usage_usd"`
	WeeklyUsageUSD  float64 `json:"weekly_usage_usd"`
	MonthlyUsageUSD float64 `json:"monthly_usage_usd"`

	DailyLimitUSD   float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD  float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD float64 `json:"monthly_limit_usd"`

	DailyCarryoverInUSD        float64 `json:"daily_carryover_in_usd"`
	DailyCarryoverRemainingUSD float64 `json:"daily_carryover_remaining_usd"`

	Groups []SubscriptionProductGroup `json:"groups"`
}

type SubscriptionProductGroup struct {
	GroupID         int64   `json:"group_id"`
	GroupName       string  `json:"group_name"`
	DebitMultiplier float64 `json:"debit_multiplier"`
	Status          string  `json:"status"`
	SortOrder       int     `json:"sort_order"`
}

type SubscriptionProductSummary struct {
	ActiveCount          int                         `json:"active_count"`
	TotalMonthlyUsageUSD float64                     `json:"total_monthly_usage_usd"`
	TotalMonthlyLimitUSD float64                     `json:"total_monthly_limit_usd"`
	Products             []ActiveSubscriptionProduct `json:"products"`
}

type AdminSubscriptionProduct struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`

	DefaultValidityDays int     `json:"default_validity_days"`
	DailyLimitUSD       float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD      float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD     float64 `json:"monthly_limit_usd"`
	SortOrder           int     `json:"sort_order"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AdminSubscriptionProductBinding struct {
	ProductID       int64     `json:"product_id"`
	GroupID         int64     `json:"group_id"`
	GroupName       string    `json:"group_name"`
	DebitMultiplier float64   `json:"debit_multiplier"`
	Status          string    `json:"status"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AdminUserProductSubscription struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	StartsAt  time.Time `json:"starts_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Status    string    `json:"status"`

	DailyWindowStart   *time.Time `json:"daily_window_start"`
	WeeklyWindowStart  *time.Time `json:"weekly_window_start"`
	MonthlyWindowStart *time.Time `json:"monthly_window_start"`

	DailyUsageUSD              float64 `json:"daily_usage_usd"`
	WeeklyUsageUSD             float64 `json:"weekly_usage_usd"`
	MonthlyUsageUSD            float64 `json:"monthly_usage_usd"`
	DailyCarryoverInUSD        float64 `json:"daily_carryover_in_usd"`
	DailyCarryoverRemainingUSD float64 `json:"daily_carryover_remaining_usd"`

	AssignedBy *int64    `json:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ActiveSubscriptionProductFromService(item *service.ActiveSubscriptionProduct) *ActiveSubscriptionProduct {
	if item == nil {
		return nil
	}
	out := &ActiveSubscriptionProduct{
		ProductID:                  item.Product.ID,
		SubscriptionID:             item.Subscription.ID,
		Code:                       item.Product.Code,
		Name:                       item.Product.Name,
		Description:                item.Product.Description,
		Status:                     item.Subscription.Status,
		DailyUsageUSD:              item.Subscription.DailyUsageUSD,
		WeeklyUsageUSD:             item.Subscription.WeeklyUsageUSD,
		MonthlyUsageUSD:            item.Subscription.MonthlyUsageUSD,
		DailyLimitUSD:              item.Product.DailyLimitUSD,
		WeeklyLimitUSD:             item.Product.WeeklyLimitUSD,
		MonthlyLimitUSD:            item.Product.MonthlyLimitUSD,
		DailyCarryoverInUSD:        item.Subscription.DailyCarryoverInUSD,
		DailyCarryoverRemainingUSD: item.Subscription.DailyCarryoverRemainingUSD,
		Groups:                     make([]SubscriptionProductGroup, 0, len(item.Groups)),
	}
	if !item.Subscription.ExpiresAt.IsZero() {
		expiresAt := item.Subscription.ExpiresAt
		out.ExpiresAt = &expiresAt
	}
	for _, group := range item.Groups {
		out.Groups = append(out.Groups, SubscriptionProductGroup{
			GroupID:         group.GroupID,
			GroupName:       group.GroupName,
			DebitMultiplier: group.DebitMultiplier,
			Status:          group.Status,
			SortOrder:       group.SortOrder,
		})
	}
	return out
}

func ActiveSubscriptionProductsFromService(items []service.ActiveSubscriptionProduct) []ActiveSubscriptionProduct {
	out := make([]ActiveSubscriptionProduct, 0, len(items))
	for i := range items {
		out = append(out, *ActiveSubscriptionProductFromService(&items[i]))
	}
	return out
}

func SubscriptionProductSummaryFromService(summary *service.SubscriptionProductSummary) *SubscriptionProductSummary {
	if summary == nil {
		return nil
	}
	return &SubscriptionProductSummary{
		ActiveCount:          summary.ActiveCount,
		TotalMonthlyUsageUSD: summary.TotalMonthlyUsageUSD,
		TotalMonthlyLimitUSD: summary.TotalMonthlyLimitUSD,
		Products:             ActiveSubscriptionProductsFromService(summary.Products),
	}
}

func AdminSubscriptionProductFromService(product *service.SubscriptionProduct) *AdminSubscriptionProduct {
	if product == nil {
		return nil
	}
	return &AdminSubscriptionProduct{
		ID:                  product.ID,
		Code:                product.Code,
		Name:                product.Name,
		Description:         product.Description,
		Status:              product.Status,
		DefaultValidityDays: product.DefaultValidityDays,
		DailyLimitUSD:       product.DailyLimitUSD,
		WeeklyLimitUSD:      product.WeeklyLimitUSD,
		MonthlyLimitUSD:     product.MonthlyLimitUSD,
		SortOrder:           product.SortOrder,
		CreatedAt:           product.CreatedAt,
		UpdatedAt:           product.UpdatedAt,
	}
}

func AdminSubscriptionProductsFromService(products []service.SubscriptionProduct) []AdminSubscriptionProduct {
	out := make([]AdminSubscriptionProduct, 0, len(products))
	for i := range products {
		out = append(out, *AdminSubscriptionProductFromService(&products[i]))
	}
	return out
}

func AdminSubscriptionProductBindingsFromService(bindings []service.SubscriptionProductBindingDetail) []AdminSubscriptionProductBinding {
	out := make([]AdminSubscriptionProductBinding, 0, len(bindings))
	for _, binding := range bindings {
		out = append(out, AdminSubscriptionProductBinding{
			ProductID:       binding.ProductID,
			GroupID:         binding.GroupID,
			GroupName:       binding.GroupName,
			DebitMultiplier: binding.DebitMultiplier,
			Status:          binding.Status,
			SortOrder:       binding.SortOrder,
			CreatedAt:       binding.CreatedAt,
			UpdatedAt:       binding.UpdatedAt,
		})
	}
	return out
}

func AdminUserProductSubscriptionsFromService(subscriptions []service.UserProductSubscription) []AdminUserProductSubscription {
	out := make([]AdminUserProductSubscription, 0, len(subscriptions))
	for _, sub := range subscriptions {
		out = append(out, AdminUserProductSubscription{
			ID:                         sub.ID,
			UserID:                     sub.UserID,
			ProductID:                  sub.ProductID,
			StartsAt:                   sub.StartsAt,
			ExpiresAt:                  sub.ExpiresAt,
			Status:                     sub.Status,
			DailyWindowStart:           sub.DailyWindowStart,
			WeeklyWindowStart:          sub.WeeklyWindowStart,
			MonthlyWindowStart:         sub.MonthlyWindowStart,
			DailyUsageUSD:              sub.DailyUsageUSD,
			WeeklyUsageUSD:             sub.WeeklyUsageUSD,
			MonthlyUsageUSD:            sub.MonthlyUsageUSD,
			DailyCarryoverInUSD:        sub.DailyCarryoverInUSD,
			DailyCarryoverRemainingUSD: sub.DailyCarryoverRemainingUSD,
			AssignedBy:                 sub.AssignedBy,
			AssignedAt:                 sub.AssignedAt,
			Notes:                      sub.Notes,
			CreatedAt:                  sub.CreatedAt,
			UpdatedAt:                  sub.UpdatedAt,
		})
	}
	return out
}
