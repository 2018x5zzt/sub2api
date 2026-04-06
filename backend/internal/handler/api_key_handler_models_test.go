package handler

import (
	"reflect"
	"sort"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestCollectGroupModelIDs_ExpandsWildcardMappingSelectors(t *testing.T) {
	defaults := supportedModelsFromOpenAI(openai.DefaultModels)
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

	modelIDs, hasAnyMapping, includeDefaults := collectGroupModelIDs(service.PlatformOpenAI, accounts, defaults)
	if !hasAnyMapping {
		t.Fatal("expected wildcard mapping to be detected")
	}
	if includeDefaults {
		t.Fatal("did not expect wildcard-only API key account to force default models")
	}

	want := make([]string, 0, len(defaults))
	for _, model := range defaults {
		want = append(want, model.ID)
	}
	sort.Strings(want)
	if !reflect.DeepEqual(modelIDs, want) {
		t.Fatalf("collectGroupModelIDs() = %v, want %v", modelIDs, want)
	}
}

func TestCollectGroupModelIDs_OpenAIPassthroughIgnoresStaleMappings(t *testing.T) {
	defaults := supportedModelsFromOpenAI(openai.DefaultModels)
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

	modelIDs, hasAnyMapping, includeDefaults := collectGroupModelIDs(service.PlatformOpenAI, accounts, defaults)
	if hasAnyMapping {
		t.Fatalf("expected passthrough account mappings to be ignored, got %v", modelIDs)
	}
	if !includeDefaults {
		t.Fatal("expected passthrough account to fall back to default models")
	}
	if len(modelIDs) != 0 {
		t.Fatalf("expected no explicit mapped IDs, got %v", modelIDs)
	}
}

func TestCollectGroupModelIDs_DefaultAccountsKeepDefaultsAlongsideMappedModels(t *testing.T) {
	defaults := supportedModelsFromClaude(claude.DefaultModels)
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
					"claude-preview-custom": "upstream-preview",
				},
			},
		},
	}

	modelIDs, hasAnyMapping, includeDefaults := collectGroupModelIDs(service.PlatformAnthropic, accounts, defaults)
	if !hasAnyMapping {
		t.Fatal("expected API key mapping to be detected")
	}
	if !includeDefaults {
		t.Fatal("expected OAuth account to keep default models visible")
	}

	merged := mergeDefaultAndMappedModels(defaults, modelIDs)
	got := make([]string, 0, len(merged))
	for _, model := range merged {
		got = append(got, model.ID)
	}

	want := make([]string, 0, len(defaults)+1)
	for _, model := range defaults {
		want = append(want, model.ID)
	}
	want = append(want, "claude-preview-custom")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeDefaultAndMappedModels() = %v, want %v", got, want)
	}
}
