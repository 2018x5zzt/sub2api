package claude

import "testing"

func TestDefaultModels_ContainsHaiku45Release(t *testing.T) {
	for _, model := range DefaultModels {
		if model.ID == "claude-haiku-4-5-20251001" {
			return
		}
	}

	t.Fatalf("expected claude-haiku-4-5-20251001 to be exposed in DefaultModels, got %v", DefaultModelIDs())
}
