//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionProductRepositoryAssignOrExtendProductSubscription(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

	user := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("product-assign-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
	})

	var productID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, default_validity_days, daily_limit_usd)
		VALUES ($1, $2, 'active', 14, 45)
		RETURNING id
	`, "assign-"+uuid.NewString(), "assign product").Scan(&productID))

	sub, reused, err := repo.AssignOrExtendProductSubscription(ctx, &service.AssignProductSubscriptionInput{
		UserID:       user.ID,
		ProductID:    productID,
		ValidityDays: 7,
		Notes:        "first grant",
	})
	require.NoError(t, err)
	require.False(t, reused)
	require.Equal(t, user.ID, sub.UserID)
	require.Equal(t, productID, sub.ProductID)
	require.Equal(t, service.SubscriptionStatusActive, sub.Status)
	require.Contains(t, sub.Notes, "first grant")
	require.WithinDuration(t, time.Now().AddDate(0, 0, 7), sub.ExpiresAt, 5*time.Second)

	extended, reused, err := repo.AssignOrExtendProductSubscription(ctx, &service.AssignProductSubscriptionInput{
		UserID:       user.ID,
		ProductID:    productID,
		ValidityDays: 3,
		Notes:        "second grant",
	})
	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, sub.ID, extended.ID)
	require.WithinDuration(t, sub.ExpiresAt.AddDate(0, 0, 3), extended.ExpiresAt, 2*time.Second)
	require.Contains(t, extended.Notes, "first grant")
	require.Contains(t, extended.Notes, "second grant")
}

func TestSubscriptionProductRepositoryAdminProductManagement(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := &subscriptionProductRepository{client: client, sql: tx.Client()}

	groupA, err := client.Group.Create().
		SetName(uniqueTestValue(t, "product-admin-a")).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)
	groupB, err := client.Group.Create().
		SetName(uniqueTestValue(t, "product-admin-b")).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().
		SetEmail(uniqueTestValue(t, "product-admin") + "@example.com").
		SetPasswordHash("hash").
		Save(ctx)
	require.NoError(t, err)

	product, err := repo.CreateProduct(ctx, &service.CreateSubscriptionProductInput{
		Code:                "admin-" + uuid.NewString(),
		Name:                "admin product",
		Status:              service.SubscriptionProductStatusActive,
		DefaultValidityDays: 10,
		DailyLimitUSD:       12,
		MonthlyLimitUSD:     120,
	})
	require.NoError(t, err)
	require.Equal(t, "admin product", product.Name)

	renamed := "admin product updated"
	updated, err := repo.UpdateProduct(ctx, product.ID, &service.UpdateSubscriptionProductInput{Name: &renamed})
	require.NoError(t, err)
	require.Equal(t, renamed, updated.Name)

	bindings, err := repo.SyncProductBindings(ctx, product.ID, []service.SubscriptionProductBindingInput{
		{GroupID: groupA.ID, DebitMultiplier: 1, Status: service.SubscriptionProductBindingStatusActive, SortOrder: 2},
		{GroupID: groupB.ID, DebitMultiplier: 2.5, Status: service.SubscriptionProductBindingStatusActive, SortOrder: 1},
	})
	require.NoError(t, err)
	require.Len(t, bindings, 2)
	require.Equal(t, groupB.ID, bindings[0].GroupID)
	require.Equal(t, 2.5, bindings[0].DebitMultiplier)

	products, err := repo.ListProducts(ctx)
	require.NoError(t, err)
	require.Contains(t, productIDs(products), product.ID)

	sub, reused, err := repo.AssignOrExtendProductSubscription(ctx, &service.AssignProductSubscriptionInput{
		UserID:       user.ID,
		ProductID:    product.ID,
		ValidityDays: 7,
		Notes:        "admin assignment",
	})
	require.NoError(t, err)
	require.False(t, reused)
	require.Equal(t, product.ID, sub.ProductID)

	subs, err := repo.ListProductSubscriptions(ctx, product.ID)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	require.Equal(t, user.ID, subs[0].UserID)
}

func TestSubscriptionProductRepositoryListUserProductSubscriptionsForAdmin(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := &subscriptionProductRepository{client: client, sql: tx.Client()}

	user, err := client.User.Create().
		SetEmail(uniqueTestValue(t, "product-sub-admin") + "@example.com").
		SetPasswordHash("hash").
		SetUsername("Product User").
		Save(ctx)
	require.NoError(t, err)

	product, err := repo.CreateProduct(ctx, &service.CreateSubscriptionProductInput{
		Code:                "admin-list-" + uuid.NewString(),
		Name:                "admin list product",
		Status:              service.SubscriptionProductStatusActive,
		DefaultValidityDays: 30,
		DailyLimitUSD:       45,
	})
	require.NoError(t, err)

	now := time.Now().UTC()
	_, _, err = repo.AssignOrExtendProductSubscription(ctx, &service.AssignProductSubscriptionInput{
		UserID:       user.ID,
		ProductID:    product.ID,
		ValidityDays: 7,
		Notes:        "admin list subscription",
	})
	require.NoError(t, err)
	_, err = tx.Client().ExecContext(ctx, `
