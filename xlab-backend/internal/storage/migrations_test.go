package storage

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrationSQLContainsPhase3MirrorTables(t *testing.T) {
	sql, err := migrationSQL()
	if err != nil {
		t.Fatalf("migrationSQL error: %v", err)
	}
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS xlab_subscription_products",
		"core_product_id BIGINT PRIMARY KEY",
		"CREATE TABLE IF NOT EXISTS xlab_subscription_product_groups",
		"core_product_id BIGINT NOT NULL REFERENCES xlab_subscription_products(core_product_id) ON DELETE CASCADE",
		"CREATE TABLE IF NOT EXISTS xlab_user_product_subscriptions",
		"CREATE INDEX IF NOT EXISTS idx_xlab_user_product_subscriptions_user_active",
		"CREATE TABLE IF NOT EXISTS xlab_sync_state",
		"source_name TEXT PRIMARY KEY",
		"updated_at TIMESTAMPTZ NOT NULL DEFAULT now()",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration SQL missing %s", fragment)
		}
	}
}

func TestRunMigrationsWrapsExecutionErrorsWithMigrationName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	migrationErr := errors.New("boom")
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS xlab_subscription_products").WillReturnError(migrationErr)

	err = RunMigrations(context.Background(), db)
	if err == nil {
		t.Fatal("expected RunMigrations error")
	}
	if !strings.Contains(err.Error(), "execute migration 001_product_subscription_read_mirror.sql") {
		t.Fatalf("error = %q, want migration execution context", err.Error())
	}
	if !errors.Is(err, migrationErr) {
		t.Fatalf("error should wrap migration error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
