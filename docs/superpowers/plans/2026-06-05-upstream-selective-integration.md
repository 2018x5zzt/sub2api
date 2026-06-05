# Upstream Selective Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port selected high-value upstream fixes into `xlabapi` without overwriting xlabapi-specific production customizations.

**Architecture:** Use an isolated worktree and apply upstream commits in small batches. Prefer `git cherry-pick -n` for traceable source diffs, but resolve conflicts by preserving `xlabapi` behavior and taking only the scoped fix. Keep low-risk UI/deploy/security fixes separate from gateway stability fixes so each batch can be tested and reverted independently.

**Tech Stack:** Go 1.26.2 backend, Ent, Gin, React/Vite `frontend-v2`, Vue/Vite legacy `frontend`, Docker Compose deployment files, git worktrees.

---

## File Structure

This plan is mostly upstream patch integration. Expected touched areas are:

- `backend/internal/handler/admin/account_data.go`, `backend/internal/handler/dto/*`, `backend/internal/service/*`
  - Account credential redaction from admin APIs.
- `backend/internal/handler/page_handler.go`, `backend/internal/server/router.go`, `backend/internal/service/setting_service.go`
  - Custom page JWT/visibility checks.
- `deploy/docker-compose.yml`
  - Remove production PostgreSQL/Redis host port exposure only.
- `frontend/src/...`
  - Legacy frontend fixes for redeem copy, groups counts, non-Antigravity model scopes, Ops dashboard link initialization, Settings dark-mode tabs, TOTP autocomplete, and CC-Switch Codex import metadata.
- `backend/internal/service/openai_endpoint_url.go`, OpenAI gateway/image/chat-completions files and tests
  - Versioned compatible base URL handling and image upstream context detachment.
- `backend/internal/config/*`, `backend/internal/repository/http_upstream*`, `backend/internal/service/http_upstream_profile*`, deploy env/config examples
  - HTTP/2 response header timeout configuration.

Files that must not be deleted or overwritten in this plan:

- `backend/internal/enterprisebff/**`
- `frontend-v2/**` custom xlabapi work unless a scoped test explicitly requires it
- `backend/ent/schema/**` and `backend/migrations/**` except if a listed task explicitly names the file; no listed task does.

---

### Task 1: Create integration worktree and verify baseline

**Files:**
- No source changes.

- [ ] **Step 1: Create isolated worktree**

Run from `/root/sub2api-src`:

```bash
git worktree add "/root/.config/superpowers/worktrees/sub2api-src/upstream-selective-xlabapi-20260605" -b "upstream-selective-xlabapi-20260605" xlabapi
```

Expected: new worktree on branch `upstream-selective-xlabapi-20260605` at current `xlabapi` tip.

- [ ] **Step 2: Confirm clean baseline**

Run:

```bash
git -C "/root/.config/superpowers/worktrees/sub2api-src/upstream-selective-xlabapi-20260605" status --short --branch
```

Expected:

```text
## upstream-selective-xlabapi-20260605
```

- [ ] **Step 3: Run a focused baseline test set**

Run:

```bash
cd "/root/.config/superpowers/worktrees/sub2api-src/upstream-selective-xlabapi-20260605/backend"
go test ./internal/handler/dto ./internal/service ./internal/repository
```

Expected: PASS. If it fails before any upstream patch is applied, stop and report the baseline failure.

---

### Task 2: Port admin account credential redaction

**Files:**
- Modify: `backend/internal/handler/admin/account_data.go`
- Modify/Create: `backend/internal/handler/dto/credentials_redact.go`
- Modify/Create tests under `backend/internal/handler/dto/*redact*_test.go`
- Modify/Create: `backend/internal/service/account_credentials_redact.go`
- Modify/Create tests under `backend/internal/service/*credentials*_test.go`
- Modify: `backend/internal/service/admin_service.go`

- [ ] **Step 1: Apply upstream patch without committing**

Run from the worktree root:

```bash
git cherry-pick -n 0f8e2d09
```

Expected: either clean apply or conflicts only in the files listed above. If a conflict touches `enterprisebff`, frontend-v2, migrations, or unrelated schema files, run `git cherry-pick --abort` and stop for review.

- [ ] **Step 2: Resolve conflicts preserving xlabapi fields**

If conflicts occur, keep xlabapi-specific account fields and apply only the redaction behavior:

```bash
git status --short
git diff --check
```

Expected: no unresolved conflict markers and no whitespace errors.

