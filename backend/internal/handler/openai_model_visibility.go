package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func filterOpenAIModelIDsForGroup(group *service.Group, modelIDs []string) []string {
	if group == nil || group.AllowsOpenAIImageGeneration() {
		return append([]string(nil), modelIDs...)
	}

	filtered := make([]string, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		if service.IsOpenAIImageGenerationModel(modelID) {
			continue
		}
		filtered = append(filtered, modelID)
	}
	return filtered
}

func filterOpenAIModelsForGroup(group *service.Group, models []openai.Model) []openai.Model {
	if group == nil || group.AllowsOpenAIImageGeneration() {
		return append([]openai.Model(nil), models...)
	}

	filtered := make([]openai.Model, 0, len(models))
	for _, model := range models {
		if service.IsOpenAIImageGenerationModel(model.ID) {
			continue
		}
		filtered = append(filtered, model)
	}
	return filtered
}

func filterSupportedModelsForGroup(group *service.Group, models []dto.SupportedModel) []dto.SupportedModel {
	if group == nil || group.AllowsOpenAIImageGeneration() {
		return append([]dto.SupportedModel(nil), models...)
	}

	filtered := make([]dto.SupportedModel, 0, len(models))
	for _, model := range models {
		if service.IsOpenAIImageGenerationModel(model.ID) {
			continue
		}
		filtered = append(filtered, model)
	}
	return filtered
}
