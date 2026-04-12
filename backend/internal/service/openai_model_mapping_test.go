package service

import "testing"

func TestResolveOpenAIForwardModel(t *testing.T) {
	tests := []struct {
		name               string
		account            *Account
		requestedModel     string
		defaultMappedModel string
		expectedModel      string
	}{
		{
			name: "falls back to group default when account has no mapping",
			account: &Account{
				Credentials: map[string]any{},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-4o-mini",
		},
		{
			name: "preserves exact passthrough mapping instead of group default",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5.4": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "preserves wildcard passthrough mapping instead of group default",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-*": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "uses account remap when explicit target differs",
			account: &Account{
				Platform: PlatformOpenAI,
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "matches canonical codex mapping for spark alias",
			account: &Account{
				Platform: PlatformOpenAI,
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5.3-codex": "gpt-4.1",
					},
				},
			},
			requestedModel:     "gpt-5.3-codex-spark",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-4.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveOpenAIForwardModel(tt.account, tt.requestedModel, tt.defaultMappedModel); got != tt.expectedModel {
				t.Fatalf("resolveOpenAIForwardModel(...) = %q, want %q", got, tt.expectedModel)
			}
		})
	}
}

func TestResolveOpenAIForwardModel_PreventsClaudeModelFromFallingBackToGpt51(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{},
	}

	withoutDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "")
	if got := normalizeCodexModel(withoutDefault); got != "claude-opus-4-6" {
		t.Fatalf("normalizeCodexModel(%q) = %q, want %q", withoutDefault, got, "claude-opus-4-6")
	}

	withDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "gpt-5.4")
	if got := normalizeCodexModel(withDefault); got != "gpt-5.4" {
		t.Fatalf("normalizeCodexModel(%q) = %q, want %q", withDefault, got, "gpt-5.4")
	}
}
