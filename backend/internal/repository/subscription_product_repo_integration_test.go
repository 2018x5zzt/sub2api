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
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

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

func TestSubscriptionProductRepositorySyncProductBindingsRejectsStandardGroup(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

	standardGroup, err := client.Group.Create().
		SetName(uniqueTestValue(t, "product-standard-binding")).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		Save(ctx)
	require.NoError(t, err)

	product, err := repo.CreateProduct(ctx, &service.CreateSubscriptionProductInput{
		Code:          "reject-standard-" + uuid.NewString(),
		Name:          "reject standard product",
		Status:        service.SubscriptionProductStatusActive,
		DailyLimitUSD: 12,
	})
	require.NoError(t, err)

	_, err = repo.SyncProductBindings(ctx, product.ID, []service.SubscriptionProductBindingInput{
		{GroupID: standardGroup.ID, DebitMultiplier: 1, Status: service.SubscriptionProductBindingStatusActive},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "subscription groups")
}

func TestSubscriptionProductRepositoryListActiveProductsNormalizesRolling30DayUsage(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

	var userID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, status, role, created_at, updated_at)
		VALUES ($1, 'hash', 'active', 'user', NOW(), NOW())
		RETURNING id
	`, uniqueTestValue(t, "product-list-month-user")+"@example.com").Scan(&userID))

	var productID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, daily_limit_usd, monthly_limit_usd)
		VALUES ($1, 'monthly display product', 'active', 45, 150)
		RETURNING id
	`, "list-month-"+uuid.NewString()).Scan(&productID))

	_, err := integrationDB.ExecContext(ctx, `
		INSERT INTO user_product_subscriptions (
			user_id,
			product_id,
			starts_at,
			expires_at,
			status,
			daily_window_start,
			weekly_window_start,
			monthly_window_start,
			daily_usage_usd,
			weekly_usage_usd,
			monthly_usage_usd
		)
		VALUES ($1, $2, NOW() - INTERVAL '40 days', NOW() + INTERVAL '20 days', 'active',
			date_trunc('day', NOW()), date_trunc('week', NOW()), NOW() - INTERVAL '31 days',
			0, 0, 149)
	`, userID, productID)
	require.NoError(t, err)

	items, err := repo.ListActiveProductsByUserID(ctx, userID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.InDelta(t, 0, items[0].Subscription.MonthlyUsageUSD, 0.000001)

	var wantMonthlyWindowStart time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT date_trunc('day', NOW())`).Scan(&wantMonthlyWindowStart))
	require.NotNil(t, items[0].Subscription.MonthlyWindowStart)
	require.True(t, items[0].Subscription.MonthlyWindowStart.Equal(wantMonthlyWindowStart), "monthly window = %s, want %s", items[0].Subscription.MonthlyWindowStart, wantMonthlyWindowStart)
}

func TestSubscriptionProductRepositoryListProductSubscriptionsNormalizesExpiredWindows(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

	var userID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, status, role, created_at, updated_at)
		VALUES ($1, 'hash', 'active', 'user', NOW(), NOW())
		RETURNING id
	`, uniqueTestValue(t, "product-detail-normalize-user")+"@example.com").Scan(&userID))

	var productID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, daily_limit_usd, monthly_limit_usd)
		VALUES ($1, 'detail normalize product', 'active', 45, 150)
		RETURNING id
	`, "detail-normalize-"+uuid.NewString()).Scan(&productID))

	_, err := integrationDB.ExecContext(ctx, `
		INSERT INTO user_product_subscriptions (
			user_id,
			product_id,
			starts_at,
			expires_at,
			status,
			daily_window_start,
			weekly_window_start,
			monthly_window_start,
			daily_usage_usd,
			weekly_usage_usd,
			monthly_usage_usd,
			daily_carryover_in_usd,
			daily_carryover_remaining_usd
		)
		VALUES ($1, $2, NOW() - INTERVAL '40 days', NOW() + INTERVAL '20 days', 'active',
			date_trunc('day', NOW()) - INTERVAL '1 day',
			date_trunc('week', NOW()),
			NOW() - INTERVAL '31 days',
			10, 20, 149, 5, 2)
	`, userID, productID)
	require.NoError(t, err)

	subs, err := repo.ListProductSubscriptions(ctx, productID)
	require.NoError(t, err)
	require.Len(t, subs, 1)

	var todayStart time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT date_trunc('day', NOW())`).Scan(&todayStart))

	got := subs[0]
	require.NotNil(t, got.DailyWindowStart)
	require.True(t, got.DailyWindowStart.Equal(todayStart), "daily_window_start = %s, want %s", got.DailyWindowStart, todayStart)
	require.InDelta(t, 0, got.DailyUsageUSD, 0.000001)
	require.InDelta(t, 38, got.DailyCarryoverInUSD, 0.000001)
	require.InDelta(t, 38, got.DailyCarryoverRemainingUSD, 0.000001)
	require.NotNil(t, got.MonthlyWindowStart)
	require.True(t, got.MonthlyWindowStart.Equal(todayStart), "monthly_window_start = %s, want %s", got.MonthlyWindowStart, todayStart)
	require.InDelta(t, 0, got.MonthlyUsageUSD, 0.000001)
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

