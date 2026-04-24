package handler

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type groupModelAccountRepoStub struct {
	accountsByGroup map[int64][]service.Account
}

func (r *groupModelAccountRepoStub) Create(context.Context, *service.Account) error { return nil }
func (r *groupModelAccountRepoStub) GetByID(context.Context, int64) (*service.Account, error) {
	return nil, service.ErrAccountNotFound
}
func (r *groupModelAccountRepoStub) GetByIDs(context.Context, []int64) ([]*service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ExistsByID(context.Context, int64) (bool, error) {
	return false, nil
}
func (r *groupModelAccountRepoStub) GetByCRSAccountID(context.Context, string) (*service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) FindByExtraField(context.Context, string, any) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListCRSAccountIDs(context.Context) (map[string]int64, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) Update(context.Context, *service.Account) error { return nil }
func (r *groupModelAccountRepoStub) Delete(context.Context, int64) error            { return nil }
func (r *groupModelAccountRepoStub) List(context.Context, pagination.PaginationParams) ([]service.Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *groupModelAccountRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, string, int64, string) ([]service.Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *groupModelAccountRepoStub) ListByGroup(_ context.Context, groupID int64) ([]service.Account, error) {
	return append([]service.Account(nil), r.accountsByGroup[groupID]...), nil
}
func (r *groupModelAccountRepoStub) ListActive(context.Context) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListByPlatform(context.Context, string) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) UpdateLastUsed(context.Context, int64) error { return nil }
func (r *groupModelAccountRepoStub) BatchUpdateLastUsed(context.Context, map[int64]time.Time) error {
	return nil
}
func (r *groupModelAccountRepoStub) SetError(context.Context, int64, string) error { return nil }
func (r *groupModelAccountRepoStub) ClearError(context.Context, int64) error       { return nil }
func (r *groupModelAccountRepoStub) SetSchedulable(context.Context, int64, bool) error {
	return nil
}
func (r *groupModelAccountRepoStub) AutoPauseExpiredAccounts(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (r *groupModelAccountRepoStub) BindGroups(context.Context, int64, []int64) error { return nil }
func (r *groupModelAccountRepoStub) ListSchedulable(context.Context) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListSchedulableByGroupID(context.Context, int64) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListSchedulableByPlatform(context.Context, string) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListSchedulableByGroupIDAndPlatform(context.Context, int64, string) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListSchedulableByPlatforms(context.Context, []string) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListSchedulableByGroupIDAndPlatforms(context.Context, int64, []string) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListSchedulableUngroupedByPlatform(context.Context, string) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) ListSchedulableUngroupedByPlatforms(context.Context, []string) ([]service.Account, error) {
	return nil, nil
}
func (r *groupModelAccountRepoStub) SetRateLimited(context.Context, int64, time.Time) error {
	return nil
}
func (r *groupModelAccountRepoStub) SetModelRateLimit(context.Context, int64, string, time.Time) error {
	return nil
}
func (r *groupModelAccountRepoStub) SetOverloaded(context.Context, int64, time.Time) error {
	return nil
}
func (r *groupModelAccountRepoStub) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	return nil
}
func (r *groupModelAccountRepoStub) ClearTempUnschedulable(context.Context, int64) error {
	return nil
}
func (r *groupModelAccountRepoStub) ClearRateLimit(context.Context, int64) error { return nil }
func (r *groupModelAccountRepoStub) ClearAntigravityQuotaScopes(context.Context, int64) error {
	return nil
}
func (r *groupModelAccountRepoStub) ClearModelRateLimits(context.Context, int64) error {
	return nil
}
func (r *groupModelAccountRepoStub) UpdateSessionWindow(context.Context, int64, *time.Time, *time.Time, string) error {
	return nil
}
func (r *groupModelAccountRepoStub) UpdateExtra(context.Context, int64, map[string]any) error {
	return nil
}
func (r *groupModelAccountRepoStub) BulkUpdate(context.Context, []int64, service.AccountBulkUpdate) (int64, error) {
	return 0, nil
}
func (r *groupModelAccountRepoStub) IncrementQuotaUsed(context.Context, int64, float64) error {
	return nil
}
func (r *groupModelAccountRepoStub) ResetQuotaUsed(context.Context, int64) error { return nil }

