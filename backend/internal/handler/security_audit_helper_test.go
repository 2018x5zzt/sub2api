package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCachesSecurityAuditCompletionSkipsWebSocketStages(t *testing.T) {
	require.True(t, cachesSecurityAuditCompletion("http"))
	require.True(t, cachesSecurityAuditCompletion(""))
	require.False(t, cachesSecurityAuditCompletion("first_turn"))
	require.False(t, cachesSecurityAuditCompletion("subsequent_turn"))
}

func TestRunSecurityAuditDoesNotSkipSubsequentWebSocketTurns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	subject := middleware2.AuthSubject{UserID: 7, Concurrency: 1}
	first := runSecurityAudit(c, nil, coordinator, nil, nil, subject, "openai_responses", "gpt-test",
		[]byte(`{"type":"response.create","response":{"input":"benign"}}`), "first_turn")
	require.NotNil(t, first)
	require.True(t, first.AllowNextStage)
	require.Equal(t, int64(1), engine.enqueues.Load())
	_, cached := c.Get(securityAuditCompletedContextKey)
	require.False(t, cached, "WebSocket stages must not set the HTTP completion cache")

	// Even if an HTTP path previously cached completion on this Context, WS turns
	// must still audit every response.create payload.
	c.Set(securityAuditCompletedContextKey, securityAuditCompletionToken{})

	second := runSecurityAudit(c, nil, coordinator, nil, nil, subject, "openai_responses", "gpt-test",
		[]byte(`{"type":"response.create","response":{"input":"malicious follow-up"}}`), "subsequent_turn")
	require.NotNil(t, second)
	require.Equal(t, int64(2), engine.enqueues.Load(), "subsequent WebSocket turns must be audited again")
}

func TestRunSecurityAuditCompletionCacheRequiresExactInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(3)
	changedGroupID := int64(4)
	baseAPIKey := &service.APIKey{ID: 9, UserID: 7, GroupID: &groupID, Group: &service.Group{ID: groupID}}
	baseSubject := middleware2.AuthSubject{UserID: 7, Concurrency: 1}
	baseBody := []byte(`{"prompt":"same input"}`)

	tests := []struct {
		name     string
		apiKey   *service.APIKey
		subject  middleware2.AuthSubject
		protocol string
		model    string
		body     []byte
		wantRuns int64
	}{
		{name: "same input", apiKey: baseAPIKey, subject: baseSubject, protocol: "openai_images", model: "gpt-image-2", body: baseBody, wantRuns: 1},
		{name: "changed body", apiKey: baseAPIKey, subject: baseSubject, protocol: "openai_images", model: "gpt-image-2", body: []byte(`{"prompt":"changed"}`), wantRuns: 2},
		{name: "changed protocol", apiKey: baseAPIKey, subject: baseSubject, protocol: "openai_responses", model: "gpt-image-2", body: baseBody, wantRuns: 2},
		{name: "changed model", apiKey: baseAPIKey, subject: baseSubject, protocol: "openai_images", model: "gpt-image-3", body: baseBody, wantRuns: 2},
		{name: "changed user", apiKey: baseAPIKey, subject: middleware2.AuthSubject{UserID: 8, Concurrency: 1}, protocol: "openai_images", model: "gpt-image-2", body: baseBody, wantRuns: 2},
		{name: "changed api key", apiKey: &service.APIKey{ID: 10, UserID: 7, GroupID: &groupID, Group: &service.Group{ID: groupID}}, subject: baseSubject, protocol: "openai_images", model: "gpt-image-2", body: baseBody, wantRuns: 2},
		{name: "changed group", apiKey: &service.APIKey{ID: 9, UserID: 7, GroupID: &changedGroupID, Group: &service.Group{ID: changedGroupID}}, subject: baseSubject, protocol: "openai_images", model: "gpt-image-2", body: baseBody, wantRuns: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
			coordinator := securityaudit.NewCoordinator(nil, engine)
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)

			first := runSecurityAudit(c, nil, coordinator, nil, baseAPIKey, baseSubject, "openai_images", "gpt-image-2", baseBody, "http")
			require.NotNil(t, first)
			require.True(t, first.AllowNextStage)
			second := runSecurityAudit(c, nil, coordinator, nil, tt.apiKey, tt.subject, tt.protocol, tt.model, tt.body, "http")
			if tt.wantRuns == 1 {
				require.Nil(t, second)
			} else {
				require.NotNil(t, second)
			}
			require.Equal(t, tt.wantRuns, engine.enqueues.Load())
		})
	}
}

type turnCountingEngine struct {
	mode     securityaudit.Mode
	enqueues atomic.Int64
}

func (e *turnCountingEngine) EffectiveMode() securityaudit.Mode { return e.mode }
func (e *turnCountingEngine) Enqueue(context.Context, securityaudit.Request) error {
	e.enqueues.Add(1)
	return nil
}
func (e *turnCountingEngine) Evaluate(context.Context, securityaudit.Request) (*securityaudit.PromptDecision, error) {
	return &securityaudit.PromptDecision{Kind: securityaudit.DecisionAllow, AllowNextStage: true}, nil
}
