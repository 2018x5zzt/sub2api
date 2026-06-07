package subscriptions

import (
	"context"
	"database/sql"
	"time"
)

const coreProductsSQL = `
SELECT id, code, name, COALESCE(description, ''), status, product_family,
       daily_limit_usd, weekly_limit_usd, monthly_limit_usd, created_at, updated_at
FROM subscription_products
WHERE deleted_at IS NULL`

const coreProductGroupsSQL = `
SELECT spg.id, spg.product_id, spg.group_id, g.name, COALESCE(g.platform, ''),
       g.balance_fallback_group_id, fg.name,
       spg.debit_multiplier, spg.status, spg.sort_order, spg.created_at, spg.updated_at
FROM subscription_product_groups spg
JOIN groups g ON g.id = spg.group_id AND g.deleted_at IS NULL
LEFT JOIN groups fg ON fg.id = g.balance_fallback_group_id AND fg.deleted_at IS NULL
WHERE spg.deleted_at IS NULL`

const coreUserProductSubscriptionsSQL = `
SELECT id, user_id, product_id, status, starts_at, expires_at,
       daily_usage_usd, weekly_usage_usd, monthly_usage_usd,
       daily_carryover_in_usd, daily_carryover_remaining_usd, created_at, updated_at
FROM user_product_subscriptions
WHERE deleted_at IS NULL
  AND (status = 'active' OR updated_at > NOW() - interval '30 days')`

const upsertProductSQL = `
INSERT INTO xlab_subscription_products (
    core_product_id, code, name, description, status, product_family,
    daily_limit_usd, weekly_limit_usd, monthly_limit_usd,
    source_created_at, source_updated_at, synced_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (core_product_id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    product_family = EXCLUDED.product_family,
    daily_limit_usd = EXCLUDED.daily_limit_usd,
    weekly_limit_usd = EXCLUDED.weekly_limit_usd,
    monthly_limit_usd = EXCLUDED.monthly_limit_usd,
    source_created_at = EXCLUDED.source_created_at,
    source_updated_at = EXCLUDED.source_updated_at,
    synced_at = EXCLUDED.synced_at`

const upsertProductGroupSQL = `
INSERT INTO xlab_subscription_product_groups (
    core_binding_id, core_product_id, core_group_id, group_name, group_platform,
    balance_fallback_group_id, balance_fallback_group_name,
    debit_multiplier, status, sort_order, source_created_at, source_updated_at, synced_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (core_binding_id) DO UPDATE SET
    core_product_id = EXCLUDED.core_product_id,
    core_group_id = EXCLUDED.core_group_id,
    group_name = EXCLUDED.group_name,
    group_platform = EXCLUDED.group_platform,
    balance_fallback_group_id = EXCLUDED.balance_fallback_group_id,
    balance_fallback_group_name = EXCLUDED.balance_fallback_group_name,
    debit_multiplier = EXCLUDED.debit_multiplier,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    source_created_at = EXCLUDED.source_created_at,
    source_updated_at = EXCLUDED.source_updated_at,
    synced_at = EXCLUDED.synced_at`

const upsertUserProductSubscriptionSQL = `
INSERT INTO xlab_user_product_subscriptions (
    core_subscription_id, core_user_id, core_product_id, status, started_at, expires_at,
    daily_usage_usd, weekly_usage_usd, monthly_usage_usd,
    daily_carryover_in_usd, daily_carryover_remaining_usd,
    source_created_at, source_updated_at, synced_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
ON CONFLICT (core_subscription_id) DO UPDATE SET
    core_user_id = EXCLUDED.core_user_id,
    core_product_id = EXCLUDED.core_product_id,
    status = EXCLUDED.status,
    started_at = EXCLUDED.started_at,
    expires_at = EXCLUDED.expires_at,
    daily_usage_usd = EXCLUDED.daily_usage_usd,
    weekly_usage_usd = EXCLUDED.weekly_usage_usd,
    monthly_usage_usd = EXCLUDED.monthly_usage_usd,
    daily_carryover_in_usd = EXCLUDED.daily_carryover_in_usd,
    daily_carryover_remaining_usd = EXCLUDED.daily_carryover_remaining_usd,
    source_created_at = EXCLUDED.source_created_at,
    source_updated_at = EXCLUDED.source_updated_at,
    synced_at = EXCLUDED.synced_at`