func TestCollectGroupModelIDs_IgnoresWildcardMappingSelectors(t *testing.T) {
	accounts := []service.Account{
		{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-*": "gpt-5.4",
				},
			},
		},
	}

	modelIDs := collectGroupModelIDs(service.PlatformOpenAI, accounts)
	var want []string
	if !reflect.DeepEqual(modelIDs, want) {
		t.Fatalf("collectGroupModelIDs() = %v, want %v", modelIDs, want)
	}
}

func TestCollectGroupModelIDs_OpenAIPassthroughIgnoresStaleMappings(t *testing.T) {
	accounts := []service.Account{
		{
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeOAuth,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"legacy-alias": "gpt-5.4",
				},
			},
			Extra: map[string]any{
				"openai_passthrough": true,
			},
		},
	}

	modelIDs := collectGroupModelIDs(service.PlatformOpenAI, accounts)
	if len(modelIDs) != 0 {
		t.Fatalf("expected passthrough account mappings to be ignored, got %v", modelIDs)
	}
}

func TestCollectGroupModelIDs_OnlyKeepsConcreteModelsFromExplicitAPIKeyMappings(t *testing.T) {
	accounts := []service.Account{
		{
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeOAuth,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-opus-ignored": "upstream",
				},
			},
		},
		{
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeAPIKey,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-*":              "wildcard-ignored",
					"claude-preview-custom": "upstream-preview",
				},
			},
		},
	}

	modelIDs := collectGroupModelIDs(service.PlatformAnthropic, accounts)
	want := []string{"claude-preview-custom"}
	if !reflect.DeepEqual(modelIDs, want) {
		t.Fatalf("collectGroupModelIDs() = %v, want %v", modelIDs, want)
	}
}

func TestFilterMappedModelIDsByCatalog_OpenAIOnlyKeepsCatalogModels(t *testing.T) {
	modelIDs := []string{"gpt-5.4", "legacy-gpt-5-preview"}

	got := filterMappedModelIDsByCatalog(service.PlatformOpenAI, modelIDs)
	want := []string{"gpt-5.4"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterMappedModelIDsByCatalog() = %v, want %v", got, want)
	}
}

func TestGetGroupSupportedModels_UsesStaticOpenAICatalog(t *testing.T) {
	handler := &APIKeyHandler{}

	models, source, err := handler.getGroupSupportedModels(context.Background(), &service.Group{
		Platform: service.PlatformOpenAI,
	})
	if err != nil {
		t.Fatalf("getGroupSupportedModels() error = %v", err)
	}
	if source != "default" {
		t.Fatalf("getGroupSupportedModels() source = %q, want %q", source, "default")
	}
	if len(models) == 0 {
		t.Fatal("expected static OpenAI model catalog to be non-empty")
	}

	found := false
	for _, model := range models {
		if model.ID == "gpt-5" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected gpt-5 to be present in static catalog, got %v", models)
	}
}

func TestGetGroupSupportedModels_HidesImageModelsForNonGPTImageGroup(t *testing.T) {
	handler := &APIKeyHandler{
		accountRepo: &groupModelAccountRepoStub{
			accountsByGroup: map[int64][]service.Account{
				29: {
					{
						Platform: service.PlatformOpenAI,
						Type:     service.AccountTypeAPIKey,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5":       "gpt-5",
								"gpt-image-2": "gpt-image-2",
							},
						},
					},
				},
			},
		},
	}

	models, source, err := handler.getGroupSupportedModels(context.Background(), &service.Group{
		ID:       29,
		Name:     "pro号池",
		Platform: service.PlatformOpenAI,
	})
	require.NoError(t, err)
	require.Equal(t, "mixed", source)

	modelIDs := make([]string, 0, len(models))
	for _, model := range models {
		modelIDs = append(modelIDs, model.ID)
	}

	require.Contains(t, modelIDs, "gpt-5")
	require.NotContains(t, modelIDs, "gpt-image-2")
}

