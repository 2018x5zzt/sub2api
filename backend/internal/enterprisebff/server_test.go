package enterprisebff

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProxyPreservesTraceHeadersAndTransformsUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var seenRequestID string
	var seenContractVersion string
	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		seenRequestID = r.Header.Get("X-Request-ID")
		seenContractVersion = r.Header.Get("X-Contract-Version")
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")
		_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"items":[{"id":1,"total_cost":0.34}]}}`))
		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, newFakeEnterpriseStore(), newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodGet, "/admin/usage", nil)
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("X-Contract-Version", "2026-04-10")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "req-123", seenRequestID)
	require.Equal(t, "2026-04-10", seenContractVersion)
	require.Contains(t, recorder.Body.String(), `"billable_cost":0.34`)
}

func TestTraceMiddlewareGeneratesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		require.NotEmpty(t, r.Header.Get("X-Request-ID"))
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(recorder, `{"code":0,"message":"success","data":{}}`)
		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, newFakeEnterpriseStore(), newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"a@example.com","password":"pw"}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
}

func TestEnterpriseLoginRejectsCompanyMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("upstream login should not be called when company does not match")
		return nil, nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.matchByEmail["owner@example.com"] = &EnterpriseProfile{
		Name:        "acme",
		DisplayName: "ACME Corp",
		UserID:      42,
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"company_name":"otherco","email":"owner@example.com","password":"pw"}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Invalid company credentials")
}

func TestAuthMeIncludesEnterpriseMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")
		_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":42,"email":"owner@example.com","username":"owner","role":"user","balance":12.5,"concurrency":3,"status":"active"}}`))
		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[42] = &EnterpriseProfile{
		Name:        "acme",
		DisplayName: "ACME Corp",
		SupportInfo: "ops@acme.test",
		UserID:      42,
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	data := payload["data"].(map[string]any)
	require.Equal(t, "acme", data["enterprise_name"])
	require.Equal(t, "ACME Corp", data["enterprise_display_name"])
	require.Equal(t, "ops@acme.test", data["enterprise_support_contact"])
}

func TestPublicSettingsUseEnterpriseBrandingForAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/me":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"id":42,"email":"owner@example.com","username":"owner","role":"user","balance":12.5,"concurrency":3,"status":"active"}}`))
		case "/api/v1/settings/public":
			_, _ = recorder.Write([]byte(`{"code":0,"message":"success","data":{"site_name":"Bus2API","contact_info":"default"}}`))
		default:
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}

		return recorder.Result(), nil
	})

	baseURL, err := url.Parse("http://core.example/api/v1")
	require.NoError(t, err)

	store := newFakeEnterpriseStore()
	store.profiles[42] = &EnterpriseProfile{
		Name:        "acme",
		DisplayName: "ACME Corp",
		SupportInfo: "ops@acme.test",
		UserID:      42,
	}

	server := New(&Config{
		ListenAddr:     "127.0.0.1:0",
		CoreBaseURL:    baseURL,
		RequestTimeout: 0,
	}, nil, store, newNoopGroupHealthSnapshotRepo())
	server.httpClient = &http.Client{Transport: transport}

	req := httptest.NewRequest(http.MethodGet, "/settings/public", nil)
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	server.Router().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	data := payload["data"].(map[string]any)
	require.Equal(t, "ACME Corp", data["site_name"])
	require.Equal(t, "acme", data["enterprise_name"])
	require.Equal(t, "ops@acme.test", data["contact_info"])
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type visibleGroupSeed struct {
	ID       int64
	Name     string
	Platform string
}

type fakeEnterpriseStore struct {
	matchByEmail  map[string]*EnterpriseProfile
	profiles      map[int64]*EnterpriseProfile
	visibleGroups map[int64][]visibleGroupSeed
}

func newFakeEnterpriseStore() *fakeEnterpriseStore {
	return &fakeEnterpriseStore{
		matchByEmail:  map[string]*EnterpriseProfile{},
		profiles:      map[int64]*EnterpriseProfile{},
		visibleGroups: map[int64][]visibleGroupSeed{},
	}
}

func (s *fakeEnterpriseStore) MatchUserByEmailAndCompany(_ context.Context, email, companyName string) (*EnterpriseProfile, error) {
	profile, ok := s.matchByEmail[email]
	if !ok || normalizeCompanyName(profile.Name) != normalizeCompanyName(companyName) {
		return nil, nil
	}
	return profile, nil
}

func (s *fakeEnterpriseStore) GetByUserID(_ context.Context, userID int64) (*EnterpriseProfile, error) {
	return s.profiles[userID], nil
}

func (s *fakeEnterpriseStore) ListVisibleGroups(_ context.Context, userID int64) ([]EnterpriseVisibleGroup, error) {
	seeds := s.visibleGroups[userID]
	out := make([]EnterpriseVisibleGroup, 0, len(seeds))
	for _, seed := range seeds {
		out = append(out, EnterpriseVisibleGroup{
			ID:       seed.ID,
			Name:     seed.Name,
			Platform: seed.Platform,
		})
	}
	return out, nil
}

func (s *fakeEnterpriseStore) SameEnterprise(_ context.Context, actorUserID, targetUserID int64) (bool, error) {
	actor := s.profiles[actorUserID]
	target := s.profiles[targetUserID]
	if actor == nil || target == nil {
		return false, nil
	}
	return normalizeCompanyName(actor.Name) == normalizeCompanyName(target.Name), nil
}

type noopGroupHealthSnapshotRepo struct{}

func newNoopGroupHealthSnapshotRepo() *noopGroupHealthSnapshotRepo {
	return &noopGroupHealthSnapshotRepo{}
}

func (noopGroupHealthSnapshotRepo) UpsertBatch(context.Context, []service.GroupHealthSnapshot) error {
	return nil
}

func (noopGroupHealthSnapshotRepo) ListRecentByGroupIDs(context.Context, []int64, time.Time) (map[int64][]service.GroupHealthSnapshot, error) {
	return map[int64][]service.GroupHealthSnapshot{}, nil
}

func (noopGroupHealthSnapshotRepo) DeleteBefore(context.Context, time.Time) (int, error) {
	return 0, nil
}
