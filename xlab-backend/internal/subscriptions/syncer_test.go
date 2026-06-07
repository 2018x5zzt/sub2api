package subscriptions

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSyncerFullSnapshotUpdatesSyncState(t *testing.T) {
	coreDB, coreMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("core sqlmock.New error: %v", err)
	}
	defer coreDB.Close()
	xlabDB, xlabMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("xlab sqlmock.New error: %v", err)
	}
	defer xlabDB.Close()

	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	coreMock.ExpectQuery(regexp.QuoteMeta(coreProductsSQL)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "code", "name", "description", "status", "product_family", "daily_limit_usd", "weekly_limit_usd", "monthly_limit_usd", "created_at", "updated_at"}).
			AddRow(int64(10), "gpt-pro", "GPT Pro", "desc", "active", "gpt", 10.0, 20.0, 30.0, now, now),
	)
	coreMock.ExpectQuery(regexp.QuoteMeta(coreProductGroupsSQL)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "product_id", "group_id", "group_name", "group_platform", "balance_fallback_group_id", "balance_fallback_group_name", "debit_multiplier", "status", "sort_order", "created_at", "updated_at"}).
			AddRow(int64(100), int64(10), int64(20), "GPT-4", "openai", int64(30), "GPT Balance", 1.0, "active", 1, now, now),
	)
	coreMock.ExpectQuery(regexp.QuoteMeta(coreUserProductSubscriptionsSQL)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "product_id", "status", "starts_at", "expires_at", "daily_usage_usd", "weekly_usage_usd", "monthly_usage_usd", "daily_carryover_in_usd", "daily_carryover_remaining_usd", "created_at", "updated_at"}).
			AddRow(int64(99), int64(7), int64(10), "active", now, now.Add(24*time.Hour), 1.0, 2.0, 3.0, 0.0, 0.0, now, now),
	)

	xlabMock.ExpectBegin()
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertProductSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertProductGroupSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertUserProductSubscriptionSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertSyncStateSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectCommit()

	syncer := NewSyncer(coreDB, xlabDB, func() time.Time { return now })
	if err := syncer.SyncOnce(context.Background()); err != nil {
		t.Fatalf("SyncOnce error: %v", err)
	}
	if err := coreMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet core expectations: %v", err)
	}
	if err := xlabMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet xlab expectations: %v", err)
	}
}