func TestGetGroupSupportedModels_KeepsImageModelsForGPTImageGroup(t *testing.T) {
	handler := &APIKeyHandler{}

	models, source, err := handler.getGroupSupportedModels(context.Background(), &service.Group{
		ID:       30,
		Name:     "gpt-image",
		Platform: service.PlatformOpenAI,
	})
	require.NoError(t, err)
	require.Equal(t, "default", source)

	modelIDs := make([]string, 0, len(models))
	for _, model := range models {
		modelIDs = append(modelIDs, model.ID)
	}

	require.Contains(t, modelIDs, "gpt-image-2")
	require.NotContains(t, modelIDs, "gpt-5")
	for _, modelID := range modelIDs {
		require.True(t, service.IsOpenAIImageGenerationModel(modelID), "unexpected non-image model %s", modelID)
	}
}

func TestGetGroupSupportedModels_AnthropicMergesDefaultsWithExplicitMappedCustomModels(t *testing.T) {
	handler := &APIKeyHandler{
		accountRepo: &groupModelAccountRepoStub{
			accountsByGroup: map[int64][]service.Account{
				42: {
					{
						Platform: service.PlatformAnthropic,
						Type:     service.AccountTypeAPIKey,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"claude-opus-4-7":   "upstream-opus-4-7",
								"claude-sonnet-4-6": "claude-sonnet-4-6",
								"claude-*":          "wildcard-ignored",
							},
						},
					},
					{
						Platform: service.PlatformAnthropic,
						Type:     service.AccountTypeOAuth,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"claude-oauth-shadow": "should-not-surface",
							},
						},
					},
				},
			},
		},
	}

	models, source, err := handler.getGroupSupportedModels(context.Background(), &service.Group{
		ID:       42,
		Platform: service.PlatformAnthropic,
	})
	require.NoError(t, err)
	require.Equal(t, "mixed", source)

	modelIDs := make([]string, 0, len(models))
	for _, model := range models {
		modelIDs = append(modelIDs, model.ID)
	}

	require.Contains(t, modelIDs, "claude-haiku-4-5-20251001")
	require.Contains(t, modelIDs, "claude-sonnet-4-6")
	require.Contains(t, modelIDs, "claude-opus-4-6")
	require.Contains(t, modelIDs, "claude-opus-4-7")
	require.NotContains(t, modelIDs, "claude-*")
	require.NotContains(t, modelIDs, "claude-oauth-shadow")
}

func TestGetGroupSupportedModels_AntigravityScopesStillApply(t *testing.T) {
	handler := &APIKeyHandler{}

	models, source, err := handler.getGroupSupportedModels(context.Background(), &service.Group{
		Platform:             service.PlatformAntigravity,
		SupportedModelScopes: []string{"claude"},
	})
	if err != nil {
		t.Fatalf("getGroupSupportedModels() error = %v", err)
	}
	if source != "default" {
		t.Fatalf("getGroupSupportedModels() source = %q, want %q", source, "default")
	}
	if len(models) == 0 {
		t.Fatal("expected scoped antigravity model catalog to be non-empty")
	}
	for _, model := range models {
		if strings.Contains(strings.ToLower(model.ID), "gemini") {
			t.Fatalf("expected gemini models to be filtered out by scopes, got %v", models)
		}
	}
}

func TestResolveGroupRateMultiplier_UsesUserOverride(t *testing.T) {
	group := &service.Group{
		ID:             42,
		RateMultiplier: 1.8,
	}

	effective, userRate := resolveGroupRateMultiplier(group, map[int64]float64{
		42: 1.25,
	})

	require.InDelta(t, 1.25, effective, 1e-12)
	require.NotNil(t, userRate)
	require.InDelta(t, 1.25, *userRate, 1e-12)
}

func testFloat64Ptr(v float64) *float64 { return &v }
func testIntPtr(v int) *int             { return &v }

