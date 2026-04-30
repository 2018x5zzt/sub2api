# Xlabapi Upstream Main Rebase Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an integration branch based on `upstream/main` and replay required `xlabapi` behavior, preserving production database compatibility and adding the requested upstream features.

**Architecture:** Use `upstream/main` as the new branch root, then reconcile `xlabapi` by domain. Start with classification and migration mapping before code movement. Each implementation phase gets its own focused plan and verification gate so schema, gateway, channels, bulk edit, Vertex, WebSearch/notifications, and runtime stability can be reviewed independently.

**Tech Stack:** Go 1.26.1, Gin, Ent, Wire, Redis/go-redis, Testify, Vue 3, TypeScript, Vite, Vitest, pnpm.

---

## Scope Check

The design spec covers multiple independent subsystems. This master plan intentionally does not ask one worker to resolve all conflicts in one pass. It creates the integration branch, classification ledgers, and child plan boundaries. The child implementation plans are required before each code-heavy phase starts.

Design source:

```text
docs/superpowers/specs/2026-04-30-xlabapi-upstream-main-rebase-design.md
```

Integration branch:

```text
integrate/xlabapi-upstream-main-rebase-20260430
```

Integration worktree:

```text
/root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430
```

## File Structure

### Master planning and ledgers

- Create: `docs/superpowers/plans/2026-04-30-xlabapi-upstream-main-rebase.md`
  - this master plan
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-baseline.md`
  - source refs, branch refs, divergence counts, and conflict scan summary
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`
  - upstream migration to xlabapi migration mapping
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-local-patch-ledger.md`
  - local commit classification and replay status
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-conflict-map.md`
  - dry merge conflict surface grouped by ownership area

### Child plans to create before implementation phases

