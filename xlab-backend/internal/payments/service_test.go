package payments

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type fakePaymentMirrorRepo struct {
	list  *OrderList
	order OrderSnapshot
	state *SyncState

	listErr  error
	getErr   error
	stateErr error

	listCalls int
	getCalls  int
	userID    int64
	orderID   int64
	params    ListParams
	source    string
}

func (f *fakePaymentMirrorRepo) ListOrdersByUser(ctx context.Context, userID int64, params ListParams) (*OrderList, error) {
	f.listCalls++
	f.userID = userID
	f.params = params
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.list, nil
}

func (f *fakePaymentMirrorRepo) GetOrderByUser(ctx context.Context, orderID int64, userID int64) (OrderSnapshot, error) {
	f.getCalls++
	f.orderID = orderID
	f.userID = userID
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.order, nil
}

func (f *fakePaymentMirrorRepo) SyncState(ctx context.Context, sourceName string) (*SyncState, error) {
	f.source = sourceName
	if f.stateErr != nil {
		return nil, f.stateErr
	}
	if f.state == nil {
		return nil, sql.ErrNoRows
	}
	return f.state, nil
}

type fakePaymentCoreProxy struct {
	path  string
	calls int
}

func (f *fakePaymentCoreProxy) ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error) {
	f.calls++
	f.path = path
	return json.RawMessage(`{"source":"core"}`), nil
}

func TestServiceCoreModeListUsesCoreProxyWithDefaults(t *testing.T) {
	repo := &fakePaymentMirrorRepo{}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceCore, repo, coreProxy, 10*time.Minute, time.Now)

	out, err := svc.ListOrders(context.Background(), &core.User{ID: 7}, "tok", ListParams{})
	if err != nil {
		t.Fatalf("ListOrders error: %v", err)
	}

	if coreProxy.path != "/payment/orders/my?page=1&page_size=20" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
	if repo.listCalls != 0 {
		t.Fatalf("repo list calls = %d", repo.listCalls)
	}
	if _, ok := out.(json.RawMessage); !ok {
		t.Fatalf("output type = %T", out)
	}
}

func TestServiceCoreModeNilCoreReturnsError(t *testing.T) {
	svc := NewService(ReadSourceCore, &fakePaymentMirrorRepo{}, nil, 10*time.Minute, time.Now)

	_, err := svc.ListOrders(context.Background(), &core.User{ID: 7}, "tok", ListParams{})
	if err == nil {
		t.Fatal("ListOrders error = nil")
	}
	if !strings.Contains(err.Error(), "missing core proxy") {
		t.Fatalf("ListOrders error = %v", err)
	}
}

func TestServiceCoreModeFilteredListUsesNormalizedCappedPath(t *testing.T) {
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceCore, &fakePaymentMirrorRepo{}, coreProxy, 10*time.Minute, time.Now)

	_, err := svc.ListOrders(context.Background(), &core.User{ID: 7}, "tok", ListParams{
		Page:        0,
		PageSize:    500,
		Status:      "PENDING",
		OrderType:   "subscription",
		PaymentType: "stripe",
	})
	if err != nil {
		t.Fatalf("ListOrders error: %v", err)
	}

	want := "/payment/orders/my?page=1&page_size=100&status=PENDING&order_type=subscription&payment_type=stripe"
	if coreProxy.path != want {
		t.Fatalf("core path = %s, want %s", coreProxy.path, want)
	}
}

func TestServiceHybridFreshMirrorUsesRepositoryList(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakePaymentMirrorRepo{
		state: freshPaymentSyncState(now),
		list: &OrderList{
			Items:    []OrderSnapshot{{"out_trade_no": "otn_101", "status": "PAID"}},
			Total:    1,
			Page:     2,
			PageSize: 5,
			Pages:    1,
		},
	}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })

	out, err := svc.ListOrders(context.Background(), &core.User{ID: 7}, "tok", ListParams{Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("ListOrders error: %v", err)
	}

	list, ok := out.(*OrderList)
	if !ok || len(list.Items) != 1 || list.Items[0]["out_trade_no"] != "otn_101" {
		t.Fatalf("unexpected output: %#v", out)
	}
	if coreProxy.calls != 0 {
		t.Fatalf("core calls = %d", coreProxy.calls)
	}
	if repo.source != syncSourcePaymentOrders {
		t.Fatalf("sync source = %s", repo.source)
	}
	if repo.userID != 7 || repo.params.Page != 2 || repo.params.PageSize != 5 {
		t.Fatalf("repo call user=%d params=%+v", repo.userID, repo.params)
	}
}

func TestServiceHybridEmptyUnfilteredFirstPageFallsBackToCore(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakePaymentMirrorRepo{
		state: freshPaymentSyncState(now),
		list:  &OrderList{Items: []OrderSnapshot{}, Total: 0, Page: 1, PageSize: 20, Pages: 1},
	}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })

	_, err := svc.ListOrders(context.Background(), &core.User{ID: 7}, "tok", ListParams{})
	if err != nil {
		t.Fatalf("ListOrders error: %v", err)
	}

	if coreProxy.path != "/payment/orders/my?page=1&page_size=20" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}

