package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type SubscriptionReadSource string

const (
	SubscriptionReadSourceCore   SubscriptionReadSource = "core"
	SubscriptionReadSourceHybrid SubscriptionReadSource = "hybrid"
	SubscriptionReadSourceXlab   SubscriptionReadSource = "xlab"
)

type PaymentReadSource string

const (
	PaymentReadSourceCore   PaymentReadSource = "core"
	PaymentReadSourceHybrid PaymentReadSource = "hybrid"
	PaymentReadSourceXlab   PaymentReadSource = "xlab"
)

type Config struct {
	ServerAddr                 string
	CoreAPIBaseURL             string
	CoreTimeout                time.Duration
	XlabDatabaseURL            string
	CoreDatabaseURL            string
	SubscriptionReadSource     SubscriptionReadSource
	SubscriptionSyncEnabled    bool
	SubscriptionSyncInterval   time.Duration
	SubscriptionSyncStaleAfter time.Duration
	PaymentReadSource          PaymentReadSource
	PaymentSyncEnabled         bool
	PaymentSyncInterval        time.Duration
	PaymentSyncStaleAfter      time.Duration
}

func Load() Config {
	addr := strings.TrimSpace(os.Getenv("XLAB_SERVER_ADDR"))
	if addr == "" {
		addr = ":8090"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("CORE_API_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080/api/v1"
	}

	timeoutSeconds := positiveIntEnv("XLAB_CORE_TIMEOUT_SECONDS", 10)

	return Config{
		ServerAddr:                 addr,
		CoreAPIBaseURL:             baseURL,
		CoreTimeout:                time.Duration(timeoutSeconds) * time.Second,
		XlabDatabaseURL:            strings.TrimSpace(os.Getenv("XLAB_DATABASE_URL")),
		CoreDatabaseURL:            strings.TrimSpace(os.Getenv("CORE_DATABASE_URL")),
		SubscriptionReadSource:     parseReadSource(os.Getenv("XLAB_SUBSCRIPTION_READ_SOURCE")),
		SubscriptionSyncEnabled:    boolEnv("XLAB_SUBSCRIPTION_SYNC_ENABLED", false),
		SubscriptionSyncInterval:   time.Duration(positiveIntEnv("XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS", 300)) * time.Second,
		SubscriptionSyncStaleAfter: time.Duration(positiveIntEnv("XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS", 600)) * time.Second,
		PaymentReadSource:          parsePaymentReadSource(os.Getenv("XLAB_PAYMENT_READ_SOURCE")),
		PaymentSyncEnabled:         boolEnv("XLAB_PAYMENT_SYNC_ENABLED", false),
		PaymentSyncInterval:        time.Duration(positiveIntEnv("XLAB_PAYMENT_SYNC_INTERVAL_SECONDS", 300)) * time.Second,
		PaymentSyncStaleAfter:      time.Duration(positiveIntEnv("XLAB_PAYMENT_SYNC_STALE_SECONDS", 600)) * time.Second,
	}
}

func parseReadSource(raw string) SubscriptionReadSource {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(SubscriptionReadSourceHybrid):
		return SubscriptionReadSourceHybrid
	case string(SubscriptionReadSourceXlab):
		return SubscriptionReadSourceXlab
	default:
		return SubscriptionReadSourceCore
	}
}

func parsePaymentReadSource(raw string) PaymentReadSource {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(PaymentReadSourceHybrid):
		return PaymentReadSourceHybrid
	case string(PaymentReadSourceXlab):
		return PaymentReadSourceXlab
	default:
		return PaymentReadSourceCore
	}
}

func positiveIntEnv(name string, fallback int) int {
	if raw := strings.TrimSpace(os.Getenv(name)); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func boolEnv(name string, fallback bool) bool {
	if raw := strings.TrimSpace(os.Getenv(name)); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			return parsed
		}
	}
	return fallback
}
