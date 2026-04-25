package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type validateOptions struct {
	ProductCode    string
	SourceGroupIDs []int64
	MigrationBatch string
	DSN            string
}

func main() {
	opts, err := parseValidateFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := runValidate(context.Background(), opts); err != nil {
		fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
		os.Exit(1)
	}
}

func parseValidateFlags(args []string) (validateOptions, error) {
	var sourceGroupIDs string
	opts := validateOptions{}
	fs := flag.NewFlagSet("shared-subscription-products-validate", flag.ContinueOnError)
	fs.StringVar(&opts.ProductCode, "product-code", "", "subscription product code")
	fs.StringVar(&sourceGroupIDs, "source-group-ids", "", "comma-separated legacy source group IDs")
	fs.StringVar(&opts.MigrationBatch, "migration-batch", "", "stable migration batch identifier")
	fs.StringVar(&opts.DSN, "dsn", "", "PostgreSQL DSN override")
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	var err error
	opts.SourceGroupIDs, err = parseInt64CSV(sourceGroupIDs)
	if err != nil {
		return opts, err
	}
	if opts.ProductCode == "" {
		return opts, errors.New("--product-code is required")
	}
	if len(opts.SourceGroupIDs) == 0 {
		return opts, errors.New("--source-group-ids is required")
	}
	if opts.MigrationBatch == "" {
		return opts, errors.New("--migration-batch is required")
	}
	return opts, nil
}

func runValidate(ctx context.Context, opts validateOptions) error {
	if opts.DSN == "" {
		cfg, err := config.LoadForBootstrap()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		opts.DSN = cfg.Database.DSNWithTimezone(cfg.Timezone)
	}
	db, err := sql.Open("postgres", opts.DSN)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	var productID int64
	if err := db.QueryRowContext(ctx, `
SELECT id
FROM subscription_products
WHERE code = $1
  AND deleted_at IS NULL
`, opts.ProductCode).Scan(&productID); err != nil {
		return fmt.Errorf("resolve product %q: %w", opts.ProductCode, err)
	}

	var migratedRows int64
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM user_product_subscriptions
WHERE product_id = $1
  AND deleted_at IS NULL
`, productID).Scan(&migratedRows); err != nil {
		return fmt.Errorf("count migrated rows: %w", err)
	}

	var auditRows int64
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM product_subscription_migration_sources
WHERE migration_batch = $1
`, opts.MigrationBatch).Scan(&auditRows); err != nil {
		return fmt.Errorf("count audit rows: %w", err)
	}

	var duplicateActiveRows int64
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM (
  SELECT user_id, product_id, COUNT(*) AS count
  FROM user_product_subscriptions
  WHERE deleted_at IS NULL
    AND status = 'active'
    AND product_id = $1
  GROUP BY user_id, product_id
  HAVING COUNT(*) > 1
) duplicates
`, productID).Scan(&duplicateActiveRows); err != nil {
		return fmt.Errorf("count duplicate active product subscriptions: %w", err)
	}

	var sourceRows int64
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM user_subscriptions
WHERE deleted_at IS NULL
  AND group_id = ANY($1)
  AND status = 'active'
`, pq.Array(opts.SourceGroupIDs)).Scan(&sourceRows); err != nil {
		return fmt.Errorf("count source rows: %w", err)
	}

	fmt.Printf("product_code=%s product_id=%d migration_batch=%s\n", opts.ProductCode, productID, opts.MigrationBatch)
	fmt.Printf("source_active_rows=%d migrated_rows=%d audit_rows=%d duplicate_active_product_rows=%d\n", sourceRows, migratedRows, auditRows, duplicateActiveRows)
	if duplicateActiveRows > 0 {
		return fmt.Errorf("duplicate active user_product_subscriptions found: %d", duplicateActiveRows)
	}
	return nil
}

func parseInt64CSV(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || value <= 0 {
			return nil, fmt.Errorf("invalid int64 CSV value %q", part)
		}
		out = append(out, value)
	}
	return out, nil
}
