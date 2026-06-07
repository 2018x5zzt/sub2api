# Xlab Product Subscription Read Mirror Phase 3A Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move product-subscription read traffic behind `/xapi/v1` from a pure core proxy to an xlab DB read mirror with safe core fallback.

**Architecture:** `xlab-backend` adds Postgres configuration, embedded xlab mirror migrations, a repository for mirrored subscription reads, a full-snapshot syncer from core DB, and a read service with `core`, `hybrid`, and `xlab` modes. Core remains the write and entitlement source of truth; payment, redeem, admin writes, gateway billing, and API key authorization stay in core.

**Tech Stack:** Go 1.26.2, `database/sql`, Postgres via `github.com/lib/pq`, SQL-focused tests with `github.com/DATA-DOG/go-sqlmock`, `net/http`, Docker/Bash deployment scripts.

---

## File Structure

- Modify `xlab-backend/go.mod` and create `xlab-backend/go.sum`
  - Add Postgres driver and SQL mock test dependency.
- Modify `xlab-backend/internal/config/config.go`
  - Add xlab/core DB URLs, subscription read-source mode, sync interval, stale threshold, and sync enablement.
- Modify `xlab-backend/internal/config/config_test.go`
  - Cover defaults, env overrides, and invalid read-source fallback.
- Create `xlab-backend/internal/storage/db.go`
  - Open and health-check Postgres connections.
- Create `xlab-backend/internal/storage/migrations.go`
  - Run embedded xlab DB migrations.
- Create `xlab-backend/internal/storage/migrations/001_product_subscription_read_mirror.sql`
  - Create mirror tables and indexes.
- Modify `xlab-backend/internal/httpapi/auth.go`
  - Store the validated core user in request context.
- Modify `xlab-backend/internal/httpapi/router.go`
  - Accept an optional subscription read service argument explicitly.
- Modify `xlab-backend/internal/httpapi/subscription_products.go`
  - Delegate reads to the read service when configured.
- Modify `xlab-backend/internal/httpapi/router_test.go`
  - Cover user context and read-service routing.
- Create `xlab-backend/internal/subscriptions/types.go`
  - Define response contract, repository interfaces, and read-source constants.
- Create `xlab-backend/internal/subscriptions/repository.go`
  - Query xlab mirror tables and return the frontend-compatible response shape.
- Create `xlab-backend/internal/subscriptions/repository_test.go`
  - Verify active-product mapping, group mapping, summary aggregation, and sync-state freshness data.
- Create `xlab-backend/internal/subscriptions/service.go`
  - Implement `core`, `hybrid`, and `xlab` read selection with fallback.
- Create `xlab-backend/internal/subscriptions/service_test.go`
  - Verify read-source and fallback behavior.
- Create `xlab-backend/internal/subscriptions/syncer.go`
  - Read core product-subscription tables and mirror them to xlab DB.
- Create `xlab-backend/internal/subscriptions/syncer_test.go`
  - Verify full-snapshot transaction behavior and sync-state update.
- Modify `xlab-backend/cmd/server/main.go`
  - Open DBs, run migrations, start optional sync loop, and wire the read service.
- Create `xlab-backend/cmd/server/main_test.go`
  - Verify read-source mapping.
- Modify `deploy-xlab-backend.sh`
  - Pass Phase 3A runtime env vars into the container.
- Modify `docs/superpowers/specs/2026-06-07-xlab-backend-docker-phase2-runbook.md`
  - Document Phase 3A rollout and rollback modes.

This plan does not modify core product-subscription writers, core migrations, gateway billing, payment fulfillment, redeem behavior, or frontend-v2 request paths.

---

### Task 1: Add Phase 3A configuration

**Files:**
- Modify: `xlab-backend/go.mod`
- Create: `xlab-backend/go.sum`
- Modify: `xlab-backend/internal/config/config.go`
- Modify: `xlab-backend/internal/config/config_test.go`

- [ ] **Step 1: Add dependencies**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: `go.mod` contains `github.com/lib/pq` and `github.com/DATA-DOG/go-sqlmock`; `go.sum` exists.

- [ ] **Step 2: Write failing config tests**

Append to `xlab-backend/internal/config/config_test.go`:

```go
func TestLoadPhase3Defaults(t *testing.T) {
	t.Setenv("XLAB_DATABASE_URL", "")
	t.Setenv("CORE_DATABASE_URL", "")
	t.Setenv("XLAB_SUBSCRIPTION_READ_SOURCE", "")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS", "")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS", "")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_ENABLED", "")

	cfg := Load()
	if cfg.XlabDatabaseURL != "" {
		t.Fatalf("XlabDatabaseURL = %q", cfg.XlabDatabaseURL)
	}
	if cfg.CoreDatabaseURL != "" {
		t.Fatalf("CoreDatabaseURL = %q", cfg.CoreDatabaseURL)
	}
	if cfg.SubscriptionReadSource != SubscriptionReadSourceCore {
		t.Fatalf("SubscriptionReadSource = %q", cfg.SubscriptionReadSource)
	}
	if cfg.SubscriptionSyncInterval != 5*time.Minute {
		t.Fatalf("SubscriptionSyncInterval = %s", cfg.SubscriptionSyncInterval)
	}
	if cfg.SubscriptionSyncStaleAfter != 10*time.Minute {
		t.Fatalf("SubscriptionSyncStaleAfter = %s", cfg.SubscriptionSyncStaleAfter)
	}
	if cfg.SubscriptionSyncEnabled {
		t.Fatal("SubscriptionSyncEnabled should default false")
	}
}

func TestLoadPhase3FromEnv(t *testing.T) {
	t.Setenv("XLAB_DATABASE_URL", "postgres://xlab-db")
	t.Setenv("CORE_DATABASE_URL", "postgres://core-db")
	t.Setenv("XLAB_SUBSCRIPTION_READ_SOURCE", "hybrid")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS", "17")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS", "31")
	t.Setenv("XLAB_SUBSCRIPTION_SYNC_ENABLED", "true")

	cfg := Load()
	if cfg.XlabDatabaseURL != "postgres://xlab-db" {
		t.Fatalf("XlabDatabaseURL = %q", cfg.XlabDatabaseURL)
	}
	if cfg.CoreDatabaseURL != "postgres://core-db" {
		t.Fatalf("CoreDatabaseURL = %q", cfg.CoreDatabaseURL)
	}
	if cfg.SubscriptionReadSource != SubscriptionReadSourceHybrid {
		t.Fatalf("SubscriptionReadSource = %q", cfg.SubscriptionReadSource)
	}
	if cfg.SubscriptionSyncInterval != 17*time.Second {
		t.Fatalf("SubscriptionSyncInterval = %s", cfg.SubscriptionSyncInterval)
	}
	if cfg.SubscriptionSyncStaleAfter != 31*time.Second {
		t.Fatalf("SubscriptionSyncStaleAfter = %s", cfg.SubscriptionSyncStaleAfter)
	}
	if !cfg.SubscriptionSyncEnabled {
		t.Fatal("SubscriptionSyncEnabled should be true")
	}
}

func TestLoadInvalidReadSourceFallsBackToCore(t *testing.T) {
	t.Setenv("XLAB_SUBSCRIPTION_READ_SOURCE", "banana")
	cfg := Load()
	if cfg.SubscriptionReadSource != SubscriptionReadSourceCore {
		t.Fatalf("SubscriptionReadSource = %q", cfg.SubscriptionReadSource)
	}
}
```

