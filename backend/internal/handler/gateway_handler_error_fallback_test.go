package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatewayEnsureForwardErrorResponse_WritesFallbackWhenNotWritten(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	h := &GatewayHandler{}
	wrote := h.ensureForwardErrorResponse(c, false)

	require.True(t, wrote)
	require.Equal(t, http.StatusBadGateway, w.Code)

	var parsed map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "error", parsed["type"])
	errorObj, ok := parsed["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "upstream_error", errorObj["type"])
	assert.Equal(t, "Upstream request failed", errorObj["message"])
}

func TestGatewayEnsureForwardErrorResponse_UsesUpstreamBadRequestMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	service.SetOpsUpstreamError(c, http.StatusBadRequest, "暂不支持", "")

	h := &GatewayHandler{}
	wrote := h.ensureForwardErrorResponse(c, false)

	require.True(t, wrote)
	require.Equal(t, http.StatusBadRequest, w.Code)

	var parsed map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "error", parsed["type"])
	errorObj, ok := parsed["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "invalid_request_error", errorObj["type"])
	assert.Equal(t, "暂不支持", errorObj["message"])
}

func TestGatewayEnsureForwardErrorResponse_DoesNotOverrideWrittenResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.String(http.StatusTeapot, "already written")

	h := &GatewayHandler{}
	wrote := h.ensureForwardErrorResponse(c, false)

	require.False(t, wrote)
	require.Equal(t, http.StatusTeapot, w.Code)
	assert.Equal(t, "already written", w.Body.String())
}

func TestGatewayHandleFailoverExhausted_UsesUpstreamBadRequestMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	h := &GatewayHandler{}
	failoverErr := &service.UpstreamFailoverError{
		StatusCode: http.StatusBadRequest,
		ResponseBody: []byte(`{
			"error": {
				"type": "invalid_request_error",
				"message": "暂不支持"
			}
		}`),
	}

	h.handleFailoverExhausted(c, failoverErr, service.PlatformAnthropic, false)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var parsed map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "error", parsed["type"])
	errorObj, ok := parsed["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "invalid_request_error", errorObj["type"])
	assert.Equal(t, "暂不支持", errorObj["message"])
}
