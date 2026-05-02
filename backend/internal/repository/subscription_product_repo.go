package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type subscriptionProductRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewSubscriptionProductRepository(client *dbent.Client, sqlDB *sql.DB) service.ProductSubscriptionRepository {
	return &subscriptionProductRepository{client: client, sql: sqlDB}
}

func (r *subscriptionProductRepository) GetActiveProductSubscriptionByUserAndGroupID(ctx context.Context, userID, groupID int64) (*service.SubscriptionProductBinding, *service.UserProductSubscription, error) {
	var binding service.SubscriptionProductBinding
	var sub service.UserProductSubscription
	var dailyWindowStart sql.NullTime
	var weeklyWindowStart sql.NullTime
	var monthlyWindowStart sql.NullTime

	err := scanSingleRow(ctx, r.sql, `
SELECT
	sp.id,
	sp.code,
	sp.name,
	sp.status,
	sp.default_validity_days,
	sp.daily_limit_usd,
	sp.weekly_limit_usd,
	sp.monthly_limit_usd,
	spg.group_id,
	g.name,
	g.platform,
	g.status,
	g.subscription_type,
	spg.debit_multiplier,
	spg.status,
	ups.id,
	ups.user_id,
	ups.product_id,
	ups.starts_at,
	ups.expires_at,
	ups.status,
	ups.daily_window_start,
	ups.weekly_window_start,
	ups.monthly_window_start,
	ups.daily_usage_usd,
	ups.weekly_usage_usd,
	ups.monthly_usage_usd,
	ups.daily_carryover_in_usd,
	ups.daily_carryover_remaining_usd
FROM subscription_product_groups spg
JOIN subscription_products sp
	ON sp.id = spg.product_id
	AND sp.deleted_at IS NULL
	AND sp.status = 'active'
JOIN groups g
	ON g.id = spg.group_id
	AND g.deleted_at IS NULL
JOIN user_product_subscriptions ups
	ON ups.product_id = sp.id
	AND ups.user_id = $1
	AND ups.status = 'active'
	AND ups.expires_at > NOW()
	AND ups.deleted_at IS NULL
WHERE spg.group_id = $2
  AND spg.deleted_at IS NULL
  AND spg.status = 'active'
ORDER BY ups.expires_at DESC, ups.id DESC
LIMIT 1
`, []any{userID, groupID},
		&binding.ProductID,
		&binding.ProductCode,
		&binding.ProductName,
		&binding.ProductStatus,
		&binding.DefaultValidityDays,
		&binding.DailyLimitUSD,
		&binding.WeeklyLimitUSD,
		&binding.MonthlyLimitUSD,
		&binding.GroupID,
		&binding.GroupName,
		&binding.GroupPlatform,
		&binding.GroupStatus,
		&binding.GroupSubscription,
		&binding.DebitMultiplier,
		&binding.BindingStatus,
		&sub.ID,
		&sub.UserID,
		&sub.ProductID,
		&sub.StartsAt,
		&sub.ExpiresAt,
		&sub.Status,
		&dailyWindowStart,
		&weeklyWindowStart,
		&monthlyWindowStart,
		&sub.DailyUsageUSD,
		&sub.WeeklyUsageUSD,
		&sub.MonthlyUsageUSD,
		&sub.DailyCarryoverInUSD,
		&sub.DailyCarryoverRemainingUSD,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, service.ErrSubscriptionNotFound
		}
		return nil, nil, err
	}
	sub.DailyWindowStart = nullableTimePtr(dailyWindowStart)
	sub.WeeklyWindowStart = nullableTimePtr(weeklyWindowStart)
	sub.MonthlyWindowStart = nullableTimePtr(monthlyWindowStart)
	return &binding, &sub, nil
}

func nullableTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	return &v.Time
}
