package repository

import (
	"context"
	"database/sql"
	"math"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const beijingDayStartSQL = "date_trunc('day', NOW() AT TIME ZONE 'Asia/Shanghai') AT TIME ZONE 'Asia/Shanghai'"

func beijingDayStartSQLFromColumn(column string) string {
	return "date_trunc('day', " + column + " AT TIME ZONE 'Asia/Shanghai') AT TIME ZONE 'Asia/Shanghai'"
}

type subscriptionUsageSQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func advanceAndIncrementProductSubscriptionUsage(ctx context.Context, exec subscriptionUsageSQLExecutor, productSubscriptionID int64, costUSD float64) error {
	updateSQL := `
		WITH prepared AS (
			SELECT
				ups.id,
				ups.status,
				ups.expires_at,
				ups.daily_window_start,
				ups.daily_usage_usd,
				ups.weekly_window_start,
				ups.weekly_usage_usd,
				ups.monthly_window_start,
				ups.monthly_usage_usd,
				ups.daily_carryover_in_usd,
				ups.daily_carryover_remaining_usd,
				COALESCE(sp.daily_limit_usd, 0) AS daily_limit_usd,
				` + beijingDayStartSQL + ` AS today_start,
				CASE
					WHEN ups.daily_window_start IS NULL THEN true
					WHEN ` + beijingDayStartSQL + ` > (` + beijingDayStartSQLFromColumn("ups.daily_window_start") + `) THEN true
					ELSE false
				END AS should_advance_daily,
				CASE
					WHEN ups.weekly_window_start IS NULL THEN true
					WHEN ups.weekly_window_start + INTERVAL '7 days' <= NOW() THEN true
					ELSE false
				END AS should_advance_weekly,
				CASE
					WHEN ups.monthly_window_start IS NULL THEN true
					WHEN ups.monthly_window_start + INTERVAL '30 days' <= NOW() THEN true
					ELSE false
				END AS should_advance_monthly,
				CASE
					WHEN ups.daily_window_start IS NULL THEN 0
					WHEN ` + beijingDayStartSQL + ` > (` + beijingDayStartSQLFromColumn("ups.daily_window_start") + `)
						THEN GREATEST(CAST(EXTRACT(EPOCH FROM (` + beijingDayStartSQL + ` - (` + beijingDayStartSQLFromColumn("ups.daily_window_start") + `))) / 86400 AS bigint), 0)
					ELSE 0
				END AS elapsed_days,
				GREATEST(ups.daily_carryover_in_usd - ups.daily_carryover_remaining_usd, 0) AS carryover_consumed,
				GREATEST(ups.daily_usage_usd - GREATEST(ups.daily_carryover_in_usd - ups.daily_carryover_remaining_usd, 0), 0) AS fresh_used
			FROM user_product_subscriptions ups
			JOIN subscription_products sp ON sp.id = ups.product_id AND sp.deleted_at IS NULL
			WHERE ups.id = $2
				AND ups.deleted_at IS NULL
		),
		next_state AS (
			SELECT
				id,
				CASE
					WHEN should_advance_daily THEN today_start
					ELSE daily_window_start
				END AS next_daily_window_start,
				CASE
					WHEN should_advance_daily THEN 0
					ELSE daily_usage_usd
				END AS base_daily_usage,
				CASE
					WHEN should_advance_weekly AND weekly_window_start IS NOT NULL THEN weekly_window_start + (FLOOR(EXTRACT(EPOCH FROM (NOW() - weekly_window_start)) / EXTRACT(EPOCH FROM INTERVAL '7 days')) * INTERVAL '7 days')
					WHEN should_advance_weekly THEN NOW()
					ELSE weekly_window_start
				END AS next_weekly_window_start,
				CASE
					WHEN should_advance_weekly THEN 0
					ELSE weekly_usage_usd
				END AS base_weekly_usage,
				CASE
					WHEN should_advance_monthly AND monthly_window_start IS NOT NULL THEN monthly_window_start + (FLOOR(EXTRACT(EPOCH FROM (NOW() - monthly_window_start)) / EXTRACT(EPOCH FROM INTERVAL '30 days')) * INTERVAL '30 days')
					WHEN should_advance_monthly THEN NOW()
					ELSE monthly_window_start
				END AS next_monthly_window_start,
				CASE
					WHEN should_advance_monthly THEN 0
					ELSE monthly_usage_usd
				END AS base_monthly_usage,
				CASE
					WHEN should_advance_daily
						AND elapsed_days = 1
						AND status = $3
						AND expires_at > today_start
						THEN GREATEST(daily_limit_usd - fresh_used, 0)
					WHEN should_advance_daily THEN 0
					ELSE daily_carryover_in_usd
				END AS next_carryover_in,
				CASE
					WHEN should_advance_daily
						AND elapsed_days = 1
						AND status = $3
						AND expires_at > today_start
						THEN GREATEST(daily_limit_usd - fresh_used, 0)
					WHEN should_advance_daily THEN 0
					ELSE daily_carryover_remaining_usd
				END AS next_carryover_remaining
			FROM prepared
		)
		UPDATE user_product_subscriptions ups
		SET
			daily_window_start = next_state.next_daily_window_start,
			weekly_window_start = next_state.next_weekly_window_start,
			monthly_window_start = next_state.next_monthly_window_start,
			daily_usage_usd = next_state.base_daily_usage + $1,
			weekly_usage_usd = next_state.base_weekly_usage + $1,
			monthly_usage_usd = next_state.base_monthly_usage + $1,
			daily_carryover_in_usd = next_state.next_carryover_in,
			daily_carryover_remaining_usd = GREATEST(
				next_state.next_carryover_remaining - LEAST(next_state.next_carryover_remaining, $1),
				0
			),
			updated_at = NOW()
		FROM next_state
		WHERE ups.id = next_state.id
	`

	res, err := exec.ExecContext(ctx, updateSQL, costUSD, productSubscriptionID, service.SubscriptionStatusActive)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	return service.ErrSubscriptionNotFound
}

func splitAndIncrementProductSubscriptionUsage(ctx context.Context, tx *sql.Tx, userID, productSubscriptionID, groupID int64, costUSD float64) error {
	if costUSD <= 0 {
		return nil
	}

	candidates, err := listProductDebitCandidates(ctx, tx, userID, productSubscriptionID, groupID)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return service.ErrSubscriptionNotFound
	}

	remaining := costUSD
	for _, candidate := range candidates {
		debit := math.Min(remaining, candidate.remaining)
		if debit <= 0 {
			continue
		}
		if err := advanceAndIncrementProductSubscriptionUsage(ctx, tx, candidate.subscriptionID, debit); err != nil {
			return err
		}
		remaining -= debit
		if remaining <= 0.0000001 {
			return nil
		}
	}

	return service.ErrDailyLimitExceeded
}

