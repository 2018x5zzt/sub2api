//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type xlabOAuthCodeStoreStub struct {
	storedTokenID   string
	storedTTL       time.Duration
	consumedTokenID string
	consumeResult   bool
	err             error
}

func (s *xlabOAuthCodeStoreStub) StoreCode(_ context.Context, tokenID string, ttl time.Duration) error {
	s.storedTokenID = tokenID
	s.storedTTL = ttl
	return s.err
}

func (s *xlabOAuthCodeStoreStub) ConsumeCode(_ context.Context, tokenID string) (bool, error) {
	s.consumedTokenID = tokenID
	return s.consumeResult, s.err
}

func TestXlabOAuthProviderAuthorizeIssuesCodeForMikuCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{
		ID:           42,
		Email:        "artist@example.com",
		Username:     "artist",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		TokenVersion: 1,
	}
	handler := newXlabOAuthProviderTestHandler(user)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/authorize", strings.NewReader(`{
		"client_id":"miku-client",
		"redirect_uri":"https://ai.mikuapi.org/auth/xlab/callback",
		"response_type":"code",
		"scope":"profile email",
		"state":"state-from-miku"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.XlabOAuthAuthorize(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			RedirectURI string `json:"redirect_uri"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)

	redirectURL, err := url.Parse(resp.Data.RedirectURI)
	require.NoError(t, err)
	require.Equal(t, "https", redirectURL.Scheme)
	require.Equal(t, "ai.mikuapi.org", redirectURL.Host)
	require.Equal(t, "/auth/xlab/callback", redirectURL.Path)
	require.NotEmpty(t, redirectURL.Query().Get("code"))
	require.Equal(t, "state-from-miku", redirectURL.Query().Get("state"))
}

func TestXlabOAuthProviderRejectsUntrustedRedirectURI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newXlabOAuthProviderTestHandler(&service.User{ID: 42, Email: "artist@example.com", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 1})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/authorize", strings.NewReader(`{
		"client_id":"miku-client",
		"redirect_uri":"https://evil.example/auth/xlab/callback",
		"response_type":"code"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 42})

	handler.XlabOAuthAuthorize(c)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestXlabOAuthProviderAllowsMikuCallbackLayoutParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{ID: 42, Email: "artist@example.com", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 1}
	handler := newXlabOAuthProviderTestHandler(user)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/authorize", strings.NewReader(`{
		"client_id":"miku-app",
		"redirect_uri":"https://ai.mikuapi.org/auth/xlab/callback?layout=horizontal",
		"response_type":"code",
		"state":"image-studio"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.XlabOAuthAuthorize(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var resp struct {
		Data struct {
			RedirectURI string `json:"redirect_uri"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	redirectURL, err := url.Parse(resp.Data.RedirectURI)
	require.NoError(t, err)
	require.Equal(t, "horizontal", redirectURL.Query().Get("layout"))
	require.NotEmpty(t, redirectURL.Query().Get("code"))
	require.Equal(t, "image-studio", redirectURL.Query().Get("state"))
}

func TestXlabOAuthProviderAllowsItstudioCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{ID: 42, Email: "artist@example.com", Username: "artist", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 1}
	handler := newXlabOAuthProviderTestHandler(user)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/authorize", strings.NewReader(`{
		"client_id":"miku-app",
		"redirect_uri":"https://miku.app.itstudio.club/auth/xlab/callback",
		"response_type":"code",
		"scope":"profile email",
		"state":"a1d70692"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.XlabOAuthAuthorize(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var resp struct {
		Data struct {
			RedirectURI string `json:"redirect_uri"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	redirectURL, err := url.Parse(resp.Data.RedirectURI)
	require.NoError(t, err)
	require.Equal(t, "miku.app.itstudio.club", redirectURL.Host)
	require.Equal(t, "/auth/xlab/callback", redirectURL.Path)
	require.NotEmpty(t, redirectURL.Query().Get("code"))
	require.Equal(t, "a1d70692", redirectURL.Query().Get("state"))
}

func TestXlabOAuthProviderTokenAndUserInfoUsePlainOAuthJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{
		ID:           42,
		Email:        "artist@example.com",
		Username:     "artist",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		TokenVersion: 1,
	}
	handler := newXlabOAuthProviderTestHandler(user)

	code, err := handler.issueXlabOAuthCode(user, "miku-client", "https://ai.mikuapi.org/auth/xlab/callback")
	require.NoError(t, err)

	tokenRecorder := httptest.NewRecorder()
	tokenCtx, _ := gin.CreateTestContext(tokenRecorder)
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"https://ai.mikuapi.org/auth/xlab/callback"},
		"client_id":     {"miku-client"},
		"client_secret": {"miku-production-secret"},
	}
	tokenCtx.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	tokenCtx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.XlabOAuthToken(tokenCtx)

	require.Equal(t, http.StatusOK, tokenRecorder.Code)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	require.NoError(t, json.Unmarshal(tokenRecorder.Body.Bytes(), &tokenResp))
	require.NotEmpty(t, tokenResp.AccessToken)
	require.Equal(t, "Bearer", tokenResp.TokenType)
	require.Positive(t, tokenResp.ExpiresIn)

	userInfoRecorder := httptest.NewRecorder()
	userInfoCtx, _ := gin.CreateTestContext(userInfoRecorder)
	userInfoCtx.Request = httptest.NewRequest(http.MethodGet, "/oauth/userinfo", nil)
	userInfoCtx.Request.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	handler.XlabOAuthUserInfo(userInfoCtx)

	require.Equal(t, http.StatusOK, userInfoRecorder.Code)
	var userInfo struct {
		Sub               string `json:"sub"`
		ID                int64  `json:"id"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	require.NoError(t, json.Unmarshal(userInfoRecorder.Body.Bytes(), &userInfo))
	require.Equal(t, "42", userInfo.Sub)
	require.Equal(t, int64(42), userInfo.ID)
	require.Equal(t, "artist@example.com", userInfo.Email)
	require.Equal(t, "artist", userInfo.Name)
	require.Equal(t, "artist", userInfo.PreferredUsername)
}

func TestXlabOAuthProviderRejectsInvalidClientSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{ID: 42, Email: "artist@example.com", Username: "artist", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 1}
	handler := newXlabOAuthProviderTestHandler(user)

	code, err := handler.issueXlabOAuthCode(user, "miku-client", "https://ai.mikuapi.org/auth/xlab/callback")
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"https://ai.mikuapi.org/auth/xlab/callback"},
		"client_id":     {"miku-client"},
		"client_secret": {"wrong-secret"},
	}
	c.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.XlabOAuthToken(c)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestXlabOAuthProviderAuthorizationCodeCanOnlyBeRedeemedOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{ID: 42, Email: "artist@example.com", Username: "artist", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 1}
	handler := newXlabOAuthProviderTestHandler(user)

	code, err := handler.issueXlabOAuthCode(user, "miku-client", "https://ai.mikuapi.org/auth/xlab/callback")
	require.NoError(t, err)

	redeem := func() int {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {"https://ai.mikuapi.org/auth/xlab/callback"},
			"client_id":     {"miku-client"},
			"client_secret": {"miku-production-secret"},
		}
		c.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler.XlabOAuthToken(c)
		return recorder.Code
	}

	require.Equal(t, http.StatusOK, redeem())
	require.Equal(t, http.StatusUnauthorized, redeem())
}

func TestXlabOAuthProviderDelegatesAuthorizationCodeStorage(t *testing.T) {
	store := &xlabOAuthCodeStoreStub{consumeResult: true}
	handler := &AuthHandler{xlabOAuthCodeStore: store}
	ctx := context.Background()

	require.NoError(t, handler.storeXlabOAuthCode(ctx, "code-id", time.Minute))
	require.Equal(t, "code-id", store.storedTokenID)
	require.Equal(t, time.Minute, store.storedTTL)
	require.True(t, handler.consumeXlabOAuthCode(ctx, "code-id"))
	require.Equal(t, "code-id", store.consumedTokenID)

	store.err = errors.New("redis unavailable")
	require.Error(t, handler.storeXlabOAuthCode(ctx, "failed-code", time.Minute))
	require.False(t, handler.consumeXlabOAuthCode(ctx, "failed-code"))
}

func newXlabOAuthProviderTestHandler(user *service.User) *AuthHandler {
	repo := &userHandlerRepoStub{user: user}
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-jwt-secret-32bytes-long!!!"
	cfg.JWT.AccessTokenExpireMinutes = 60
	cfg.JWT.ExpireHour = 1
	cfg.XlabOAuthProvider.Clients = []config.XlabOAuthProviderClientConfig{
		{ClientID: "miku-client", ClientSecret: "miku-production-secret"},
		{ClientID: "miku-app", ClientSecret: "miku-production-secret"},
		{ClientID: "miku-prod", ClientSecret: "miku-production-secret"},
	}
	return &AuthHandler{
		cfg:         cfg,
		authService: service.NewAuthService(nil, repo, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil),
		userService: service.NewUserService(repo, nil, nil, nil),
	}
}

func TestXlabOAuthProviderCodeExpires(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{ID: 42, Email: "artist@example.com", Username: "artist", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 1}
	handler := newXlabOAuthProviderTestHandler(user)
	code, err := handler.issueXlabOAuthCodeWithTTL(user, "miku-client", "https://ai.mikuapi.org/auth/xlab/callback", -time.Second)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"https://ai.mikuapi.org/auth/xlab/callback"},
		"client_id":     {"miku-client"},
		"client_secret": {"miku-production-secret"},
	}
	c.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.XlabOAuthToken(c)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestXlabOAuthProviderAuthorizationCodeStaysCompactWithLargeProfileFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{
		ID:           42,
		Email:        "artist@example.com",
		Username:     strings.Repeat("u", 512),
		AvatarURL:    "https://cdn.example.com/avatar/" + strings.Repeat("a", 9000),
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		TokenVersion: 1,
	}
	handler := newXlabOAuthProviderTestHandler(user)

	code, err := handler.issueXlabOAuthCode(user, "miku-client", "https://ai.mikuapi.org/auth/xlab/callback?layout=horizontal")
	require.NoError(t, err)

	// The authorization code is embedded in browser callback URLs. Keep it small
	// so callback URLs stay below common proxy URI limits.
	require.Less(t, len(code), 1024)
}
