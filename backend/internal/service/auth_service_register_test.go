//go:build unit

package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type settingRepoStub struct {
	values map[string]string
	err    error
}

func (s *settingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}

func (s *settingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *settingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

type emailCacheStub struct {
	data *VerificationCodeData
	err  error
}

type defaultSubscriptionAssignerStub struct {
	calls []AssignSubscriptionInput
	err   error
}

type defaultProductSubscriptionAssignerStub struct {
	calls []AssignProductSubscriptionInput
	err   error
}

type legacyInvitationLookupStub struct {
	code     *RedeemCode
	err      error
	lastCode string
}

func (s *legacyInvitationLookupStub) GetByCode(_ context.Context, code string) (*RedeemCode, error) {
	s.lastCode = code
	if s.err != nil {
		return nil, s.err
	}
	if s.code == nil {
		return nil, ErrRedeemCodeNotFound
	}
	return s.code, nil
}

func (s *defaultSubscriptionAssignerStub) AssignOrExtendSubscription(_ context.Context, input *AssignSubscriptionInput) (*UserSubscription, bool, error) {
	if input != nil {
		s.calls = append(s.calls, *input)
	}
	if s.err != nil {
		return nil, false, s.err
	}
	return &UserSubscription{UserID: input.UserID, GroupID: input.GroupID}, false, nil
}

func (s *defaultProductSubscriptionAssignerStub) AssignOrExtendProductSubscription(_ context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error) {
	if input != nil {
		s.calls = append(s.calls, *input)
	}
	if s.err != nil {
		return nil, false, s.err
	}
	return &UserProductSubscription{UserID: input.UserID, ProductID: input.ProductID}, false, nil
}

func (s *emailCacheStub) GetVerificationCode(ctx context.Context, email string) (*VerificationCodeData, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

func (s *emailCacheStub) SetVerificationCode(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error {
	return nil
}

func (s *emailCacheStub) DeleteVerificationCode(ctx context.Context, email string) error {
	return nil
}

func (s *emailCacheStub) GetPasswordResetToken(ctx context.Context, email string) (*PasswordResetTokenData, error) {
	return nil, nil
}

func (s *emailCacheStub) SetPasswordResetToken(ctx context.Context, email string, data *PasswordResetTokenData, ttl time.Duration) error {
	return nil
}

func (s *emailCacheStub) DeletePasswordResetToken(ctx context.Context, email string) error {
	return nil
}

func (s *emailCacheStub) IsPasswordResetEmailInCooldown(ctx context.Context, email string) bool {
	return false
}

func (s *emailCacheStub) SetPasswordResetEmailCooldown(ctx context.Context, email string, ttl time.Duration) error {
	return nil
}

func newAuthService(repo *userRepoStub, settings map[string]string, emailCache EmailCache) *AuthService {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpireHour: 1,
		},
		Default: config.DefaultConfig{
			UserBalance:     3.5,
			UserConcurrency: 2,
		},
	}

	var settingService *SettingService
	if settings != nil {
		settingService = NewSettingService(&settingRepoStub{values: settings}, cfg)
	}

	var emailService *EmailService
	if emailCache != nil {
		emailService = NewEmailService(&settingRepoStub{values: settings}, emailCache)
	}

	return NewAuthService(
		nil, // entClient
		repo,
		nil, // redeemRepo
		nil, // refreshTokenCache
		cfg,
		settingService,
		emailService,
		nil,
		nil,
		nil, // promoService
		nil, // defaultSubAssigner
		nil, // affiliateService
	)
}

type inviteAuthUserRepoStub struct {
	userRepoStub
	usersByEmail      map[string]*User
	usersByInviteCode map[string]*User
}

func (s *inviteAuthUserRepoStub) Create(ctx context.Context, user *User) error {
	if err := s.userRepoStub.Create(ctx, user); err != nil {
		return err
	}
	if s.usersByEmail == nil {
		s.usersByEmail = map[string]*User{}
	}
	s.usersByEmail[user.Email] = user
	if s.usersByInviteCode == nil {
		s.usersByInviteCode = map[string]*User{}
	}
	if user.InviteCode != "" {
		s.usersByInviteCode[user.InviteCode] = user
	}
	return nil
}

func (s *inviteAuthUserRepoStub) GetByEmail(_ context.Context, email string) (*User, error) {
	u, ok := s.usersByEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *inviteAuthUserRepoStub) GetByInviteCode(_ context.Context, code string) (*User, error) {
	u, ok := s.usersByInviteCode[code]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *inviteAuthUserRepoStub) ExistsByInviteCode(_ context.Context, code string) (bool, error) {
	_, ok := s.usersByInviteCode[code]
	return ok, nil
}

func (s *inviteAuthUserRepoStub) CountInviteesByInviter(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

type authAffiliateRepoStub struct {
	summariesByUserID map[int64]*AffiliateSummary
	summariesByCode   map[string]*AffiliateSummary
	nextCode          string
}

func newAuthAffiliateRepoStub(nextCode string, summaries ...*AffiliateSummary) *authAffiliateRepoStub {
	repo := &authAffiliateRepoStub{
		summariesByUserID: map[int64]*AffiliateSummary{},
		summariesByCode:   map[string]*AffiliateSummary{},
		nextCode:          nextCode,
	}
	for _, summary := range summaries {
		cp := *summary
		repo.summariesByUserID[cp.UserID] = &cp
		if cp.AffCode != "" {
			repo.summariesByCode[strings.ToUpper(cp.AffCode)] = &cp
		}
	}
	return repo
}

func (s *authAffiliateRepoStub) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	if summary, ok := s.summariesByUserID[userID]; ok {
		return summary, nil
	}
	code := s.nextCode
	if code == "" {
		code = "DEFAULTAFF01"
	}
	summary := &AffiliateSummary{UserID: userID, AffCode: code}
	s.summariesByUserID[userID] = summary
	s.summariesByCode[strings.ToUpper(code)] = summary
	return summary, nil
}

func (s *authAffiliateRepoStub) GetAffiliateByCode(_ context.Context, code string) (*AffiliateSummary, error) {
	summary, ok := s.summariesByCode[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return nil, ErrAffiliateProfileNotFound
	}
	return summary, nil
}

func (s *authAffiliateRepoStub) BindInviter(_ context.Context, userID, inviterID int64) (bool, error) {
	self, ok := s.summariesByUserID[userID]
	if !ok {
		return false, ErrAffiliateProfileNotFound
	}
	if self.InviterID != nil {
		return false, nil
	}
	self.InviterID = &inviterID
	if inviter, ok := s.summariesByUserID[inviterID]; ok {
		inviter.AffCount++
	}
	return true, nil
}

func (s *authAffiliateRepoStub) AccrueQuota(context.Context, int64, int64, float64, int) (bool, error) {
	panic("unexpected AccrueQuota call")
}

func (s *authAffiliateRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	panic("unexpected GetAccruedRebateFromInvitee call")
}

func (s *authAffiliateRepoStub) CountEffectiveInvitees(context.Context, int64) (int64, error) {
	panic("unexpected CountEffectiveInvitees call")
}

func (s *authAffiliateRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	panic("unexpected ThawFrozenQuota call")
}

func (s *authAffiliateRepoStub) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	panic("unexpected TransferQuotaToBalance call")
}

