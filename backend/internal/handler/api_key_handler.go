// Package handler provides HTTP request handlers for the application.
package handler

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler handles API key-related requests
type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
	accountRepo   service.AccountRepository
}

// NewAPIKeyHandler creates a new APIKeyHandler
func NewAPIKeyHandler(apiKeyService *service.APIKeyService, accountRepo service.AccountRepository) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
		accountRepo:   accountRepo,
	}
}

// CreateAPIKeyRequest represents the create API key request payload
type CreateAPIKeyRequest struct {
	Name          string   `json:"name" binding:"required"`
	GroupID       *int64   `json:"group_id"`        // nullable
	CustomKey     *string  `json:"custom_key"`      // 可选的自定义key
	IPWhitelist   []string `json:"ip_whitelist"`    // IP 白名单
	IPBlacklist   []string `json:"ip_blacklist"`    // IP 黑名单
	Quota         *float64 `json:"quota"`           // 配额限制 (USD)
	ExpiresInDays *int     `json:"expires_in_days"` // 过期天数

	// Rate limit fields (0 = unlimited)
	RateLimit5h *float64 `json:"rate_limit_5h"`
	RateLimit1d *float64 `json:"rate_limit_1d"`
	RateLimit7d *float64 `json:"rate_limit_7d"`
}

// UpdateAPIKeyRequest represents the update API key request payload
type UpdateAPIKeyRequest struct {
	Name        string   `json:"name"`
	GroupID     *int64   `json:"group_id"`
	Status      string   `json:"status" binding:"omitempty,oneof=active inactive"`
	IPWhitelist []string `json:"ip_whitelist"` // IP 白名单
	IPBlacklist []string `json:"ip_blacklist"` // IP 黑名单
	Quota       *float64 `json:"quota"`        // 配额限制 (USD), 0=无限制
	ExpiresAt   *string  `json:"expires_at"`   // 过期时间 (ISO 8601)
	ResetQuota  *bool    `json:"reset_quota"`  // 重置已用配额

	// Rate limit fields (nil = no change, 0 = unlimited)
	RateLimit5h         *float64 `json:"rate_limit_5h"`
	RateLimit1d         *float64 `json:"rate_limit_1d"`
	RateLimit7d         *float64 `json:"rate_limit_7d"`
	ResetRateLimitUsage *bool    `json:"reset_rate_limit_usage"` // 重置限速用量
}

// List handles listing user's API keys with pagination
// GET /api/v1/api-keys
func (h *APIKeyHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	// Parse filter parameters
	var filters service.APIKeyListFilters
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		if len(search) > 100 {
			search = search[:100]
		}
		filters.Search = search
	}
	filters.Status = c.Query("status")
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		gid, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err == nil {
			filters.GroupID = &gid
		}
	}

	keys, result, err := h.apiKeyService.List(c.Request.Context(), subject.UserID, params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.APIKey, 0, len(keys))
	for i := range keys {
		out = append(out, *dto.APIKeyFromService(&keys[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// GetByID handles getting a single API key
// GET /api/v1/api-keys/:id
func (h *APIKeyHandler) GetByID(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	key, err := h.apiKeyService.GetByID(c.Request.Context(), keyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 验证所有权
	if key.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this key")
		return
	}

	response.Success(c, dto.APIKeyFromService(key))
}

// Create handles creating a new API key
// POST /api/v1/api-keys
func (h *APIKeyHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	svcReq := service.CreateAPIKeyRequest{
		Name:          req.Name,
		GroupID:       req.GroupID,
		CustomKey:     req.CustomKey,
		IPWhitelist:   req.IPWhitelist,
		IPBlacklist:   req.IPBlacklist,
		ExpiresInDays: req.ExpiresInDays,
	}
	if req.Quota != nil {
		svcReq.Quota = *req.Quota
	}
	if req.RateLimit5h != nil {
		svcReq.RateLimit5h = *req.RateLimit5h
	}
	if req.RateLimit1d != nil {
		svcReq.RateLimit1d = *req.RateLimit1d
	}
	if req.RateLimit7d != nil {
		svcReq.RateLimit7d = *req.RateLimit7d
	}

	executeUserIdempotentJSON(c, "user.api_keys.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		key, err := h.apiKeyService.Create(ctx, subject.UserID, svcReq)
		if err != nil {
			return nil, err
		}
		return dto.APIKeyFromService(key), nil
	})
}

// Update handles updating an API key
// PUT /api/v1/api-keys/:id
func (h *APIKeyHandler) Update(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	var req UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	svcReq := service.UpdateAPIKeyRequest{
		IPWhitelist:         req.IPWhitelist,
		IPBlacklist:         req.IPBlacklist,
		Quota:               req.Quota,
		ResetQuota:          req.ResetQuota,
		RateLimit5h:         req.RateLimit5h,
		RateLimit1d:         req.RateLimit1d,
		RateLimit7d:         req.RateLimit7d,
		ResetRateLimitUsage: req.ResetRateLimitUsage,
	}
	if req.Name != "" {
		svcReq.Name = &req.Name
	}
	svcReq.GroupID = req.GroupID
	if req.Status != "" {
		svcReq.Status = &req.Status
	}
	// Parse expires_at if provided
	if req.ExpiresAt != nil {
		if *req.ExpiresAt == "" {
			// Empty string means clear expiration
			svcReq.ExpiresAt = nil
			svcReq.ClearExpiration = true
		} else {
			t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				response.BadRequest(c, "Invalid expires_at format: "+err.Error())
				return
			}
			svcReq.ExpiresAt = &t
		}
	}

	key, err := h.apiKeyService.Update(c.Request.Context(), keyID, subject.UserID, svcReq)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.APIKeyFromService(key))
}

// Delete handles deleting an API key
// DELETE /api/v1/api-keys/:id
func (h *APIKeyHandler) Delete(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	err = h.apiKeyService.Delete(c.Request.Context(), keyID, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "API key deleted successfully"})
}

