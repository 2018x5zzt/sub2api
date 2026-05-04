package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type subscriptionProductRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewSubscriptionProductRepository(client *dbent.Client, sqlDB *sql.DB) service.ProductSubscriptionRepository {
	return &subscriptionProductRepository{client: client, sql: sqlDB}
}

func (r *subscriptionProductRepository) execForContext(ctx context.Context) sqlExecutor {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return r.sql
}

func (r *subscriptionProductRepository) GetActiveProductSubscriptionByUserAndGroupID(ctx context.Context, userID, groupID int64, productFamily *string) (*service.SubscriptionProductBinding, *service.UserProductSubscription, error) {
	normalizedFamily := ""
	if productFamily != nil {
		normalizedFamily = normalizeProductFamily(*productFamily)
	}
	rows, err := r.sql.QueryContext(ctx, `
SELECT
	sp.id,
	sp.code,
	sp.name,
	sp.product_family,
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
  AND ($3 = '' OR sp.product_family = $3)
ORDER BY sp.product_family ASC, sp.sort_order ASC, ups.starts_at ASC, ups.id ASC
`, userID, groupID, normalizedFamily)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	var firstFamily string
	var firstLimitErr error
	var firstAvailableBinding *service.SubscriptionProductBinding
	var firstAvailableSub *service.UserProductSubscription
	seenFamilies := make(map[string]struct{})
	seen := false
	for rows.Next() {
		seen = true
		var binding service.SubscriptionProductBinding
		var sub service.UserProductSubscription
		var dailyWindowStart sql.NullTime
		var weeklyWindowStart sql.NullTime
		var monthlyWindowStart sql.NullTime
		if err := rows.Scan(
			&binding.ProductID,
			&binding.ProductCode,
			&binding.ProductName,
			&binding.ProductFamily,
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
		); err != nil {
			return nil, nil, err
		}
		if firstFamily == "" {
			firstFamily = strings.TrimSpace(binding.ProductFamily)
			if firstFamily == "" {
				firstFamily = "gpt"
			}
		}
		family := strings.TrimSpace(binding.ProductFamily)
		if family == "" {
			family = "gpt"
		}
		seenFamilies[family] = struct{}{}
		if normalizedFamily == "" && len(seenFamilies) > 1 {
			return nil, nil, service.ErrProductFamilyRequired
		}
		if family != firstFamily {
			break
		}
		sub.DailyWindowStart = nullableTimePtr(dailyWindowStart)
		sub.WeeklyWindowStart = nullableTimePtr(weeklyWindowStart)
		sub.MonthlyWindowStart = nullableTimePtr(monthlyWindowStart)
		product := binding.Product()
		service.NormalizeExpiredProductSubscriptionWindowForRepository(&sub, product, time.Now())
		if err := productSubscriptionHasRemaining(&sub, product); err != nil {
			if firstLimitErr == nil {
				firstLimitErr = err
			}
			continue
		}
		if normalizedFamily != "" {
			return &binding, &sub, nil
		}
		if firstAvailableBinding == nil {
			bindingCopy := binding
			subCopy := sub
			firstAvailableBinding = &bindingCopy
			firstAvailableSub = &subCopy
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if normalizedFamily == "" {
		switch len(seenFamilies) {
		case 0:
			// handled below
		case 1:
			if firstAvailableBinding != nil && firstAvailableSub != nil {
				return firstAvailableBinding, firstAvailableSub, nil
			}
		default:
			return nil, nil, service.ErrProductFamilyRequired
		}
	}
	if !seen {
		return nil, nil, service.ErrSubscriptionNotFound
	}
	if firstAvailableBinding != nil && firstAvailableSub != nil {
		return firstAvailableBinding, firstAvailableSub, nil
	}
	if firstLimitErr != nil {
		return nil, nil, firstLimitErr
	}
	return nil, nil, service.ErrSubscriptionNotFound
}

func (r *subscriptionProductRepository) ListActiveProductsByUserID(ctx context.Context, userID int64) ([]service.ActiveSubscriptionProduct, error) {
	rows, err := r.sql.QueryContext(ctx, `
SELECT
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
	ups.updated_at,
	sp.id,
	sp.code,
	sp.name,
	COALESCE(sp.description, ''),
	sp.status,
	sp.product_family,
	sp.default_validity_days,
	sp.daily_limit_usd,
	sp.weekly_limit_usd,
	sp.monthly_limit_usd,
	sp.sort_order,
	sp.created_at,
	sp.updated_at
FROM user_product_subscriptions ups
JOIN subscription_products sp
	ON sp.id = ups.product_id
	AND sp.deleted_at IS NULL
	AND sp.status = 'active'
WHERE ups.user_id = $1
  AND ups.status = 'active'
  AND ups.expires_at > NOW()
  AND ups.deleted_at IS NULL
ORDER BY sp.sort_order ASC, ups.expires_at DESC, ups.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.ActiveSubscriptionProduct, 0)
	productIDs := make([]int64, 0)
	for rows.Next() {
		var item service.ActiveSubscriptionProduct
		var dailyWindowStart, weeklyWindowStart, monthlyWindowStart sql.NullTime
		var assignedBy sql.NullInt64
		var notes sql.NullString
		if err := rows.Scan(
			&item.Subscription.ID,
			&item.Subscription.UserID,
			&item.Subscription.ProductID,
			&item.Subscription.StartsAt,
			&item.Subscription.ExpiresAt,
			&item.Subscription.Status,
			&dailyWindowStart,
			&weeklyWindowStart,
			&monthlyWindowStart,
			&item.Subscription.DailyUsageUSD,
			&item.Subscription.WeeklyUsageUSD,
			&item.Subscription.MonthlyUsageUSD,
			&item.Subscription.DailyCarryoverInUSD,
			&item.Subscription.DailyCarryoverRemainingUSD,
			&assignedBy,
			&item.Subscription.AssignedAt,
			&notes,
			&item.Subscription.CreatedAt,
			&item.Subscription.UpdatedAt,
			&item.Product.ID,
			&item.Product.Code,
			&item.Product.Name,
			&item.Product.Description,
			&item.Product.Status,
			&item.Product.ProductFamily,
			&item.Product.DefaultValidityDays,
			&item.Product.DailyLimitUSD,
			&item.Product.WeeklyLimitUSD,
			&item.Product.MonthlyLimitUSD,
			&item.Product.SortOrder,
			&item.Product.CreatedAt,
			&item.Product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.Subscription.DailyWindowStart = nullableTimePtr(dailyWindowStart)
		item.Subscription.WeeklyWindowStart = nullableTimePtr(weeklyWindowStart)
		item.Subscription.MonthlyWindowStart = nullableTimePtr(monthlyWindowStart)
		if assignedBy.Valid {
			v := assignedBy.Int64
			item.Subscription.AssignedBy = &v
		}
		item.Subscription.Notes = notes.String
		service.NormalizeExpiredProductSubscriptionWindowForRepository(&item.Subscription, &item.Product, time.Now())
		item.Groups = []service.SubscriptionProductGroupSummary{}
		items = append(items, item)
		productIDs = append(productIDs, item.Product.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []service.ActiveSubscriptionProduct{}, nil
	}

	groupsByProduct, err := r.productGroupsByProductID(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if groups := groupsByProduct[items[i].Product.ID]; groups != nil {
			items[i].Groups = groups
		}
	}
	return items, nil
}

func (r *subscriptionProductRepository) ListVisibleGroupsByUserID(ctx context.Context, userID int64) ([]service.Group, error) {
	rows, err := r.sql.QueryContext(ctx, `
SELECT DISTINCT
	g.id,
	g.name,
	COALESCE(g.description, ''),
	g.platform,
	g.rate_multiplier,
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
JOIN subscription_products sp
	ON sp.id = ups.product_id
	AND sp.deleted_at IS NULL
	AND sp.status = 'active'
JOIN subscription_product_groups spg
	ON spg.product_id = sp.id
	AND spg.deleted_at IS NULL
	AND spg.status = 'active'
JOIN groups g
	ON g.id = spg.group_id
	AND g.deleted_at IS NULL
	AND g.status = 'active'
WHERE ups.user_id = $1
  AND ups.status = 'active'
  AND ups.expires_at > NOW()
  AND ups.deleted_at IS NULL
ORDER BY g.sort_order ASC, g.id ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	groups := make([]service.Group, 0)
	for rows.Next() {
		var g service.Group
		var dailyLimit, weeklyLimit, monthlyLimit sql.NullFloat64
		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.Description,
			&g.Platform,
			&g.RateMultiplier,
			&g.IsExclusive,
			&g.Status,
			&g.SubscriptionType,
			&dailyLimit,
			&weeklyLimit,
			&monthlyLimit,
			&g.DefaultValidityDays,
			&g.SortOrder,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		g.DailyLimitUSD = nullableProductFloat64Ptr(dailyLimit)
		g.WeeklyLimitUSD = nullableProductFloat64Ptr(weeklyLimit)
		g.MonthlyLimitUSD = nullableProductFloat64Ptr(monthlyLimit)
		g.Hydrated = true
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *subscriptionProductRepository) productGroupsByProductID(ctx context.Context, productIDs []int64) (map[int64][]service.SubscriptionProductGroupSummary, error) {
	out := make(map[int64][]service.SubscriptionProductGroupSummary, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}
	rows, err := r.sql.QueryContext(ctx, `
SELECT
	spg.product_id,
	spg.group_id,
	g.name,
	COALESCE(g.platform, ''),
	spg.debit_multiplier,
	spg.status,
	spg.sort_order
FROM subscription_product_groups spg
JOIN groups g
	ON g.id = spg.group_id
	AND g.deleted_at IS NULL
WHERE spg.product_id = ANY($1)
  AND spg.deleted_at IS NULL
  AND spg.status = 'active'
ORDER BY spg.product_id ASC, spg.sort_order ASC, spg.group_id ASC`, pqInt64Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var productID int64
		var group service.SubscriptionProductGroupSummary
		if err := rows.Scan(
			&productID,
			&group.GroupID,
			&group.GroupName,
			&group.GroupPlatform,
			&group.DebitMultiplier,
			&group.Status,
			&group.SortOrder,
		); err != nil {
			return nil, err
		}
		out[productID] = append(out[productID], group)
	}
	return out, rows.Err()
}

func (r *subscriptionProductRepository) ListProducts(ctx context.Context) ([]service.SubscriptionProduct, error) {
	rows, err := r.execForContext(ctx).QueryContext(ctx, `
SELECT
	id,
	code,
	name,
	COALESCE(description, ''),
	status,
	product_family,
	default_validity_days,
	daily_limit_usd,
	weekly_limit_usd,
	monthly_limit_usd,
	sort_order,
	created_at,
	updated_at
FROM subscription_products
WHERE deleted_at IS NULL
ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	products := make([]service.SubscriptionProduct, 0)
	for rows.Next() {
		var product service.SubscriptionProduct
		if err := rows.Scan(
			&product.ID,
			&product.Code,
			&product.Name,
			&product.Description,
			&product.Status,
			&product.ProductFamily,
			&product.DefaultValidityDays,
			&product.DailyLimitUSD,
			&product.WeeklyLimitUSD,
			&product.MonthlyLimitUSD,
			&product.SortOrder,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *subscriptionProductRepository) ResolveActiveProductByGroupID(ctx context.Context, groupID int64) (*service.SubscriptionProduct, error) {
	if groupID <= 0 {
		return nil, service.ErrSubscriptionNotFound
	}
	exec := r.execForContext(ctx)
	var product service.SubscriptionProduct
	err := scanSingleRow(ctx, exec, `
SELECT
	sp.id,
	sp.code,
	sp.name,
	COALESCE(sp.description, ''),
	sp.status,
	sp.product_family,
	sp.default_validity_days,
	sp.daily_limit_usd,
	sp.weekly_limit_usd,
	sp.monthly_limit_usd,
	sp.sort_order,
	sp.created_at,
	sp.updated_at
FROM subscription_product_groups spg
JOIN subscription_products sp
	ON sp.id = spg.product_id
	AND sp.deleted_at IS NULL
	AND sp.status = 'active'
WHERE spg.group_id = $1
  AND spg.deleted_at IS NULL
  AND spg.status = 'active'
ORDER BY sp.sort_order ASC, spg.sort_order ASC, sp.id ASC
LIMIT 1`, []any{groupID},
		&product.ID,
		&product.Code,
		&product.Name,
		&product.Description,
		&product.Status,
		&product.ProductFamily,
		&product.DefaultValidityDays,
		&product.DailyLimitUSD,
		&product.WeeklyLimitUSD,
		&product.MonthlyLimitUSD,
		&product.SortOrder,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("resolve subscription product by group: %w", err)
	}
	return &product, nil
}

func (r *subscriptionProductRepository) CreateProduct(ctx context.Context, input *service.CreateSubscriptionProductInput) (*service.SubscriptionProduct, error) {
	if input == nil {
		return nil, service.ErrSubscriptionNilInput
	}
	code := strings.TrimSpace(input.Code)
	name := strings.TrimSpace(input.Name)
	if code == "" || name == "" {
		return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "code and name are required")
	}
	status, err := normalizeSubscriptionProductStatus(input.Status)
	if err != nil {
		return nil, err
	}
	if err := validateProductLimits(input.DailyLimitUSD, input.WeeklyLimitUSD, input.MonthlyLimitUSD); err != nil {
		return nil, err
	}
	validityDays := normalizeProductValidityDays(input.DefaultValidityDays, 30)

	var createdID int64
	err = scanSingleRow(ctx, r.execForContext(ctx), `
INSERT INTO subscription_products (
	code,
	name,
	description,
	status,
	product_family,
	default_validity_days,
	daily_limit_usd,
	weekly_limit_usd,
	monthly_limit_usd,
	sort_order,
	created_at,
	updated_at
) VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
RETURNING id`, []any{
		code,
		name,
		strings.TrimSpace(input.Description),
		status,
		normalizeProductFamily(input.ProductFamily),
		validityDays,
		input.DailyLimitUSD,
		input.WeeklyLimitUSD,
		input.MonthlyLimitUSD,
		input.SortOrder,
	}, &createdID)
	if err != nil {
		return nil, fmt.Errorf("create subscription product: %w", err)
	}
	return r.getProductByID(ctx, r.execForContext(ctx), createdID)
}

func (r *subscriptionProductRepository) UpdateProduct(ctx context.Context, productID int64, input *service.UpdateSubscriptionProductInput) (*service.SubscriptionProduct, error) {
	if productID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "product_id is required")
	}
	if input == nil {
		return nil, service.ErrSubscriptionNilInput
	}

	exec := r.execForContext(ctx)
	if _, err := r.getProductByID(ctx, exec, productID); err != nil {
		return nil, err
	}

	sets := []string{"updated_at = NOW()"}
	args := make([]any, 0, 10)
	addSet := func(field string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", field, len(args)))
	}

	if input.Code != nil {
		code := strings.TrimSpace(*input.Code)
		if code == "" {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "code cannot be empty")
		}
		addSet("code", code)
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "name cannot be empty")
		}
		addSet("name", name)
	}
	if input.Description != nil {
		addSet("description", strings.TrimSpace(*input.Description))
	}
	if input.Status != nil {
		status, err := normalizeSubscriptionProductStatus(*input.Status)
		if err != nil {
			return nil, err
		}
		addSet("status", status)
	}
	if input.ProductFamily != nil {
		addSet("product_family", normalizeProductFamily(*input.ProductFamily))
	}
	if input.DefaultValidityDays != nil {
		if *input.DefaultValidityDays <= 0 || *input.DefaultValidityDays > service.MaxValidityDays {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "default_validity_days is out of range")
		}
		addSet("default_validity_days", *input.DefaultValidityDays)
	}
	if input.DailyLimitUSD != nil {
		if *input.DailyLimitUSD < 0 {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "daily_limit_usd cannot be negative")
		}
		addSet("daily_limit_usd", *input.DailyLimitUSD)
	}
	if input.WeeklyLimitUSD != nil {
		if *input.WeeklyLimitUSD < 0 {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "weekly_limit_usd cannot be negative")
		}
		addSet("weekly_limit_usd", *input.WeeklyLimitUSD)
	}
	if input.MonthlyLimitUSD != nil {
		if *input.MonthlyLimitUSD < 0 {
			return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "monthly_limit_usd cannot be negative")
		}
		addSet("monthly_limit_usd", *input.MonthlyLimitUSD)
	}
	if input.SortOrder != nil {
		addSet("sort_order", *input.SortOrder)
	}

	args = append(args, productID)
	if _, err := exec.ExecContext(ctx, `
UPDATE subscription_products
SET `+strings.Join(sets, ", ")+`
WHERE id = $`+strconv.Itoa(len(args))+`
  AND deleted_at IS NULL`, args...); err != nil {
		return nil, fmt.Errorf("update subscription product: %w", err)
	}
	return r.getProductByID(ctx, exec, productID)
}

func (r *subscriptionProductRepository) SyncProductBindings(ctx context.Context, productID int64, inputs []service.SubscriptionProductBindingInput) ([]service.SubscriptionProductBindingDetail, error) {
	if productID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDINGS", "product_id is required")
	}
	exec := r.execForContext(ctx)
	if _, err := r.getProductByID(ctx, exec, productID); err != nil {
		return nil, err
	}

	groupIDs := make([]int64, 0, len(inputs))
	for _, input := range inputs {
		groupIDs = append(groupIDs, input.GroupID)
	}
	if err := r.ensureSubscriptionGroups(ctx, exec, groupIDs); err != nil {
		return nil, err
	}

	seen := make(map[int64]struct{}, len(inputs))
	for _, input := range inputs {
		normalized, err := normalizeProductBindingInput(input)
		if err != nil {
			return nil, err
		}
		seen[normalized.GroupID] = struct{}{}
		var existingID int64
		err = scanSingleRow(ctx, exec, `
SELECT id
FROM subscription_product_groups
WHERE product_id = $1
  AND group_id = $2
  AND deleted_at IS NULL
ORDER BY id DESC
LIMIT 1`, []any{productID, normalized.GroupID}, &existingID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get subscription product binding: %w", err)
		}
		if err == nil {
			if _, err := exec.ExecContext(ctx, `
UPDATE subscription_product_groups
SET debit_multiplier = $1,
    status = $2,
    sort_order = $3,
    updated_at = NOW()
WHERE id = $4`, normalized.DebitMultiplier, normalized.Status, normalized.SortOrder, existingID); err != nil {
				return nil, fmt.Errorf("update subscription product binding: %w", err)
			}
			continue
		}
		if _, err := exec.ExecContext(ctx, `
INSERT INTO subscription_product_groups (
	product_id,
	group_id,
	debit_multiplier,
	status,
	sort_order,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`, productID, normalized.GroupID, normalized.DebitMultiplier, normalized.Status, normalized.SortOrder); err != nil {
			return nil, fmt.Errorf("create subscription product binding: %w", err)
		}
	}

	existingRows, err := exec.QueryContext(ctx, `
SELECT group_id
FROM subscription_product_groups
WHERE product_id = $1
  AND deleted_at IS NULL
  AND status <> $2`, productID, service.SubscriptionProductBindingStatusInactive)
	if err != nil {
		return nil, fmt.Errorf("list subscription product bindings: %w", err)
	}
	defer func() { _ = existingRows.Close() }()
	for existingRows.Next() {
		var groupID int64
		if err := existingRows.Scan(&groupID); err != nil {
			return nil, err
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		if _, err := exec.ExecContext(ctx, `
UPDATE subscription_product_groups
SET status = $1,
    updated_at = NOW()
WHERE product_id = $2
  AND group_id = $3
  AND deleted_at IS NULL`, service.SubscriptionProductBindingStatusInactive, productID, groupID); err != nil {
			return nil, fmt.Errorf("deactivate subscription product binding: %w", err)
		}
	}
	if err := existingRows.Err(); err != nil {
		return nil, err
	}

	return r.listProductBindingDetails(ctx, exec, productID)
}

func (r *subscriptionProductRepository) ensureSubscriptionGroups(ctx context.Context, exec sqlExecutor, groupIDs []int64) error {
	if len(groupIDs) == 0 {
		return nil
	}
	rows, err := exec.QueryContext(ctx, `
SELECT id, subscription_type
FROM groups
WHERE id = ANY($1)
  AND deleted_at IS NULL`, pqInt64Array(groupIDs))
	if err != nil {
		return fmt.Errorf("validate subscription product bindings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	found := make(map[int64]string, len(groupIDs))
	for rows.Next() {
		var groupID int64
		var subscriptionType string
		if err := rows.Scan(&groupID, &subscriptionType); err != nil {
			return err
		}
		found[groupID] = subscriptionType
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, groupID := range groupIDs {
		subscriptionType, ok := found[groupID]
		if !ok {
			return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDINGS", fmt.Sprintf("group %d not found", groupID))
		}
		if strings.TrimSpace(subscriptionType) != service.SubscriptionTypeSubscription {
			return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDINGS", "subscription product bindings can only target subscription groups")
		}
	}
	return nil
}

func (r *subscriptionProductRepository) ListProductBindings(ctx context.Context, productID int64) ([]service.SubscriptionProductBindingDetail, error) {
	if productID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDINGS", "product_id is required")
	}
	exec := r.execForContext(ctx)
	if _, err := r.getProductByID(ctx, exec, productID); err != nil {
		return nil, err
	}
	return r.listProductBindingDetails(ctx, exec, productID)
}

func (r *subscriptionProductRepository) ListProductSubscriptions(ctx context.Context, productID int64) ([]service.UserProductSubscription, error) {
	if productID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_PRODUCT_SUBSCRIPTION_LIST", "product_id is required")
	}
	exec := r.execForContext(ctx)
	product, err := r.getProductByID(ctx, exec, productID)
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, `
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
WHERE product_id = $1
  AND deleted_at IS NULL
ORDER BY expires_at DESC, id DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	subs := make([]service.UserProductSubscription, 0)
	now := time.Now()
	for rows.Next() {
		sub, err := scanUserProductSubscriptionFromRows(rows)
		if err != nil {
			return nil, err
		}
		if sub.Status == service.SubscriptionStatusActive && sub.ExpiresAt.After(now) {
			service.NormalizeExpiredProductSubscriptionWindowForRepository(sub, product, now)
		}
		subs = append(subs, *sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *subscriptionProductRepository) ListUserProductSubscriptionsForAdmin(ctx context.Context, params service.AdminProductSubscriptionListParams) ([]service.AdminProductSubscriptionListItem, *pagination.PaginationResult, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	sortOrder := pagination.NormalizeSortOrder(params.SortOrder, pagination.SortOrderDesc)
	sortBy := strings.TrimSpace(params.SortBy)
	orderExpr := "ups.expires_at DESC, ups.id DESC"
	switch sortBy {
	case "created_at":
		if sortOrder == pagination.SortOrderAsc {
			orderExpr = "ups.created_at ASC, ups.id ASC"
		} else {
			orderExpr = "ups.created_at DESC, ups.id DESC"
		}
	case "daily_usage_usd":
		if sortOrder == pagination.SortOrderAsc {
			orderExpr = "ups.daily_usage_usd ASC, ups.id ASC"
		} else {
			orderExpr = "ups.daily_usage_usd DESC, ups.id DESC"
		}
	case "expires_at", "":
		if sortOrder == pagination.SortOrderAsc {
			orderExpr = "ups.expires_at ASC, ups.id ASC"
		}
	}

	where := []string{"ups.deleted_at IS NULL", "sp.deleted_at IS NULL", "u.deleted_at IS NULL"}
	args := make([]any, 0, 8)
	if params.ProductID > 0 {
		args = append(args, params.ProductID)
		where = append(where, fmt.Sprintf("ups.product_id = $%d", len(args)))
	}
	if params.UserID > 0 {
		args = append(args, params.UserID)
		where = append(where, fmt.Sprintf("ups.user_id = $%d", len(args)))
	}
	if status := strings.TrimSpace(params.Status); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("ups.status = $%d", len(args)))
	}
	if search := strings.TrimSpace(params.Search); search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf("(LOWER(u.email) LIKE $%d OR LOWER(u.username) LIKE $%d OR CAST(u.id AS TEXT) LIKE $%d OR LOWER(sp.name) LIKE $%d OR LOWER(sp.code) LIKE $%d)", idx, idx, idx, idx, idx))
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM user_product_subscriptions ups
JOIN users u ON u.id = ups.user_id
JOIN subscription_products sp ON sp.id = ups.product_id
WHERE ` + whereSQL
	if err := scanSingleRow(ctx, r.execForContext(ctx), countQuery, args, &total); err != nil {
		return nil, nil, fmt.Errorf("count admin product subscriptions: %w", err)
	}

	limit := pageSize
	offset := (page - 1) * pageSize
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitPos := len(listArgs) - 1
	offsetPos := len(listArgs)
	query := `
SELECT
	ups.id,
	ups.user_id,
	u.email,
	u.username,
	ups.product_id,
	sp.code,
	sp.name,
	sp.daily_limit_usd,
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
	GREATEST(ups.daily_carryover_in_usd - ups.daily_carryover_remaining_usd, 0) AS carryover_used_usd,
	GREATEST(ups.daily_usage_usd - GREATEST(ups.daily_carryover_in_usd - ups.daily_carryover_remaining_usd, 0), 0) AS fresh_daily_usage_usd,
	ups.assigned_by,
	ups.assigned_at,
	ups.notes,
	ups.created_at,
	ups.updated_at
FROM user_product_subscriptions ups
JOIN users u ON u.id = ups.user_id
JOIN subscription_products sp ON sp.id = ups.product_id
WHERE ` + whereSQL + `
ORDER BY ` + orderExpr + `
LIMIT $` + strconv.Itoa(limitPos) + ` OFFSET $` + strconv.Itoa(offsetPos)

	rows, err := r.execForContext(ctx).QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("list admin product subscriptions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.AdminProductSubscriptionListItem, 0, limit)
	for rows.Next() {
		var item service.AdminProductSubscriptionListItem
		var dailyWindow, weeklyWindow, monthlyWindow sql.NullTime
		var assignedBy sql.NullInt64
		var notes sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.UserEmail,
			&item.UserUsername,
			&item.ProductID,
			&item.ProductCode,
			&item.ProductName,
			&item.DailyLimitUSD,
			&item.StartsAt,
			&item.ExpiresAt,
			&item.Status,
			&dailyWindow,
			&weeklyWindow,
			&monthlyWindow,
			&item.DailyUsageUSD,
			&item.WeeklyUsageUSD,
			&item.MonthlyUsageUSD,
			&item.DailyCarryoverInUSD,
			&item.DailyCarryoverRemainingUSD,
			&item.CarryoverUsedUSD,
			&item.FreshDailyUsageUSD,
			&assignedBy,
			&item.AssignedAt,
			&notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		item.DailyWindowStart = nullableTimePtr(dailyWindow)
		item.WeeklyWindowStart = nullableTimePtr(weeklyWindow)
		item.MonthlyWindowStart = nullableTimePtr(monthlyWindow)
		if assignedBy.Valid {
			v := assignedBy.Int64
			item.AssignedBy = &v
		}
		item.Notes = notes.String
		normalizeAdminProductSubscriptionListItem(&item)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	pages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pages < 1 {
		pages = 1
	}
	return items, &pagination.PaginationResult{Total: total, Page: page, PageSize: pageSize, Pages: pages}, nil
}

func normalizeAdminProductSubscriptionListItem(item *service.AdminProductSubscriptionListItem) {
	if item == nil {
		return
	}
	product := &service.SubscriptionProduct{DailyLimitUSD: item.DailyLimitUSD}
	service.NormalizeExpiredProductSubscriptionWindowForRepository(&item.UserProductSubscription, product, time.Now())
	item.CarryoverUsedUSD = item.DailyCarryoverInUSD - item.DailyCarryoverRemainingUSD
	if item.CarryoverUsedUSD < 0 {
		item.CarryoverUsedUSD = 0
	}
	item.FreshDailyUsageUSD = item.DailyUsageUSD - item.CarryoverUsedUSD
	if item.FreshDailyUsageUSD < 0 {
		item.FreshDailyUsageUSD = 0
	}
}

func (r *subscriptionProductRepository) AssignOrExtendProductSubscription(ctx context.Context, input *service.AssignProductSubscriptionInput) (*service.UserProductSubscription, bool, error) {
	if input == nil {
		return nil, false, service.ErrSubscriptionNilInput
	}
	if input.UserID <= 0 || input.ProductID <= 0 {
		return nil, false, infraerrors.BadRequest("INVALID_PRODUCT_SUBSCRIPTION_ASSIGNMENT", "user_id and product_id are required")
	}

	exec := r.execForContext(ctx)
	validityDays := input.ValidityDays
	if validityDays <= 0 {
		var defaultDays int
		err := scanSingleRow(ctx, exec, `
SELECT default_validity_days
FROM subscription_products
WHERE id = $1
  AND deleted_at IS NULL
  AND status = 'active'
LIMIT 1`, []any{input.ProductID}, &defaultDays)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, false, service.ErrSubscriptionNotFound
			}
			return nil, false, fmt.Errorf("get subscription product: %w", err)
		}
		validityDays = defaultDays
	}
	if validityDays <= 0 {
		validityDays = 30
	}

	var existing service.UserProductSubscription
	var existingNotes sql.NullString
	var existingDailyWindow, existingWeeklyWindow, existingMonthlyWindow sql.NullTime
	err := scanSingleRow(ctx, exec, `
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
  AND deleted_at IS NULL
ORDER BY id DESC
LIMIT 1`, []any{input.UserID, input.ProductID},
		&existing.ID,
		&existing.UserID,
		&existing.ProductID,
		&existing.StartsAt,
		&existing.ExpiresAt,
		&existing.Status,
		&existingDailyWindow,
		&existingWeeklyWindow,
		&existingMonthlyWindow,
		&existing.DailyUsageUSD,
		&existing.WeeklyUsageUSD,
		&existing.MonthlyUsageUSD,
		&existing.DailyCarryoverInUSD,
		&existing.DailyCarryoverRemainingUSD,
		&existing.AssignedBy,
		&existing.AssignedAt,
		&existingNotes,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("get user product subscription: %w", err)
	}
	if err == nil {
		now := time.Now()
		base := existing.ExpiresAt
		if base.Before(now) {
			base = now
		}
		newExpiresAt := base.AddDate(0, 0, validityDays)
		newNotes := appendSubscriptionNotes(existingNotes.String, input.Notes)
		if _, err := exec.ExecContext(ctx, `
UPDATE user_product_subscriptions
SET starts_at = CASE WHEN starts_at > NOW() THEN NOW() ELSE starts_at END,
    expires_at = $1,
    status = 'active',
    assigned_by = NULLIF($2, 0),
    assigned_at = NOW(),
    notes = $3,
    updated_at = NOW()
WHERE id = $4`, newExpiresAt, input.AssignedBy, newNotes, existing.ID); err != nil {
			return nil, false, fmt.Errorf("extend user product subscription: %w", err)
		}
		sub, getErr := r.getUserProductSubscriptionByID(ctx, exec, existing.ID)
		return sub, true, getErr
	}

	var productExists int
	if err := scanSingleRow(ctx, exec, `
SELECT 1
FROM subscription_products
WHERE id = $1
  AND deleted_at IS NULL
  AND status = 'active'
LIMIT 1`, []any{input.ProductID}, &productExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, service.ErrSubscriptionNotFound
		}
		return nil, false, fmt.Errorf("get subscription product: %w", err)
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 0, validityDays)
	var createdID int64
	err = scanSingleRow(ctx, exec, `
INSERT INTO user_product_subscriptions (
	user_id,
	product_id,
	starts_at,
	expires_at,
	status,
	assigned_by,
	assigned_at,
	notes,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, 'active', NULLIF($5, 0), NOW(), $6, NOW(), NOW())
RETURNING id`, []any{input.UserID, input.ProductID, now, expiresAt, input.AssignedBy, strings.TrimSpace(input.Notes)}, &createdID)
	if err != nil {
		return nil, false, fmt.Errorf("create user product subscription: %w", err)
	}
	sub, getErr := r.getUserProductSubscriptionByID(ctx, exec, createdID)
	return sub, false, getErr
}

func (r *subscriptionProductRepository) AdjustProductSubscription(ctx context.Context, subscriptionID int64, input *service.AdjustProductSubscriptionInput) (*service.UserProductSubscription, error) {
	if subscriptionID <= 0 || input == nil {
		return nil, infraerrors.BadRequest("INVALID_PRODUCT_SUBSCRIPTION", "subscription_id is required")
	}
	sets := []string{"updated_at = NOW()"}
	args := make([]any, 0, 4)
	addSet := func(field string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", field, len(args)))
	}
	if input.ExpiresAt != nil {
		addSet("expires_at", *input.ExpiresAt)
	}
	if input.Status != nil {
		addSet("status", strings.TrimSpace(*input.Status))
	}
	if input.Notes != nil {
		addSet("notes", strings.TrimSpace(*input.Notes))
	}
	if len(sets) == 1 {
		return r.getUserProductSubscriptionByID(ctx, r.execForContext(ctx), subscriptionID)
	}
	args = append(args, subscriptionID)
	if _, err := r.execForContext(ctx).ExecContext(ctx, `
UPDATE user_product_subscriptions
SET `+strings.Join(sets, ", ")+`
WHERE id = $`+strconv.Itoa(len(args))+`
  AND deleted_at IS NULL`, args...); err != nil {
		return nil, fmt.Errorf("adjust user product subscription: %w", err)
	}
	return r.getUserProductSubscriptionByID(ctx, r.execForContext(ctx), subscriptionID)
}

func (r *subscriptionProductRepository) ResetProductSubscriptionQuota(ctx context.Context, subscriptionID int64, input *service.ResetProductSubscriptionQuotaInput) (*service.UserProductSubscription, error) {
	if subscriptionID <= 0 || input == nil || (!input.Daily && !input.Weekly && !input.Monthly) {
		return nil, infraerrors.BadRequest("INVALID_PRODUCT_SUBSCRIPTION_RESET", "at least one quota window must be selected")
	}
	sets := []string{"updated_at = NOW()"}
	if input.Daily {
		sets = append(sets,
			"daily_usage_usd = 0",
			"daily_carryover_in_usd = 0",
			"daily_carryover_remaining_usd = 0",
			"daily_window_start = date_trunc('day', NOW() AT TIME ZONE 'Asia/Shanghai') AT TIME ZONE 'Asia/Shanghai'",
		)
	}
	if input.Weekly {
		sets = append(sets,
			"weekly_usage_usd = 0",
			"weekly_window_start = NOW()",
		)
	}
	if input.Monthly {
		sets = append(sets,
			"monthly_usage_usd = 0",
			"monthly_window_start = NOW()",
		)
	}
	if _, err := r.execForContext(ctx).ExecContext(ctx, `
UPDATE user_product_subscriptions
SET `+strings.Join(sets, ", ")+`
WHERE id = $1
  AND deleted_at IS NULL`, subscriptionID); err != nil {
		return nil, fmt.Errorf("reset user product subscription quota: %w", err)
	}
	return r.getUserProductSubscriptionByID(ctx, r.execForContext(ctx), subscriptionID)
}

func (r *subscriptionProductRepository) RevokeProductSubscription(ctx context.Context, subscriptionID int64) error {
	if subscriptionID <= 0 {
		return infraerrors.BadRequest("INVALID_PRODUCT_SUBSCRIPTION", "subscription_id is required")
	}
	result, err := r.execForContext(ctx).ExecContext(ctx, `
UPDATE user_product_subscriptions
SET status = $1,
    updated_at = NOW()
WHERE id = $2
  AND deleted_at IS NULL`, service.ProductSubscriptionStatusRevoked, subscriptionID)
	if err != nil {
		return fmt.Errorf("revoke user product subscription: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return service.ErrSubscriptionNotFound
	}
	return nil
}

func (r *subscriptionProductRepository) getUserProductSubscriptionByID(ctx context.Context, exec sqlExecutor, id int64) (*service.UserProductSubscription, error) {
	var sub service.UserProductSubscription
	var dailyWindow, weeklyWindow, monthlyWindow sql.NullTime
	var assignedBy sql.NullInt64
	var notes sql.NullString
	err := scanSingleRow(ctx, exec, `
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
WHERE id = $1
  AND deleted_at IS NULL`, []any{id},
		&sub.ID,
		&sub.UserID,
		&sub.ProductID,
		&sub.StartsAt,
		&sub.ExpiresAt,
		&sub.Status,
		&dailyWindow,
		&weeklyWindow,
		&monthlyWindow,
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
			return nil, service.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("get user product subscription: %w", err)
	}
	sub.DailyWindowStart = nullableTimePtr(dailyWindow)
	sub.WeeklyWindowStart = nullableTimePtr(weeklyWindow)
	sub.MonthlyWindowStart = nullableTimePtr(monthlyWindow)
	if assignedBy.Valid {
		v := assignedBy.Int64
		sub.AssignedBy = &v
	}
	sub.Notes = notes.String
	return &sub, nil
}

func (r *subscriptionProductRepository) getProductByID(ctx context.Context, exec sqlExecutor, id int64) (*service.SubscriptionProduct, error) {
	var product service.SubscriptionProduct
	err := scanSingleRow(ctx, exec, `
SELECT
	id,
	code,
	name,
	COALESCE(description, ''),
	status,
	product_family,
	default_validity_days,
	daily_limit_usd,
	weekly_limit_usd,
	monthly_limit_usd,
	sort_order,
	created_at,
	updated_at
FROM subscription_products
WHERE id = $1
  AND deleted_at IS NULL`, []any{id},
		&product.ID,
		&product.Code,
		&product.Name,
		&product.Description,
		&product.Status,
		&product.ProductFamily,
		&product.DefaultValidityDays,
		&product.DailyLimitUSD,
		&product.WeeklyLimitUSD,
		&product.MonthlyLimitUSD,
		&product.SortOrder,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("get subscription product: %w", err)
	}
	return &product, nil
}

func (r *subscriptionProductRepository) listProductBindingDetails(ctx context.Context, exec sqlExecutor, productID int64) ([]service.SubscriptionProductBindingDetail, error) {
	rows, err := exec.QueryContext(ctx, `
SELECT
	spg.product_id,
	spg.group_id,
	g.name,
	spg.debit_multiplier,
	spg.status,
	spg.sort_order,
	spg.created_at,
	spg.updated_at
FROM subscription_product_groups spg
JOIN groups g
	ON g.id = spg.group_id
	AND g.deleted_at IS NULL
WHERE spg.product_id = $1
  AND spg.deleted_at IS NULL
ORDER BY spg.sort_order ASC, spg.group_id ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	bindings := make([]service.SubscriptionProductBindingDetail, 0)
	for rows.Next() {
		var binding service.SubscriptionProductBindingDetail
		if err := rows.Scan(
			&binding.ProductID,
			&binding.GroupID,
			&binding.GroupName,
			&binding.DebitMultiplier,
			&binding.Status,
			&binding.SortOrder,
			&binding.CreatedAt,
			&binding.UpdatedAt,
		); err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bindings, nil
}

type userProductSubscriptionScanner interface {
	Scan(dest ...any) error
}

func scanUserProductSubscriptionFromRows(row userProductSubscriptionScanner) (*service.UserProductSubscription, error) {
	var sub service.UserProductSubscription
	var dailyWindow, weeklyWindow, monthlyWindow sql.NullTime
	var assignedBy sql.NullInt64
	var notes sql.NullString
	if err := row.Scan(
		&sub.ID,
		&sub.UserID,
		&sub.ProductID,
		&sub.StartsAt,
		&sub.ExpiresAt,
		&sub.Status,
		&dailyWindow,
		&weeklyWindow,
		&monthlyWindow,
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
	); err != nil {
		return nil, err
	}
	sub.DailyWindowStart = nullableTimePtr(dailyWindow)
	sub.WeeklyWindowStart = nullableTimePtr(weeklyWindow)
	sub.MonthlyWindowStart = nullableTimePtr(monthlyWindow)
	if assignedBy.Valid {
		v := assignedBy.Int64
		sub.AssignedBy = &v
	}
	sub.Notes = notes.String
	return &sub, nil
}

func normalizeSubscriptionProductStatus(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return service.SubscriptionProductStatusDraft, nil
	}
	switch status {
	case service.SubscriptionProductStatusDraft, service.SubscriptionProductStatusActive, service.SubscriptionProductStatusDisabled:
		return status, nil
	default:
		return "", infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_STATUS", "invalid subscription product status")
	}
}

func normalizeProductValidityDays(days, fallback int) int {
	if days <= 0 {
		days = fallback
	}
	if days <= 0 {
		days = 30
	}
	if days > service.MaxValidityDays {
		return service.MaxValidityDays
	}
	return days
}

func normalizeProductBindingInput(input service.SubscriptionProductBindingInput) (service.SubscriptionProductBindingInput, error) {
	if input.GroupID <= 0 {
		return input, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDINGS", "group_id is required")
	}
	if input.DebitMultiplier <= 0 {
		input.DebitMultiplier = 1
	}
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = service.SubscriptionProductBindingStatusActive
	}
	switch input.Status {
	case service.SubscriptionProductBindingStatusActive, service.SubscriptionProductBindingStatusInactive:
		return input, nil
	default:
		return input, infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT_BINDING_STATUS", "invalid subscription product binding status")
	}
}

func validateProductLimits(daily, weekly, monthly float64) error {
	if daily < 0 {
		return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "daily_limit_usd cannot be negative")
	}
	if weekly < 0 {
		return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "weekly_limit_usd cannot be negative")
	}
	if monthly < 0 {
		return infraerrors.BadRequest("INVALID_SUBSCRIPTION_PRODUCT", "monthly_limit_usd cannot be negative")
	}
	return nil
}

func normalizeProductFamily(input string) string {
	out := strings.TrimSpace(input)
	if out == "" {
		return "gpt"
	}
	return out
}

func productSubscriptionHasRemaining(sub *service.UserProductSubscription, product *service.SubscriptionProduct) error {
	if sub == nil || product == nil {
		return service.ErrSubscriptionNotFound
	}
	if product.HasDailyLimit() && sub.DailyRemainingTotal(product) <= 0 {
		return service.ErrDailyLimitExceeded
	}
	if product.HasWeeklyLimit() && sub.WeeklyUsageUSD >= product.WeeklyLimitUSD {
		return service.ErrWeeklyLimitExceeded
	}
	if product.HasMonthlyLimit() && sub.MonthlyUsageUSD >= product.MonthlyLimitUSD {
		return service.ErrMonthlyLimitExceeded
	}
	return nil
}

func appendSubscriptionNotes(existing, next string) string {
	existing = strings.TrimSpace(existing)
	next = strings.TrimSpace(next)
	if existing == "" {
		return next
	}
	if next == "" {
		return existing
	}
	return existing + "\n" + next
}

func nullableTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	return &v.Time
}

func nullableProductFloat64Ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}

type pqInt64Array []int64

func (a pqInt64Array) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	parts := make([]string, len(a))
	for i, v := range a {
		parts[i] = strconv.FormatInt(v, 10)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}
