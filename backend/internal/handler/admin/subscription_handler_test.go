package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type subscriptionAdminRepoStub struct {
	listCalls    int
	lastStatus   string
	lastGroupID  *int64
	resetCalls   []int64
	listResponse []service.UserSubscription
}

func (r *subscriptionAdminRepoStub) Create(context.Context, *service.UserSubscription) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) GetByID(context.Context, int64) (*service.UserSubscription, error) {
	return nil, errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) GetByUserIDAndGroupID(context.Context, int64, int64) (*service.UserSubscription, error) {
	return nil, errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) GetActiveByUserIDAndGroupID(context.Context, int64, int64) (*service.UserSubscription, error) {
	return nil, errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) Update(context.Context, *service.UserSubscription) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) Delete(context.Context, int64) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) ListByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return nil, errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) ListActiveByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return nil, errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]service.UserSubscription, *pagination.PaginationResult, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) List(_ context.Context, params pagination.PaginationParams, _ *int64, groupID *int64, status, _, _, _ string) ([]service.UserSubscription, *pagination.PaginationResult, error) {
	r.listCalls++
	r.lastStatus = status
	r.lastGroupID = groupID
	if params.Page > 1 {
		return nil, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
	}
	return append([]service.UserSubscription(nil), r.listResponse...), &pagination.PaginationResult{
		Total:    int64(len(r.listResponse)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    1,
	}, nil
}

func (r *subscriptionAdminRepoStub) ExistsByUserIDAndGroupID(context.Context, int64, int64) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) ExtendExpiry(context.Context, int64, time.Time) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) UpdateStatus(context.Context, int64, string) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) UpdateNotes(context.Context, int64, string) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) ActivateWindows(context.Context, int64, time.Time) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) AdvanceDailyWindow(context.Context, int64, time.Time, float64, float64) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) ResetDailyUsage(_ context.Context, id int64, _ time.Time) error {
	r.resetCalls = append(r.resetCalls, id)
	return nil
}

func (r *subscriptionAdminRepoStub) ResetWeeklyUsage(context.Context, int64, time.Time) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) ResetMonthlyUsage(context.Context, int64, time.Time) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) IncrementUsage(context.Context, int64, float64) error {
	return errors.New("not implemented")
}

func (r *subscriptionAdminRepoStub) BatchUpdateExpiredStatus(context.Context) (int64, error) {
	return 0, errors.New("not implemented")
}

func TestSubscriptionHandlerResetDailyQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &subscriptionAdminRepoStub{
		listResponse: []service.UserSubscription{
			{ID: 11, UserID: 101, GroupID: 88},
			{ID: 12, UserID: 102, GroupID: 88},
		},
	}
	svc := service.NewSubscriptionService(nil, repo, nil, nil, nil)
	handler := NewSubscriptionHandler(svc)

	router := gin.New()
	router.POST("/api/v1/admin/subscriptions/reset-daily", handler.ResetDailyQuota)

	body, err := json.Marshal(map[string]any{"group_id": 88})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/subscriptions/reset-daily", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, service.SubscriptionStatusActive, repo.lastStatus)
	require.NotNil(t, repo.lastGroupID)
	require.Equal(t, int64(88), *repo.lastGroupID)
	require.Equal(t, []int64{11, 12}, repo.resetCalls)
	require.Contains(t, rec.Body.String(), "\"reset_count\":2")
}
