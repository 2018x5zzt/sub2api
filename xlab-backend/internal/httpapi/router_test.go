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

func TestHealth(t *testing.T) {
	r := NewRouter(&fakeCoreClient{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsRequireBearer(t *testing.T) {
	r := NewRouter(&fakeCoreClient{})
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsProxyActive(t *testing.T) {
	fake := &fakeCoreClient{}
	r := NewRouter(fake)
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
