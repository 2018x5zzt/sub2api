package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	openAIImageJobDefaultConcurrency = 2
	openAIImageJobDefaultTimeout     = 10 * time.Minute
	openAIImageJobDefaultTTL         = 24 * time.Hour
)

type openAIImageJobStatus string

const (
	openAIImageJobStatusQueued    openAIImageJobStatus = "queued"
	openAIImageJobStatusRunning   openAIImageJobStatus = "running"
	openAIImageJobStatusSucceeded openAIImageJobStatus = "succeeded"
	openAIImageJobStatusFailed    openAIImageJobStatus = "failed"
)

type openAIImageJobOwner struct {
	UserID   int64
	APIKeyID int64
}

type openAIImageJobRequest struct {
	Endpoint    string
	ContentType string
	RemoteAddr  string
	Headers     http.Header
	Body        []byte
}

func (r openAIImageJobRequest) clone() openAIImageJobRequest {
	return openAIImageJobRequest{
		Endpoint:    r.Endpoint,
		ContentType: r.ContentType,
		RemoteAddr:  r.RemoteAddr,
		Headers:     r.Headers.Clone(),
		Body:        append([]byte(nil), r.Body...),
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
	Concurrency int
	Timeout     time.Duration
	TTL         time.Duration
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

type openAIImageJob struct {
	id          string
	owner       openAIImageJobOwner
	status      openAIImageJobStatus
	statusCode  int
	headers     http.Header
	body        []byte
	createdAt   time.Time
	updatedAt   time.Time
	completedAt time.Time
}

type openAIImageJobStore struct {
	mu      sync.RWMutex
	jobs    map[string]*openAIImageJob
	sem     chan struct{}
	timeout time.Duration
	ttl     time.Duration
	now     func() time.Time
}

func newOpenAIImageJobStore(options openAIImageJobStoreOptions) *openAIImageJobStore {
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
	return &openAIImageJobStore{
		jobs:    make(map[string]*openAIImageJob),
		sem:     make(chan struct{}, concurrency),
		timeout: timeout,
		ttl:     ttl,
		now:     time.Now,
	}
}

func (s *openAIImageJobStore) submit(owner openAIImageJobOwner, req openAIImageJobRequest, runner openAIImageJobRunner) *openAIImageJobSnapshot {
	if s == nil {
		return nil
	}
	now := s.now()
	job := &openAIImageJob{
		id:        "imgjob_" + uuid.NewString(),
		owner:     owner,
		status:    openAIImageJobStatusQueued,
		createdAt: now,
		updatedAt: now,
	}

	s.mu.Lock()
	s.cleanupExpiredLocked(now)
	s.jobs[job.id] = job
	snapshot := job.snapshotLocked()
	s.mu.Unlock()

	go s.run(job.id, req.clone(), runner)
	return snapshot
}

func (s *openAIImageJobStore) get(id string, owner openAIImageJobOwner) (*openAIImageJobSnapshot, bool) {
	if s == nil || id == "" {
		return nil, false
	}
	now := s.now()
	s.mu.Lock()
	s.cleanupExpiredLocked(now)
	job, ok := s.jobs[id]
	if !ok || job.owner.UserID != owner.UserID || job.owner.APIKeyID != owner.APIKeyID {
		s.mu.Unlock()
		return nil, false
	}
	snapshot := job.snapshotLocked()
	s.mu.Unlock()
	return snapshot, true
}

func (s *openAIImageJobStore) run(id string, req openAIImageJobRequest, runner openAIImageJobRunner) {
	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	s.setRunning(id)

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	var result openAIImageJobResult
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				result = openAIImageJobResult{
					StatusCode: http.StatusInternalServerError,
					Headers:    http.Header{"Content-Type": []string{"application/json"}},
					Body:       openAIImageJobErrorBody("api_error", "Image job failed"),
				}
			}
		}()
		if runner == nil {
			result = openAIImageJobResult{
				StatusCode: http.StatusInternalServerError,
				Headers:    http.Header{"Content-Type": []string{"application/json"}},
				Body:       openAIImageJobErrorBody("api_error", "Image job runner is not configured"),
			}
			return
		}
		result = runner(ctx, req.clone())
	}()

	if result.StatusCode == 0 {
		status := http.StatusInternalServerError
		message := "Image job failed"
		if ctx.Err() != nil {
			status = http.StatusGatewayTimeout
			message = "Image job timed out"
		}
		result = openAIImageJobResult{
			StatusCode: status,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       openAIImageJobErrorBody("api_error", message),
		}
	}
	if len(result.Body) == 0 && result.StatusCode >= 400 {
		result.Body = openAIImageJobErrorBody("api_error", http.StatusText(result.StatusCode))
	}
	s.complete(id, result.clone())
}

func (s *openAIImageJobStore) setRunning(id string) {
	now := s.now()
	s.mu.Lock()
	if job, ok := s.jobs[id]; ok {
		job.status = openAIImageJobStatusRunning
		job.updatedAt = now
	}
	s.mu.Unlock()
}

func (s *openAIImageJobStore) complete(id string, result openAIImageJobResult) {
	now := s.now()
	s.mu.Lock()
	if job, ok := s.jobs[id]; ok {
		job.status = openAIImageJobStatusFailed
		if result.StatusCode >= 200 && result.StatusCode < 300 {
			job.status = openAIImageJobStatusSucceeded
		}
		job.statusCode = result.StatusCode
		job.headers = result.Headers.Clone()
		job.body = append([]byte(nil), result.Body...)
		job.updatedAt = now
		job.completedAt = now
	}
	s.mu.Unlock()
}

func (s *openAIImageJobStore) cleanupExpiredLocked(now time.Time) {
	if s.ttl <= 0 {
		return
	}
	for id, job := range s.jobs {
		if now.Sub(job.updatedAt) > s.ttl {
			delete(s.jobs, id)
		}
	}
}

func (j *openAIImageJob) snapshotLocked() *openAIImageJobSnapshot {
	if j == nil {
		return nil
	}
	return &openAIImageJobSnapshot{
		ID:          j.id,
		Status:      j.status,
		StatusCode:  j.statusCode,
		Headers:     j.headers.Clone(),
		Body:        append([]byte(nil), j.body...),
		CreatedAt:   j.createdAt,
		UpdatedAt:   j.updatedAt,
		CompletedAt: j.completedAt,
	}
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

// ImageJobCreate submits a long-running OpenAI Images request and returns a
// pollable job id. The worker reuses Images so billing stays success-only.
func (h *OpenAIGatewayHandler) ImageJobCreate(c *gin.Context) {
	if h == nil {
		return
	}
	store := h.ensureImageJobStore()
	if store == nil {
		h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Image jobs are not available")
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

	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	productSettlement := GetProductSettlement(c)
	jobReq := openAIImageJobRequest{
		Endpoint:    endpoint,
		ContentType: c.GetHeader("Content-Type"),
		RemoteAddr:  c.Request.RemoteAddr,
		Headers:     c.Request.Header.Clone(),
		Body:        append([]byte(nil), body...),
	}
	owner := openAIImageJobOwner{UserID: subject.UserID, APIKeyID: apiKey.ID}
	job := store.submit(owner, jobReq, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return h.runOpenAIImageJob(ctx, req, apiKey, subject, subscription, productSettlement)
	})
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
	case openAIImageJobStatusFailed:
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
