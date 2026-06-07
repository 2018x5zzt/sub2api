package subscriptions

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

const activeProductsByUserSQL = `
SELECT
    ups.core_subscription_id,
    ups.core_product_id,
    sp.code,
    sp.name,
    sp.description,
    ups.status,
    ups.expires_at,
    ups.daily_usage_usd,
    ups.weekly_usage_usd,
    ups.monthly_usage_usd,
    COALESCE(ups.daily_limit_usd, sp.daily_limit_usd, 0),
    COALESCE(ups.weekly_limit_usd, sp.weekly_limit_usd, 0),
    COALESCE(ups.monthly_limit_usd, sp.monthly_limit_usd, 0),
    ups.daily_carryover_in_usd,
    ups.daily_carryover_remaining_usd
FROM xlab_user_product_subscriptions ups
JOIN xlab_subscription_products sp
  ON sp.core_product_id = ups.core_product_id
WHERE ups.core_user_id = $1
  AND ups.status = 'active'
  AND ups.expires_at > NOW()
  AND sp.status = 'active'
ORDER BY sp.name ASC, ups.expires_at DESC, ups.core_subscription_id DESC`

const groupsByProductSQL = `
SELECT
    core_product_id,
    core_group_id,
    group_name,
    COALESCE(group_platform, ''),
    balance_fallback_group_id,
    balance_fallback_group_name,
    debit_multiplier,
    status,
    sort_order
FROM xlab_subscription_product_groups
WHERE core_product_id = ANY($1)
  AND status = 'active'
ORDER BY core_product_id ASC, sort_order ASC, core_group_id ASC`

const syncStateSQL = `
SELECT source_name, last_success_at, row_count
FROM xlab_sync_state
WHERE source_name = $1`

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListActiveProductsByUser(ctx context.Context, userID int64) ([]ActiveProduct, error) {
	rows, err := r.db.QueryContext(ctx, activeProductsByUserSQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ActiveProduct, 0)
	productIDs := make([]int64, 0)
	for rows.Next() {
		var item ActiveProduct
		var expiresAt sql.NullTime
		if err := rows.Scan(
			&item.SubscriptionID,
			&item.ProductID,
			&item.Code,
			&item.Name,
			&item.Description,
			&item.Status,
			&expiresAt,
			&item.DailyUsageUSD,
			&item.WeeklyUsageUSD,
			&item.MonthlyUsageUSD,
			&item.DailyLimitUSD,
			&item.WeeklyLimitUSD,
			&item.MonthlyLimitUSD,
			&item.DailyCarryoverInUSD,
			&item.DailyCarryoverRemainingUSD,
		); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}
		item.Groups = []Group{}
		items = append(items, item)
		productIDs = append(productIDs, item.ProductID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []ActiveProduct{}, nil
	}

	groupsByProduct, err := r.groupsByProductID(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Groups = groupsByProduct[items[i].ProductID]
		if items[i].Groups == nil {
			items[i].Groups = []Group{}
		}
	}
	return items, nil
}

func (r *Repository) groupsByProductID(ctx context.Context, productIDs []int64) (map[int64][]Group, error) {
	rows, err := r.db.QueryContext(ctx, groupsByProductSQL, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64][]Group)
	for rows.Next() {
		var productID int64
		var group Group
		var fallbackGroupID sql.NullInt64
		var fallbackGroupName sql.NullString
		if err := rows.Scan(
			&productID,
			&group.GroupID,
			&group.GroupName,
			&group.GroupPlatform,
			&fallbackGroupID,
			&fallbackGroupName,
			&group.DebitMultiplier,
			&group.Status,
			&group.SortOrder,
		); err != nil {
			return nil, err
		}
		if fallbackGroupID.Valid {
			v := fallbackGroupID.Int64
			group.BalanceFallbackGroupID = &v
		}
		if fallbackGroupName.Valid {
			v := fallbackGroupName.String
			group.BalanceFallbackGroupName = &v
		}
		out[productID] = append(out[productID], group)
	}
	return out, rows.Err()
}

func (r *Repository) SyncState(ctx context.Context, sourceName string) (*SyncState, error) {
	row := r.db.QueryRowContext(ctx, syncStateSQL, sourceName)
	var state SyncState
	var lastSuccess sql.NullTime
	if err := row.Scan(&state.SourceName, &lastSuccess, &state.RowCount); err != nil {
		return nil, err
	}
	if lastSuccess.Valid {
		state.LastSuccessAt = &lastSuccess.Time
	}
	return &state, nil
}
