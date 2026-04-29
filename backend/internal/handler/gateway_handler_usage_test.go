package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUsageUnrestrictedReportsLegacySubscriptionWithoutCarryover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dailyLimit := 45.0
	group := &service.Group{
		ID:               1,
		Name:             "backend",
		SubscriptionType: service.SubscriptionTypeSubscription,
		DailyLimitUSD:    &dailyLimit,
	}
	subscription := &service.UserSubscription{
		ID:                         10,
		DailyUsageUSD:              33.54,
		DailyCarryoverInUSD:        40.35,
		DailyCarryoverRemainingUSD: 6.81,
		ExpiresAt:                  time.Now().Add(24 * time.Hour),
		Group:                      group,
	}

	resp := runUsageUnrestricted(t, &service.APIKey{Group: group}, nil, subscription)

	require.InDelta(t, 11.46, resp["remaining"], 1e-6)
	payload := requireMap(t, resp["subscription"])
	require.InDelta(t, 33.54, payload["daily_usage_usd"], 1e-6)
	require.InDelta(t, 45.0, payload["daily_limit_usd"], 1e-6)
	require.InDelta(t, 0.0, payload["daily_carryover_in_usd"], 1e-6)
	require.InDelta(t, 45.0, payload["daily_effective_limit_usd"], 1e-6)
	require.InDelta(t, 11.46, payload["daily_remaining_total_usd"], 1e-6)
	require.InDelta(t, 0.0, payload["daily_remaining_carryover_usd"], 1e-6)
}

func TestUsageUnrestrictedReportsProductSubscriptionCarryover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	group := &service.Group{
		ID:               1,
		Name:             "backend",
		SubscriptionType: service.SubscriptionTypeSubscription,
	}
	productCtx := &service.ProductSettlementContext{
		Binding: &service.SubscriptionProductBinding{
			ProductID:     20,
			ProductCode:   "backend_pool",
			ProductName:   "Backend Pool",
			ProductStatus: service.SubscriptionProductStatusActive,
			BindingStatus: service.SubscriptionProductBindingStatusActive,
			DailyLimitUSD: 45,
			GroupID:       group.ID,
		},
		Subscription: &service.UserProductSubscription{
			ID:                         30,
			DailyUsageUSD:              33.54,
			DailyCarryoverInUSD:        40.35,
			DailyCarryoverRemainingUSD: 6.81,
			ExpiresAt:                  time.Now().Add(24 * time.Hour),
		},
	}

	resp := runUsageUnrestricted(t, &service.APIKey{Group: group}, productCtx, nil)

	require.Equal(t, "backend", resp["planName"])
	require.InDelta(t, 51.81, resp["remaining"], 1e-6)
	payload := requireMap(t, resp["subscription"])
	require.InDelta(t, 20, payload["product_id"], 1e-6)
	require.Equal(t, "backend_pool", payload["product_code"])
	require.Equal(t, "Backend Pool", payload["product_name"])
	require.InDelta(t, 33.54, payload["daily_usage_usd"], 1e-6)
	require.InDelta(t, 45.0, payload["daily_limit_usd"], 1e-6)
	require.InDelta(t, 40.35, payload["daily_carryover_in_usd"], 1e-6)
	require.InDelta(t, 85.35, payload["daily_effective_limit_usd"], 1e-6)
	require.InDelta(t, 51.81, payload["daily_remaining_total_usd"], 1e-6)
	require.InDelta(t, 6.81, payload["daily_remaining_carryover_usd"], 1e-6)
}

func runUsageUnrestricted(t *testing.T, apiKey *service.APIKey, productCtx *service.ProductSettlementContext, subscription *service.UserSubscription) map[string]any {
	t.Helper()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/v1/usage", nil)
	if productCtx != nil {
		req = req.WithContext(service.ContextWithProductSettlement(req.Context(), productCtx))
	}
	c.Request = req
	if subscription != nil {
		c.Set(string(middleware.ContextKeySubscription), subscription)
	}

	handler := &GatewayHandler{}
	handler.usageUnrestricted(c, req.Context(), apiKey, middleware.AuthSubject{UserID: 1}, nil, nil)

	require.Equal(t, http.StatusOK, recorder.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	return resp
}

func requireMap(t *testing.T, value any) map[string]any {
	t.Helper()
	payload, ok := value.(map[string]any)
	require.True(t, ok, "expected map payload, got %T", value)
	return payload
}