- [ ] **Step 3: Run config tests and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because Phase 3A config fields and constants are not defined.

- [ ] **Step 4: Replace config implementation**

Replace `xlab-backend/internal/config/config.go` with:

```go
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type SubscriptionReadSource string

const (
	SubscriptionReadSourceCore   SubscriptionReadSource = "core"
	SubscriptionReadSourceHybrid SubscriptionReadSource = "hybrid"
	SubscriptionReadSourceXlab   SubscriptionReadSource = "xlab"
)

type Config struct {
	ServerAddr                 string
	CoreAPIBaseURL             string
	CoreTimeout                time.Duration
	XlabDatabaseURL            string
	CoreDatabaseURL            string
	SubscriptionReadSource     SubscriptionReadSource
	SubscriptionSyncEnabled    bool
	SubscriptionSyncInterval   time.Duration
	SubscriptionSyncStaleAfter time.Duration
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

	timeoutSeconds := positiveIntEnv("XLAB_CORE_TIMEOUT_SECONDS", 10)

	return Config{
		ServerAddr:                 addr,
		CoreAPIBaseURL:             baseURL,
		CoreTimeout:                time.Duration(timeoutSeconds) * time.Second,
		XlabDatabaseURL:            strings.TrimSpace(os.Getenv("XLAB_DATABASE_URL")),
		CoreDatabaseURL:            strings.TrimSpace(os.Getenv("CORE_DATABASE_URL")),
		SubscriptionReadSource:     parseReadSource(os.Getenv("XLAB_SUBSCRIPTION_READ_SOURCE")),
		SubscriptionSyncEnabled:    boolEnv("XLAB_SUBSCRIPTION_SYNC_ENABLED", false),
		SubscriptionSyncInterval:   time.Duration(positiveIntEnv("XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS", 300)) * time.Second,
		SubscriptionSyncStaleAfter: time.Duration(positiveIntEnv("XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS", 600)) * time.Second,
	}
}

func parseReadSource(raw string) SubscriptionReadSource {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(SubscriptionReadSourceHybrid):
		return SubscriptionReadSourceHybrid
	case string(SubscriptionReadSourceXlab):
		return SubscriptionReadSourceXlab
	default:
		return SubscriptionReadSourceCore
	}
}

func positiveIntEnv(name string, fallback int) int {
	if raw := strings.TrimSpace(os.Getenv(name)); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func boolEnv(name string, fallback bool) bool {
	if raw := strings.TrimSpace(os.Getenv(name)); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			return parsed
		}
	}
	return fallback
}
```

- [ ] **Step 5: Verify config tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 6: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): add subscription mirror runtime config

Introduce database and read-source settings needed to run product subscription reads from an xlab mirror with core fallback.
EOF
)"
```

---

### Task 2: Add xlab DB migrations and storage helpers

**Files:**
- Create: `xlab-backend/internal/storage/db.go`
- Create: `xlab-backend/internal/storage/migrations.go`
- Create: `xlab-backend/internal/storage/migrations/001_product_subscription_read_mirror.sql`
- Create: `xlab-backend/internal/storage/migrations_test.go`

- [ ] **Step 1: Write failing migration test**

Create `xlab-backend/internal/storage/migrations_test.go`:

```go
package storage

import (
	"strings"
	"testing"
)

