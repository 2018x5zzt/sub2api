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

func TestForwardAsChatCompletions_BufferedDeltaOnlyTerminalResponsePreservesContentAndReasoning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5.1","messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_buffered_gap","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.reasoning_summary_text.delta","output_index":0,"summary_index":0,"delta":"plan first"}`,
		"",
		`data: {"type":"response.output_text.delta","output_index":1,"content_index":0,"delta":"final answer"}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_buffered_gap","status":"completed","usage":{"input_tokens":11,"output_tokens":7,"total_tokens":18}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-buffered-gap"}},
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

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "plan first", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.reasoning_content").String())
	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.content").String())
	require.Equal(t, int64(11), gjson.GetBytes(rec.Body.Bytes(), "usage.prompt_tokens").Int())
	require.Equal(t, int64(7), gjson.GetBytes(rec.Body.Bytes(), "usage.completion_tokens").Int())
}

func TestForwardAsChatCompletions_BufferedMessageItemAndTextDoneDoNotDuplicateContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5.4","messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_dup_text","model":"gpt-5.4","status":"in_progress"}}`,
		"",
		`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"final answer"}]}}`,
		"",
		`data: {"type":"response.output_text.done","output_index":0,"content_index":0,"text":"final answer"}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_dup_text","status":"completed","usage":{"input_tokens":4,"output_tokens":2,"total_tokens":6}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-dup-text"}},
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

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.content").String())
}

func TestForwardAsChatCompletions_BufferedToolCallDeltaOnlyTerminalResponsePreservesToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5.1","messages":[{"role":"user","content":"Call the function"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_tool_gap","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","call_id":"call_1","name":"get_weather"}}`,
		"",
		`data: {"type":"response.function_call_arguments.delta","output_index":0,"delta":"{\"city\":\"Ber"}`,
		"",
		`data: {"type":"response.function_call_arguments.delta","output_index":0,"delta":"lin\"}"}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_tool_gap","status":"completed","usage":{"input_tokens":5,"output_tokens":3,"total_tokens":8}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-tool-gap"}},
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

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "tool_calls", gjson.GetBytes(rec.Body.Bytes(), "choices.0.finish_reason").String())
	require.Equal(t, "call_1", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.tool_calls.0.id").String())
	require.Equal(t, "get_weather", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.tool_calls.0.function.name").String())
	require.Equal(t, `{"city":"Berlin"}`, gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.tool_calls.0.function.arguments").String())
}

func TestForwardAsChatCompletions_BufferedDoneTerminalResponsePreservesUsageInResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5.1","messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_done_usage","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"final answer"}`,
		"",
		`data: {"type":"response.done","response":{"id":"resp_done_usage","model":"gpt-5.1","status":"completed","usage":{"input_tokens":12,"output_tokens":6,"input_tokens_details":{"cached_tokens":2}}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-done-usage"}},
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

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.content").String())
	require.Equal(t, 12, result.Usage.InputTokens)
	require.Equal(t, 6, result.Usage.OutputTokens)
	require.Equal(t, 2, result.Usage.CacheReadInputTokens)
}

func TestForwardAsChatCompletions_BufferedDoneTerminalResponsePreservesOptionalCacheCreationUsageInResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5.1","messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_done_cache_create","model":"gpt-5.1","status":"in_progress"}}`,
		"",
		`data: {"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"final answer"}`,
		"",
		`data: {"type":"response.done","response":{"id":"resp_done_cache_create","model":"gpt-5.1","status":"completed","usage":{"input_tokens":12,"output_tokens":6,"cache_creation_input_tokens":5,"input_tokens_details":{"cached_tokens":2}}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-done-cache-create"}},
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

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.content").String())
	require.Equal(t, 12, result.Usage.InputTokens)
	require.Equal(t, 6, result.Usage.OutputTokens)
	require.Equal(t, 5, result.Usage.CacheCreationInputTokens)
	require.Equal(t, 2, result.Usage.CacheReadInputTokens)
}