- [ ] **Step 3: Run targeted backend tests**

Run from `backend`:

```bash
go test ./internal/handler/dto ./internal/service -run 'Credential|Redact|AdminService'
```

Expected: PASS. This proves sensitive credential fields are redacted and admin update/merge behavior remains valid.

- [ ] **Step 4: Commit redaction port**

Run from the worktree root:

```bash
git add backend/internal/handler/admin/account_data.go backend/internal/handler/dto backend/internal/service
git commit -m "$(cat <<'EOF'
fix(security): redact admin account credentials

Port upstream credential redaction for admin account responses while preserving xlabapi account fields and enterprise behavior.
EOF
)"
```

Expected: commit succeeds.

---

### Task 3: Port custom page auth and visibility checks

**Files:**
- Modify: `backend/internal/handler/page_handler.go`
- Modify: `backend/internal/server/router.go`
- Modify: `backend/internal/service/setting_service.go`
- Modify: `frontend/src/views/user/CustomPageView.vue`

- [ ] **Step 1: Apply upstream patch without committing**

Run from the worktree root:

```bash
git cherry-pick -n cf2d5067
```

Expected: clean apply or conflicts only in listed files. If router conflicts with xlabapi custom routes, keep all xlabapi routes and add only the required auth/visibility handling.

- [ ] **Step 2: Check for conflicts and forbidden changes**

Run:

```bash
git status --short
git diff --name-status HEAD
git diff --check
```

Expected: only the files listed in this task are modified and there are no conflict markers.

- [ ] **Step 3: Run targeted page/backend tests**

Run from `backend`:

```bash
go test ./internal/handler ./internal/server ./internal/service -run 'Page|CustomPage|Visibility|Setting'
```

Expected: PASS. If no tests match in one package, the command still exits successfully for other package tests.

- [ ] **Step 4: Commit page security port**

Run from the worktree root:

```bash
git add backend/internal/handler/page_handler.go backend/internal/server/router.go backend/internal/service/setting_service.go frontend/src/views/user/CustomPageView.vue
git commit -m "$(cat <<'EOF'
fix(security): enforce custom page visibility

Port upstream JWT and visibility checks for custom pages without changing xlabapi route registration semantics.
EOF
)"
```

Expected: commit succeeds.

---

### Task 4: Port low-risk legacy frontend and deploy fixes

**Files:**
- Modify: `deploy/docker-compose.yml`
- Modify: legacy frontend files touched by these commits:
  - `frontend/src/views/admin/RedeemView.vue`
  - `frontend/src/views/admin/GroupsView.vue`
  - `frontend/src/components/payment/SubscriptionPlanCard.vue`
  - `frontend/src/components/payment/__tests__/SubscriptionPlanCard.spec.ts`
  - `frontend/src/views/admin/groupsSupportedModelScopes.ts`
  - `frontend/src/views/admin/__tests__/groupsSupportedModelScopes.spec.ts`
  - `frontend/src/views/admin/ops/OpsDashboard.vue`
  - `frontend/src/views/admin/SettingsView.vue`
  - `frontend/src/components/auth/TotpLoginModal.vue`
  - `frontend/src/utils/ccswitchImport.ts`
  - `frontend/src/utils/__tests__/ccswitchImport.spec.ts`
  - `frontend/src/views/user/KeysView.vue`

- [ ] **Step 1: Apply upstream patches without committing**

Run from the worktree root:

```bash
git cherry-pick -n 18790386 4d51e53d 360f8dec 26ca73a4 e46d2c21 b0c77233 44679221 65493df9
```

Expected: clean apply or conflicts only in `deploy/docker-compose.yml` and legacy `frontend/src/**`. Do not accept changes to `frontend-v2/**` in this task.

- [ ] **Step 2: Confirm docker-compose production ports only changed**

Run:

```bash
git diff -- deploy/docker-compose.yml
git diff --name-status HEAD | rg 'docker-compose|frontend/src|deploy/docker-compose.yml'
```

Expected: `deploy/docker-compose.yml` removes host port exposure for PostgreSQL/Redis only; local/dev compose files are unchanged.

- [ ] **Step 3: Run legacy frontend targeted tests**

Run from the worktree root:

```bash
npm --prefix frontend exec -- vitest --root frontend run \
  src/utils/__tests__/ccswitchImport.spec.ts \
  src/components/payment/__tests__/SubscriptionPlanCard.spec.ts \
  src/views/admin/__tests__/groupsSupportedModelScopes.spec.ts
```

