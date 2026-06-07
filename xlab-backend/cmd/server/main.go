package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/core"
	"github.com/2018x5zzt/xlab-backend/internal/httpapi"
	"github.com/2018x5zzt/xlab-backend/internal/storage"
	"github.com/2018x5zzt/xlab-backend/internal/subscriptions"
)

func main() {
	cfg := config.Load()
	client := core.NewClient(cfg.CoreAPIBaseURL, cfg.CoreTimeout)
	router := httpapi.NewRouter(client, nil)

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

		repo := subscriptions.NewRepository(xlabDB)
		readSvc := subscriptions.NewService(subscriptionReadSource(cfg.SubscriptionReadSource), repo, client, cfg.SubscriptionSyncStaleAfter, time.Now)
		router = httpapi.NewRouter(client, readSvc)

		if cfg.SubscriptionSyncEnabled {
			if cfg.CoreDatabaseURL == "" {
				log.Fatal("CORE_DATABASE_URL is required when XLAB_SUBSCRIPTION_SYNC_ENABLED=true")
			}
			coreDB, err := storage.OpenPostgres(ctx, cfg.CoreDatabaseURL)
			if err != nil {
				log.Fatalf("open core db: %v", err)
			}
			defer coreDB.Close()
			syncer := subscriptions.NewSyncer(coreDB, xlabDB, time.Now)
			go runSubscriptionSyncLoop(ctx, syncer, cfg.SubscriptionSyncInterval)
		}
	}

	log.Printf("xlab backend listening on %s, core=%s", cfg.ServerAddr, cfg.CoreAPIBaseURL)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
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