func TestMigrationSQLContainsPhase3MirrorTables(t *testing.T) {
	sql, err := migrationSQL()
	if err != nil {
		t.Fatalf("migrationSQL error: %v", err)
	}
	for _, table := range []string{
		"xlab_subscription_products",
		"xlab_subscription_product_groups",
		"xlab_user_product_subscriptions",
		"xlab_sync_state",
	} {
		if !strings.Contains(sql, table) {
			t.Fatalf("migration SQL missing %s", table)
		}
	}
}
```

- [ ] **Step 2: Run migration test and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because package `internal/storage` is absent.

- [ ] **Step 3: Add migration SQL**

Create `xlab-backend/internal/storage/migrations/001_product_subscription_read_mirror.sql`:

```sql
CREATE TABLE IF NOT EXISTS xlab_subscription_products (
    core_product_id BIGINT PRIMARY KEY,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    product_family TEXT NOT NULL DEFAULT 'gpt',
    daily_limit_usd NUMERIC(18,6),
    weekly_limit_usd NUMERIC(18,6),
    monthly_limit_usd NUMERIC(18,6),
    daily_carryover_enabled BOOLEAN NOT NULL DEFAULT false,
    daily_carryover_limit_usd NUMERIC(18,6),
    source_created_at TIMESTAMPTZ,
    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_products_status
    ON xlab_subscription_products(status);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_products_synced_at
    ON xlab_subscription_products(synced_at);

CREATE TABLE IF NOT EXISTS xlab_subscription_product_groups (
    core_binding_id BIGINT PRIMARY KEY,
    core_product_id BIGINT NOT NULL REFERENCES xlab_subscription_products(core_product_id) ON DELETE CASCADE,
    core_group_id BIGINT NOT NULL,
    group_name TEXT NOT NULL,
    group_platform TEXT,
    balance_fallback_group_id BIGINT,
    balance_fallback_group_name TEXT,
    debit_multiplier NUMERIC(18,6) NOT NULL DEFAULT 1,
    status TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    source_created_at TIMESTAMPTZ,
    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_product_groups_product
    ON xlab_subscription_product_groups(core_product_id, sort_order, core_group_id);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_product_groups_group
    ON xlab_subscription_product_groups(core_group_id);

CREATE TABLE IF NOT EXISTS xlab_user_product_subscriptions (
    core_subscription_id BIGINT PRIMARY KEY,
    core_user_id BIGINT NOT NULL,
    core_product_id BIGINT NOT NULL REFERENCES xlab_subscription_products(core_product_id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    daily_usage_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    weekly_usage_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    monthly_usage_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    daily_limit_usd NUMERIC(18,6),
    weekly_limit_usd NUMERIC(18,6),
    monthly_limit_usd NUMERIC(18,6),
    daily_carryover_in_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    daily_carryover_remaining_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    source_created_at TIMESTAMPTZ,
    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_xlab_user_product_subscriptions_user_active
    ON xlab_user_product_subscriptions(core_user_id, status, expires_at);

CREATE INDEX IF NOT EXISTS idx_xlab_user_product_subscriptions_product
    ON xlab_user_product_subscriptions(core_product_id);

CREATE INDEX IF NOT EXISTS idx_xlab_user_product_subscriptions_synced_at
    ON xlab_user_product_subscriptions(synced_at);

CREATE TABLE IF NOT EXISTS xlab_sync_state (
    source_name TEXT PRIMARY KEY,
    last_success_at TIMESTAMPTZ,
    last_watermark TEXT,
    last_error TEXT,
    last_error_at TIMESTAMPTZ,
    row_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- [ ] **Step 4: Add storage helpers**

Create `xlab-backend/internal/storage/db.go`:

```go
package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func OpenPostgres(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
```

Create `xlab-backend/internal/storage/migrations.go`:

```go
package storage

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(ctx context.Context, db *sql.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return err
		}
	}
	return nil
}

func migrationSQL() (string, error) {
	content, err := migrationsFS.ReadFile("migrations/001_product_subscription_read_mirror.sql")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
```

- [ ] **Step 5: Verify storage tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 6: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): add product subscription mirror migrations

Create xlab mirror tables and storage helpers for Phase 3A product subscription read migration.
EOF
)"
```

---

### Task 3: Preserve authenticated core user in request context

**Files:**
- Modify: `xlab-backend/internal/httpapi/auth.go`
- Modify: `xlab-backend/internal/httpapi/router_test.go`

- [ ] **Step 1: Write failing auth context test**

Append to `xlab-backend/internal/httpapi/router_test.go`:

```go
func TestAuthStoresCurrentUserInContext(t *testing.T) {
	fake := &fakeCoreClient{}
	api := &API{core: fake}
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()

	api.auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userFromContext(r.Context())
		if user == nil || user.ID != 7 || user.Email != "u@example.com" {
			t.Fatalf("userFromContext = %+v", user)
		}
		if tokenFromContext(r.Context()) != "abc" {
			t.Fatalf("tokenFromContext = %q", tokenFromContext(r.Context()))
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
```

- [ ] **Step 2: Run auth context test and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because `userFromContext` is undefined.

- [ ] **Step 3: Store core user in context**

Modify `xlab-backend/internal/httpapi/auth.go`.

Add import:

```go
"github.com/2018x5zzt/xlab-backend/internal/core"
```

Replace the context-key constants with:

```go
const tokenContextKey contextKey = "token"
const userContextKey contextKey = "user"
```

Replace the token validation block inside `auth` with:

```go
		user, err := a.core.CurrentUser(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), tokenContextKey, token)
		ctx = context.WithValue(ctx, userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
```

Add:

```go
func userFromContext(ctx context.Context) *core.User {
	value, _ := ctx.Value(userContextKey).(*core.User)
	return value
}
```

- [ ] **Step 4: Verify HTTP tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 5: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): retain authenticated core user context

Store the validated core user alongside the bearer token so xlab subscription reads can query by core user id.
EOF
)"
```

---

### Task 4: Add subscription response types and mirror repository

**Files:**
- Create: `xlab-backend/internal/subscriptions/types.go`
- Create: `xlab-backend/internal/subscriptions/repository.go`
- Create: `xlab-backend/internal/subscriptions/repository_test.go`

- [ ] **Step 1: Write failing repository tests**

Create `xlab-backend/internal/subscriptions/repository_test.go`:

```go
package subscriptions

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryListActiveProductsByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	productRows := sqlmock.NewRows([]string{
		"core_subscription_id", "core_product_id", "code", "name", "description", "status", "expires_at",
		"daily_usage_usd", "weekly_usage_usd", "monthly_usage_usd",
		"daily_limit_usd", "weekly_limit_usd", "monthly_limit_usd",
		"daily_carryover_in_usd", "daily_carryover_remaining_usd",
	}).AddRow(int64(99), int64(10), "gpt-pro", "GPT Pro", "desc", "active", expires, 1.25, 2.5, 3.75, 10.0, 20.0, 30.0, 0.5, 0.25)
	groupRows := sqlmock.NewRows([]string{"core_product_id", "core_group_id", "group_name", "group_platform", "debit_multiplier", "status", "sort_order"}).
		AddRow(int64(10), int64(20), "GPT-4", "openai", 1.2, "active", 1)

	mock.ExpectQuery(regexp.QuoteMeta(activeProductsByUserSQL)).WithArgs(int64(7)).WillReturnRows(productRows)
	mock.ExpectQuery(regexp.QuoteMeta(groupsByProductSQL)).WithArgs(sqlmock.AnyArg()).WillReturnRows(groupRows)

	repo := NewRepository(db)
	items, err := repo.ListActiveProductsByUser(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListActiveProductsByUser error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d", len(items))
	}
	item := items[0]
	if item.SubscriptionID != 99 || item.ProductID != 10 || item.Name != "GPT Pro" || item.Description != "desc" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if len(item.Groups) != 1 || item.Groups[0].GroupName != "GPT-4" || item.Groups[0].DebitMultiplier != 1.2 {
		t.Fatalf("unexpected groups: %+v", item.Groups)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositorySummaryAggregatesMonthlyUsageAndLimit(t *testing.T) {
	items := []ActiveProduct{
		{MonthlyUsageUSD: 1.5, MonthlyLimitUSD: 10},
		{MonthlyUsageUSD: 2.5, MonthlyLimitUSD: 20},
	}
	summary := SummaryFromActiveProducts(items)
	if summary.ActiveCount != 2 || summary.TotalMonthlyUsageUSD != 4.0 || summary.TotalMonthlyLimitUSD != 30.0 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestRepositorySyncState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(syncStateSQL)).WithArgs("product_subscriptions").WillReturnRows(
		sqlmock.NewRows([]string{"source_name", "last_success_at", "row_count"}).AddRow("product_subscriptions", now, 3),
	)
	repo := NewRepository(db)
	state, err := repo.SyncState(context.Background(), "product_subscriptions")
	if err != nil {
		t.Fatalf("SyncState error: %v", err)
	}
	if state.LastSuccessAt == nil || state.RowCount != 3 {
		t.Fatalf("state = %+v", state)
	}
}
```

- [ ] **Step 2: Run repository tests and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because package `internal/subscriptions` is absent.

- [ ] **Step 3: Add response types**

Create `xlab-backend/internal/subscriptions/types.go`:

```go
package subscriptions

import (
	"context"
	"encoding/json"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type ReadSource string

const (
	ReadSourceCore   ReadSource = "core"
	ReadSourceHybrid ReadSource = "hybrid"
	ReadSourceXlab   ReadSource = "xlab"
)

type Group struct {
	GroupID         int64   `json:"group_id"`
	GroupName       string  `json:"group_name"`
	GroupPlatform   string  `json:"group_platform"`
	DebitMultiplier float64 `json:"debit_multiplier"`
	Status          string  `json:"status"`
	SortOrder       int     `json:"sort_order"`
}

type ActiveProduct struct {
	ProductID                  int64      `json:"product_id"`
	SubscriptionID             int64      `json:"subscription_id"`
	Code                       string     `json:"code"`
	Name                       string     `json:"name"`
	Description                string     `json:"description"`
	Status                     string     `json:"status"`
	ExpiresAt                  *time.Time `json:"expires_at,omitempty"`
	DailyUsageUSD              float64    `json:"daily_usage_usd"`
	WeeklyUsageUSD             float64    `json:"weekly_usage_usd"`
	MonthlyUsageUSD            float64    `json:"monthly_usage_usd"`
	DailyLimitUSD              float64    `json:"daily_limit_usd"`
	WeeklyLimitUSD             float64    `json:"weekly_limit_usd"`
	MonthlyLimitUSD            float64    `json:"monthly_limit_usd"`
	DailyCarryoverInUSD        float64    `json:"daily_carryover_in_usd"`
	DailyCarryoverRemainingUSD float64    `json:"daily_carryover_remaining_usd"`
	Groups                     []Group    `json:"groups"`
}

type Summary struct {
	ActiveCount          int             `json:"active_count"`
	TotalMonthlyUsageUSD float64         `json:"total_monthly_usage_usd"`
	TotalMonthlyLimitUSD float64         `json:"total_monthly_limit_usd"`
	Products             []ActiveProduct `json:"products"`
}

func SummaryFromActiveProducts(items []ActiveProduct) Summary {
	summary := Summary{ActiveCount: len(items), Products: items}
	for _, item := range items {
		summary.TotalMonthlyUsageUSD += item.MonthlyUsageUSD
		summary.TotalMonthlyLimitUSD += item.MonthlyLimitUSD
	}
	return summary
}

type SyncState struct {
	SourceName    string
	LastSuccessAt *time.Time
	RowCount      int
}

type CoreProxy interface {
	ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error)
}

type MirrorRepository interface {
	ListActiveProductsByUser(ctx context.Context, userID int64) ([]ActiveProduct, error)
	SyncState(ctx context.Context, sourceName string) (*SyncState, error)
}

type ReadService interface {
	Active(ctx context.Context, user *core.User, token string) (any, error)
	Summary(ctx context.Context, user *core.User, token string) (any, error)
	Progress(ctx context.Context, user *core.User, token string) (any, error)
}
```

- [ ] **Step 4: Add repository implementation**

Create `xlab-backend/internal/subscriptions/repository.go`:

```go
package subscriptions

import (
	"context"
	"database/sql"
)

const activeProductsByUserSQL = `
SELECT
    ups.core_subscription_id,
    ups.core_product_id,
    sp.code,
    sp.name,
    sp.description,
    ups.status,
    ups.expires_at,
    ups.daily_usage_usd,
    ups.weekly_usage_usd,
    ups.monthly_usage_usd,
    COALESCE(ups.daily_limit_usd, sp.daily_limit_usd, 0),
    COALESCE(ups.weekly_limit_usd, sp.weekly_limit_usd, 0),
    COALESCE(ups.monthly_limit_usd, sp.monthly_limit_usd, 0),
    ups.daily_carryover_in_usd,
    ups.daily_carryover_remaining_usd
FROM xlab_user_product_subscriptions ups
JOIN xlab_subscription_products sp
  ON sp.core_product_id = ups.core_product_id
WHERE ups.core_user_id = $1
  AND ups.status = 'active'
  AND ups.expires_at > NOW()
  AND sp.status = 'active'
ORDER BY sp.name ASC, ups.expires_at DESC, ups.core_subscription_id DESC`

const groupsByProductSQL = `
SELECT
    core_product_id,
    core_group_id,
    group_name,
    COALESCE(group_platform, ''),
    debit_multiplier,
    status,
    sort_order
FROM xlab_subscription_product_groups
WHERE core_product_id = ANY($1)
  AND status = 'active'
ORDER BY core_product_id ASC, sort_order ASC, core_group_id ASC`

const syncStateSQL = `
SELECT source_name, last_success_at, row_count
FROM xlab_sync_state
WHERE source_name = $1`

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListActiveProductsByUser(ctx context.Context, userID int64) ([]ActiveProduct, error) {
	rows, err := r.db.QueryContext(ctx, activeProductsByUserSQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ActiveProduct, 0)
	productIDs := make([]int64, 0)
	for rows.Next() {
		var item ActiveProduct
		var expiresAt sql.NullTime
		if err := rows.Scan(
			&item.SubscriptionID,
			&item.ProductID,
			&item.Code,
			&item.Name,
			&item.Description,
			&item.Status,
			&expiresAt,
			&item.DailyUsageUSD,
			&item.WeeklyUsageUSD,
			&item.MonthlyUsageUSD,
			&item.DailyLimitUSD,
			&item.WeeklyLimitUSD,
			&item.MonthlyLimitUSD,
			&item.DailyCarryoverInUSD,
			&item.DailyCarryoverRemainingUSD,
		); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}
		item.Groups = []Group{}
		items = append(items, item)
		productIDs = append(productIDs, item.ProductID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []ActiveProduct{}, nil
	}

	groupsByProduct, err := r.groupsByProductID(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Groups = groupsByProduct[items[i].ProductID]
		if items[i].Groups == nil {
			items[i].Groups = []Group{}
		}
	}
	return items, nil
}

func (r *Repository) groupsByProductID(ctx context.Context, productIDs []int64) (map[int64][]Group, error) {
	rows, err := r.db.QueryContext(ctx, groupsByProductSQL, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64][]Group)
	for rows.Next() {
		var productID int64
		var group Group
		if err := rows.Scan(&productID, &group.GroupID, &group.GroupName, &group.GroupPlatform, &group.DebitMultiplier, &group.Status, &group.SortOrder); err != nil {
			return nil, err
		}
		out[productID] = append(out[productID], group)
	}
	return out, rows.Err()
}

func (r *Repository) SyncState(ctx context.Context, sourceName string) (*SyncState, error) {
	row := r.db.QueryRowContext(ctx, syncStateSQL, sourceName)
	var state SyncState
	var lastSuccess sql.NullTime
	if err := row.Scan(&state.SourceName, &lastSuccess, &state.RowCount); err != nil {
		return nil, err
	}
	if lastSuccess.Valid {
		state.LastSuccessAt = &lastSuccess.Time
	}
	return &state, nil
}
```

- [ ] **Step 5: Verify repository tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 6: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): add product subscription mirror repository

Read mirrored product subscriptions from xlab DB using the existing frontend response contract.
EOF
)"
```

---

### Task 5: Add read service with source modes and fallback

**Files:**
- Create: `xlab-backend/internal/subscriptions/service.go`
- Create: `xlab-backend/internal/subscriptions/service_test.go`

- [ ] **Step 1: Write failing service tests**

Create `xlab-backend/internal/subscriptions/service_test.go`:

```go
package subscriptions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type fakeMirrorRepo struct {
	products []ActiveProduct
	state    *SyncState
	err      error
}

func (f *fakeMirrorRepo) ListActiveProductsByUser(ctx context.Context, userID int64) ([]ActiveProduct, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.products, nil
}

func (f *fakeMirrorRepo) SyncState(ctx context.Context, sourceName string) (*SyncState, error) {
	if f.state == nil {
		return nil, sql.ErrNoRows
	}
	return f.state, nil
}

type fakeCoreProxy struct {
	path string
}

func (f *fakeCoreProxy) ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error) {
	f.path = path
	return json.RawMessage(`[{"subscription_id":44,"name":"Core"}]`), nil
}

func TestServiceCoreModeUsesCoreProxy(t *testing.T) {
	coreProxy := &fakeCoreProxy{}
	svc := NewService(ReadSourceCore, &fakeMirrorRepo{}, coreProxy, 10*time.Minute, time.Now)
	out, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active error: %v", err)
	}
	if coreProxy.path != "/subscription-products/active" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
	if _, ok := out.(json.RawMessage); !ok {
		t.Fatalf("output type = %T", out)
	}
}

func TestServiceHybridFreshMirrorUsesRepository(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakeMirrorRepo{
		state:    &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &now, RowCount: 1},
		products: []ActiveProduct{{ProductID: 10, SubscriptionID: 99, Name: "Mirror", Groups: []Group{}}},
	}
	svc := NewService(ReadSourceHybrid, repo, &fakeCoreProxy{}, 10*time.Minute, func() time.Time { return now })
	out, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active error: %v", err)
	}
	items, ok := out.([]ActiveProduct)
	if !ok || len(items) != 1 || items[0].Name != "Mirror" {
		t.Fatalf("unexpected output: %#v", out)
	}
}

