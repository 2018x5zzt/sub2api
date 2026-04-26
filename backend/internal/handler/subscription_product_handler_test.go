package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type subscriptionProductServiceFake struct {
	lastUserID int64
	active     []service.ActiveSubscriptionProduct
	summary    *service.SubscriptionProductSummary
}

func (f *subscriptionProductServiceFake) ListActiveUserProducts(_ context.Context, userID int64) ([]service.ActiveSubscriptionProduct, error) {
	f.lastUserID = userID
	return f.active, nil
}

func (f *subscriptionProductServiceFake) GetUserProductSummary(_ context.Context, userID int64) (*service.SubscriptionProductSummary, error) {
	f.lastUserID = userID
	return f.summary, nil
}

func (f *subscriptionProductServiceFake) GetUserProductProgress(_ context.Context, userID int64) (*service.SubscriptionProductSummary, error) {
	f.lastUserID = userID
	return f.summary, nil
}

type subscriptionProductEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func TestSubscriptionProductHandler_GetActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expiresAt := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	fake := &subscriptionProductServiceFake{
		active: []service.ActiveSubscriptionProduct{
			{
				Product: service.SubscriptionProduct{
					ID:              101,
					Code:            "gpt_team",
					Name:            "GPT Team",
					DailyLimitUSD:   10,
					MonthlyLimitUSD: 100,
				},
				Subscription: service.UserProductSubscription{
					ID:                         501,
					ExpiresAt:                  expiresAt,
					DailyUsageUSD:              4,
					MonthlyUsageUSD:            17.5,
					DailyCarryoverInUSD:        2,
					DailyCarryoverRemainingUSD: 1.25,
				},
				Groups: []service.SubscriptionProductGroupSummary{
					{GroupID: 201, GroupName: "gpt-4", DebitMultiplier: 1.5},
				},
			},
		},
	}
	handler := NewSubscriptionProductHandler(fake)

	rec := performSubscriptionProductRequest(handler.GetActive, 42)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(42), fake.lastUserID)

	var envelope subscriptionProductEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload []map[string]any
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Len(t, payload, 1)
	require.Equal(t, float64(101), payload[0]["product_id"])
	require.Equal(t, "gpt_team", payload[0]["code"])
	require.Equal(t, "GPT Team", payload[0]["name"])
	require.Equal(t, float64(17.5), payload[0]["monthly_usage_usd"])
	require.Equal(t, float64(100), payload[0]["monthly_limit_usd"])
	require.Equal(t, float64(2), payload[0]["daily_carryover_in_usd"])
	require.Equal(t, float64(1.25), payload[0]["daily_carryover_remaining_usd"])
	require.Equal(t, float64(12), payload[0]["daily_effective_limit_usd"])
	require.Equal(t, float64(8), payload[0]["daily_remaining_total_usd"])
	require.Equal(t, float64(1.25), payload[0]["daily_remaining_carryover_usd"])

	groups, ok := payload[0]["groups"].([]any)
	require.True(t, ok)
	require.Len(t, groups, 1)
	group := groups[0].(map[string]any)
	require.Equal(t, float64(201), group["group_id"])
	require.Equal(t, "gpt-4", group["group_name"])
	require.Equal(t, float64(1.5), group["debit_multiplier"])
}

func TestSubscriptionProductHandler_GetSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &subscriptionProductServiceFake{
		summary: &service.SubscriptionProductSummary{
			ActiveCount:          1,
			TotalMonthlyUsageUSD: 17.5,
			TotalMonthlyLimitUSD: 100,
			Products: []service.ActiveSubscriptionProduct{
				{
					Product: service.SubscriptionProduct{
						ID:              101,
						Code:            "gpt_team",
						Name:            "GPT Team",
						MonthlyLimitUSD: 100,
					},
					Subscription: service.UserProductSubscription{
						ID:              501,
						MonthlyUsageUSD: 17.5,
					},
				},
			},
		},
	}
	handler := NewSubscriptionProductHandler(fake)

	rec := performSubscriptionProductRequest(handler.GetSummary, 42)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(42), fake.lastUserID)

	var envelope subscriptionProductEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Equal(t, float64(1), payload["active_count"])
	require.Equal(t, float64(17.5), payload["total_monthly_usage_usd"])
	require.Equal(t, float64(100), payload["total_monthly_limit_usd"])

	products, ok := payload["products"].([]any)
	require.True(t, ok)
	require.Len(t, products, 1)
	product := products[0].(map[string]any)
	require.Equal(t, float64(101), product["product_id"])
	require.Equal(t, "gpt_team", product["code"])
}

func performSubscriptionProductRequest(handler gin.HandlerFunc, userID int64) *httptest.ResponseRecorder {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID})
		c.Next()
	})
	router.GET("/subscription-products", handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/subscription-products", nil)
	router.ServeHTTP(rec, req)
	return rec
}
