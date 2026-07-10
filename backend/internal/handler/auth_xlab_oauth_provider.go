package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	xlabOAuthProviderClientMiku       = "miku-client"
	xlabOAuthProviderClientMikuApp    = "miku-app"
	xlabOAuthProviderClientMikuProd   = "miku-prod"
	xlabOAuthProviderCodeTTL          = 2 * time.Minute
	xlabOAuthProviderAccessTokenTTL   = 15 * time.Minute
	xlabOAuthProviderPurposeCode      = "xlab_oauth_code"
	xlabOAuthProviderPurposeUserToken = "xlab_oauth_access_token"
)

var xlabOAuthProviderMemoryCodes = newXlabOAuthProviderMemoryCodeStore()

type xlabOAuthAuthorizeRequest struct {
	ClientID     string `json:"client_id" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	ResponseType string `json:"response_type" binding:"required"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
}

type xlabOAuthAuthorizeResponse struct {
	RedirectURI string `json:"redirect_uri"`
}

type xlabOAuthProviderClaims struct {
	Purpose     string `json:"purpose"`
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Email       string `json:"email,omitempty"`
	Username    string `json:"preferred_username,omitempty"`
	AvatarURL   string `json:"picture,omitempty"`
	Role        string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

func (h *AuthHandler) XlabOAuthAuthorize(c *gin.Context) {
	var req xlabOAuthAuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.RedirectURI = strings.TrimSpace(req.RedirectURI)
	req.ResponseType = strings.TrimSpace(req.ResponseType)

	if req.ResponseType != "code" {
		response.BadRequest(c, "unsupported response_type")
		return
	}
	if !isAllowedXlabOAuthClient(req.ClientID) {
		response.BadRequest(c, "oauth client not found")
		return
	}
	if !isAllowedXlabOAuthRedirectURI(req.RedirectURI) {
		response.BadRequest(c, "invalid redirect_uri")
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "login required")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := ensureLoginUserActive(user); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	code, err := h.issueXlabOAuthCode(user, req.ClientID, req.RedirectURI)
	if err != nil {
		response.InternalError(c, "failed to create authorization code")
		return
	}

	redirectURL, err := url.Parse(req.RedirectURI)
	if err != nil {
		response.BadRequest(c, "invalid redirect_uri")
		return
	}
	query := redirectURL.Query()
	query.Set("code", code)
	if req.State != "" {
		query.Set("state", req.State)
	}
	redirectURL.RawQuery = query.Encode()

	response.Success(c, xlabOAuthAuthorizeResponse{RedirectURI: redirectURL.String()})
}

func (h *AuthHandler) XlabOAuthToken(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "invalid form body"})
		return
	}

	grantType := strings.TrimSpace(c.PostForm("grant_type"))
	code := strings.TrimSpace(c.PostForm("code"))
	redirectURI := strings.TrimSpace(c.PostForm("redirect_uri"))
	clientID := strings.TrimSpace(c.PostForm("client_id"))
	clientSecret := strings.TrimSpace(c.PostForm("client_secret"))

	if grantType != "authorization_code" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
		return
	}
	if !h.isAllowedXlabOAuthClientSecret(clientID, clientSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client", "error_description": "oauth client not found"})
		return
	}

	claims, err := h.parseXlabOAuthToken(code, xlabOAuthProviderPurposeCode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_grant", "error_description": "authorization code is invalid or expired"})
		return
	}
	if claims.ClientID != clientID || claims.RedirectURI != redirectURI || !isAllowedXlabOAuthRedirectURI(redirectURI) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_grant", "error_description": "authorization code does not match this client"})
		return
	}
	if claims.ID == "" || !h.consumeXlabOAuthCode(c.Request.Context(), claims.ID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_grant", "error_description": "authorization code is invalid or expired"})
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_grant", "error_description": "authorization code subject is invalid"})
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil || !user.IsActive() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_grant", "error_description": "user is not available"})
		return
	}

	accessToken, err := h.issueXlabOAuthAccessToken(user, clientID, claims.Scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(xlabOAuthProviderAccessTokenTTL.Seconds()),
	})
}

func (h *AuthHandler) XlabOAuthUserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}

	claims, err := h.parseXlabOAuthToken(strings.TrimSpace(parts[1]), xlabOAuthProviderPurposeUserToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil || !user.IsActive() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}

	username := strings.TrimSpace(user.Username)
	if username == "" {
		username = strings.TrimSpace(claims.Username)
	}
	if username == "" {
		username = "xlab_" + strconv.FormatInt(user.ID, 10)
	}
	picture := strings.TrimSpace(user.AvatarURL)
	if picture == "" {
		picture = strings.TrimSpace(claims.AvatarURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"sub":                strconv.FormatInt(user.ID, 10),
		"id":                 user.ID,
		"email":              user.Email,
		"name":               username,
		"preferred_username": username,
		"picture":            picture,
	})
}

func (h *AuthHandler) issueXlabOAuthCode(user *service.User, clientID, redirectURI string) (string, error) {
	return h.issueXlabOAuthCodeWithTTL(user, clientID, redirectURI, xlabOAuthProviderCodeTTL)
}

func (h *AuthHandler) issueXlabOAuthCodeWithTTL(user *service.User, clientID, redirectURI string, ttl time.Duration) (string, error) {
	jti, err := newXlabOAuthProviderTokenID()
	if err != nil {
		return "", err
	}
	token, err := h.issueXlabOAuthTokenWithID(user, clientID, redirectURI, "", xlabOAuthProviderPurposeCode, ttl, jti)
	if err != nil {
		return "", err
	}
	if err := h.storeXlabOAuthCode(context.Background(), jti, ttl); err != nil {
		return "", err
	}
	return token, nil
}

