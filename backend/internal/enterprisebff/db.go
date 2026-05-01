package enterprisebff

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	coreconfig "github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
)

type DBResources struct {
	Client *ent.Client
	SQLDB  *sql.DB
}

func OpenDB(cfg *coreconfig.Config) (*DBResources, error) {
	if err := timezone.Init(cfg.Timezone); err != nil {
		return nil, fmt.Errorf("init timezone: %w", err)
	}

	dsn := cfg.Database.DSNWithTimezone(cfg.Timezone)
	drv, err := entsql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB := drv.DB()
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeMinutes) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.ConnMaxIdleTimeMinutes) * time.Minute)

	return &DBResources{
		Client: ent.NewClient(ent.Driver(drv)),
		SQLDB:  sqlDB,
	}, nil
}

func (r *DBResources) Close() error {
	if r == nil || r.Client == nil {
		return nil
	}
	return r.Client.Close()
}
