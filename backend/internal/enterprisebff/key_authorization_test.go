package enterprisebff

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEnterpriseUserKeyCreateRejectsUnauthorizedGroupID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var upstreamCreateCalled bool
	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":42,"email":"owner@example.com","username":"owner","role":"user","balance":12.5,"concurrency":3,"status":"active"}}`))
		case "/api/v1/keys":
			upstreamCreateCalled = true
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":11,"user_id":42,"key":"sk-test","name":"demo","group_id":1,"status":"active","quota":0,"quota_used":0,"created_at":"2026-04-10T12:00:00Z","updated_at":"2026-04-10T12:00:00Z"}}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[42] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 42}
	store.visibleGroups[42] = []visibleGroupSeed{
		{ID: 2, Name: "enterprise-private", Platform: "openai"},
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodPost, "/keys", jsonBody(t, map[string]any{
		"name":     "demo",
		"group_id": 1,
	}))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "无权使用该号池")
	require.False(t, upstreamCreateCalled)
}

func TestEnterpriseUserKeyUpdateRejectsGroupRebind(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var upstreamUpdateCalled bool
	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":42,"email":"owner@example.com","username":"owner","role":"user","balance":12.5,"concurrency":3,"status":"active"}}`))
		case "/api/v1/keys/11":
			upstreamUpdateCalled = true
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":11,"user_id":42,"key":"sk-test","name":"demo","group_id":2,"status":"active","quota":0,"quota_used":0,"created_at":"2026-04-10T12:00:00Z","updated_at":"2026-04-10T12:00:00Z"}}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[42] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 42}
	store.visibleGroups[42] = []visibleGroupSeed{
		{ID: 2, Name: "enterprise-private", Platform: "openai"},
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodPatch, "/keys/11", jsonBody(t, map[string]any{
		"group_id": 2,
	}))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "现有 Key 不支持修改绑定号池")
	require.False(t, upstreamUpdateCalled)
}

func TestEnterpriseAdminKeyCreateRejectsCrossEnterpriseTargetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":7,"email":"admin@example.com","username":"admin","role":"admin","balance":12.5,"concurrency":3,"status":"active"}}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[7] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 7}
	store.profiles[99] = &EnterpriseProfile{Name: "otherco", DisplayName: "OtherCo", UserID: 99}
	store.visibleGroups[99] = []visibleGroupSeed{
		{ID: 5, Name: "otherco-private", Platform: "openai"},
	}

	adminStore := newFakeAdminKeyStore()
	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, adminStore, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/keys", jsonBody(t, map[string]any{
		"name":     "demo",
		"user_id":  99,
		"group_id": 5,
	}))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "无权使用该号池")
	require.Len(t, adminStore.createCalls, 0)
}

func TestEnterpriseAdminKeyUpdateAllowsNilGroupIDButRejectsUnauthorizedRebind(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":7,"email":"admin@example.com","username":"admin","role":"admin","balance":12.5,"concurrency":3,"status":"active"}}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[7] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 7}
	store.profiles[42] = &EnterpriseProfile{Name: "acme", DisplayName: "ACME", UserID: 42}
	store.visibleGroups[42] = []visibleGroupSeed{
		{ID: 2, Name: "enterprise-private", Platform: "openai"},
	}

	adminStore := newFakeAdminKeyStore()
	adminStore.keys[123] = &service.APIKey{
		ID:     123,
		UserID: 42,
		Key:    "sk-test",
		Name:   "demo",
		Status: "active",
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, adminStore, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	nilGroupReq := httptest.NewRequest(http.MethodPatch, "/v1/admin/keys/123", jsonBody(t, map[string]any{
		"group_id": nil,
	}))
	nilGroupReq.Header.Set("Authorization", "Bearer token")
	nilGroupReq.Header.Set("Content-Type", "application/json")

	nilGroupRecorder := httptest.NewRecorder()
	server.Router().ServeHTTP(nilGroupRecorder, nilGroupReq)

	require.Equal(t, http.StatusOK, nilGroupRecorder.Code)
	require.Len(t, adminStore.updateCalls, 1)
	require.Nil(t, adminStore.updateCalls[0].GroupID)

	rebindReq := httptest.NewRequest(http.MethodPatch, "/v1/admin/keys/123", jsonBody(t, map[string]any{
		"group_id": 9,
	}))
	rebindReq.Header.Set("Authorization", "Bearer token")
	rebindReq.Header.Set("Content-Type", "application/json")

	rebindRecorder := httptest.NewRecorder()
	server.Router().ServeHTTP(rebindRecorder, rebindReq)

	require.Equal(t, http.StatusForbidden, rebindRecorder.Code)
	require.Contains(t, rebindRecorder.Body.String(), "无权使用该号池")
	require.Len(t, adminStore.updateCalls, 1)
}

type fakeAdminKeyStore struct {
	keys        map[int64]*service.APIKey
	createCalls []fakeAdminKeyCreateCall
	updateCalls []service.UpdateAPIKeyRequest
}

type fakeAdminKeyCreateCall struct {
	OwnerID int64
	Req     service.CreateAPIKeyRequest
}

func newFakeAdminKeyStore() *fakeAdminKeyStore {
	return &fakeAdminKeyStore{
		keys: map[int64]*service.APIKey{},
	}
}

func (s *fakeAdminKeyStore) List(context.Context, pagination.PaginationParams, AdminKeyListFilters) ([]service.APIKey, int64, error) {
	return nil, 0, nil
}

func (s *fakeAdminKeyStore) Get(_ context.Context, id int64) (*service.APIKey, error) {
	return s.keys[id], nil
}

func (s *fakeAdminKeyStore) Create(_ context.Context, ownerID int64, req service.CreateAPIKeyRequest) (*service.APIKey, error) {
	s.createCalls = append(s.createCalls, fakeAdminKeyCreateCall{
		OwnerID: ownerID,
		Req:     req,
	})
	key := &service.APIKey{
		ID:      int64(len(s.createCalls)),
		UserID:  ownerID,
		Key:     "sk-created",
		Name:    req.Name,
		GroupID: req.GroupID,
		Status:  "active",
	}
	return key, nil
}

func (s *fakeAdminKeyStore) Update(_ context.Context, id int64, req service.UpdateAPIKeyRequest) (*service.APIKey, error) {
	s.updateCalls = append(s.updateCalls, req)
	key := s.keys[id]
	if key == nil {
		key = &service.APIKey{ID: id, Status: "active"}
	}
	key.GroupID = req.GroupID
	return key, nil
}

func (s *fakeAdminKeyStore) Delete(context.Context, int64) error {
	return nil
}

func jsonBody(t *testing.T, value any) *bytes.Buffer {
	t.Helper()

	body, err := json.Marshal(value)
	require.NoError(t, err)
	return bytes.NewBuffer(body)
}
