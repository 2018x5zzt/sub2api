package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

type productSettlementBillingFields struct {
	productID             int64
	productSubscriptionID int64
	groupID               int64
	groupDebitMultiplier  float64
	productDebitCost      float64
}

func hasProductSettlement(settlement *ProductSettlementContext) bool {
	return settlement != nil &&
		settlement.Binding != nil &&
		settlement.Subscription != nil &&
		settlement.Binding.DebitMultiplier > 0
}

func applyProductSettlementUsageLog(log *UsageLog, settlement *ProductSettlementContext, totalCost float64) {
	if log == nil || !hasProductSettlement(settlement) {
		return
	}

	productID := settlement.Binding.ProductID
	productSubscriptionID := settlement.Subscription.ID
	groupDebitMultiplier := settlement.Binding.DebitMultiplier
	log.SubscriptionID = nil
	log.ProductID = &productID
	log.ProductSubscriptionID = &productSubscriptionID
	log.GroupDebitMultiplier = &groupDebitMultiplier
	if totalCost > 0 {
		productDebitCost := totalCost * groupDebitMultiplier
		log.ProductDebitCost = &productDebitCost
	}
}

func productSettlementBilling(settlement *ProductSettlementContext, totalCost float64) (productSettlementBillingFields, bool) {
	if !hasProductSettlement(settlement) || totalCost <= 0 {
		return productSettlementBillingFields{}, false
	}
	multiplier := settlement.Binding.DebitMultiplier
	return productSettlementBillingFields{
		productID:             settlement.Binding.ProductID,
		productSubscriptionID: settlement.Subscription.ID,
		groupID:               settlement.Binding.GroupID,
		groupDebitMultiplier:  multiplier,
		productDebitCost:      totalCost * multiplier,
	}, true
}

func ContextWithProductSettlement(ctx context.Context, settlement *ProductSettlementContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if settlement == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxkey.ProductSettlement, settlement)
}

func ProductSettlementFromContext(ctx context.Context) (*ProductSettlementContext, bool) {
	if ctx == nil {
		return nil, false
	}
	settlement, ok := ctx.Value(ctxkey.ProductSettlement).(*ProductSettlementContext)
	return settlement, ok && settlement != nil
}

func ContextWithSubscriptionBalanceFallback(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxkey.SubscriptionBalanceFallback, true)
}

func SubscriptionBalanceFallbackFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, _ := ctx.Value(ctxkey.SubscriptionBalanceFallback).(bool)
	return v
}