func (s *authAffiliateRepoStub) ListInvitees(context.Context, int64, int) ([]AffiliateInvitee, error) {
	panic("unexpected ListInvitees call")
}

func (s *authAffiliateRepoStub) UpdateUserAffCode(context.Context, int64, string) error {
	panic("unexpected UpdateUserAffCode call")
}

func (s *authAffiliateRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	panic("unexpected ResetUserAffCode call")
}

func (s *authAffiliateRepoStub) SetUserRebateRate(context.Context, int64, *float64) error {
	panic("unexpected SetUserRebateRate call")
}

func (s *authAffiliateRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	panic("unexpected BatchSetUserRebateRate call")
}

func (s *authAffiliateRepoStub) ListUsersWithCustomSettings(context.Context, AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	panic("unexpected ListUsersWithCustomSettings call")
}

func newAuthServiceWithAffiliate(repo *inviteAuthUserRepoStub, affiliateRepo *authAffiliateRepoStub, settings map[string]string, emailCache EmailCache) *AuthService {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpireHour: 1,
		},
		Default: config.DefaultConfig{
			UserBalance:     3.5,
			UserConcurrency: 2,
		},
	}

	var settingService *SettingService
	if settings != nil {
		settingService = NewSettingService(&settingRepoStub{values: settings}, cfg)
	}

	var emailService *EmailService
	if emailCache != nil {
		emailService = NewEmailService(&settingRepoStub{values: settings}, emailCache)
	}

	var affiliateService *AffiliateService
	if affiliateRepo != nil {
		affiliateService = NewAffiliateService(affiliateRepo, settingService, nil, nil)
	}

	return NewAuthService(
		nil,
		repo,
		nil,
		nil,
		cfg,
		settingService,
		emailService,
		nil,
		nil,
		nil,
		nil,
		affiliateService,
	)
}

