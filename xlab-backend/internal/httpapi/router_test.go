package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/2018x5zzt/xlab-backend/internal/core"
	"github.com/2018x5zzt/xlab-backend/internal/payments"
)

type fakeCoreClient struct {
	userCalls int
	getPath   string
	token     string
}

type fakeReadService struct {
	activeCalled bool
}

type fakePaymentReadService struct {
	listCalled   bool
	detailCalled bool
	user         *core.User
	token        string
	params       payments.ListParams
	orderID      int64
}

func (f *fakeCoreClient) CurrentUser(ctx context.Context, token string) (*core.User, error) {
	f.userCalls++
	f.token = token
	return &core.User{ID: 7, Email: "u@example.com", Role: "user"}, nil
}

func (f *fakeCoreClient) ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error) {
	f.token = token
	f.getPath = path
	return json.RawMessage(`[{"subscription_id":99,"name":"Pro"}]`), nil
}

func (f *fakeReadService) Active(ctx context.Context, user *core.User, token string) (any, error) {
	f.activeCalled = true
	return []map[string]any{{"subscription_id": 123, "name": "Mirror"}}, nil
}

func (f *fakeReadService) Summary(ctx context.Context, user *core.User, token string) (any, error) {
	return map[string]any{"active_count": 1, "total_monthly_usage_usd": 0, "total_monthly_limit_usd": 0, "products": []any{}}, nil
}

func (f *fakeReadService) Progress(ctx context.Context, user *core.User, token string) (any, error) {
	return map[string]any{"active_count": 1, "total_monthly_usage_usd": 0, "total_monthly_limit_usd": 0, "products": []any{}}, nil
}

func (f *fakePaymentReadService) ListOrders(ctx context.Context, user *core.User, token string, params payments.ListParams) (any, error) {
	f.listCalled = true
	f.user = user
	f.token = token
	f.params = params
	return map[string]any{"items": []any{map[string]any{"id": 456}}, "total": 1}, nil
}

func (f *fakePaymentReadService) GetOrder(ctx context.Context, user *core.User, token string, orderID int64) (any, error) {
	f.detailCalled = true
	f.user = user
	f.token = token
	f.orderID = orderID
	return map[string]any{"id": orderID, "status": "PAID"}, nil
}

