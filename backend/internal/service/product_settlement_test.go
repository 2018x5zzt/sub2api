package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyProductSettlementUsageLogClearsLegacySubscription(t *testing.T) {
	legacySubID := int64(10)
	log := &UsageLog{SubscriptionID: &legacySubID}
	settlement := &ProductSettlementContext{
		Binding:      &SubscriptionProductBinding{ProductID: 20, GroupID: 30, DebitMultiplier: 1.5},
		Subscription: &UserProductSubscription{ID: 40},
	}

	applyProductSettlementUsageLog(log, settlement, 2)

	require.Nil(t, log.SubscriptionID)
	require.Equal(t, int64(20), *log.ProductID)
	require.Equal(t, int64(40), *log.ProductSubscriptionID)
	require.Equal(t, 1.5, *log.GroupDebitMultiplier)
	require.Equal(t, 3.0, *log.ProductDebitCost)
}

func TestProductSettlementBillingRequiresPositiveTotalCost(t *testing.T) {
	settlement := &ProductSettlementContext{
		Binding:      &SubscriptionProductBinding{ProductID: 20, GroupID: 30, DebitMultiplier: 1.5},
		Subscription: &UserProductSubscription{ID: 40},
	}

	_, ok := productSettlementBilling(settlement, 0)
	require.False(t, ok)

	fields, ok := productSettlementBilling(settlement, 2)
	require.True(t, ok)
	require.Equal(t, int64(20), fields.productID)
	require.Equal(t, int64(40), fields.productSubscriptionID)
	require.Equal(t, int64(30), fields.groupID)
	require.Equal(t, 1.5, fields.groupDebitMultiplier)
	require.Equal(t, 2.0, fields.standardTotalCost)
	require.Equal(t, 3.0, fields.productDebitCost)
}