func TestAuthService_Register_Disabled(t *testing.T) {
	repo := &userRepoStub{}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "false",
	}, nil)

	_, _, err := service.Register(context.Background(), "user@test.com", "password")
	require.ErrorIs(t, err, ErrRegDisabled)
}

func TestAuthService_Register_DisabledByDefault(t *testing.T) {
	// 当 settings 为 nil（设置项不存在）时，注册应该默认关闭
	repo := &userRepoStub{}
	service := newAuthService(repo, nil, nil)

	_, _, err := service.Register(context.Background(), "user@test.com", "password")
	require.ErrorIs(t, err, ErrRegDisabled)
}

func TestAuthService_Register_EmailVerifyEnabledButServiceNotConfigured(t *testing.T) {
	repo := &userRepoStub{}
	// 邮件验证开启但 emailCache 为 nil（emailService 未配置）
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
		SettingKeyEmailVerifyEnabled:  "true",
	}, nil)

	// 应返回服务不可用错误，而不是允许绕过验证
	_, _, err := service.RegisterWithVerification(context.Background(), "user@test.com", "password", "any-code", "", "")
	require.ErrorIs(t, err, ErrServiceUnavailable)
}

func TestAuthService_Register_EmailVerifyRequired(t *testing.T) {
	repo := &userRepoStub{}
	cache := &emailCacheStub{} // 配置 emailService
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
		SettingKeyEmailVerifyEnabled:  "true",
	}, cache)

	_, _, err := service.RegisterWithVerification(context.Background(), "user@test.com", "password", "", "", "")
	require.ErrorIs(t, err, ErrEmailVerifyRequired)
}

func TestAuthService_Register_EmailVerifyInvalid(t *testing.T) {
	repo := &userRepoStub{}
	cache := &emailCacheStub{
		data: &VerificationCodeData{Code: "expected", Attempts: 0},
	}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
		SettingKeyEmailVerifyEnabled:  "true",
	}, cache)

	_, _, err := service.RegisterWithVerification(context.Background(), "user@test.com", "password", "wrong", "", "")
	require.ErrorIs(t, err, ErrInvalidVerifyCode)
	require.ErrorContains(t, err, "verify code")
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	repo := &userRepoStub{exists: true}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, _, err := service.Register(context.Background(), "user@test.com", "password")
	require.ErrorIs(t, err, ErrEmailExists)
}

func TestAuthService_Register_CheckEmailError(t *testing.T) {
	repo := &userRepoStub{existsErr: errors.New("db down")}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, _, err := service.Register(context.Background(), "user@test.com", "password")
	require.ErrorIs(t, err, ErrServiceUnavailable)
}