func TestServiceHybridStaleMirrorFallsBackToCore(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	old := now.Add(-time.Hour)
	coreProxy := &fakeCoreProxy{}
	repo := &fakeMirrorRepo{state: &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &old, RowCount: 1}}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })
	_, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active error: %v", err)
	}
	if coreProxy.path != "/subscription-products/active" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}

func TestServiceXlabFallsBackWhenRepositoryFails(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	coreProxy := &fakeCoreProxy{}
	repo := &fakeMirrorRepo{state: &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &now, RowCount: 1}, err: errors.New("db down")}
	svc := NewService(ReadSourceXlab, repo, coreProxy, 10*time.Minute, func() time.Time { return now })
	_, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active should fallback instead of failing: %v", err)
	}
	if coreProxy.path != "/subscription-products/active" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}
```

- [ ] **Step 2: Run service tests and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because `NewService` is undefined.

- [ ] **Step 3: Add service implementation**

Create `xlab-backend/internal/subscriptions/service.go`:

```go
package subscriptions

import (
	"context"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

const syncSourceProductSubscriptions = "product_subscriptions"

type Service struct {
	source     ReadSource
	repo       MirrorRepository
	core       CoreProxy
	staleAfter time.Duration
	now        func() time.Time
}

func NewService(source ReadSource, repo MirrorRepository, core CoreProxy, staleAfter time.Duration, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{source: source, repo: repo, core: core, staleAfter: staleAfter, now: now}
}

func (s *Service) Active(ctx context.Context, user *core.User, token string) (any, error) {
	if s.source == ReadSourceCore || user == nil || s.repo == nil {
		return s.core.ProxyGET(ctx, token, "/subscription-products/active")
	}
	if !s.mirrorFresh(ctx) {
		return s.core.ProxyGET(ctx, token, "/subscription-products/active")
	}
	items, err := s.repo.ListActiveProductsByUser(ctx, user.ID)
	if err != nil {
		return s.core.ProxyGET(ctx, token, "/subscription-products/active")
	}
	if s.source == ReadSourceHybrid && len(items) == 0 {
		return s.core.ProxyGET(ctx, token, "/subscription-products/active")
	}
	return items, nil
}

func (s *Service) Summary(ctx context.Context, user *core.User, token string) (any, error) {
	active, err := s.Active(ctx, user, token)
	if err != nil {
		return nil, err
	}
	items, ok := active.([]ActiveProduct)
	if !ok {
		return active, nil
	}
	summary := SummaryFromActiveProducts(items)
	return summary, nil
}

func (s *Service) Progress(ctx context.Context, user *core.User, token string) (any, error) {
	return s.Summary(ctx, user, token)
}

func (s *Service) mirrorFresh(ctx context.Context) bool {
	state, err := s.repo.SyncState(ctx, syncSourceProductSubscriptions)
	if err != nil || state == nil || state.LastSuccessAt == nil {
		return false
	}
	return s.now().Sub(*state.LastSuccessAt) <= s.staleAfter
}
```

- [ ] **Step 4: Verify service tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 5: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): add subscription read source fallback service

Route product subscription reads through core, hybrid mirror, or xlab mirror modes with stale-data fallback.
EOF
)"
```

---

### Task 6: Wire HTTP routes to the read service

**Files:**
- Modify: `xlab-backend/internal/httpapi/auth.go`
- Modify: `xlab-backend/internal/httpapi/router.go`
- Modify: `xlab-backend/internal/httpapi/subscription_products.go`
- Modify: `xlab-backend/internal/httpapi/router_test.go`

- [ ] **Step 1: Write failing HTTP read-service test**

Add to `xlab-backend/internal/httpapi/router_test.go`:

```go
type fakeReadService struct {
	activeCalled bool
}

func (f *fakeReadService) Active(ctx context.Context, user *core.User, token string) (any, error) {
	f.activeCalled = true
	return []map[string]any{{"subscription_id": 123, "name": "Mirror"}}, nil
}

func (f *fakeReadService) Summary(ctx context.Context, user *core.User, token string) (any, error) {
	return map[string]any{"active_count": 1, "total_monthly_usage_usd": 0, "total_monthly_limit_usd": 0, "products": []any{}}, nil
}

func (f *fakeReadService) Progress(ctx context.Context, user *core.User, token string) (any, error) {
	return map[string]any{"active_count": 1, "total_monthly_usage_usd": 0, "total_monthly_limit_usd": 0, "products": []any{}}, nil
}

func TestSubscriptionProductsUseReadServiceWhenProvided(t *testing.T) {
	fakeCore := &fakeCoreClient{}
	fakeReads := &fakeReadService{}
	r := NewRouter(fakeCore, fakeReads)
	req := httptest.NewRequest(http.MethodGet, "/xapi/v1/subscription-products/active", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !fakeReads.activeCalled {
		t.Fatal("read service was not called")
	}
	if !strings.Contains(rec.Body.String(), `"subscription_id":123`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run HTTP tests and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because `NewRouter` accepts only one argument.

- [ ] **Step 3: Update HTTP API wiring**

Modify `xlab-backend/internal/httpapi/router.go` to add the interface and explicit second argument:

```go
type SubscriptionReadService interface {
	Active(ctx context.Context, user *core.User, token string) (any, error)
	Summary(ctx context.Context, user *core.User, token string) (any, error)
	Progress(ctx context.Context, user *core.User, token string) (any, error)
}

func NewRouter(coreClient CoreClient, readService SubscriptionReadService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	api := &API{core: coreClient, subscriptionReads: readService}
	mux.Handle("/xapi/v1/subscription-products/active", api.auth(http.HandlerFunc(api.getActiveProducts)))
	mux.Handle("/xapi/v1/subscription-products/summary", api.auth(http.HandlerFunc(api.getProductSummary)))
	mux.Handle("/xapi/v1/subscription-products/progress", api.auth(http.HandlerFunc(api.getProductProgress)))
	return mux
}
```

Update all existing tests that call `NewRouter(fake)` to call `NewRouter(fake, nil)`.

Modify `xlab-backend/internal/httpapi/auth.go` so `API` has both fields:

```go
type API struct {
	core              CoreClient
	subscriptionReads SubscriptionReadService
}
```

Replace `xlab-backend/internal/httpapi/subscription_products.go` with:

```go
package httpapi

import (
	"context"
	"net/http"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

func (a *API) getActiveProducts(w http.ResponseWriter, r *http.Request) {
	if a.subscriptionReads != nil {
		a.writeReadServiceJSON(w, r, a.subscriptionReads.Active)
		return
	}
	a.proxyCoreJSON(w, r, "/subscription-products/active")
}

func (a *API) getProductSummary(w http.ResponseWriter, r *http.Request) {
	if a.subscriptionReads != nil {
		a.writeReadServiceJSON(w, r, a.subscriptionReads.Summary)
		return
	}
	a.proxyCoreJSON(w, r, "/subscription-products/summary")
}

func (a *API) getProductProgress(w http.ResponseWriter, r *http.Request) {
	if a.subscriptionReads != nil {
		a.writeReadServiceJSON(w, r, a.subscriptionReads.Progress)
		return
	}
	a.proxyCoreJSON(w, r, "/subscription-products/progress")
}

func (a *API) writeReadServiceJSON(w http.ResponseWriter, r *http.Request, fn func(context.Context, *core.User, string) (any, error)) {
	out, err := fn(r.Context(), userFromContext(r.Context()), tokenFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusBadGateway, "SUBSCRIPTION_READ_UNAVAILABLE", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
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

- [ ] **Step 4: Verify HTTP tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 5: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): route subscription endpoints through read service

Allow /xapi/v1 product subscription endpoints to use xlab mirror reads while preserving core proxy compatibility.
EOF
)"
```

---

### Task 7: Add full snapshot syncer

**Files:**
- Create: `xlab-backend/internal/subscriptions/syncer.go`
- Create: `xlab-backend/internal/subscriptions/syncer_test.go`

- [ ] **Step 1: Write failing syncer transaction test**

Create `xlab-backend/internal/subscriptions/syncer_test.go`:

```go
package subscriptions

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSyncerFullSnapshotUpdatesSyncState(t *testing.T) {
	coreDB, coreMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("core sqlmock.New error: %v", err)
	}
	defer coreDB.Close()
	xlabDB, xlabMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("xlab sqlmock.New error: %v", err)
	}
	defer xlabDB.Close()

	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	coreMock.ExpectQuery(regexp.QuoteMeta(coreProductsSQL)).WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "status", "product_family", "daily_limit_usd", "weekly_limit_usd", "monthly_limit_usd", "created_at", "updated_at"}).AddRow(int64(10), "gpt-pro", "GPT Pro", "desc", "active", "gpt", 10.0, 20.0, 30.0, now, now))
	coreMock.ExpectQuery(regexp.QuoteMeta(coreProductGroupsSQL)).WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "group_id", "group_name", "group_platform", "debit_multiplier", "status", "sort_order", "created_at", "updated_at"}).AddRow(int64(100), int64(10), int64(20), "GPT-4", "openai", 1.0, "active", 1, now, now))
	coreMock.ExpectQuery(regexp.QuoteMeta(coreUserProductSubscriptionsSQL)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "product_id", "status", "starts_at", "expires_at", "daily_usage_usd", "weekly_usage_usd", "monthly_usage_usd", "daily_carryover_in_usd", "daily_carryover_remaining_usd", "created_at", "updated_at"}).AddRow(int64(99), int64(7), int64(10), "active", now, now.Add(24*time.Hour), 1.0, 2.0, 3.0, 0.0, 0.0, now, now))

	xlabMock.ExpectBegin()
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertProductSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertProductGroupSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertUserProductSubscriptionSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectExec(regexp.QuoteMeta(upsertSyncStateSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
	xlabMock.ExpectCommit()

	syncer := NewSyncer(coreDB, xlabDB, func() time.Time { return now })
	if err := syncer.SyncOnce(context.Background()); err != nil {
		t.Fatalf("SyncOnce error: %v", err)
	}
	if err := coreMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet core expectations: %v", err)
	}
	if err := xlabMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet xlab expectations: %v", err)
	}
}
```

- [ ] **Step 2: Run syncer test and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because `NewSyncer` is undefined.

- [ ] **Step 3: Add syncer implementation with concrete SQL**

Create `xlab-backend/internal/subscriptions/syncer.go`:

```go
package subscriptions