func TestSubscriptionProductRepositoryListUserProductSubscriptionsForAdminNormalizesExpiredWindows(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

	user := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("product-sub-admin-normalize-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
	})

	var productID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, daily_limit_usd, monthly_limit_usd)
		VALUES ($1, 'admin normalize product', 'active', 45, 150)
		RETURNING id
	`, "admin-normalize-"+uuid.NewString()).Scan(&productID))

	_, err := integrationDB.ExecContext(ctx, `
		INSERT INTO user_product_subscriptions (
			user_id,
			product_id,
			starts_at,
			expires_at,
			status,
			daily_window_start,
			weekly_window_start,
			monthly_window_start,
			daily_usage_usd,
			weekly_usage_usd,
			monthly_usage_usd,
			daily_carryover_in_usd,
			daily_carryover_remaining_usd
		)
		VALUES ($1, $2, NOW() - INTERVAL '40 days', NOW() + INTERVAL '20 days', 'active',
			date_trunc('day', NOW()) - INTERVAL '1 day',
			date_trunc('week', NOW()),
			NOW() - INTERVAL '31 days',
			10, 20, 149, 5, 2)
	`, user.ID, productID)
	require.NoError(t, err)

	items, page, err := repo.ListUserProductSubscriptionsForAdmin(ctx, service.AdminProductSubscriptionListParams{
		Page:     1,
		PageSize: 20,
		UserID:   user.ID,
		Status:   service.SubscriptionStatusActive,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, items, 1)

	var todayStart time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT date_trunc('day', NOW())`).Scan(&todayStart))

	got := items[0]
	require.NotNil(t, got.DailyWindowStart)
	require.True(t, got.DailyWindowStart.Equal(todayStart), "daily_window_start = %s, want %s", got.DailyWindowStart, todayStart)
	require.InDelta(t, 0, got.DailyUsageUSD, 0.000001)
	require.InDelta(t, 38, got.DailyCarryoverInUSD, 0.000001)
	require.InDelta(t, 38, got.DailyCarryoverRemainingUSD, 0.000001)
	require.InDelta(t, 0, got.CarryoverUsedUSD, 0.000001)
	require.InDelta(t, 0, got.FreshDailyUsageUSD, 0.000001)
	require.NotNil(t, got.MonthlyWindowStart)
	require.True(t, got.MonthlyWindowStart.Equal(todayStart), "monthly_window_start = %s, want %s", got.MonthlyWindowStart, todayStart)
	require.InDelta(t, 0, got.MonthlyUsageUSD, 0.000001)
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