// GetAvailableGroups 获取用户可以绑定的分组列表
// GET /api/v1/groups/available
func (h *APIKeyHandler) GetAvailableGroups(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	groups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.Group, 0, len(groups))
	for i := range groups {
		out = append(out, *dto.GroupFromService(&groups[i]))
	}
	response.Success(c, out)
}

// GetAvailableGroupModels 获取当前用户可用分组的模型列表
// GET /api/v1/groups/models
func (h *APIKeyHandler) GetAvailableGroupModels(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	groups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.GroupModelCatalog, 0, len(groups))
	for i := range groups {
		models, source, err := h.getGroupSupportedModels(c.Request.Context(), &groups[i])
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		out = append(out, dto.GroupModelCatalog{
			Group:  *dto.GroupFromService(&groups[i]),
			Models: models,
			Source: source,
		})
	}
	response.Success(c, out)
}

// GetUserGroupRates 获取当前用户的专属分组倍率配置
// GET /api/v1/groups/rates
func (h *APIKeyHandler) GetUserGroupRates(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	rates, err := h.apiKeyService.GetUserGroupRates(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, rates)
}

func (h *APIKeyHandler) getGroupSupportedModels(ctx context.Context, group *service.Group) ([]dto.SupportedModel, string, error) {
	if group == nil {
		return nil, "mapping", nil
	}

	if h.accountRepo == nil {
		return nil, "mapping", nil
	}

	accounts, err := h.accountRepo.ListSchedulableByGroupID(ctx, group.ID)
	if err != nil {
		return nil, "", err
	}

	mappedModelIDs := collectGroupModelIDs(group.Platform, accounts)
	catalog := staticCatalogModelsForPlatform(group.Platform)
	mappedModelIDs = filterMappedModelIDsByCatalog(group.Platform, mappedModelIDs)
	return buildMappedModels(mappedModelIDs, catalog), "mapping", nil
}

func staticCatalogModelsForPlatform(platform string) []dto.SupportedModel {
	switch platform {
	case service.PlatformOpenAI:
		return supportedModelsFromOpenAI(openai.DefaultModels)
	default:
		return nil
	}
}

func filterMappedModelIDsByCatalog(platform string, modelIDs []string) []string {
	if len(modelIDs) == 0 {
		return nil
	}

	switch platform {
	case service.PlatformOpenAI:
		filtered := make([]string, 0, len(modelIDs))
		for _, modelID := range modelIDs {
			if openai.IsDefaultModel(modelID) {
				filtered = append(filtered, modelID)
			}
		}
		if len(filtered) == 0 {
			return nil
		}
		return filtered
	default:
		return modelIDs
	}
}

func collectGroupModelIDs(platform string, accounts []service.Account) []string {
	modelSet := make(map[string]struct{})

	for i := range accounts {
		account := &accounts[i]
		mapping := configuredModelMapping(account)
		if accountUsesImplicitModelCatalog(platform, account, mapping) {
			continue
		}
		addConcreteMappedModelIDs(modelSet, mapping)
	}

	if len(modelSet) == 0 {
		return nil
	}

	models := make([]string, 0, len(modelSet))
	for modelID := range modelSet {
		models = append(models, modelID)
	}
	sort.Strings(models)
	return models
}

func configuredModelMapping(account *service.Account) map[string]string {
	if account == nil || account.Credentials == nil {
		return nil
	}

	rawMapping, _ := account.Credentials["model_mapping"].(map[string]any)
	if len(rawMapping) == 0 {
		return nil
	}

	mapping := make(map[string]string, len(rawMapping))
	for selector, target := range rawMapping {
		targetModel, ok := target.(string)
		if !ok {
			continue
		}
		mapping[selector] = targetModel
	}
	if len(mapping) == 0 {
		return nil
	}
	return mapping
}

func accountUsesImplicitModelCatalog(platform string, account *service.Account, mapping map[string]string) bool {
	if account == nil {
		return false
	}
	if len(mapping) == 0 {
		return true
	}

	switch platform {
	case service.PlatformOpenAI:
		return account.IsOpenAIPassthroughEnabled()
	case service.PlatformGemini:
		return account.IsOAuth()
	default:
		return account.IsOAuth()
	}
}