import (
	"context"
	"database/sql"
	"time"
)

const coreProductsSQL = `
SELECT id, code, name, COALESCE(description, ''), status, product_family,
       daily_limit_usd, weekly_limit_usd, monthly_limit_usd, created_at, updated_at
FROM subscription_products
WHERE deleted_at IS NULL`

const coreProductGroupsSQL = `
SELECT spg.id, spg.product_id, spg.group_id, g.name, COALESCE(g.platform, ''),
       spg.debit_multiplier, spg.status, spg.sort_order, spg.created_at, spg.updated_at
FROM subscription_product_groups spg
JOIN groups g ON g.id = spg.group_id AND g.deleted_at IS NULL
WHERE spg.deleted_at IS NULL`

const coreUserProductSubscriptionsSQL = `
SELECT id, user_id, product_id, status, starts_at, expires_at,
       daily_usage_usd, weekly_usage_usd, monthly_usage_usd,
       daily_carryover_in_usd, daily_carryover_remaining_usd, created_at, updated_at
FROM user_product_subscriptions
WHERE deleted_at IS NULL
  AND (status = 'active' OR updated_at > NOW() - interval '30 days')`

const upsertProductSQL = `
INSERT INTO xlab_subscription_products (
    core_product_id, code, name, description, status, product_family,
    daily_limit_usd, weekly_limit_usd, monthly_limit_usd,
    source_created_at, source_updated_at, synced_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (core_product_id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    product_family = EXCLUDED.product_family,
    daily_limit_usd = EXCLUDED.daily_limit_usd,
    weekly_limit_usd = EXCLUDED.weekly_limit_usd,
    monthly_limit_usd = EXCLUDED.monthly_limit_usd,
    source_created_at = EXCLUDED.source_created_at,
    source_updated_at = EXCLUDED.source_updated_at,
    synced_at = EXCLUDED.synced_at`

