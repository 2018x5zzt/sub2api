# Xlabapi Upstream v0.1.121 Stability Gateway Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a verified integration branch that ports upstream `v0.1.119..v0.1.121` stability, scheduler, security, and gateway/API compatibility fixes into `xlabapi` without importing large upstream feature migrations.

**Architecture:** Work in an isolated git worktree from `xlabapi`, classify each candidate upstream patch before changing behavior, then land focused commits by concern. Prefer behavioral transplants over broad merges for gateway and scheduler files because `xlabapi` has branch-specific routing, billing, image, and model-compatibility logic.

**Tech Stack:** Go 1.26.1, Gin, Redis/go-redis, Wire, Ent, Testify, Vue 3, Vitest, TypeScript.

---

## Scope Boundary

Use the design spec at `docs/superpowers/specs/2026-04-30-xlabapi-upstream-v0121-stability-gateway-sync-design.md`.

Included upstream source window:

```bash
git -C /root/sub2api-src log --oneline --reverse v0.1.119..v0.1.121
```

Primary included commits:

```text
798fd673 feat(httputil): decode compressed request bodies (zstd/gzip/deflate)
40feb86b fix(httputil): add decompression bomb guard and fix errcheck lint
8bf2a7b8 fix(scheduler): resolve SetSnapshot race conditions and remove usage throttle
733627cf fix: improve sticky session scheduling
53f919f8 fix(api-key): reset rate limit usage cache
30220903 fix(anthropic): drop empty Read.pages in responses-to-anthropic tool input
615557ec fix(openai): avoid implicit image sticky sessions
9fe02bba fix(openai): strip unsupported passthrough fields
3d4ca5e8 fix(openai): preserve current Codex compact payload fields
7452fad8 fix(openai): drop reasoning items from /v1/responses input on OAuth path
04b2866f fix: use Responses-compatible function tool_choice format
63275735 fix(gateway): wrap Anthropic stream EOF as failover error before client output
4c474616 fix(gateway): emit Anthropic-standard SSE error events and failover body
d78478e8 fix(gateway): sanitize stream errors to avoid leaking infrastructure topology
28dc34b6 fix(openai): avoid inferred WS continuation on explicit tool replay
094e1171 fix(openai): infer previous response for item references
```

Low-risk candidates to classify and either include or defer:

```text
73b87299 feat: add Anthropic cache TTL injection switch
f084d30d fix: restore table page-size localStorage persistence
4b6954f9 feat(ops): allow retention days = 0 to wipe table on each scheduled cleanup
9d801595 test: update admin settings contract fields
```

Excluded from this plan:

```text
30f55a1f OpenAI Fast/Flex Policy
25c7b0d9 / 2ab6b34f / a161f9d0 account bulk edit chain
6d11f9ed / 489a4d93 / 93d91e20 Vertex service account chain
payment, channel monitor, profile identity, OAuth adoption, broad settings-page migrations
backend/cmd/server/VERSION bump to 0.1.121
```

## File Map

### Documentation and branch hygiene

- Create: `docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md`
- Worktree: `/root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430`

### HTTP request body safety

- Modify: `backend/internal/pkg/httputil/body.go`
- Create or modify: `backend/internal/pkg/httputil/body_test.go`
- Check existing callers under `backend/internal/handler/*` that use `ReadRequestBodyWithPrealloc`.

### Scheduler and sticky sessions

- Modify: `backend/internal/repository/scheduler_cache.go`
- Modify: `backend/internal/service/scheduler_cache.go`
- Modify: `backend/internal/service/scheduler_snapshot_service.go`
- Modify: `backend/internal/service/gateway_service.go`
- Modify: `backend/internal/handler/gateway_handler.go`
- Modify or create focused tests:
  - `backend/internal/repository/scheduler_cache_unit_test.go`
  - `backend/internal/service/scheduler_snapshot_hydration_test.go`
  - `backend/internal/service/gateway_service_test.go` or existing sticky/session selection tests

### API key rate-limit cache reset

- Modify: `backend/internal/handler/admin/apikey_handler.go`
- Modify: `backend/internal/handler/admin/apikey_handler_test.go`
- Modify: `backend/internal/handler/admin/admin_service_stub_test.go`
- Modify: `backend/internal/service/admin_service.go`
- Modify: `backend/internal/service/billing_cache_service.go`
- Modify: `backend/internal/service/wire.go`
- Regenerate: `backend/cmd/server/wire_gen.go`

### OpenAI, Anthropic, Codex, Responses compatibility

