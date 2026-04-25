package repository

import (
	"context"
	"database/sql"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type subscriptionUsageSQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func advanceAndIncrementSubscriptionUsage(ctx context.Context, exec subscriptionUsageSQLExecutor, subscriptionID int64, costUSD float64) error {
	const updateSQL = `
		WITH prepared AS (
			SELECT
				us.id,
				us.status,
				us.expires_at,
				us.daily_window_start,
				us.daily_usage_usd,
				us.daily_carryover_in_usd,
				us.daily_carryover_remaining_usd,
				COALESCE(g.daily_limit_usd, 0) AS daily_limit_usd,
				date_trunc('day', NOW()) AS today_start,
				CASE
					WHEN us.daily_window_start IS NULL THEN true
					WHEN us.daily_window_start + INTERVAL '24 hours' <= NOW() THEN true
					ELSE false
				END AS should_advance,
				CASE
					WHEN us.daily_window_start IS NULL THEN 0
					WHEN us.daily_window_start + INTERVAL '24 hours' <= NOW()
						THEN GREATEST(CAST(EXTRACT(EPOCH FROM (date_trunc('day', NOW()) - us.daily_window_start)) / 86400 AS bigint), 0)
					ELSE 0
				END AS elapsed_days,
				GREATEST(us.daily_carryover_in_usd - us.daily_carryover_remaining_usd, 0) AS carryover_consumed,
				GREATEST(us.daily_usage_usd - GREATEST(us.daily_carryover_in_usd - us.daily_carryover_remaining_usd, 0), 0) AS fresh_used
			FROM user_subscriptions us
			JOIN groups g ON g.id = us.group_id AND g.deleted_at IS NULL
			WHERE us.id = $2
				AND us.deleted_at IS NULL
		),
		next_state AS (
			SELECT
				id,
				CASE
					WHEN should_advance THEN today_start
					ELSE daily_window_start
				END AS next_daily_window_start,
				CASE
					WHEN should_advance THEN 0
					ELSE daily_usage_usd
				END AS base_daily_usage,
				CASE
					WHEN should_advance
						AND elapsed_days = 1
						AND status = $3
						AND expires_at > today_start
						THEN GREATEST(daily_limit_usd - fresh_used, 0)
					WHEN should_advance THEN 0
					ELSE daily_carryover_in_usd
				END AS next_carryover_in,
				CASE
					WHEN should_advance
						AND elapsed_days = 1
						AND status = $3
						AND expires_at > today_start
						THEN GREATEST(daily_limit_usd - fresh_used, 0)
					WHEN should_advance THEN 0
					ELSE daily_carryover_remaining_usd
				END AS next_carryover_remaining
			FROM prepared
		)
		UPDATE user_subscriptions us
		SET
			daily_window_start = next_state.next_daily_window_start,
			daily_usage_usd = next_state.base_daily_usage + $1,
			weekly_usage_usd = us.weekly_usage_usd + $1,
			monthly_usage_usd = us.monthly_usage_usd + $1,
			daily_carryover_in_usd = next_state.next_carryover_in,
			daily_carryover_remaining_usd = GREATEST(
				next_state.next_carryover_remaining - LEAST(next_state.next_carryover_remaining, $1),
				0
			),
			updated_at = NOW()
		FROM next_state
		WHERE us.id = next_state.id
	`

	res, err := exec.ExecContext(ctx, updateSQL, costUSD, subscriptionID, service.SubscriptionStatusActive)
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
