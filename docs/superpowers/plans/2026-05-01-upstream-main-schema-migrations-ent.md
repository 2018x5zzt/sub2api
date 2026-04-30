# Upstream Main Schema Migrations Ent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reconcile `upstream/main` schema with `xlabapi` production migrations without rewriting existing migration history.

**Architecture:** Treat upstream Ent schema as the new base, add `xlabapi`-only schema fields where required, write compatibility migrations after the `xlabapi` high-water mark, then regenerate Ent and Wire. Use migration tests and schema drift ledgers before changing generated code.

**Tech Stack:** Go 1.26.1, Ent, Atlas, Wire, PostgreSQL migration SQL, Testify.

---

## Scope Boundary

This plan owns only schema, migrations, generated Ent, and provider wiring required by schema changes. It does not implement gateway compatibility, channel UI, account bulk edit behavior, Vertex request handling, or WebSearch/notification runtime behavior. Those phases consume the schema decisions made here.

## Source Context

- Master spec: `docs/superpowers/specs/2026-04-30-xlabapi-upstream-main-rebase-design.md`
- Master plan: `docs/superpowers/plans/2026-04-30-xlabapi-upstream-main-rebase.md`
- Migration map: `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`
- Baseline: `docs/superpowers/context/2026-04-30-upstream-main-rebase-baseline.md`

## File Map

### Planning and ledgers

- Modify: `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`
- Create: `docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md`

### Schema and generated code

- Modify: `backend/ent/schema/api_key.go`
- Modify: `backend/ent/schema/group.go`
- Preserve or add as needed: `backend/ent/schema/auth_identity*.go`
- Preserve or add as needed: `backend/ent/schema/channel_monitor*.go`
- Preserve or add as needed: `backend/ent/schema/payment_*.go`
- Preserve or add as needed: `backend/ent/schema/pending_auth_session.go`
- Preserve or add as needed: `backend/ent/schema/subscription_plan.go`
- Add or port as needed: `backend/ent/schema/subscription_product*.go`
- Add or port as needed: `backend/ent/schema/user_product_subscription.go`
- Modify: `backend/ent/schema/usage_log.go`
- Modify: `backend/ent/schema/user.go`
- Modify: `backend/ent/schema/user_subscription.go`
- Regenerate: `backend/ent/*`
- Regenerate as needed: `backend/cmd/server/wire_gen.go`

### Migrations

- Preserve upstream migrations already present in the branch through `133_affiliate_rebate_freeze.sql`.
- Add compatibility migrations after the `xlabapi` high-water mark. Start with the next safe number after `137_clear_legacy_subscription_carryover.sql`.
- Do not overwrite upstream migration files in place unless the change is a verified correction that upstream-main branch already requires.
- Modify: `backend/migrations/migrations.go` if embedded migration registration changes require it.
- Add tests: `backend/migrations/*_test.go` or `backend/internal/repository/migrations_schema_integration_test.go`

## Required Tests

- `cd backend && go test ./migrations -count=1`
- `cd backend && go test -tags=unit ./internal/repository -run 'Migration|Schema|UsageLog|Dashboard' -count=1`
- `cd backend && go test -tags=unit ./ent/... -count=1`
- `cd backend && go test -tags=unit ./internal/service -run 'SubscriptionProduct|DailyCarryover|Affiliate|Channel' -count=1`

## Required Outputs

- Updated migration map at `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`
- New drift ledger at `docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md`
- Compatibility migrations numbered after the current `xlabapi` high-water mark
- Regenerated Ent files
- Regenerated Wire files when provider sets change

---

### Task 1: Create Schema And Migration Drift Ledger

**Files:**
- Create: `docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md`
- Modify: `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`

- [ ] **Step 1: Capture schema drift from the source branch**

Run:

```bash
git -C /root/sub2api-src diff --name-status upstream/main..xlabapi -- backend/ent/schema backend/migrations backend/internal/repository/migrations_schema_integration_test.go > /tmp/schema-migration-drift.txt
```

Expected:

```text
```

The file should contain schema and migration drift entries. A non-empty file is expected.

- [ ] **Step 2: Create the drift ledger**

Create `docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md` with this content:

