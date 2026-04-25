package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type subscriptionProductAdminServiceFake struct {
	createInput       *service.CreateSubscriptionProductInput
	syncProductID     int64
	syncBindingInputs []service.SubscriptionProductBindingInput
}

func (f *subscriptionProductAdminServiceFake) ListProducts(context.Context) ([]service.SubscriptionProduct, error) {
	return nil, nil
}

func (f *subscriptionProductAdminServiceFake) CreateProduct(_ context.Context, input *service.CreateSubscriptionProductInput) (*service.SubscriptionProduct, error) {
	f.createInput = input
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	return &service.SubscriptionProduct{
		ID:                  101,
		Code:                input.Code,
		Name:                input.Name,
		Description:         input.Description,
		Status:              input.Status,
		DefaultValidityDays: input.DefaultValidityDays,
		DailyLimitUSD:       input.DailyLimitUSD,
		WeeklyLimitUSD:      input.WeeklyLimitUSD,
		MonthlyLimitUSD:     input.MonthlyLimitUSD,
		SortOrder:           input.SortOrder,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

func (f *subscriptionProductAdminServiceFake) UpdateProduct(context.Context, int64, *service.UpdateSubscriptionProductInput) (*service.SubscriptionProduct, error) {
	return nil, nil
}

func (f *subscriptionProductAdminServiceFake) SyncProductBindings(_ context.Context, productID int64, inputs []service.SubscriptionProductBindingInput) ([]service.SubscriptionProductBindingDetail, error) {
	f.syncProductID = productID
	f.syncBindingInputs = inputs
	return []service.SubscriptionProductBindingDetail{
		{GroupID: 201, GroupName: "gpt-4", DebitMultiplier: 1.5, Status: service.SubscriptionProductBindingStatusActive, SortOrder: 2},
		{GroupID: 202, GroupName: "gpt-4o", DebitMultiplier: 1, Status: service.SubscriptionProductBindingStatusInactive, SortOrder: 3},
	}, nil
}

func (f *subscriptionProductAdminServiceFake) ListProductSubscriptions(context.Context, int64) ([]service.UserProductSubscription, error) {
	return nil, nil
}

type subscriptionProductAdminEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func TestAdminSubscriptionProductHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &subscriptionProductAdminServiceFake{}
	handler := NewSubscriptionProductHandler(fake)

	rec := performAdminSubscriptionProductJSONRequest(handler.Create, http.MethodPost, "/admin/subscription-products", map[string]any{
		"code":                  "gpt_team",
		"name":                  "GPT Team",
		"description":           "shared GPT access",
		"status":                "active",
		"default_validity_days": 30,
		"daily_limit_usd":       10,
		"weekly_limit_usd":      50,
		"monthly_limit_usd":     100,
		"sort_order":            7,
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, fake.createInput)
	require.Equal(t, "gpt_team", fake.createInput.Code)
	require.Equal(t, "GPT Team", fake.createInput.Name)
	require.Equal(t, 30, fake.createInput.DefaultValidityDays)
	require.Equal(t, 100.0, fake.createInput.MonthlyLimitUSD)

	var envelope subscriptionProductAdminEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Equal(t, float64(101), payload["id"])
	require.Equal(t, "gpt_team", payload["code"])
	require.Equal(t, float64(100), payload["monthly_limit_usd"])
}

func TestAdminSubscriptionProductHandler_UpdateBindings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &subscriptionProductAdminServiceFake{}
	handler := NewSubscriptionProductHandler(fake)

	rec := performAdminSubscriptionProductJSONRequest(handler.SyncBindings, http.MethodPut, "/admin/subscription-products/101/bindings", map[string]any{
		"bindings": []map[string]any{
			{"group_id": 201, "debit_multiplier": 1.5, "status": "active", "sort_order": 2},
			{"group_id": 202, "debit_multiplier": 1.0, "status": "inactive", "sort_order": 3},
		},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(101), fake.syncProductID)
	require.Len(t, fake.syncBindingInputs, 2)
	require.Equal(t, int64(201), fake.syncBindingInputs[0].GroupID)
	require.Equal(t, 1.5, fake.syncBindingInputs[0].DebitMultiplier)
	require.Equal(t, service.SubscriptionProductBindingStatusInactive, fake.syncBindingInputs[1].Status)

	var envelope subscriptionProductAdminEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var payload []map[string]any
	require.NoError(t, json.Unmarshal(envelope.Data, &payload))
	require.Len(t, payload, 2)
	require.Equal(t, float64(201), payload[0]["group_id"])
	require.Equal(t, "gpt-4", payload[0]["group_name"])
	require.Equal(t, float64(1.5), payload[0]["debit_multiplier"])
}

func performAdminSubscriptionProductJSONRequest(handler gin.HandlerFunc, method, path string, body any) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, "/admin/subscription-products", handler)
	router.Handle(method, "/admin/subscription-products/:id/bindings", handler)

	raw, _ := json.Marshal(body)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	return rec
}
