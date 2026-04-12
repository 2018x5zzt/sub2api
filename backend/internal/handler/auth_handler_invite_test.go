//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type inviteValidationSettingRepoStub struct {
	invitationCodeEnabled bool
}

type inviteValidationUserRepoStub struct {
	usersByInviteCode map[string]*service.User
}

func (s *inviteValidationUserRepoStub) GetByID(ctx context.Context, id int64) (*service.User, error) {
	return nil, service.ErrUserNotFound
}

func (s *inviteValidationUserRepoStub) GetByInviteCode(ctx context.Context, code string) (*service.User, error) {
	u, ok := s.usersByInviteCode[code]
	if !ok {
		return nil, service.ErrUserNotFound
	}
	return u, nil
}

func (s *inviteValidationUserRepoStub) ExistsByInviteCode(ctx context.Context, code string) (bool, error) {
	_, ok := s.usersByInviteCode[code]
	return ok, nil
}

func (s *inviteValidationUserRepoStub) CountInviteesByInviter(ctx context.Context, inviterID int64) (int64, error) {
	return 0, nil
}

func (s *inviteValidationUserRepoStub) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	return nil
}

type inviteValidationRewardRepoStub struct{}

type inviteValidationRedeemRepoStub struct {
	code *service.RedeemCode
	err  error
}

func (s *inviteValidationRedeemRepoStub) GetByCode(ctx context.Context, code string) (*service.RedeemCode, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.code == nil {
		return nil, service.ErrRedeemCodeNotFound
	}
	return s.code, nil
}

func (s *inviteValidationRewardRepoStub) CreateBatch(ctx context.Context, records []service.InviteRewardRecord) error {
	return nil
}

func (s *inviteValidationRewardRepoStub) ListByRewardTarget(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.InviteRewardRecord, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{Total: 0}, nil
}

func (s *inviteValidationRewardRepoStub) ListByAdminActionID(ctx context.Context, adminActionID int64) ([]service.InviteRewardRecord, error) {
	return nil, nil
}

func (s *inviteValidationRewardRepoStub) SumBaseRewardsByTargetAndRole(ctx context.Context, userID int64, rewardRole string) (float64, error) {
	return 0, nil
}

func (s *inviteValidationSettingRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get")
}

func (s *inviteValidationSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if key == service.SettingKeyInvitationCodeEnabled {
		if s.invitationCodeEnabled {
			return "true", nil
		}
		return "false", nil
	}
	return "", service.ErrSettingNotFound
}

func (s *inviteValidationSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set")
}

func (s *inviteValidationSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple")
}

func (s *inviteValidationSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple")
}

func (s *inviteValidationSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll")
}

func (s *inviteValidationSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete")
}

func TestValidateInvitationCode_UsesPermanentInviteCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-invitation-code", strings.NewReader(`{"code":"INVITER07"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	inviteSvc := service.NewInviteService(&inviteValidationUserRepoStub{
		usersByInviteCode: map[string]*service.User{
			"INVITER07": {ID: 7, InviteCode: "INVITER07", Status: service.StatusActive},
		},
	}, &inviteValidationRewardRepoStub{})
	authService := service.NewAuthService(
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		inviteSvc,
	)
	handler := &AuthHandler{
		settingSvc:  service.NewSettingService(&inviteValidationSettingRepoStub{invitationCodeEnabled: true}, &config.Config{}),
		authService: authService,
	}

	handler.ValidateInvitationCode(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"valid":true`)
}

func TestValidateInvitationCode_IgnoresLegacyToggleWhenPermanentInviteCodeExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-invitation-code", strings.NewReader(`{"code":"INVITER07"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	inviteSvc := service.NewInviteService(&inviteValidationUserRepoStub{
		usersByInviteCode: map[string]*service.User{
			"INVITER07": {ID: 7, InviteCode: "INVITER07", Status: service.StatusActive},
		},
	}, &inviteValidationRewardRepoStub{})
	authService := service.NewAuthService(
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		inviteSvc,
	)
	handler := &AuthHandler{
		settingSvc:  service.NewSettingService(&inviteValidationSettingRepoStub{invitationCodeEnabled: false}, &config.Config{}),
		authService: authService,
	}

	handler.ValidateInvitationCode(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"valid":true`)
	require.NotContains(t, rec.Body.String(), `INVITATION_CODE_DISABLED`)
}

func TestValidateInvitationCode_ReturnsRemovedErrorForLegacyInvitationRedeemCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-invitation-code", strings.NewReader(`{"code":"legacy-123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	inviteSvc := service.NewInviteService(&inviteValidationUserRepoStub{
		usersByInviteCode: map[string]*service.User{},
	}, &inviteValidationRewardRepoStub{})
	authService := service.NewAuthService(
		nil,
		nil,
		&inviteValidationRedeemRepoStub{
			code: &service.RedeemCode{
				Code: "LEGACY-123",
				Type: service.RedeemTypeInvitation,
			},
		},
		nil,
		&config.Config{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		inviteSvc,
	)
	handler := &AuthHandler{
		authService: authService,
	}

	handler.ValidateInvitationCode(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"valid":false`)
	require.Contains(t, rec.Body.String(), `"error_code":"INVITATION_CODE_REMOVED"`)
}
