# Xlab Shell Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the first xlab backend boundary by adding a small read-only xlab service that authenticates with core JWTs, proxies product subscription reads to the current core, and lets frontend-v2 call `/xapi/v1` through an adapter.

**Architecture:** Phase 1 does not migrate data or payment fulfillment. The new `xlab-backend` is a separate Go service that verifies user identity by calling the current core API, then forwards product subscription read endpoints to core using the caller token. `frontend-v2` gets an `xlabClient` and routes subscription product reads through `src/api/xlab/*`, while the existing core endpoints remain available as fallback during rollout.

**Tech Stack:** Go 1.26.2 for `xlab-backend`, standard `net/http`, React/Vite `frontend-v2`, Axios, Vitest, Docker Compose for optional local wiring.

---

## File Structure

- Create `xlab-backend/go.mod`
  - Independent module for the xlab service so it can evolve outside upstream core.
- Create `xlab-backend/cmd/server/main.go`
  - Starts the xlab HTTP service.
- Create `xlab-backend/internal/config/config.go`
  - Reads `XLAB_SERVER_ADDR`, `CORE_API_BASE_URL`, and request timeout env vars.
- Create `xlab-backend/internal/core/client.go`
  - Thin HTTP client for verifying current user and proxying subscription products.
- Create `xlab-backend/internal/httpapi/router.go`
  - Registers `/health` and `/xapi/v1/subscription-products/*` routes.
- Create `xlab-backend/internal/httpapi/auth.go`
  - Bearer token extraction and core identity verification middleware.
- Create `xlab-backend/internal/httpapi/subscription_products.go`
  - Read-only product subscription proxy handlers.
- Create tests next to each package.
- Create `frontend-v2/src/api/xlabClient.ts`
  - Axios client for `/xapi/v1`, matching core client token and error behavior.
- Create `frontend-v2/src/api/xlab/subscriptionProducts.ts`
  - Xlab product subscription API wrapper.
- Modify `frontend-v2/src/api/subscriptionProducts.ts`
  - Route existing consumers through the xlab API wrapper while preserving exported API shape.
- Modify or add frontend-v2 tests for the adapter.
- Create `docs/superpowers/specs/2026-06-07-xlab-shell-phase1-runtime.md`
  - Runtime wiring notes for reverse proxy and local deployment.

This phase must not modify existing core product subscription services, core migrations, payment fulfillment, or API key authorization logic.

---

### Task 1: Scaffold xlab-backend service and config

**Files:**
- Create: `xlab-backend/go.mod`
- Create: `xlab-backend/cmd/server/main.go`
- Create: `xlab-backend/internal/config/config.go`
- Test: `xlab-backend/internal/config/config_test.go`

- [ ] **Step 1: Write config tests**

Create `xlab-backend/internal/config/config_test.go`:

```go
package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("XLAB_SERVER_ADDR", "")
	t.Setenv("CORE_API_BASE_URL", "")
	t.Setenv("XLAB_CORE_TIMEOUT_SECONDS", "")

	cfg := Load()
	if cfg.ServerAddr != ":8090" {
		t.Fatalf("ServerAddr = %q, want :8090", cfg.ServerAddr)
	}
	if cfg.CoreAPIBaseURL != "http://127.0.0.1:8080/api/v1" {
		t.Fatalf("CoreAPIBaseURL = %q", cfg.CoreAPIBaseURL)
	}
	if cfg.CoreTimeout != 10*time.Second {
		t.Fatalf("CoreTimeout = %s", cfg.CoreTimeout)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("XLAB_SERVER_ADDR", ":19090")
	t.Setenv("CORE_API_BASE_URL", "https://core.example.com/api/v1/")
	t.Setenv("XLAB_CORE_TIMEOUT_SECONDS", "3")

	cfg := Load()
	if cfg.ServerAddr != ":19090" {
		t.Fatalf("ServerAddr = %q", cfg.ServerAddr)
	}
	if cfg.CoreAPIBaseURL != "https://core.example.com/api/v1" {
		t.Fatalf("CoreAPIBaseURL = %q", cfg.CoreAPIBaseURL)
	}
	if cfg.CoreTimeout != 3*time.Second {
		t.Fatalf("CoreTimeout = %s", cfg.CoreTimeout)
	}
}
```

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
cd /root/sub2api-src/xlab-backend
go test ./internal/config
```

Expected: FAIL because `xlab-backend` and `Load` do not exist yet.

- [ ] **Step 3: Add module and config implementation**

Create `xlab-backend/go.mod`:

```go
module github.com/2018x5zzt/xlab-backend

