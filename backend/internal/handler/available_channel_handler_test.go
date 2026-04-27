package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAvailableChannelFeatureDisabledByDefaultWithoutSettingService(t *testing.T) {
	h := &AvailableChannelHandler{settingService: nil}
	require.False(t, h.featureEnabled(nil))
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