func TestSupportedModelPricingFromResolved_IncludesTokenIntervals(t *testing.T) {
	pricing := supportedModelPricingFromResolved(&service.ResolvedPricing{
		Mode: service.BillingModeToken,
		BasePricing: &service.ModelPricing{
			InputPricePerToken:  1e-6,
			OutputPricePerToken: 2e-6,
		},
		Intervals: []service.PricingInterval{
			{
				MinTokens:   0,
				MaxTokens:   testIntPtr(128000),
				InputPrice:  testFloat64Ptr(3e-6),
				OutputPrice: testFloat64Ptr(4e-6),
			},
		},
	}, 2)

	require.NotNil(t, pricing)
	require.Equal(t, string(service.BillingModeToken), pricing.BillingMode)
	require.NotNil(t, pricing.InputPricePerMillionTokens)
	require.NotNil(t, pricing.OutputPricePerMillionTokens)
	require.InDelta(t, 2, *pricing.InputPricePerMillionTokens, 1e-12)
	require.InDelta(t, 4, *pricing.OutputPricePerMillionTokens, 1e-12)
	require.Len(t, pricing.TokenIntervals, 1)
	require.Equal(t, 0, pricing.TokenIntervals[0].MinTokens)
	require.Equal(t, testIntPtr(128000), pricing.TokenIntervals[0].MaxTokens)
	require.NotNil(t, pricing.TokenIntervals[0].InputPricePerMillionTokens)
	require.NotNil(t, pricing.TokenIntervals[0].OutputPricePerMillionTokens)
	require.InDelta(t, 6, *pricing.TokenIntervals[0].InputPricePerMillionTokens, 1e-12)
	require.InDelta(t, 8, *pricing.TokenIntervals[0].OutputPricePerMillionTokens, 1e-12)
}

func TestSupportedModelPricingFromResolved_UsesEffectiveRateMultiplierForImagePricing(t *testing.T) {
	pricing := supportedModelPricingFromResolved(&service.ResolvedPricing{
		Mode:                   service.BillingModeImage,
		DefaultPerRequestPrice: 0.08,
		RequestTiers: []service.PricingInterval{
			{
				TierLabel:       "1K",
				PerRequestPrice: testFloat64Ptr(0.04),
			},
		},
	}, 1.5)

	require.NotNil(t, pricing)
	require.Equal(t, string(service.BillingModeImage), pricing.BillingMode)
	require.NotNil(t, pricing.DefaultPricePerRequest)
	require.InDelta(t, 0.12, *pricing.DefaultPricePerRequest, 1e-12)
	require.Len(t, pricing.RequestTiers, 1)
	require.Equal(t, "1K", pricing.RequestTiers[0].TierLabel)
	require.NotNil(t, pricing.RequestTiers[0].PricePerRequest)
	require.InDelta(t, 0.06, *pricing.RequestTiers[0].PricePerRequest, 1e-12)
}

func TestSupportedModelPricingFromService_UsesEffectiveRateMultiplier(t *testing.T) {
	h := NewAPIKeyHandler(nil, nil)
	h.SetBillingService(service.NewBillingService(&config.Config{}, nil))

	pricing := h.buildSupportedModelPricing(context.Background(), 0, "gpt-5.4", 1.5)
	require.NotNil(t, pricing)
	require.Equal(t, "USD", pricing.Currency)
	require.NotNil(t, pricing.InputPricePerMillionTokens)
	require.NotNil(t, pricing.OutputPricePerMillionTokens)
	require.InDelta(t, 3.75, *pricing.InputPricePerMillionTokens, 1e-9)
	require.InDelta(t, 22.5, *pricing.OutputPricePerMillionTokens, 1e-9)
}

func TestSupportedModelPricingFromService_ZeroMultiplierShowsZeroEffectivePrice(t *testing.T) {
	h := NewAPIKeyHandler(nil, nil)
	h.SetBillingService(service.NewBillingService(&config.Config{}, nil))

	pricing := h.buildSupportedModelPricing(context.Background(), 0, "gpt-5.4-mini", 0)
	require.NotNil(t, pricing)
	require.NotNil(t, pricing.InputPricePerMillionTokens)
	require.NotNil(t, pricing.OutputPricePerMillionTokens)
	require.InDelta(t, 0, *pricing.InputPricePerMillionTokens, 1e-12)
	require.InDelta(t, 0, *pricing.OutputPricePerMillionTokens, 1e-12)
}
