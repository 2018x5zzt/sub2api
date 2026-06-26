package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePricingData_ParsesPriorityAndServiceTierFields(t *testing.T) {
	svc := &PricingService{}
	body := []byte(`{
		"gpt-5.4": {
			"input_cost_per_token": 0.0000025,
			"input_cost_per_token_priority": 0.000005,
			"output_cost_per_token": 0.000015,
			"output_cost_per_token_priority": 0.00003,
			"cache_creation_input_token_cost": 0.0000025,
			"cache_read_input_token_cost": 0.00000025,
			"cache_read_input_token_cost_priority": 0.0000005,
			"supports_service_tier": true,
			"supports_prompt_caching": true,
			"litellm_provider": "openai",
			"mode": "chat"
		}
	}`)

	data, err := svc.parsePricingData(body)
	require.NoError(t, err)
	pricing := data["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 5e-6, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 3e-5, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 5e-7, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

func TestGetModelPricing_Gpt52HighUsesGpt52Pricing(t *testing.T) {
	gpt52Pricing := &LiteLLMModelPricing{InputCostPerToken: 1}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.2": gpt52Pricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.2-high")
	require.Same(t, gpt52Pricing, got)
}

func TestGetModelPricing_Gpt53CodexFallbackStillUsesGpt52Codex(t *testing.T) {
	gpt52CodexPricing := &LiteLLMModelPricing{InputCostPerToken: 2}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.2-codex": gpt52CodexPricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex")
	require.Same(t, gpt52CodexPricing, got)
}

func TestGetModelPricing_Gpt5CodexMiniFallsBackToGpt51CodexMini(t *testing.T) {
	gpt51CodexMiniPricing := &LiteLLMModelPricing{InputCostPerToken: 0.25e-6}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex-mini": gpt51CodexMiniPricing,
		},
	}

	got := svc.GetModelPricing("gpt-5-codex-mini")
	require.Same(t, gpt51CodexMiniPricing, got)
}

func TestGetModelPricing_OpenAIFallbackMatchedLoggedAsInfo(t *testing.T) {
	logSink, restore := captureStructuredLog(t)
	defer restore()

	gpt52CodexPricing := &LiteLLMModelPricing{InputCostPerToken: 2}
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.2-codex": gpt52CodexPricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex-high")
	require.Same(t, gpt52CodexPricing, got)

	require.True(t, logSink.ContainsMessageAtLevel("[Pricing] OpenAI fallback matched gpt-5.3-codex-high -> gpt-5.2-codex", "info"))
	require.False(t, logSink.ContainsMessageAtLevel("[Pricing] OpenAI fallback matched gpt-5.3-codex-high -> gpt-5.2-codex", "warn"))
}

func TestGetModelPricing_Gpt54UsesStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": &LiteLLMModelPricing{InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4")
	require.NotNil(t, got)
	require.InDelta(t, 2.5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 2.5e-7, got.CacheReadInputTokenCost, 1e-12)
	require.Equal(t, 272000, got.LongContextInputTokenThreshold)
	require.InDelta(t, 2.0, got.LongContextInputCostMultiplier, 1e-12)
	require.InDelta(t, 1.5, got.LongContextOutputCostMultiplier, 1e-12)
}

func TestGetModelPricing_OpenAICompactAliasUsesStaticFallback(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("openai/gpt5.5")
	require.NotNil(t, got)
	require.InDelta(t, 2.5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, got.OutputCostPerToken, 1e-12)
}

func TestGetModelPricing_Gpt54MiniUsesDedicatedStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4-mini")
	require.NotNil(t, got)
	require.InDelta(t, 7.5e-7, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 4.5e-6, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 7.5e-8, got.CacheReadInputTokenCost, 1e-12)
	require.Zero(t, got.LongContextInputTokenThreshold)
}

func TestGetModelPricing_Gpt54NanoUsesDedicatedStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4-nano")
	require.NotNil(t, got)
	require.InDelta(t, 2e-7, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.25e-6, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 2e-8, got.CacheReadInputTokenCost, 1e-12)
	require.Zero(t, got.LongContextInputTokenThreshold)
}

func TestGetModelPricing_UnsupportedOpenAIModelDoesNotFallBackToDifferentFamily(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5":   &LiteLLMModelPricing{InputCostPerToken: 1.25e-6},
			"gpt-5.4": &LiteLLMModelPricing{InputCostPerToken: 2.5e-6},
		},
	}

	// 吸收上游 0.1.115 codex 归一化后，gpt-5-pro 会被路由/归一化到 gpt-5.4，
	// 因此按 gpt-5.4 计费与实际上游用量一致。（PHASE2 吸收 xlabapi 的混合归一化时
	// 会恢复“未知模型不静默 remap”的保守策略，届时此断言回归 nil。）
	got := svc.GetModelPricing("gpt-5-pro")
	require.NotNil(t, got)
	require.Equal(t, 2.5e-6, got.InputCostPerToken)
}

func TestGetModelPricing_ImageModelDoesNotFallbackToTextModel(t *testing.T) {
	imagePricing := &LiteLLMModelPricing{InputCostPerToken: 3}
	textPricing := &LiteLLMModelPricing{InputCostPerToken: 9}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-image-2": imagePricing,
			"gpt-5.4":     textPricing,
		},
	}

	got := svc.GetModelPricing("gpt-image-3")
	require.Same(t, imagePricing, got)
}

func TestParsePricingData_PreservesPriorityAndServiceTierFields(t *testing.T) {
	raw := map[string]any{
		"gpt-5.4": map[string]any{
			"input_cost_per_token":                 2.5e-6,
			"input_cost_per_token_priority":        5e-6,
			"output_cost_per_token":                15e-6,
			"output_cost_per_token_priority":       30e-6,
			"cache_read_input_token_cost":          0.25e-6,
			"cache_read_input_token_cost_priority": 0.5e-6,
			"supports_service_tier":                true,
			"supports_prompt_caching":              true,
			"litellm_provider":                     "openai",
			"mode":                                 "chat",
		},
	}
	body, err := json.Marshal(raw)
	require.NoError(t, err)

	svc := &PricingService{}
	pricingMap, err := svc.parsePricingData(body)
	require.NoError(t, err)

	pricing := pricingMap["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 2.5e-6, pricing.InputCostPerToken, 1e-12)
	require.InDelta(t, 5e-6, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 15e-6, pricing.OutputCostPerToken, 1e-12)
	require.InDelta(t, 30e-6, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.25e-6, pricing.CacheReadInputTokenCost, 1e-12)
	require.InDelta(t, 0.5e-6, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

func TestParsePricingData_PreservesServiceTierPriorityFields(t *testing.T) {
	svc := &PricingService{}
	pricingData, err := svc.parsePricingData([]byte(`{
		"gpt-5.4": {
			"input_cost_per_token": 0.0000025,
			"input_cost_per_token_priority": 0.000005,
			"output_cost_per_token": 0.000015,
			"output_cost_per_token_priority": 0.00003,
			"cache_read_input_token_cost": 0.00000025,
			"cache_read_input_token_cost_priority": 0.0000005,
			"supports_service_tier": true,
			"litellm_provider": "openai",
			"mode": "chat"
		}
	}`))
	require.NoError(t, err)

	pricing := pricingData["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 0.0000025, pricing.InputCostPerToken, 1e-12)
	require.InDelta(t, 0.000005, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.000015, pricing.OutputCostPerToken, 1e-12)
	require.InDelta(t, 0.00003, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.00000025, pricing.CacheReadInputTokenCost, 1e-12)
	require.InDelta(t, 0.0000005, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}