- Modify: `backend/internal/pkg/apicompat/responses_to_anthropic.go`
- Modify: `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
- Modify: `backend/internal/pkg/apicompat/anthropic_to_responses.go`
- Modify: `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- Modify tests under `backend/internal/pkg/apicompat/*_test.go`
- Modify: `backend/internal/service/openai_codex_transform.go`
- Modify: `backend/internal/service/openai_gateway_service.go`
- Modify: `backend/internal/service/openai_gateway_chat_completions.go`
- Modify: `backend/internal/service/openai_gateway_messages.go`
- Modify: `backend/internal/service/openai_ws_forwarder.go`
- Modify: `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- Modify or create focused tests under `backend/internal/service/*openai*test.go`, `backend/internal/service/gateway_streaming_test.go`, and `backend/internal/service/openai_passthrough_normalization_test.go`

### Low-risk optional items

- Anthropic cache TTL:
  - `backend/internal/handler/admin/setting_handler.go`
  - `backend/internal/handler/dto/settings.go`
  - `backend/internal/service/domain_constants.go`
  - `backend/internal/service/gateway_service.go`
  - `backend/internal/service/setting_service.go`
  - `backend/internal/service/settings_view.go`
- Frontend page-size persistence:
  - `frontend/src/components/common/Pagination.vue`
  - `frontend/src/composables/usePersistedPageSize.ts`
  - `frontend/src/composables/useTableLoader.ts`
- Ops retention:
  - `backend/internal/service/ops_cleanup_service.go`
  - `backend/internal/service/ops_cleanup_service_test.go`
  - `backend/internal/service/ops_settings.go`
  - `frontend/src/views/admin/ops/components/OpsSettingsDialog.vue`

---

### Task 1: Create Isolated Integration Worktree

**Files:**
- Create worktree: `/root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430`

- [ ] **Step 1: Confirm source worktree is clean enough for branch creation**

Run:

```bash
git -C /root/sub2api-src status --short --branch
```

Expected:

```text
## xlabapi...origin/xlabapi [ahead 1]
```

The only ahead commit should be `4d14bcba docs: add upstream v0121 stability sync design`.

- [ ] **Step 2: Create the isolated worktree**

Run:

```bash
git -C /root/sub2api-src worktree add /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 -b integrate/xlabapi-upstream-v0121-stability-gateway-20260430 xlabapi
```

Expected:

```text
Preparing worktree (new branch 'integrate/xlabapi-upstream-v0121-stability-gateway-20260430')
HEAD is now at 4d14bcba docs: add upstream v0121 stability sync design
```

- [ ] **Step 3: Verify worktree branch and cleanliness**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 status --short --branch
```

Expected:

```text
## integrate/xlabapi-upstream-v0121-stability-gateway-20260430
```

No modified files should be listed.

---

### Task 2: Create Patch Classification Notes

**Files:**
- Create: `docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md`

- [ ] **Step 1: Create the initial classification file**

Create `docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md` in the integration worktree with this content:

```markdown
# Upstream v0.1.121 Stability Gateway Patch Classification

**Date:** 2026-04-30
**Branch:** integrate/xlabapi-upstream-v0121-stability-gateway-20260430
**Base:** xlabapi at 4d14bcba

## Classification Legend

- already equivalent: xlabapi already has behavior and tests or stronger local behavior
- partially present: xlabapi has part of the behavior; this branch adds the missing behavior
- applicable: this branch ports the behavior
- not applicable: upstream path does not exist here or is intentionally replaced
- deferred: useful but belongs to a later product-feature migration

## Primary Fixes

| Commit | Subject | Classification | Evidence |
| --- | --- | --- | --- |
| 798fd673 | decode compressed request bodies | applicable | `body.go` currently only copies raw bytes |
| 40feb86b | decompression bomb guard | applicable | apply together with `798fd673` |
| 8bf2a7b8 | scheduler SetSnapshot race fixes | applicable | current `SetSnapshot` switches active version without CAS and deletes old snapshot immediately |
| 733627cf | sticky session scheduling | partially present | current xlabapi has sticky prefetch but still checks group membership in snapshot path |
| 53f919f8 | admin API key rate-limit reset cache invalidation | partially present | user key update invalidates cache; admin key update lacks reset field/method |
| 30220903 | drop empty Read.pages | applicable | no `sanitizeAnthropicToolUseInput` helper exists in apicompat |
| 615557ec | avoid implicit image sticky sessions | classify during Task 6 | compare current async image-job behavior before editing |
| 9fe02bba | strip unsupported passthrough fields | classify during Task 6 | compare current Codex and passthrough normalization tests |
| 3d4ca5e8 | preserve Codex compact payload fields | classify during Task 6 | compare current compact payload handling |
| 7452fad8 | drop reasoning items on OAuth path | classify during Task 6 | compare current `openai_codex_transform.go` |
| 04b2866f | Responses-compatible function tool_choice | classify during Task 6 | compare current apicompat tool-choice converters |
| 63275735 | wrap Anthropic stream EOF failover | classify during Task 7 | compare current `gateway_streaming_test.go` |
| 4c474616 | Anthropic-standard SSE error events | classify during Task 7 | compare current stream error emission |
| d78478e8 | sanitize stream errors | classify during Task 7 | compare current `UpstreamFailoverError` response text |
| 28dc34b6 | avoid inferred WS continuation on explicit tool replay | classify during Task 8 | compare current ingress session tests |
| 094e1171 | infer previous response for item references | classify during Task 8 | compare current item reference logic |

## Low-Risk Candidates

| Commit | Subject | Classification | Evidence |
| --- | --- | --- | --- |
| 73b87299 | Anthropic cache TTL switch | deferred until Task 9 inspection | settings and gateway behavior must stay small |
| f084d30d | table page-size localStorage | deferred until Task 9 inspection | frontend-only fix may be safe |
| 4b6954f9 | ops retention days = 0 | deferred until Task 9 inspection | backend and ops UI both touched |
| 9d801595 | admin settings contract tests | deferred until Task 9 inspection | include only if settings fields change |

## Exclusions Confirmed

- OpenAI Fast/Flex Policy is not included.
- Account bulk edit is not included.
- Vertex service account support is not included.
- Payment, monitor, profile identity, OAuth adoption, and broad settings-page migrations are not included.
- `backend/cmd/server/VERSION` is not bumped to `0.1.121`.
```

- [ ] **Step 2: Commit the classification baseline**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "docs: classify upstream v0121 stability candidates"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] docs: classify upstream v0121 stability candidates
```

---

### Task 3: HTTP Body Compression And Bomb Guard

**Files:**
- Modify: `backend/internal/pkg/httputil/body.go`
- Create: `backend/internal/pkg/httputil/body_test.go`

- [ ] **Step 1: Write failing body tests**

Create `backend/internal/pkg/httputil/body_test.go` with focused tests adapted from upstream `798fd673` plus one explicit size-limit test:

```go
package httputil

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"net/http"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"
)

const samplePayload = `{"model":"gpt-5.5","input":"hi","stream":false}`

func newRequestWithBody(t *testing.T, body []byte, encoding string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	require.NoError(t, err)
	if encoding != "" {
		req.Header.Set("Content-Encoding", encoding)
	}
	req.ContentLength = int64(len(body))
	return req
}

func TestReadRequestBodyWithPrealloc_DecodesZstd(t *testing.T) {
	enc, err := zstd.NewWriter(nil)
	require.NoError(t, err)
	compressed := enc.EncodeAll([]byte(samplePayload), nil)
	enc.Close()

	req := newRequestWithBody(t, compressed, "zstd")
	got, err := ReadRequestBodyWithPrealloc(req)
	require.NoError(t, err)
	require.Equal(t, samplePayload, string(got))
	require.Empty(t, req.Header.Get("Content-Encoding"))
	require.Equal(t, int64(len(samplePayload)), req.ContentLength)
}

func TestReadRequestBodyWithPrealloc_DecodesGzip(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(samplePayload))
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	req := newRequestWithBody(t, buf.Bytes(), "gzip")
	got, err := ReadRequestBodyWithPrealloc(req)
	require.NoError(t, err)
	require.Equal(t, samplePayload, string(got))
}

func TestReadRequestBodyWithPrealloc_DecodesDeflate(t *testing.T) {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	_, err := zw.Write([]byte(samplePayload))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	req := newRequestWithBody(t, buf.Bytes(), "deflate")
	got, err := ReadRequestBodyWithPrealloc(req)
	require.NoError(t, err)
	require.Equal(t, samplePayload, string(got))
}

func TestReadRequestBodyWithPrealloc_RejectsUnsupportedEncoding(t *testing.T) {
	req := newRequestWithBody(t, []byte(samplePayload), "br")
	_, err := ReadRequestBodyWithPrealloc(req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "br")
}

func TestReadRequestBodyWithPrealloc_RejectsCorruptZstd(t *testing.T) {
	req := newRequestWithBody(t, []byte("not actually zstd"), "zstd")
	_, err := ReadRequestBodyWithPrealloc(req)
	require.Error(t, err)
}

func TestReadRequestBodyWithPrealloc_RespectsIdentityEncoding(t *testing.T) {
	req := newRequestWithBody(t, []byte(samplePayload), "identity")
	got, err := ReadRequestBodyWithPrealloc(req)
	require.NoError(t, err)
	require.Equal(t, samplePayload, string(got))
}

func TestReadRequestBodyWithPrealloc_LimitsDecompressedBody(t *testing.T) {
	huge := strings.Repeat("x", maxDecompressedBodySize+1)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(huge))
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	req := newRequestWithBody(t, buf.Bytes(), "gzip")
	got, err := ReadRequestBodyWithPrealloc(req)
	require.NoError(t, err)
	require.Len(t, got, maxDecompressedBodySize)
}
```

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/pkg/httputil -run 'ReadRequestBodyWithPrealloc' -count=1
```

Expected:

```text
FAIL
```

The failure should mention `undefined: maxDecompressedBodySize` or show zstd/gzip/deflate bodies are not decoded.

- [ ] **Step 3: Implement compressed body decoding and decompression limit**

Modify `backend/internal/pkg/httputil/body.go` so the imports and helper logic match this structure:

```go
import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/klauspost/compress/zstd"
)

const (
	requestBodyReadInitCap    = 512
	requestBodyReadMaxInitCap = 1 << 20
	maxDecompressedBodySize   = 64 << 20
)
```

Add this logic after the raw body is copied:

```go
raw := buf.Bytes()

enc := strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Encoding")))
if enc == "" || enc == "identity" {
	return raw, nil
}

decoded, err := decompressRequestBody(enc, raw)
if err != nil {
	return nil, fmt.Errorf("decode Content-Encoding %q: %w", enc, err)
}

req.Header.Del("Content-Encoding")
req.Header.Del("Content-Length")
req.ContentLength = int64(len(decoded))

return decoded, nil
```

Add the helper:

```go
func decompressRequestBody(encoding string, raw []byte) ([]byte, error) {
	switch encoding {
	case "zstd":
		dec, err := zstd.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		defer dec.Close()
		return io.ReadAll(io.LimitReader(dec, maxDecompressedBodySize))
	case "gzip", "x-gzip":
		gr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		defer func() { _ = gr.Close() }()
		return io.ReadAll(io.LimitReader(gr, maxDecompressedBodySize))
	case "deflate":
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		defer func() { _ = zr.Close() }()
		return io.ReadAll(io.LimitReader(zr, maxDecompressedBodySize))
	default:
		return nil, errors.New("unsupported Content-Encoding")
	}
}
```

- [ ] **Step 4: Run focused and caller tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/pkg/httputil ./internal/handler -run 'ReadRequestBodyWithPrealloc' -count=1
```

Expected:

```text
ok  	github.com/Wei-Shaw/sub2api/internal/pkg/httputil
ok  	github.com/Wei-Shaw/sub2api/internal/handler
```

- [ ] **Step 5: Commit HTTP body fix**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/pkg/httputil/body.go backend/internal/pkg/httputil/body_test.go
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(httputil): decode compressed request bodies safely"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(httputil): decode compressed request bodies safely
```

---

### Task 4: Scheduler Snapshot Race Fixes

**Files:**
- Modify: `backend/internal/repository/scheduler_cache.go`
- Modify: `backend/internal/service/scheduler_cache.go`
- Modify: `backend/internal/service/scheduler_snapshot_service.go`
- Modify: `backend/internal/repository/scheduler_cache_unit_test.go`

- [ ] **Step 1: Write unit tests for the new scheduler cache contract**

Append to `backend/internal/repository/scheduler_cache_unit_test.go`:

```go
func TestSchedulerSnapshotKeyPrefixForActivation(t *testing.T) {
	bucket := service.SchedulerBucket{GroupID: 7, Platform: service.PlatformAnthropic, Mode: service.SchedulerModeSingle}
	require.Equal(t, "sched:7:anthropic:single:v42", schedulerSnapshotKey(bucket, "42"))
	require.Equal(t, "sched:7:anthropic:single:v", schedulerSnapshotKeyPrefix(bucket))
}

func TestSchedulerCacheUnlockBucketMethodExists(t *testing.T) {
	var cache service.SchedulerCache = &schedulerCache{}
	require.Implements(t, (*service.SchedulerCache)(nil), cache)
}
```

- [ ] **Step 2: Run the scheduler unit tests and verify RED**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/repository -run 'SchedulerSnapshotKeyPrefixForActivation|SchedulerCacheUnlockBucketMethodExists' -count=1
```

Expected:

```text
FAIL
```

The failure should mention `undefined: schedulerSnapshotKeyPrefix` or that `*schedulerCache` does not implement `service.SchedulerCache`.

- [ ] **Step 3: Add unlock support to the service interface**

In `backend/internal/service/scheduler_cache.go`, add this method after `TryLockBucket`:

```go
// UnlockBucket releases a bucket rebuild lock.
UnlockBucket(ctx context.Context, bucket SchedulerBucket) error
```

- [ ] **Step 4: Add CAS activation script and unlock implementation**

In `backend/internal/repository/scheduler_cache.go`, add:

```go
const (
	snapshotGraceTTLSeconds = 60
)

var activateSnapshotScript = redis.NewScript(`
local currentActive = redis.call('GET', KEYS[1])
local newVersion = tonumber(ARGV[1])

if currentActive ~= false then
	local curVersion = tonumber(currentActive)
	if curVersion and newVersion < curVersion then
		redis.call('DEL', KEYS[4])
		return 0
	end
end

redis.call('SET', KEYS[1], ARGV[1])
redis.call('SET', KEYS[2], '1')
redis.call('SADD', KEYS[3], ARGV[2])

if currentActive ~= false and currentActive ~= ARGV[1] then
	redis.call('EXPIRE', ARGV[3] .. currentActive, tonumber(ARGV[4]))
end

return 1
`)
```

Add:

```go
func schedulerSnapshotKeyPrefix(bucket service.SchedulerBucket) string {
	return fmt.Sprintf("%s%d:%s:%s:v", schedulerSnapshotPrefix, bucket.GroupID, bucket.Platform, bucket.Mode)
}

func (c *schedulerCache) UnlockBucket(ctx context.Context, bucket service.SchedulerBucket) error {
	key := schedulerBucketKey(schedulerLockPrefix, bucket)
	return c.rdb.Del(ctx, key).Err()
}
```

- [ ] **Step 5: Replace active-version switching inside `SetSnapshot`**

In `SetSnapshot`, remove the `oldActive` lookup and remove the immediate `Del` of the old snapshot. Keep `xlabapi`'s existing account serialization model; do not import upstream `schedulerAccountMetaPrefix` unless a later test proves it is required.

After writing account keys and snapshot members, run:

```go
activeKey := schedulerBucketKey(schedulerActivePrefix, bucket)
readyKey := schedulerBucketKey(schedulerReadyPrefix, bucket)
keys := []string{activeKey, readyKey, schedulerBucketSetKey, snapshotKey}
args := []any{versionStr, bucket.String(), schedulerSnapshotKeyPrefix(bucket), snapshotGraceTTLSeconds}
if _, err := activateSnapshotScript.Run(ctx, c.rdb, keys, args...).Result(); err != nil {
	return err
}
```

The final `SetSnapshot` must preserve the current branch's behavior of treating empty snapshots as cache misses in `GetSnapshot`.

- [ ] **Step 6: Release rebuild locks immediately**

In `backend/internal/service/scheduler_snapshot_service.go`, inside `rebuildBucket`, add this immediately after a successful `TryLockBucket`:

```go
defer func() {
	_ = s.cache.UnlockBucket(ctx, bucket)
}()
```

- [ ] **Step 7: Run scheduler-focused tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/repository ./internal/service -run 'Scheduler|Snapshot' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 8: Commit scheduler race fix**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/repository/scheduler_cache.go backend/internal/service/scheduler_cache.go backend/internal/service/scheduler_snapshot_service.go backend/internal/repository/scheduler_cache_unit_test.go
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(scheduler): activate snapshots atomically"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(scheduler): activate snapshots atomically
```

---

### Task 5: Sticky Session Scheduling Compatibility

**Files:**
- Modify: `backend/internal/service/gateway_service.go`
- Modify or add: focused sticky session tests in `backend/internal/service/gateway_service_test.go`
- Modify classification file from Task 2

- [ ] **Step 1: Inspect upstream sticky patch against current branch**

Run:

```bash
git -C /root/sub2api-src show --no-ext-diff --unified=80 733627cf -- backend/internal/service/gateway_service.go backend/internal/handler/gateway_handler.go
```

Expected:

```text
commit 733627cf...
```

Record in the classification file that the upstream handler patch contains debug logging that should not be ported as-is.

- [ ] **Step 2: Write a focused sticky regression test**

Add a test in the existing `gateway_service` test file that already defines a lightweight `GatewayService` account-selection harness. If no suitable harness exists, create `backend/internal/service/gateway_sticky_selection_test.go` with a narrow package-level test around the no-routing sticky gate using local helper functions already present in `gateway_service.go`.

The test requirement is:

```go
// Given a sticky-bound account loaded from a scheduler snapshot without AccountGroups,
// SelectAccountWithLoadAwareness must still allow the sticky account when platform,
// model, quota, window-cost, RPM, and schedulability checks pass.
```

Use an account with:

```go
service.Account{
	ID:          101,
	Platform:    service.PlatformAnthropic,
	Type:        service.AccountTypeOAuth,
	Status:      service.StatusActive,
	Schedulable: true,
	Concurrency: 1,
	GroupIDs:    []int64{7},
}
```

Expected RED before implementation:

```text
selected account is not the sticky account or selection returns ErrNoAvailableAccounts
```

- [ ] **Step 3: Remove the stale group-membership gate in no-routing sticky path**

In `backend/internal/service/gateway_service.go`, locate the no-routing sticky block near:

```go
if len(routingAccountIDs) == 0 && sessionHash != "" && stickyAccountID > 0 && !isExcluded(stickyAccountID) {
```

Replace the condition that includes `s.isAccountInGroup(account, groupID)` with explicit checks:

```go
platformOK := s.isAccountAllowedForPlatform(account, platform, useMixed)
modelSupported := requestedModel == "" || s.isModelSupportedByAccountWithContext(ctx, account, requestedModel)
modelSchedulable := s.isAccountSchedulableForModelSelection(ctx, account, requestedModel)
channelAllowed := s.isAccountAllowedByChannelPricing(ctx, groupID, account, requestedModel, needsUpstreamRestrictionCheck)
quotaOK := s.isAccountSchedulableForQuota(account)
dynamicPricingOK := s.isAccountSchedulableForDynamicPricing(ctx, account, groupID)
windowCostOK := s.isAccountSchedulableForWindowCost(ctx, account, true)
rpmOK := s.isAccountSchedulableForRPM(ctx, account, true)
schedulable := s.isAccountSchedulableForSelection(account)

if !clearSticky && platformOK && modelSupported && channelAllowed && modelSchedulable && quotaOK && dynamicPricingOK && windowCostOK && rpmOK && schedulable {
	// keep the existing tryAcquireAccountSlot and wait-plan body
}
```

Do not port `[DEBUG-STICKY]` handler log additions from upstream.

- [ ] **Step 4: Run sticky-focused tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service -run 'Sticky|SelectAccountWithLoadAwareness' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit sticky session fix**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/service/gateway_service.go backend/internal/service/*sticky*test.go backend/internal/service/gateway_service_test.go docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(gateway): preserve sticky sessions from scheduler snapshots"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(gateway): preserve sticky sessions from scheduler snapshots
```

---

### Task 6: Admin API Key Rate-Limit Reset Cache Invalidation

**Files:**
- Modify: `backend/internal/handler/admin/apikey_handler.go`
- Modify: `backend/internal/handler/admin/apikey_handler_test.go`
- Modify: `backend/internal/handler/admin/admin_service_stub_test.go`
- Modify: `backend/internal/service/admin_service.go`
- Modify: `backend/internal/service/billing_cache_service.go`
- Modify: `backend/internal/service/wire.go`
- Modify: `backend/cmd/server/wire_gen.go`

- [ ] **Step 1: Write failing admin handler and service tests**

Add handler coverage proving admin update accepts `reset_rate_limit_usage` without changing group:

```go
func TestAdminAPIKeyHandlerUpdateGroup_ResetRateLimitUsageOnly(t *testing.T) {
	reset := true
	body := `{"reset_rate_limit_usage":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "123"}}

	stub := &adminServiceStub{apiKey: &service.APIKey{ID: 123, Key: "sk-test"}}
	h := NewAdminAPIKeyHandler(stub)
	h.UpdateGroup(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, stub.resetRateLimitCalled)
	require.Nil(t, stub.lastGroupID)
	require.True(t, reset)
}
```

Add service coverage proving the Redis rate-limit cache is invalidated when admin reset runs:

```go
func TestAdminResetAPIKeyRateLimitUsage_InvalidatesBillingCache(t *testing.T) {
	key := &APIKey{
		ID:            123,
		Key:           "sk-test",
		Usage5h:       10,
		Usage1d:       20,
		Usage7d:       30,
		Window5hStart: ptrTime(time.Now()),
		Window1dStart: ptrTime(time.Now()),
		Window7dStart: ptrTime(time.Now()),
	}
	repo := &adminAPIKeyRepoStub{key: key}
	cache := &billingCacheInvalidationStub{}
	svc := &adminServiceImpl{apiKeyRepo: repo, billingCacheService: cache}

	got, err := svc.AdminResetAPIKeyRateLimitUsage(context.Background(), 123)
	require.NoError(t, err)
	require.Equal(t, float64(0), got.Usage5h)
	require.Equal(t, float64(0), got.Usage1d)
	require.Equal(t, float64(0), got.Usage7d)
	require.Nil(t, got.Window5hStart)
	require.Nil(t, got.Window1dStart)
	require.Nil(t, got.Window7dStart)
	require.Equal(t, int64(123), cache.invalidatedKeyID)
}
```

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/handler/admin ./internal/service -run 'AdminAPIKeyHandlerUpdateGroup_ResetRateLimitUsageOnly|AdminResetAPIKeyRateLimitUsage_InvalidatesBillingCache' -count=1
```

Expected:

```text
FAIL
```

The failure should mention missing `ResetRateLimitUsage`, missing `AdminResetAPIKeyRateLimitUsage`, or missing billing cache invalidation method.

- [ ] **Step 3: Port admin reset behavior**

In `AdminUpdateAPIKeyGroupRequest`, add:

```go
ResetRateLimitUsage *bool `json:"reset_rate_limit_usage"`
```

In `AdminService`, add:

```go
AdminResetAPIKeyRateLimitUsage(ctx context.Context, keyID int64) (*APIKey, error)
```

In `adminServiceImpl`, add:

```go
func (s *adminServiceImpl) AdminResetAPIKeyRateLimitUsage(ctx context.Context, keyID int64) (*APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return nil, err
	}
	apiKey.Usage5h = 0
	apiKey.Usage1d = 0
	apiKey.Usage7d = 0
	apiKey.Window5hStart = nil
	apiKey.Window1dStart = nil
	apiKey.Window7dStart = nil
	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("reset api key rate limit usage: %w", err)
	}
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByKey(ctx, apiKey.Key)
	}
	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateAPIKeyRateLimit(ctx, apiKey.ID)
	}
	return apiKey, nil
}
```

In `BillingCacheService`, add:

```go
func (s *BillingCacheService) InvalidateAPIKeyRateLimit(ctx context.Context, keyID int64) error {
	if s.cache == nil {
		return nil
	}
	if err := s.cache.InvalidateAPIKeyRateLimit(ctx, keyID); err != nil {
		logger.LegacyPrintf("service.billing_cache", "Warning: invalidate api key rate limit cache failed for key %d: %v", keyID, err)
		return err
	}
	return nil
}
```

Wire `APIKeyService.SetRateLimitCacheInvalidator(billingCacheService)` in the existing `ProvideAPIKeyService` without dropping the current `SubscriptionProductService` wiring:

```go
svc := NewAPIKeyService(apiKeyRepo, userRepo, groupRepo, userSubRepo, userGroupRateRepo, cache, cfg)
svc.subscriptionProductService = subscriptionProductService
svc.SetRateLimitCacheInvalidator(billingCacheService)
return svc
```

- [ ] **Step 4: Regenerate Wire**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go generate ./cmd/server
```

Expected:

```text
```

No output is expected on success.

- [ ] **Step 5: Run admin and service tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/handler/admin ./internal/service -run 'APIKey|RateLimit|AdminReset' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 6: Commit API key cache reset fix**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/handler/admin/apikey_handler.go backend/internal/handler/admin/apikey_handler_test.go backend/internal/handler/admin/admin_service_stub_test.go backend/internal/service/admin_service.go backend/internal/service/billing_cache_service.go backend/internal/service/wire.go backend/cmd/server/wire_gen.go
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(api-key): reset admin rate-limit cache"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(api-key): reset admin rate-limit cache
```

---

### Task 7: Apicompat Tool Input And Tool Choice Fixes

**Files:**
- Modify: `backend/internal/pkg/apicompat/responses_to_anthropic.go`
- Modify: `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
- Modify: `backend/internal/pkg/apicompat/anthropic_to_responses.go`
- Modify: `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- Modify: `backend/internal/pkg/apicompat/anthropic_responses_test.go`
- Modify: `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`

- [ ] **Step 1: Write failing `Read.pages` tests**

Add tests to `backend/internal/pkg/apicompat/anthropic_responses_test.go`:

```go
func TestResponsesToAnthropic_ReadToolDropsEmptyPages(t *testing.T) {
	resp := &ResponsesResponse{
		ID:     "resp_read",
		Model:  "gpt-5.5",
		Status: "completed",
		Output: []ResponsesOutput{{
			Type:      "function_call",
			CallID:    "call_read",
			Name:      "Read",
			Arguments: `{"file_path":"/tmp/demo.py","limit":2000,"offset":0,"pages":""}`,
		}},
	}

	anth := ResponsesToAnthropic(resp, "claude-opus-4-6")
	require.Len(t, anth.Content, 1)
	require.JSONEq(t, `{"file_path":"/tmp/demo.py","limit":2000,"offset":0}`, string(anth.Content[0].Input))
}

func TestResponsesToAnthropic_PreservesEmptyStringsForOtherTools(t *testing.T) {
	resp := &ResponsesResponse{
		ID:     "resp_other",
		Model:  "gpt-5.5",
		Status: "completed",
		Output: []ResponsesOutput{{
			Type:      "function_call",
			CallID:    "call_other",
			Name:      "Search",
			Arguments: `{"query":""}`,
		}},
	}

	anth := ResponsesToAnthropic(resp, "claude-opus-4-6")
	require.Len(t, anth.Content, 1)
	require.JSONEq(t, `{"query":""}`, string(anth.Content[0].Input))
}
```

- [ ] **Step 2: Run apicompat tests to verify RED**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/pkg/apicompat -run 'ReadToolDropsEmptyPages|PreservesEmptyStringsForOtherTools|ToolChoice' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Add Read tool input sanitizer**

In `responses_to_anthropic.go`, add:

```go
func sanitizeAnthropicToolUseInput(name string, raw string) json.RawMessage {
	if name != "Read" || raw == "" {
		return json.RawMessage(raw)
	}

	var input map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return json.RawMessage(raw)
	}

	if pages, ok := input["pages"]; !ok || string(pages) != `""` {
		return json.RawMessage(raw)
	}

	delete(input, "pages")
	sanitized, err := json.Marshal(input)
	if err != nil {
		return json.RawMessage(raw)
	}
	return sanitized
}
```

Replace function-call output mapping:

```go
Input: sanitizeAnthropicToolUseInput(item.Name, item.Arguments),
```

- [ ] **Step 4: Port tool-choice compatibility behavior**

Inspect upstream source:

```bash
git -C /root/sub2api-src show --no-ext-diff --unified=80 04b2866f -- backend/internal/pkg/apicompat/anthropic_to_responses.go backend/internal/pkg/apicompat/chatcompletions_to_responses.go backend/internal/pkg/apicompat/responses_to_anthropic_request.go backend/internal/service/openai_codex_transform.go
```

Apply only the converter behavior from `04b2866f`; do not import unrelated OpenAI Fast/Flex code.

Behavior that must be true after the change:

```text
Anthropic tool_choice is converted to Responses-compatible function tool_choice format.
Chat Completions tool_choice is converted to Responses-compatible function tool_choice format.
Responses tool_choice is converted back to Anthropic format without dropping function name.
Codex transform uses the same Responses-compatible format.
```

- [ ] **Step 5: Run apicompat package tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/pkg/apicompat ./internal/service -run 'ResponsesToAnthropic|ToolChoice|Codex' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 6: Commit apicompat fixes**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/pkg/apicompat backend/internal/service/openai_codex_transform.go backend/internal/service/openai_codex_transform_test.go
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(apicompat): align tool inputs and choices"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(apicompat): align tool inputs and choices
```

---

### Task 8: OpenAI Passthrough, Codex, And Image Sticky Fixes

**Files:**
- Modify: `backend/internal/service/openai_codex_transform.go`
- Modify: `backend/internal/service/openai_codex_transform_test.go`
- Modify: `backend/internal/service/openai_gateway_service.go`
- Modify: `backend/internal/service/openai_gateway_service_test.go`
- Modify: `backend/internal/service/openai_passthrough_normalization_test.go`
- Modify: `backend/internal/handler/openai_images.go`
- Modify classification file from Task 2

- [ ] **Step 1: Inspect and classify the four OpenAI/Codex commits**

Run:

```bash
git -C /root/sub2api-src show --name-status --oneline --no-renames 615557ec 9fe02bba 3d4ca5e8 7452fad8
git -C /root/sub2api-src show --no-ext-diff --unified=80 615557ec 9fe02bba 3d4ca5e8 7452fad8 -- backend/internal/service/openai_codex_transform.go backend/internal/service/openai_gateway_service.go backend/internal/handler/openai_images.go
```

Expected:

```text
615557ec fix(openai): avoid implicit image sticky sessions
9fe02bba fix(openai): strip unsupported passthrough fields
3d4ca5e8 fix(openai): preserve current Codex compact payload fields
7452fad8 fix(openai): drop reasoning items from /v1/responses input on OAuth path
```

Update the classification file with one of these values for each commit: `already equivalent`, `partially present`, `applicable`, `not applicable`, or `deferred`.

- [ ] **Step 2: Apply applicable OpenAI/Codex behavior**

Use upstream tests as the source of truth:

```bash
git -C /root/sub2api-src show --no-ext-diff --unified=80 9fe02bba -- backend/internal/service/openai_codex_transform_test.go backend/internal/service/openai_passthrough_normalization_test.go
git -C /root/sub2api-src show --no-ext-diff --unified=80 3d4ca5e8 -- backend/internal/service/openai_gateway_service_test.go
git -C /root/sub2api-src show --no-ext-diff --unified=80 7452fad8 -- backend/internal/service/openai_codex_transform_test.go
git -C /root/sub2api-src show --no-ext-diff --unified=80 615557ec -- backend/internal/service/openai_gateway_service_test.go
```

Port only tests and code matching these behaviors:

```text
unsupported passthrough fields are stripped before OAuth/native upstream calls
Codex compact payload fields already accepted by xlabapi remain preserved
reasoning items are removed from OAuth Responses input where upstream rejects them
image generation requests do not create implicit sticky sessions unless xlabapi's async image job path explicitly needs a session key
```

- [ ] **Step 3: Run focused OpenAI/Codex tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service ./internal/handler -run 'Codex|Passthrough|Compact|Reasoning|Image|Sticky' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 4: Commit OpenAI/Codex fixes**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/service/openai_codex_transform.go backend/internal/service/openai_codex_transform_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_test.go backend/internal/service/openai_passthrough_normalization_test.go backend/internal/handler/openai_images.go docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(openai): align passthrough and codex normalization"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(openai): align passthrough and codex normalization
```

---

### Task 9: Gateway Stream Failover And Error Sanitization

**Files:**
- Modify: `backend/internal/service/gateway_service.go`
- Modify: `backend/internal/service/gateway_streaming_test.go`
- Modify classification file from Task 2

- [ ] **Step 1: Inspect upstream stream failover commits**

Run:

```bash
git -C /root/sub2api-src show --no-ext-diff --unified=100 63275735 4c474616 d78478e8 -- backend/internal/service/gateway_service.go backend/internal/service/gateway_streaming_test.go
```

Expected:

```text
63275735 fix(gateway): wrap Anthropic stream EOF as failover error before client output
4c474616 fix(gateway): emit Anthropic-standard SSE error events and failover body
d78478e8 fix(gateway): sanitize stream errors to avoid leaking infrastructure topology
```

- [ ] **Step 2: Port stream tests first**

Add or adapt upstream tests proving:

```text
EOF before client output returns an UpstreamFailoverError so handler can retry another account.
Anthropic stream errors emitted after output use Anthropic-standard SSE error event shape.
Infrastructure details such as proxy host, upstream URL, local network address, and raw transport topology do not appear in stream error bodies.
```

- [ ] **Step 3: Run tests to verify RED**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service -run 'Gateway.*Stream|Anthropic.*EOF|Sanitize|Failover' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 4: Port minimal stream failover behavior**

Modify only the stream forwarding paths in `gateway_service.go` so these behavior checks pass:

```text
Before client output: upstream EOF or malformed pre-output stream returns UpstreamFailoverError.
After client output: stream writes an Anthropic-compatible SSE error event and does not attempt account failover.
All emitted stream errors use sanitized client-facing messages.
Internal upstream/proxy topology remains in logs only, not response bodies.
```

Do not refactor unrelated non-streaming gateway paths.

- [ ] **Step 5: Run focused gateway streaming tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service -run 'Gateway.*Stream|Anthropic.*EOF|Sanitize|Failover' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 6: Commit gateway stream failover fixes**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/service/gateway_service.go backend/internal/service/gateway_streaming_test.go docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(gateway): sanitize stream failover errors"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(gateway): sanitize stream failover errors
```

---

### Task 10: OpenAI WebSocket Continuation And Item References

**Files:**
- Modify: `backend/internal/service/openai_ws_forwarder.go`
- Modify: `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- Modify: `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- Modify classification file from Task 2

- [ ] **Step 1: Inspect upstream WebSocket commits**

Run:

```bash
git -C /root/sub2api-src show --no-ext-diff --unified=100 28dc34b6 094e1171 -- backend/internal/service/openai_ws_forwarder.go backend/internal/service/openai_ws_forwarder_ingress_session_test.go backend/internal/service/openai_ws_forwarder_ingress_test.go
```

Expected:

```text
28dc34b6 fix(openai): avoid inferred WS continuation on explicit tool replay
094e1171 fix(openai): infer previous response for item references
```

- [ ] **Step 2: Port upstream WS tests first**

Add or adapt tests proving:

```text
Explicit tool replay does not infer an unrelated previous response.
Item references that need previous-response context infer the previous response when no explicit tool replay is present.
Existing xlabapi WebSocket passthrough and native relay modes continue to pass.
```

- [ ] **Step 3: Run tests to verify RED or already equivalent**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service -run 'WS|WebSocket|Ingress|Continuation|ItemRef|PreviousResponse' -count=1
```

Expected:

```text
FAIL
```

If every newly added upstream test already passes before code changes, update the classification file to `already equivalent` for the matching commit and skip the implementation edit for that commit.

- [ ] **Step 4: Port minimal WS continuation behavior**

Modify `openai_ws_forwarder.go` to satisfy the upstream tests while preserving current xlabapi modes:

```text
When a client sends explicit tool output replay, do not infer an unrelated previous response.
When input item references require context and no explicit replay exists, infer the previous response from session state.
Do not alter xlabapi's `openai_oauth_responses_websockets_v2_mode` behavior.
```

- [ ] **Step 5: Run focused WebSocket tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service -run 'WS|WebSocket|Ingress|Continuation|ItemRef|PreviousResponse' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 6: Commit WS continuation fixes**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add backend/internal/service/openai_ws_forwarder.go backend/internal/service/openai_ws_forwarder_ingress_session_test.go backend/internal/service/openai_ws_forwarder_ingress_test.go docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix(openai): align websocket continuation inference"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix(openai): align websocket continuation inference
```

---

### Task 11: Low-Risk Candidate Decision

**Files:**
- Modify classification file from Task 2
- Modify optional files only when the patch stays small

- [ ] **Step 1: Inspect low-risk candidate patches**

Run:

```bash
git -C /root/sub2api-src show --name-status --oneline --no-renames 73b87299 f084d30d 4b6954f9 9d801595
git -C /root/sub2api-src show --stat --oneline --no-renames 73b87299 f084d30d 4b6954f9 9d801595
```

Expected:

```text
73b87299 feat: 添加 Anthropic 缓存 TTL 注入开关
f084d30d fix: 恢复表格分页大小 localStorage 持久化
4b6954f9 feat(ops): allow retention days = 0 to wipe table on each scheduled cleanup
9d801595 test: 更新管理员设置契约字段
```

- [ ] **Step 2: Apply the inclusion rules**

Use these concrete rules:

```text
Include `f084d30d` if it changes only frontend table persistence helpers and tests pass.
Include `4b6954f9` if backend ops cleanup can be changed without changing unrelated ops UI structure.
Defer `73b87299` if it requires broader admin settings UI restructuring; include only if it is a small settings DTO/service addition plus gateway read.
Include `9d801595` only when `73b87299` changes settings contracts.
```

Update the classification file with the final decision and evidence.

- [ ] **Step 3: Run relevant tests for any included low-risk candidate**

For frontend page-size persistence:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/frontend && npm run test:run -- usePersistedPageSize Pagination
```

For ops retention:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service -run 'OpsCleanup|Retention' -count=1
```

For Anthropic cache TTL:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/service ./internal/handler/admin ./internal/server -run 'Anthropic.*Cache|Setting|Contract' -count=1
```

Expected for each command that is run:

```text
ok
```

- [ ] **Step 4: Commit low-risk candidate result**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md backend/internal/service backend/internal/handler backend/internal/server frontend/src/components/common/Pagination.vue frontend/src/composables/usePersistedPageSize.ts frontend/src/composables/useTableLoader.ts frontend/src/views/admin/ops/components/OpsSettingsDialog.vue frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "fix: document v0121 low-risk candidate decisions"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] fix: document v0121 low-risk candidate decisions
```

If only the classification file changed, the commit is still valid and records the deferral evidence.

---

### Task 12: Final Verification And Exclusion Audit

**Files:**
- All changed files

- [ ] **Step 1: Verify excluded features did not leak in**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 diff --name-only xlabapi...HEAD
```

Expected:

```text
docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md
backend/internal/pkg/httputil/body.go
backend/internal/pkg/httputil/body_test.go
...
```

The output must not include:

```text
backend/ent/schema/payment_order.go
backend/internal/handler/admin/channel_monitor_handler.go
backend/internal/service/vertex_service_account.go
frontend/src/views/user/PaymentView.vue
frontend/src/views/user/ChannelStatusView.vue
frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue
```

- [ ] **Step 2: Verify `VERSION` was not bumped**

Run:

```bash
cat /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend/cmd/server/VERSION
```

Expected:

```text
0.1.105
```

- [ ] **Step 3: Run backend focused package tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/pkg/apicompat ./internal/pkg/httputil ./internal/repository ./internal/service ./internal/handler ./internal/handler/admin -count=1
```

Expected:

```text
ok
```

- [ ] **Step 4: Run server contract tests if settings or routes changed**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/backend && go test -tags unit ./internal/server -run 'Contract|API' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Run frontend tests only when frontend changed**

Run this only if `git diff --name-only xlabapi...HEAD` includes `frontend/` files:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/frontend && npm run test:run -- usePersistedPageSize Pagination
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430/frontend && npm run typecheck
```

Expected:

```text
PASS
```

and:

```text
vue-tsc --noEmit
```

exits with code 0.

- [ ] **Step 6: Run diff hygiene checks**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 diff --check xlabapi...HEAD
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 status --short --branch
```

Expected:

```text
```

`diff --check` should print no output. `status` should show the integration branch with no uncommitted files.

- [ ] **Step 7: Create final verification commit only if verification notes changed**

If verification updates the classification file, run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 add docs/superpowers/context/2026-04-30-v0121-stability-gateway-classification.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-v0121-stability-gateway-20260430 commit -m "docs: record v0121 stability sync verification"
```

Expected:

```text
[integrate/xlabapi-upstream-v0121-stability-gateway-20260430 <sha>] docs: record v0121 stability sync verification
```

---

## Plan Self-Review

Spec coverage:

- Stability/security HTTP body fixes are covered by Task 3.
- Scheduler race and lock-release fixes are covered by Task 4.
- Sticky session scheduling compatibility is covered by Task 5.
- API key rate-limit cache reset is covered by Task 6.
- OpenAI/Anthropic/Codex/Responses compatibility fixes are covered by Tasks 7, 8, 9, and 10.
- Low-risk candidate classification is covered by Task 11.
- Exclusion and version-bump protections are covered by Task 12.

Placeholder scan:

- No task uses unresolved implementation markers or vague deferred-work wording.
- Steps that require code changes provide either concrete snippets or exact upstream commits and tests used as behavioral references.

Type consistency:

- `SchedulerCache.UnlockBucket(ctx, bucket)` is added to both interface and repository implementation.
- `BillingCacheService.InvalidateAPIKeyRateLimit(ctx, keyID)` matches the existing `RateLimitCacheInvalidator` interface used by `APIKeyService`.
- `AdminResetAPIKeyRateLimitUsage(ctx, keyID)` is added to `AdminService` and implemented on `adminServiceImpl`.
