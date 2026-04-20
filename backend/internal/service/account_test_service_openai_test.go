//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type openAIAccountTestRepo struct {
	mockAccountRepoForGemini
	updatedExtra  map[string]any
	rateLimitedID int64
	rateLimitedAt *time.Time
	setErrorCalls int
	lastErrorMsg  string
}

func (r *openAIAccountTestRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.updatedExtra = updates
	return nil
}

func (r *openAIAccountTestRepo) SetRateLimited(_ context.Context, id int64, resetAt time.Time) error {
	r.rateLimitedID = id
	r.rateLimitedAt = &resetAt
	return nil
}

func (r *openAIAccountTestRepo) SetError(_ context.Context, _ int64, errorMsg string) error {
	r.setErrorCalls++
	r.lastErrorMsg = errorMsg
	return nil
}

func TestAccountTestService_OpenAISuccessPersistsSnapshotFromHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, recorder := newSoraTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader(`data: {"type":"response.completed"}

`))
	resp.Header.Set("x-codex-primary-used-percent", "88")
	resp.Header.Set("x-codex-primary-reset-after-seconds", "604800")
	resp.Header.Set("x-codex-primary-window-minutes", "10080")
	resp.Header.Set("x-codex-secondary-used-percent", "42")
	resp.Header.Set("x-codex-secondary-reset-after-seconds", "18000")
	resp.Header.Set("x-codex-secondary-window-minutes", "300")

	repo := &openAIAccountTestRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
	account := &Account{
		ID:          89,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4")
	require.NoError(t, err)
	require.NotEmpty(t, repo.updatedExtra)
	require.Equal(t, 42.0, repo.updatedExtra["codex_5h_used_percent"])
	require.Equal(t, 88.0, repo.updatedExtra["codex_7d_used_percent"])
	require.Contains(t, recorder.Body.String(), "test_complete")
}

func TestAccountTestService_OpenAI429PersistsSnapshotWithoutRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newSoraTestContext()

	resp := newJSONResponse(http.StatusTooManyRequests, `{"error":{"type":"usage_limit_reached","message":"limit reached"}}`)
	resp.Header.Set("x-codex-primary-used-percent", "100")
	resp.Header.Set("x-codex-primary-reset-after-seconds", "604800")
	resp.Header.Set("x-codex-primary-window-minutes", "10080")
	resp.Header.Set("x-codex-secondary-used-percent", "100")
	resp.Header.Set("x-codex-secondary-reset-after-seconds", "18000")
	resp.Header.Set("x-codex-secondary-window-minutes", "300")

	repo := &openAIAccountTestRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
	account := &Account{
		ID:          88,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4")
	require.Error(t, err)
	require.NotEmpty(t, repo.updatedExtra)
	require.Equal(t, 100.0, repo.updatedExtra["codex_5h_used_percent"])
	require.Zero(t, repo.rateLimitedID)
	require.Nil(t, repo.rateLimitedAt)
	require.Nil(t, account.RateLimitResetAt)
}

func TestAccountTestService_OpenAIRemapsLegacyOAuthModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, recorder := newSoraTestContext()

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"id":"resp_ok","usage":{"input_tokens":1,"output_tokens":1}}`),
		},
	}
	svc := &AccountTestService{httpUpstream: upstream}
	account := &Account{
		ID:          87,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.2-codex")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Contains(t, recorder.Body.String(), "test_complete")

	body, readErr := io.ReadAll(upstream.requests[0].Body)
	require.NoError(t, readErr)
	require.Equal(t, "gpt-5.3-codex", gjson.GetBytes(body, "model").String())
}

func TestAccountTestService_OpenAI401MarksPermanentError(t *testing.T) {
	t.Run("oauth accounts are marked permanent error", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ctx, _ := newSoraTestContext()

		resp := newJSONResponse(http.StatusUnauthorized, `{"detail":"Unauthorized"}`)
		repo := &openAIAccountTestRepo{}
		upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
		svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
		account := &Account{
			ID:          90,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Concurrency: 1,
			Credentials: map[string]any{"access_token": "test-token"},
		}

		err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4")
		require.Error(t, err)
		require.Equal(t, 1, repo.setErrorCalls)
		require.Contains(t, repo.lastErrorMsg, "Authentication failed (401)")
		require.Contains(t, repo.lastErrorMsg, "Unauthorized")
	})

	t.Run("api key accounts are not marked permanent error", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ctx, _ := newSoraTestContext()

		resp := newJSONResponse(http.StatusUnauthorized, `{"detail":"Unauthorized"}`)
		repo := &openAIAccountTestRepo{}
		upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
		svc := &AccountTestService{
			accountRepo:  repo,
			httpUpstream: upstream,
			cfg: &config.Config{
				Security: config.SecurityConfig{
					URLAllowlist: config.URLAllowlistConfig{Enabled: false},
				},
			},
		}
		account := &Account{
			ID:          91,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeAPIKey,
			Concurrency: 1,
			Credentials: map[string]any{
				"api_key":  "test-token",
				"base_url": "https://api.openai.com/responses",
			},
		}

		err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4")
		require.Error(t, err)
		require.Zero(t, repo.setErrorCalls)
		require.Empty(t, repo.lastErrorMsg)
	})
}