func TestAuthService_Register_ReservedEmail(t *testing.T) {
	repo := &userRepoStub{}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, _, err := service.Register(context.Background(), "linuxdo-123@linuxdo-connect.invalid", "password")
	require.ErrorIs(t, err, ErrEmailReserved)
}

func TestAuthService_Register_EmailSuffixNotAllowed(t *testing.T) {
	repo := &userRepoStub{}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled:              "true",
		SettingKeyRegistrationEmailSuffixWhitelist: `["@example.com","@company.com"]`,
	}, nil)

	_, _, err := service.Register(context.Background(), "user@other.com", "password")
	require.ErrorIs(t, err, ErrEmailSuffixNotAllowed)
	appErr := infraerrors.FromError(err)
	require.Contains(t, appErr.Message, "@example.com")
	require.Contains(t, appErr.Message, "@company.com")
	require.Equal(t, "EMAIL_SUFFIX_NOT_ALLOWED", appErr.Reason)
	require.Equal(t, "2", appErr.Metadata["allowed_suffix_count"])
	require.Equal(t, "@example.com,@company.com", appErr.Metadata["allowed_suffixes"])
}

func TestAuthService_Register_EmailSuffixAllowed(t *testing.T) {
	repo := &userRepoStub{nextID: 8}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled:              "true",
		SettingKeyRegistrationEmailSuffixWhitelist: `["example.com"]`,
	}, nil)

	_, user, err := service.Register(context.Background(), "user@example.com", "password")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(8), user.ID)
}

func TestAuthService_SendVerifyCode_EmailSuffixNotAllowed(t *testing.T) {
	repo := &userRepoStub{}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled:              "true",
		SettingKeyRegistrationEmailSuffixWhitelist: `["@example.com","@company.com"]`,
	}, nil)

	err := service.SendVerifyCode(context.Background(), "user@other.com")
	require.ErrorIs(t, err, ErrEmailSuffixNotAllowed)
	appErr := infraerrors.FromError(err)
	require.Contains(t, appErr.Message, "@example.com")
	require.Contains(t, appErr.Message, "@company.com")
	require.Equal(t, "2", appErr.Metadata["allowed_suffix_count"])
}

func TestAuthService_Register_CreateError(t *testing.T) {
	repo := &userRepoStub{createErr: errors.New("create failed")}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, _, err := service.Register(context.Background(), "user@test.com", "password")
	require.ErrorIs(t, err, ErrServiceUnavailable)
}

func TestAuthService_Register_CreateEmailExistsRace(t *testing.T) {
	// 模拟竞态条件：ExistsByEmail 返回 false，但 Create 时因唯一约束失败
	repo := &userRepoStub{createErr: ErrEmailExists}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, _, err := service.Register(context.Background(), "user@test.com", "password")
	require.ErrorIs(t, err, ErrEmailExists)
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := &userRepoStub{nextID: 5}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	token, user, err := service.Register(context.Background(), "user@test.com", "password")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, user)
	require.Equal(t, int64(5), user.ID)
	require.Equal(t, "user@test.com", user.Email)
	require.Equal(t, RoleUser, user.Role)
	require.Equal(t, StatusActive, user.Status)
	require.Equal(t, 3.5, user.Balance)
	require.Equal(t, 2, user.Concurrency)
	require.Len(t, repo.created, 1)
	require.True(t, user.CheckPassword("password"))
}

