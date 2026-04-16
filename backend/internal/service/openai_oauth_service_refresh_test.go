package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type openaiOAuthClientRefreshStub struct {
	refreshCalls     int32
	refreshTokenResp *openai.TokenResponse
	refreshErr       error
}

func (s *openaiOAuthClientRefreshStub) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI, proxyURL, clientID string) (*openai.TokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *openaiOAuthClientRefreshStub) RefreshToken(ctx context.Context, refreshToken, proxyURL string) (*openai.TokenResponse, error) {
	atomic.AddInt32(&s.refreshCalls, 1)
	if s.refreshErr != nil {
		return nil, s.refreshErr
	}
	if s.refreshTokenResp != nil {
		return s.refreshTokenResp, nil
	}
	return nil, errors.New("not implemented")
}

func (s *openaiOAuthClientRefreshStub) RefreshTokenWithClientID(ctx context.Context, refreshToken, proxyURL string, clientID string) (*openai.TokenResponse, error) {
	atomic.AddInt32(&s.refreshCalls, 1)
	if s.refreshErr != nil {
		return nil, s.refreshErr
	}
	if s.refreshTokenResp != nil {
		return s.refreshTokenResp, nil
	}
	return nil, errors.New("not implemented")
}

func TestOpenAIOAuthService_RefreshAccountToken_NoRefreshTokenUsesExistingAccessToken(t *testing.T) {
	client := &openaiOAuthClientRefreshStub{}
	svc := NewOpenAIOAuthService(nil, client)

	expiresAt := time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)
	account := &Account{
		ID:       77,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "existing-access-token",
			"expires_at":   expiresAt,
			"client_id":    "client-id-1",
		},
	}

	info, err := svc.RefreshAccountToken(context.Background(), account)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, "existing-access-token", info.AccessToken)
	require.Equal(t, "client-id-1", info.ClientID)
	require.Zero(t, atomic.LoadInt32(&client.refreshCalls), "existing access token should be reused without calling refresh")
}

func TestOpenAIOAuthService_RefreshTokenWithClientID_RefreshesPlanTypeFromBackendAPI(t *testing.T) {
	var checkCalls int32
	var privacyCalls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/backend-api/accounts/check/v4-2023-04-27":
			atomic.AddInt32(&checkCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"accounts": {
					"org_123": {
						"account": {
							"plan_type": "pro",
							"is_default": true
						}
					}
				}
			}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/backend-api/settings/account_user_setting":
			atomic.AddInt32(&privacyCalls, 1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &openaiOAuthClientRefreshStub{
		refreshTokenResp: &openai.TokenResponse{
			AccessToken:  buildFakeOpenAIJWT(t, map[string]any{"https://api.openai.com/auth": map[string]any{"poid": "org_123"}}),
			RefreshToken: "refresh-token-1",
			IDToken: buildFakeOpenAIJWT(t, map[string]any{
				"email": "stale@example.com",
				"https://api.openai.com/auth": map[string]any{
					"chatgpt_account_id": "acct_123",
					"chatgpt_user_id":    "user_123",
					"chatgpt_plan_type":  "free",
					"organizations": []map[string]any{
						{
							"id":         "org_123",
							"is_default": true,
						},
					},
				},
			}),
			ExpiresIn: 3600,
		},
	}
	svc := NewOpenAIOAuthService(nil, client)
	svc.SetPrivacyClientFactory(func(proxyURL string) (*req.Client, error) {
		target, err := url.Parse(server.URL)
		require.NoError(t, err)

		httpClient := req.C()
		httpClient.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
			return func(r *req.Request) (*req.Response, error) {
				rewritten := *r.URL
				rewritten.Scheme = target.Scheme
				rewritten.Host = target.Host
				r.URL = &rewritten
				if r.RawRequest != nil {
					r.RawRequest.URL = &rewritten
					r.RawRequest.Host = target.Host
				}
				return rt.RoundTrip(r)
			}
		})
		return httpClient, nil
	})

	info, err := svc.RefreshTokenWithClientID(context.Background(), "refresh-token-1", "", "client-id-1")
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, "pro", info.PlanType)
	require.Equal(t, "stale@example.com", info.Email)
	require.Equal(t, "client-id-1", info.ClientID)
	require.Equal(t, PrivacyModeTrainingOff, info.PrivacyMode)
	require.Equal(t, int32(1), atomic.LoadInt32(&checkCalls))
	require.Equal(t, int32(1), atomic.LoadInt32(&privacyCalls))
	require.Equal(t, int32(1), atomic.LoadInt32(&client.refreshCalls))
}

func buildFakeOpenAIJWT(t *testing.T, claims map[string]any) string {
	t.Helper()

	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["iat"] = time.Now().Unix()

	headerBytes, err := json.Marshal(map[string]any{"alg": "none", "typ": "JWT"})
	require.NoError(t, err)
	payloadBytes, err := json.Marshal(claims)
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(headerBytes) + "." +
		base64.RawURLEncoding.EncodeToString(payloadBytes) + ".sig"
}
