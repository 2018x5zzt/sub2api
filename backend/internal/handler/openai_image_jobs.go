package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	openAIImageJobDefaultConcurrency = 2
	openAIImageJobDefaultTimeout     = 20 * time.Minute
	openAIImageJobDefaultTTL         = 24 * time.Hour
	openAIImageJobDefaultMaxAttempts = 3
)

var openAIImageJobDefaultRetryBackoff = []time.Duration{10 * time.Second, 30 * time.Second}

type openAIImageJobStatus = service.OpenAIImageJobStatus

const (
	openAIImageJobStatusQueued    openAIImageJobStatus = service.OpenAIImageJobStatusQueued
	openAIImageJobStatusRunning   openAIImageJobStatus = service.OpenAIImageJobStatusRunning
	openAIImageJobStatusSucceeded openAIImageJobStatus = service.OpenAIImageJobStatusSucceeded
	openAIImageJobStatusFailed    openAIImageJobStatus = service.OpenAIImageJobStatusFailed
	openAIImageJobStatusTimeout   openAIImageJobStatus = service.OpenAIImageJobStatusTimeout
)

type openAIImageJobOwner = service.OpenAIImageJobOwner

type openAIImageJobRequest struct {
	Endpoint                   string
	ContentType                string
	RemoteAddr                 string
	Headers                    http.Header
	Body                       []byte
	SecurityAuditCompletion    securityAuditCompletionToken
	HasSecurityAuditCompletion bool
}

func (r openAIImageJobRequest) clone() openAIImageJobRequest {
	return openAIImageJobRequest{
		Endpoint:                   r.Endpoint,
		ContentType:                r.ContentType,
		RemoteAddr:                 r.RemoteAddr,
		Headers:                    r.Headers.Clone(),
		Body:                       append([]byte(nil), r.Body...),
		SecurityAuditCompletion:    r.SecurityAuditCompletion,
		HasSecurityAuditCompletion: r.HasSecurityAuditCompletion,
	}
}

type openAIImageJobResult struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (r openAIImageJobResult) clone() openAIImageJobResult {
	return openAIImageJobResult{
		StatusCode: r.StatusCode,
		Headers:    r.Headers.Clone(),
		Body:       append([]byte(nil), r.Body...),
	}
}

type openAIImageJobRunner func(context.Context, openAIImageJobRequest) openAIImageJobResult

type openAIImageJobStoreOptions struct {
	Concurrency  int
	Timeout      time.Duration
	TTL          time.Duration
	MaxAttempts  int
	RetryBackoff []time.Duration
}

type openAIImageJobSnapshot struct {
	ID          string
	Status      openAIImageJobStatus
	StatusCode  int
	Headers     http.Header
	Body        []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt time.Time
}

type openAIImageJobStore struct {
	sem          chan struct{}
	timeout      time.Duration
	ttl          time.Duration
	maxAttempts  int
	retryBackoff []time.Duration
	now          func() time.Time
	persistence  service.OpenAIImageJobStore
}

func newOpenAIImageJobStore(options openAIImageJobStoreOptions) *openAIImageJobStore {
	return newOpenAIImageJobStoreWithPersistence(nil, options)
}

func newOpenAIImageJobStoreWithPersistence(persistence service.OpenAIImageJobStore, options openAIImageJobStoreOptions) *openAIImageJobStore {
	concurrency := options.Concurrency
	if concurrency <= 0 {
		concurrency = openAIImageJobDefaultConcurrency
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = openAIImageJobDefaultTimeout
	}
	ttl := options.TTL
	if ttl <= 0 {
		ttl = openAIImageJobDefaultTTL
	}
	maxAttempts := options.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = openAIImageJobDefaultMaxAttempts
	}
	retryBackoff := append([]time.Duration(nil), options.RetryBackoff...)
	if len(retryBackoff) == 0 {
		retryBackoff = append([]time.Duration(nil), openAIImageJobDefaultRetryBackoff...)
	}
	if persistence == nil {
		persistence = newInMemoryOpenAIImageJobPersistence()
	}
	return &openAIImageJobStore{
		sem:          make(chan struct{}, concurrency),
		timeout:      timeout,
		ttl:          ttl,
		maxAttempts:  maxAttempts,
		retryBackoff: retryBackoff,
		now:          time.Now,
		persistence:  persistence,
	}
}