func (h *AuthHandler) issueXlabOAuthAccessToken(user *service.User, clientID, scope string) (string, error) {
	return h.issueXlabOAuthToken(user, clientID, "", scope, xlabOAuthProviderPurposeUserToken, xlabOAuthProviderAccessTokenTTL)
}

func (h *AuthHandler) issueXlabOAuthToken(user *service.User, clientID, redirectURI, scope, purpose string, ttl time.Duration) (string, error) {
	return h.issueXlabOAuthTokenWithID(user, clientID, redirectURI, scope, purpose, ttl, "")
}

func (h *AuthHandler) issueXlabOAuthTokenWithID(user *service.User, clientID, redirectURI, scope, purpose string, ttl time.Duration, tokenID string) (string, error) {
	if h == nil || h.cfg == nil || strings.TrimSpace(h.cfg.JWT.Secret) == "" {
		return "", fmt.Errorf("jwt secret is not configured")
	}
	if user == nil || user.ID <= 0 {
		return "", infraerrors.Unauthorized("INVALID_USER", "user not found")
	}
	now := time.Now()
	claims := xlabOAuthProviderClaims{
		Purpose:     purpose,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Audience:  []string{clientID},
			Issuer:    "xlabapi",
		},
	}
	// Keep authorization-code JWTs compact. They are returned in browser URLs,
	// so embedding large profile fields can exceed common URI length limits.
	if purpose == xlabOAuthProviderPurposeUserToken {
		claims.Email = user.Email
		claims.Username = user.Username
		claims.AvatarURL = user.AvatarURL
		claims.Role = user.Role
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.cfg.JWT.Secret))
}

func (h *AuthHandler) parseXlabOAuthToken(rawToken, purpose string) (*xlabOAuthProviderClaims, error) {
	if h == nil || h.cfg == nil || strings.TrimSpace(h.cfg.JWT.Secret) == "" {
		return nil, service.ErrInvalidToken
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	token, err := parser.ParseWithClaims(rawToken, &xlabOAuthProviderClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*xlabOAuthProviderClaims)
	if !ok || !token.Valid || claims.Purpose != purpose {
		return nil, service.ErrInvalidToken
	}
	return claims, nil
}

func isAllowedXlabOAuthClient(clientID string) bool {
	switch strings.TrimSpace(clientID) {
	case xlabOAuthProviderClientMiku, xlabOAuthProviderClientMikuApp, xlabOAuthProviderClientMikuProd:
		return true
	default:
		return false
	}
}

func (h *AuthHandler) isAllowedXlabOAuthClientSecret(clientID, clientSecret string) bool {
	clientID = strings.TrimSpace(clientID)
	if !isAllowedXlabOAuthClient(clientID) {
		return false
	}
	expectedSecret, ok := h.xlabOAuthClientSecret(clientID)
	if !ok || expectedSecret == "" || clientSecret == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expectedSecret), []byte(strings.TrimSpace(clientSecret))) == 1
}

func (h *AuthHandler) xlabOAuthClientSecret(clientID string) (string, bool) {
	if h == nil || h.cfg == nil {
		return "", false
	}
	for _, client := range h.cfg.XlabOAuthProvider.Clients {
		if strings.TrimSpace(client.ClientID) == clientID {
			return strings.TrimSpace(client.ClientSecret), true
		}
	}
	return "", false
}

func newXlabOAuthProviderTokenID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (h *AuthHandler) storeXlabOAuthCode(ctx context.Context, jti string, ttl time.Duration) error {
	jti = strings.TrimSpace(jti)
	if jti == "" {
		return fmt.Errorf("authorization code id is empty")
	}
	if h != nil && h.xlabOAuthCodeStore != nil {
		return h.xlabOAuthCodeStore.StoreCode(ctx, jti, ttl)
	}
	xlabOAuthProviderMemoryCodes.store(jti, ttl)
	return nil
}

func (h *AuthHandler) consumeXlabOAuthCode(ctx context.Context, jti string) bool {
	jti = strings.TrimSpace(jti)
	if jti == "" {
		return false
	}
	if h != nil && h.xlabOAuthCodeStore != nil {
		consumed, err := h.xlabOAuthCodeStore.ConsumeCode(ctx, jti)
		return err == nil && consumed
	}
	return xlabOAuthProviderMemoryCodes.consume(jti)
}

type xlabOAuthProviderMemoryCodeStore struct {
	mu    sync.Mutex
	codes map[string]time.Time
}

func newXlabOAuthProviderMemoryCodeStore() *xlabOAuthProviderMemoryCodeStore {
	return &xlabOAuthProviderMemoryCodeStore{codes: make(map[string]time.Time)}
}

func (s *xlabOAuthProviderMemoryCodeStore) store(jti string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[jti] = time.Now().Add(ttl)
}

func (s *xlabOAuthProviderMemoryCodeStore) consume(jti string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, ok := s.codes[jti]
	if !ok {
		return false
	}
	delete(s.codes, jti)
	return time.Now().Before(expiresAt)
}

func isAllowedXlabOAuthRedirectURI(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" {
		return false
	}
	if !isAllowedXlabOAuthRedirectHost(u.Host) {
		return false
	}
	if u.Path != "/auth/xlab/callback" {
		return false
	}
	for key, values := range u.Query() {
		if key != "layout" {
			return false
		}
		if len(values) != 1 || (values[0] != "horizontal" && values[0] != "vertical") {
			return false
		}
	}
	return true
}

func isAllowedXlabOAuthRedirectHost(host string) bool {
	switch host {
	case "ai.mikuapi.org", "mikuapi.org", "miku.app.itstudio.club":
		return true
	default:
		return false
	}
}