UPDATE user_product_subscriptions
SET daily_usage_usd = 12.5,
    weekly_usage_usd = 25,
    monthly_usage_usd = 50,
    daily_carryover_in_usd = 8,
    daily_carryover_remaining_usd = 3,
    starts_at = $1
WHERE user_id = $2
  AND product_id = $3`, now.Add(-time.Hour), user.ID, product.ID)
	require.NoError(t, err)

	items, page, err := repo.ListUserProductSubscriptionsForAdmin(ctx, service.AdminProductSubscriptionListParams{
		Page:     1,
		PageSize: 20,
		Search:   user.Email,
		Status:   service.SubscriptionStatusActive,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, items, 1)
	require.Equal(t, user.ID, items[0].UserID)
	require.Equal(t, user.Email, items[0].UserEmail)
	require.Equal(t, product.ID, items[0].ProductID)
	require.Equal(t, product.Code, items[0].ProductCode)
	require.Equal(t, product.Name, items[0].ProductName)
	require.InDelta(t, 45, items[0].DailyLimitUSD, 0.000001)
	require.InDelta(t, 12.5, items[0].DailyUsageUSD, 0.000001)
	require.InDelta(t, 8, items[0].DailyCarryoverInUSD, 0.000001)
	require.InDelta(t, 3, items[0].DailyCarryoverRemainingUSD, 0.000001)
	require.InDelta(t, 5, items[0].CarryoverUsedUSD, 0.000001)
	require.InDelta(t, 7.5, items[0].FreshDailyUsageUSD, 0.000001)
	require.Contains(t, items[0].Notes, "admin list subscription")
}

func TestSubscriptionProductRepositoryResolveActiveProductByGroupID(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := &subscriptionProductRepository{client: client, sql: tx.Client()}

	group, err := client.Group.Create().
		SetName(uniqueTestValue(t, "product-resolve")).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)

	product, err := repo.CreateProduct(ctx, &service.CreateSubscriptionProductInput{
		Code:                "resolve-" + uuid.NewString(),
		Name:                "resolve product",
		Status:              service.SubscriptionProductStatusActive,
		DefaultValidityDays: 10,
		SortOrder:           20,
	})
	require.NoError(t, err)
	_, err = repo.SyncProductBindings(ctx, product.ID, []service.SubscriptionProductBindingInput{
		{GroupID: group.ID, DebitMultiplier: 1, Status: service.SubscriptionProductBindingStatusActive, SortOrder: 1},
	})
	require.NoError(t, err)

	resolved, err := repo.ResolveActiveProductByGroupID(ctx, group.ID)
	require.NoError(t, err)
	require.Equal(t, product.ID, resolved.ID)
	require.Equal(t, product.Code, resolved.Code)
}

func productIDs(products []service.SubscriptionProduct) []int64 {
	out := make([]int64, 0, len(products))
	for _, product := range products {
		out = append(out, product.ID)
	}
	return out
}
