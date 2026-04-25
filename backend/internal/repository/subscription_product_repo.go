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
	return newSubscriptionProductRepositoryWithSQL(client, sqlDB)
}

func newSubscriptionProductRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *subscriptionProductRepository {
	return &subscriptionProductRepository{client: client, sql: sqlq}
}

func (r *subscriptionProductRepository) GetActiveProductBindingByGroupID(ctx context.Context, groupID int64) (*service.SubscriptionProductBinding, error) {
	var binding service.SubscriptionProductBinding
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
	spg.status
FROM subscription_product_groups spg
JOIN subscription_products sp ON sp.id = spg.product_id
JOIN groups g ON g.id = spg.group_id
WHERE spg.group_id = $1
  AND spg.deleted_at IS NULL
  AND sp.deleted_at IS NULL
  AND g.deleted_at IS NULL
  AND sp.status = 'active'
  AND spg.status = 'active'
LIMIT 1
`, []any{groupID},
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
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.NewSubscriptionProductNotFoundError(groupID).WithCause(err)
		}
		return nil, err
	}
	return &binding, nil
}

func (r *subscriptionProductRepository) GetActiveUserProductSubscription(ctx context.Context, userID, productID int64) (*service.UserProductSubscription, error) {
	var sub service.UserProductSubscription
	var dailyWindowStart sql.NullTime
	var weeklyWindowStart sql.NullTime
	var monthlyWindowStart sql.NullTime
	var assignedBy sql.NullInt64
	var notes sql.NullString
	err := scanSingleRow(ctx, r.sql, `
SELECT
	id,
	user_id,
	product_id,
	starts_at,
	expires_at,
	status,
	daily_window_start,
	weekly_window_start,
	monthly_window_start,
	daily_usage_usd,
	weekly_usage_usd,
	monthly_usage_usd,
	daily_carryover_in_usd,
	daily_carryover_remaining_usd,
	assigned_by,
	assigned_at,
	notes,
	created_at,
	updated_at
FROM user_product_subscriptions
WHERE user_id = $1
  AND product_id = $2
  AND status = 'active'
  AND expires_at > NOW()
  AND deleted_at IS NULL
LIMIT 1
`, []any{userID, productID},
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
		&assignedBy,
		&sub.AssignedAt,
		&notes,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.NewProductSubscriptionInvalidError(&service.SubscriptionProductBinding{ProductID: productID}, nil).WithCause(err)
		}
		return nil, err
	}
	sub.DailyWindowStart = nullableTimePtr(dailyWindowStart)
	sub.WeeklyWindowStart = nullableTimePtr(weeklyWindowStart)
	sub.MonthlyWindowStart = nullableTimePtr(monthlyWindowStart)
	sub.AssignedBy = nullableInt64Ptr(assignedBy)
	sub.Notes = notes.String
	return &sub, nil
}

func (r *subscriptionProductRepository) ListVisibleGroupsByUserID(ctx context.Context, userID int64) ([]service.Group, error) {
	rows, err := r.sql.QueryContext(ctx, `
SELECT
	g.id,
	g.name,
	g.description,
	g.platform,
	g.rate_multiplier,
	g.pricing_mode,
	g.default_budget_multiplier,
	g.is_exclusive,
	g.status,
	g.subscription_type,
	g.daily_limit_usd,
	g.weekly_limit_usd,
	g.monthly_limit_usd,
	g.default_validity_days,
	g.sort_order,
	g.created_at,
	g.updated_at
FROM user_product_subscriptions ups
JOIN subscription_products sp ON sp.id = ups.product_id
JOIN subscription_product_groups spg ON spg.product_id = sp.id
JOIN groups g ON g.id = spg.group_id
WHERE ups.user_id = $1
  AND ups.status = 'active'
  AND ups.expires_at > NOW()
  AND ups.deleted_at IS NULL
  AND sp.status = 'active'
  AND sp.deleted_at IS NULL
  AND spg.status = 'active'
  AND spg.deleted_at IS NULL
  AND g.status = 'active'
  AND g.deleted_at IS NULL
ORDER BY sp.sort_order ASC, spg.sort_order ASC, g.sort_order ASC, g.id ASC
`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	groups := make([]service.Group, 0)
	for rows.Next() {
		var g service.Group
		var description sql.NullString
		var pricingMode sql.NullString
		var defaultBudgetMultiplier sql.NullFloat64
		var dailyLimit sql.NullFloat64
		var weeklyLimit sql.NullFloat64
		var monthlyLimit sql.NullFloat64
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&description,
			&g.Platform,
			&g.RateMultiplier,
			&pricingMode,
			&defaultBudgetMultiplier,
			&g.IsExclusive,
			&g.Status,
			&g.SubscriptionType,
			&dailyLimit,
			&weeklyLimit,
			&monthlyLimit,
			&g.DefaultValidityDays,
			&g.SortOrder,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		g.Description = description.String
		g.PricingMode = pricingMode.String
		g.DefaultBudgetMultiplier = nullableFloat64Ptr(defaultBudgetMultiplier)
		g.DailyLimitUSD = nullableFloat64Ptr(dailyLimit)
		g.WeeklyLimitUSD = nullableFloat64Ptr(weeklyLimit)
		g.MonthlyLimitUSD = nullableFloat64Ptr(monthlyLimit)
		g.CreatedAt = createdAt
		g.UpdatedAt = updatedAt
		g.Hydrated = true
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func nullableFloat64Ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}

func nullableTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	return &v.Time
}

func nullableInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}