Expected: PASS. If `npm` reports missing dependencies, run `npm --prefix frontend install --package-lock=false`, then rerun this command and ensure no lockfile was created or modified.

- [ ] **Step 4: Run legacy frontend typecheck**

Run:

```bash
npm --prefix frontend run typecheck
```

Expected: PASS. If pre-existing typecheck failures outside touched files occur, capture output and stop for review.

- [ ] **Step 5: Commit low-risk frontend/deploy ports**

Run from the worktree root:

```bash
git add deploy/docker-compose.yml frontend/src
git commit -m "$(cat <<'EOF'
fix(upstream): port low-risk legacy UI and deploy fixes

Bring selected upstream legacy frontend compatibility fixes and production compose hardening into xlabapi without changing frontend-v2 behavior.
EOF
)"
```

Expected: commit succeeds.

---

### Task 5: Port versioned OpenAI compatible base URL handling

**Files:**
- Create/Modify: `backend/internal/service/openai_endpoint_url.go`
- Modify: `backend/internal/service/openai_gateway_chat_completions.go`
- Modify: `backend/internal/service/openai_gateway_chat_completions_raw.go`
- Modify: `backend/internal/service/openai_gateway_service.go`
- Modify: `backend/internal/service/openai_images.go`
- Modify tests under `backend/internal/service/*openai*test.go`

- [ ] **Step 1: Apply upstream patch without committing**

Run from the worktree root:

```bash
git cherry-pick -n 679c0865
```

Expected: likely conflicts in OpenAI service files. Keep xlabapi custom image diagnostics, compact keepalive, and frontend-v2 recent changes; apply only the versioned URL resolution helper and call-site adjustments.

- [ ] **Step 2: Confirm no unrelated gateway refactor leaked in**

Run:

```bash
git diff --name-status HEAD
git diff -- backend/internal/service/openai_gateway_service.go backend/internal/service/openai_images.go
git diff --check
```

Expected: no `enterprisebff`, no ent schema, no migration changes; no conflict markers.

- [ ] **Step 3: Run targeted OpenAI URL tests**

Run from `backend`:

```bash
go test ./internal/service -run 'OpenAI.*Endpoint|Versioned|CompatibleBase|ChatCompletionsRaw|Images'
```

Expected: PASS.

- [ ] **Step 4: Commit versioned URL port**

Run from the worktree root:

```bash
git add backend/internal/service/openai_endpoint_url.go backend/internal/service/openai_gateway_chat_completions.go backend/internal/service/openai_gateway_chat_completions_raw.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_images.go backend/internal/service/*openai*test.go
git commit -m "$(cat <<'EOF'
fix(openai): support versioned compatible base urls

Port upstream OpenAI endpoint URL normalization while preserving xlabapi gateway and image diagnostics behavior.
EOF
)"
```

Expected: commit succeeds.

---

### Task 6: Port detached upstream context for image generation

**Files:**
- Modify: `backend/internal/service/openai_images.go`
- Modify: `backend/internal/service/openai_images_responses.go`
- Modify tests under `backend/internal/service/openai_images*_test.go`

- [ ] **Step 1: Apply upstream patch without committing**

Run from the worktree root:

```bash
git cherry-pick -n a6117429
```

Expected: small diff in image request context handling. If conflicts occur, preserve xlabapi logging and missing-output diagnostics while detaching upstream request context.

- [ ] **Step 2: Run image tests**

Run from `backend`:

```bash
go test ./internal/service -run 'OpenAI.*Image|Images|Detach|Context'
```

Expected: PASS.

- [ ] **Step 3: Commit image context port**

Run from the worktree root:

```bash
git add backend/internal/service/openai_images.go backend/internal/service/openai_images_responses.go backend/internal/service/openai_images*_test.go
git commit -m "$(cat <<'EOF'
fix(openai): detach image upstream requests from clients

Port upstream image-generation context detachment while retaining xlabapi image diagnostics and failover handling.
EOF
)"
```

Expected: commit succeeds.

---

