package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	entgroup "github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionproduct"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionproductgroup"
	"github.com/Wei-Shaw/sub2api/ent/userproductsubscription"
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

func (r *subscriptionProductRepository) GetActiveProductSubscriptionByUserAndGroupID(ctx context.Context, userID, groupID int64) (*service.SubscriptionProductBinding, *service.UserProductSubscription, error) {
	var binding service.SubscriptionProductBinding
	var sub service.UserProductSubscription
	var dailyWindowStart sql.NullTime
	var weeklyWindowStart sql.NullTime
	var monthlyWindowStart sql.NullTime
	var assignedBy sql.NullInt64
	var notes sql.NullString
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
	ups.daily_carryover_remaining_usd,
	ups.assigned_by,
	ups.assigned_at,
	ups.notes,
	ups.created_at,
	ups.updated_at
FROM user_product_subscriptions ups
JOIN subscription_products sp ON sp.id = ups.product_id
JOIN subscription_product_groups spg ON spg.product_id = sp.id
JOIN groups g ON g.id = spg.group_id
WHERE ups.user_id = $1
  AND spg.group_id = $2
  AND ups.status = 'active'
  AND ups.expires_at > NOW()
  AND ups.deleted_at IS NULL
  AND sp.status = 'active'
  AND sp.deleted_at IS NULL
  AND spg.status = 'active'
  AND spg.deleted_at IS NULL
  AND g.deleted_at IS NULL
ORDER BY ups.assigned_at DESC NULLS LAST, ups.id DESC
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
		&assignedBy,
		&sub.AssignedAt,
		&notes,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	sub.DailyWindowStart = nullableTimePtr(dailyWindowStart)
	sub.WeeklyWindowStart = nullableTimePtr(weeklyWindowStart)
	sub.MonthlyWindowStart = nullableTimePtr(monthlyWindowStart)
	sub.AssignedBy = nullableInt64Ptr(assignedBy)
	sub.Notes = notes.String
	return &binding, &sub, nil
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

func (r *subscriptionProductRepository) ListActiveProductsByUserID(ctx context.Context, userID int64) ([]service.ActiveSubscriptionProduct, error) {
	subs, err := r.client.UserProductSubscription.Query().
		Where(
			userproductsubscription.UserIDEQ(userID),
			userproductsubscription.StatusEQ(service.SubscriptionStatusActive),
			userproductsubscription.ExpiresAtGT(time.Now()),
			userproductsubscription.HasProductWith(subscriptionproduct.StatusEQ(service.SubscriptionProductStatusActive)),
		).
		WithProduct().
		Order(dbent.Asc(userproductsubscription.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return []service.ActiveSubscriptionProduct{}, nil
	}

	productIDs := make([]int64, 0, len(subs))
	for _, sub := range subs {
		productIDs = append(productIDs, sub.ProductID)
	}

	bindings, err := r.client.SubscriptionProductGroup.Query().
		Where(
			subscriptionproductgroup.ProductIDIn(productIDs...),
			subscriptionproductgroup.StatusEQ(service.SubscriptionProductBindingStatusActive),
		).
		Order(dbent.Asc(subscriptionproductgroup.FieldSortOrder), dbent.Asc(subscriptionproductgroup.FieldGroupID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]int64, 0, len(bindings))
	for _, binding := range bindings {
		groupIDs = append(groupIDs, binding.GroupID)
	}
	groupNames, err := r.groupNamesByID(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	bindingsByProduct := make(map[int64][]service.SubscriptionProductGroupSummary, len(productIDs))
	for _, binding := range bindings {
		bindingsByProduct[binding.ProductID] = append(bindingsByProduct[binding.ProductID], service.SubscriptionProductGroupSummary{
			GroupID:         binding.GroupID,
			GroupName:       groupNames[binding.GroupID],
			DebitMultiplier: binding.DebitMultiplier,
			Status:          binding.Status,
			SortOrder:       binding.SortOrder,
		})
	}

	items := make([]service.ActiveSubscriptionProduct, 0, len(subs))
	for _, sub := range subs {
		item := service.ActiveSubscriptionProduct{
			Subscription: userProductSubscriptionEntToService(sub),
			Groups:       bindingsByProduct[sub.ProductID],
		}
		if item.Groups == nil {
			item.Groups = []service.SubscriptionProductGroupSummary{}
		}
		if sub.Edges.Product != nil {
			item.Product = subscriptionProductEntToService(sub.Edges.Product)
			item.Subscription.Product = &item.Product
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *subscriptionProductRepository) groupNamesByID(ctx context.Context, ids []int64) (map[int64]string, error) {
	out := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	groups, err := r.client.Group.Query().
		Where(entgroup.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		out[group.ID] = group.Name
	}
	return out, nil
}

func subscriptionProductEntToService(m *dbent.SubscriptionProduct) service.SubscriptionProduct {
	if m == nil {
		return service.SubscriptionProduct{}
	}
	product := service.SubscriptionProduct{
		ID:                  m.ID,
		Code:                m.Code,
		Name:                m.Name,
		Status:              m.Status,
		DefaultValidityDays: m.DefaultValidityDays,
		DailyLimitUSD:       m.DailyLimitUsd,
		WeeklyLimitUSD:      m.WeeklyLimitUsd,
		MonthlyLimitUSD:     m.MonthlyLimitUsd,
		SortOrder:           m.SortOrder,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
	if m.Description != nil {
		product.Description = *m.Description
	}
	return product
}

func userProductSubscriptionEntToService(m *dbent.UserProductSubscription) service.UserProductSubscription {
	if m == nil {
		return service.UserProductSubscription{}
	}
	sub := service.UserProductSubscription{
		ID:                         m.ID,
		UserID:                     m.UserID,
		ProductID:                  m.ProductID,
		StartsAt:                   m.StartsAt,
		ExpiresAt:                  m.ExpiresAt,
		Status:                     m.Status,
		DailyWindowStart:           m.DailyWindowStart,
		WeeklyWindowStart:          m.WeeklyWindowStart,
		MonthlyWindowStart:         m.MonthlyWindowStart,
		DailyUsageUSD:              m.DailyUsageUsd,
		WeeklyUsageUSD:             m.WeeklyUsageUsd,
		MonthlyUsageUSD:            m.MonthlyUsageUsd,
		DailyCarryoverInUSD:        m.DailyCarryoverInUsd,
		DailyCarryoverRemainingUSD: m.DailyCarryoverRemainingUsd,
		AssignedBy:                 m.AssignedBy,
		AssignedAt:                 m.AssignedAt,
		CreatedAt:                  m.CreatedAt,
		UpdatedAt:                  m.UpdatedAt,
	}
	if m.Notes != nil {
		sub.Notes = *m.Notes
	}
	return sub
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