func buildOpenAIImageJobStoreOptions(cfg *config.Config) openAIImageJobStoreOptions {
	if cfg == nil {
		return openAIImageJobStoreOptions{}
	}
	jobCfg := cfg.Gateway.OpenAIImageJobs.WithDefaults()
	retryBackoff := make([]time.Duration, 0, len(jobCfg.RetryBackoffSeconds))
	for _, seconds := range jobCfg.RetryBackoffSeconds {
		retryBackoff = append(retryBackoff, time.Duration(seconds)*time.Second)
	}
	return openAIImageJobStoreOptions{
		Concurrency:  jobCfg.Concurrency,
		Timeout:      time.Duration(jobCfg.TimeoutSeconds) * time.Second,
		TTL:          time.Duration(jobCfg.TTLSeconds) * time.Second,
		MaxAttempts:  jobCfg.MaxAttempts,
		RetryBackoff: retryBackoff,
	}
}

func (s *openAIImageJobStore) submit(owner openAIImageJobOwner, req openAIImageJobRequest, runner openAIImageJobRunner) (*openAIImageJobSnapshot, error) {
	if s == nil {
		return nil, nil
	}
	now := s.now()
	job := &service.OpenAIImageJobRecord{
		ID:        "imgjob_" + uuid.NewString(),
		Owner:     owner,
		Status:    openAIImageJobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.persistence.Create(context.Background(), job, s.ttl); err != nil {
		return nil, err
	}

	go s.run(job.ID, req.clone(), runner)
	return openAIImageJobSnapshotFromRecord(job), nil
}

func (s *openAIImageJobStore) get(id string, owner openAIImageJobOwner) (*openAIImageJobSnapshot, bool) {
	if s == nil || s.persistence == nil || id == "" {
		return nil, false
	}
	job, ok, err := s.persistence.Get(context.Background(), id, owner)
	if err != nil || !ok {
		return nil, false
	}
	return openAIImageJobSnapshotFromRecord(job), true
}

func (s *openAIImageJobStore) run(id string, req openAIImageJobRequest, runner openAIImageJobRunner) {
	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	if !s.setRunning(id) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resultCh := make(chan openAIImageJobResult, 1)
	go func() {
		resultCh <- s.runWithRetry(ctx, req, runner)
	}()

	var result openAIImageJobResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		result = openAIImageJobTimeoutResult()
	}
	if result.StatusCode == 0 {
		result = openAIImageJobErrorResult(ctx.Err())
	}
	if len(result.Body) == 0 && result.StatusCode >= 400 {
		result.Body = openAIImageJobErrorBody("api_error", http.StatusText(result.StatusCode))
	}
	s.complete(id, result.clone())
}

func openAIImageJobErrorResult(err error) openAIImageJobResult {
	if err != nil {
		return openAIImageJobTimeoutResult()
	}
	return openAIImageJobResult{
		StatusCode: http.StatusInternalServerError,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       openAIImageJobErrorBody("api_error", "Image job failed"),
	}
}

func openAIImageJobTimeoutResult() openAIImageJobResult {
	return openAIImageJobResult{
		StatusCode: http.StatusGatewayTimeout,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       openAIImageJobErrorBody("api_error", "Image job timed out"),
	}
}