func TestServiceHybridStaleMirrorFallsBackForDetail(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	stale := now.Add(-time.Hour)
	repo := &fakePaymentMirrorRepo{state: &SyncState{SourceName: syncSourcePaymentOrders, LastSuccessAt: &stale, RowCount: 1}}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })

	_, err := svc.GetOrder(context.Background(), &core.User{ID: 7}, "tok", 101)
	if err != nil {
		t.Fatalf("GetOrder error: %v", err)
	}

	if coreProxy.path != "/payment/orders/101" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
	if repo.getCalls != 0 {
		t.Fatalf("repo get calls = %d", repo.getCalls)
	}
}

func TestServiceStaleFallbackNilCoreReturnsError(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	stale := now.Add(-time.Hour)
	repo := &fakePaymentMirrorRepo{state: &SyncState{SourceName: syncSourcePaymentOrders, LastSuccessAt: &stale, RowCount: 1}}
	svc := NewService(ReadSourceHybrid, repo, nil, 10*time.Minute, func() time.Time { return now })

	_, err := svc.GetOrder(context.Background(), &core.User{ID: 7}, "tok", 101)
	if err == nil {
		t.Fatal("GetOrder error = nil")
	}
	if !strings.Contains(err.Error(), "missing core proxy") {
		t.Fatalf("GetOrder error = %v", err)
	}
	if repo.getCalls != 0 {
		t.Fatalf("repo get calls = %d", repo.getCalls)
	}
}

func TestServiceRepositoryErrorFallsBackForDetail(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakePaymentMirrorRepo{state: freshPaymentSyncState(now), getErr: errors.New("db down")}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })

	_, err := svc.GetOrder(context.Background(), &core.User{ID: 7}, "tok", 101)
	if err != nil {
		t.Fatalf("GetOrder should fallback instead of failing: %v", err)
	}

	if coreProxy.path != "/payment/orders/101" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}

func TestServiceFallbackLogsOnlySafeReason(t *testing.T) {
	var logs bytes.Buffer
	restore := capturePaymentServiceLogs(&logs)
	defer restore()

	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakePaymentMirrorRepo{
		state:  freshPaymentSyncState(now),
		getErr: errors.New("db down for user@example.com out_trade_no=otn_secret"),
	}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })

	_, err := svc.GetOrder(context.Background(), &core.User{ID: 7, Email: "user@example.com"}, "tok_secret", 101)
	if err != nil {
		t.Fatalf("GetOrder error: %v", err)
	}

	got := logs.String()
	if !strings.Contains(got, "repo_error") {
		t.Fatalf("logs = %s", got)
	}
	for _, forbidden := range []string{"tok_secret", "user@example.com", "otn_secret", "db down"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("logs contain sensitive value %q: %s", forbidden, got)
		}
	}
}

func TestServiceFreshMirrorDetailReturnsRepositoryOrder(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakePaymentMirrorRepo{
		state: freshPaymentSyncState(now),
		order: OrderSnapshot{"out_trade_no": "otn_101", "status": "PAID"},
	}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceXlab, repo, coreProxy, 10*time.Minute, func() time.Time { return now })

	out, err := svc.GetOrder(context.Background(), &core.User{ID: 7}, "tok", 101)
	if err != nil {
		t.Fatalf("GetOrder error: %v", err)
	}

	order, ok := out.(OrderSnapshot)
	if !ok || order["out_trade_no"] != "otn_101" {
		t.Fatalf("unexpected output: %#v", out)
	}
	if coreProxy.calls != 0 {
		t.Fatalf("core calls = %d", coreProxy.calls)
	}
	if repo.orderID != 101 || repo.userID != 7 {
		t.Fatalf("repo call order=%d user=%d", repo.orderID, repo.userID)
	}
}

func TestServiceHybridFilteredEmptyListDoesNotFallback(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakePaymentMirrorRepo{
		state: freshPaymentSyncState(now),
		list:  &OrderList{Items: []OrderSnapshot{}, Total: 0, Page: 1, PageSize: 20, Pages: 1},
	}
	coreProxy := &fakePaymentCoreProxy{}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })

	out, err := svc.ListOrders(context.Background(), &core.User{ID: 7}, "tok", ListParams{Status: "PAID"})
	if err != nil {
		t.Fatalf("ListOrders error: %v", err)
	}

	list, ok := out.(*OrderList)
	if !ok || len(list.Items) != 0 {
		t.Fatalf("unexpected output: %#v", out)
	}
	if coreProxy.calls != 0 {
		t.Fatalf("core calls = %d", coreProxy.calls)
	}
}

func freshPaymentSyncState(ts time.Time) *SyncState {
	return &SyncState{SourceName: syncSourcePaymentOrders, LastSuccessAt: &ts, RowCount: 1}
}

func capturePaymentServiceLogs(buf *bytes.Buffer) func() {
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return func() {
		slog.SetDefault(previous)
	}
}
