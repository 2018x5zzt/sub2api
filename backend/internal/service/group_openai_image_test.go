package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroup_AllowsOpenAIImageGeneration_UsesDedicatedGroupID(t *testing.T) {
	require.True(t, (&Group{
		ID:       30,
		Name:     "【限时半价】gpt-image",
		Platform: PlatformOpenAI,
	}).AllowsOpenAIImageGeneration())

	require.False(t, (&Group{
		ID:       30,
		Name:     "【限时半价】gpt-image",
		Platform: PlatformAnthropic,
	}).AllowsOpenAIImageGeneration())

	require.False(t, (&Group{
		ID:       1,
		Name:     "gpt-image",
		Platform: PlatformOpenAI,
	}).AllowsOpenAIImageGeneration())
}
