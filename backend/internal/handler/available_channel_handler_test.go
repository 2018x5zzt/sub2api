package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAvailableChannelList_FeatureDisabledByDefaultWithoutSettingServiceReturnsEmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/available-channels", nil)
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})

	h := &AvailableChannelHandler{
		channelService: nil,
		apiKeyService:  nil,
		settingService: nil,
	}

	h.List(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var envelope struct {
		Code int                    `json:"code"`
		Data []userAvailableChannel `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)
	require.Empty(t, envelope.Data)
}

func TestBuildPlatformSections_OnlyReturnsModelsForVisiblePlatforms(t *testing.T) {
	price := 0.000001
	ch := service.AvailableChannel{
		Name: "official",
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: service.PlatformAnthropic, Pricing: &service.ChannelModelPricing{InputPrice: &price}},
			{Name: "gpt-5.4", Platform: service.PlatformOpenAI, Pricing: &service.ChannelModelPricing{InputPrice: &price}},
		},
	}
	visibleGroups := []userAvailableGroup{
		{ID: 1, Name: "anthropic-sub", Platform: service.PlatformAnthropic},
	}

	sections := buildPlatformSections(ch, visibleGroups)
	require.Len(t, sections, 1)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Len(t, sections[0].SupportedModels, 1)
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
}