func addConcreteMappedModelIDs(modelSet map[string]struct{}, mapping map[string]string) {
	for selector := range mapping {
		modelID := strings.TrimSpace(selector)
		if modelID == "" || strings.Contains(modelID, "*") {
			continue
		}
		modelSet[modelID] = struct{}{}
	}
}

func buildMappedModels(mappedModelIDs []string, defaults []dto.SupportedModel) []dto.SupportedModel {
	if len(mappedModelIDs) == 0 {
		return nil
	}

	defaultMap := make(map[string]dto.SupportedModel, len(defaults))
	for _, model := range defaults {
		defaultMap[model.ID] = model
	}

	models := make([]dto.SupportedModel, 0, len(mappedModelIDs))
	remaining := make(map[string]struct{}, len(mappedModelIDs))
	for _, modelID := range mappedModelIDs {
		remaining[modelID] = struct{}{}
	}

	for _, model := range defaults {
		if _, ok := remaining[model.ID]; ok {
			models = append(models, model)
			delete(remaining, model.ID)
		}
	}

	if len(remaining) == 0 {
		return models
	}

	extraIDs := make([]string, 0, len(remaining))
	for modelID := range remaining {
		extraIDs = append(extraIDs, modelID)
	}
	sort.Strings(extraIDs)
	for _, modelID := range extraIDs {
		model, ok := defaultMap[modelID]
		if ok {
			models = append(models, model)
			continue
		}
		models = append(models, dto.SupportedModel{
			ID:          modelID,
			DisplayName: modelID,
		})
	}
	return models
}

func mergeDefaultAndMappedModels(defaults []dto.SupportedModel, mappedModelIDs []string) []dto.SupportedModel {
	models := make([]dto.SupportedModel, 0, len(defaults)+len(mappedModelIDs))
	models = append(models, defaults...)

	seen := make(map[string]struct{}, len(defaults))
	for _, model := range defaults {
		seen[model.ID] = struct{}{}
	}

	extraIDs := make([]string, 0, len(mappedModelIDs))
	for _, modelID := range mappedModelIDs {
		if _, ok := seen[modelID]; ok {
			continue
		}
		seen[modelID] = struct{}{}
		extraIDs = append(extraIDs, modelID)
	}
	sort.Strings(extraIDs)
	for _, modelID := range extraIDs {
		models = append(models, dto.SupportedModel{
			ID:          modelID,
			DisplayName: modelID,
		})
	}
	return models
}

func supportedModelsFromOpenAI(models []openai.Model) []dto.SupportedModel {
	out := make([]dto.SupportedModel, 0, len(models))
	for _, model := range models {
		displayName := model.DisplayName
		if displayName == "" {
			displayName = model.ID
		}
		out = append(out, dto.SupportedModel{
			ID:                 model.ID,
			DisplayName:        displayName,
			InputPricePerMTok:  model.InputPricePerMTok,
			OutputPricePerMTok: model.OutputPricePerMTok,
		})
	}
	return out
}

func supportedModelsFromClaude(models []claude.Model) []dto.SupportedModel {
	out := make([]dto.SupportedModel, 0, len(models))
	for _, model := range models {
		displayName := model.DisplayName
		if displayName == "" {
			displayName = model.ID
		}
		out = append(out, dto.SupportedModel{
			ID:          model.ID,
			DisplayName: displayName,
		})
	}
	return out
}

func supportedModelsFromGemini(models []geminicli.Model) []dto.SupportedModel {
	out := make([]dto.SupportedModel, 0, len(models))
	for _, model := range models {
		displayName := model.DisplayName
		if displayName == "" {
			displayName = model.ID
		}
		out = append(out, dto.SupportedModel{
			ID:          model.ID,
			DisplayName: displayName,
		})
	}
	return out
}

func supportedModelsFromAntigravity(models []antigravity.ClaudeModel) []dto.SupportedModel {
	out := make([]dto.SupportedModel, 0, len(models))
	for _, model := range models {
		displayName := model.DisplayName
		if displayName == "" {
			displayName = model.ID
		}
		out = append(out, dto.SupportedModel{
			ID:          model.ID,
			DisplayName: displayName,
		})
	}
	return out
}

func filterAntigravityModelsByScopes(models []dto.SupportedModel, scopes []string) []dto.SupportedModel {
	if len(scopes) == 0 {
		return models
	}

	allowedScopes := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		allowedScopes[strings.TrimSpace(scope)] = struct{}{}
	}

	filtered := make([]dto.SupportedModel, 0, len(models))
	for _, model := range models {
		if antigravityScopeAllowsModel(model.ID, allowedScopes) {
			filtered = append(filtered, model)
		}
	}
	return filtered
}

func antigravityScopeAllowsModel(modelID string, allowedScopes map[string]struct{}) bool {
	if len(allowedScopes) == 0 {
		return true
	}

	modelLower := strings.ToLower(modelID)
	if strings.Contains(modelLower, "claude") {
		_, ok := allowedScopes["claude"]
		return ok
	}

	if strings.Contains(modelLower, "gemini") {
		if strings.Contains(modelLower, "image") {
			_, ok := allowedScopes["gemini_image"]
			return ok
		}
		_, ok := allowedScopes["gemini_text"]
		return ok
	}

	return true
}