func (s *openAIImageJobStore) runWithRetry(ctx context.Context, req openAIImageJobRequest, runner openAIImageJobRunner) openAIImageJobResult {
	if runner == nil {
		return openAIImageJobResult{
			StatusCode: http.StatusInternalServerError,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       openAIImageJobErrorBody("api_error", "Image job runner is not configured"),
		}
	}

	maxAttempts := s.maxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	var result openAIImageJobResult
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return openAIImageJobResult{}
		}
		result = runOpenAIImageJobAttempt(ctx, req, runner)
		if !isOpenAIImageJobRetryableStatus(result.StatusCode) || attempt >= maxAttempts {
			break
		}
		if !s.waitBeforeOpenAIImageJobRetry(ctx, attempt) {
			return openAIImageJobResult{}
		}
	}
	if maxAttempts > 1 && isOpenAIImageJobRetryableStatus(result.StatusCode) {
		result.Headers = http.Header{"Content-Type": []string{"application/json"}}
		result.Body = openAIImageJobAttemptsExhaustedBody(result.StatusCode, maxAttempts)
	}
	return result
}

func runOpenAIImageJobAttempt(ctx context.Context, req openAIImageJobRequest, runner openAIImageJobRunner) (result openAIImageJobResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = openAIImageJobResult{
				StatusCode: http.StatusInternalServerError,
				Headers:    http.Header{"Content-Type": []string{"application/json"}},
				Body:       openAIImageJobErrorBody("api_error", "Image job failed"),
			}
		}
	}()
	return runner(ctx, req.clone())
}

func (s *openAIImageJobStore) waitBeforeOpenAIImageJobRetry(ctx context.Context, attempt int) bool {
	backoff := time.Duration(0)
	if attempt > 0 && attempt <= len(s.retryBackoff) {
		backoff = s.retryBackoff[attempt-1]
	}
	if backoff <= 0 {
		return ctx.Err() == nil
	}
	select {
	case <-ctx.Done():
		return false
	case <-time.After(backoff):
		return true
	}
}

func isOpenAIImageJobRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, 524:
		return true
	default:
		return false
	}
}

func (s *openAIImageJobStore) setRunning(id string) bool {
	if s == nil || s.persistence == nil {
		return false
	}
	now := s.now()
	ok, err := s.persistence.SetRunning(context.Background(), id, now, s.ttl)
	return err == nil && ok
}

func (s *openAIImageJobStore) complete(id string, result openAIImageJobResult) {
	if s == nil || s.persistence == nil {
		return
	}
	now := s.now()
	status := openAIImageJobStatusFailed
	if result.StatusCode >= 200 && result.StatusCode < 300 {
		status = openAIImageJobStatusSucceeded
	} else if result.StatusCode == http.StatusGatewayTimeout {
		status = openAIImageJobStatusTimeout
	}
	_ = s.persistence.Complete(context.Background(), id, status, result.StatusCode, result.Headers, result.Body, now, s.ttl)
}

func openAIImageJobSnapshotFromRecord(j *service.OpenAIImageJobRecord) *openAIImageJobSnapshot {
	if j == nil {
		return nil
	}
	return &openAIImageJobSnapshot{
		ID:          j.ID,
		Status:      j.Status,
		StatusCode:  j.StatusCode,
		Headers:     http.Header(j.Headers).Clone(),
		Body:        append([]byte(nil), j.Body...),
		CreatedAt:   j.CreatedAt,
		UpdatedAt:   j.UpdatedAt,
		CompletedAt: j.CompletedAt,
	}
}

type inMemoryOpenAIImageJobPersistence struct {
	mu   sync.RWMutex
	jobs map[string]*service.OpenAIImageJobRecord
}

func newInMemoryOpenAIImageJobPersistence() *inMemoryOpenAIImageJobPersistence {
	return &inMemoryOpenAIImageJobPersistence{jobs: make(map[string]*service.OpenAIImageJobRecord)}
}

func (s *inMemoryOpenAIImageJobPersistence) Create(_ context.Context, record *service.OpenAIImageJobRecord, _ time.Duration) error {
	if s == nil || record == nil || record.ID == "" {
		return nil
	}
	s.mu.Lock()
	s.jobs[record.ID] = cloneOpenAIImageJobRecord(record)
	s.mu.Unlock()
	return nil
}

