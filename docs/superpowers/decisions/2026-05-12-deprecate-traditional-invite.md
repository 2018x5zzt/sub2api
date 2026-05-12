# Decision Draft: Deprecate Traditional Invite

> **Status: superseded by [`2026-05-12-deprecate-traditional-invite-v2.md`](./2026-05-12-deprecate-traditional-invite-v2.md) on 2026-05-12. Do not implement; see v2.**

- Date: 2026-05-12
- Status: Superseded by v2 on 2026-05-12. Do not implement; see v2.
- Decision intent: deprecate the traditional invite-growth system backed by `users.invite_code`; keep the affiliate rebate system backed by `user_affiliates.aff_code`.
- Scope rule: this document is planning only. No code, migration, UI, or config changes are authorized by this draft.
- Primary references: `docs/superpowers/audits/2026-05-12-semantic-dual-tracks.md`; `origin/test/xlabapi` backend/frontend grep results.

## 1. Deprecation Scope

### To Deprecate: Traditional Invite

| Layer | Item | File:line evidence | Notes |
|---|---|---|---|
| DB field | `users.invite_code` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:1`; `origin/test/xlabapi:backend/ent/schema/user.go:57` | User-owned traditional invite relation code. Planned deprecated source of truth. |
| DB field | `users.invited_by_user_id` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:3`; `origin/test/xlabapi:backend/ent/schema/user.go:62` | Traditional inviter relation. Needs historical compatibility decision before drop/freeze. |
| DB field | `users.invite_bound_at` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:4`; `origin/test/xlabapi:backend/ent/schema/user.go:65` | Traditional invite binding timestamp. |
| DB table | `invite_code_aliases` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:10`; `origin/test/xlabapi:backend/internal/repository/user_repo.go:886` | Legacy alias lookup for migrated invite codes. |
| DB table | `invite_relationship_events` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:97` | Audit table for traditional invite relationship changes. Keep as history or migrate to archive. |
| DB table | `invite_reward_records` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:138`; `origin/test/xlabapi:backend/internal/service/invite.go:30` | Traditional base/manual/recompute reward ledger. Do not merge with affiliate ledger without explicit migration decision. |
| DB table | `invite_admin_actions` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:86`; `origin/test/xlabapi:backend/internal/service/invite.go:59` | Traditional invite admin action audit. |
| Ent/generated | `ent.User.InviteCode` and generated mutation/predicate helpers | `origin/test/xlabapi:backend/ent/user.go:38`; `origin/test/xlabapi:backend/ent/user/user.go:36`; `origin/test/xlabapi:backend/ent/mutation.go:41846` | Generated code will disappear only if schema field is removed and ent is regenerated. |
| Repository | `GetByInviteCode`, `ExistsByInviteCode`, alias lookup | `origin/test/xlabapi:backend/internal/repository/user_repo.go:170`; `origin/test/xlabapi:backend/internal/repository/user_repo.go:795`; `origin/test/xlabapi:backend/internal/repository/user_repo.go:881` | Read path for traditional codes and alias compatibility. |
| Service | `InviteService` | `origin/test/xlabapi:backend/internal/service/invite_service.go:35` | Main traditional invite service. Top-level deprecation banner should land here in S4. |
| Service method | `GenerateUniqueInviteCode` | `origin/test/xlabapi:backend/internal/service/invite_service.go:119` | Writes/generates traditional 8-letter invite code. |
| Service method | `ResolveInviterByCode` | `origin/test/xlabapi:backend/internal/service/invite_service.go:136` | Traditional code to inviter lookup. |
| Service method | `GetSummary`, `ListRewards` | `origin/test/xlabapi:backend/internal/service/invite_service.go:175`; `origin/test/xlabapi:backend/internal/service/invite_service.go:199` | User-facing traditional invite summary/reward read path. |
| Service method | `ApplyBaseRechargeRewards` | `origin/test/xlabapi:backend/internal/service/invite_service.go:210` | Traditional 3% base reward path for commercial balance redeem codes. |
| Admin service | `RebindInviter`, `CreateManualInviteGrant`, recompute | `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:52`; `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:114`; `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:168`; `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:228` | Admin-only traditional invite management. |
| User API | `GET /api/v1/invite/summary`, `GET /api/v1/invite/rewards` | `origin/test/xlabapi:backend/internal/server/routes/user.go:106`; `origin/test/xlabapi:backend/internal/handler/invite_handler.go:19`; `origin/test/xlabapi:backend/internal/handler/invite_handler.go:35` | User-facing traditional invite endpoints. |
| Admin API | `/api/v1/admin/invites/*` | `origin/test/xlabapi:backend/internal/server/routes/admin.go:619`; `origin/test/xlabapi:backend/internal/handler/admin/invite_handler.go:61`; `origin/test/xlabapi:backend/internal/handler/admin/invite_handler.go:115` | Traditional invite stats/relationships/rewards/actions/rebind/manual-grant/recompute. |
| DTO | `InviteSummary`, `InviteRewardRecord`, admin invite DTOs | `origin/test/xlabapi:backend/internal/handler/dto/invite.go:9`; `origin/test/xlabapi:backend/internal/handler/dto/invite.go:17`; `origin/test/xlabapi:backend/internal/handler/dto/admin_invite.go:12`; `origin/test/xlabapi:backend/internal/handler/dto/admin_invite.go:20` | DTOs for deprecated endpoints. |
| Wiring | `ProvideInviteService`, `NewInviteHandler` | `origin/test/xlabapi:backend/internal/service/wire.go:32`; `origin/test/xlabapi:backend/cmd/server/wire_gen.go:96` | Injector/wire entrypoints to seal in S4. |
| Frontend route | `/invite` redirect | `origin/test/xlabapi:frontend-v2/src/router/index.tsx:86`; `origin/test/xlabapi:frontend/src/router/index.ts:222` | Already redirects to `/affiliate`; keep as compatibility redirect until user approves removal. |
| Admin frontend route | `/admin/invites` redirect | `origin/test/xlabapi:frontend-v2/src/router/index.tsx:121`; `origin/test/xlabapi:frontend/src/router/index.ts:536` | Already redirects away; should remain sealed or be removed later. |
| UI/i18n copy | Traditional-looking invite labels under affiliate and registration contexts | `origin/test/xlabapi:frontend/src/i18n/locales/zh.ts:1002`; `origin/test/xlabapi:frontend/src/i18n/locales/zh.ts:1039`; `origin/test/xlabapi:frontend-v2/src/i18n/locales/zh.ts:5301`; `origin/test/xlabapi:frontend-v2/src/i18n/locales/zh.ts:5302` | Must be renamed carefully: affiliate can say “推广码/返利码”; registration invitation remains separate. |

### Not Deprecated: Keep Affiliate

| Layer | Item | File:line evidence | Notes |
|---|---|---|---|
| DB table | `user_affiliates` | `origin/test/xlabapi:backend/migrations/130_add_user_affiliates.sql:1` | Kept source of truth for invite rebate. |
| DB field | `user_affiliates.aff_code` | `origin/test/xlabapi:backend/migrations/130_add_user_affiliates.sql:3`; `origin/test/xlabapi:backend/migrations/130_add_user_affiliates.sql:16` | Kept affiliate code. UI should avoid bare “邀请码” if it causes confusion. |
| Service | `AffiliateService` and repository port | `origin/test/xlabapi:backend/internal/service/affiliate_service.go:60`; `origin/test/xlabapi:backend/internal/service/affiliate_service.go:98` | Kept. Handles profile, aff code lookup, binding, rebate accrual, transfer, admin settings. |
| Auth signup binding | `affiliateCode` binding on email/OAuth registration | `origin/test/xlabapi:backend/internal/service/auth_service.go:137`; `origin/test/xlabapi:backend/internal/service/auth_service.go:237`; `origin/test/xlabapi:backend/internal/service/auth_service.go:570`; `origin/test/xlabapi:backend/internal/service/auth_service.go:789` | Keep this path. It is not `users.invite_code`. |
| User API | `GET /api/v1/user/aff`, `POST /api/v1/user/aff/transfer` | `origin/test/xlabapi:backend/internal/handler/user_handler.go:171`; `origin/test/xlabapi:backend/internal/handler/user_handler.go:188`; `origin/test/xlabapi:backend/internal/server/routes/user.go:28` | Kept. |
| Admin API | `/api/v1/admin/affiliates/*` | `origin/test/xlabapi:backend/internal/server/routes/admin.go:599`; `origin/test/xlabapi:backend/internal/handler/admin/affiliate_handler.go:15`; `origin/test/xlabapi:backend/internal/handler/admin/affiliate_handler.go:49` | Kept. |
| Frontend route | `/affiliate`, `/admin/affiliates/*` | `origin/test/xlabapi:frontend-v2/src/router/index.tsx:87`; `origin/test/xlabapi:frontend-v2/src/router/index.tsx:122`; `origin/test/xlabapi:frontend/src/router/index.ts:226`; `origin/test/xlabapi:frontend/src/router/index.ts:576` | Kept. |
| UI copy | Affiliate user/admin pages | `origin/test/xlabapi:frontend-v2/src/pages/user/Affiliate.tsx:55`; `origin/test/xlabapi:frontend-v2/src/pages/admin/AffiliateRecords.tsx:102`; `origin/test/xlabapi:frontend/src/views/user/AffiliateView.vue:164`; `origin/test/xlabapi:frontend/src/views/admin/affiliates/AdminAffiliateRecordsTable.vue:256` | Keep behavior; rename labels if needed to “affiliate code / 返利码”. |

### Boundary Cases

| Case | Decision Draft |
|---|---|
| User has `users.invite_code` and `user_affiliates.aff_code` | Treat `aff_code` as the only future shareable code. Keep `invite_code` only for historical audit/compatibility until chosen DB option runs. |
| User was registered via traditional `invited_by_user_id`, but also has affiliate inviter | Do not auto-merge in S0-S3. In S4, choose explicit migration policy: preserve history only, or backfill affiliate inviter where affiliate relation is empty. Never double-pay rewards. |
| Historical `invite_reward_records` and affiliate ledger both exist for same recharge | Keep ledgers separate. If backfilling to affiliate later, mark migrated rows and exclude already rewarded transactions. |
| Existing `/invite?invite=...` marketing links | During transition, keep redirect to `/affiliate` or `/register` with explicit affiliate-code mapping only if a safe mapping exists. If no mapping exists, show clear “legacy invite retired” state rather than silently accepting as affiliate. |
| Registration invitation code `redeem_codes.type=invitation` | Not in scope. It remains a registration gate. Its UI text should be renamed to “registration invitation / 注册准入码” if ambiguity persists. |

## 2. Database Compatibility Plan

### Option X1: Drop `users.invite_code` Directly

**Summary**: remove traditional invite fields/tables from live schema after code paths are sealed and removed.

**Migration skeleton (pseudo SQL)**

```sql
-- S4 only, after code no longer references traditional invite fields.
BEGIN;

ALTER TABLE users DROP COLUMN IF EXISTS invite_code;
ALTER TABLE users DROP COLUMN IF EXISTS invited_by_user_id;
ALTER TABLE users DROP COLUMN IF EXISTS invite_bound_at;

DROP TABLE IF EXISTS invite_code_aliases;
DROP TABLE IF EXISTS invite_relationship_events;
DROP TABLE IF EXISTS invite_reward_records;
DROP TABLE IF EXISTS invite_admin_actions;

-- ent schema: remove fields and indexes from backend/ent/schema/user.go, then regenerate ent.
COMMIT;
```

**Rollback skeleton**

```sql
BEGIN;

ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_code VARCHAR(32);
ALTER TABLE users ADD COLUMN IF NOT EXISTS invited_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_bound_at TIMESTAMPTZ;

-- Tables can be recreated from migration 139, but dropped historical rows are gone unless restored from backup.
-- Restore table data from pre-migration backup before re-enabling old code.
COMMIT;
```

**Risk / migration cost / rollback difficulty**

- Risk: highest. Directly destroys historical relationship/reward data unless archived first.
- Migration cost: medium for schema, high for code because ent/generated code and every service/API reference must be removed first.
- Rollback difficulty: high. Schema can be recreated, but data rollback requires backup restore.
- Stop-the-world needed: not necessarily for the DDL if DB supports online operations, but operationally should be a maintenance window because code and schema must switch atomically.
- Visible user impact: old `/invite` links and admin invite history stop working unless replaced by explicit archive/read-only screens.

### Option X2: Keep Fields, Freeze Writes, Add Deprecated Comments

**Summary**: retain traditional invite schema for compatibility/history, block new writes and mark every layer as deprecated.

**Migration skeleton (pseudo SQL)**

```sql
BEGIN;

COMMENT ON COLUMN users.invite_code IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite code. Do not write or re-enable. See docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md';
COMMENT ON COLUMN users.invited_by_user_id IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite relation. Read only for historical audit.';
COMMENT ON COLUMN users.invite_bound_at IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite bind timestamp.';
COMMENT ON TABLE invite_code_aliases IS 'DEPRECATED 2026-05-12: legacy invite aliases retained for audit only.';
COMMENT ON TABLE invite_relationship_events IS 'DEPRECATED 2026-05-12: legacy traditional invite relationship audit, read only.';
COMMENT ON TABLE invite_reward_records IS 'DEPRECATED 2026-05-12: legacy traditional invite rewards, read only.';
COMMENT ON TABLE invite_admin_actions IS 'DEPRECATED 2026-05-12: legacy traditional invite admin audit, read only.';

-- Optional enforcement after code paths are removed:
-- CREATE TRIGGER reject_legacy_invite_writes BEFORE INSERT OR UPDATE OF invite_code, invited_by_user_id, invite_bound_at ON users ...

COMMIT;
```

**Rollback skeleton**

```sql
BEGIN;

COMMENT ON COLUMN users.invite_code IS NULL;
COMMENT ON COLUMN users.invited_by_user_id IS NULL;
COMMENT ON COLUMN users.invite_bound_at IS NULL;
COMMENT ON TABLE invite_code_aliases IS NULL;
COMMENT ON TABLE invite_relationship_events IS NULL;
COMMENT ON TABLE invite_reward_records IS NULL;
COMMENT ON TABLE invite_admin_actions IS NULL;

-- Drop reject-write trigger if installed.
-- Re-enable code paths only after explicit user approval.
COMMIT;
```

**Risk / migration cost / rollback difficulty**

- Risk: lowest. Historical data remains in place and old reads can still be inspected during transition.
- Migration cost: low to medium. Requires code-level freeze and CI checks, not immediate destructive DDL.
- Rollback difficulty: low. Removing comments/triggers and reverting deprecation code can restore behavior if needed.
- Stop-the-world needed: no, if implemented as comments plus application-level write freeze. Trigger rollout should still be coordinated.
- Visible user impact: user/admin traditional invite mutation endpoints can be hidden/disabled while old history remains available or archived. Old links can receive controlled redirect/retirement behavior.

### Option X3: Move History to `legacy_user_invites`, Then Drop Original Fields

**Summary**: archive traditional invite data into a legacy table, then remove live fields from `users`.

**Migration skeleton (pseudo SQL)**

```sql
BEGIN;

CREATE TABLE IF NOT EXISTS legacy_user_invites (
  user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  invite_code VARCHAR(32),
  invited_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
  invite_bound_at TIMESTAMPTZ,
  archived_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  source VARCHAR(64) NOT NULL DEFAULT 'deprecate_traditional_invite_2026_05_12'
);

INSERT INTO legacy_user_invites (user_id, invite_code, invited_by_user_id, invite_bound_at)
SELECT id, invite_code, invited_by_user_id, invite_bound_at
FROM users
ON CONFLICT (user_id) DO UPDATE SET
  invite_code = EXCLUDED.invite_code,
  invited_by_user_id = EXCLUDED.invited_by_user_id,
  invite_bound_at = EXCLUDED.invite_bound_at;

-- Either keep original invite_* audit/reward tables as read-only, or copy them to legacy_* tables too.
ALTER TABLE users DROP COLUMN IF EXISTS invite_code;
ALTER TABLE users DROP COLUMN IF EXISTS invited_by_user_id;
ALTER TABLE users DROP COLUMN IF EXISTS invite_bound_at;

COMMIT;
```

**Rollback skeleton**

```sql
BEGIN;

ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_code VARCHAR(32);
ALTER TABLE users ADD COLUMN IF NOT EXISTS invited_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_bound_at TIMESTAMPTZ;

UPDATE users u
SET invite_code = l.invite_code,
    invited_by_user_id = l.invited_by_user_id,
    invite_bound_at = l.invite_bound_at
FROM legacy_user_invites l
WHERE u.id = l.user_id;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_invite_code_not_null ON users(invite_code) WHERE invite_code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_invited_by_user_id ON users(invited_by_user_id);

COMMIT;
```

**Risk / migration cost / rollback difficulty**

- Risk: medium. Live schema is cleaned, but all reads that need history must be retargeted to archive tables.
- Migration cost: high. Requires archive schema, backfill validation, ent regeneration, and admin/history read path decisions.
- Rollback difficulty: medium if archive is complete; high if reward/action tables are partially dropped or transformed.
- Stop-the-world needed: recommended during the archive/drop step, or use dual-read validation before final drop.
- Visible user impact: old live fields disappear; old history can still be shown through archive-only screens if implemented.

### Recommendation

Recommend **X2 first, then optionally X3 later**.

- X2 matches the user constraint: S0/S1/S2/S3 do not change code, and S4 can start with a reversible freeze/deprecation seal.
- X2 gives the strongest protection against agent confusion without a destructive migration.
- X1 is too destructive for the first implementation step because historical rewards and marketing links may still need audit.
- X3 is the clean end-state if the user wants schema hygiene, but it should be a second decision after X2 has proven no live code or users depend on traditional invite.

## 3. Historical Data Handling

### Users Already Registered Through Traditional Invite

| Topic | Draft Decision |
|---|---|
| Existing `users.invited_by_user_id` | Preserve as historical relationship in S4-1/S4-2. Do not erase or rewrite until the user chooses X1/X2/X3. |
| Existing `invite_relationship_events` | Preserve for audit. These events explain when a traditional invite relationship was bound or admin-rebound. |
| Existing `invite_reward_records` | Preserve as a separate legacy ledger. Do not merge into `user_affiliate_ledger` by default. |
| Existing balances credited by `ApplyBaseRechargeRewards` | Treat as already-paid user balance. Never claw back during deprecation. |
| Future base rewards from traditional invite | Stop generating new traditional invite rewards once S4 freeze is implemented. Affiliate recharge rebates remain active if affiliate is enabled. |

### Should Historical Rewards Be Backfilled Into Affiliate?

Default recommendation: **do not backfill historical traditional invite rewards into affiliate automatically**.

Reasons:

- Traditional invite paid both inviter and invitee using `InviteBaseRewardRate = 0.03`; affiliate uses different settings, freeze windows, quotas, tiering, and transfer semantics.
- Traditional rewards update `users.balance` immediately in `InviteService.ApplyBaseRechargeRewards`; affiliate accrues quota in `user_affiliates` / `user_affiliate_ledger` before transfer.
- Automatic backfill can double-pay if a user has both `invited_by_user_id` and `user_affiliates.inviter_id`.

Optional manual backfill policy if the user explicitly asks later:

1. Only backfill relationships where `user_affiliates.inviter_id IS NULL` and a valid `users.invited_by_user_id` exists.
2. Do not backfill historical money by default; only bind future affiliate relationship.
3. If money backfill is required, create a one-time `legacy_invite_migration` ledger action and exclude all `invite_reward_records` already reflected in `users.balance`.
4. Produce a dry-run report: affected users, conflicting users, estimated ledger amounts, and rows skipped due to existing affiliate inviter.

### Old Invite Links Already Shared

| Link type | Recommended handling |
|---|---|
| `/invite` route | Keep redirect to `/affiliate` during transition. Evidence: `origin/test/xlabapi:frontend-v2/src/router/index.tsx:86`, `origin/test/xlabapi:frontend/src/router/index.ts:222`. |
| `/register?invite=<legacy_invite_code>` | Do not silently accept as affiliate. If a mapping exists and user approves, resolve legacy `invite_code` to the user's `aff_code` and redirect to `/register?aff=<aff_code>` or equivalent. Otherwise show a retired-code message. |
| Admin `/admin/invites` | Keep redirect away or remove menu entry. Evidence: `origin/test/xlabapi:frontend-v2/src/router/index.tsx:121`, `origin/test/xlabapi:frontend/src/router/index.ts:536`. |
| Marketing pages mentioning “邀请码” | Rename based on meaning: registration gate = “注册准入码”; affiliate = “返利码/推广码”; traditional invite = “legacy invite code” only in archive/admin contexts. |

### 404 vs Redirect vs Silent Accept

Recommendation: **redirect with explicit messaging, not 404 and not silent accept**.

- 404 is too abrupt for existing shared links and creates support load.
- Silent accept is dangerous because it hides a semantic migration from `invite_code` to `aff_code`, and can bind the wrong reward model.
- Redirect to `/affiliate` for logged-in users or `/register` with a visible “legacy invite retired” banner is safest.

### Data Retention

- Keep traditional invite data for at least one billing/rebate dispute window after S4 freeze. Exact retention period: `[待用户确认]`.
- If X3 is selected later, archive `users.invite_code`, `invited_by_user_id`, `invite_bound_at` into `legacy_user_invites` before drop.
- If X1 is selected, require a verified backup and a dry-run report before destructive migration.

## 4. Seal Mechanism Against Accidental Re-Enablement

The goal is to prevent future agents from seeing dormant `InviteService`/`invite_code` code and “helpfully” turning it back on. The seal should be redundant across code, DB, CI, tests, and docs.

### 4.1 Code-Layer Seal

Add a deprecation banner at the top of `backend/internal/service/invite_service.go` in S4, not before.

```go
// DEPRECATED 2026-05-12: Traditional invite (`users.invite_code`) is frozen.
// Do not add new call sites, routes, UI, or reward writes for this service.
// Kept only for legacy read/audit until the user selects DB option X1/X2/X3.
// Decision: docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md
```

Add method-level comments to every public method that can be called by handlers/services:

```go
// DEPRECATED 2026-05-12: legacy traditional invite. Do not use for new signup or rebate flows.
func (s *InviteService) GenerateUniqueInviteCode(ctx context.Context) (string, error) { ... }

// DEPRECATED 2026-05-12: legacy traditional invite. Use AffiliateService.BindInviterByCode for affiliate.
func (s *InviteService) ResolveInviterByCode(ctx context.Context, code string) (*User, error) { ... }

// DEPRECATED 2026-05-12: legacy traditional invite read path. Keep only for archive/history until removal.
func (s *InviteService) GetSummary(ctx context.Context, userID int64) (*InviteSummary, error) { ... }

// DEPRECATED 2026-05-12: legacy traditional invite reward. Must not be called for new payments/redeems.
func (s *InviteService) ApplyBaseRechargeRewards(ctx context.Context, inviteeID int64, redeemCode *RedeemCode) error { ... }
```

Also seal these files/routes in S4:

- `backend/internal/handler/invite_handler.go`: banner that `/api/v1/invite/*` is legacy/read-only or disabled.
- `backend/internal/handler/admin/invite_handler.go`: banner that `/api/v1/admin/invites/*` is legacy admin audit only unless user explicitly reopens.
- `backend/internal/service/admin_service_invite.go`: banner that rebind/manual/recompute writes are deprecated.
- `backend/internal/service/wire.go` and generated `backend/cmd/server/wire_gen.go`: comments around `ProvideInviteService` usage if service remains wired for read-only history.

### 4.2 Database-Layer Seal

For X2 freeze, add comments to the freeze migration file:

```sql
-- frozen 2026-05-12 — DO NOT re-enable traditional invite (`users.invite_code`).
-- See docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md

COMMENT ON COLUMN users.invite_code IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite code. Do not write or re-enable. See docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md';
COMMENT ON COLUMN users.invited_by_user_id IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite relation. Read only for historical audit.';
COMMENT ON COLUMN users.invite_bound_at IS 'DEPRECATED 2026-05-12: frozen legacy invite bind timestamp.';
```

Optional hard trigger after code freeze:

```sql
CREATE OR REPLACE FUNCTION reject_legacy_invite_write()
RETURNS trigger AS $$
BEGIN
  IF NEW.invite_code IS DISTINCT FROM OLD.invite_code
     OR NEW.invited_by_user_id IS DISTINCT FROM OLD.invited_by_user_id
     OR NEW.invite_bound_at IS DISTINCT FROM OLD.invite_bound_at THEN
    RAISE EXCEPTION 'legacy traditional invite fields are frozen; see docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_reject_legacy_invite_write
BEFORE UPDATE OF invite_code, invited_by_user_id, invite_bound_at ON users
FOR EACH ROW EXECUTE FUNCTION reject_legacy_invite_write();
```

For MySQL-compatible deployments, column comments would look like:

```sql
ALTER TABLE users MODIFY invite_code VARCHAR(32)
  COMMENT 'DEPRECATED 2026-05-12: frozen legacy traditional invite code; do not re-enable; see docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md';
```

### 4.3 CI-Layer Seal

Add a grep-based guard after the S4 code freeze. Suggested landing location: `.github/workflows/ci.yml` or an existing backend lint workflow. Optional local mirror: `.git/hooks/pre-commit` or repo-managed `scripts/check-legacy-invite-seal.sh`.

Allowlist should include this decision doc, migration comments, explicit deprecation banners, generated ent until removal, and tests that enforce the seal.

Initial CI command sketch:

```sh
#!/usr/bin/env bash
set -euo pipefail

# Fail new non-allowlisted traditional invite code references.
rg -n "\bInviteService\b|\binvite_code\b|\binvited_by_user_id\b|\binvite_bound_at\b" \
  backend frontend frontend-v2 \
  -g '!backend/ent/**' \
  -g '!backend/migrations/*deprecate*invite*' \
  -g '!backend/internal/service/invite_service.go' \
  -g '!backend/internal/handler/invite_handler.go' \
  -g '!backend/internal/handler/admin/invite_handler.go' \
  -g '!**/*legacy*invite*test*' \
  && {
    echo "legacy traditional invite reference detected; use affiliate aff_code or update allowlist with explicit deprecation rationale" >&2
    exit 1
  } || true
