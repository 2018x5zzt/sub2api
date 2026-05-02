//go:build unit

package service

import (
	"context"
	"testing"
)

// TestBuildUsageBillingCommand_SubscriptionAppliesRateMultiplier locks in the fix
// that subscription-mode billing honours the group (and any user-specific) rate
// multiplier — i.e. cmd.SubscriptionCost tracks ActualCost (= TotalCost *
// RateMultiplier), not raw TotalCost.
func TestBuildUsageBillingCommand_SubscriptionAppliesRateMultiplier(t *testing.T) {
	t.Parallel()

	groupID := int64(7)
	subID := int64(42)

	tests := []struct {
		name           string
		totalCost      float64
		actualCost     float64
		isSubscription bool
		wantSub        float64
		wantBalance    float64
	}{
		{
			name:           "subscription with 2x multiplier consumes 2x quota",
			totalCost:      1.0,
			actualCost:     2.0,
			isSubscription: true,
			wantSub:        2.0,
			wantBalance:    0,
		},
		{
			name:           "subscription with 0.5x multiplier consumes 0.5x quota",
			totalCost:      1.0,
			actualCost:     0.5,
			isSubscription: true,
			wantSub:        0.5,
			wantBalance:    0,
		},
		{
			name:           "free subscription (multiplier 0) consumes no quota",
			totalCost:      1.0,
			actualCost:     0,
			isSubscription: true,
			wantSub:        0,
			wantBalance:    0,
		},
		{
			name:           "balance billing keeps using ActualCost (regression)",
			totalCost:      1.0,
			actualCost:     2.0,
			isSubscription: false,
			wantSub:        0,
			wantBalance:    2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &postUsageBillingParams{
				Cost:               &CostBreakdown{TotalCost: tt.totalCost, ActualCost: tt.actualCost},
				User:               &User{ID: 1},
				APIKey:             &APIKey{ID: 2, GroupID: &groupID},
				Account:            &Account{ID: 3},
				Subscription:       &UserSubscription{ID: subID},
				IsSubscriptionBill: tt.isSubscription,
			}

			cmd := buildUsageBillingCommand("req-1", nil, p)
			if cmd == nil {
				t.Fatal("buildUsageBillingCommand returned nil")
			}
			if cmd.SubscriptionCost != tt.wantSub {
				t.Errorf("SubscriptionCost = %v, want %v", cmd.SubscriptionCost, tt.wantSub)
			}
			if cmd.BalanceCost != tt.wantBalance {
				t.Errorf("BalanceCost = %v, want %v", cmd.BalanceCost, tt.wantBalance)
			}
		})
	}
}

func TestBuildUsageBillingCommand_ProductSubscriptionUsesSharedProductDebit(t *testing.T) {
	t.Parallel()

	groupID := int64(7)
	productID := int64(88)
	productSubID := int64(99)
	multiplier := 1.5
	settlement := &ProductSettlementContext{
		Binding: &SubscriptionProductBinding{
			ProductID:       productID,
			GroupID:         groupID,
			DebitMultiplier: multiplier,
		},
		Subscription: &UserProductSubscription{ID: productSubID},
	}

	p := &postUsageBillingParams{
		Cost:               &CostBreakdown{TotalCost: 10, ActualCost: 20},
		User:               &User{ID: 1},
		APIKey:             &APIKey{ID: 2, GroupID: &groupID},
		Account:            &Account{ID: 3},
		Subscription:       &UserSubscription{ID: 42},
		ProductSettlement:  settlement,
		IsSubscriptionBill: true,
	}
	log := &UsageLog{
		SubscriptionID: &p.Subscription.ID,
	}

	cmd := buildUsageBillingCommand("req-product", log, p)
	if cmd == nil {
		t.Fatal("buildUsageBillingCommand returned nil")
	}
	if cmd.SubscriptionID != nil || cmd.SubscriptionCost != 0 {
		t.Fatalf("group subscription billing should be suppressed, got id=%v cost=%v", cmd.SubscriptionID, cmd.SubscriptionCost)
	}
	if cmd.ProductSubscriptionID == nil || *cmd.ProductSubscriptionID != productSubID {
		t.Fatalf("ProductSubscriptionID = %v, want %d", cmd.ProductSubscriptionID, productSubID)
	}
	if cmd.ProductDebitCost != 15 {
		t.Fatalf("ProductDebitCost = %v, want 15", cmd.ProductDebitCost)
	}
	if log.SubscriptionID != nil {
		t.Fatalf("usage log SubscriptionID should be cleared for product settlement")
	}
	if log.ProductID == nil || *log.ProductID != productID {
		t.Fatalf("usage log ProductID = %v, want %d", log.ProductID, productID)
	}
	if log.ProductDebitCost == nil || *log.ProductDebitCost != 15 {
		t.Fatalf("usage log ProductDebitCost = %v, want 15", log.ProductDebitCost)
	}
}

func TestBuildUsageBillingCommand_SubscriptionBillingDoesNotConsumeAPIKeyQuota(t *testing.T) {
	t.Parallel()

	groupID := int64(7)
	subID := int64(42)
	p := &postUsageBillingParams{
		Cost:               &CostBreakdown{TotalCost: 10, ActualCost: 10},
		User:               &User{ID: 1},
		APIKey:             &APIKey{ID: 2, GroupID: &groupID, Quota: 45, RateLimit1d: 45},
		Account:            &Account{ID: 3},
		Subscription:       &UserSubscription{ID: subID},
		IsSubscriptionBill: true,
		APIKeyService:      &noopAPIKeyQuotaUpdater{},
	}

	cmd := buildUsageBillingCommand("req-sub-key-quota", nil, p)

	if cmd == nil {
		t.Fatal("buildUsageBillingCommand returned nil")
	}
	if cmd.SubscriptionID == nil || *cmd.SubscriptionID != subID || cmd.SubscriptionCost != 10 {
		t.Fatalf("Subscription billing = id %v cost %v, want id %d cost 10", cmd.SubscriptionID, cmd.SubscriptionCost, subID)
	}
	if cmd.APIKeyQuotaCost != 0 {
		t.Fatalf("APIKeyQuotaCost = %v, want 0 for subscription billing", cmd.APIKeyQuotaCost)
	}
	if cmd.APIKeyRateLimitCost != 0 {
		t.Fatalf("APIKeyRateLimitCost = %v, want 0 for subscription billing", cmd.APIKeyRateLimitCost)
	}
}

func TestShouldQueueAPIKeyRateLimitCacheSkipsSubscriptionBilling(t *testing.T) {
	t.Parallel()

	p := &postUsageBillingParams{
		Cost:               &CostBreakdown{TotalCost: 10, ActualCost: 10},
		APIKey:             &APIKey{ID: 2, RateLimit1d: 45},
		IsSubscriptionBill: true,
	}

	if shouldQueueAPIKeyRateLimitCache(p) {
		t.Fatal("subscription billing should not update API key rate-limit cache")
	}
}

type noopAPIKeyQuotaUpdater struct{}

func (n *noopAPIKeyQuotaUpdater) UpdateQuotaUsed(_ context.Context, _ int64, _ float64) error {
	return nil
}

func (n *noopAPIKeyQuotaUpdater) UpdateRateLimitUsage(_ context.Context, _ int64, _ float64) error {
	return nil
}