const upsertSyncStateSQL = `
INSERT INTO xlab_sync_state (source_name, last_success_at, row_count, created_at, updated_at)
VALUES ('product_subscriptions', $1, $2, $1, $1)
ON CONFLICT (source_name) DO UPDATE SET
    last_success_at = EXCLUDED.last_success_at,
    row_count = EXCLUDED.row_count,
    last_error = NULL,
    updated_at = EXCLUDED.updated_at`

type Syncer struct {
	coreDB *sql.DB
	xlabDB *sql.DB
	now    func() time.Time
}

func NewSyncer(coreDB *sql.DB, xlabDB *sql.DB, now func() time.Time) *Syncer {
	if now == nil {
		now = time.Now
	}
	return &Syncer{coreDB: coreDB, xlabDB: xlabDB, now: now}
}

func (s *Syncer) SyncOnce(ctx context.Context) error {
	syncedAt := s.now()
	productRows, err := s.coreDB.QueryContext(ctx, coreProductsSQL)
	if err != nil {
		return err
	}
	defer productRows.Close()

	tx, err := s.xlabDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	rowCount := 0
	for productRows.Next() {
		var id int64
		var code, name, description, status, family string
		var daily, weekly, monthly float64
		var createdAt, updatedAt time.Time
		if err := productRows.Scan(&id, &code, &name, &description, &status, &family, &daily, &weekly, &monthly, &createdAt, &updatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, upsertProductSQL, id, code, name, description, status, family, daily, weekly, monthly, createdAt, updatedAt, syncedAt); err != nil {
			return err
		}
		rowCount++
	}
	if err := productRows.Err(); err != nil {
		return err
	}

	if err := s.syncGroups(ctx, tx, syncedAt, &rowCount); err != nil {
		return err
	}
	if err := s.syncSubscriptions(ctx, tx, syncedAt, &rowCount); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, upsertSyncStateSQL, syncedAt, rowCount); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Syncer) syncGroups(ctx context.Context, tx *sql.Tx, syncedAt time.Time, rowCount *int) error {
	rows, err := s.coreDB.QueryContext(ctx, coreProductGroupsSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, productID, groupID int64
		var name, platform, status string
		var fallbackGroupID sql.NullInt64
		var fallbackGroupName sql.NullString
		var multiplier float64
		var sortOrder int
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &productID, &groupID, &name, &platform, &fallbackGroupID, &fallbackGroupName, &multiplier, &status, &sortOrder, &createdAt, &updatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, upsertProductGroupSQL, id, productID, groupID, name, platform, nullableInt64(fallbackGroupID), nullableString(fallbackGroupName), multiplier, status, sortOrder, createdAt, updatedAt, syncedAt); err != nil {
			return err
		}
		(*rowCount)++
	}
	return rows.Err()
}

func nullableInt64(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func nullableString(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
}

func (s *Syncer) syncSubscriptions(ctx context.Context, tx *sql.Tx, syncedAt time.Time, rowCount *int) error {
	rows, err := s.coreDB.QueryContext(ctx, coreUserProductSubscriptionsSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID, productID int64
		var status string
		var startsAt, expiresAt, createdAt, updatedAt time.Time
		var dailyUsage, weeklyUsage, monthlyUsage, carryoverIn, carryoverRemaining float64
		if err := rows.Scan(&id, &userID, &productID, &status, &startsAt, &expiresAt, &dailyUsage, &weeklyUsage, &monthlyUsage, &carryoverIn, &carryoverRemaining, &createdAt, &updatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, upsertUserProductSubscriptionSQL, id, userID, productID, status, startsAt, expiresAt, dailyUsage, weeklyUsage, monthlyUsage, carryoverIn, carryoverRemaining, createdAt, updatedAt, syncedAt); err != nil {
			return err
		}
		(*rowCount)++
	}
	return rows.Err()
}
