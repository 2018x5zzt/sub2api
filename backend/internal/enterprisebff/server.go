package enterprisebff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Server struct {
	cfg             *Config
	httpClient      *http.Client
	router          *gin.Engine
	adminKeySvc     AdminKeyStore
	enterpriseStore EnterpriseStore
	healthRepo      service.GroupHealthSnapshotRepository
}

func New(cfg *Config, adminKeySvc AdminKeyStore, enterpriseStore EnterpriseStore, healthRepo service.GroupHealthSnapshotRepository) *Server {
	s := &Server{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		adminKeySvc:     adminKeySvc,
		enterpriseStore: enterpriseStore,
		healthRepo:      healthRepo,
	}
	s.router = s.newRouter()
	return s
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) newRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(s.traceRequests())

	if len(s.cfg.TrustedProxies) > 0 {
		if err := r.SetTrustedProxies(s.cfg.TrustedProxies); err != nil {
			log.Printf("enterprise-bff: failed to set trusted proxies: %v", err)
		}
	} else {
		if err := r.SetTrustedProxies(nil); err != nil {
			log.Printf("enterprise-bff: failed to disable trusted proxies: %v", err)
		}
	}

	r.GET("/healthz", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status":  "ok",
			"service": "enterprise-bff",
		})
	})

	// Core compatibility endpoints used by the current frontend.
	r.POST("/auth/login", s.handleEnterpriseLogin)
	r.POST("/auth/login/2fa", s.handleEnterpriseLogin2FA)
	r.POST("/auth/refresh", func(c *gin.Context) { s.proxy(c, "/auth/refresh", nil) })
	r.POST("/auth/logout", func(c *gin.Context) { s.proxy(c, "/auth/logout", nil) })
	r.GET("/auth/me", s.handleEnterpriseMe)
	r.GET("/settings/public", s.handleEnterprisePublicSettings)
	r.GET("/groups/available", s.handleEnterpriseVisibleGroups)
	r.GET("/groups/pool-status", s.handleEnterprisePoolStatus)

	r.GET("/usage", s.handleRoleAwareUsageList)
	r.GET("/usage/stats", s.handleRoleAwareUsageStats)
	r.GET("/usage/:id", func(c *gin.Context) { s.proxy(c, buildPathf("/usage/%s", c.Param("id")), transformUsageEnvelope) })
	r.GET("/usage/dashboard/stats", func(c *gin.Context) { s.proxy(c, "/usage/dashboard/stats", nil) })
	r.GET("/usage/dashboard/trend", func(c *gin.Context) { s.proxy(c, "/usage/dashboard/trend", nil) })
	r.GET("/usage/dashboard/models", func(c *gin.Context) { s.proxy(c, "/usage/dashboard/models", nil) })
	r.POST("/usage/dashboard/api-keys-usage", func(c *gin.Context) { s.proxy(c, "/usage/dashboard/api-keys-usage", nil) })

	r.GET("/keys", s.handleRoleAwareKeyList)
	r.GET("/keys/:id", s.handleRoleAwareKeyGet)
	r.POST("/keys", s.handleRoleAwareKeyCreate)
	r.PUT("/keys/:id", s.handleRoleAwareKeyUpdate)
	r.PATCH("/keys/:id", s.handleRoleAwareKeyUpdate)
	r.DELETE("/keys/:id", s.handleRoleAwareKeyDelete)

	// Existing admin frontend routes in the current repo.
	r.GET("/admin/usage", func(c *gin.Context) { s.proxy(c, "/admin/usage", transformUsageEnvelope) })
	r.GET("/admin/usage/stats", func(c *gin.Context) { s.proxy(c, "/admin/usage/stats", nil) })
	r.GET("/admin/usage/search-users", func(c *gin.Context) { s.proxy(c, "/admin/usage/search-users", nil) })
	r.GET("/admin/usage/search-api-keys", func(c *gin.Context) { s.proxy(c, "/admin/usage/search-api-keys", nil) })
	r.GET("/admin/usage/cleanup-tasks", func(c *gin.Context) { s.proxy(c, "/admin/usage/cleanup-tasks", nil) })
	r.POST("/admin/usage/cleanup-tasks", func(c *gin.Context) { s.proxy(c, "/admin/usage/cleanup-tasks", nil) })
	r.POST("/admin/usage/cleanup-tasks/:id/cancel", func(c *gin.Context) {
		s.proxy(c, buildPathf("/admin/usage/cleanup-tasks/%s/cancel", c.Param("id")), nil)
	})
	r.PUT("/admin/api-keys/:id", func(c *gin.Context) {
		s.proxy(c, buildPathf("/admin/api-keys/%s", c.Param("id")), transformKeysEnvelope)
	})

	// Canonical v1 aliases.
	r.GET("/v1/session/me", s.handleEnterpriseMe)
	r.GET("/v1/branding", s.handleEnterprisePublicSettings)
	r.GET("/v1/admin/usage", func(c *gin.Context) { s.proxy(c, "/admin/usage", transformUsageEnvelope) })
	r.GET("/v1/admin/keys", s.handleAdminKeyList)
	r.POST("/v1/admin/keys", s.handleAdminKeyCreate)
	r.PATCH("/v1/admin/keys/:keyId", s.handleAdminKeyUpdate)
	r.PUT("/v1/admin/keys/:keyId", s.handleAdminKeyUpdate)
	r.DELETE("/v1/admin/keys/:keyId", s.handleAdminKeyDelete)

	return r
}