const upsertProductGroupSQL = `
INSERT INTO xlab_subscription_product_groups (
    core_binding_id, core_product_id, core_group_id, group_name, group_platform,
    debit_multiplier, status, sort_order, source_created_at, source_updated_at, synced_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (core_binding_id) DO UPDATE SET
    core_product_id = EXCLUDED.core_product_id,
    core_group_id = EXCLUDED.core_group_id,
    group_name = EXCLUDED.group_name,
    group_platform = EXCLUDED.group_platform,
    debit_multiplier = EXCLUDED.debit_multiplier,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    source_created_at = EXCLUDED.source_created_at,
    source_updated_at = EXCLUDED.source_updated_at,
    synced_at = EXCLUDED.synced_at`

const upsertUserProductSubscriptionSQL = `
INSERT INTO xlab_user_product_subscriptions (
    core_subscription_id, core_user_id, core_product_id, status, started_at, expires_at,
    daily_usage_usd, weekly_usage_usd, monthly_usage_usd,
    daily_carryover_in_usd, daily_carryover_remaining_usd,
    source_created_at, source_updated_at, synced_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
ON CONFLICT (core_subscription_id) DO UPDATE SET
    core_user_id = EXCLUDED.core_user_id,
    core_product_id = EXCLUDED.core_product_id,
    status = EXCLUDED.status,
    started_at = EXCLUDED.started_at,
    expires_at = EXCLUDED.expires_at,
    daily_usage_usd = EXCLUDED.daily_usage_usd,
    weekly_usage_usd = EXCLUDED.weekly_usage_usd,
    monthly_usage_usd = EXCLUDED.monthly_usage_usd,
    daily_carryover_in_usd = EXCLUDED.daily_carryover_in_usd,
    daily_carryover_remaining_usd = EXCLUDED.daily_carryover_remaining_usd,
    source_created_at = EXCLUDED.source_created_at,
    source_updated_at = EXCLUDED.source_updated_at,
    synced_at = EXCLUDED.synced_at`

const upsertSyncStateSQL = `
INSERT INTO xlab_sync_state (source_name, last_success_at, row_count, created_at, updated_at)
VALUES ('product_subscriptions', $1, $2, $1, $1)
ON CONFLICT (source_name) DO UPDATE SET
    last_success_at = EXCLUDED.last_success_at,
    row_count = EXCLUDED.row_count,
    last_error = NULL,
    updated_at = EXCLUDED.updated_at`

type Syncer struct {
	coreDB *sql.DB
	xlabDB *sql.DB
	now    func() time.Time
}

