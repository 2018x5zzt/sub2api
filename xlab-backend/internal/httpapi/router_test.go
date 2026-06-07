package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type fakeCoreClient struct {
	userCalls int
	getPath   string
	token     string
}

type fakeReadService struct {
	activeCalled bool
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

func TestHealth(t *testing.T) {
	r := NewRouter(&fakeCoreClient{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsRequireBearer(t *testing.T) {
	r := NewRouter(&fakeCoreClient{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsProxyActive(t *testing.T) {
	fake := &fakeCoreClient{}
	r := NewRouter(fake, nil)
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
	r := NewRouter(fakeCore, fakeReads)
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