func (s *inMemoryOpenAIImageJobPersistence) Get(_ context.Context, id string, owner service.OpenAIImageJobOwner) (*service.OpenAIImageJobRecord, bool, error) {
	if s == nil || id == "" {
		return nil, false, nil
	}
	s.mu.RLock()
	job, ok := s.jobs[id]
	if !ok || job.Owner.UserID != owner.UserID || job.Owner.APIKeyID != owner.APIKeyID {
		s.mu.RUnlock()
		return nil, false, nil
	}
	out := cloneOpenAIImageJobRecord(job)
	s.mu.RUnlock()
	return out, true, nil
}

func (s *inMemoryOpenAIImageJobPersistence) SetRunning(_ context.Context, id string, updatedAt time.Time, _ time.Duration) (bool, error) {
	if s == nil || id == "" {
		return false, nil
	}
	started := false
	s.mu.Lock()
	if job, ok := s.jobs[id]; ok && isOpenAIImageJobActiveStatus(job.Status) {
		job.Status = openAIImageJobStatusRunning
		job.UpdatedAt = updatedAt
		started = true
	}
	s.mu.Unlock()
	return started, nil
}

func (s *inMemoryOpenAIImageJobPersistence) Complete(
	_ context.Context,
	id string,
	status service.OpenAIImageJobStatus,
	statusCode int,
	headers map[string][]string,
	body []byte,
	completedAt time.Time,
	_ time.Duration,
) error {
	if s == nil || id == "" {
		return nil
	}
	s.mu.Lock()
	if job, ok := s.jobs[id]; ok && isOpenAIImageJobActiveStatus(job.Status) {
		job.Status = status
		job.StatusCode = statusCode
		job.Headers = http.Header(headers).Clone()
		job.Body = append([]byte(nil), body...)
		job.UpdatedAt = completedAt
		job.CompletedAt = completedAt
	}
	s.mu.Unlock()
	return nil
}

func (s *inMemoryOpenAIImageJobPersistence) MarkStaleTimeouts(_ context.Context, _ time.Time, _ time.Duration, _ int, _ time.Duration) (int64, error) {
	return 0, nil
}

func cloneOpenAIImageJobRecord(in *service.OpenAIImageJobRecord) *service.OpenAIImageJobRecord {
	if in == nil {
		return nil
	}
	out := *in
	out.Headers = http.Header(in.Headers).Clone()
	out.Body = append([]byte(nil), in.Body...)
	return &out
}

func isOpenAIImageJobActiveStatus(status service.OpenAIImageJobStatus) bool {
	return status == openAIImageJobStatusQueued || status == openAIImageJobStatusRunning
}

func openAIImageJobErrorBody(code string, message string) []byte {
	body, err := json.Marshal(map[string]any{
		"error": map[string]any{
			"type":    code,
			"message": message,
		},
	})
	if err != nil {
		return []byte(`{"error":{"type":"api_error","message":"Image job failed"}}`)
	}
	return body
}

func openAIImageJobAttemptsExhaustedBody(status int, attempts int) []byte {
	body, err := json.Marshal(map[string]any{
		"error": map[string]any{
			"type":        "upstream_error",
			"message":     fmt.Sprintf("Image job failed after %d attempts", attempts),
			"last_status": status,
			"attempts":    attempts,
		},
	})
	if err != nil {
		return openAIImageJobErrorBody("upstream_error", "Image job failed after retries")
	}
	return body
}

