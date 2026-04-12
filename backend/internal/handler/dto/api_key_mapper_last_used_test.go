package dto

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func TestAPIKeyFromService_MapsLastUsedAt(t *testing.T) {
	lastUsed := time.Now().UTC().Truncate(time.Second)
	src := &service.APIKey{
		ID:         1,
		UserID:     2,
		Key:        "sk-map-last-used",
		Name:       "Mapper",
		Status:     service.StatusActive,
		LastUsedAt: &lastUsed,
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.NotNil(t, out.LastUsedAt)
	require.WithinDuration(t, lastUsed, *out.LastUsedAt, time.Second)
}

func TestAPIKeyFromService_MapsNilLastUsedAt(t *testing.T) {
	src := &service.APIKey{
		ID:     1,
		UserID: 2,
		Key:    "sk-map-last-used-nil",
		Name:   "MapperNil",
		Status: service.StatusActive,
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.Nil(t, out.LastUsedAt)
}

func TestAPIKeyFromService_MapsBudgetMultiplier(t *testing.T) {
	src := &service.APIKey{
		ID:               1,
		UserID:           2,
		Key:              "sk-budget-mapper",
		Name:             "BudgetMapper",
		Status:           service.StatusActive,
		BudgetMultiplier: float64Ptr(8.5),
		Group: &service.Group{
			ID:                      11,
			Name:                    "claude-dynamic",
			PricingMode:             service.GroupPricingModeDynamic,
			DefaultBudgetMultiplier: float64Ptr(8.0),
		},
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.NotNil(t, out.BudgetMultiplier)
	require.InDelta(t, 8.5, *out.BudgetMultiplier, 0.0001)
	require.NotNil(t, out.Group)
	require.Equal(t, service.GroupPricingModeDynamic, out.Group.PricingMode)
	require.NotNil(t, out.Group.DefaultBudgetMultiplier)
	require.InDelta(t, 8.0, *out.Group.DefaultBudgetMultiplier, 0.0001)
}