func TestAuthService_ValidateToken_ExpiredReturnsClaimsWithError(t *testing.T) {
	repo := &userRepoStub{}
	service := newAuthService(repo, nil, nil)

	// 创建用户并生成 token
	user := &User{
		ID:           1,
		Email:        "test@test.com",
		Role:         RoleUser,
		Status:       StatusActive,
		TokenVersion: 1,
	}
	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	// 验证有效 token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.Equal(t, int64(1), claims.UserID)

	// 模拟过期 token（通过创建一个过期很久的 token）
	service.cfg.JWT.ExpireHour = -1 // 设置为负数使 token 立即过期
	expiredToken, err := service.GenerateToken(user)
	require.NoError(t, err)
	service.cfg.JWT.ExpireHour = 1 // 恢复

	// 验证过期 token 应返回 claims 和 ErrTokenExpired
	claims, err = service.ValidateToken(expiredToken)
	require.ErrorIs(t, err, ErrTokenExpired)
	require.NotNil(t, claims, "claims should not be nil when token is expired")
	require.Equal(t, int64(1), claims.UserID)
	require.Equal(t, "test@test.com", claims.Email)
}

func TestAuthService_RefreshToken_ExpiredTokenNoPanic(t *testing.T) {
	user := &User{
		ID:           1,
		Email:        "test@test.com",
		Role:         RoleUser,
		Status:       StatusActive,
		TokenVersion: 1,
	}
	repo := &userRepoStub{user: user}
	service := newAuthService(repo, nil, nil)

	// 创建过期 token
	service.cfg.JWT.ExpireHour = -1
	expiredToken, err := service.GenerateToken(user)
	require.NoError(t, err)
	service.cfg.JWT.ExpireHour = 1

	// RefreshToken 使用过期 token 不应 panic
	require.NotPanics(t, func() {
		newToken, err := service.RefreshToken(context.Background(), expiredToken)
		require.NoError(t, err)
		require.NotEmpty(t, newToken)
	})
}

func TestAuthService_GetAccessTokenExpiresIn_FallbackToExpireHour(t *testing.T) {
	service := newAuthService(&userRepoStub{}, nil, nil)
	service.cfg.JWT.ExpireHour = 24
	service.cfg.JWT.AccessTokenExpireMinutes = 0

	require.Equal(t, 24*3600, service.GetAccessTokenExpiresIn())
}

func TestAuthService_GetAccessTokenExpiresIn_MinutesHasPriority(t *testing.T) {
	service := newAuthService(&userRepoStub{}, nil, nil)
	service.cfg.JWT.ExpireHour = 24
	service.cfg.JWT.AccessTokenExpireMinutes = 90

	require.Equal(t, 90*60, service.GetAccessTokenExpiresIn())
}

func TestAuthService_GenerateToken_UsesExpireHourWhenMinutesZero(t *testing.T) {
	service := newAuthService(&userRepoStub{}, nil, nil)
	service.cfg.JWT.ExpireHour = 24
	service.cfg.JWT.AccessTokenExpireMinutes = 0

	user := &User{
		ID:           1,
		Email:        "test@test.com",
		Role:         RoleUser,
		Status:       StatusActive,
		TokenVersion: 1,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.NotNil(t, claims.IssuedAt)
	require.NotNil(t, claims.ExpiresAt)

	require.WithinDuration(t, claims.IssuedAt.Time.Add(24*time.Hour), claims.ExpiresAt.Time, 2*time.Second)
}

func TestAuthService_GenerateToken_UsesMinutesWhenConfigured(t *testing.T) {
	service := newAuthService(&userRepoStub{}, nil, nil)
	service.cfg.JWT.ExpireHour = 24
	service.cfg.JWT.AccessTokenExpireMinutes = 90

	user := &User{
		ID:           2,
		Email:        "test2@test.com",
		Role:         RoleUser,
		Status:       StatusActive,
		TokenVersion: 1,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.NotNil(t, claims.IssuedAt)
	require.NotNil(t, claims.ExpiresAt)

	require.WithinDuration(t, claims.IssuedAt.Time.Add(90*time.Minute), claims.ExpiresAt.Time, 2*time.Second)
}

func TestAuthService_Register_AssignsDefaultSubscriptions(t *testing.T) {
	repo := &userRepoStub{nextID: 42}
	assigner := &defaultSubscriptionAssignerStub{}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled:  "true",
		SettingKeyDefaultSubscriptions: `[{"group_id":11,"validity_days":30},{"group_id":12,"validity_days":7}]`,
	}, nil)
	service.defaultSubAssigner = assigner

	_, user, err := service.Register(context.Background(), "default-sub@test.com", "password")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, assigner.calls, 2)
	require.Equal(t, int64(42), assigner.calls[0].UserID)
	require.Equal(t, int64(11), assigner.calls[0].GroupID)
	require.Equal(t, 30, assigner.calls[0].ValidityDays)
	require.Equal(t, int64(12), assigner.calls[1].GroupID)
	require.Equal(t, 7, assigner.calls[1].ValidityDays)
}