- Create: `docs/superpowers/plans/2026-05-01-upstream-main-schema-migrations-ent.md`
  - schema, migration, Ent, Wire reconciliation
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-gateway-compat.md`
  - OpenAI, Codex, Claude, Anthropic, Responses, Messages, image, Sora, WebSocket compatibility
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-channels-model-surface.md`
  - channel-first public surface and model-plaza compatibility
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-account-bulk-edit.md`
  - upstream account bulk edit chain adapted to xlabapi account fields
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-vertex.md`
  - Vertex service-account support
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-websearch-notifications.md`
  - WebSearch tri-state, Anthropic web-search emulation, balance/quota notifications
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-runtime-stability.md`
  - scheduler, compressed body, stream failover, cache reset, table persistence, ops retention checks
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-final-verification.md`
  - full integration verification and cutover checklist

### Implementation work areas by child plan

Schema and generated code:

- `backend/ent/schema/*`
- `backend/ent/*`
- `backend/migrations/*`
- `backend/internal/repository/migrations_schema_integration_test.go`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/**/wire.go`

Gateway compatibility:

- `backend/internal/pkg/apicompat/*`
- `backend/internal/pkg/openai/*`
- `backend/internal/handler/gateway_handler*.go`
- `backend/internal/handler/openai_*.go`
- `backend/internal/service/openai_*.go`
- `backend/internal/service/gateway_*.go`
- `backend/internal/service/*anthropic*.go`
- `backend/internal/service/*sora*.go`
- `backend/resources/model-pricing/*`

Channels and model-surface compatibility:

- `backend/internal/service/channel*.go`
- `backend/internal/handler/available_channel_handler.go`
- `backend/internal/handler/admin/channel_handler.go`
- `frontend/src/views/user/AvailableChannelsView.vue`
- `frontend/src/views/user/ModelHubView.vue` if restored as a compatibility wrapper
- `frontend/src/router/index.ts`
- `frontend/src/components/channels/*`
- `frontend/src/components/layout/AppSidebar.vue`

Account bulk edit:

- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_bulk_update_test.go`
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/admin/account_data.go`
- `frontend/src/components/account/BulkEditAccountModal.vue`
- `frontend/src/components/admin/account/*`
- `frontend/src/api/admin/*`
- `frontend/src/types/index.ts`

Vertex:

- `backend/internal/service/account*.go`
- `backend/internal/service/account_test_service*.go`
- `backend/internal/handler/admin/account_handler.go`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/AccountTestModal.vue`
- `frontend/src/views/user/UsageView.vue`

WebSearch and notifications:

- `backend/internal/service/account_websearch*.go`
- `backend/internal/service/balance_notify*.go`
- `backend/internal/service/email_service.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/dto/notify_email_entry.go`
- `backend/internal/handler/admin/setting_handler.go`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/types/index.ts`

Runtime stability:

- `backend/internal/pkg/httputil/body.go`
- `backend/internal/repository/scheduler_cache.go`
- `backend/internal/service/scheduler_cache.go`
- `backend/internal/service/scheduler_snapshot_service.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/billing_cache_service.go`
- `frontend/src/composables/*`
- `frontend/src/components/common/Pagination.vue`

---

### Task 1: Create The Upstream-Main Integration Worktree

**Files:**
- Create worktree: `/root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430`

- [ ] **Step 1: Confirm source branch status**

Run:

```bash
git -C /root/sub2api-src status --short --branch
```

Expected:

```text
## xlabapi...origin/xlabapi [ahead 4]
```

The ahead commits should be documentation/planning commits only.

- [ ] **Step 2: Refresh upstream**

Run:

```bash
git -C /root/sub2api-src fetch upstream
```

Expected:

```text
```

No output is acceptable. If output appears, confirm `upstream/main` still resolves.

- [ ] **Step 3: Create the worktree from upstream/main**

Run:

```bash
git -C /root/sub2api-src worktree add /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 -b integrate/xlabapi-upstream-main-rebase-20260430 upstream/main
```

Expected:

```text
Preparing worktree (new branch 'integrate/xlabapi-upstream-main-rebase-20260430')
HEAD is now at 48912014 chore: sync VERSION to 0.1.121 [skip ci]
```

If `upstream/main` has advanced, record the new commit in the baseline file in Task 2 and use that commit as the integration baseline.

- [ ] **Step 4: Verify integration branch status**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 status --short --branch
```

Expected:

```text
## integrate/xlabapi-upstream-main-rebase-20260430
```

No modified files should be listed.

- [ ] **Step 5: Commit checkpoint**

No commit is needed for worktree creation because no repository file changes are expected.

---

### Task 2: Create Baseline And Conflict Ledgers

**Files:**
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-baseline.md`
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-conflict-map.md`

- [ ] **Step 1: Capture baseline refs**

Run:

```bash
git -C /root/sub2api-src rev-parse xlabapi origin/xlabapi upstream/main
git -C /root/sub2api-src rev-list --left-right --count xlabapi...upstream/main
git -C /root/sub2api-src merge-base xlabapi upstream/main
```

Expected values from the design pass:

```text
xlabapi: 7752d2e6 or a descendant containing this plan
origin/xlabapi: cca30da8c7f7314469e58cef8d17ddcd38442684
upstream/main: 48912014a16e2dd1cfca8b7cad785d0e8e7bfeec
divergence: approximately 246 770 after this plan commit
merge-base: 6a2cf09ee05ff4833c93592f6c68cf21415febde
```

- [ ] **Step 2: Create baseline file**

Create `docs/superpowers/context/2026-04-30-upstream-main-rebase-baseline.md` in the integration worktree with this content, replacing refs if Step 1 differs:

```markdown
# Upstream Main Rebase Baseline

**Date:** 2026-04-30
**Integration branch:** integrate/xlabapi-upstream-main-rebase-20260430
**Integration root:** upstream/main

## Source Refs

| Ref | Commit | Role |
| --- | --- | --- |
| upstream/main | 48912014a16e2dd1cfca8b7cad785d0e8e7bfeec | new baseline |
| xlabapi | 7752d2e6 | local behavior source |
| origin/xlabapi | cca30da8c7f7314469e58cef8d17ddcd38442684 | published xlabapi baseline before local docs |
| merge-base | 6a2cf09ee05ff4833c93592f6c68cf21415febde | common ancestor |

## Required Outcomes

- Use upstream/main as the integration root.
- Preserve xlabapi production migration compatibility.
- Preserve OpenAI/Codex/Claude compatibility.
- Migrate model-plaza behavior toward channel-first Available Channels with compatibility routes.
- Include account bulk edit, Vertex, WebSearch/notifications, and runtime stability.

## Verification Gate

This branch cannot replace xlabapi until all child plans complete their targeted tests and final verification passes.
```

- [ ] **Step 3: Generate conflict map input**

Run:

```bash
git -C /root/sub2api-src merge-tree --write-tree --name-only xlabapi upstream/main > /tmp/xlabapi-upstream-main-merge-tree.txt
```

Expected:

```text
```

The command writes conflict details to `/tmp/xlabapi-upstream-main-merge-tree.txt`.

- [ ] **Step 4: Create conflict map file**

Create `docs/superpowers/context/2026-04-30-upstream-main-rebase-conflict-map.md` with this content:

```markdown
# Upstream Main Rebase Conflict Map

**Date:** 2026-04-30

## High-Risk Conflict Areas

| Area | Conflict type | Resolution owner |
| --- | --- | --- |
| Ent generated code | content conflicts across `backend/ent/*` | schema/migration phase; regenerate after schema decisions |
| Migrations | overlapping numbered migrations and local production history | schema/migration phase; preserve local high-water mark and add compatibility migrations |
| Account admin handlers | content conflicts and upstream bulk edit changes | account bulk edit phase |
| Channels and available channels | add/add and content conflicts | channels/model-surface phase |
| Affiliate and invite | add/add and content conflicts | schema/migration plus channel phase; preserve xlabapi affiliate-only invite retirement semantics |
| Settings DTO and public settings | content conflicts | WebSearch/notifications and channels phases |
| Gateway/OpenAI/Codex/Claude | content conflicts | gateway compatibility phase |
| Sora/image paths | upstream removals vs xlabapi modifications | gateway compatibility phase; preserve xlabapi compatibility where still used |
| Frontend account/settings/channel pages | content conflicts | child plan owned by related feature area |

## Raw Conflict Capture

The raw dry-merge output was generated with:

```bash
git -C /root/sub2api-src merge-tree --write-tree --name-only xlabapi upstream/main > /tmp/xlabapi-upstream-main-merge-tree.txt
```

The raw output is intentionally not committed because it is machine output and may change when either source ref advances.
```

- [ ] **Step 5: Commit baseline ledgers**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add docs/superpowers/context/2026-04-30-upstream-main-rebase-baseline.md docs/superpowers/context/2026-04-30-upstream-main-rebase-conflict-map.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: capture upstream main rebase baseline"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: capture upstream main rebase baseline
```

---

### Task 3: Create Migration Map Ledger

**Files:**
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`

- [ ] **Step 1: Capture migration file lists**

Run:

```bash
git -C /root/sub2api-src ls-tree -r --name-only xlabapi backend/migrations > /tmp/xlabapi-migrations.txt
git -C /root/sub2api-src ls-tree -r --name-only upstream/main backend/migrations > /tmp/upstream-main-migrations.txt
```

Expected:

```text
```

Both files should be created under `/tmp`.

- [ ] **Step 2: Create migration map file**

Create `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md` in the integration worktree with this content:

```markdown
# Upstream Main Rebase Migration Map

**Date:** 2026-04-30

## Policy

- Upstream schema is the desired target schema.
- xlabapi production migration history is preserved.
- Existing xlabapi migration files are not overwritten by upstream files.
- New compatibility migrations are added after the xlabapi high-water mark.
- Generated Ent code is regenerated after schema and SQL decisions are complete.

## Current High-Water Marks

| Source | High-water mark | Notes |
| --- | --- | --- |
| xlabapi | 137_clear_legacy_subscription_carryover.sql | includes local shared subscription products, GPT image group, and carryover cleanup |
| upstream/main | 133_affiliate_rebate_freeze.sql | includes payment, auth identity, channel monitor, affiliate, WebSearch/notification-related schema |

## Required Mapping Categories

| Category | Action |
| --- | --- |
| identical migration | record as already equivalent |
| upstream migration absent from xlabapi | port as compatibility migration after xlabapi high-water mark if schema is required |
| xlabapi migration absent from upstream | preserve and ensure target schema still includes behavior |
| overlapping semantic migration | write one compatibility migration that reaches target schema without breaking existing installs |
| generated schema drift | resolve schema files first, then regenerate Ent |

## Initial Required Local Migrations To Preserve

| xlabapi migration | Required outcome |
| --- | --- |
| 109_add_subscription_daily_carryover.sql | daily carryover fields remain available |
| 110_add_shared_subscription_products.sql | shared subscription product schema remains available |
| 134_xlabapi_invite_to_affiliate_compat.sql | affiliate compatibility remains available |
| 135_add_channel_model_pricing_platform.sql | channel pricing platform remains available |
| 136_add_subscription_gpt_image_group.sql | GPT image subscription group seed remains available |
| 137_clear_legacy_subscription_carryover.sql | legacy carryover cleanup remains safe |

## Initial Required Upstream Schema Families

| Upstream family | Required outcome |
| --- | --- |
| account stats pricing | needed for WebSearch/pricing/account stats work |
| balance notify fields | needed for notifications |
| notify email entries | needed for notification recipient toggles |
| WebSearch tri-state | needed for WebSearch migration |
| auth identity and pending auth | included by upstream baseline; preserve if required by settings/auth flow |
| payment provider/order/audit | included by upstream baseline; do not regress compile or migrations even if not product-focused |
| channel monitor/request templates | included by upstream baseline; keep compile and migrations coherent |
| affiliate rebate hardening | included by upstream baseline; reconcile with xlabapi affiliate-only invite semantics |

## Final Mapping Table

Populate this table during the schema child plan before writing SQL:

| Upstream migration | xlabapi equivalent | Final action | Verification |
| --- | --- | --- | --- |
| 101_add_balance_notify_fields.sql | none confirmed | classify in schema child plan | migration test |
| 104_migrate_notify_emails_to_struct.sql | none confirmed | classify in schema child plan | migration test |
| 105_migrate_websearch_emulation_to_tristate.sql | none confirmed | classify in schema child plan | migration test |
| 108_auth_identity_foundation_core.sql | none confirmed | classify in schema child plan | migration test |
| 125_add_channel_monitors.sql | none confirmed | classify in schema child plan | migration test |
| 130_add_user_affiliates.sql | 134_xlabapi_invite_to_affiliate_compat.sql partial | classify in schema child plan | migration test |
```

- [ ] **Step 3: Commit migration map ledger**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: map upstream main migration strategy"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: map upstream main migration strategy
```

---

### Task 4: Create Local Patch Replay Ledger

**Files:**
- Create: `docs/superpowers/context/2026-04-30-upstream-main-rebase-local-patch-ledger.md`

- [ ] **Step 1: Capture local-only commit list**

Run:

```bash
git -C /root/sub2api-src log --oneline --reverse upstream/main..xlabapi > /tmp/xlabapi-local-only-commits.txt
git -C /root/sub2api-src log --oneline --reverse --cherry-pick --left-only xlabapi...upstream/main > /tmp/xlabapi-local-only-cherry-commits.txt
```

Expected:

```text
```

The files should list local `xlabapi` commits not present in upstream.

- [ ] **Step 2: Create local patch ledger**

Create `docs/superpowers/context/2026-04-30-upstream-main-rebase-local-patch-ledger.md` with this content:

```markdown
# Upstream Main Rebase Local Patch Ledger

**Date:** 2026-04-30

## Classification Legend

- replay: required local behavior must be restored on the upstream-main branch
- superseded: upstream already includes equivalent or stronger behavior
- fuse: upstream and xlabapi both contain partial behavior; implement the union
- preserve migration only: SQL/history must remain compatible, but runtime code may be represented by upstream
- obsolete: no runtime or migration behavior should be carried forward
- documentation: keep only if it documents current integration behavior

## Required Replay Domains

| Domain | Classification | Notes |
| --- | --- | --- |
| OpenAI/Codex/Claude compatibility | replay/fuse | preserve local default instructions, reasoning relay, gpt-5.4/gpt-5.5, image/Sora, Claude/Sonnet behavior |
| shared subscription products | replay/fuse | preserve product schema, billing authorization, usage settlement, frontend views |
| subscription daily carryover | replay | preserve behavior and hide migrated legacy carryover as already patched |
| subscription pro models and image group | replay | preserve seed and group/model behavior |
| available channels and affiliate transition | fuse | channel-first upstream surface plus xlabapi compatibility decisions |
| model plaza compatibility | replay as compatibility | old route remains redirect/alias; primary product becomes Available Channels |
| dynamic group budget multiplier | replay if not upstream-equivalent | preserve local billing behavior |
| enterprise visible groups | replay if still required by product | preserve if current production depends on it |
| token refresh reused-token terminal behavior | replay if not upstream-equivalent | preserve operational safety |
| local migration checksum compatibility | preserve migration only | do not break existing production migrations |

## Required Upstream Domains

| Domain | Upstream source | Integration action |
| --- | --- | --- |
| account bulk edit | 65c27d2c..53b24bc2 | child plan |
| Vertex service accounts | 6d11f9ed, 489a4d93, 93d91e20 | child plan |
| WebSearch and notifications | upstream WebSearch/notify chain | child plan |
| runtime stability | v0.1.119..v0.1.121 stability fixes | child plan |

## Commit Classification Table

Populate this table before replaying code:

| Commit | Subject | Domain | Classification | Destination child plan |
| --- | --- | --- | --- | --- |
| d208624d | fix: add default responses instructions in openai compat | gateway | replay | gateway compatibility |
| bd0837aa | fix: enable subscription pro models and image group | subscription | replay | schema and channels |
| 7426265b | fix(openai): bump codex cli version | gateway | fuse | gateway compatibility |
| 24a570ac | fix: hide migrated legacy subscription carryover | subscription | replay | schema and channels |
| cca30da8 | fix: support sonnet 4.6 thinking model | gateway | replay | gateway compatibility |
```

- [ ] **Step 3: Commit local patch ledger**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add docs/superpowers/context/2026-04-30-upstream-main-rebase-local-patch-ledger.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: classify xlabapi replay domains"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: classify xlabapi replay domains
```

---

### Task 5: Write Schema/Migration Child Plan

**Files:**
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-schema-migrations-ent.md`

- [ ] **Step 1: Inspect schema and migration drift**

Run:

```bash
git -C /root/sub2api-src diff --name-status upstream/main..xlabapi -- backend/ent/schema backend/migrations backend/internal/repository/migrations_schema_integration_test.go > /tmp/schema-migration-drift.txt
```

Expected:

```text
```

- [ ] **Step 2: Write the schema child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-schema-migrations-ent.md` with this header and task structure:

```markdown
# Upstream Main Schema Migrations Ent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reconcile upstream/main schema with xlabapi production migrations without rewriting existing migration history.

**Architecture:** Treat upstream Ent schema as the new base, add xlabapi-only schema fields where required, write compatibility migrations after the xlabapi high-water mark, then regenerate Ent and Wire.

**Tech Stack:** Go 1.26.1, Ent, Atlas, Wire, PostgreSQL migration SQL, Testify.

---

## Required Tests

- `cd backend && go test ./migrations -count=1`
- `cd backend && go test -tags=unit ./internal/repository -run 'Migration|Schema|UsageLog|Dashboard' -count=1`
- `cd backend && go test -tags=unit ./ent/... -count=1`

## Required Outputs

- Updated migration map at `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`
- Compatibility migrations numbered after the current xlabapi high-water mark
- Regenerated Ent files
- Regenerated Wire files when provider sets change
```

Then add tasks for:

- comparing `backend/ent/schema`
- writing migration compatibility tests
- adding compatibility SQL
- regenerating Ent with `go generate ./ent`
- regenerating Wire with `go generate ./cmd/server ./internal/...` if the local provider pattern supports it
- running focused tests
- committing schema work

- [ ] **Step 3: Commit schema child plan**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add docs/superpowers/plans/2026-05-01-upstream-main-schema-migrations-ent.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: plan upstream main schema reconciliation"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: plan upstream main schema reconciliation
```

---

### Task 6: Write Gateway Compatibility Child Plan

**Files:**
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-gateway-compat.md`

- [ ] **Step 1: Inspect gateway drift**

Run:

```bash
git -C /root/sub2api-src diff --name-status upstream/main..xlabapi -- backend/internal/pkg/apicompat backend/internal/pkg/openai backend/internal/handler/gateway_handler.go backend/internal/handler/gateway_handler_chat_completions.go backend/internal/handler/gateway_handler_responses.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_images.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_codex_transform.go backend/internal/service/openai_ws_forwarder.go backend/internal/service/openai_images.go backend/resources/model-pricing > /tmp/gateway-drift.txt
```

Expected:

```text
```

- [ ] **Step 2: Write the gateway child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-gateway-compat.md` with this header:

```markdown
# Upstream Main Gateway Compatibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore xlabapi OpenAI/Codex/Claude compatibility on the upstream/main baseline while preserving upstream 0.1.121 gateway stability fixes.

**Architecture:** Add tests for each xlabapi compatibility contract first, then port the smallest code slices needed to pass those tests. Preserve upstream fixes unless tests prove a stronger local behavior is required.

**Tech Stack:** Go 1.26.1, Gin, Testify, OpenAI Responses/Chat Completions, Anthropic Messages, WebSocket forwarding.

---

## Required Tests

- `cd backend && go test -tags=unit ./internal/pkg/apicompat -count=1`
- `cd backend && go test -tags=unit ./internal/service -run 'OpenAI|Codex|Claude|Anthropic|Responses|WebSocket|Image|Sora|Reasoning|Instructions' -count=1`
- `cd backend && go test -tags=unit ./internal/handler -run 'OpenAI|Gateway|Responses|ChatCompletions|Images|Sora' -count=1`

## Required Contracts

- default responses instructions are preserved
- Codex compact payload fields are preserved
- gpt-5.4 and gpt-5.5 aliases remain supported
- Sonnet 4.6 thinking behavior remains supported
- reasoning relay safety remains supported
- upstream 0.1.121 stream failover and sanitized error behavior remains supported
- Sora/image compatibility remains supported where current xlabapi exposes it
```

Then add explicit test-first tasks for default instructions, Codex compact payload, model aliases, Claude/Sonnet, image/Sora, stream failover, and item-reference inference.

- [ ] **Step 3: Commit gateway child plan**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add docs/superpowers/plans/2026-05-01-upstream-main-gateway-compat.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: plan upstream main gateway compatibility"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: plan upstream main gateway compatibility
```

---

### Task 7: Write Channels And Model-Surface Child Plan

**Files:**
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-channels-model-surface.md`

- [ ] **Step 1: Inspect channel and frontend route drift**

Run:

```bash
git -C /root/sub2api-src diff --name-status upstream/main..xlabapi -- backend/internal/service/channel.go backend/internal/service/channel_available.go backend/internal/handler/available_channel_handler.go backend/internal/handler/admin/channel_handler.go frontend/src/router/index.ts frontend/src/views/user frontend/src/components/channels frontend/src/components/layout/AppSidebar.vue > /tmp/channels-model-surface-drift.txt
```

Expected:

```text
```

- [ ] **Step 2: Write the channels child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-channels-model-surface.md` with this header:

```markdown
# Upstream Main Channels Model Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make upstream Available Channels the primary discovery surface while keeping xlabapi model-plaza compatibility routes and local subscription/channel billing behavior.

**Architecture:** Keep channel-first backend DTOs, preserve `/models` as a compatibility route, and test that users can still search by model name inside available channels.

**Tech Stack:** Go 1.26.1, Gin, Vue 3, TypeScript, Vitest.

---

## Required Tests

- `cd backend && go test -tags=unit ./internal/service -run 'Channel|Available' -count=1`
- `cd backend && go test -tags=unit ./internal/handler -run 'AvailableChannel|Channel' -count=1`
- `cd frontend && pnpm run test:run -- AvailableChannels AppSidebar`
- `cd frontend && pnpm run typecheck`

## Required Contracts

- `/available-channels` is primary
- `/models` remains compatible
- model search still works through channel rows or sections
- subscription groups, image groups, group multipliers, and channel billing remain visible and correct
```

Then add test-first tasks for route compatibility, model search, group multiplier visibility, and backend public channel filtering.

- [ ] **Step 3: Commit channels child plan**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add docs/superpowers/plans/2026-05-01-upstream-main-channels-model-surface.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: plan channel-first model compatibility"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: plan channel-first model compatibility
```

---

### Task 8: Write Feature Child Plans For Bulk Edit, Vertex, WebSearch, Runtime, And Final Verification

**Files:**
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-account-bulk-edit.md`
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-vertex.md`
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-websearch-notifications.md`
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-runtime-stability.md`
- Create: `docs/superpowers/plans/2026-05-01-upstream-main-final-verification.md`

- [ ] **Step 1: Write account bulk edit child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-account-bulk-edit.md` with this header:

```markdown
# Upstream Main Account Bulk Edit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port upstream account bulk edit into the upstream-main xlabapi integration branch and adapt it to local account fields.

**Architecture:** Port backend filter-target update first, then handler contract, then frontend modal. Every supported field gets a focused service or component test.

**Tech Stack:** Go 1.26.1, Gin, Testify, Vue 3, TypeScript, Vitest.

---

## Required Upstream Commits

- 65c27d2c docs: add account bulk edit scope design
- 54de4e00 docs: add account bulk edit implementation plan
- 25c7b0d9 feat: support filter-target account bulk update
- 2ab6b34f feat: add filtered-result account bulk edit
- a161f9d0 feat: align OpenAI bulk edit compact settings
- 53b24bc2 fix: tighten account bulk edit target typing

## Required Tests

- `cd backend && go test -tags=unit ./internal/service -run 'Bulk|Account' -count=1`
- `cd backend && go test -tags=unit ./internal/handler/admin -run 'Bulk|Account' -count=1`
- `cd frontend && pnpm run test:run -- BulkEditAccountModal`
- `cd frontend && pnpm run typecheck`
```

- [ ] **Step 2: Write Vertex child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-vertex.md` with this header:

```markdown
# Upstream Main Vertex Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port upstream Vertex service-account support and usage-window visibility into the upstream-main xlabapi integration branch.

**Architecture:** Add schema and DTO support first, then account create/edit/test, then scheduling and usage stats, then frontend controls.

**Tech Stack:** Go 1.26.1, Gin, Testify, Vue 3, TypeScript, Vitest.

---

## Required Upstream Commits

- 6d11f9ed Add Vertex service account support
- 489a4d93 Show today stats for Vertex usage window
- 93d91e20 fix(vertex): audit fixes for Vertex Service Account feature (#1977)

## Required Tests

- `cd backend && go test -tags=unit ./internal/service -run 'Vertex|AccountTest|UsageWindow' -count=1`
- `cd backend && go test -tags=unit ./internal/handler/admin -run 'Vertex|Account' -count=1`
- `cd frontend && pnpm run test:run -- AccountTestModal CreateAccountModal EditAccountModal UsageView`
- `cd frontend && pnpm run typecheck`
```

- [ ] **Step 3: Write WebSearch/notifications child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-websearch-notifications.md` with this header:

```markdown
# Upstream Main WebSearch Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port upstream WebSearch tri-state, Anthropic API Key web-search emulation, and balance/quota notification features into the integration branch.

**Architecture:** Add schema/settings DTO support first, then service behavior, then account/settings frontend controls. Notification delivery keeps existing email service safety and timeout behavior.

**Tech Stack:** Go 1.26.1, Gin, Testify, Vue 3, TypeScript, Vitest.

---

## Required Contracts

- WebSearch supports tri-state account behavior.
- Anthropic API Key accounts can use upstream web-search emulation behavior.
- balance notification supports recipient entries and recharge URL context.
- quota notification uses remaining quota semantics.
- public settings expose only user-safe notification flags.

## Required Tests

- `cd backend && go test -tags=unit ./internal/service -run 'WebSearch|Notify|Balance|Quota|Email' -count=1`
- `cd backend && go test -tags=unit ./internal/handler -run 'Settings|PublicSettings|Notify' -count=1`
- `cd frontend && pnpm run test:run -- SettingsView CreateAccountModal EditAccountModal`
- `cd frontend && pnpm run typecheck`
```

- [ ] **Step 4: Write runtime stability child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-runtime-stability.md` with this header:

```markdown
# Upstream Main Runtime Stability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Verify and finish upstream 0.1.121 runtime stability fixes on top of the reconciled xlabapi integration branch.

**Architecture:** Treat stability fixes as behavior contracts. Add focused tests for each missing contract, then port the minimal upstream code needed to pass.

**Tech Stack:** Go 1.26.1, Gin, Redis/go-redis, Testify, Vue 3, Vitest.

---

## Required Contracts

- compressed request bodies decode safely with decompression bomb protection.
- scheduler snapshots activate atomically.
- sticky sessions do not select invalid accounts.
- API key rate-limit cache resets after admin reset.
- Anthropic stream EOF failover happens before client output is committed.
- stream errors do not leak internal topology.
- table page-size localStorage persists.

## Required Tests

- `cd backend && go test -tags=unit ./internal/pkg/httputil -count=1`
- `cd backend && go test -tags=unit ./internal/repository -run 'Scheduler|Snapshot|RateLimit' -count=1`
- `cd backend && go test -tags=unit ./internal/service -run 'Scheduler|Sticky|Stream|Failover|RateLimit' -count=1`
- `cd frontend && pnpm run test:run -- Pagination useTableLoader`
```

- [ ] **Step 5: Write final verification child plan**

Create `docs/superpowers/plans/2026-05-01-upstream-main-final-verification.md` with this header:

```markdown
# Upstream Main Final Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove the upstream-main xlabapi integration branch is ready to promote to xlabapi.

**Architecture:** Run focused tests from every child plan, then broader backend/frontend checks, then produce a rollout note with migration and behavior changes.

**Tech Stack:** Go 1.26.1, pnpm, Vite, Vitest, Docker if available.

---

## Required Verification Commands

- `cd backend && go test -tags=unit ./internal/pkg/apicompat ./internal/pkg/httputil -count=1`
- `cd backend && go test -tags=unit ./internal/service -count=1`
- `cd backend && go test -tags=unit ./internal/handler ./internal/handler/admin -count=1`
- `cd backend && go test ./migrations -count=1`
- `cd frontend && pnpm run typecheck`
- `cd frontend && pnpm run test:run`
- `make build`
```

- [ ] **Step 6: Commit feature child plans**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add docs/superpowers/plans/2026-05-01-upstream-main-account-bulk-edit.md docs/superpowers/plans/2026-05-01-upstream-main-vertex.md docs/superpowers/plans/2026-05-01-upstream-main-websearch-notifications.md docs/superpowers/plans/2026-05-01-upstream-main-runtime-stability.md docs/superpowers/plans/2026-05-01-upstream-main-final-verification.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: plan upstream main feature phases"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: plan upstream main feature phases
```

---

### Task 9: Master Plan Verification

**Files:**
- Read: all context and child plan files created by this plan

- [ ] **Step 1: Confirm no child plan is missing**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 ls-files docs/superpowers/plans docs/superpowers/context | sort
```

Expected includes:

```text
docs/superpowers/context/2026-04-30-upstream-main-rebase-baseline.md
docs/superpowers/context/2026-04-30-upstream-main-rebase-conflict-map.md
docs/superpowers/context/2026-04-30-upstream-main-rebase-local-patch-ledger.md
docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md
docs/superpowers/plans/2026-05-01-upstream-main-account-bulk-edit.md
docs/superpowers/plans/2026-05-01-upstream-main-channels-model-surface.md
docs/superpowers/plans/2026-05-01-upstream-main-final-verification.md
docs/superpowers/plans/2026-05-01-upstream-main-gateway-compat.md
docs/superpowers/plans/2026-05-01-upstream-main-runtime-stability.md
docs/superpowers/plans/2026-05-01-upstream-main-schema-migrations-ent.md
docs/superpowers/plans/2026-05-01-upstream-main-vertex.md
docs/superpowers/plans/2026-05-01-upstream-main-websearch-notifications.md
```

- [ ] **Step 2: Check docs for unresolved planning markers**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 grep -nE 'T[B]D|TO[D]O|待[定]' -- docs/superpowers/plans docs/superpowers/context
```

Expected:

```text
```

No output should be produced.

- [ ] **Step 3: Confirm branch cleanliness**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 status --short --branch
```

Expected:

```text
## integrate/xlabapi-upstream-main-rebase-20260430
```

No modified files should be listed.

- [ ] **Step 4: Report execution readiness**

Report the following:

```text
Master plan complete. Integration branch is based on upstream/main. Classification ledgers and child plans are present. Ready to execute schema/migration child plan first.
```

No code implementation begins until the schema/migration child plan is selected for execution.
