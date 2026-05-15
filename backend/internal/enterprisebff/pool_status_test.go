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
	defaultBudget := 12.0
	dailyLimit := 10.0
	weeklyLimit := 70.0
	monthlyLimit := 300.0
	image1K := 0.02
	image2K := 0.04
	image4K := 0.08
	fallbackID := int64(9)
	fallbackOnInvalidID := int64(10)

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
		{
			ID:                              2,
			Name:                            "enterprise-private",
			Description:                     "enterprise only",
			Platform:                        "openai",
			RateMultiplier:                  1.8,
			PricingMode:                     "dynamic",
			DefaultBudgetMultiplier:         &defaultBudget,
			IsExclusive:                     true,
			Status:                          "active",
			SubscriptionType:                "standard",
			DailyLimitUSD:                   &dailyLimit,
			WeeklyLimitUSD:                  &weeklyLimit,
			MonthlyLimitUSD:                 &monthlyLimit,
			ImagePrice1K:                    &image1K,
			ImagePrice2K:                    &image2K,
			ImagePrice4K:                    &image4K,
			ClaudeCodeOnly:                  true,
			AllowMessagesDispatch:           true,
			FallbackGroupID:                 &fallbackID,
			FallbackGroupIDOnInvalidRequest: &fallbackOnInvalidID,
			RequireOAuthOnly:                true,
			RequirePrivacySet:               true,
			RPMLimit:                        15,
		},
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
			ID                              int64    `json:"id"`
			Name                            string   `json:"name"`
			Description                     string   `json:"description"`
			Platform                        string   `json:"platform"`
			RateMultiplier                  float64  `json:"rate_multiplier"`
			PricingMode                     string   `json:"pricing_mode"`
			DefaultBudgetMultiplier         *float64 `json:"default_budget_multiplier"`
			IsExclusive                     bool     `json:"is_exclusive"`
			Status                          string   `json:"status"`
			SubscriptionType                string   `json:"subscription_type"`
			DailyLimitUSD                   *float64 `json:"daily_limit_usd"`
			WeeklyLimitUSD                  *float64 `json:"weekly_limit_usd"`
			MonthlyLimitUSD                 *float64 `json:"monthly_limit_usd"`
			ImagePrice1K                    *float64 `json:"image_price_1k"`
			ImagePrice2K                    *float64 `json:"image_price_2k"`
			ImagePrice4K                    *float64 `json:"image_price_4k"`
			ClaudeCodeOnly                  bool     `json:"claude_code_only"`
			AllowMessagesDispatch           bool     `json:"allow_messages_dispatch"`
			FallbackGroupID                 *int64   `json:"fallback_group_id"`
			FallbackGroupIDOnInvalidRequest *int64   `json:"fallback_group_id_on_invalid_request"`
			RequireOAuthOnly                bool     `json:"require_oauth_only"`
			RequirePrivacySet               bool     `json:"require_privacy_set"`
			RPMLimit                        int      `json:"rpm_limit"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.Len(t, payload.Data, 1)
	require.Equal(t, int64(2), payload.Data[0].ID)
	require.Equal(t, "enterprise-private", payload.Data[0].Name)
	require.Equal(t, "openai", payload.Data[0].Platform)
	require.Equal(t, "enterprise only", payload.Data[0].Description)
	require.Equal(t, 1.8, payload.Data[0].RateMultiplier)
	require.Equal(t, "dynamic", payload.Data[0].PricingMode)
	require.NotNil(t, payload.Data[0].DefaultBudgetMultiplier)
	require.Equal(t, 12.0, *payload.Data[0].DefaultBudgetMultiplier)
	require.True(t, payload.Data[0].IsExclusive)
	require.Equal(t, "active", payload.Data[0].Status)
	require.Equal(t, "standard", payload.Data[0].SubscriptionType)
	require.NotNil(t, payload.Data[0].DailyLimitUSD)
	require.NotNil(t, payload.Data[0].WeeklyLimitUSD)
	require.NotNil(t, payload.Data[0].MonthlyLimitUSD)
	require.Equal(t, 10.0, *payload.Data[0].DailyLimitUSD)
	require.Equal(t, 70.0, *payload.Data[0].WeeklyLimitUSD)
	require.Equal(t, 300.0, *payload.Data[0].MonthlyLimitUSD)
	require.NotNil(t, payload.Data[0].ImagePrice1K)
	require.NotNil(t, payload.Data[0].ImagePrice2K)
	require.NotNil(t, payload.Data[0].ImagePrice4K)
	require.Equal(t, 0.02, *payload.Data[0].ImagePrice1K)
	require.Equal(t, 0.04, *payload.Data[0].ImagePrice2K)
	require.Equal(t, 0.08, *payload.Data[0].ImagePrice4K)
	require.True(t, payload.Data[0].ClaudeCodeOnly)
	require.True(t, payload.Data[0].AllowMessagesDispatch)
	require.NotNil(t, payload.Data[0].FallbackGroupID)
	require.NotNil(t, payload.Data[0].FallbackGroupIDOnInvalidRequest)
	require.EqualValues(t, 9, *payload.Data[0].FallbackGroupID)
	require.EqualValues(t, 10, *payload.Data[0].FallbackGroupIDOnInvalidRequest)
	require.True(t, payload.Data[0].RequireOAuthOnly)
	require.True(t, payload.Data[0].RequirePrivacySet)
	require.Equal(t, 15, payload.Data[0].RPMLimit)
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
