package openai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookupDefaultModel_GPT55Pricing(t *testing.T) {
	model, ok := LookupDefaultModel("gpt-5.5")
	require.True(t, ok)
	require.InDelta(t, 5.0, model.InputPricePerMTok, 1e-12)
	require.InDelta(t, 30.0, model.OutputPricePerMTok, 1e-12)
}
