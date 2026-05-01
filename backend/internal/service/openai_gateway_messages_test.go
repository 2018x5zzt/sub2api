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

func TestForwardAsAnthropic_BufferedDeltaOnlyTerminalResponsePreservesThinkingAndText(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"claude-opus-4-1","max_tokens":1024,"messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_messages_gap","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.reasoning_summary_text.delta","output_index":0,"summary_index":0,"delta":"plan first"}`,
		"",
		`data: {"type":"response.output_text.delta","output_index":1,"content_index":0,"delta":"final answer"}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_messages_gap","status":"completed","usage":{"input_tokens":11,"output_tokens":7}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-messages-gap"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}
	upstream := &httpUpstreamRecorder{resp: resp}

	cfg := &config.Config{Gateway: config.GatewayConfig{}}
	svc := &OpenAIGatewayService{
		cfg:                  cfg,
		httpUpstream:         upstream,
		responseHeaderFilter: compileResponseHeaderFilter(cfg),
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

	result, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.Equal(t, "thinking", gjson.GetBytes(rec.Body.Bytes(), "content.0.type").String())
	require.Equal(t, "plan first", gjson.GetBytes(rec.Body.Bytes(), "content.0.thinking").String())
	require.Equal(t, "text", gjson.GetBytes(rec.Body.Bytes(), "content.1.type").String())
	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "content.1.text").String())
	require.Equal(t, int64(11), gjson.GetBytes(rec.Body.Bytes(), "usage.input_tokens").Int())
	require.Equal(t, int64(7), gjson.GetBytes(rec.Body.Bytes(), "usage.output_tokens").Int())
}

func TestForwardAsAnthropic_APIKeyAddsDefaultInstructionsWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"claude-opus-4-1","max_tokens":1024,"messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.completed","response":{"id":"resp_messages_default_instructions","model":"gpt-5.5","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"final answer"}]}],"usage":{"input_tokens":1,"output_tokens":2}}}`,
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-messages-default-instructions"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}
	upstream := &httpUpstreamRecorder{resp: resp}

	cfg := &config.Config{Gateway: config.GatewayConfig{}}
	svc := &OpenAIGatewayService{
		cfg:                  cfg,
		httpUpstream:         upstream,
		responseHeaderFilter: compileResponseHeaderFilter(cfg),
	}

	account := &Account{
		ID:             123,
		Name:           "acc",
		Platform:       PlatformOpenAI,
		Type:           AccountTypeAPIKey,
		Concurrency:    1,
		Credentials:    map[string]any{"api_key": "test-api-key"},
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
	}

	result, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "gpt-5.5")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, openAIResponsesDefaultInstructions, gjson.GetBytes(upstream.lastBody, "instructions").String())
	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "content.0.text").String())
}

func TestForwardAsAnthropic_APIKeyStripsConvertedMaxOutputTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5.4","max_tokens":4096,"max_outputs_tokens":4096,"messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.completed","response":{"id":"resp_messages_no_max_output","model":"gpt-5.4","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"final answer"}]}],"usage":{"input_tokens":1,"output_tokens":2}}}`,
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-messages-no-max-output"}},
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
		Type:           AccountTypeAPIKey,
		Concurrency:    1,
		Credentials:    map[string]any{"api_key": "test-api-key"},
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
	}

	result, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "gpt-5.4")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, gjson.GetBytes(upstream.lastBody, "max_output_tokens").Exists(), "API key messages compatibility path must not forward max_output_tokens")
	require.False(t, gjson.GetBytes(upstream.lastBody, "max_outputs_tokens").Exists(), "API key messages compatibility path must not forward misspelled max_outputs_tokens")
	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "content.0.text").String())
}

func TestForwardAsAnthropic_BufferedToolCallDeltaOnlyTerminalResponsePreservesToolUse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"claude-opus-4-1","max_tokens":1024,"messages":[{"role":"user","content":"Call the function"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_messages_tool","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","call_id":"call_1","name":"get_weather"}}`,
		"",
		`data: {"type":"response.function_call_arguments.delta","output_index":0,"delta":"{\"city\":\"Ber"}`,
		"",
		`data: {"type":"response.function_call_arguments.delta","output_index":0,"delta":"lin\"}"}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_messages_tool","status":"completed","usage":{"input_tokens":5,"output_tokens":3}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-messages-tool"}},
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

	result, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "tool_use", gjson.GetBytes(rec.Body.Bytes(), "content.0.type").String())
	require.Equal(t, "call_1", gjson.GetBytes(rec.Body.Bytes(), "content.0.id").String())
	require.Equal(t, "get_weather", gjson.GetBytes(rec.Body.Bytes(), "content.0.name").String())
	require.Equal(t, "Berlin", gjson.GetBytes(rec.Body.Bytes(), "content.0.input.city").String())
	require.Equal(t, "tool_use", gjson.GetBytes(rec.Body.Bytes(), "stop_reason").String())
}

func TestForwardAsAnthropic_BufferedDoneTerminalResponsePreservesUsageInResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"claude-opus-4-1","max_tokens":1024,"messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_messages_done","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"final answer"}`,
		"",
		`data: {"type":"response.done","response":{"id":"resp_messages_done","model":"gpt-5.1","status":"completed","usage":{"input_tokens":10,"output_tokens":4,"input_tokens_details":{"cached_tokens":1}}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-messages-done"}},
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

	result, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "content.0.text").String())
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, 4, result.Usage.OutputTokens)
	require.Equal(t, 1, result.Usage.CacheReadInputTokens)
}

func TestForwardAsAnthropic_BufferedDoneTerminalResponsePreservesOptionalCacheCreationUsageInResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"claude-opus-4-1","max_tokens":1024,"messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_messages_cache_create","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"final answer"}`,
		"",
		`data: {"type":"response.done","response":{"id":"resp_messages_cache_create","model":"gpt-5.1","status":"completed","usage":{"input_tokens":10,"output_tokens":4,"cache_creation_input_tokens":7,"input_tokens_details":{"cached_tokens":1}}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-messages-cache-create"}},
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

	result, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "content.0.text").String())
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, 4, result.Usage.OutputTokens)
	require.Equal(t, 7, result.Usage.CacheCreationInputTokens)
	require.Equal(t, 1, result.Usage.CacheReadInputTokens)
}
