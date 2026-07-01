package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSubmitUsageRecordTaskCopiesRequestContext(t *testing.T) {
	parent := context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-request-123")
	parent = context.WithValue(parent, ctxkey.RequestID, "request-456")

	var gotClientRequestID string
	var gotRequestID string
	h := &GatewayHandler{}
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		gotClientRequestID, _ = ctx.Value(ctxkey.ClientRequestID).(string)
		gotRequestID, _ = ctx.Value(ctxkey.RequestID).(string)
	})

	require.Equal(t, "client-request-123", gotClientRequestID)
	require.Equal(t, "request-456", gotRequestID)
}

func TestOpenAISubmitUsageRecordTaskCopiesRequestContext(t *testing.T) {
	parent := context.WithValue(context.Background(), ctxkey.ClientRequestID, "openai-client-request-123")
	parent = context.WithValue(parent, ctxkey.RequestID, "openai-request-456")

	var gotClientRequestID string
	var gotRequestID string
	h := &OpenAIGatewayHandler{}
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		gotClientRequestID, _ = ctx.Value(ctxkey.ClientRequestID).(string)
		gotRequestID, _ = ctx.Value(ctxkey.RequestID).(string)
	})

	require.Equal(t, "openai-client-request-123", gotClientRequestID)
	require.Equal(t, "openai-request-456", gotRequestID)
}

func TestOpenAISubmitUsageRecordTaskCopiesProductSettlementContext(t *testing.T) {
	settlement := &service.ProductSettlementContext{
		Binding: &service.SubscriptionProductBinding{
			ProductID:       99,
			GroupID:         42,
			DebitMultiplier: 1,
		},
		Subscription: &service.UserProductSubscription{
			ID:        1001,
			UserID:    2002,
			ProductID: 99,
			Status:    service.SubscriptionStatusActive,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	parent := service.ContextWithProductSettlement(context.Background(), settlement)

	var got *service.ProductSettlementContext
	h := &OpenAIGatewayHandler{}
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		got, _ = service.ProductSettlementFromContext(ctx)
	})

	require.Same(t, settlement, got)
}

func TestUsageUnrestrictedReportsProductSubscriptionSettlement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(42)
	apiKey := &service.APIKey{
		ID:      7,
		GroupID: &groupID,
		Group: &service.Group{
			ID:               groupID,
			Name:             "Claude Pro",
			SubscriptionType: service.SubscriptionTypeSubscription,
		},
	}
	expiresAt := time.Date(2026, 7, 31, 23, 59, 59, 0, time.UTC)
	settlement := &service.ProductSettlementContext{
		Binding: &service.SubscriptionProductBinding{
			ProductID:       99,
			ProductCode:     "xlab-pro",
			ProductName:     "xlab Pro",
			DailyLimitUSD:   10,
			WeeklyLimitUSD:  50,
			MonthlyLimitUSD: 120,
			GroupID:         groupID,
		},
		Subscription: &service.UserProductSubscription{
			ID:                  1001,
			UserID:              2002,
			ProductID:           99,
			Status:              service.SubscriptionStatusActive,
			ExpiresAt:           expiresAt,
			DailyUsageUSD:       4,
			WeeklyUsageUSD:      25,
			MonthlyUsageUSD:     80,
			DailyCarryoverInUSD: 2,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(string(middleware2.ContextKeyProductSettlement), settlement)

	(&GatewayHandler{}).usageUnrestricted(c, context.Background(), apiKey, middleware2.AuthSubject{UserID: 2002}, nil, nil, nil)

	require.Equal(t, 200, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "xlab Pro", body["planName"])
	require.InDelta(t, 12.0-4.0, body["remaining"], 1e-9)

	sub, ok := body["subscription"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "product", sub["source"])
	require.InDelta(t, float64(1001), sub["subscription_id"], 1e-9)
	require.InDelta(t, float64(99), sub["product_id"], 1e-9)
	require.Equal(t, "xlab-pro", sub["product_code"])
	require.Equal(t, "xlab Pro", sub["product_name"])
	require.InDelta(t, 4.0, sub["daily_usage_usd"], 1e-9)
	require.InDelta(t, 25.0, sub["weekly_usage_usd"], 1e-9)
	require.InDelta(t, 80.0, sub["monthly_usage_usd"], 1e-9)
	require.InDelta(t, 10.0, sub["daily_limit_usd"], 1e-9)
	require.InDelta(t, 12.0, sub["daily_effective_limit_usd"], 1e-9)
	require.InDelta(t, 50.0, sub["weekly_limit_usd"], 1e-9)
	require.InDelta(t, 120.0, sub["monthly_limit_usd"], 1e-9)
}
