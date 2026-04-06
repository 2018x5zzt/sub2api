package handler

import (
	"reflect"
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