### Task 7: Port HTTP/2 response header timeout configuration

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/internal/repository/http_upstream.go`
- Modify: `backend/internal/repository/http_upstream_test.go`
- Modify: `backend/internal/service/http_upstream_profile.go`
- Modify: `backend/internal/service/http_upstream_profile_test.go`
- Modify OpenAI call sites that use upstream profiles
- Modify deploy env/config examples for the new timeout variables.

- [ ] **Step 1: Apply upstream patch without committing**

Run from the worktree root:

```bash
git cherry-pick -n 33ac8eb2
```

Expected: conflicts possible in config/deploy files. Preserve xlabapi deploy script and custom config fields; add only the HTTP/2 response header timeout settings and transport wiring.

- [ ] **Step 2: Confirm diff scope**

Run:

```bash
git diff --name-status HEAD
git diff -- backend/internal/config/config.go backend/internal/repository/http_upstream.go backend/internal/service/http_upstream_profile.go
git diff --check
```

Expected: no schema/migration changes and no frontend-v2 changes.

- [ ] **Step 3: Run HTTP upstream tests**

Run from `backend`:

```bash
go test ./internal/config ./internal/repository ./internal/service -run 'HTTP|Upstream|Timeout|Profile|ResponseHeader'
```

Expected: PASS.

- [ ] **Step 4: Commit HTTP timeout port**

Run from the worktree root:

```bash
git add backend/internal/config backend/internal/repository/http_upstream* backend/internal/service/http_upstream_profile* backend/internal/service/openai* deploy
git commit -m "$(cat <<'EOF'
fix(upstream): add HTTP/2 response header timeout

Port upstream transport timeout configuration to improve compatibility with slow OpenAI-compatible providers.
EOF
)"
```

Expected: commit succeeds.

---

### Task 8: Final verification, merge, push, and deploy

**Files:**
- No source changes unless verification reveals issues.

- [ ] **Step 1: Verify no forbidden upstream deletions were introduced**

Run from the worktree root:

```bash
git diff --name-status xlabapi...HEAD | rg 'enterprisebff|backend/ent/schema|backend/migrations' || true
```

Expected: no output. If output appears, inspect and remove those changes unless they are explicitly from a scoped task; this plan has no scoped ent/migration task.

- [ ] **Step 2: Run backend package verification**

Run from `backend`:

```bash
go test ./internal/handler/... ./internal/service/... ./internal/repository/... ./internal/config/...
```

Expected: PASS.

- [ ] **Step 3: Run frontend-v2 verification**

Run from the worktree root:

```bash
npm --prefix frontend-v2 exec -- vitest --root frontend-v2 run src/pages/user/__tests__/ccswitch.spec.ts src/pages/user/__tests__/keyUsageConfig.spec.ts src/pages/user/__tests__/Keys.spec.tsx
npm --prefix frontend-v2 run typecheck
npm --prefix frontend-v2 run build
```

Expected: PASS; Vite chunk-size warning is acceptable.

- [ ] **Step 4: Run legacy frontend scoped verification**

Run from the worktree root:

```bash
npm --prefix frontend exec -- vitest --root frontend run \
  src/utils/__tests__/ccswitchImport.spec.ts \
  src/components/payment/__tests__/SubscriptionPlanCard.spec.ts \
  src/views/admin/__tests__/groupsSupportedModelScopes.spec.ts
```

Expected: PASS. If dependency installation is missing, use `npm --prefix frontend install --package-lock=false` and rerun; do not commit generated lockfiles.

- [ ] **Step 5: Merge worktree branch back to xlabapi**

Run:

```bash
git -C /root/sub2api-src status --short --branch
git -C /root/sub2api-src merge --no-ff upstream-selective-xlabapi-20260605
```

Expected: merge succeeds and `/root/sub2api-src` remains on `xlabapi`.

- [ ] **Step 6: Re-run key verification on merged xlabapi**

Run:

```bash
cd /root/sub2api-src/backend
go test ./internal/handler/... ./internal/service/... ./internal/repository/... ./internal/config/...
cd /root/sub2api-src
npm --prefix frontend-v2 run typecheck
npm --prefix frontend-v2 run build
```

Expected: PASS.

- [ ] **Step 7: Push and deploy**

Run:

```bash
git -C /root/sub2api-src push origin xlabapi
cd /root/sub2api-src && bash ./deploy.sh
```

Expected: push succeeds; deploy script builds frontend/backend, uploads binary, restarts `sub2api`, and prints running container status.

---

## Self-Review

- Spec coverage: The plan covers all first-batch low-risk fixes and exactly the three gateway fixes approved in the design. It explicitly excludes schema-heavy and product-policy-heavy upstream changes.
- Placeholder scan: No TBD/TODO placeholders remain. Conflict handling is explicit: preserve xlabapi scope, abort and stop if conflicts leave the allowed file set.
- Type consistency: Branch names, commit hashes, commands, and file paths are used consistently across tasks.