```

Stricter variant after X1/X3 removal:

```sh
rg -n "\bInviteService\b|\binvite_code\b|\binvited_by_user_id\b|\binvite_bound_at\b" backend frontend frontend-v2 \
  -g '!docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md' \
  -g '!backend/migrations/*legacy*invite*'
# This command must return no matches.
```

### 4.4 Test-Layer Seal

Add a red assertion that fails if new code writes traditional invite fields.

Pseudo Go test:

```go
func TestLegacyTraditionalInviteFieldsRemainFrozen(t *testing.T) {
    ctx := context.Background()
    user := createTestUser(t, ctx)

    before := loadUser(t, ctx, user.ID)

    // Exercise all active signup/payment/redeem flows that should now use affiliate only.
    registerUserThroughAffiliateCode(t, ctx)
    completeCommercialRecharge(t, ctx)

    after := loadUser(t, ctx, user.ID)
    require.Equal(t, before.InviteCode, after.InviteCode, "legacy users.invite_code must not be written")
    require.Equal(t, before.InvitedByUserID, after.InvitedByUserID, "legacy invited_by_user_id must not be written")
    require.Equal(t, before.InviteBoundAt, after.InviteBoundAt, "legacy invite_bound_at must not be written")
}
```

Pseudo repository-level static test:

```go
func TestNoNewTraditionalInviteCallsites(t *testing.T) {
    out := run(t, "rg", "-n", "InviteService|invite_code|invited_by_user_id", "backend/internal", "frontend", "frontend-v2")
    violations := filterAllowedLegacyInviteReferences(out)
    require.Empty(t, violations, "traditional invite is deprecated; use AffiliateService / aff_code")
}
```

Expected red line: if a future agent re-adds `/invite` UI, writes `users.invite_code`, calls `InviteService.ApplyBaseRechargeRewards`, or adds new admin traditional invite mutation, tests fail before merge.

### 4.5 Documentation-Layer Seal

Add visible references in S4 after user approves this decision:

- `MIGRATION_TODO.md`: add “Traditional invite (`users.invite_code`) frozen; use affiliate `aff_code`; see this decision doc.”
- `CLAUDE.md` / `AGENTS.md` / `DEV_GUIDE.md`: add a short red-line note for agents: do not resurrect InviteService or `/invite`; affiliate is the supported invite/rebate system.
- `docs/superpowers/contracts/API-CONTRACT.md`: mark `/api/v1/invite/*` and `/api/v1/admin/invites/*` as legacy/deprecated once S4 starts.
- Coordination control plane: add to `coordination.brief.constraints` as a red line. Foreman owns this update, per task card.

Recommended wording:

```md
Red line: Traditional invite (`users.invite_code`, `InviteService`, `/api/v1/invite`, `/api/v1/admin/invites`) is deprecated/frozen. Do not reopen it. New invite/rebate work must use affiliate (`user_affiliates.aff_code`, `/user/aff`, `/admin/affiliates/*`). See docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md.
```

## 5. Implementation Route

No implementation happens before user review. Suggested S4-only sequence:

| Step | Action | Exit criteria |
|---|---|---|
| S4-1 | User reviews this decision doc and chooses DB option X1/X2/X3. Foreman adds coordination red line. | Explicit user approval and chosen DB path recorded. |
| S4-2 | Add code-layer deprecation banners and method comments. If X2 selected, add DB comments/freeze migration. | `InviteService` and admin/user invite handlers visibly sealed; no behavior change except optional write-freeze if approved. |
| S4-3 | Remove or hide traditional invite UI entry points. Keep affiliate `/affiliate` and `/admin/affiliates/*`. Rename ambiguous i18n copy. | `/invite` and `/admin/invites` cannot look like active product features; affiliate copy uses `aff_code` semantics. |
| S4-4 | Run chosen DB migration. Recommended first migration is X2 freeze; X3 archive/drop only after validation. | Migration applied; rollback tested; historical data report generated. |
| S4-5 | Clean call chain and dead code. Remove `InviteService` routes/wiring/tests only after DB strategy is complete. | CI seal passes; no active code writes/reads traditional invite except approved archive path. |

### Sequencing Rules

- Do not remove DB fields before code stops referencing them.
- Do not remove old links before marketing/support agrees on redirect message.
- Do not backfill affiliate relationships or money without a separate dry-run and user approval.
- Do not delete `invite_reward_records` until retention/dispute window is approved.

## 6. Risks and Rollback

### Irreversible or High-Risk Points

| Action | Risk | Reversibility |
|---|---|---|
| Dropping `users.invite_code`, `invited_by_user_id`, `invite_bound_at` | Breaks any live code or report still reading traditional invite fields. | Schema reversible; data reversible only from backup/archive. |
| Dropping `invite_reward_records` / `invite_relationship_events` | Loses historical reward and relationship audit. | High-risk; needs backup or archive. |
| Deleting `InviteService` and admin invite service code | Makes emergency legacy inspection/recompute harder. | Git revert possible; runtime restore depends on DB choice. |
| Silent mapping from `invite_code` to `aff_code` | Can bind wrong reward model and double-pay or misattribute users. | Hard to audit after the fact. Avoid by default. |
| Auto-backfilling money into affiliate quota | Double-pay risk because traditional rewards may already be in `users.balance`. | Hard; requires transaction-level audit. Avoid by default. |

### Where the User Can Stop

| Stop point | What remains true |
|---|---|
| Before S4-1 approval | Nothing changes; this is only a draft. |
| After S4-2 banners/comments | Behavior can remain unchanged; deprecation is visible and reversible. |
| After X2 freeze/comments | Historical data remains in place; write-freeze can be rolled back by removing trigger/comments and reverting code. |
| Before X3 archive/drop | Live schema still has original fields; safest point to pause for more audit. |
| Before deleting service code | Legacy code remains available for inspection, even if hidden from UI. |

### Rollback Playbook

1. If UI removal causes support issues: restore `/invite` redirect and show a legacy-retired explanation; do not restore traditional reward writes.
2. If X2 trigger blocks a legitimate admin audit tool: temporarily disable only the trigger, keep code deprecation banners, and open a follow-up issue.
3. If X3 archive has data mismatch: stop before dropping fields, rerun archive dry-run, compare row counts and checksums.
4. If X1 drop already ran and data is needed: restore from database backup into `legacy_user_invites` rather than re-enabling live `users.invite_code`.
5. Any rollback that re-enables traditional invite mutations requires explicit user approval because it violates the chosen product direction.
