package dto

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupFromServiceAdminIncludesDynamicPricingFields(t *testing.T) {
	t.Parallel()

	defaultBudgetMultiplier := 2.4
	group := &service.Group{
		ID:                      42,
		Name:                    "dynamic-openai",
		Platform:                service.PlatformOpenAI,
		Status:                  service.StatusActive,
		PricingMode:             service.GroupPricingModeDynamic,
		DefaultBudgetMultiplier: &defaultBudgetMultiplier,
	}

	out := GroupFromServiceAdmin(group)

	require.NotNil(t, out)
	require.Equal(t, service.GroupPricingModeDynamic, out.PricingMode)
	require.NotNil(t, out.DefaultBudgetMultiplier)
	require.InDelta(t, defaultBudgetMultiplier, *out.DefaultBudgetMultiplier, 1e-12)

	payload, err := json.Marshal(out)
	require.NoError(t, err)
	require.Contains(t, string(payload), `"pricing_mode":"dynamic"`)
	require.Contains(t, string(payload), `"default_budget_multiplier":2.4`)
}
