//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type inviteValidationAffiliateRepoStub struct {
	summariesByCode map[string]*service.AffiliateSummary
}

func newInviteValidationAffiliateRepoStub(summaries ...*service.AffiliateSummary) *inviteValidationAffiliateRepoStub {
	repo := &inviteValidationAffiliateRepoStub{
		summariesByCode: map[string]*service.AffiliateSummary{},
	}
	for _, summary := range summaries {
		cp := *summary
		repo.summariesByCode[strings.ToUpper(cp.AffCode)] = &cp
	}
	return repo
}

func (s *inviteValidationAffiliateRepoStub) EnsureUserAffiliate(context.Context, int64) (*service.AffiliateSummary, error) {
	panic("unexpected EnsureUserAffiliate call")
}

func (s *inviteValidationAffiliateRepoStub) GetAffiliateByCode(_ context.Context, code string) (*service.AffiliateSummary, error) {
	summary, ok := s.summariesByCode[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return nil, service.ErrAffiliateProfileNotFound
	}
	return summary, nil
}

func (s *inviteValidationAffiliateRepoStub) BindInviter(context.Context, int64, int64) (bool, error) {
	panic("unexpected BindInviter call")
}

func (s *inviteValidationAffiliateRepoStub) AccrueQuota(context.Context, int64, int64, float64, int) (bool, error) {
	panic("unexpected AccrueQuota call")
}

func (s *inviteValidationAffiliateRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	panic("unexpected GetAccruedRebateFromInvitee call")
}

func (s *inviteValidationAffiliateRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	panic("unexpected ThawFrozenQuota call")
}

func (s *inviteValidationAffiliateRepoStub) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	panic("unexpected TransferQuotaToBalance call")
}

func (s *inviteValidationAffiliateRepoStub) ListInvitees(context.Context, int64, int) ([]service.AffiliateInvitee, error) {
	panic("unexpected ListInvitees call")
}

func (s *inviteValidationAffiliateRepoStub) UpdateUserAffCode(context.Context, int64, string) error {
	panic("unexpected UpdateUserAffCode call")
}

func (s *inviteValidationAffiliateRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	panic("unexpected ResetUserAffCode call")
}

func (s *inviteValidationAffiliateRepoStub) SetUserRebateRate(context.Context, int64, *float64) error {
	panic("unexpected SetUserRebateRate call")
}

func (s *inviteValidationAffiliateRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	panic("unexpected BatchSetUserRebateRate call")
}

func (s *inviteValidationAffiliateRepoStub) ListUsersWithCustomSettings(context.Context, service.AffiliateAdminFilter) ([]service.AffiliateAdminEntry, int64, error) {
	panic("unexpected ListUsersWithCustomSettings call")
}

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

func TestValidateInvitationCode_UsesPermanentInviteCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-invitation-code", strings.NewReader(`{"code":"INVITER07"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	affiliateService := service.NewAffiliateService(newInviteValidationAffiliateRepoStub(&service.AffiliateSummary{
		UserID:  7,
		AffCode: "INVITER07",
	}), nil, nil, nil)
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
		affiliateService,
	)
	handler := &AuthHandler{
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

	affiliateService := service.NewAffiliateService(newInviteValidationAffiliateRepoStub(&service.AffiliateSummary{
		UserID:  7,
		AffCode: "INVITER07",
	}), nil, nil, nil)
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
		affiliateService,
	)
	handler := &AuthHandler{
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

	affiliateService := service.NewAffiliateService(newInviteValidationAffiliateRepoStub(), nil, nil, nil)
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
		affiliateService,
	)
	handler := &AuthHandler{
		authService: authService,
	}

	handler.ValidateInvitationCode(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"valid":false`)
	require.Contains(t, rec.Body.String(), `"error_code":"INVITATION_CODE_REMOVED"`)
}