func NewSyncer(coreDB *sql.DB, xlabDB *sql.DB, now func() time.Time) *Syncer {
	if now == nil {
		now = time.Now
	}
	return &Syncer{coreDB: coreDB, xlabDB: xlabDB, now: now}
}

func (s *Syncer) SyncOnce(ctx context.Context) error {
	now := s.now()
	productRows, err := s.coreDB.QueryContext(ctx, coreProductsSQL)
	if err != nil {
		return err
	}
	defer productRows.Close()
	tx, err := s.xlabDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	rowCount := 0
	for productRows.Next() {
		var id int64
		var code, name, description, status, family string
		var daily, weekly, monthly float64
		var createdAt, updatedAt time.Time
		if err := productRows.Scan(&id, &code, &name, &description, &status, &family, &daily, &weekly, &monthly, &createdAt, &updatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, upsertProductSQL, id, code, name, description, status, family, daily, weekly, monthly, createdAt, updatedAt, now); err != nil {
			return err
		}
		rowCount++
	}
	if err := productRows.Err(); err != nil {
		return err
	}

	if err := s.syncGroups(ctx, tx, now, &rowCount); err != nil {
		return err
	}
	if err := s.syncSubscriptions(ctx, tx, now, &rowCount); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, upsertSyncStateSQL, now, rowCount); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Syncer) syncGroups(ctx context.Context, tx *sql.Tx, syncedAt time.Time, rowCount *int) error {
	rows, err := s.coreDB.QueryContext(ctx, coreProductGroupsSQL)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, productID, groupID int64
		var name, platform, status string
		var multiplier float64
		var sortOrder int
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &productID, &groupID, &name, &platform, &multiplier, &status, &sortOrder, &createdAt, &updatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, upsertProductGroupSQL, id, productID, groupID, name, platform, multiplier, status, sortOrder, createdAt, updatedAt, syncedAt); err != nil {
			return err
		}
		*rowCount++
	}
	return rows.Err()
}

func (s *Syncer) syncSubscriptions(ctx context.Context, tx *sql.Tx, syncedAt time.Time, rowCount *int) error {
	rows, err := s.coreDB.QueryContext(ctx, coreUserProductSubscriptionsSQL)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, userID, productID int64
		var status string
		var startsAt, expiresAt, createdAt, updatedAt time.Time
		var dailyUsage, weeklyUsage, monthlyUsage, carryoverIn, carryoverRemaining float64
		if err := rows.Scan(&id, &userID, &productID, &status, &startsAt, &expiresAt, &dailyUsage, &weeklyUsage, &monthlyUsage, &carryoverIn, &carryoverRemaining, &createdAt, &updatedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, upsertUserProductSubscriptionSQL, id, userID, productID, status, startsAt, expiresAt, dailyUsage, weeklyUsage, monthlyUsage, carryoverIn, carryoverRemaining, createdAt, updatedAt, syncedAt); err != nil {
			return err
		}
		*rowCount++
	}
	return rows.Err()
}
```

- [ ] **Step 4: Verify syncer tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 5: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): sync product subscription mirror data

Copy core product subscription read data into xlab mirror tables with sync-state tracking.
EOF
)"
```

---

### Task 8: Wire runtime in server main

**Files:**
- Modify: `xlab-backend/cmd/server/main.go`
- Create: `xlab-backend/cmd/server/main_test.go`

- [ ] **Step 1: Write failing read-source mapping test**

Create `xlab-backend/cmd/server/main_test.go`:

```go
package main

import (
	"testing"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/subscriptions"
)

func TestSubscriptionReadSourceMapping(t *testing.T) {
	cases := []struct {
		in   config.SubscriptionReadSource
		want subscriptions.ReadSource
	}{
		{config.SubscriptionReadSourceCore, subscriptions.ReadSourceCore},
		{config.SubscriptionReadSourceHybrid, subscriptions.ReadSourceHybrid},
		{config.SubscriptionReadSourceXlab, subscriptions.ReadSourceXlab},
	}
	for _, tc := range cases {
		if got := subscriptionReadSource(tc.in); got != tc.want {
			t.Fatalf("subscriptionReadSource(%q) = %q", tc.in, got)
		}
	}
}
```

- [ ] **Step 2: Run server test and verify failure**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: FAIL because `subscriptionReadSource` is undefined.

- [ ] **Step 3: Replace main wiring**

Replace `xlab-backend/cmd/server/main.go` with:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/core"
	"github.com/2018x5zzt/xlab-backend/internal/httpapi"
	"github.com/2018x5zzt/xlab-backend/internal/storage"
	"github.com/2018x5zzt/xlab-backend/internal/subscriptions"
)

func main() {
	cfg := config.Load()
	client := core.NewClient(cfg.CoreAPIBaseURL, cfg.CoreTimeout)
	router := httpapi.NewRouter(client, nil)

	ctx := context.Background()
	if cfg.XlabDatabaseURL != "" {
		xlabDB, err := storage.OpenPostgres(ctx, cfg.XlabDatabaseURL)
		if err != nil {
			log.Fatalf("open xlab db: %v", err)
		}
		defer xlabDB.Close()
		if err := storage.RunMigrations(ctx, xlabDB); err != nil {
			log.Fatalf("run xlab migrations: %v", err)
		}

		repo := subscriptions.NewRepository(xlabDB)
		readSvc := subscriptions.NewService(subscriptionReadSource(cfg.SubscriptionReadSource), repo, client, cfg.SubscriptionSyncStaleAfter, time.Now)
		router = httpapi.NewRouter(client, readSvc)

		if cfg.SubscriptionSyncEnabled {
			if cfg.CoreDatabaseURL == "" {
				log.Fatal("CORE_DATABASE_URL is required when XLAB_SUBSCRIPTION_SYNC_ENABLED=true")
			}
			coreDB, err := storage.OpenPostgres(ctx, cfg.CoreDatabaseURL)
			if err != nil {
				log.Fatalf("open core db: %v", err)
			}
			defer coreDB.Close()
			syncer := subscriptions.NewSyncer(coreDB, xlabDB, time.Now)
			go runSubscriptionSyncLoop(ctx, syncer, cfg.SubscriptionSyncInterval)
		}
	}

	log.Printf("xlab backend listening on %s, core=%s", cfg.ServerAddr, cfg.CoreAPIBaseURL)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
}

func subscriptionReadSource(source config.SubscriptionReadSource) subscriptions.ReadSource {
	switch source {
	case config.SubscriptionReadSourceHybrid:
		return subscriptions.ReadSourceHybrid
	case config.SubscriptionReadSourceXlab:
		return subscriptions.ReadSourceXlab
	default:
		return subscriptions.ReadSourceCore
	}
}

func runSubscriptionSyncLoop(ctx context.Context, syncer *subscriptions.Syncer, interval time.Duration) {
	if err := syncer.SyncOnce(ctx); err != nil {
		log.Printf("product subscription sync failed: %v", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := syncer.SyncOnce(ctx); err != nil {
				log.Printf("product subscription sync failed: %v", err)
			}
		}
	}
}
```

- [ ] **Step 4: Verify server and internal tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 5: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
feat(xlab): wire subscription mirror runtime

Open xlab DB connections, run mirror migrations, optionally sync from core DB, and serve subscription reads through the configured source mode.
EOF
)"
```

---

### Task 9: Update deployment script and runbook

**Files:**
- Modify: `deploy-xlab-backend.sh`
- Modify: `docs/superpowers/specs/2026-06-07-xlab-backend-docker-phase2-runbook.md`