func TestSubscriptionProductRepositorySelectsEligibleProductByFamilyPriority(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

	var userID, groupID int64
	var err error
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, status, role, created_at, updated_at)
		VALUES ($1, 'hash', 'active', 'user', NOW(), NOW())
		RETURNING id
	`, uniqueTestValue(t, "product-select-user")+"@example.com").Scan(&userID))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO groups (name, status, subscription_type, created_at, updated_at)
		VALUES ($1, 'active', 'subscription', NOW(), NOW())
		RETURNING id
	`, uniqueTestValue(t, "product-select-group")).Scan(&groupID))

	var exhaustedProductID, nextProductID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, product_family, default_validity_days, daily_limit_usd, sort_order)
		VALUES ($1, '45 weekly', 'active', 'gpt_shared', 7, 45, 10)
		RETURNING id
	`, "select-exhausted-"+uuid.NewString()).Scan(&exhaustedProductID))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, product_family, default_validity_days, daily_limit_usd, sort_order)
		VALUES ($1, '150 monthly', 'active', 'gpt_shared', 30, 150, 20)
		RETURNING id
	`, "select-next-"+uuid.NewString()).Scan(&nextProductID))
	_, err = integrationDB.ExecContext(ctx, `
		INSERT INTO subscription_product_groups (product_id, group_id, debit_multiplier, status, sort_order)
		VALUES ($1, $3, 1, 'active', 1), ($2, $3, 1, 'active', 1)
	`, exhaustedProductID, nextProductID, groupID)
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = integrationDB.ExecContext(ctx, `
		INSERT INTO user_product_subscriptions (
			user_id, product_id, starts_at, expires_at, status,
			daily_window_start, weekly_window_start, monthly_window_start, daily_usage_usd
		)
		VALUES
			($1, $2, $4::timestamptz - INTERVAL '2 days', $4::timestamptz + INTERVAL '7 days', 'active', date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), 45),
			($1, $3, $4::timestamptz - INTERVAL '1 day', $4::timestamptz + INTERVAL '30 days', 'active', date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), 12)
	`, userID, exhaustedProductID, nextProductID, now)
	require.NoError(t, err)

	binding, sub, err := repo.GetActiveProductSubscriptionByUserAndGroupID(ctx, userID, groupID, nil)
	require.NoError(t, err)
	require.Equal(t, nextProductID, binding.ProductID)
	require.Equal(t, nextProductID, sub.ProductID)
}

func TestSubscriptionProductRepositoryDoesNotFallbackAcrossProductFamilies(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewSubscriptionProductRepository(client, integrationDB)

	var userID, groupID int64
	var err error
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, status, role, created_at, updated_at)
		VALUES ($1, 'hash', 'active', 'user', NOW(), NOW())
		RETURNING id
	`, uniqueTestValue(t, "product-family-user")+"@example.com").Scan(&userID))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO groups (name, status, subscription_type, created_at, updated_at)
		VALUES ($1, 'active', 'subscription', NOW(), NOW())
		RETURNING id
	`, uniqueTestValue(t, "product-family-group")).Scan(&groupID))

	var exhaustedProductID, otherFamilyProductID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, product_family, default_validity_days, daily_limit_usd, sort_order)
		VALUES ($1, 'exhausted family', 'active', 'gpt_shared', 7, 45, 10)
		RETURNING id
	`, "family-exhausted-"+uuid.NewString()).Scan(&exhaustedProductID))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO subscription_products (code, name, status, product_family, default_validity_days, daily_limit_usd, sort_order)
		VALUES ($1, 'other family', 'active', 'image_only', 30, 150, 20)
		RETURNING id
	`, "family-other-"+uuid.NewString()).Scan(&otherFamilyProductID))
	_, err = integrationDB.ExecContext(ctx, `
		INSERT INTO subscription_product_groups (product_id, group_id, debit_multiplier, status, sort_order)
		VALUES ($1, $3, 1, 'active', 1), ($2, $3, 1, 'active', 1)
	`, exhaustedProductID, otherFamilyProductID, groupID)
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = integrationDB.ExecContext(ctx, `
		INSERT INTO user_product_subscriptions (
			user_id, product_id, starts_at, expires_at, status,
			daily_window_start, weekly_window_start, monthly_window_start, daily_usage_usd
		)
		VALUES
			($1, $2, $4::timestamptz - INTERVAL '2 days', $4::timestamptz + INTERVAL '7 days', 'active', date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), 45),
			($1, $3, $4::timestamptz - INTERVAL '1 day', $4::timestamptz + INTERVAL '30 days', 'active', date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), date_trunc('day', $4::timestamptz), 12)
	`, userID, exhaustedProductID, otherFamilyProductID, now)
	require.NoError(t, err)

	_, _, err = repo.GetActiveProductSubscriptionByUserAndGroupID(ctx, userID, groupID, nil)
	require.ErrorIs(t, err, service.ErrDailyLimitExceeded)
}

func productIDs(products []service.SubscriptionProduct) []int64 {
	out := make([]int64, 0, len(products))
	for _, product := range products {
		out = append(out, product.ID)
	}
	return out
}