type productDebitCandidate struct {
	subscriptionID int64
	remaining      float64
}

func listProductDebitCandidates(ctx context.Context, tx *sql.Tx, userID, productSubscriptionID, groupID int64) ([]productDebitCandidate, error) {
	rows, err := tx.QueryContext(ctx, `
		WITH anchor AS (
			SELECT
				ups.user_id,
				sp.product_family
			FROM user_product_subscriptions ups
			JOIN subscription_products sp
				ON sp.id = ups.product_id
				AND sp.deleted_at IS NULL
			WHERE ups.id = $2
				AND ups.user_id = $1
				AND ups.deleted_at IS NULL
		),
		prepared AS (
			SELECT
				ups.id,
				COALESCE(sp.daily_limit_usd, 0) AS daily_limit_usd,
				COALESCE(sp.weekly_limit_usd, 0) AS weekly_limit_usd,
				COALESCE(sp.monthly_limit_usd, 0) AS monthly_limit_usd,
				CASE
					WHEN ups.daily_window_start IS NULL THEN 0
					WHEN `+beijingDayStartSQL+` > (`+beijingDayStartSQLFromColumn("ups.daily_window_start")+`)
						THEN GREATEST(CAST(EXTRACT(EPOCH FROM (`+beijingDayStartSQL+` - (`+beijingDayStartSQLFromColumn("ups.daily_window_start")+`))) / 86400 AS bigint), 0)
					ELSE 0
				END AS elapsed_days,
				CASE
					WHEN ups.daily_window_start IS NULL OR `+beijingDayStartSQL+` > (`+beijingDayStartSQLFromColumn("ups.daily_window_start")+`)
						THEN 0
					ELSE ups.daily_usage_usd
				END AS base_daily_usage,
				CASE
					WHEN ups.weekly_window_start IS NULL OR ups.weekly_window_start + INTERVAL '7 days' <= NOW()
						THEN 0
					ELSE ups.weekly_usage_usd
				END AS base_weekly_usage,
				CASE
					WHEN ups.monthly_window_start IS NULL OR ups.monthly_window_start + INTERVAL '30 days' <= NOW()
						THEN 0
					ELSE ups.monthly_usage_usd
				END AS base_monthly_usage,
				CASE
					WHEN ups.daily_window_start IS NULL OR `+beijingDayStartSQL+` = (`+beijingDayStartSQLFromColumn("ups.daily_window_start")+`)
						THEN ups.daily_carryover_in_usd
					WHEN ups.status = $4
						AND ups.expires_at > `+beijingDayStartSQL+`
						AND GREATEST(CAST(EXTRACT(EPOCH FROM (`+beijingDayStartSQL+` - (`+beijingDayStartSQLFromColumn("ups.daily_window_start")+`))) / 86400 AS bigint), 0) = 1
						THEN GREATEST(COALESCE(sp.daily_limit_usd, 0) - GREATEST(ups.daily_usage_usd - GREATEST(ups.daily_carryover_in_usd - ups.daily_carryover_remaining_usd, 0), 0), 0)
					ELSE 0
				END AS effective_carryover
			FROM user_product_subscriptions ups
			JOIN subscription_products sp
				ON sp.id = ups.product_id
				AND sp.deleted_at IS NULL
				AND sp.status = 'active'
			JOIN subscription_product_groups spg
				ON spg.product_id = sp.id
				AND spg.group_id = $3
				AND spg.deleted_at IS NULL
				AND spg.status = 'active'
			JOIN anchor
				ON anchor.user_id = ups.user_id
				AND anchor.product_family = sp.product_family
			WHERE ups.user_id = $1
				AND ups.status = $4
				AND ups.expires_at > NOW()
				AND ups.deleted_at IS NULL
			ORDER BY sp.sort_order ASC, ups.starts_at ASC, ups.id ASC
			FOR UPDATE OF ups
		)
		SELECT
			id,
			LEAST(
				CASE
					WHEN daily_limit_usd > 0 THEN GREATEST(daily_limit_usd + GREATEST(effective_carryover, 0) - base_daily_usage, 0)
					ELSE 1.0e100
				END,
				CASE
					WHEN weekly_limit_usd > 0 THEN GREATEST(weekly_limit_usd - base_weekly_usage, 0)
					ELSE 1.0e100
				END,
				CASE
					WHEN monthly_limit_usd > 0 THEN GREATEST(monthly_limit_usd - base_monthly_usage, 0)
					ELSE 1.0e100
				END
			) AS remaining
		FROM prepared
	`, userID, productSubscriptionID, groupID, service.SubscriptionStatusActive)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	candidates := make([]productDebitCandidate, 0)
	for rows.Next() {
		var candidate productDebitCandidate
		if err := rows.Scan(&candidate.subscriptionID, &candidate.remaining); err != nil {
			return nil, err
		}
		if candidate.remaining > 0 {
			candidates = append(candidates, candidate)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, service.ErrDailyLimitExceeded
	}
	return candidates, nil
}
