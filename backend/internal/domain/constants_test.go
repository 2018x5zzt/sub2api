package domain

import "testing"

func TestDefaultAntigravityModelMapping_ImageCompatibilityAliases(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"gemini-2.5-flash-image":         "gemini-2.5-flash-image",
		"gemini-2.5-flash-image-preview": "gemini-2.5-flash-image",
		"gemini-3.1-flash-image":         "gemini-3.1-flash-image",
		"gemini-3.1-flash-image-preview": "gemini-3.1-flash-image",
		"gemini-3-pro-image":             "gemini-3.1-flash-image",
		"gemini-3-pro-image-preview":     "gemini-3.1-flash-image",
	}

	for from, want := range cases {
		got, ok := DefaultAntigravityModelMapping[from]
		if !ok {
			t.Fatalf("expected mapping for %q to exist", from)
		}
		if got != want {
			t.Fatalf("unexpected mapping for %q: got %q want %q", from, got, want)
		}
	}
}

func TestDefaultModelMappings_ContainOpus47(t *testing.T) {
	t.Parallel()

	if got, ok := DefaultAntigravityModelMapping["claude-opus-4-7"]; !ok {
		t.Fatal("expected antigravity mapping for claude-opus-4-7 to exist")
	} else if got != "claude-opus-4-7" {
		t.Fatalf("unexpected antigravity mapping for claude-opus-4-7: got %q want %q", got, "claude-opus-4-7")
	}

	if got, ok := DefaultBedrockModelMapping["claude-opus-4-7"]; !ok {
		t.Fatal("expected bedrock mapping for claude-opus-4-7 to exist")
	} else if got != "us.anthropic.claude-opus-4-7-v1" {
		t.Fatalf("unexpected bedrock mapping for claude-opus-4-7: got %q want %q", got, "us.anthropic.claude-opus-4-7-v1")
	}
}
