//go:build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestUsageLogRepositoryCreateBestEffort_PersistsProductSettlementFields(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := newUsageLogRepositoryWithSQL(client, integrationDB)

	user := mustCreateUser(t, client, &service.User{Email: fmt.Sprintf("usage-product-%d@example.com", time.Now().UnixNano())})
	apiKey := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-usage-product-" + uuid.NewString(), Name: "k"})
	account := mustCreateAccount(t, client, &service.Account{Name: "acc-usage-product-" + uuid.NewString()})
	product := mustCreateSubscriptionProduct(t, "usage_log_product", "Usage Log Product", service.SubscriptionProductStatusActive)
	productSub := mustCreateUserProductSubscription(t, user.ID, product.ID, service.SubscriptionStatusActive, time.Now().Add(24*time.Hour))
	groupDebitMultiplier := 1.5
	productDebitCost := 0.75

	log := &service.UsageLog{
		UserID:                user.ID,
		APIKeyID:              apiKey.ID,
		AccountID:             account.ID,
		RequestID:             uuid.NewString(),
		Model:                 "gpt-5.5",
		InputTokens:           10,
		OutputTokens:          20,
		TotalCost:             0.5,
		ActualCost:            0.5,
		ProductID:             &product.ID,
		ProductSubscriptionID: &productSub.ID,
		GroupDebitMultiplier:  &groupDebitMultiplier,
		ProductDebitCost:      &productDebitCost,
		BillingType:           service.BillingTypeSubscription,
		CreatedAt:             time.Now().UTC(),
	}

	require.NoError(t, repo.CreateBestEffort(ctx, log))

	var gotProductID, gotProductSubID sql.NullInt64
	var gotGroupDebitMultiplier, gotProductDebitCost sql.NullFloat64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		SELECT product_id, product_subscription_id, group_debit_multiplier, product_debit_cost
		FROM usage_logs
		WHERE request_id = $1 AND api_key_id = $2
	`, log.RequestID, apiKey.ID).Scan(
		&gotProductID,
		&gotProductSubID,
		&gotGroupDebitMultiplier,
		&gotProductDebitCost,
	))
	require.Equal(t, sql.NullInt64{Int64: product.ID, Valid: true}, gotProductID)
	require.Equal(t, sql.NullInt64{Int64: productSub.ID, Valid: true}, gotProductSubID)
	require.True(t, gotGroupDebitMultiplier.Valid)
	require.InDelta(t, groupDebitMultiplier, gotGroupDebitMultiplier.Float64, 1e-12)
	require.True(t, gotProductDebitCost.Valid)
	require.InDelta(t, productDebitCost, gotProductDebitCost.Float64, 1e-12)
}