func TestRegisterUser_AppliesDefaultProductSubscriptions(t *testing.T) {
	repo := &userRepoStub{nextID: 43}
	assigner := &defaultProductSubscriptionAssignerStub{}
	service := newAuthService(repo, map[string]string{
		SettingKeyRegistrationEnabled:         "true",
		SettingKeyDefaultSubscriptionProducts: `[{"product_id":21,"validity_days":30},{"product_id":22,"validity_days":7}]`,
	}, nil)
	service.defaultProductSubAssigner = assigner

	_, user, err := service.Register(context.Background(), "default-product-sub@test.com", "password")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, assigner.calls, 2)
	require.Equal(t, int64(43), assigner.calls[0].UserID)
	require.Equal(t, int64(21), assigner.calls[0].ProductID)
	require.Equal(t, 30, assigner.calls[0].ValidityDays)
	require.Equal(t, int64(22), assigner.calls[1].ProductID)
	require.Equal(t, 7, assigner.calls[1].ValidityDays)
}

func TestAuthService_RegisterWithVerification_BindsAffiliateInviterCode(t *testing.T) {
	repo := &inviteAuthUserRepoStub{
		userRepoStub: userRepoStub{nextID: 8},
	}
	affiliateRepo := newAuthAffiliateRepoStub("NEWCODE08", &AffiliateSummary{
		UserID:  7,
		AffCode: "INVITER07",
	})
	service := newAuthServiceWithAffiliate(repo, affiliateRepo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, user, err := service.RegisterWithVerification(context.Background(), "user@test.com", "password", "", "", "INVITER07")
	require.NoError(t, err)
	require.Equal(t, int64(8), user.ID)
	require.Equal(t, "NEWCODE08", affiliateRepo.summariesByUserID[8].AffCode)
	require.NotNil(t, affiliateRepo.summariesByUserID[8].InviterID)
	require.EqualValues(t, 7, *affiliateRepo.summariesByUserID[8].InviterID)
	require.Equal(t, 1, affiliateRepo.summariesByUserID[7].AffCount)
}

func TestAuthService_RegisterWithVerification_BindsMixedCaseInviterCode(t *testing.T) {
	repo := &inviteAuthUserRepoStub{
		userRepoStub: userRepoStub{nextID: 18},
	}
	affiliateRepo := newAuthAffiliateRepoStub("QWERTYUI", &AffiliateSummary{
		UserID:  17,
		AffCode: "AbCdEfGh",
	})
	service := newAuthServiceWithAffiliate(repo, affiliateRepo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, user, err := service.RegisterWithVerification(context.Background(), "mixed-user@test.com", "password", "", "", " AbCdEfGh ")
	require.NoError(t, err)
	require.Equal(t, int64(18), user.ID)
	require.NotNil(t, affiliateRepo.summariesByUserID[18].InviterID)
	require.EqualValues(t, 17, *affiliateRepo.summariesByUserID[18].InviterID)
}

func TestAuthService_ValidateInvitationCode_RejectsUnknownAffiliateCode(t *testing.T) {
	repo := &inviteAuthUserRepoStub{userRepoStub: userRepoStub{nextID: 9}}
	affiliateRepo := newAuthAffiliateRepoStub("NEWCODE09")
	service := newAuthServiceWithAffiliate(repo, affiliateRepo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	err := service.ValidateInvitationCode(context.Background(), "MISSING")
	require.ErrorIs(t, err, ErrInvitationCodeInvalid)
}

func TestAuthService_RegisterWithVerification_IgnoresUnknownAffiliateCode(t *testing.T) {
	repo := &inviteAuthUserRepoStub{userRepoStub: userRepoStub{nextID: 10}}
	affiliateRepo := newAuthAffiliateRepoStub("NEWCODE10")
	service := newAuthServiceWithAffiliate(repo, affiliateRepo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)

	_, user, err := service.RegisterWithVerification(context.Background(), "user@test.com", "password", "", "", "MISSING")
	require.NoError(t, err)
	require.Equal(t, int64(10), user.ID)
	require.Nil(t, affiliateRepo.summariesByUserID[10].InviterID)
}

func TestAuthService_ValidateInvitationCode_ReturnsRemovedErrorForLegacyInvitationRedeemCode(t *testing.T) {
	repo := &inviteAuthUserRepoStub{userRepoStub: userRepoStub{nextID: 10}}
	affiliateRepo := newAuthAffiliateRepoStub("NEWCODE10")
	legacyRepo := &legacyInvitationLookupStub{
		code: &RedeemCode{
			Code: "LEGACY-A1B2",
			Type: RedeemTypeInvitation,
		},
	}
	service := newAuthServiceWithAffiliate(repo, affiliateRepo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)
	service.redeemRepo = legacyRepo

	err := service.ValidateInvitationCode(context.Background(), " legacy-a1b2 ")
	require.ErrorIs(t, err, ErrInvitationCodeRemoved)
	require.Equal(t, "LEGACY-A1B2", legacyRepo.lastCode)
}

type refreshTokenCacheStub struct{}

func (s *refreshTokenCacheStub) StoreRefreshToken(ctx context.Context, tokenHash string, data *RefreshTokenData, ttl time.Duration) error {
	return nil
}

func (s *refreshTokenCacheStub) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenData, error) {
	return nil, ErrRefreshTokenNotFound
}

func (s *refreshTokenCacheStub) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	return nil
}

func (s *refreshTokenCacheStub) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	return nil
}