func (s *Server) traceRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.NewString()
			c.Request.Header.Set("X-Request-ID", requestID)
		}

		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()

		contractVersion := strings.TrimSpace(c.GetHeader("X-Contract-Version"))
		idempotencyKey := strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
		log.Printf(
			"enterprise-bff request_id=%s method=%s path=%s status=%d duration=%s contract_version=%q idempotency_key=%q",
			requestID,
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			time.Since(start).Round(time.Millisecond),
			contractVersion,
			idempotencyKey,
		)
	}
}

func (s *Server) handleRoleAwareUsageList(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}
	if user.Role == "admin" {
		s.proxy(c, "/admin/usage", transformUsageEnvelope)
		return
	}
	s.proxy(c, "/usage", transformUsageEnvelope)
}

func (s *Server) handleRoleAwareUsageStats(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}
	if user.Role == "admin" {
		s.proxy(c, "/admin/usage/stats", nil)
		return
	}
	s.proxy(c, "/usage/stats", nil)
}

func (s *Server) handleRoleAwareKeyList(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}
	if user.Role == "admin" {
		s.renderAdminKeyList(c)
		return
	}
	s.proxy(c, "/keys", transformKeysEnvelope)
}

func (s *Server) handleRoleAwareKeyGet(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}
	if user.Role == "admin" {
		s.renderAdminKeyGet(c)
		return
	}
	s.proxy(c, buildPathf("/keys/%s", c.Param("id")), transformKeysEnvelope)
}

func (s *Server) handleRoleAwareKeyCreate(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}
	if user.Role == "admin" {
		s.renderAdminKeyCreate(c, user)
		return
	}
	s.proxyValidatedKeyMutation(c, user, user.ID, "/keys", transformKeysEnvelope, true)
}

func (s *Server) handleRoleAwareKeyUpdate(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}
	if user.Role == "admin" {
		s.renderAdminKeyUpdate(c, user)
		return
	}
	s.proxyValidatedKeyMutation(c, user, user.ID, buildPathf("/keys/%s", c.Param("id")), transformKeysEnvelope, false)
}

func (s *Server) handleRoleAwareKeyDelete(c *gin.Context) {
	user, ok := s.requireCurrentUser(c)
	if !ok {
		return
	}
	if user.Role == "admin" {
		s.renderAdminKeyDelete(c)
		return
	}
	s.proxy(c, buildPathf("/keys/%s", c.Param("id")), nil)
}

func (s *Server) handleAdminKeyList(c *gin.Context) {
	if _, ok := s.requireAdmin(c); !ok {
		return
	}
	s.renderAdminKeyList(c)
}

func (s *Server) handleAdminKeyCreate(c *gin.Context) {
	user, ok := s.requireAdmin(c)
	if !ok {
		return
	}
	s.renderAdminKeyCreate(c, user)
}

func (s *Server) handleAdminKeyUpdate(c *gin.Context) {
	user, ok := s.requireAdmin(c)
	if !ok {
		return
	}
	s.renderAdminKeyUpdate(c, user)
}

func (s *Server) handleAdminKeyDelete(c *gin.Context) {
	if _, ok := s.requireAdmin(c); !ok {
		return
	}
	s.renderAdminKeyDelete(c)
}

type adminCreateAPIKeyRequest struct {
	handler.CreateAPIKeyRequest
	UserID *int64 `json:"user_id"`
}

func (s *Server) renderAdminKeyList(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	filters, err := parseAdminKeyListFilters(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	keys, total, err := s.adminKeySvc.List(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]map[string]any, 0, len(keys))
	for i := range keys {
		items = append(items, compatKeyMap(dto.APIKeyFromService(&keys[i])))
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (s *Server) renderAdminKeyGet(c *gin.Context) {
	keyID, err := parsePathID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	key, err := s.adminKeySvc.Get(c.Request.Context(), keyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, compatKeyMap(dto.APIKeyFromService(key)))
}

func (s *Server) renderAdminKeyCreate(c *gin.Context, currentUser *currentUser) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	binding, err := parseRequestedGroupBinding(body)
	if err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req adminCreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	ownerID := currentUser.ID
	if req.UserID != nil && *req.UserID > 0 {
		ownerID = *req.UserID
	}
	if err := s.authorizeRequestedGroup(c.Request.Context(), currentUser, ownerID, binding); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": err.Error(),
		})
		return
	}

	created, err := s.adminKeySvc.Create(c.Request.Context(), ownerID, service.CreateAPIKeyRequest{
		Name:          req.Name,
		GroupID:       req.GroupID,
		CustomKey:     req.CustomKey,
		IPWhitelist:   req.IPWhitelist,
		IPBlacklist:   req.IPBlacklist,
		ExpiresInDays: req.ExpiresInDays,
		Quota:         derefFloat64(req.Quota),
		RateLimit5h:   derefFloat64(req.RateLimit5h),
		RateLimit1d:   derefFloat64(req.RateLimit1d),
		RateLimit7d:   derefFloat64(req.RateLimit7d),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, compatKeyMap(dto.APIKeyFromService(created)))
}

