package enterprisebff

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEnterpriseVisibleGroupsOmitsPublicOnlyGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":42,"email":"owner@example.com","username":"owner","role":"user","balance":12.5,"concurrency":3,"status":"active"}}`))
		case "/api/v1/groups/available":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":[{"id":1,"name":"public-default","platform":"anthropic"},{"id":2,"name":"enterprise-private","platform":"openai"}]}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[42] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 42}
	store.visibleGroups[42] = []visibleGroupSeed{
		{ID: 2, Name: "enterprise-private", Platform: "openai"},
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodGet, "/groups/available", nil)
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload struct {
		Code int `json:"code"`
		Data []struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			Platform string `json:"platform"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.Len(t, payload.Data, 1)
	require.Equal(t, int64(2), payload.Data[0].ID)
	require.Equal(t, "enterprise-private", payload.Data[0].Name)
	require.Equal(t, "openai", payload.Data[0].Platform)
}

func TestEnterprisePoolStatusFiltersToExplicitGroupsOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":42,"email":"owner@example.com","username":"owner","role":"user","balance":12.5,"concurrency":3,"status":"active"}}`))
		case "/api/v1/groups/pool-status":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"checked_at":"2026-04-10T11:22:33Z","groups":[{"group_id":1,"group_name":"public-default","platform":"anthropic","total_accounts":10,"active_account_count":10,"rate_limited_account_count":0,"available_account_count":10,"availability_ratio":1,"status":"healthy"},{"group_id":2,"group_name":"enterprise-private","platform":"openai","total_accounts":4,"active_account_count":2,"rate_limited_account_count":1,"available_account_count":1,"availability_ratio":0.25,"status":"degraded"}]}}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[42] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 42}
	store.visibleGroups[42] = []visibleGroupSeed{
		{ID: 2, Name: "enterprise-private", Platform: "openai"},
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodGet, "/groups/pool-status", nil)
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)

	groupsRaw, ok := payload.Data["groups"].([]any)
	require.True(t, ok)
	require.Len(t, groupsRaw, 1)

	group, ok := groupsRaw[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(2), group["group_id"])
	require.Equal(t, "enterprise-private", group["group_name"])
	require.Equal(t, float64(25), group["health_percent"])
	require.Equal(t, "degraded", group["health_state"])
	require.Equal(t, "2026-04-10T11:22:33Z", group["updated_at"])
	require.NotContains(t, group, "total_accounts")
	require.NotContains(t, group, "active_account_count")
	require.NotContains(t, group, "rate_limited_account_count")
	require.NotContains(t, group, "available_account_count")
}

func TestEnterprisePoolStatusReturnsEmptyGroupsWhenNoAssignmentsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":42,"email":"owner@example.com","username":"owner","role":"user","balance":12.5,"concurrency":3,"status":"active"}}`))
		case "/api/v1/groups/pool-status":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"checked_at":"2026-04-10T11:22:33Z","groups":[{"group_id":1,"group_name":"public-default","platform":"anthropic","total_accounts":10,"active_account_count":10,"rate_limited_account_count":0,"available_account_count":10,"availability_ratio":1,"status":"healthy"}]}}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[42] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 42}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodGet, "/groups/pool-status", nil)
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.Equal(t, float64(0), payload.Data["visible_group_count"])

	groupsRaw, ok := payload.Data["groups"].([]any)
	require.True(t, ok)
	require.Len(t, groupsRaw, 0)
}
