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

	require.True(t, (&Group{
		ID:       35,
		Name:     "【订阅】gpt-image",
		Platform: PlatformOpenAI,
	}).AllowsOpenAIImageGeneration())

	require.False(t, (&Group{
		ID:       35,
		Name:     "【订阅】gpt-image",
		Platform: PlatformAnthropic,
	}).AllowsOpenAIImageGeneration())

	require.False(t, (&Group{
		ID:       1,
		Name:     "pro号池",
		Platform: PlatformOpenAI,
	}).AllowsOpenAIImageGeneration())
}
