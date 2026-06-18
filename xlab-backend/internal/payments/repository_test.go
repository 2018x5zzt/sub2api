package payments

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryListOrdersByUserAppliesFiltersAndPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	snapshot := []byte(`{"out_trade_no":"otn_101","status":"PENDING"}`)
	mock.ExpectQuery(regexp.QuoteMeta(countOrdersByUserSQL)).
		WithArgs(int64(7), "PENDING", "subscription", "stripe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(listOrdersByUserSQL)).
		WithArgs(int64(7), "PENDING", "subscription", "stripe", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"response_snapshot"}).AddRow(snapshot))

	repo := NewRepository(db)
	result, err := repo.ListOrdersByUser(context.Background(), 7, ListParams{
		Status:      "PENDING",
		OrderType:   "subscription",
		PaymentType: "stripe",
	})
	if err != nil {
		t.Fatalf("ListOrdersByUser error: %v", err)
	}
	if result.Total != 1 || result.Page != 1 || result.PageSize != 20 || result.Pages != 1 {
		t.Fatalf("result metadata = %+v", result)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(items) = %d", len(result.Items))
	}
	if result.Items[0]["out_trade_no"] != "otn_101" {
		t.Fatalf("out_trade_no = %v", result.Items[0]["out_trade_no"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetOrderByUserDecodesSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	snapshot := []byte(`{"out_trade_no":"otn_102","status":"COMPLETED"}`)
	mock.ExpectQuery(regexp.QuoteMeta(orderByUserSQL)).
		WithArgs(int64(101), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"response_snapshot"}).AddRow(snapshot))

	repo := NewRepository(db)
	result, err := repo.GetOrderByUser(context.Background(), 101, 7)
	if err != nil {
		t.Fatalf("GetOrderByUser error: %v", err)
	}
	if result["status"] != "COMPLETED" {
		t.Fatalf("status = %v", result["status"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositorySyncStateReadsPaymentOrdersState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(syncStateSQL)).
		WithArgs(syncSourcePaymentOrders).
		WillReturnRows(sqlmock.NewRows([]string{"source_name", "last_success_at", "row_count"}).AddRow(syncSourcePaymentOrders, now, 3))

	repo := NewRepository(db)
	state, err := repo.SyncState(context.Background(), syncSourcePaymentOrders)
	if err != nil {
		t.Fatalf("SyncState error: %v", err)
	}
	if state.LastSuccessAt == nil || state.RowCount != 3 {
		t.Fatalf("state = %+v", state)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
