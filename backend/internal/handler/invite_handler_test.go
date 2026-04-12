//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type inviteHandlerUserRepoStub struct {
	users        map[int64]*service.User
	inviteeCount int64
}

func (s *inviteHandlerUserRepoStub) GetByID(ctx context.Context, id int64) (*service.User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, service.ErrUserNotFound
	}
	return u, nil
}

func (s *inviteHandlerUserRepoStub) GetByInviteCode(ctx context.Context, code string) (*service.User, error) {
	for _, user := range s.users {
		if user.InviteCode == code {
			return user, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (s *inviteHandlerUserRepoStub) ExistsByInviteCode(ctx context.Context, code string) (bool, error) {
	_, err := s.GetByInviteCode(ctx, code)
	return err == nil, nil
}

func (s *inviteHandlerUserRepoStub) CountInviteesByInviter(ctx context.Context, inviterID int64) (int64, error) {
	return s.inviteeCount, nil
}

func (s *inviteHandlerUserRepoStub) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	return nil
}

type inviteHandlerRewardRepoStub struct {
	records   []service.InviteRewardRecord
	totalBase float64
}

type inviteHandlerSettingRepoStub struct{}

func (s *inviteHandlerRewardRepoStub) CreateBatch(ctx context.Context, records []service.InviteRewardRecord) error {
	return nil
}

func (s *inviteHandlerRewardRepoStub) ListByRewardTarget(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.InviteRewardRecord, *pagination.PaginationResult, error) {
	return s.records, &pagination.PaginationResult{Total: int64(len(s.records))}, nil
}

func (s *inviteHandlerRewardRepoStub) ListByAdminActionID(ctx context.Context, adminActionID int64) ([]service.InviteRewardRecord, error) {
	return nil, nil
}

func (s *inviteHandlerRewardRepoStub) SumBaseRewardsByTargetAndRole(ctx context.Context, userID int64, rewardRole string) (float64, error) {
	return s.totalBase, nil
}

func (s *inviteHandlerSettingRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get")
}

func (s *inviteHandlerSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if key == service.SettingKeyFrontendURL {
		return "https://portal.example.com", nil
	}
	return "", service.ErrSettingNotFound
}

func (s *inviteHandlerSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set")
}

func (s *inviteHandlerSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple")
}

func (s *inviteHandlerSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple")
}

func (s *inviteHandlerSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll")
}

func (s *inviteHandlerSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete")
}

func TestInviteHandler_GetSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/invite/summary", nil)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 8})

	settingService := service.NewSettingService(&inviteHandlerSettingRepoStub{}, &config.Config{})
	handler := NewInviteHandler(service.ProvideInviteService(
		&inviteHandlerUserRepoStub{
			users: map[int64]*service.User{
				8: {ID: 8, InviteCode: "HELLO123", Status: service.StatusActive},
			},
			inviteeCount: 4,
		},
		&inviteHandlerRewardRepoStub{totalBase: 9},
		settingService,
		nil,
	))

	handler.GetSummary(c)
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope struct {
		Code int `json:"code"`
		Data struct {
			InviteCode            string  `json:"invite_code"`
			InviteLink            string  `json:"invite_link"`
			InvitedUsersTotal     int64   `json:"invited_users_total"`
			InviteesRechargeTotal float64 `json:"invitees_recharge_total"`
			BaseRewardsTotal      float64 `json:"base_rewards_total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)
	require.Equal(t, "HELLO123", envelope.Data.InviteCode)
	require.Equal(t, "https://portal.example.com/register?invite=HELLO123", envelope.Data.InviteLink)
	require.EqualValues(t, 4, envelope.Data.InvitedUsersTotal)
	require.Equal(t, 300.0, envelope.Data.InviteesRechargeTotal)
	require.Equal(t, 9.0, envelope.Data.BaseRewardsTotal)
}

func TestInviteHandler_ListRewards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/invite/rewards?page=1&page_size=20", nil)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 8})

	settingService := service.NewSettingService(&inviteHandlerSettingRepoStub{}, &config.Config{})
	handler := NewInviteHandler(service.ProvideInviteService(
		&inviteHandlerUserRepoStub{
			users: map[int64]*service.User{
				8: {ID: 8, InviteCode: "HELLO123", Status: service.StatusActive},
			},
		},
		&inviteHandlerRewardRepoStub{
			records: []service.InviteRewardRecord{
				{
					RewardTargetUserID: 8,
					RewardRole:         service.InviteRewardRoleInvitee,
					RewardType:         service.InviteRewardTypeBase,
					RewardAmount:       5,
					CreatedAt:          time.Date(2026, 4, 11, 8, 0, 0, 0, time.UTC),
				},
			},
		},
		settingService,
		nil,
	))

	handler.ListRewards(c)
	require.Equal(t, http.StatusOK, rec.Code)

	var envelope struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				RewardRole   string    `json:"reward_role"`
				RewardType   string    `json:"reward_type"`
				RewardAmount float64   `json:"reward_amount"`
				CreatedAt    time.Time `json:"created_at"`
			} `json:"items"`
			Total int64 `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)
	require.EqualValues(t, 1, envelope.Data.Total)
	require.Len(t, envelope.Data.Items, 1)
	require.Equal(t, service.InviteRewardRoleInvitee, envelope.Data.Items[0].RewardRole)
	require.Equal(t, service.InviteRewardTypeBase, envelope.Data.Items[0].RewardType)
	require.Equal(t, 5.0, envelope.Data.Items[0].RewardAmount)
}
