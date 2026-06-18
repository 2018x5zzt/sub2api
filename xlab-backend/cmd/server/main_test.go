package main

import (
	"strings"
	"testing"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/payments"
	"github.com/2018x5zzt/xlab-backend/internal/subscriptions"
)

func TestSubscriptionReadSourceMapping(t *testing.T) {
	cases := []struct {
		in   config.SubscriptionReadSource
		want subscriptions.ReadSource
	}{
		{config.SubscriptionReadSourceCore, subscriptions.ReadSourceCore},
		{config.SubscriptionReadSourceHybrid, subscriptions.ReadSourceHybrid},
		{config.SubscriptionReadSourceXlab, subscriptions.ReadSourceXlab},
	}
	for _, tc := range cases {
		if got := subscriptionReadSource(tc.in); got != tc.want {
			t.Fatalf("subscriptionReadSource(%q) = %q", tc.in, got)
		}
	}
}

func TestValidateSyncConfigRequiresXlabDatabaseForSubscriptionSync(t *testing.T) {
	err := validateSyncConfig(config.Config{SubscriptionSyncEnabled: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "XLAB_DATABASE_URL") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateSyncConfigRequiresXlabDatabaseForPaymentSync(t *testing.T) {
	err := validateSyncConfig(config.Config{PaymentSyncEnabled: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "XLAB_DATABASE_URL") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateSyncConfigAllowsNoSyncWithoutXlabDatabase(t *testing.T) {
	if err := validateSyncConfig(config.Config{}); err != nil {
		t.Fatalf("validateSyncConfig returned error: %v", err)
	}
}

func TestPaymentReadSourceMapping(t *testing.T) {
	cases := []struct {
		in   config.PaymentReadSource
		want payments.ReadSource
	}{
		{config.PaymentReadSourceCore, payments.ReadSourceCore},
		{config.PaymentReadSourceHybrid, payments.ReadSourceHybrid},
		{config.PaymentReadSourceXlab, payments.ReadSourceXlab},
	}
	for _, tc := range cases {
		if got := paymentReadSource(tc.in); got != tc.want {
			t.Fatalf("paymentReadSource(%q) = %q", tc.in, got)
		}
	}
}