// ImageJobCreate submits a long-running OpenAI Images request and returns a
// pollable job id. The worker reuses Images so billing stays success-only.
func (h *OpenAIGatewayHandler) ImageJobCreate(c *gin.Context) {
	setOpenAIImageJobNoStoreHeaders(c)
	if h == nil {
		return
	}

	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	endpoint, ok := openAIImageJobEndpointFromPath(c.Request.URL.Path)
	if !ok {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "Image job endpoint not found")
		return
	}

	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.errorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}
	if !service.GroupAllowsImageGeneration(apiKey.Group) {
		h.errorResponse(c, http.StatusForbidden, "permission_error", service.ImageGenerationPermissionMessage())
		return
	}
	if h.gatewayService == nil {
		h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Image gateway is not available")
		return
	}

	parseCtx := c.Copy()
	parseRequest := c.Request.Clone(c.Request.Context())
	parseURL := *parseRequest.URL
	parseURL.Path = endpoint
	parseRequest.URL = &parseURL
	parseCtx.Request = parseRequest
	parsed, err := h.gatewayService.ParseOpenAIImagesRequest(parseCtx, body)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	moderationBody := parsed.ModerationBody()
	if len(moderationBody) == 0 {
		cacheSecurityAuditCompletion(c, apiKey, subject, service.ContentModerationProtocolOpenAIImages, parsed.Model, moderationBody)
	} else {
		reqLog := requestLogger(c, "handler.openai_image_job.security_audit",
			zap.Int64("user_id", subject.UserID), zap.Int64("api_key_id", apiKey.ID), zap.String("model", parsed.Model))
		if decision := h.checkSecurityAudit(c, reqLog, apiKey, subject, service.ContentModerationProtocolOpenAIImages, parsed.Model, moderationBody); decision != nil && !decision.AllowNextStage {
			h.openAISecurityAuditError(c, decision)
			return
		}
	}

	store := h.ensureImageJobStore()
	if store == nil {
		h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Image jobs are not available")
		return
	}

	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	productSettlement, _ := middleware2.GetProductSettlementFromContext(c)
	securityAuditCompletion, hasSecurityAuditCompletion := securityAuditCompletionFromContext(c)

	jobReq := openAIImageJobRequest{
		Endpoint:                   endpoint,
		ContentType:                c.GetHeader("Content-Type"),
		RemoteAddr:                 c.Request.RemoteAddr,
		Headers:                    c.Request.Header.Clone(),
		Body:                       append([]byte(nil), body...),
		SecurityAuditCompletion:    securityAuditCompletion,
		HasSecurityAuditCompletion: hasSecurityAuditCompletion,
	}
	owner := openAIImageJobOwner{UserID: subject.UserID, APIKeyID: apiKey.ID}
	job, err := store.submit(owner, jobReq, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return h.runOpenAIImageJob(ctx, req, apiKey, subject, subscription, productSettlement)
	})
	if err != nil {
		h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Image jobs are not available")
		return
	}
	if job == nil {
		h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Image jobs are not available")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"id":         job.ID,
		"job_id":     job.ID,
		"status":     string(job.Status),
		"created_at": job.CreatedAt.Unix(),
		"updated_at": job.UpdatedAt.Unix(),
	})
}

// ImageJobStatus returns the current state of an image job owned by the API key.
func (h *OpenAIGatewayHandler) ImageJobStatus(c *gin.Context) {
	setOpenAIImageJobNoStoreHeaders(c)
	if h == nil {
		return
	}
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	store := h.ensureImageJobStore()
	jobID := strings.TrimSpace(c.Param("id"))
	job, ok := store.get(jobID, openAIImageJobOwner{UserID: subject.UserID, APIKeyID: apiKey.ID})
	if !ok {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "Image job not found")
		return
	}
	c.JSON(http.StatusOK, openAIImageJobStatusPayload(job))
}

func setOpenAIImageJobNoStoreHeaders(c *gin.Context) {
	if c == nil {
		return
	}
	c.Header("Cache-Control", "no-store, no-cache, max-age=0, must-revalidate, private")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Surrogate-Control", "no-store")
	c.Header("X-Accel-Expires", "0")
}

func (h *OpenAIGatewayHandler) ensureImageJobStore() *openAIImageJobStore {
	if h == nil {
		return nil
	}
	if h.imageJobStore == nil {
		h.imageJobStore = newOpenAIImageJobStore(openAIImageJobStoreOptions{})
	}
	return h.imageJobStore
}

