package repository

import (
	"context"
	"database/sql"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type subscriptionUsageSQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func advanceAndIncrementProductSubscriptionUsage(ctx context.Context, exec subscriptionUsageSQLExecutor, productSubscriptionID int64, costUSD float64) error {
	const updateSQL = `
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
				date_trunc('day', NOW()) AS today_start,
				date_trunc('week', NOW()) AS week_start,
				date_trunc('day', NOW()) AS month_start,
				CASE
					WHEN ups.daily_window_start IS NULL THEN true
					WHEN ups.daily_window_start + INTERVAL '24 hours' <= NOW() THEN true
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
					WHEN ups.daily_window_start + INTERVAL '24 hours' <= NOW()
						THEN GREATEST(CAST(EXTRACT(EPOCH FROM (date_trunc('day', NOW()) - ups.daily_window_start)) / 86400 AS bigint), 0)
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
					WHEN should_advance_weekly THEN week_start
					ELSE weekly_window_start
				END AS next_weekly_window_start,
				CASE
					WHEN should_advance_weekly THEN 0
					ELSE weekly_usage_usd
				END AS base_weekly_usage,
				CASE
					WHEN should_advance_monthly THEN month_start
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
