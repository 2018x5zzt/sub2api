package payments

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

const corePaymentOrdersSQL = `
SELECT id, user_id, amount, pay_amount, fee_rate,
       out_trade_no, status, order_type, payment_type, plan_id, provider_instance_id,
       COALESCE(provider_snapshot->>'currency', '') AS provider_currency,
       created_at, updated_at, expires_at, paid_at, completed_at,
       refund_amount, refund_reason, refund_requested_at, refund_requested_by, refund_request_reason
FROM payment_orders
WHERE updated_at >= NOW() - interval '180 days'
   OR status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED', 'REFUNDED')
ORDER BY updated_at ASC, id ASC`

const upsertPaymentOrderSQL = `
INSERT INTO xlab_payment_orders (
    core_order_id, core_user_id, out_trade_no, status, order_type, payment_type,
    amount, pay_amount, currency, provider_instance_id, created_at, updated_at,
    expires_at, paid_at, completed_at, source_created_at, source_updated_at,
    response_snapshot, synced_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
ON CONFLICT (core_order_id) DO UPDATE SET
    core_user_id = EXCLUDED.core_user_id,
    out_trade_no = EXCLUDED.out_trade_no,
    status = EXCLUDED.status,
    order_type = EXCLUDED.order_type,
    payment_type = EXCLUDED.payment_type,
    amount = EXCLUDED.amount,
    pay_amount = EXCLUDED.pay_amount,
    currency = EXCLUDED.currency,
    provider_instance_id = EXCLUDED.provider_instance_id,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    expires_at = EXCLUDED.expires_at,
    paid_at = EXCLUDED.paid_at,
    completed_at = EXCLUDED.completed_at,
    source_created_at = EXCLUDED.source_created_at,
    source_updated_at = EXCLUDED.source_updated_at,
    response_snapshot = EXCLUDED.response_snapshot,
    synced_at = EXCLUDED.synced_at`

const upsertPaymentSyncStateSQL = `
INSERT INTO xlab_sync_state (source_name, last_success_at, row_count, created_at, updated_at)
VALUES ('payment_orders', $1, $2, $1, $1)
ON CONFLICT (source_name) DO UPDATE SET
    last_success_at = EXCLUDED.last_success_at,
    row_count = EXCLUDED.row_count,
    last_error = NULL,
    updated_at = EXCLUDED.updated_at`

