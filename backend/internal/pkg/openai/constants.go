// Package openai provides helpers and types for OpenAI API integration.
package openai

import _ "embed"

// Model represents an OpenAI model
type Model struct {
	ID                 string  `json:"id"`
	Object             string  `json:"object"`
	Created            int64   `json:"created"`
	OwnedBy            string  `json:"owned_by"`
	Type               string  `json:"type"`
	DisplayName        string  `json:"display_name"`
	InputPricePerMTok  float64 `json:"input_price_per_mtoken,omitempty"`
	OutputPricePerMTok float64 `json:"output_price_per_mtoken,omitempty"`
}

// DefaultModels OpenAI models list
var DefaultModels = []Model{
	{ID: "gpt-5", Object: "model", Created: 1722988800, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5", InputPricePerMTok: 1.25, OutputPricePerMTok: 10.00},
	{ID: "gpt-5-codex", Object: "model", Created: 1722988800, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5 Codex", InputPricePerMTok: 1.25, OutputPricePerMTok: 10.00},
	{ID: "gpt-5-codex-mini", Object: "model", Created: 1722988800, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5 Codex Mini", InputPricePerMTok: 0.25, OutputPricePerMTok: 2.00},
	{ID: "gpt-5.1", Object: "model", Created: 1731456000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.1", InputPricePerMTok: 1.25, OutputPricePerMTok: 10.00},
	{ID: "gpt-5.1-codex", Object: "model", Created: 1730419200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.1 Codex", InputPricePerMTok: 1.25, OutputPricePerMTok: 10.00},
	{ID: "gpt-5.1-codex-max", Object: "model", Created: 1730419200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.1 Codex Max", InputPricePerMTok: 1.25, OutputPricePerMTok: 10.00},
	{ID: "gpt-5.1-codex-mini", Object: "model", Created: 1730419200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.1 Codex Mini", InputPricePerMTok: 0.25, OutputPricePerMTok: 2.00},
	{ID: "gpt-5.2", Object: "model", Created: 1733875200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.2", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.2-codex", Object: "model", Created: 1733011200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.2 Codex", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.2-high", Object: "model", Created: 1733875200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.2 High", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.2-low", Object: "model", Created: 1733875200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.2 Low", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.2-medium", Object: "model", Created: 1733875200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.2 Medium", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.2-xhigh", Object: "model", Created: 1733875200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.2 XHigh", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.3-codex", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.3 Codex", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.3-codex-high", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.3 Codex High", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.3-codex-low", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.3 Codex Low", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.3-codex-medium", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.3 Codex Medium", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.3-codex-xhigh", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.3 Codex XHigh", InputPricePerMTok: 1.75, OutputPricePerMTok: 14.00},
	{ID: "gpt-5.4", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4", InputPricePerMTok: 2.50, OutputPricePerMTok: 15.00},
	{ID: "gpt-5.4-high", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4 High", InputPricePerMTok: 2.50, OutputPricePerMTok: 15.00},
	{ID: "gpt-5.4-medium", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4 Medium", InputPricePerMTok: 2.50, OutputPricePerMTok: 15.00},
	{ID: "gpt-5.4-mini", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4 Mini", InputPricePerMTok: 0.75, OutputPricePerMTok: 4.50},
	{ID: "gpt-5.4-nano", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4 Nano", InputPricePerMTok: 0.20, OutputPricePerMTok: 1.25},
	{ID: "gpt-5.4-xhigh", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4 XHigh", InputPricePerMTok: 2.50, OutputPricePerMTok: 15.00},
	{ID: "gpt-image-2", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 2"},
}

var defaultModelsByID = func() map[string]Model {
	out := make(map[string]Model, len(DefaultModels))
	for _, model := range DefaultModels {
		out[model.ID] = model
	}
	return out
}()

// LookupDefaultModel returns the static OpenAI catalog entry for a model ID.
func LookupDefaultModel(modelID string) (Model, bool) {
	model, ok := defaultModelsByID[modelID]
	return model, ok
}

// IsDefaultModel reports whether the model ID exists in the static OpenAI catalog.
func IsDefaultModel(modelID string) bool {
	_, ok := defaultModelsByID[modelID]
	return ok
}

// DefaultModelIDs returns the default model ID list
func DefaultModelIDs() []string {
	ids := make([]string, len(DefaultModels))
	for i, m := range DefaultModels {
		ids[i] = m.ID
	}
	return ids
}

// DefaultTestModel default model for testing OpenAI accounts
const DefaultTestModel = "gpt-5.4"

// DefaultInstructions default instructions for non-Codex CLI requests
// Content loaded from instructions.txt at compile time
//
//go:embed instructions.txt
var DefaultInstructions string