go 1.26.2
```

Create `xlab-backend/internal/config/config.go`:

```go
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr     string
	CoreAPIBaseURL string
	CoreTimeout    time.Duration
}

func Load() Config {
	addr := strings.TrimSpace(os.Getenv("XLAB_SERVER_ADDR"))
	if addr == "" {
		addr = ":8090"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("CORE_API_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080/api/v1"
	}

	timeoutSeconds := 10
	if raw := strings.TrimSpace(os.Getenv("XLAB_CORE_TIMEOUT_SECONDS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	return Config{
		ServerAddr:     addr,
		CoreAPIBaseURL: baseURL,
		CoreTimeout:    time.Duration(timeoutSeconds) * time.Second,
	}
}
```

Create `xlab-backend/cmd/server/main.go`:

```go
package main

import (
	"log"
	"net/http"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/core"
	"github.com/2018x5zzt/xlab-backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	client := core.NewClient(cfg.CoreAPIBaseURL, cfg.CoreTimeout)
	router := httpapi.NewRouter(client)

	log.Printf("xlab backend listening on %s, core=%s", cfg.ServerAddr, cfg.CoreAPIBaseURL)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 4: Run config tests to verify GREEN**

Run:

```bash
cd /root/sub2api-src/xlab-backend
go test ./internal/config
```

Expected: PASS.

- [ ] **Step 5: Commit scaffold**

Run:

```bash
git add xlab-backend/go.mod xlab-backend/cmd/server/main.go xlab-backend/internal/config
git commit -m "$(cat <<'EOF'
feat(xlab): scaffold backend service

Add a minimal standalone xlab backend entrypoint and environment-based core API configuration.
EOF
)"
```

Expected: commit succeeds.

---

### Task 2: Add core API client and auth verification

**Files:**
- Create: `xlab-backend/internal/core/client.go`
- Test: `xlab-backend/internal/core/client_test.go`

- [ ] **Step 1: Write core client tests**

Create `xlab-backend/internal/core/client_test.go`:

```go
package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCurrentUserForwardsBearerTokenAndUnwrapsCoreEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user/profile" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"id":42,"email":"u@example.com","role":"user"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/api/v1", time.Second)
	user, err := client.CurrentUser(context.Background(), "token-123")
	if err != nil {
		t.Fatalf("CurrentUser error: %v", err)
	}
	if user.ID != 42 || user.Email != "u@example.com" || user.Role != "user" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestCurrentUserRejectsCoreError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"message":"bad token"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/api/v1", time.Second)
	_, err := client.CurrentUser(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
cd /root/sub2api-src/xlab-backend
go test ./internal/core
```

Expected: FAIL because `NewClient` does not exist.

- [ ] **Step 3: Implement core client**

Create `xlab-backend/internal/core/client.go`:

```go
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Envelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) CurrentUser(ctx context.Context, token string) (*User, error) {
	var user User
	if err := c.getEnvelope(ctx, token, "/user/profile", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.getEnvelope(ctx, token, path, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (c *Client) getEnvelope(ctx context.Context, token string, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("core status %d: %s", resp.StatusCode, string(body))
	}

	var env Envelope[json.RawMessage]
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&env); err != nil {
		return err
	}
	if env.Code != 0 {
		return fmt.Errorf("core code %d: %s", env.Code, env.Message)
	}
	return json.Unmarshal(env.Data, out)
}
```

- [ ] **Step 4: Run core tests to verify GREEN**

Run:

```bash
cd /root/sub2api-src/xlab-backend
go test ./internal/core
```

Expected: PASS.

- [ ] **Step 5: Commit core client**

Run:

```bash
git add xlab-backend/internal/core
git commit -m "$(cat <<'EOF'
feat(xlab): verify users through core API

Add a core API client that validates bearer tokens and proxies read-only core responses for xlab endpoints.
EOF
)"
```

Expected: commit succeeds.

---

### Task 3: Add xlab HTTP API routes and subscription product proxy

**Files:**
- Create: `xlab-backend/internal/httpapi/router.go`
- Create: `xlab-backend/internal/httpapi/auth.go`
- Create: `xlab-backend/internal/httpapi/subscription_products.go`
- Test: `xlab-backend/internal/httpapi/router_test.go`

- [ ] **Step 1: Write router tests**

Create `xlab-backend/internal/httpapi/router_test.go`:

```go
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type fakeCoreClient struct {
	userCalls int
	getPath   string
	token     string
}

func (f *fakeCoreClient) CurrentUser(ctx context.Context, token string) (*core.User, error) {
	f.userCalls++
	f.token = token
	return &core.User{ID: 7, Email: "u@example.com", Role: "user"}, nil
}

func (f *fakeCoreClient) ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error) {
	f.token = token
	f.getPath = path
	return json.RawMessage(`[{"subscription_id":99,"name":"Pro"}]`), nil
}

func TestHealth(t *testing.T) {
	r := NewRouter(&fakeCoreClient{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsRequireBearer(t *testing.T) {
	r := NewRouter(&fakeCoreClient{})
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSubscriptionProductsProxyActive(t *testing.T) {
	fake := &fakeCoreClient{}
	r := NewRouter(fake)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if fake.getPath != "/subscription-products/active" {
		t.Fatalf("proxied path = %s", fake.getPath)
	}
	if fake.token != "abc" {
		t.Fatalf("token = %s", fake.token)
	}
	if !strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
cd /root/sub2api-src/xlab-backend
go test ./internal/httpapi
```

Expected: FAIL because `NewRouter` does not exist.

- [ ] **Step 3: Implement HTTP API**

Create `xlab-backend/internal/httpapi/router.go`:

```go
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type CoreClient interface {
	CurrentUser(ctx context.Context, token string) (*core.User, error)
	ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error)
}

func NewRouter(coreClient CoreClient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	api := &API{core: coreClient}
	mux.Handle("/xapi/v1/subscription-products/active", api.auth(http.HandlerFunc(api.getActiveProducts)))
	mux.Handle("/xapi/v1/subscription-products/summary", api.auth(http.HandlerFunc(api.getProductSummary)))
	mux.Handle("/xapi/v1/subscription-products/progress", api.auth(http.HandlerFunc(api.getProductProgress)))
	return mux
}
```

Create `xlab-backend/internal/httpapi/auth.go`:

```go
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type contextKey string

const tokenContextKey contextKey = "token"

type API struct {
	core CoreClient
}

func (a *API) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
			return
		}
		if _, err := a.core.CurrentUser(r.Context(), token); err != nil {
			writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), tokenContextKey, token)))
	})
}

func bearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func tokenFromContext(ctx context.Context) string {
	value, _ := ctx.Value(tokenContextKey).(string)
	return value
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data})
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": message})
}

var _ = core.User{}
```

Create `xlab-backend/internal/httpapi/subscription_products.go`:

```go
package httpapi

import "net/http"

func (a *API) getActiveProducts(w http.ResponseWriter, r *http.Request) {
	a.proxyCoreJSON(w, r, "/subscription-products/active")
}

func (a *API) getProductSummary(w http.ResponseWriter, r *http.Request) {
	a.proxyCoreJSON(w, r, "/subscription-products/summary")
}

func (a *API) getProductProgress(w http.ResponseWriter, r *http.Request) {
	a.proxyCoreJSON(w, r, "/subscription-products/progress")
}

func (a *API) proxyCoreJSON(w http.ResponseWriter, r *http.Request, path string) {
	raw, err := a.core.ProxyGET(r.Context(), tokenFromContext(r.Context()), path)
	if err != nil {
		writeError(w, http.StatusBadGateway, "CORE_UNAVAILABLE", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, raw)
}
```

- [ ] **Step 4: Fix imports and run tests to verify GREEN**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS. If compile errors mention missing `context`, `encoding/json`, or `core`, add those imports to `router.go` exactly where used.

- [ ] **Step 5: Commit HTTP API**

Run:

```bash
git add xlab-backend/internal/httpapi
git commit -m "$(cat <<'EOF'
feat(xlab): proxy subscription product reads

Expose read-only xlab subscription product endpoints that authenticate via core and proxy current core responses.
EOF
)"
```

Expected: commit succeeds.

---

### Task 4: Add frontend-v2 xlab API adapter

**Files:**
- Create: `frontend-v2/src/api/xlabClient.ts`
- Create: `frontend-v2/src/api/xlab/subscriptionProducts.ts`
- Modify: `frontend-v2/src/api/subscriptionProducts.ts`
- Test: `frontend-v2/src/api/xlab/__tests__/subscriptionProducts.spec.ts`

- [ ] **Step 1: Write frontend adapter test**

Create `frontend-v2/src/api/xlab/__tests__/subscriptionProducts.spec.ts`:

```ts
import { describe, expect, it, vi } from 'vitest'

vi.mock('../xlabClient', () => ({
  xlabClient: {
    get: vi.fn(async (path: string) => ({ data: path }))
  }
}))

import { xlabClient } from '../xlabClient'
import { getActive, getProgress, getSummary } from '../subscriptionProducts'

describe('xlab subscriptionProducts API', () => {
  it('uses xlab endpoints for product subscription reads', async () => {
    await expect(getActive()).resolves.toBe('/subscription-products/active')
    await expect(getSummary()).resolves.toBe('/subscription-products/summary')
    await expect(getProgress()).resolves.toBe('/subscription-products/progress')
    expect(xlabClient.get).toHaveBeenCalledTimes(3)
  })
})
```

- [ ] **Step 2: Run test to verify RED**

Run:

```bash
npm --prefix frontend-v2 exec -- vitest --root frontend-v2 run src/api/xlab/__tests__/subscriptionProducts.spec.ts
```

Expected: FAIL because `src/api/xlab/subscriptionProducts.ts` does not exist.

- [ ] **Step 3: Implement xlab client and adapter**

Create `frontend-v2/src/api/xlabClient.ts`:

```ts
import axios, { type AxiosError, type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import type { ApiResponse } from '@/types'
import { getLocale } from '@/i18n'

const XLAB_API_BASE_URL = import.meta.env.VITE_XLAB_API_BASE_URL || '/xapi/v1'

export const xlabClient: AxiosInstance = axios.create({
  baseURL: XLAB_API_BASE_URL,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

xlabClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('auth_token')
  if (token && config.headers) config.headers.Authorization = `Bearer ${token}`
  if (config.headers) config.headers['Accept-Language'] = getLocale()
  return config
})

xlabClient.interceptors.response.use(
  (response: AxiosResponse) => {
    const apiResponse = response.data as ApiResponse<unknown>
    if (apiResponse && typeof apiResponse === 'object' && 'code' in apiResponse) {
      if (apiResponse.code === 0) {
        response.data = apiResponse.data
      } else {
        return Promise.reject({ status: response.status, code: apiResponse.code, message: apiResponse.message || 'Unknown error' })
      }
    }
    return response
  },
  (error: AxiosError<ApiResponse<unknown>>) => {
    const data = error.response?.data
    return Promise.reject({
      status: error.response?.status || 0,
      code: data?.code,
      message: data?.message || error.message || 'Network error'
    })
  }
)

export default xlabClient
```

Create `frontend-v2/src/api/xlab/subscriptionProducts.ts`:

```ts
import { xlabClient } from '../xlabClient'
import type { ActiveSubscriptionProduct, SubscriptionProductSummary } from '@/types'

export async function getActive(): Promise<ActiveSubscriptionProduct[]> {
  const { data } = await xlabClient.get<ActiveSubscriptionProduct[]>('/subscription-products/active')
  return data
}

export async function getSummary(): Promise<SubscriptionProductSummary> {
  const { data } = await xlabClient.get<SubscriptionProductSummary>('/subscription-products/summary')
  return data
}

export async function getProgress(): Promise<SubscriptionProductSummary> {
  const { data } = await xlabClient.get<SubscriptionProductSummary>('/subscription-products/progress')
  return data
}

export const xlabSubscriptionProductsAPI = { getActive, getSummary, getProgress }
```

Modify `frontend-v2/src/api/subscriptionProducts.ts` to delegate to xlab:

```ts
import { xlabSubscriptionProductsAPI } from './xlab/subscriptionProducts'

export const getActive = xlabSubscriptionProductsAPI.getActive
export const getSummary = xlabSubscriptionProductsAPI.getSummary
export const getProgress = xlabSubscriptionProductsAPI.getProgress

export const subscriptionProductsAPI = { getActive, getSummary, getProgress }

export default subscriptionProductsAPI
```

- [ ] **Step 4: Run frontend adapter tests to verify GREEN**

Run:

```bash
npm --prefix frontend-v2 exec -- vitest --root frontend-v2 run src/api/xlab/__tests__/subscriptionProducts.spec.ts
```

Expected: PASS.

- [ ] **Step 5: Commit frontend adapter**

Run:

```bash
git add frontend-v2/src/api/xlabClient.ts frontend-v2/src/api/xlab frontend-v2/src/api/subscriptionProducts.ts
git commit -m "$(cat <<'EOF'
feat(frontend-v2): route product subscriptions through xlab API

Add an xlab API client and route product subscription reads through the new xlab boundary while keeping page imports stable.
EOF
)"
```

Expected: commit succeeds.

---

### Task 5: Add runtime wiring documentation

**Files:**
- Create: `docs/superpowers/specs/2026-06-07-xlab-shell-phase1-runtime.md`

- [ ] **Step 1: Write runtime doc**

Create `docs/superpowers/specs/2026-06-07-xlab-shell-phase1-runtime.md`:

```md
# Xlab Shell Phase 1 Runtime Wiring

Phase 1 introduces a standalone xlab backend with read-only product subscription endpoints.

## Services

- sub2api core: existing upstream-compatible core service.
- xlab backend: new service listening on `XLAB_SERVER_ADDR`, default `:8090`.
- frontend-v2: calls `/api/v1` for core and `/xapi/v1` for xlab.

## Required environment

For xlab backend:

```text
XLAB_SERVER_ADDR=:8090
CORE_API_BASE_URL=http://sub2api:8080/api/v1
XLAB_CORE_TIMEOUT_SECONDS=10
```

For frontend-v2 build:

```text
VITE_XLAB_API_BASE_URL=/xapi/v1
```

## Reverse proxy rules

```text
/api/v1  -> sub2api core
/v1      -> sub2api core gateway
/xapi/v1 -> xlab backend
/*       -> frontend-v2 shell
```

## Rollout note

The xlab backend initially proxies current core product subscription reads. It does not migrate product subscription data and does not change payment fulfillment.
```

- [ ] **Step 2: Commit runtime doc**

Run:

```bash
git add docs/superpowers/specs/2026-06-07-xlab-shell-phase1-runtime.md
git commit -m "$(cat <<'EOF'
docs(xlab): document phase one runtime wiring

Record the proxy and environment requirements for running frontend-v2 with the new xlab backend boundary.
EOF
)"
```

Expected: commit succeeds.

---

### Task 6: Verification

**Files:**
- No source changes unless verification fails.

- [ ] **Step 1: Run xlab backend tests**

Run:

```bash
cd xlab-backend
```

Expected: PASS.

- [ ] **Step 2: Run frontend adapter tests**

Run:

```bash
npm --prefix frontend-v2 exec -- vitest --root frontend-v2 run src/api/xlab/__tests__/subscriptionProducts.spec.ts
```

Expected: PASS.

- [ ] **Step 3: Run frontend-v2 typecheck and build**

Run:

```bash
npm --prefix frontend-v2 run typecheck
npm --prefix frontend-v2 run build
```

Expected: PASS; existing Vite chunk-size warnings are acceptable.

- [ ] **Step 4: Check no core subscription/payment code changed**

Run:

```bash
git diff --name-only xlabapi...HEAD | rg 'backend/internal/(service|handler|repository).*subscription|backend/migrations|backend/ent/schema|payment_fulfillment' || true
```

Expected: no output except docs if the grep matches documentation paths. This confirms Phase 1 only introduced a boundary, not business migration.

- [ ] **Step 5: Commit any verification fixes**

If verification required code changes, commit them with:

```bash
git add <changed-files>
git commit -m "fix(xlab): stabilize phase one shell boundary"
```

Expected: no commit is needed if all tests pass.

---

## Self-Review

- Spec coverage: The plan implements Phase 1 only: identity validation, read-only product subscription API boundary, frontend-v2 adapter, and runtime docs.
- Placeholder scan: No TBD/TODO placeholders remain; every file path and command is explicit.
- Type consistency: `xlab-backend`, `/xapi/v1`, `CORE_API_BASE_URL`, `VITE_XLAB_API_BASE_URL`, and product subscription endpoint paths are consistent across backend, frontend, and runtime docs.