const defaultPaymentCurrency = "CNY"

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
	rows, err := s.coreDB.QueryContext(ctx, corePaymentOrdersSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

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
	for rows.Next() {
		order, err := scanPaymentOrder(rows)
		if err != nil {
			return err
		}
		currency := normalizePaymentCurrency(order.providerCurrency)
		snapshot, err := buildPaymentOrderSnapshot(order, currency)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(
			ctx,
			upsertPaymentOrderSQL,
			order.id,
			order.userID,
			order.outTradeNo,
			order.status,
			order.orderType,
			order.paymentType,
			order.amount,
			order.payAmount,
			currency,
			nullableStringValue(order.providerInstanceID),
			order.createdAt,
			order.updatedAt,
			order.expiresAt,
			nullableTimeValue(order.paidAt),
			nullableTimeValue(order.completedAt),
			order.createdAt,
			order.updatedAt,
			snapshot,
			syncedAt,
		); err != nil {
			return err
		}
		rowCount++
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, upsertPaymentSyncStateSQL, syncedAt, rowCount); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

type paymentOrderRow struct {
	id                  int64
	userID              int64
	amount              float64
	payAmount           float64
	feeRate             float64
	outTradeNo          string
	status              string
	orderType           string
	paymentType         string
	planID              sql.NullInt64
	providerInstanceID  sql.NullString
	providerCurrency    string
	createdAt           time.Time
	updatedAt           time.Time
	expiresAt           time.Time
	paidAt              sql.NullTime
	completedAt         sql.NullTime
	refundAmount        float64
	refundReason        sql.NullString
	refundRequestedAt   sql.NullTime
	refundRequestedBy   sql.NullString
	refundRequestReason sql.NullString
}

func scanPaymentOrder(rows *sql.Rows) (paymentOrderRow, error) {
	var order paymentOrderRow
	err := rows.Scan(
		&order.id,
		&order.userID,
		&order.amount,
		&order.payAmount,
		&order.feeRate,
		&order.outTradeNo,
		&order.status,
		&order.orderType,
		&order.paymentType,
		&order.planID,
		&order.providerInstanceID,
		&order.providerCurrency,
		&order.createdAt,
		&order.updatedAt,
		&order.expiresAt,
		&order.paidAt,
		&order.completedAt,
		&order.refundAmount,
		&order.refundReason,
		&order.refundRequestedAt,
		&order.refundRequestedBy,
		&order.refundRequestReason,
	)
	return order, err
}

type paymentOrderSnapshot struct {
	ID                  int64      `json:"id"`
	UserID              int64      `json:"user_id"`
	Amount              float64    `json:"amount"`
	PayAmount           float64    `json:"pay_amount"`
	FeeRate             float64    `json:"fee_rate"`
	Currency            string     `json:"currency"`
	PaymentType         string     `json:"payment_type"`
	OutTradeNo          string     `json:"out_trade_no"`
	Status              string     `json:"status"`
	OrderType           string     `json:"order_type"`
	CreatedAt           time.Time  `json:"created_at"`
	ExpiresAt           time.Time  `json:"expires_at"`
	PaidAt              *time.Time `json:"paid_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	RefundAmount        float64    `json:"refund_amount"`
	RefundReason        *string    `json:"refund_reason,omitempty"`
	RefundRequestedAt   *time.Time `json:"refund_requested_at,omitempty"`
	RefundRequestedBy   *string    `json:"refund_requested_by,omitempty"`
	RefundRequestReason *string    `json:"refund_request_reason,omitempty"`
	PlanID              *int64     `json:"plan_id,omitempty"`
	ProviderInstanceID  *string    `json:"provider_instance_id,omitempty"`
}

func buildPaymentOrderSnapshot(order paymentOrderRow, currency string) ([]byte, error) {
	snapshot := paymentOrderSnapshot{
		ID:                  order.id,
		UserID:              order.userID,
		Amount:              order.amount,
		PayAmount:           order.payAmount,
		FeeRate:             order.feeRate,
		Currency:            currency,
		PaymentType:         order.paymentType,
		OutTradeNo:          order.outTradeNo,
		Status:              order.status,
		OrderType:           order.orderType,
		CreatedAt:           order.createdAt,
		ExpiresAt:           order.expiresAt,
		PaidAt:              nullableTimePointer(order.paidAt),
		CompletedAt:         nullableTimePointer(order.completedAt),
		RefundAmount:        order.refundAmount,
		RefundReason:        nullableStringPointer(order.refundReason),
		RefundRequestedAt:   nullableTimePointer(order.refundRequestedAt),
		RefundRequestedBy:   nullableStringPointer(order.refundRequestedBy),
		RefundRequestReason: nullableStringPointer(order.refundRequestReason),
		PlanID:              nullableInt64Pointer(order.planID),
		ProviderInstanceID:  nullableStringPointer(order.providerInstanceID),
	}
	return json.Marshal(snapshot)
}

func normalizePaymentCurrency(currency string) string {
	normalized := strings.ToUpper(strings.TrimSpace(currency))
	if len(normalized) != 3 {
		return defaultPaymentCurrency
	}
	for i := 0; i < len(normalized); i++ {
		if normalized[i] < 'A' || normalized[i] > 'Z' {
			return defaultPaymentCurrency
		}
	}
	return normalized
}

func nullableStringValue(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
}

func nullableTimeValue(value sql.NullTime) any {
	if !value.Valid {
		return nil
	}
	return value.Time
}

func nullableStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}
