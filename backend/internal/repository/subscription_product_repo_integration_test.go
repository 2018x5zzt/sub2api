//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionProductRepository_GetActiveProductByGroupID(t *testing.T) {
	ctx := context.Background()
	group := mustCreateGroup(t, integrationEntClient, &service.Group{
		Name:             uniqueProductTestName("product-group"),
		Platform:         service.PlatformOpenAI,
		SubscriptionType: service.SubscriptionTypeSubscription,
		Status:           service.StatusActive,
		RateMultiplier:   1,
	})
	product := mustCreateSubscriptionProduct(t, "gpt_monthly_binding", "GPT Monthly", "active")
	mustCreateSubscriptionProductGroup(t, product.ID, group.ID, 1.5, "active", 10)

	repo := NewSubscriptionProductRepository(integrationEntClient, integrationDB)
	got, err := repo.GetActiveProductBindingByGroupID(ctx, group.ID)

	require.NoError(t, err)
	require.Equal(t, product.ID, got.ProductID)
	require.Equal(t, product.Code, got.ProductCode)
	require.Equal(t, product.Name, got.ProductName)
	require.Equal(t, group.ID, got.GroupID)
	require.Equal(t, group.Name, got.GroupName)
	require.Equal(t, 1.5, got.DebitMultiplier)
	require.Equal(t, "active", got.ProductStatus)
	require.Equal(t, "active", got.BindingStatus)
}

func TestSubscriptionProductRepository_ListVisibleGroupsByUserID(t *testing.T) {
	ctx := context.Background()
	user := mustCreateUser(t, integrationEntClient, &service.User{
		Email: uniqueProductTestName("visible-user") + "@example.com",
	})
	plus := mustCreateGroup(t, integrationEntClient, &service.Group{
		Name:             uniqueProductTestName("plus"),
		Platform:         service.PlatformOpenAI,
		SubscriptionType: service.SubscriptionTypeSubscription,
		Status:           service.StatusActive,
		RateMultiplier:   1,
	})
	pro := mustCreateGroup(t, integrationEntClient, &service.Group{
		Name:             uniqueProductTestName("pro"),
		Platform:         service.PlatformOpenAI,
		SubscriptionType: service.SubscriptionTypeSubscription,
		Status:           service.StatusActive,
		RateMultiplier:   1,
	})
	product := mustCreateSubscriptionProduct(t, "gpt_monthly_visible", "GPT Monthly", "active")
	mustCreateSubscriptionProductGroup(t, product.ID, plus.ID, 1.0, "active", 10)
	mustCreateSubscriptionProductGroup(t, product.ID, pro.ID, 1.5, "active", 20)
	mustCreateUserProductSubscription(t, user.ID, product.ID, "active", time.Now().Add(24*time.Hour))

	repo := NewSubscriptionProductRepository(integrationEntClient, integrationDB)
	groups, err := repo.ListVisibleGroupsByUserID(ctx, user.ID)

	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, []int64{plus.ID, pro.ID}, []int64{groups[0].ID, groups[1].ID})
	require.Equal(t, service.SubscriptionTypeSubscription, groups[0].SubscriptionType)
	require.Equal(t, service.PlatformOpenAI, groups[1].Platform)
}

func mustCreateSubscriptionProduct(t *testing.T, codePrefix, name, status string) *service.SubscriptionProduct {
	t.Helper()
	now := time.Now()
	code := uniqueProductTestName(codePrefix)
	created, err := integrationEntClient.SubscriptionProduct.Create().
		SetCode(code).
		SetName(name).
		SetStatus(status).
		SetDefaultValidityDays(30).
		SetDailyLimitUsd(10).
		SetWeeklyLimitUsd(50).
		SetMonthlyLimitUsd(100).
		SetSortOrder(10).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(context.Background())
	require.NoError(t, err, "create subscription product")
	return &service.SubscriptionProduct{
		ID:                  created.ID,
		Code:                created.Code,
		Name:                created.Name,
		Status:              created.Status,
		DefaultValidityDays: created.DefaultValidityDays,
		DailyLimitUSD:       created.DailyLimitUsd,
		WeeklyLimitUSD:      created.WeeklyLimitUsd,
		MonthlyLimitUSD:     created.MonthlyLimitUsd,
		SortOrder:           created.SortOrder,
		CreatedAt:           created.CreatedAt,
		UpdatedAt:           created.UpdatedAt,
	}
}

func mustCreateSubscriptionProductGroup(t *testing.T, productID, groupID int64, multiplier float64, status string, sortOrder int) {
	t.Helper()
	_, err := integrationEntClient.SubscriptionProductGroup.Create().
		SetProductID(productID).
		SetGroupID(groupID).
		SetDebitMultiplier(multiplier).
		SetStatus(status).
		SetSortOrder(sortOrder).
		Save(context.Background())
	require.NoError(t, err, "create subscription product group")
}

func mustCreateUserProductSubscription(t *testing.T, userID, productID int64, status string, expiresAt time.Time) {
	t.Helper()
	now := time.Now()
	_, err := integrationEntClient.UserProductSubscription.Create().
		SetUserID(userID).
		SetProductID(productID).
		SetStartsAt(now.Add(-time.Hour)).
		SetExpiresAt(expiresAt).
		SetStatus(status).
		SetAssignedAt(now).
		SetDailyCarryoverInUsd(2).
		SetDailyCarryoverRemainingUsd(1).
		Save(context.Background())
	require.NoError(t, err, "create user product subscription")
}

func uniqueProductTestName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
