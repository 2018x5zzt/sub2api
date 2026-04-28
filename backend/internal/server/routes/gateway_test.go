package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newGatewayRoutesTestRouter() *gin.Engine {
	return newGatewayRoutesTestRouterWithGroupID(service.PlatformOpenAI, "【限时半价】gpt-image", 30)
}

func newGatewayRoutesTestRouterWithGroup(platform, groupName string) *gin.Engine {
	return newGatewayRoutesTestRouterWithGroupID(platform, groupName, 1)
}

func newGatewayRoutesTestRouterWithGroupID(platform, groupName string, groupID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	RegisterGatewayRoutes(
		router,
		&handler.Handlers{
			Gateway:       &handler.GatewayHandler{},
			OpenAIGateway: &handler.OpenAIGatewayHandler{},
			SoraGateway:   &handler.SoraGatewayHandler{},
		},
		servermiddleware.APIKeyAuthMiddleware(func(c *gin.Context) {
			c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
				GroupID: &groupID,
				Group: &service.Group{
					ID:       groupID,
					Name:     groupName,
					Platform: platform,
				},
			})
			c.Next()
		}),
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
	)

	return router
}

func TestGatewayRoutesOpenAIResponsesCompactPathIsRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	for _, path := range []string{"/v1/responses/compact", "/responses/compact"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-5"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit OpenAI responses handler", path)
	}
}

func newGatewayRoutesTestRouterWithGroupPlatform(platform string) *gin.Engine {
	return newGatewayRoutesTestRouterWithGroup(platform, "")
}

func parseErrorObject(t *testing.T, body string) map[string]any {
	t.Helper()
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	errObj, ok := payload["error"].(map[string]any)
	require.True(t, ok, "response should contain error object: %s", body)
	return errObj
}

func TestGatewayRoutesV1ResponsesDispatchesByGroupPlatform(t *testing.T) {
	testCases := []struct {
		name          string
		platform      string
		wantStatus    int
		wantFieldName string
		wantFieldVal  string
	}{
		{
			name:          "openai platform routes to openai handler envelope",
			platform:      service.PlatformOpenAI,
			wantStatus:    http.StatusInternalServerError,
			wantFieldName: "type",
			wantFieldVal:  "api_error",
		},
		{
			name:          "anthropic platform routes to gateway handler envelope",
			platform:      service.PlatformAnthropic,
			wantStatus:    http.StatusInternalServerError,
			wantFieldName: "code",
			wantFieldVal:  "api_error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := newGatewayRoutesTestRouterWithGroupPlatform(tc.platform)

			req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"gpt-5"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			require.Equal(t, tc.wantStatus, w.Code)

			errObj := parseErrorObject(t, w.Body.String())
			require.Equal(t, tc.wantFieldVal, errObj[tc.wantFieldName])
		})
	}
}

func TestGatewayRoutesV1ChatCompletionsRouteReturnsGatewayErrorEnvelope(t *testing.T) {
	router := newGatewayRoutesTestRouterWithGroupPlatform(service.PlatformAnthropic)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-5"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	errObj := parseErrorObject(t, w.Body.String())
	require.Equal(t, "api_error", errObj["type"])
	require.Equal(t, "User context not found", errObj["message"])
}

func TestGatewayRoutesOpenAIImagesPathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	for _, path := range []string{
		"/v1/images/generations",
		"/v1/images/edits",
		"/images/generations",
		"/images/edits",
		"/api-proxy/v1/images/generations",
		"/api-proxy/v1/images/edits",
		"/openai/v1/images/generations",
		"/openai/v1/images/edits",
	} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-image-2","prompt":"draw a cat"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit OpenAI images handler", path)
	}
}

func TestGatewayRoutesOpenAIImagesAllowSubscriptionImageGroup(t *testing.T) {
	router := newGatewayRoutesTestRouterWithGroupID(service.PlatformOpenAI, "【订阅】gpt-image", 35)

	req := httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"model":"gpt-image-2","prompt":"draw a cat"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestGatewayRoutesOpenAIImagesRejectNonOpenAIPlatforms(t *testing.T) {
	router := newGatewayRoutesTestRouterWithGroupPlatform(service.PlatformAnthropic)

	for _, path := range []string{"/v1/images/generations", "/images/edits"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-image-2","prompt":"draw a cat"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code, "path=%s should reject non-openai groups", path)

		errObj := parseErrorObject(t, w.Body.String())
		require.Equal(t, "not_found_error", errObj["type"])
		require.Equal(t, "Images API is not supported for this platform", errObj["message"])
	}
}

func TestGatewayRoutesOpenAIImagesRejectNonGPTImageOpenAIGroups(t *testing.T) {
	router := newGatewayRoutesTestRouterWithGroup(service.PlatformOpenAI, "pro号池")

	for _, path := range []string{"/v1/images/generations", "/images/edits"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-image-2","prompt":"draw a cat"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code, "path=%s should reject non-gpt-image openai groups", path)

		errObj := parseErrorObject(t, w.Body.String())
		require.Equal(t, "not_found_error", errObj["type"])
		require.Equal(t, "Images API is not enabled for this group", errObj["message"])
	}
}