func (h *OpenAIGatewayHandler) runOpenAIImageJob(
	ctx context.Context,
	req openAIImageJobRequest,
	apiKey *service.APIKey,
	subject middleware2.AuthSubject,
	subscription *service.UserSubscription,
	productSettlement *service.ProductSettlementContext,
) openAIImageJobResult {
	recorder := httptest.NewRecorder()
	jobCtx, _ := gin.CreateTestContext(recorder)

	requestCtx := ctx
	if apiKey != nil {
		requestCtx = context.WithValue(requestCtx, ctxkey.APIKey, apiKey)
		if apiKey.Group != nil {
			requestCtx = context.WithValue(requestCtx, ctxkey.Group, apiKey.Group)
		}
	}
	if productSettlement != nil {
		requestCtx = service.ContextWithProductSettlement(requestCtx, productSettlement)
	}
	httpReq := httptest.NewRequest(http.MethodPost, req.Endpoint, bytes.NewReader(req.Body))
	httpReq = httpReq.WithContext(requestCtx)
	httpReq.Header = req.Headers.Clone()
	if strings.TrimSpace(req.ContentType) != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}
	httpReq.RemoteAddr = req.RemoteAddr
	jobCtx.Request = httpReq
	jobCtx.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(req.Endpoint))
	if apiKey != nil {
		jobCtx.Set(string(middleware2.ContextKeyAPIKey), apiKey)
		if apiKey.User != nil {
			jobCtx.Set(string(middleware2.ContextKeyUserRole), apiKey.User.Role)
		}
	}
	jobCtx.Set(string(middleware2.ContextKeyUser), subject)
	if subscription != nil {
		jobCtx.Set(string(middleware2.ContextKeySubscription), subscription)
	}
	if productSettlement != nil {
		jobCtx.Set(string(middleware2.ContextKeyProductSettlement), productSettlement)
	}
	applyOpenAIImageJobSecurityAuditCompletion(jobCtx, req)

	h.Images(jobCtx)

	statusCode := recorder.Code
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	body := append([]byte(nil), recorder.Body.Bytes()...)
	if statusCode >= 200 && statusCode < 300 && len(body) == 0 {
		statusCode = http.StatusBadGateway
		body = openAIImageJobErrorBody("api_error", "Image job produced no response")
	}
	return openAIImageJobResult{
		StatusCode: statusCode,
		Headers:    recorder.Header().Clone(),
		Body:       body,
	}
}

func applyOpenAIImageJobSecurityAuditCompletion(c *gin.Context, req openAIImageJobRequest) {
	if c != nil && req.HasSecurityAuditCompletion {
		c.Set(securityAuditCompletedContextKey, req.SecurityAuditCompletion)
	}
}

func openAIImageJobEndpointFromPath(path string) (string, bool) {
	normalized := strings.TrimRight(strings.TrimSpace(path), "/")
	switch {
	case strings.HasSuffix(normalized, "/images/jobs/generations"):
		return EndpointImagesGenerations, true
	case strings.HasSuffix(normalized, "/images/jobs/edits"):
		return EndpointImagesEdits, true
	default:
		return "", false
	}
}

func openAIImageJobStatusPayload(job *openAIImageJobSnapshot) gin.H {
	payload := gin.H{
		"id":         job.ID,
		"job_id":     job.ID,
		"status":     string(job.Status),
		"created_at": job.CreatedAt.Unix(),
		"updated_at": job.UpdatedAt.Unix(),
	}
	if !job.CompletedAt.IsZero() {
		payload["completed_at"] = job.CompletedAt.Unix()
	}
	if job.StatusCode > 0 {
		payload["http_status"] = job.StatusCode
	}
	switch job.Status {
	case openAIImageJobStatusSucceeded:
		payload["response"] = openAIImageJobJSONOrString(job.Body)
	case openAIImageJobStatusFailed, openAIImageJobStatusTimeout:
		payload["error"] = openAIImageJobErrorPayload(job.Body)
	}
	return payload
}

func openAIImageJobJSONOrString(body []byte) any {
	var value any
	if len(body) > 0 && json.Unmarshal(body, &value) == nil {
		return value
	}
	return string(body)
}

func openAIImageJobErrorPayload(body []byte) any {
	var payload map[string]any
	if len(body) > 0 && json.Unmarshal(body, &payload) == nil {
		if errPayload, ok := payload["error"]; ok {
			return errPayload
		}
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = "Image job failed"
	}
	return gin.H{
		"type":    "api_error",
		"message": message,
	}
}
