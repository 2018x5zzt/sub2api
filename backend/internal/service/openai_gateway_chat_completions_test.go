package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestForwardAsChatCompletions_DefaultMappedModelPreservesReasoningSemantics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"plus","messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.completed","response":{"id":"resp_reasoning","status":"completed","output":[{"type":"reasoning","summary":[{"type":"summary_text","text":"plan first"}]},{"type":"message","content":[{"type":"output_text","text":"final answer"}]}],"usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}}`,
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-reasoning"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}
	upstream := &httpUpstreamRecorder{resp: resp}

	svc := &OpenAIGatewayService{
		cfg:          &config.Config{Gateway: config.GatewayConfig{}},
		httpUpstream: upstream,
	}

	account := &Account{
		ID:             123,
		Name:           "acc",
		Platform:       PlatformOpenAI,
		Type:           AccountTypeOAuth,
		Concurrency:    1,
		Credentials:    map[string]any{"access_token": "oauth-token", "chatgpt_account_id": "chatgpt-acc"},
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "gpt-5.1-high")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "gpt-5.1", gjson.GetBytes(upstream.lastBody, "model").String(), "upstream model should stay normalized")
	require.Equal(t, "high", gjson.GetBytes(upstream.lastBody, "reasoning.effort").String(), "mapped model suffix should become reasoning.effort")
	require.Equal(t, "auto", gjson.GetBytes(upstream.lastBody, "reasoning.summary").String(), "compat path should request displayable reasoning summary")

	require.Equal(t, "plan first", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.reasoning_content").String())
	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.content").String())
}
