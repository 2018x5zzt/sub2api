package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateOpenAITestPayload_NormalizesReasoningAliasModel(t *testing.T) {
	payload := createOpenAITestPayload("gpt-5.4-high", "hi", false)

	require.Equal(t, "gpt-5.4", payload["model"])
	reasoning, ok := payload["reasoning"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "high", reasoning["effort"])
	require.Equal(t, "auto", reasoning["summary"])
}
