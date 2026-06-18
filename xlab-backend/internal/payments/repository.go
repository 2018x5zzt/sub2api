package payments

import (
	"context"
	"database/sql"
	"encoding/json"
)

const countOrdersByUserSQL = `
SELECT COUNT(*)
FROM xlab_payment_orders
WHERE core_user_id = $1
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR order_type = $3)
  AND ($4 = '' OR payment_type = $4)`

const listOrdersByUserSQL = `
SELECT response_snapshot
FROM xlab_payment_orders
WHERE core_user_id = $1
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR order_type = $3)
  AND ($4 = '' OR payment_type = $4)
ORDER BY created_at DESC NULLS LAST, core_order_id DESC
LIMIT $5 OFFSET $6`

const orderByUserSQL = `
SELECT response_snapshot
FROM xlab_payment_orders
WHERE core_order_id = $1
  AND core_user_id = $2`

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

func (r *Repository) ListOrdersByUser(ctx context.Context, userID int64, params ListParams) (*OrderList, error) {
	page, pageSize := normalizePagination(params)

	var total int
	if err := r.db.QueryRowContext(
		ctx,
		countOrdersByUserSQL,
		userID,
		params.Status,
		params.OrderType,
		params.PaymentType,
	).Scan(&total); err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(
		ctx,
		listOrdersByUserSQL,
		userID,
		params.Status,
		params.OrderType,
		params.PaymentType,
		pageSize,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OrderSnapshot, 0)
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		item, err := decodeSnapshot(raw)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &OrderList{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pageCount(total, pageSize),
	}, nil
}

func (r *Repository) GetOrderByUser(ctx context.Context, orderID int64, userID int64) (OrderSnapshot, error) {
	row := r.db.QueryRowContext(ctx, orderByUserSQL, orderID, userID)
	var raw json.RawMessage
	if err := row.Scan(&raw); err != nil {
		return nil, err
	}
	return decodeSnapshot(raw)
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

func normalizePagination(params ListParams) (int, int) {
	page := params.Page
	if page <= 0 {
		page = 1
	}

	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

func decodeSnapshot(raw json.RawMessage) (OrderSnapshot, error) {
	var snapshot OrderSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func pageCount(total int, pageSize int) int {
	pages := (total + pageSize - 1) / pageSize
	if pages < 1 {
		return 1
	}
	return pages
}