```markdown
# Schema Migration Drift Ledger

**Date:** 2026-05-01
**Branch:** integrate/xlabapi-upstream-main-rebase-20260430

## Classification Legend

- upstream baseline: keep upstream-main schema or migration as-is
- xlabapi replay: add xlabapi-only schema or SQL behavior
- fused: combine upstream and xlabapi behavior into one target schema
- migration preserved: keep production SQL history but do not duplicate runtime behavior
- generated only: regenerate from schema; do not hand-edit

## High-Risk Schema Families

| Family | Initial classification | Required decision |
| --- | --- | --- |
| auth identity and pending auth | upstream baseline | keep upstream schema unless xlabapi auth tests require local fields |
| payment order/provider/audit | upstream baseline | keep compile and migrations coherent even if payment product is not the focus |
| channel monitor/request templates | upstream baseline | keep upstream schema; defer runtime feature behavior to channel plan |
| subscription products | xlabapi replay | preserve shared subscription products and user product subscriptions |
| daily carryover | xlabapi replay | preserve user subscription carryover fields and cleanup behavior |
| affiliate/invite compatibility | fused | keep upstream affiliate hardening and xlabapi affiliate-only invite migration behavior |
| channel pricing platform | fused | keep upstream channel features and xlabapi pricing platform behavior |
| usage log billing and metadata | fused | preserve upstream fields plus xlabapi billing, product, cache, endpoint, and model metadata |

## Initial File Drift

Populate from `/tmp/schema-migration-drift.txt` before changing schema.
```

- [ ] **Step 3: Update the migration map with the child-plan branch note**

Append this section to `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`:

```markdown
## Schema Child Plan Notes

- The schema child plan owns final migration numbering after `137_clear_legacy_subscription_carryover.sql`.
- New SQL must be idempotent where it can be run on databases that already received xlabapi local migrations.
- `docs/superpowers` is ignored by upstream `.gitignore`; stage context updates with `git add -f`.
```

- [ ] **Step 4: Commit the drift ledger**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add -f docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "docs: classify schema migration drift"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] docs: classify schema migration drift
```

---

### Task 2: Write Failing Migration Compatibility Tests

**Files:**
- Modify or create: `backend/migrations/xlabapi_upstream_main_compat_test.go`
- Modify as needed: `backend/internal/repository/migrations_schema_integration_test.go`

- [ ] **Step 1: Add tests for required local migration outcomes**

Create `backend/migrations/xlabapi_upstream_main_compat_test.go` with tests that assert the migrated schema contains the required local tables and columns after all migrations run. The test must cover at least:

```go
func TestXlabapiCompatibilitySchemaContainsSharedSubscriptionProducts(t *testing.T)
func TestXlabapiCompatibilitySchemaContainsDailyCarryoverFields(t *testing.T)
func TestXlabapiCompatibilitySchemaContainsChannelPricingPlatform(t *testing.T)
func TestXlabapiCompatibilitySchemaContainsAffiliateCompatibilityFields(t *testing.T)
```

Use the existing migration test helpers in the repository. If no helper exists for checking table/column existence, add small local helpers in the test file:

```go
func requireTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`, table).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "expected table %s to exist", table)
}

func requireColumn(t *testing.T, db *sql.DB, table, column string) {
	t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		)
	`, table, column).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "expected column %s.%s to exist", table, column)
}
```

- [ ] **Step 2: Run migration tests and confirm RED**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test ./migrations -run 'TestXlabapiCompatibilitySchemaContains' -count=1
```

Expected:

```text
FAIL
```

The failure should be missing local xlabapi schema such as shared subscription products, daily carryover, channel pricing platform, or affiliate compatibility fields.

- [ ] **Step 3: Commit failing tests**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add backend/migrations/xlabapi_upstream_main_compat_test.go backend/internal/repository/migrations_schema_integration_test.go
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "test: cover xlabapi schema compatibility"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] test: cover xlabapi schema compatibility
```

---

### Task 3: Add Compatibility Migrations After The Xlabapi High-Water Mark

**Files:**
- Create: `backend/migrations/138_xlabapi_upstream_main_subscription_compat.sql`
- Create: `backend/migrations/139_xlabapi_upstream_main_channel_affiliate_compat.sql`
- Modify: `backend/migrations/migrations.go` if migrations are manually registered
- Modify: `docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md`

- [ ] **Step 1: Write subscription compatibility SQL**

Create `backend/migrations/138_xlabapi_upstream_main_subscription_compat.sql`.

Required outcomes:

- preserve `subscription_products`
- preserve `subscription_product_groups`
- preserve `user_product_subscriptions`
- preserve `product_subscription_migration_sources` if the xlabapi schema still requires it
- preserve user subscription daily carryover fields
- use `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`, and guarded indexes where practical

- [ ] **Step 2: Write channel and affiliate compatibility SQL**

Create `backend/migrations/139_xlabapi_upstream_main_channel_affiliate_compat.sql`.

Required outcomes:

- preserve `channels.model_pricing` platform compatibility if the target schema still stores pricing JSON that needs platform metadata
- preserve affiliate compatibility fields required by `134_xlabapi_invite_to_affiliate_compat.sql`
- preserve seeded GPT image group behavior only if seeding belongs in schema migration after upstream reconciliation; otherwise document the seed in the channels child plan

- [ ] **Step 3: Update the migration map final table**

For each touched migration family, replace `classify in schema child plan` with one of:

- `upstream baseline`
- `covered by 138_xlabapi_upstream_main_subscription_compat.sql`
- `covered by 139_xlabapi_upstream_main_channel_affiliate_compat.sql`
- `deferred to runtime child plan`

- [ ] **Step 4: Run focused migration tests and confirm GREEN**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test ./migrations -run 'TestXlabapiCompatibilitySchemaContains' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit compatibility migrations**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add backend/migrations/138_xlabapi_upstream_main_subscription_compat.sql backend/migrations/139_xlabapi_upstream_main_channel_affiliate_compat.sql backend/migrations/migrations.go docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add -f docs/superpowers/context/2026-04-30-upstream-main-rebase-migration-map.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "feat(migrations): preserve xlabapi schema compatibility"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] feat(migrations): preserve xlabapi schema compatibility
```

