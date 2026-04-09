package handler

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

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