- [ ] **Step 1: Verify current deployment script syntax**

Run:

```bash
cd /root/sub2api-src
bash -n deploy-xlab-backend.sh
```

Expected: PASS.

- [ ] **Step 2: Add env passthrough to deployment script**

In `deploy-xlab-backend.sh`, add near existing env defaults:

```bash
XLAB_DATABASE_URL="${XLAB_DATABASE_URL:-}"
CORE_DATABASE_URL="${CORE_DATABASE_URL:-}"
SUBSCRIPTION_READ_SOURCE="${XLAB_SUBSCRIPTION_READ_SOURCE:-core}"
SUBSCRIPTION_SYNC_ENABLED="${XLAB_SUBSCRIPTION_SYNC_ENABLED:-false}"
SUBSCRIPTION_SYNC_INTERVAL_SECONDS="${XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS:-300}"
SUBSCRIPTION_SYNC_STALE_SECONDS="${XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS:-600}"
```

Add to `docker run`:

```bash
  -e XLAB_DATABASE_URL='${XLAB_DATABASE_URL}' \
  -e CORE_DATABASE_URL='${CORE_DATABASE_URL}' \
  -e XLAB_SUBSCRIPTION_READ_SOURCE='${SUBSCRIPTION_READ_SOURCE}' \
  -e XLAB_SUBSCRIPTION_SYNC_ENABLED='${SUBSCRIPTION_SYNC_ENABLED}' \
  -e XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS='${SUBSCRIPTION_SYNC_INTERVAL_SECONDS}' \
  -e XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS='${SUBSCRIPTION_SYNC_STALE_SECONDS}' \
```

- [ ] **Step 3: Update runbook with concrete runtime commands**

Append to `docs/superpowers/specs/2026-06-07-xlab-backend-docker-phase2-runbook.md`:

```md
## Phase 3A product subscription read mirror

Rollback mode keeps xlab-backend proxying core:

```bash
cd /root/sub2api-src
XLAB_SUBSCRIPTION_READ_SOURCE=core \
XLAB_SUBSCRIPTION_SYNC_ENABLED=false \
./deploy-xlab-backend.sh
```

Sync-only rollout requires DSNs to be exported first:

```bash
export XLAB_DATABASE_URL="$XLAB_DATABASE_URL"
export CORE_DATABASE_URL="$CORE_DATABASE_URL"
test -n "$XLAB_DATABASE_URL"
test -n "$CORE_DATABASE_URL"

cd /root/sub2api-src
XLAB_SUBSCRIPTION_SYNC_ENABLED=true \
XLAB_SUBSCRIPTION_READ_SOURCE=core \
./deploy-xlab-backend.sh
```

Hybrid read rollout keeps fallback enabled:

```bash
export XLAB_DATABASE_URL="$XLAB_DATABASE_URL"
export CORE_DATABASE_URL="$CORE_DATABASE_URL"
test -n "$XLAB_DATABASE_URL"
test -n "$CORE_DATABASE_URL"

cd /root/sub2api-src
XLAB_SUBSCRIPTION_SYNC_ENABLED=true \
XLAB_SUBSCRIPTION_READ_SOURCE=hybrid \
./deploy-xlab-backend.sh
```

Runtime rollback:

```bash
cd /root/sub2api-src
XLAB_SUBSCRIPTION_READ_SOURCE=core \
XLAB_SUBSCRIPTION_SYNC_ENABLED=false \
./deploy-xlab-backend.sh
```
```

- [ ] **Step 4: Verify deployment script syntax**

Run:

```bash
cd /root/sub2api-src
bash -n deploy-xlab-backend.sh
```

Expected: PASS.

- [ ] **Step 5: Commit checkpoint when commits are explicitly requested**

Run only in an execution session where the user requested commits:

```bash
cd /root/sub2api-src
docs(xlab): document phase 3a mirror rollout

Expose subscription mirror runtime settings in deployment and document core, hybrid, and rollback modes.
EOF
)"
```

---

### Task 10: Verification and production rollout gate

**Files:**
- No source changes unless verification finds issues.

- [ ] **Step 1: Run all xlab-backend tests**

Run:

```bash
cd /root/sub2api-src/xlab-backend
```

Expected: PASS.

- [ ] **Step 2: Build xlab-backend binary**

Run:

```bash
cd /root/sub2api-src/xlab-backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/xlab-backend ./cmd/server
```

Expected: PASS.

- [ ] **Step 3: Run focused core product-subscription regressions**

Run:

```bash
cd /root/sub2api-src/backend
```

Expected: PASS. If integration tests require external Postgres and the harness is unavailable, stop and report the exact failing package and setup error before production rollout.

- [ ] **Step 4: Deploy rollback mode first**

Run on production:

```bash
cd /root/sub2api-src
XLAB_SUBSCRIPTION_READ_SOURCE=core \
XLAB_SUBSCRIPTION_SYNC_ENABLED=false \
./deploy-xlab-backend.sh
```

Expected: `xlab-backend` is healthy and unauthenticated `/xapi/v1/subscription-products/active` returns JSON 401.

- [ ] **Step 5: Deploy sync enabled with reads still on core**

Before running, export production DSNs from the secret source used by the server:

```bash
test -n "$XLAB_DATABASE_URL"
test -n "$CORE_DATABASE_URL"
```

Then run:

```bash
cd /root/sub2api-src
XLAB_SUBSCRIPTION_SYNC_ENABLED=true \
XLAB_SUBSCRIPTION_READ_SOURCE=core \
./deploy-xlab-backend.sh
```

Expected: `xlab-backend` is healthy and logs a successful product subscription sync.

- [ ] **Step 6: Switch to hybrid after sync audit**

Run:

```bash
cd /root/sub2api-src
XLAB_SUBSCRIPTION_SYNC_ENABLED=true \
XLAB_SUBSCRIPTION_READ_SOURCE=hybrid \
./deploy-xlab-backend.sh
```

Expected: selected logged-in users' subscription pages load; unauthenticated `/xapi/v1/subscription-products/active` still returns JSON 401.

- [ ] **Step 7: Verify origin routes**

Run:

```bash
for domain in xlabapi.com xlabapi.top xlabapi.space apitest1.xlabapi.com; do
  curl --noproxy '*' --resolve "${domain}:443:127.0.0.1" -k -i -sS "https://${domain}/xapi/v1/subscription-products/active" | sed -n '1,12p'
done
```

Expected: all origin responses are JSON 401 when unauthenticated. Public `apitest1.xlabapi.com` can remain Cloudflare-blocked; origin validation is sufficient for Nginx and backend routing.

- [ ] **Step 8: Check repository status**

Run:

```bash
cd /root/sub2api-src
```

Expected: only intended source, docs, `go.mod`, and `go.sum` changes are present. Built binaries are not staged.

---

## Self-Review

- Spec coverage: The plan covers xlab DB config, migrations, mirror tables, sync job, stale detection, core fallback, response contract, auth user context, deployment envs, rollout, rollback, and verification gates.
- Placeholder scan: The plan avoids repository placeholders. Production DSNs are runtime environment variables and are validated with `test -n` before rollout commands.
- Type consistency: Read-source names are consistently `core`, `hybrid`, and `xlab`; sync source is consistently `product_subscriptions`; response types match frontend-v2/core DTO names and JSON fields.

