package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("XLAB_SERVER_ADDR", "")
	t.Setenv("CORE_API_BASE_URL", "")
	t.Setenv("XLAB_CORE_TIMEOUT_SECONDS", "")
	t.Setenv("XLAB_DATABASE_URL", "")
	t.Setenv("CORE_DATABASE_URL", "")
	t.Setenv("XLAB_SUBSCRIPTION_READ_SOURCE", "")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS", "")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS", "")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_ENABLED", "")
	t.Setenv("XLAB_PAYMENT_READ_SOURCE", "")
	t.Setenv("XLAB_PAYMENT_SYNC_STALE_SECONDS", "")
	t.Setenv("XLAB_PAYMENT_SYNC_INTERVAL_SECONDS", "")
	t.Setenv("XLAB_PAYMENT_SYNC_ENABLED", "")

	cfg := Load()
	if cfg.ServerAddr != ":8090" {
		t.Fatalf("ServerAddr = %q, want :8090", cfg.ServerAddr)
	}
	if cfg.CoreAPIBaseURL != "http://127.0.0.1:8080/api/v1" {
		t.Fatalf("CoreAPIBaseURL = %q", cfg.CoreAPIBaseURL)
	}
	if cfg.CoreTimeout != 10*time.Second {
		t.Fatalf("CoreTimeout = %s", cfg.CoreTimeout)
	}
	if cfg.XlabDatabaseURL != "" {
		t.Fatalf("XlabDatabaseURL = %q, want empty", cfg.XlabDatabaseURL)
	}
	if cfg.CoreDatabaseURL != "" {
		t.Fatalf("CoreDatabaseURL = %q, want empty", cfg.CoreDatabaseURL)
	}
	if cfg.SubscriptionReadSource != SubscriptionReadSourceCore {
		t.Fatalf("SubscriptionReadSource = %q, want core", cfg.SubscriptionReadSource)
	}
	if cfg.SubscriptionSyncEnabled {
		t.Fatal("SubscriptionSyncEnabled should default false")
	}
	if cfg.SubscriptionSyncStaleAfter != 10*time.Minute {
		t.Fatalf("SubscriptionSyncStaleAfter = %s", cfg.SubscriptionSyncStaleAfter)
	}
	if cfg.SubscriptionSyncInterval != 5*time.Minute {
		t.Fatalf("SubscriptionSyncInterval = %s", cfg.SubscriptionSyncInterval)
	}
	if cfg.PaymentReadSource != PaymentReadSourceCore {
		t.Fatalf("PaymentReadSource = %q, want core", cfg.PaymentReadSource)
	}
	if cfg.PaymentSyncEnabled {
		t.Fatal("PaymentSyncEnabled should default false")
	}
	if cfg.PaymentSyncStaleAfter != 10*time.Minute {
		t.Fatalf("PaymentSyncStaleAfter = %s", cfg.PaymentSyncStaleAfter)
	}
	if cfg.PaymentSyncInterval != 5*time.Minute {
		t.Fatalf("PaymentSyncInterval = %s", cfg.PaymentSyncInterval)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("XLAB_SERVER_ADDR", ":19090")
	t.Setenv("CORE_API_BASE_URL", "https://core.example.com/api/v1/")
	t.Setenv("XLAB_CORE_TIMEOUT_SECONDS", "3")
	t.Setenv("XLAB_DATABASE_URL", "postgres://xlab:secret@db/xlab?sslmode=disable")
	t.Setenv("CORE_DATABASE_URL", "postgres://core:secret@db/core?sslmode=disable")
	t.Setenv("XLAB_SUBSCRIPTION_READ_SOURCE", "hybrid")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS", "120")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS", "30")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_ENABLED", "true")
	t.Setenv("XLAB_PAYMENT_READ_SOURCE", "xlab")
	t.Setenv("XLAB_PAYMENT_SYNC_STALE_SECONDS", "45")
	t.Setenv("XLAB_PAYMENT_SYNC_INTERVAL_SECONDS", "15")
	t.Setenv("XLAB_PAYMENT_SYNC_ENABLED", "true")

	cfg := Load()
	if cfg.ServerAddr != ":19090" {
		t.Fatalf("ServerAddr = %q", cfg.ServerAddr)
	}
	if cfg.CoreAPIBaseURL != "https://core.example.com/api/v1" {
		t.Fatalf("CoreAPIBaseURL = %q", cfg.CoreAPIBaseURL)
	}
	if cfg.CoreTimeout != 3*time.Second {
		t.Fatalf("CoreTimeout = %s", cfg.CoreTimeout)
	}
	if cfg.XlabDatabaseURL != "postgres://xlab:secret@db/xlab?sslmode=disable" {
		t.Fatalf("XlabDatabaseURL = %q", cfg.XlabDatabaseURL)
	}
	if cfg.CoreDatabaseURL != "postgres://core:secret@db/core?sslmode=disable" {
		t.Fatalf("CoreDatabaseURL = %q", cfg.CoreDatabaseURL)
	}
	if cfg.SubscriptionReadSource != SubscriptionReadSourceHybrid {
		t.Fatalf("SubscriptionReadSource = %q", cfg.SubscriptionReadSource)
	}
	if !cfg.SubscriptionSyncEnabled {
		t.Fatal("SubscriptionSyncEnabled should be true")
	}
	if cfg.SubscriptionSyncStaleAfter != 2*time.Minute {
		t.Fatalf("SubscriptionSyncStaleAfter = %s", cfg.SubscriptionSyncStaleAfter)
	}
	if cfg.SubscriptionSyncInterval != 30*time.Second {
		t.Fatalf("SubscriptionSyncInterval = %s", cfg.SubscriptionSyncInterval)
	}
	if cfg.PaymentReadSource != PaymentReadSourceXlab {
		t.Fatalf("PaymentReadSource = %q", cfg.PaymentReadSource)
	}
	if !cfg.PaymentSyncEnabled {
		t.Fatal("PaymentSyncEnabled should be true")
	}
	if cfg.PaymentSyncStaleAfter != 45*time.Second {
		t.Fatalf("PaymentSyncStaleAfter = %s", cfg.PaymentSyncStaleAfter)
	}
	if cfg.PaymentSyncInterval != 15*time.Second {
		t.Fatalf("PaymentSyncInterval = %s", cfg.PaymentSyncInterval)
	}
}

func TestLoadRejectsUnknownSubscriptionReadSource(t *testing.T) {
	t.Setenv("XLAB_SUBSCRIPTION_READ_SOURCE", "mirror")

	cfg := Load()
	if cfg.SubscriptionReadSource != SubscriptionReadSourceCore {
		t.Fatalf("SubscriptionReadSource = %q, want core", cfg.SubscriptionReadSource)
	}
}

func TestLoadRejectsUnknownPaymentReadSource(t *testing.T) {
	t.Setenv("XLAB_PAYMENT_READ_SOURCE", "mirror")

	cfg := Load()
	if cfg.PaymentReadSource != PaymentReadSourceCore {
		t.Fatalf("PaymentReadSource = %q, want core", cfg.PaymentReadSource)
	}
}