func TestHealth(t *testing.T) {
	r := NewRouter(&fakeCoreClient{}, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsRequireBearer(t *testing.T) {
	r := NewRouter(&fakeCoreClient{}, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsProxyActive(t *testing.T) {
	fake := &fakeCoreClient{}
	r := NewRouter(fake, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if fake.getPath != "/subscription-products/active" {
		t.Fatalf("proxied path = %s", fake.getPath)
	}
	if fake.token != "abc" {
		t.Fatalf("token = %s", fake.token)
	}
	if !strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestAuthStoresCurrentUserInContext(t *testing.T) {
	fake := &fakeCoreClient{}
	api := &API{core: fake}
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()

	api.auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userFromContext(r.Context())
		if user == nil || user.ID != 7 || user.Email != "u@example.com" {
			t.Fatalf("userFromContext = %+v", user)
		}
		if tokenFromContext(r.Context()) != "abc" {
			t.Fatalf("tokenFromContext = %q", tokenFromContext(r.Context()))
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSubscriptionProductsUseReadServiceWhenProvided(t *testing.T) {
	fakeCore := &fakeCoreClient{}
	fakeReads := &fakeReadService{}
	r := NewRouter(fakeCore, fakeReads, nil)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !fakeReads.activeCalled {
		t.Fatal("read service was not called")
	}
	if !strings.Contains(rec.Body.String(), `"subscription_id":123`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestPaymentOrdersRequireBearer(t *testing.T) {
	r := NewRouter(&fakeCoreClient{}, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/payment/orders/my", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestPaymentOrdersRejectPostWithoutCallingServices(t *testing.T) {
	fakeCore := &fakeCoreClient{}
	fakeReads := &fakePaymentReadService{}
	r := NewRouter(fakeCore, nil, fakeReads)
	req := httptest.NewRequest(http.MethodPost, "/xapi/v1/payment/orders/my", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed && rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if fakeCore.userCalls != 0 {
		t.Fatalf("CurrentUser calls = %d", fakeCore.userCalls)
	}
	if fakeCore.getPath != "" {
		t.Fatalf("proxied path = %s", fakeCore.getPath)
	}
	if fakeReads.listCalled || fakeReads.detailCalled {
		t.Fatalf("payment read service called: list=%v detail=%v", fakeReads.listCalled, fakeReads.detailCalled)
	}
}

func TestPaymentOrdersProxyListWhenReadServiceMissing(t *testing.T) {
	fake := &fakeCoreClient{}
	r := NewRouter(fake, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/payment/orders/my?page=2&page_size=150&status=PAID&order_type=subscription&payment_type=stripe", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	want := "/payment/orders/my?page=2&page_size=100&status=PAID&order_type=subscription&payment_type=stripe"
	if fake.getPath != want {
		t.Fatalf("proxied path = %s, want %s", fake.getPath, want)
	}
}

func TestPaymentOrdersUseReadServiceForList(t *testing.T) {
	fakeReads := &fakePaymentReadService{}
	r := NewRouter(&fakeCoreClient{}, nil, fakeReads)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/payment/orders/my?page=0&page_size=0&status=PENDING&order_type=credit&payment_type=alipay", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !fakeReads.listCalled {
		t.Fatal("payment read service list was not called")
	}
	if fakeReads.user == nil || fakeReads.user.ID != 7 {
		t.Fatalf("user = %+v", fakeReads.user)
	}
	if fakeReads.token != "abc" {
		t.Fatalf("token = %s", fakeReads.token)
	}
	want := payments.ListParams{Page: 1, PageSize: 20, Status: "PENDING", OrderType: "credit", PaymentType: "alipay"}
	if fakeReads.params != want {
		t.Fatalf("params = %+v, want %+v", fakeReads.params, want)
	}
	if !strings.Contains(rec.Body.String(), `"id":456`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestPaymentOrdersUseReadServiceForDetail(t *testing.T) {
	fakeReads := &fakePaymentReadService{}
	r := NewRouter(&fakeCoreClient{}, nil, fakeReads)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/payment/orders/789", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !fakeReads.detailCalled {
		t.Fatal("payment read service detail was not called")
	}
	if fakeReads.orderID != 789 {
		t.Fatalf("orderID = %d", fakeReads.orderID)
	}
	if fakeReads.user == nil || fakeReads.user.ID != 7 {
		t.Fatalf("user = %+v", fakeReads.user)
	}
	if fakeReads.token != "abc" {
		t.Fatalf("token = %s", fakeReads.token)
	}
	if !strings.Contains(rec.Body.String(), `"id":789`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestPaymentOrdersRejectNestedPathWithoutCallingServices(t *testing.T) {
	fakeCore := &fakeCoreClient{}
	fakeReads := &fakePaymentReadService{}
	r := NewRouter(fakeCore, nil, fakeReads)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/payment/orders/123/cancel", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if fakeCore.userCalls != 0 {
		t.Fatalf("CurrentUser calls = %d", fakeCore.userCalls)
	}
	if fakeCore.getPath != "" {
		t.Fatalf("proxied path = %s", fakeCore.getPath)
	}
	if fakeReads.listCalled || fakeReads.detailCalled {
		t.Fatalf("payment read service called: list=%v detail=%v", fakeReads.listCalled, fakeReads.detailCalled)
	}
}

func TestPaymentOrdersBadID(t *testing.T) {
	r := NewRouter(&fakeCoreClient{}, nil, &fakePaymentReadService{})
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/payment/orders/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"BAD_ORDER_ID"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
