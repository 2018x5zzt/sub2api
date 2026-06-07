package subscriptions

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryListActiveProductsByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	productRows := sqlmock.NewRows([]string{
		"core_subscription_id", "core_product_id", "code", "name", "description", "status", "expires_at",
		"daily_usage_usd", "weekly_usage_usd", "monthly_usage_usd",
		"daily_limit_usd", "weekly_limit_usd", "monthly_limit_usd",
		"daily_carryover_in_usd", "daily_carryover_remaining_usd",
	}).AddRow(int64(99), int64(10), "gpt-pro", "GPT Pro", "desc", "active", expires, 1.25, 2.5, 3.75, 10.0, 20.0, 30.0, 0.5, 0.25)
	groupRows := sqlmock.NewRows([]string{"core_product_id", "core_group_id", "group_name", "group_platform", "balance_fallback_group_id", "balance_fallback_group_name", "debit_multiplier", "status", "sort_order"}).
		AddRow(int64(10), int64(20), "GPT-4", "openai", int64(30), "GPT Balance", 1.2, "active", 1)

	mock.ExpectQuery(regexp.QuoteMeta(activeProductsByUserSQL)).WithArgs(int64(7)).WillReturnRows(productRows)
	mock.ExpectQuery(regexp.QuoteMeta(groupsByProductSQL)).WithArgs(sqlmock.AnyArg()).WillReturnRows(groupRows)

	repo := NewRepository(db)
	items, err := repo.ListActiveProductsByUser(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListActiveProductsByUser error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d", len(items))
	}
	item := items[0]
	if item.SubscriptionID != 99 || item.ProductID != 10 || item.Name != "GPT Pro" || item.Description != "desc" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if len(item.Groups) != 1 || item.Groups[0].GroupName != "GPT-4" || item.Groups[0].DebitMultiplier != 1.2 {
		t.Fatalf("unexpected groups: %+v", item.Groups)
	}
	if item.Groups[0].BalanceFallbackGroupID == nil || *item.Groups[0].BalanceFallbackGroupID != 30 {
		t.Fatalf("BalanceFallbackGroupID = %v", item.Groups[0].BalanceFallbackGroupID)
	}
	if item.Groups[0].BalanceFallbackGroupName == nil || *item.Groups[0].BalanceFallbackGroupName != "GPT Balance" {
		t.Fatalf("BalanceFallbackGroupName = %v", item.Groups[0].BalanceFallbackGroupName)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositorySummaryAggregatesMonthlyUsageAndLimit(t *testing.T) {
	items := []ActiveProduct{
		{MonthlyUsageUSD: 1.5, MonthlyLimitUSD: 10},
		{MonthlyUsageUSD: 2.5, MonthlyLimitUSD: 20},
	}
	summary := SummaryFromActiveProducts(items)
	if summary.ActiveCount != 2 || summary.TotalMonthlyUsageUSD != 4.0 || summary.TotalMonthlyLimitUSD != 30.0 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestRepositorySyncState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(syncStateSQL)).WithArgs("product_subscriptions").WillReturnRows(
		sqlmock.NewRows([]string{"source_name", "last_success_at", "row_count"}).AddRow("product_subscriptions", now, 3),
	)
	repo := NewRepository(db)
	state, err := repo.SyncState(context.Background(), "product_subscriptions")
	if err != nil {
		t.Fatalf("SyncState error: %v", err)
	}
	if state.LastSuccessAt == nil || state.RowCount != 3 {
		t.Fatalf("state = %+v", state)
	}
}
