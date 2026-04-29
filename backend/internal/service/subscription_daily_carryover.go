package service

import "time"

type DailyCarryoverPreview struct {
	WindowStart           time.Time
	CarryoverInUSD        float64
	CarryoverRemainingUSD float64
	EffectiveLimitUSD     float64
	RemainingTotalUSD     float64
	ElapsedDays           int
}

func previewDailyWindowAdvance(sub *UserSubscription, group *Group, newWindowStart time.Time) DailyCarryoverPreview {
	preview := DailyCarryoverPreview{
		WindowStart: newWindowStart,
	}
	if sub == nil || group == nil || !group.HasDailyLimit() {
		return preview
	}

	preview.EffectiveLimitUSD = *group.DailyLimitUSD
	preview.RemainingTotalUSD = preview.EffectiveLimitUSD
	if sub.DailyWindowStart == nil {
		return preview
	}

	preview.ElapsedDays = elapsedDailyWindows(*sub.DailyWindowStart, newWindowStart)
	preview.EffectiveLimitUSD = *group.DailyLimitUSD
	preview.RemainingTotalUSD = preview.EffectiveLimitUSD
	return preview
}

func previewProductDailyWindowAdvance(sub *UserProductSubscription, product *SubscriptionProduct, newWindowStart time.Time) DailyCarryoverPreview {
	preview := DailyCarryoverPreview{
		WindowStart: newWindowStart,
	}
	if sub == nil || product == nil || !product.HasDailyLimit() {
		return preview
	}

	preview.EffectiveLimitUSD = product.DailyLimitUSD
	preview.RemainingTotalUSD = preview.EffectiveLimitUSD
	if sub.DailyWindowStart == nil {
		return preview
	}

	preview.ElapsedDays = elapsedDailyWindows(*sub.DailyWindowStart, newWindowStart)

	carryoverConsumed := maxFloat64(sub.DailyCarryoverInUSD-sub.DailyCarryoverRemainingUSD, 0)
	freshUsed := maxFloat64(sub.DailyUsageUSD-carryoverConsumed, 0)
	freshRemaining := maxFloat64(product.DailyLimitUSD-freshUsed, 0)

	if preview.ElapsedDays == 1 && sub.Status == SubscriptionStatusActive && sub.ExpiresAt.After(newWindowStart) {
		preview.CarryoverInUSD = freshRemaining
		preview.CarryoverRemainingUSD = freshRemaining
	}

	preview.EffectiveLimitUSD = product.DailyLimitUSD + preview.CarryoverInUSD
	preview.RemainingTotalUSD = preview.EffectiveLimitUSD
	return preview
}

func elapsedDailyWindows(windowStart, newWindowStart time.Time) int {
	if newWindowStart.Before(windowStart) {
		return 0
	}
	return int(newWindowStart.Sub(windowStart) / (24 * time.Hour))
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
