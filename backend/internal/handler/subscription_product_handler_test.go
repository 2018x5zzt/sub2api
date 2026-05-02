//go:build unit

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
)

type subscriptionProductHandlerServiceStub struct {
	products []service.ActiveSubscriptionProduct
}

func (s *subscriptionProductHandlerServiceStub) ListActiveUserProducts(context.Context, int64) ([]service.ActiveSubscriptionProduct, error) {
	out := make([]service.ActiveSubscriptionProduct, len(s.products))
	copy(out, s.products)
	return out, nil
}

func (s *subscriptionProductHandlerServiceStub) GetUserProductSummary(context.Context, int64) (*service.SubscriptionProductSummary, error) {
	return &service.SubscriptionProductSummary{ActiveCount: len(s.products), Products: s.products}, nil
}

func (s *subscriptionProductHandlerServiceStub) GetUserProductProgress(ctx context.Context, userID int64) (*service.SubscriptionProductSummary, error) {
	return s.GetUserProductSummary(ctx, userID)
}

func TestSubscriptionProductHandlerGetActiveReturnsProductWithGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expiresAt := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)
	handler := &SubscriptionProductHandler{subscriptionProductService: &subscriptionProductHandlerServiceStub{
		products: []service.ActiveSubscriptionProduct{
			{
				Product: service.SubscriptionProduct{
					ID:              88,
					Code:            "gpt_monthly",
					Name:            "GPT Monthly",
					Description:     "shared pool",
					Status:          service.SubscriptionProductStatusActive,
					DailyLimitUSD:   45,
					MonthlyLimitUSD: 1350,
				},
				Subscription: service.UserProductSubscription{
					ID:                         99,
					ExpiresAt:                  expiresAt,
					Status:                     service.SubscriptionStatusActive,
					DailyUsageUSD:              6,
					DailyCarryoverInUSD:        38,
					DailyCarryoverRemainingUSD: 32,
				},
				Groups: []service.SubscriptionProductGroupSummary{
					{GroupID: 21, GroupName: "plus-team", DebitMultiplier: 1, Status: service.SubscriptionProductBindingStatusActive},
					{GroupID: 36, GroupName: "pro", DebitMultiplier: 1.5, Status: service.SubscriptionProductBindingStatusActive},
				},
			},
		},
	}}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/subscription-products/active", nil)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 7})

	handler.GetActive(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ProductID                  int64   `json:"product_id"`
			SubscriptionID             int64   `json:"subscription_id"`
			Code                       string  `json:"code"`
			DailyCarryoverRemainingUSD float64 `json:"daily_carryover_remaining_usd"`
			Groups                     []struct {
				GroupID         int64   `json:"group_id"`
				GroupName       string  `json:"group_name"`
				DebitMultiplier float64 `json:"debit_multiplier"`
			} `json:"groups"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 || len(resp.Data) != 1 {
		t.Fatalf("response = %+v, want one success item", resp)
	}
	got := resp.Data[0]
	if got.ProductID != 88 || got.SubscriptionID != 99 || got.Code != "gpt_monthly" {
		t.Fatalf("product response = %+v, want product/subscription ids and code", got)
	}
	if got.DailyCarryoverRemainingUSD != 32 {
		t.Fatalf("daily_carryover_remaining_usd = %v, want 32", got.DailyCarryoverRemainingUSD)
	}
	if len(got.Groups) != 2 || got.Groups[1].GroupName != "pro" || got.Groups[1].DebitMultiplier != 1.5 {
		t.Fatalf("groups = %+v, want plus-team and pro with multipliers", got.Groups)
	}
}