func (s *refreshTokenCacheStub) DeleteTokenFamily(ctx context.Context, familyID string) error {
	return nil
}

func (s *refreshTokenCacheStub) AddToUserTokenSet(ctx context.Context, userID int64, tokenHash string, ttl time.Duration) error {
	return nil
}

func (s *refreshTokenCacheStub) AddToFamilyTokenSet(ctx context.Context, familyID string, tokenHash string, ttl time.Duration) error {
	return nil
}

func (s *refreshTokenCacheStub) GetUserTokenHashes(ctx context.Context, userID int64) ([]string, error) {
	return nil, nil
}

func (s *refreshTokenCacheStub) GetFamilyTokenHashes(ctx context.Context, familyID string) ([]string, error) {
	return nil, nil
}

func (s *refreshTokenCacheStub) IsTokenInFamily(ctx context.Context, familyID string, tokenHash string) (bool, error) {
	return false, nil
}

func TestAuthService_LoginOrRegisterOAuthWithTokenPair_IgnoresRetiredInvitationToggle(t *testing.T) {
	repo := &inviteAuthUserRepoStub{userRepoStub: userRepoStub{nextID: 11}}
	affiliateRepo := newAuthAffiliateRepoStub("NEWCODE11")
	service := newAuthServiceWithAffiliate(repo, affiliateRepo, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)
	service.refreshTokenCache = &refreshTokenCacheStub{}
	service.cfg.JWT.RefreshTokenExpireDays = 7

	tokenPair, user, err := service.LoginOrRegisterOAuthWithTokenPair(context.Background(), "oauth-no-invite@test.com", "oauth-user", "")
	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	require.NotNil(t, user)
	require.Nil(t, user.InvitedByUserID)
}
