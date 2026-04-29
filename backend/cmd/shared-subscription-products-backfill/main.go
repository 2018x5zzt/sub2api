package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const (
	SkipReasonAlreadyMigrated                    = "already_migrated"
	SkipReasonAmbiguousLegacySource              = "ambiguous_legacy_source"
	SkipReasonDuplicateActiveProductSubscription = "duplicate_active_product_subscription"
	SkipReasonExistingActiveProductSubscription  = "existing_active_product_subscription"
)

type BackfillOptions struct {
	DryRun         bool
	ProductCode    string
	ProductID      int64
	SourceGroupIDs []int64
	MigrationBatch string
	DSN            string
}

type LegacySubscriptionRow struct {
	ID                         int64
	UserID                     int64
	GroupID                    int64
	Status                     string
	StartsAt                   time.Time
	ExpiresAt                  time.Time
	DailyWindowStart           *time.Time
	WeeklyWindowStart          *time.Time
	MonthlyWindowStart         *time.Time
	DailyUsageUSD              float64
	WeeklyUsageUSD             float64
	MonthlyUsageUSD            float64
	DailyCarryoverInUSD        float64
	DailyCarryoverRemainingUSD float64
	AssignedBy                 *int64
	AssignedAt                 time.Time
	Notes                      string
}

type ExistingProductSubscriptionRow struct {
	ID        int64
	UserID    int64
	ProductID int64
	Status    string
}

type MigrationSourceRow struct {
	ProductSubscriptionID    int64
	LegacyUserSubscriptionID int64
	MigrationBatch           string
}

type BackfillInput struct {
	ProductID                    int64
	SourceGroupIDs               []int64
	MigrationBatch               string
	LegacySubscriptions          []LegacySubscriptionRow
	ExistingProductSubscriptions []ExistingProductSubscriptionRow
	MigrationSources             []MigrationSourceRow
}

type PlannedBackfillRow struct {
	LegacySubscription LegacySubscriptionRow
	ProductID          int64
}

type SkippedBackfillRow struct {
	LegacySubscriptionID int64
	UserID               int64
	Reason               string
}

type AppliedBackfillRow struct {
	LegacySubscriptionID  int64
	ProductSubscriptionID int64
	UserID                int64
}

type BackfillReport struct {
	TotalCandidates int
	Planned         []PlannedBackfillRow
	Skipped         []SkippedBackfillRow
	Applied         []AppliedBackfillRow
}

type InMemoryBackfillState struct {
	NextProductSubscriptionID int64
	ProductID                 int64
	SourceGroupIDs            []int64
	LegacySubscriptions       []LegacySubscriptionRow
	ProductSubscriptions      []ExistingProductSubscriptionRow
	MigrationSources          []MigrationSourceRow
}

func main() {
	opts, err := parseBackfillFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := runBackfill(context.Background(), opts, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "backfill failed: %v\n", err)
		os.Exit(1)
	}
}

