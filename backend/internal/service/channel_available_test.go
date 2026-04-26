package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelSupportedModelsIncludesPricingWithoutMapping(t *testing.T) {
	price := 0.000001
	ch := &Channel{
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-opus-*": "claude-opus-*",
			},
		},
		ModelPricing: []ChannelModelPricing{
			{
				Platform:   PlatformAnthropic,
				Models:     []string{"claude-opus-4-6", "claude-sonnet-4-6"},
				InputPrice: &price,
			},
		},
	}

	models := ch.SupportedModels()
	require.Len(t, models, 2)
	require.Equal(t, "claude-opus-4-6", models[0].Name)
	require.Equal(t, "claude-sonnet-4-6", models[1].Name)
	require.NotNil(t, models[1].Pricing, "pricing-only models must be displayed with their pricing")
}
