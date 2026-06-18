package payments

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSyncerSyncOnceUpsertsPaymentOrderAndUpdatesSyncState(t *testing.T) {
	coreDB, coreMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("core sqlmock.New error: %v", err)
	}
	defer coreDB.Close()
	xlabDB, xlabMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("xlab sqlmock.New error: %v", err)
	}
	defer xlabDB.Close()

	syncedAt := time.Date(2026, 6, 18, 5, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)
	expiresAt := time.Date(2026, 1, 4, 5, 6, 7, 0, time.UTC)
	paidAt := time.Date(2026, 1, 2, 4, 4, 5, 0, time.UTC)
	completedAt := time.Date(2026, 1, 2, 4, 5, 5, 0, time.UTC)
	refundRequestedAt := time.Date(2026, 1, 5, 6, 7, 8, 0, time.UTC)

	coreMock.ExpectQuery(regexp.QuoteMeta(corePaymentOrdersSQL)).WillReturnRows(
		corePaymentOrderRows().AddRow(
			int64(101),
			int64(7),
			12.34,
			12.70,
			0.03,
			"otn_101",
			"COMPLETED",
			"subscription",
			"stripe",
			int64(22),
			"pi_123",
			"hkd",
			createdAt,
			updatedAt,
			expiresAt,
			paidAt,
			completedAt,
			1.25,
			"customer request",
			refundRequestedAt,
			"admin",
			"duplicate purchase",
		),
	)

	xlabMock.ExpectBegin()
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertPaymentOrderSQL)).
		WithArgs(
			int64(101),
			int64(7),
			"otn_101",
			"COMPLETED",
			"subscription",
			"stripe",
			12.34,
			12.70,
			"HKD",
			"pi_123",
			createdAt,
			updatedAt,
			expiresAt,
			paidAt,
			completedAt,
			createdAt,
			updatedAt,
			snapshotArgument{expected: map[string]any{
				"id":                    float64(101),
				"user_id":               float64(7),
				"amount":                12.34,
				"pay_amount":            12.70,
				"fee_rate":              0.03,
				"currency":              "HKD",
				"payment_type":          "stripe",
				"out_trade_no":          "otn_101",
				"status":                "COMPLETED",
				"order_type":            "subscription",
				"created_at":            createdAt.Format(time.RFC3339Nano),
				"expires_at":            expiresAt.Format(time.RFC3339Nano),
				"paid_at":               paidAt.Format(time.RFC3339Nano),
				"completed_at":          completedAt.Format(time.RFC3339Nano),
				"refund_amount":         1.25,
				"refund_reason":         "customer request",
				"refund_requested_at":   refundRequestedAt.Format(time.RFC3339Nano),
				"refund_requested_by":   "admin",
				"refund_request_reason": "duplicate purchase",
				"plan_id":               float64(22),
				"provider_instance_id":  "pi_123",
			}},
			syncedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertPaymentSyncStateSQL)).
		WithArgs(syncedAt, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectCommit()

	syncer := NewSyncer(coreDB, xlabDB, func() time.Time { return syncedAt })
	if err := syncer.SyncOnce(context.Background()); err != nil {
		t.Fatalf("SyncOnce error: %v", err)
	}
	if err := coreMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet core expectations: %v", err)
	}
	if err := xlabMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet xlab expectations: %v", err)
	}
}

func TestSyncerSyncOnceRollsBackWhenUpsertFails(t *testing.T) {
	coreDB, coreMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("core sqlmock.New error: %v", err)
	}
	defer coreDB.Close()
	xlabDB, xlabMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("xlab sqlmock.New error: %v", err)
	}
	defer xlabDB.Close()

	now := time.Date(2026, 6, 18, 5, 0, 0, 0, time.UTC)
	coreMock.ExpectQuery(regexp.QuoteMeta(corePaymentOrdersSQL)).WillReturnRows(
		corePaymentOrderRows().AddRow(
			int64(101),
			int64(7),
			12.34,
			12.70,
			0.03,
			"otn_101",
			"PENDING",
			"balance",
			"stripe",
			nil,
			nil,
			"",
			now,
			now,
			now.Add(24*time.Hour),
			nil,
			nil,
			0.0,
			nil,
			nil,
			nil,
			nil,
		),
	)
	xlabMock.ExpectBegin()
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertPaymentOrderSQL)).
		WillReturnError(errors.New("upsert failed"))
	xlabMock.ExpectRollback()

	syncer := NewSyncer(coreDB, xlabDB, func() time.Time { return now })
	if err := syncer.SyncOnce(context.Background()); err == nil {
		t.Fatal("SyncOnce error = nil, want upsert failure")
	}
	if err := coreMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet core expectations: %v", err)
	}
	if err := xlabMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet xlab expectations: %v", err)
	}
}

func TestNormalizePaymentCurrency(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "uppercases three letter code", in: " hkd ", want: "HKD"},
		{name: "defaults empty", in: "", want: "CNY"},
		{name: "defaults non ascii", in: "人民币", want: "CNY"},
		{name: "defaults too short", in: "US", want: "CNY"},
		{name: "defaults too long", in: "USDT", want: "CNY"},
		{name: "defaults alphanumeric", in: "U5D", want: "CNY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePaymentCurrency(tt.in); got != tt.want {
				t.Fatalf("normalizePaymentCurrency(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func corePaymentOrderRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"amount",
		"pay_amount",
		"fee_rate",
		"out_trade_no",
		"status",
		"order_type",
		"payment_type",
		"plan_id",
		"provider_instance_id",
		"provider_currency",
		"created_at",
		"updated_at",
		"expires_at",
		"paid_at",
		"completed_at",
		"refund_amount",
		"refund_reason",
		"refund_requested_at",
		"refund_requested_by",
		"refund_request_reason",
	})
}

type snapshotArgument struct {
	expected map[string]any
}

func (s snapshotArgument) Match(value driver.Value) bool {
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return false
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		return false
	}
	return reflect.DeepEqual(got, s.expected)
}