func parseBackfillFlags(args []string) (BackfillOptions, error) {
	var sourceGroupIDs string
	opts := BackfillOptions{}
	fs := flag.NewFlagSet("shared-subscription-products-backfill", flag.ContinueOnError)
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print report without writing rows")
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

func BuildBackfillReport(input BackfillInput) BackfillReport {
	sourceGroups := make(map[int64]struct{}, len(input.SourceGroupIDs))
	for _, groupID := range input.SourceGroupIDs {
		sourceGroups[groupID] = struct{}{}
	}

	migratedLegacy := make(map[int64]struct{}, len(input.MigrationSources))
	for _, source := range input.MigrationSources {
		if input.MigrationBatch == "" || source.MigrationBatch == input.MigrationBatch {
			migratedLegacy[source.LegacyUserSubscriptionID] = struct{}{}
		}
	}

	activeProductSubsByUser := make(map[int64][]ExistingProductSubscriptionRow)
	for _, sub := range input.ExistingProductSubscriptions {
		if sub.ProductID == input.ProductID && strings.EqualFold(sub.Status, "active") {
			activeProductSubsByUser[sub.UserID] = append(activeProductSubsByUser[sub.UserID], sub)
		}
	}

	report := BackfillReport{}
	candidatesByUser := map[int64][]LegacySubscriptionRow{}
	for _, legacy := range input.LegacySubscriptions {
		if _, ok := sourceGroups[legacy.GroupID]; !ok {
			continue
		}
		if !strings.EqualFold(legacy.Status, "active") {
			continue
		}
		report.TotalCandidates++
		if _, ok := migratedLegacy[legacy.ID]; ok {
			report.Skipped = append(report.Skipped, skippedRow(legacy, SkipReasonAlreadyMigrated))
			continue
		}
		candidatesByUser[legacy.UserID] = append(candidatesByUser[legacy.UserID], legacy)
	}

	userIDs := make([]int64, 0, len(candidatesByUser))
	for userID := range candidatesByUser {
		userIDs = append(userIDs, userID)
	}
	sort.Slice(userIDs, func(i, j int) bool { return userIDs[i] < userIDs[j] })

	for _, userID := range userIDs {
		candidates := candidatesByUser[userID]
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
		if len(candidates) > 1 {
			for _, legacy := range candidates {
				report.Skipped = append(report.Skipped, skippedRow(legacy, SkipReasonAmbiguousLegacySource))
			}
			continue
		}
		legacy := candidates[0]
		activeProductSubs := activeProductSubsByUser[userID]
		if len(activeProductSubs) > 1 {
			report.Skipped = append(report.Skipped, skippedRow(legacy, SkipReasonDuplicateActiveProductSubscription))
			continue
		}
		if len(activeProductSubs) == 1 {
			report.Skipped = append(report.Skipped, skippedRow(legacy, SkipReasonExistingActiveProductSubscription))
			continue
		}
		report.Planned = append(report.Planned, PlannedBackfillRow{
			LegacySubscription: legacy,
			ProductID:          input.ProductID,
		})
	}

	return report
}

func ApplyBackfill(state *InMemoryBackfillState, opts BackfillOptions) (BackfillReport, error) {
	if state.NextProductSubscriptionID == 0 {
		state.NextProductSubscriptionID = 1
	}
	input := BackfillInput{
		ProductID:                    state.ProductID,
		SourceGroupIDs:               state.SourceGroupIDs,
		MigrationBatch:               opts.MigrationBatch,
		LegacySubscriptions:          state.LegacySubscriptions,
		ExistingProductSubscriptions: state.ProductSubscriptions,
		MigrationSources:             state.MigrationSources,
	}
	report := BuildBackfillReport(input)
	if opts.DryRun {
		return report, nil
	}
	for _, planned := range report.Planned {
		id := state.NextProductSubscriptionID
		state.NextProductSubscriptionID++
		legacy := planned.LegacySubscription
		state.ProductSubscriptions = append(state.ProductSubscriptions, ExistingProductSubscriptionRow{
			ID:        id,
			UserID:    legacy.UserID,
			ProductID: state.ProductID,
			Status:    legacy.Status,
		})
		state.MigrationSources = append(state.MigrationSources, MigrationSourceRow{
			ProductSubscriptionID:    id,
			LegacyUserSubscriptionID: legacy.ID,
			MigrationBatch:           opts.MigrationBatch,
		})
		report.Applied = append(report.Applied, AppliedBackfillRow{
			LegacySubscriptionID:  legacy.ID,
			ProductSubscriptionID: id,
			UserID:                legacy.UserID,
		})
	}
	return report, nil
}

func skippedRow(legacy LegacySubscriptionRow, reason string) SkippedBackfillRow {
	return SkippedBackfillRow{
		LegacySubscriptionID: legacy.ID,
		UserID:               legacy.UserID,
		Reason:               reason,
	}
}

func runBackfill(ctx context.Context, opts BackfillOptions, out io.Writer) error {
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

	productID, err := loadProductID(ctx, db, opts.ProductCode)
	if err != nil {
		return err
	}
	opts.ProductID = productID
	input, err := loadBackfillInput(ctx, db, opts)
	if err != nil {
		return err
	}
	report := BuildBackfillReport(input)
	if !opts.DryRun {
		if err := applyDBBackfill(ctx, db, opts, report); err != nil {
			return err
		}
	}
	printBackfillReport(out, opts, report)
	return nil
}

func loadProductID(ctx context.Context, db *sql.DB, productCode string) (int64, error) {
	var productID int64
	err := db.QueryRowContext(ctx, `
SELECT id
FROM subscription_products
WHERE code = $1
  AND deleted_at IS NULL
`, productCode).Scan(&productID)
	if err != nil {
		return 0, fmt.Errorf("resolve product code %q: %w", productCode, err)
	}
	return productID, nil
}

func loadBackfillInput(ctx context.Context, db *sql.DB, opts BackfillOptions) (BackfillInput, error) {
	legacy, err := loadLegacySubscriptions(ctx, db, opts.SourceGroupIDs)
	if err != nil {
		return BackfillInput{}, err
	}
	existing, err := loadExistingProductSubscriptions(ctx, db, opts.ProductID)
	if err != nil {
		return BackfillInput{}, err
	}
	sources, err := loadMigrationSources(ctx, db, opts.MigrationBatch)
	if err != nil {
		return BackfillInput{}, err
	}
	return BackfillInput{
		ProductID:                    opts.ProductID,
		SourceGroupIDs:               opts.SourceGroupIDs,
		MigrationBatch:               opts.MigrationBatch,
		LegacySubscriptions:          legacy,
		ExistingProductSubscriptions: existing,
		MigrationSources:             sources,
	}, nil
}

func loadLegacySubscriptions(ctx context.Context, db *sql.DB, sourceGroupIDs []int64) ([]LegacySubscriptionRow, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, user_id, group_id, status, starts_at, expires_at,
       daily_window_start, weekly_window_start, monthly_window_start,
       daily_usage_usd, weekly_usage_usd, monthly_usage_usd,
       daily_carryover_in_usd, daily_carryover_remaining_usd,
       assigned_by, assigned_at, COALESCE(notes, '')
FROM user_subscriptions
WHERE deleted_at IS NULL
  AND group_id = ANY($1)
`, pq.Array(sourceGroupIDs))
	if err != nil {
		return nil, fmt.Errorf("load legacy subscriptions: %w", err)
	}
	defer rows.Close()
	var out []LegacySubscriptionRow
	for rows.Next() {
		var row LegacySubscriptionRow
		var daily, weekly, monthly sql.NullTime
		var assignedBy sql.NullInt64
		if err := rows.Scan(
			&row.ID, &row.UserID, &row.GroupID, &row.Status, &row.StartsAt, &row.ExpiresAt,
			&daily, &weekly, &monthly,
			&row.DailyUsageUSD, &row.WeeklyUsageUSD, &row.MonthlyUsageUSD,
			&row.DailyCarryoverInUSD, &row.DailyCarryoverRemainingUSD,
			&assignedBy, &row.AssignedAt, &row.Notes,
		); err != nil {
			return nil, fmt.Errorf("scan legacy subscription: %w", err)
		}
		row.DailyWindowStart = nullTimePtr(daily)
		row.WeeklyWindowStart = nullTimePtr(weekly)
		row.MonthlyWindowStart = nullTimePtr(monthly)
		if assignedBy.Valid {
			v := assignedBy.Int64
			row.AssignedBy = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func loadExistingProductSubscriptions(ctx context.Context, db *sql.DB, productID int64) ([]ExistingProductSubscriptionRow, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, user_id, product_id, status
FROM user_product_subscriptions
WHERE deleted_at IS NULL
  AND product_id = $1
`, productID)
	if err != nil {
		return nil, fmt.Errorf("load existing product subscriptions: %w", err)
	}
	defer rows.Close()
	var out []ExistingProductSubscriptionRow
	for rows.Next() {
		var row ExistingProductSubscriptionRow
		if err := rows.Scan(&row.ID, &row.UserID, &row.ProductID, &row.Status); err != nil {
			return nil, fmt.Errorf("scan product subscription: %w", err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func loadMigrationSources(ctx context.Context, db *sql.DB, batch string) ([]MigrationSourceRow, error) {
	rows, err := db.QueryContext(ctx, `
SELECT product_subscription_id, legacy_user_subscription_id, migration_batch
FROM product_subscription_migration_sources
WHERE migration_batch = $1
`, batch)
	if err != nil {
		return nil, fmt.Errorf("load migration sources: %w", err)
	}
	defer rows.Close()
	var out []MigrationSourceRow
	for rows.Next() {
		var row MigrationSourceRow
		if err := rows.Scan(&row.ProductSubscriptionID, &row.LegacyUserSubscriptionID, &row.MigrationBatch); err != nil {
			return nil, fmt.Errorf("scan migration source: %w", err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func applyDBBackfill(ctx context.Context, db *sql.DB, opts BackfillOptions, report BackfillReport) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, planned := range report.Planned {
		legacy := planned.LegacySubscription
		var productSubscriptionID int64
		err := tx.QueryRowContext(ctx, `
INSERT INTO user_product_subscriptions (
  user_id, product_id, starts_at, expires_at, status,
  daily_window_start, weekly_window_start, monthly_window_start,
  daily_usage_usd, weekly_usage_usd, monthly_usage_usd,
  daily_carryover_in_usd, daily_carryover_remaining_usd,
  assigned_by, assigned_at, notes
) VALUES (
  $1, $2, $3, $4, $5,
  $6, $7, $8,
  $9, $10, $11,
  $12, $13,
  $14, $15, $16
)
RETURNING id
`, legacy.UserID, opts.ProductID, legacy.StartsAt, legacy.ExpiresAt, legacy.Status,
			legacy.DailyWindowStart, legacy.WeeklyWindowStart, legacy.MonthlyWindowStart,
			legacy.DailyUsageUSD, legacy.WeeklyUsageUSD, legacy.MonthlyUsageUSD,
			0, 0,
			legacy.AssignedBy, legacy.AssignedAt, "shared product backfill: "+opts.MigrationBatch,
		).Scan(&productSubscriptionID)
		if err != nil {
			return fmt.Errorf("insert product subscription for legacy %d: %w", legacy.ID, err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO product_subscription_migration_sources (
  product_subscription_id, legacy_user_subscription_id, migration_batch,
  legacy_group_id, legacy_status, legacy_starts_at, legacy_expires_at,
  legacy_daily_usage_usd, legacy_weekly_usage_usd, legacy_monthly_usage_usd,
  legacy_daily_carryover_in_usd, legacy_daily_carryover_remaining_usd
) VALUES (
  $1, $2, $3,
  $4, $5, $6, $7,
  $8, $9, $10,
  $11, $12
)
`, productSubscriptionID, legacy.ID, opts.MigrationBatch,
			legacy.GroupID, legacy.Status, legacy.StartsAt, legacy.ExpiresAt,
			legacy.DailyUsageUSD, legacy.WeeklyUsageUSD, legacy.MonthlyUsageUSD,
			legacy.DailyCarryoverInUSD, legacy.DailyCarryoverRemainingUSD,
		); err != nil {
			return fmt.Errorf("insert migration source for legacy %d: %w", legacy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func printBackfillReport(out io.Writer, opts BackfillOptions, report BackfillReport) {
	mode := "apply"
	if opts.DryRun {
		mode = "dry-run"
	}
	fmt.Fprintf(out, "shared subscription product backfill (%s)\n", mode)
	fmt.Fprintf(out, "product_code=%s product_id=%d migration_batch=%s source_group_ids=%v\n", opts.ProductCode, opts.ProductID, opts.MigrationBatch, opts.SourceGroupIDs)
	fmt.Fprintf(out, "candidates=%d planned=%d skipped=%d applied=%d\n", report.TotalCandidates, len(report.Planned), len(report.Skipped), len(report.Applied))
	for _, skipped := range report.Skipped {
		fmt.Fprintf(out, "skip legacy_subscription_id=%d user_id=%d reason=%s\n", skipped.LegacySubscriptionID, skipped.UserID, skipped.Reason)
	}
	limit := len(report.Planned)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		planned := report.Planned[i]
		fmt.Fprintf(out, "sample legacy_subscription_id=%d user_id=%d group_id=%d monthly_usage_usd=%.6f\n",
			planned.LegacySubscription.ID,
			planned.LegacySubscription.UserID,
			planned.LegacySubscription.GroupID,
			planned.LegacySubscription.MonthlyUsageUSD,
		)
	}
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

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}
