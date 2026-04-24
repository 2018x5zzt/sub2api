package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func isOpenAIModelAllowedForGroup(group *service.Group, modelID string) bool {
	if group == nil {
		return true
	}
	isImageModel := service.IsOpenAIImageGenerationModel(modelID)
	if group.AllowsOpenAIImageGeneration() {
		return isImageModel
	}
	return !isImageModel
}

func filterOpenAIModelIDsForGroup(group *service.Group, modelIDs []string) []string {
	if group == nil {
		return append([]string(nil), modelIDs...)
	}

	filtered := make([]string, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		if !isOpenAIModelAllowedForGroup(group, modelID) {
			continue
		}
		filtered = append(filtered, modelID)
	}
	return filtered
}

func filterOpenAIModelsForGroup(group *service.Group, models []openai.Model) []openai.Model {
	if group == nil {
		return append([]openai.Model(nil), models...)
	}

	filtered := make([]openai.Model, 0, len(models))
	for _, model := range models {
		if !isOpenAIModelAllowedForGroup(group, model.ID) {
			continue
		}
		filtered = append(filtered, model)
	}
	return filtered
}

func filterSupportedModelsForGroup(group *service.Group, models []dto.SupportedModel) []dto.SupportedModel {
	if group == nil {
		return append([]dto.SupportedModel(nil), models...)
	}

	filtered := make([]dto.SupportedModel, 0, len(models))
	for _, model := range models {
		if !isOpenAIModelAllowedForGroup(group, model.ID) {
			continue
		}
		filtered = append(filtered, model)
	}
	return filtered
}