func (s *Server) renderAdminKeyUpdate(c *gin.Context, currentUser *currentUser) {
	keyID, err := parsePathID(c.Param("id"))
	if err != nil {
		keyID, err = parsePathID(c.Param("keyId"))
	}
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	existing, err := s.adminKeySvc.Get(c.Request.Context(), keyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	binding, err := parseRequestedGroupBinding(body)
	if err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := s.authorizeRequestedGroup(c.Request.Context(), currentUser, existing.UserID, binding); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": err.Error(),
		})
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req handler.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	updateReq := service.UpdateAPIKeyRequest{
		GroupID:             req.GroupID,
		Status:              nilIfEmpty(req.Status),
		IPWhitelist:         req.IPWhitelist,
		IPBlacklist:         req.IPBlacklist,
		Quota:               req.Quota,
		ResetQuota:          req.ResetQuota,
		RateLimit5h:         req.RateLimit5h,
		RateLimit1d:         req.RateLimit1d,
		RateLimit7d:         req.RateLimit7d,
		ResetRateLimitUsage: req.ResetRateLimitUsage,
		ClearExpiration:     false,
	}
	if req.Name != "" {
		updateReq.Name = &req.Name
	}
	if req.ExpiresAt != nil {
		if *req.ExpiresAt == "" {
			updateReq.ClearExpiration = true
		} else {
			parsed, parseErr := time.Parse(time.RFC3339, *req.ExpiresAt)
			if parseErr != nil {
				response.BadRequest(c, "Invalid expires_at format: "+parseErr.Error())
				return
			}
			updateReq.ExpiresAt = &parsed
		}
	}

	updated, err := s.adminKeySvc.Update(c.Request.Context(), keyID, updateReq)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, compatKeyMap(dto.APIKeyFromService(updated)))
}

func (s *Server) renderAdminKeyDelete(c *gin.Context) {
	keyID, err := parsePathID(c.Param("id"))
	if err != nil {
		keyID, err = parsePathID(c.Param("keyId"))
	}
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	if err := s.adminKeySvc.Delete(c.Request.Context(), keyID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "API key deleted successfully"})
}

func parseAdminKeyListFilters(c *gin.Context) (AdminKeyListFilters, error) {
	filters := AdminKeyListFilters{
		Search: strings.TrimSpace(c.Query("search")),
		Status: strings.TrimSpace(c.Query("status")),
	}

	if userID := strings.TrimSpace(c.Query("user_id")); userID != "" {
		value, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return filters, fmt.Errorf("Invalid user_id")
		}
		filters.UserID = &value
	}

	if groupID := strings.TrimSpace(c.Query("group_id")); groupID != "" {
		value, err := strconv.ParseInt(groupID, 10, 64)
		if err != nil {
			return filters, fmt.Errorf("Invalid group_id")
		}
		filters.GroupID = &value
	}

	return filters, nil
}

func parsePathID(raw string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
}

func nilIfEmpty(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func compatKeyMap(value *dto.APIKey) map[string]any {
	return compatObjectMap(value, normalizeCompatibleKeyMap)
}

func compatObjectMap(value any, mutate func(map[string]any) map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	out := make(map[string]any)
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return mutate(out)
}

func derefFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func Run(ctx context.Context) error {
	cfg, sharedCfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if strings.EqualFold(sharedCfg.Server.Mode, gin.ReleaseMode) {
		gin.SetMode(gin.ReleaseMode)
	}

	dbResources, err := OpenDB(sharedCfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = dbResources.Close()
	}()

	server := New(
		cfg,
		NewAdminKeyStore(dbResources.Client, dbResources.SQLDB, sharedCfg),
		newEntEnterpriseStore(dbResources.Client, repository.NewSettingRepository(dbResources.Client), cfg),
		repository.NewGroupHealthSnapshotRepository(dbResources.Client, dbResources.SQLDB),
	)
	httpServer := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           server.Router(),
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	log.Printf("enterprise-bff started on %s forwarding to %s", cfg.ListenAddr, cfg.CoreBaseURL.String())
	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