---

### Task 4: Reconcile Ent Schema Before Regeneration

**Files:**
- Modify relevant files under `backend/ent/schema/`
- Modify: `docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md`

- [ ] **Step 1: Add schema tests for xlabapi-only Ent objects**

Add or update Ent schema tests so generated schema must include:

- shared subscription product entities if runtime code still depends on them
- daily carryover fields on user subscription
- usage log product and billing metadata fields
- channel pricing platform behavior if modeled outside JSON

- [ ] **Step 2: Run schema tests and confirm RED**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test -tags=unit ./ent/schema -count=1
```

Expected:

```text
FAIL
```

The failure should identify missing schema objects or fields.

- [ ] **Step 3: Reconcile schema files**

Modify `backend/ent/schema/*` so target schema is the union required by:

- upstream-main entities for auth identity, payment, channel monitor, subscription plan, pending auth, and affiliate hardening
- xlabapi entities and fields for shared subscription products, carryover, product usage settlement, local billing metadata, and channel pricing platform behavior

Do not delete upstream schema files just because `xlabapi` lacks them. Do not add xlabapi schema files if the runtime behavior is already represented by upstream schema and the migration map marks them as superseded.

- [ ] **Step 4: Run schema tests and confirm GREEN**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test -tags=unit ./ent/schema -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit schema reconciliation**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add backend/ent/schema docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add -f docs/superpowers/context/2026-05-01-schema-migration-drift-ledger.md
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "feat(ent): reconcile xlabapi schema on upstream main"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] feat(ent): reconcile xlabapi schema on upstream main
```

---

### Task 5: Regenerate Ent And Wire

**Files:**
- Regenerate: `backend/ent/*`
- Regenerate as needed: `backend/cmd/server/wire_gen.go`
- Regenerate as needed: other `wire_gen.go` files if provider sets changed

- [ ] **Step 1: Regenerate Ent**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go generate ./ent
```

Expected:

```text
```

No output is acceptable. If dependencies need download and sandbox blocks it, rerun with approved network escalation.

- [ ] **Step 2: Regenerate Wire when provider sets changed**

If any `wire.go` provider set changed in this phase, run the repository's Wire generation command. Start with:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go generate ./cmd/server
```

Expected:

```text
```

If the command is not wired for all provider sets, inspect `//go:generate` declarations and run the exact package generation commands that already exist in the repository.

- [ ] **Step 3: Check generated diff scope**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 diff --name-status -- backend/ent backend/cmd/server/wire_gen.go
```

Expected:

```text
```

The output should include generated Ent files that correspond to schema changes. Unexpected unrelated files must be inspected before committing.

- [ ] **Step 4: Commit generated code**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 add backend/ent backend/cmd/server/wire_gen.go
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 commit -m "chore(ent): regenerate schema artifacts"
```

Expected:

```text
[integrate/xlabapi-upstream-main-rebase-20260430 <hash>] chore(ent): regenerate schema artifacts
```

---

### Task 6: Run Schema Phase Verification

**Files:**
- Read-only verification across changed schema, migrations, and generated code

- [ ] **Step 1: Run migration tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test ./migrations -count=1
```

Expected:

```text
ok
```

- [ ] **Step 2: Run repository schema tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test -tags=unit ./internal/repository -run 'Migration|Schema|UsageLog|Dashboard' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 3: Run Ent package tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test -tags=unit ./ent/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 4: Run service smoke tests for schema consumers**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430/backend && go test -tags=unit ./internal/service -run 'SubscriptionProduct|DailyCarryover|Affiliate|Channel' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Confirm clean worktree**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430 status --short --branch
```

Expected:

```text
## integrate/xlabapi-upstream-main-rebase-20260430...upstream/main [ahead N]
```

No modified files should be listed.

- [ ] **Step 6: Report schema phase readiness**

Report:

```text
Schema/migration/Ent phase complete. Migration map updated, compatibility migrations added after xlabapi high-water mark, Ent/Wire regenerated, and required schema verification passed.
```

If any verification command fails, report the exact failing command and keep the phase open.
