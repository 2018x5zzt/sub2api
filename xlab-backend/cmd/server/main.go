package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/core"
	"github.com/2018x5zzt/xlab-backend/internal/httpapi"
	"github.com/2018x5zzt/xlab-backend/internal/payments"
	"github.com/2018x5zzt/xlab-backend/internal/storage"
	"github.com/2018x5zzt/xlab-backend/internal/subscriptions"
)

func main() {
	cfg := config.Load()
	if err := validateSyncConfig(cfg); err != nil {
		log.Fatal(err)
	}
	client := core.NewClient(cfg.CoreAPIBaseURL, cfg.CoreTimeout)
	router := httpapi.NewRouter(client, nil, nil)

	ctx := context.Background()
	if cfg.XlabDatabaseURL != "" {
		xlabDB, err := storage.OpenPostgres(ctx, cfg.XlabDatabaseURL)
		if err != nil {
			log.Fatalf("open xlab db: %v", err)
		}
		defer xlabDB.Close()
		if err := storage.RunMigrations(ctx, xlabDB); err != nil {
			log.Fatalf("run xlab migrations: %v", err)
		}

		subscriptionRepo := subscriptions.NewRepository(xlabDB)
		readSvc := subscriptions.NewService(subscriptionReadSource(cfg.SubscriptionReadSource), subscriptionRepo, client, cfg.SubscriptionSyncStaleAfter, time.Now)
		paymentRepo := payments.NewRepository(xlabDB)
		paymentReadSvc := payments.NewService(paymentReadSource(cfg.PaymentReadSource), paymentRepo, client, cfg.PaymentSyncStaleAfter, time.Now)
		router = httpapi.NewRouter(client, readSvc, paymentReadSvc)

		if cfg.PaymentSyncEnabled && cfg.CoreDatabaseURL == "" {
			log.Fatal("CORE_DATABASE_URL is required when XLAB_PAYMENT_SYNC_ENABLED=true")
		}
		if cfg.SubscriptionSyncEnabled && cfg.CoreDatabaseURL == "" {
			log.Fatal("CORE_DATABASE_URL is required when XLAB_SUBSCRIPTION_SYNC_ENABLED=true")
		}
		if cfg.SubscriptionSyncEnabled || cfg.PaymentSyncEnabled {
			coreDB, err := storage.OpenPostgres(ctx, cfg.CoreDatabaseURL)
			if err != nil {
				log.Fatalf("open core db: %v", err)
			}
			defer coreDB.Close()
			if cfg.SubscriptionSyncEnabled {
				syncer := subscriptions.NewSyncer(coreDB, xlabDB, time.Now)
				go runSubscriptionSyncLoop(ctx, syncer, cfg.SubscriptionSyncInterval)
			}
			if cfg.PaymentSyncEnabled {
				syncer := payments.NewSyncer(coreDB, xlabDB, time.Now)
				go runPaymentSyncLoop(ctx, syncer, cfg.PaymentSyncInterval)
			}
		}
	}

	log.Printf("xlab backend listening on %s, core=%s", cfg.ServerAddr, cfg.CoreAPIBaseURL)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
}

func validateSyncConfig(cfg config.Config) error {
	if (cfg.SubscriptionSyncEnabled || cfg.PaymentSyncEnabled) && cfg.XlabDatabaseURL == "" {
		return errors.New("XLAB_DATABASE_URL is required when subscription or payment sync is enabled")
	}
	return nil
}

func subscriptionReadSource(source config.SubscriptionReadSource) subscriptions.ReadSource {
	switch source {
	case config.SubscriptionReadSourceHybrid:
		return subscriptions.ReadSourceHybrid
	case config.SubscriptionReadSourceXlab:
		return subscriptions.ReadSourceXlab
	default:
		return subscriptions.ReadSourceCore
	}
}

func paymentReadSource(source config.PaymentReadSource) payments.ReadSource {
	switch source {
	case config.PaymentReadSourceHybrid:
		return payments.ReadSourceHybrid
	case config.PaymentReadSourceXlab:
		return payments.ReadSourceXlab
	default:
		return payments.ReadSourceCore
	}
}

func runSubscriptionSyncLoop(ctx context.Context, syncer *subscriptions.Syncer, interval time.Duration) {
	if err := syncer.SyncOnce(ctx); err != nil {
		log.Printf("product subscription sync failed: %v", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := syncer.SyncOnce(ctx); err != nil {
				log.Printf("product subscription sync failed: %v", err)
			}
		}
	}
}

func runPaymentSyncLoop(ctx context.Context, syncer *payments.Syncer, interval time.Duration) {
	if err := syncer.SyncOnce(ctx); err != nil {
		log.Printf("payment order sync failed: %v", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := syncer.SyncOnce(ctx); err != nil {
				log.Printf("payment order sync failed: %v", err)
			}
		}
	}
}
