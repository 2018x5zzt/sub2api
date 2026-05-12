# Decision: Deprecate Traditional Invite (v2)

- Date: 2026-05-12
- Status: Active pending reviewer re-review. Do not implement in S0/S1/S2/S3; S4 implementation still requires reviewer pass and user-approved task scope.
- Decision intent: deprecate the traditional invite-growth system backed by `users.invite_code`; keep the affiliate rebate system backed by `user_affiliates.aff_code`.
- Scope rule: this document authorizes planning and S4 prerequisites only. It does not authorize code, migration, UI, config, or `API-CONTRACT.md` changes before S4.
- Primary references: `docs/superpowers/audits/2026-05-12-semantic-dual-tracks.md`; `docs/superpowers/reviews/2026-05-12-review-deprecate-invite.md`; `docs/superpowers/contracts/API-CONTRACT.md`; `origin/test/xlabapi` backend/frontend grep evidence.

## v2 Revision Record: T013 P0/P1 Fixes

This v2 keeps the v1 product direction unchanged: traditional invite is deprecated, affiliate is retained, and X2 freeze is recommended before any destructive X1/X3 path.

| Review item | v2 decision |
|---|---|
| P0-1 HTTP compatibility matrix | Added `S4 Endpoint Behavior Matrix`. Every listed endpoint is assigned exactly one S4 behavior: `keep`, `read-only`, or `deprecation-response`. |
| P0-2 X2 database freeze | X2 is split into comment seal first and hard freeze as S4 exit condition. Hard freeze covers `users` INSERT/UPDATE plus four legacy invite tables. |
| P0-3 CI static seal baseline | Replaced whole-file allowlist with exact baseline comparison and required route/wire scans plus a negative self-test. |
| P1-4 feature flag explicit ban | Added a ban on new traditional-invite enable flags without new user-approved decision. |
| P1-5 affiliate continuity | Added four S4 entry tests that must pass before traditional reward writes are disabled. |
| P1-6 X3 archive completeness | X3 now archives all legacy invite tables, has dry-run checks, and requires admin/history read paths to explain legacy disputes without live dropped fields. |

## 1. Deprecation Scope

### To Deprecate: Traditional Invite

| Layer | Item | File:line evidence | Notes |
|---|---|---|---|
| DB field | `users.invite_code` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:1`; `origin/test/xlabapi:backend/ent/schema/user.go:57` | User-owned traditional invite relation code. Planned deprecated source of truth. |
| DB field | `users.invited_by_user_id` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:3`; `origin/test/xlabapi:backend/ent/schema/user.go:62` | Traditional inviter relation. Freeze first, archive/drop only after S4 decision. |
| DB field | `users.invite_bound_at` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:4`; `origin/test/xlabapi:backend/ent/schema/user.go:65` | Traditional invite binding timestamp. |
| DB table | `invite_code_aliases` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:10`; `origin/test/xlabapi:backend/internal/repository/user_repo.go:886` | Legacy alias lookup for migrated invite codes. |
| DB table | `invite_relationship_events` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:97` | Audit table for traditional invite relationship changes. |
| DB table | `invite_reward_records` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:138`; `origin/test/xlabapi:backend/internal/service/invite.go:30` | Traditional base/manual/recompute reward ledger. Do not merge with affiliate ledger without explicit migration decision. |
| DB table | `invite_admin_actions` | `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:86`; `origin/test/xlabapi:backend/internal/service/invite.go:59` | Traditional invite admin action audit. |
| Ent/generated | `ent.User.InviteCode` and generated mutation/predicate helpers | `origin/test/xlabapi:backend/ent/user.go:38`; `origin/test/xlabapi:backend/ent/user/user.go:36`; `origin/test/xlabapi:backend/ent/mutation.go:41846` | Generated code disappears only if schema fields are removed and ent is regenerated. |
| Repository | `GetByInviteCode`, `ExistsByInviteCode`, alias lookup | `origin/test/xlabapi:backend/internal/repository/user_repo.go:170`; `origin/test/xlabapi:backend/internal/repository/user_repo.go:795`; `origin/test/xlabapi:backend/internal/repository/user_repo.go:881` | Read path for traditional codes and alias compatibility. |
| Service | `InviteService` | `origin/test/xlabapi:backend/internal/service/invite_service.go:35` | Main traditional invite service. Top-level deprecation banner should land here in S4. |
| Service method | `GenerateUniqueInviteCode` | `origin/test/xlabapi:backend/internal/service/invite_service.go:119` | Writes/generates traditional 8-letter invite code. Must be unreachable after S4 freeze. |
| Service method | `ResolveInviterByCode` | `origin/test/xlabapi:backend/internal/service/invite_service.go:136` | Traditional code-to-inviter lookup. Retain only if a read-only legacy link policy explicitly needs it. |
| Service method | `GetSummary`, `ListRewards` | `origin/test/xlabapi:backend/internal/service/invite_service.go:175`; `origin/test/xlabapi:backend/internal/service/invite_service.go:199` | User-facing traditional invite summary/reward read path. S4 behavior is read-only. |
| Service method | `ApplyBaseRechargeRewards` | `origin/test/xlabapi:backend/internal/service/invite_service.go:210` | Traditional 3% base reward path for commercial balance redeem codes. Must not be called after S4 freeze. |
| Admin service | `RebindInviter`, `CreateManualInviteGrant`, recompute | `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:52`; `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:114`; `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:168`; `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:228` | Admin-only traditional invite mutation. Disabled by default after S4 freeze. |
| User API | `GET /api/v1/invite/summary`, `GET /api/v1/invite/rewards` | `origin/test/xlabapi:backend/internal/server/routes/user.go:106`; `origin/test/xlabapi:backend/internal/handler/invite_handler.go:19`; `origin/test/xlabapi:backend/internal/handler/invite_handler.go:35` | User-facing traditional invite endpoints. S4 behavior is read-only history. |
| Admin API | `/api/v1/admin/invites/*` | `origin/test/xlabapi:backend/internal/server/routes/admin.go:619`; `origin/test/xlabapi:backend/internal/handler/admin/invite_handler.go:61`; `origin/test/xlabapi:backend/internal/handler/admin/invite_handler.go:115` | Admin traditional invite history remains read-only; mutations return deprecation response. |
| DTO | `InviteSummary`, `InviteRewardRecord`, admin invite DTOs | `origin/test/xlabapi:backend/internal/handler/dto/invite.go:9`; `origin/test/xlabapi:backend/internal/handler/dto/invite.go:17`; `origin/test/xlabapi:backend/internal/handler/dto/admin_invite.go:12`; `origin/test/xlabapi:backend/internal/handler/dto/admin_invite.go:20` | DTOs for deprecated/read-only endpoints. |
| Wiring | `ProvideInviteService`, `NewInviteHandler` | `origin/test/xlabapi:backend/internal/service/wire.go:32`; `origin/test/xlabapi:backend/cmd/server/wire_gen.go:96` | Injector/wire entrypoints must be included in CI seal scans. |
| Frontend route | `/invite` redirect | `origin/test/xlabapi:frontend-v2/src/router/index.tsx:86`; `origin/test/xlabapi:frontend/src/router/index.ts:222` | Already redirects to `/affiliate`; keep compatibility redirect until user approves removal. |
| Admin frontend route | `/admin/invites` redirect | `origin/test/xlabapi:frontend-v2/src/router/index.tsx:121`; `origin/test/xlabapi:frontend/src/router/index.ts:536` | Already redirects away; should remain sealed or be removed later. |
| UI/i18n copy | Traditional-looking invite labels under affiliate and registration contexts | `origin/test/xlabapi:frontend/src/i18n/locales/zh.ts:1002`; `origin/test/xlabapi:frontend/src/i18n/locales/zh.ts:1039`; `origin/test/xlabapi:frontend-v2/src/i18n/locales/zh.ts:5301`; `origin/test/xlabapi:frontend-v2/src/i18n/locales/zh.ts:5302` | Rename by meaning: affiliate is promotion/rebate code; registration invitation remains a separate gate. |

### Not Deprecated: Keep Affiliate

| Layer | Item | File:line evidence | Notes |
|---|---|---|---|
| DB table | `user_affiliates` | `origin/test/xlabapi:backend/migrations/130_add_user_affiliates.sql:1` | Kept source of truth for rebate/referral. |
| DB field | `user_affiliates.aff_code` | `origin/test/xlabapi:backend/migrations/130_add_user_affiliates.sql:3`; `origin/test/xlabapi:backend/migrations/130_add_user_affiliates.sql:16` | Kept affiliate code. UI should avoid bare "邀请码" if it causes confusion. |
| Service | `AffiliateService` and repository port | `origin/test/xlabapi:backend/internal/service/affiliate_service.go:60`; `origin/test/xlabapi:backend/internal/service/affiliate_service.go:98` | Kept. Handles profile, aff code lookup, binding, rebate accrual, transfer, admin settings. |
| Auth signup binding | `affiliateCode` binding on email/OAuth registration | `origin/test/xlabapi:backend/internal/service/auth_service.go:137`; `origin/test/xlabapi:backend/internal/service/auth_service.go:237`; `origin/test/xlabapi:backend/internal/service/auth_service.go:570`; `origin/test/xlabapi:backend/internal/service/auth_service.go:789` | Keep this path. It is not `users.invite_code`. |
| User API | `GET /api/v1/user/aff`, `POST /api/v1/user/aff/transfer` | `origin/test/xlabapi:backend/internal/handler/user_handler.go:171`; `origin/test/xlabapi:backend/internal/handler/user_handler.go:188`; `origin/test/xlabapi:backend/internal/server/routes/user.go:28` | Kept. |
| Admin API | `/api/v1/admin/affiliates/*` | `origin/test/xlabapi:backend/internal/server/routes/admin.go:599`; `origin/test/xlabapi:backend/internal/handler/admin/affiliate_handler.go:15`; `origin/test/xlabapi:backend/internal/handler/admin/affiliate_handler.go:49` | Kept. |
| Frontend route | `/affiliate`, `/admin/affiliates/*` | `origin/test/xlabapi:frontend-v2/src/router/index.tsx:87`; `origin/test/xlabapi:frontend-v2/src/router/index.tsx:122`; `origin/test/xlabapi:frontend/src/router/index.ts:226`; `origin/test/xlabapi:frontend/src/router/index.ts:576` | Kept. |
| UI copy | Affiliate user/admin pages | `origin/test/xlabapi:frontend-v2/src/pages/user/Affiliate.tsx:55`; `origin/test/xlabapi:frontend-v2/src/pages/admin/AffiliateRecords.tsx:102`; `origin/test/xlabapi:frontend/src/views/user/AffiliateView.vue:164`; `origin/test/xlabapi:frontend/src/views/admin/affiliates/AdminAffiliateRecordsTable.vue:256` | Keep behavior; rename labels if needed to "affiliate code / 推广码 / 返利码". |

### Boundary Cases

| Case | Decision |
|---|---|
| User has `users.invite_code` and `user_affiliates.aff_code` | Treat `aff_code` as the only future shareable code. Keep `invite_code` only for historical audit/compatibility until chosen DB option runs. |
| User was registered via traditional `invited_by_user_id`, but also has affiliate inviter | Preserve legacy relation as history. Future reward writes go through affiliate only when affiliate relation is active. Never double-pay rewards. |
| Historical `invite_reward_records` and affiliate ledger both exist for same recharge | Keep ledgers separate. If backfilling later, mark migrated rows and exclude already rewarded transactions. |
| Existing `/invite?invite=...` marketing links | Do not silently accept traditional codes as affiliate codes. The default post-S4 behavior is a clear deprecation page/response or a redirect that does not bind a legacy inviter. |
| Registration invitation code `redeem_codes.type=invitation` | Not in scope. It remains a registration gate. Its UI text should be "registration invitation / 注册准入码" if ambiguity persists. |

## 2. S4 Endpoint Behavior Matrix

Allowed S4 behavior values are exactly `keep`, `read-only`, and `deprecation-response`.

| Endpoint group | Contract identifier | S4 behavior | Exact S4 rule | Evidence / current contract risk |
|---|---|---|---|---|
| Affiliate user routes | `/api/v1/user/aff` and all subroutes, including `GET /api/v1/user/aff` and `POST /api/v1/user/aff/transfer` | `keep` | No traditional-invite deprecation behavior applies. Existing request/response contract remains active. | `origin/test/xlabapi:backend/internal/server/routes/user.go:28`; `origin/test/xlabapi:backend/internal/handler/user_handler.go:171`; `origin/test/xlabapi:backend/internal/handler/user_handler.go:188` |
| Affiliate admin routes | `/api/v1/admin/affiliates/*` | `keep` | No traditional-invite deprecation behavior applies. Admin affiliate list/settings/records/rebate actions remain active. | `origin/test/xlabapi:backend/internal/server/routes/admin.go:599`; `origin/test/xlabapi:backend/internal/handler/admin/affiliate_handler.go:15`; `origin/test/xlabapi:backend/internal/handler/admin/affiliate_handler.go:49` |
| User legacy invite summary | `GET /api/v1/invite/summary` | `read-only` | Return legacy summary/history only. Must not call or indirectly trigger `GenerateUniqueInviteCode`; must not ensure, generate, or update `users.invite_code`, `users.invited_by_user_id`, or `users.invite_bound_at`. If the user has no legacy data, return empty/null legacy fields with a stable success response. | Current T006 marks this endpoint active and notes it may ensure/generate a traditional invite code: `docs/superpowers/contracts/API-CONTRACT.md:166`. Handler evidence: `origin/test/xlabapi:backend/internal/handler/invite_handler.go:19`. |
| User legacy invite rewards | `GET /api/v1/invite/rewards` | `read-only` | Return historical `invite_reward_records` only. Must not create rewards, recompute, backfill, or mutate legacy invite tables. | Handler evidence: `origin/test/xlabapi:backend/internal/handler/invite_handler.go:35`. |
| Admin legacy invite stats | `GET /api/v1/admin/invites/stats` | `read-only` | Return archive/history statistics from frozen legacy data only. No recompute, no repair writes. | Route group evidence: `origin/test/xlabapi:backend/internal/server/routes/admin.go:619`; T006 marks admin invite endpoints active: `docs/superpowers/contracts/API-CONTRACT.md:179`. |
| Admin legacy invite relationships | `GET /api/v1/admin/invites/relationships` | `read-only` | Return archive/history relationship rows only. No rebind side effect. | `origin/test/xlabapi:backend/internal/handler/admin/invite_handler.go:61`. |
| Admin legacy invite rewards | `GET /api/v1/admin/invites/rewards` | `read-only` | Return archive/history reward rows only. No manual grant, no recompute side effect. | `origin/test/xlabapi:backend/internal/handler/admin/invite_handler.go:115`. |
| Admin legacy invite actions | `GET /api/v1/admin/invites/actions` | `read-only` | Return archive/history admin action rows only. No mutation. | `origin/test/xlabapi:backend/internal/server/routes/admin.go:619`. |
| Admin rebind legacy inviter | `POST /api/v1/admin/invites/rebind` | `deprecation-response` | Disabled by default after S4 freeze. Return `410 Gone` with stable body `{ "error": "legacy_invite_deprecated", "message": "Traditional invite mutations are retired" }`. Re-enable only by new user-approved decision. | `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:52`. |
| Admin manual legacy grant | `POST /api/v1/admin/invites/manual-grants` | `deprecation-response` | Disabled by default after S4 freeze. Return `410 Gone` with the same stable deprecation body. | `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:114`. |
| Admin recompute preview/execute | `POST /api/v1/admin/invites/recompute/execute` and any mutation under `/api/v1/admin/invites/recompute/*` | `deprecation-response` | Mutation execution disabled by default after S4 freeze. Return `410 Gone` with the same stable deprecation body. Read-only preview may exist only if explicitly represented as a GET/read-only contract; otherwise keep disabled. | `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:168`; `origin/test/xlabapi:backend/internal/service/admin_service_invite.go:228`. |

S4 must update contract tests to prove the chosen behavior. In particular, `GET /api/v1/invite/summary` must be tested against a user without `users.invite_code` and must not create one.

## 3. Database Compatibility Plan

### Recommendation

Use X2 first: retain legacy schema for readable history, freeze writes in two stages, and defer destructive cleanup. X1 and X3 are not S4 defaults unless the user explicitly chooses them after reviewing dry-run evidence.

### Option X1: Drop `users.invite_code` Directly

**Summary**: remove traditional invite fields/tables from live schema after code paths are sealed and removed.

```sql
-- S4+ only, after code no longer references traditional invite fields.
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

Rollback: recreate columns/tables from migration 139 and restore data from backup. Schema rollback is possible; data rollback is not reliable without pre-migration backup.

Operational judgement: not recommended as first S4 step because it is the highest-risk and least reversible option.

### Option X2: Keep Fields, Freeze Writes, Add Deprecated Comments

**Summary**: retain traditional invite schema for compatibility/history, block new writes, and mark every layer as deprecated.

X2 is split into two required stages:

| Stage | Name | Timing | Rule |
|---|---|---|---|
| X2-A | Comment seal | First S4 DB change | Add DB comments and migration banners. No app behavior change should depend solely on comments. |
| X2-B | Hard freeze | S4 exit condition before claiming deprecation complete | Add DB enforcement that rejects new legacy invite writes. S4 cannot close until this is either deployed or explicitly waived by user with a new decision. |

**X2-A migration skeleton**

```sql
BEGIN;

COMMENT ON COLUMN users.invite_code IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite code. Do not write or re-enable. See docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md';
COMMENT ON COLUMN users.invited_by_user_id IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite relation. Read only for historical audit.';
COMMENT ON COLUMN users.invite_bound_at IS 'DEPRECATED 2026-05-12: frozen legacy traditional invite bind timestamp.';
COMMENT ON TABLE invite_code_aliases IS 'DEPRECATED 2026-05-12: legacy invite aliases retained for audit only.';
COMMENT ON TABLE invite_relationship_events IS 'DEPRECATED 2026-05-12: legacy traditional invite relationship audit, read only.';
COMMENT ON TABLE invite_reward_records IS 'DEPRECATED 2026-05-12: legacy traditional invite rewards, read only.';
COMMENT ON TABLE invite_admin_actions IS 'DEPRECATED 2026-05-12: legacy traditional invite admin audit, read only.';

COMMIT;
```

**X2-B hard-freeze skeleton**

```sql
-- Pseudo PostgreSQL shape; exact dialect must match the live DB.
CREATE OR REPLACE FUNCTION reject_users_legacy_invite_insert()
RETURNS trigger AS $$
BEGIN
  IF NEW.invite_code IS NOT NULL
     OR NEW.invited_by_user_id IS NOT NULL
     OR NEW.invite_bound_at IS NOT NULL THEN
    RAISE EXCEPTION 'legacy traditional invite fields are frozen; see docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION reject_users_legacy_invite_update()
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

CREATE TRIGGER users_legacy_invite_insert_frozen
BEFORE INSERT ON users
FOR EACH ROW EXECUTE FUNCTION reject_users_legacy_invite_insert();

CREATE TRIGGER users_legacy_invite_update_frozen
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION reject_users_legacy_invite_update();

CREATE OR REPLACE FUNCTION reject_legacy_invite_table_write()
RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'legacy traditional invite tables are read-only; see docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER invite_code_aliases_read_only
BEFORE INSERT OR UPDATE OR DELETE ON invite_code_aliases
FOR EACH ROW EXECUTE FUNCTION reject_legacy_invite_table_write();

CREATE TRIGGER invite_relationship_events_read_only
BEFORE INSERT OR UPDATE OR DELETE ON invite_relationship_events
FOR EACH ROW EXECUTE FUNCTION reject_legacy_invite_table_write();

CREATE TRIGGER invite_reward_records_read_only
BEFORE INSERT OR UPDATE OR DELETE ON invite_reward_records
FOR EACH ROW EXECUTE FUNCTION reject_legacy_invite_table_write();

CREATE TRIGGER invite_admin_actions_read_only
BEFORE INSERT OR UPDATE OR DELETE ON invite_admin_actions
FOR EACH ROW EXECUTE FUNCTION reject_legacy_invite_table_write();
```

Rollback: drop the triggers/functions and remove comments. No data restore is needed because X2 does not delete history. If rollback re-enables writes, it requires a new user-approved decision because this document intentionally freezes the feature.

Stop-the-world: not normally required for comments/triggers, but X2-B should be deployed in a low-traffic window and preceded by a dry-run scan that proves current app flows no longer attempt legacy writes.

Visible user impact: legacy history can remain readable; new traditional invite binding/reward writes stop. Affiliate behavior must remain available if `affiliate_enabled=true`.

### Option X3: Archive Legacy Data, Then Drop Live Fields

**Summary**: move legacy traditional invite data to archive tables, keep read paths pointed at archives, then drop live fields/tables only after dry-run checks pass.

X3 policy: archive all legacy invite tables rather than leaving some in place. The archive set is:

| Live source | Archive target | Required preservation |
|---|---|---|
| `users.invite_code`, `users.invited_by_user_id`, `users.invite_bound_at` | `legacy_user_invites` | User ID, invite code, inviter user ID, bound timestamp, archival timestamp. |
| `invite_code_aliases` | `legacy_invite_code_aliases` | Alias, canonical code/user, created/updated metadata. |
| `invite_relationship_events` | `legacy_invite_relationship_events` | Relationship event history sufficient to explain bind/rebind/unbind disputes. |
| `invite_reward_records` | `legacy_invite_reward_records` | Reward type, source transaction/order, amount, status, created metadata. |
| `invite_admin_actions` | `legacy_invite_admin_actions` | Actor, action type, before/after payload, reason, created metadata. |

**Migration skeleton**

```sql
BEGIN;

CREATE TABLE legacy_user_invites AS
SELECT id AS user_id, invite_code, invited_by_user_id, invite_bound_at, NOW() AS archived_at
FROM users
WHERE invite_code IS NOT NULL OR invited_by_user_id IS NOT NULL OR invite_bound_at IS NOT NULL;

CREATE TABLE legacy_invite_code_aliases AS SELECT *, NOW() AS archived_at FROM invite_code_aliases;
CREATE TABLE legacy_invite_relationship_events AS SELECT *, NOW() AS archived_at FROM invite_relationship_events;
CREATE TABLE legacy_invite_reward_records AS SELECT *, NOW() AS archived_at FROM invite_reward_records;
CREATE TABLE legacy_invite_admin_actions AS SELECT *, NOW() AS archived_at FROM invite_admin_actions;

-- Only after dry-run checks, read-path changes, and backups:
-- ALTER TABLE users DROP COLUMN invite_code;
-- ALTER TABLE users DROP COLUMN invited_by_user_id;
-- ALTER TABLE users DROP COLUMN invite_bound_at;
-- DROP TABLE invite_code_aliases;
-- DROP TABLE invite_relationship_events;
-- DROP TABLE invite_reward_records;
-- DROP TABLE invite_admin_actions;

COMMIT;
```

**X3 dry-run checks**

| Check | Pass condition |
|---|---|
| Row counts | Source row counts match archive row counts for every source table. |
| Null inviter count | Count and sample rows where `invited_by_user_id IS NULL` are reported and accepted as explainable history. |
| Duplicate invite code | Duplicate `users.invite_code` and duplicate aliases are reported before archive/drop. |
| Alias coverage | Every alias points to a retained canonical user/code or is explicitly marked orphaned. |
| Reward/action/event counts | Reward, action, and event counts match source tables after archive copy. |
| Sample chain trace | Sample `invitee -> inviter -> reward -> admin action/event` chains can be followed from archive tables alone. |

X3 exit condition: admin/history read paths must not query dropped live fields or dropped live tables, yet must still explain a legacy invite dispute from archive data. If that read path is not implemented and tested, X3 cannot drop live columns/tables.

Rollback: before drop, rollback by discarding archive tables. After drop, rollback requires restoring live fields/tables and source data from backup or archive tables; treat this as a maintenance-window migration.

Stop-the-world: likely yes for the final drop/read-path switch.

Visible user impact: legacy history remains available in archive/admin contexts only. Shareable traditional invite codes are not restored.

## 4. Historical Data Handling

| Topic | Decision |
|---|---|
| Users already registered via traditional invite | Preserve relation as legacy history. Do not convert automatically into affiliate unless user approves a separate migration/backfill decision. |
| Historical base rewards | Keep `invite_reward_records` readable as legacy reward history. Do not backfill into affiliate ledger by default. |
| Future paid order/recharge rewards | After S4 freeze, do not call `InviteService.ApplyBaseRechargeRewards`. Affiliate ledger/quota accrual remains the future path when affiliate is enabled. |
| Old invite_code links | Do not silently accept as affiliate. The default post-S4 behavior is a clear deprecation page/response or a redirect that does not bind a legacy inviter. |
| Old admin corrections | Read-only archive/history remains available. Mutating rebind/manual grant/recompute endpoints return `410 Gone` unless user re-approves them in a new decision. |

Open policy point for user review: if marketing has active legacy invite links, the S4 implementation task should inventory them and choose a product message before disabling UI entry points. This does not reopen traditional invite writes.

## 5. Seal Mechanism Against Accidental Re-Enablement

### 5.1 Code Layer Seal

S4 should add a deprecation banner at the top of `backend/internal/service/invite_service.go` and public legacy invite methods.

```go
// DEPRECATED 2026-05-12: traditional invite is retired.
// Do not add new callers or re-enable writes. See:
// docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md
```

Required method-level comments:

| Method/path | Required S4 comment |
|---|---|
| `GenerateUniqueInviteCode` | `DEPRECATED 2026-05-12: must not be called after S4 freeze.` |
| `ResolveInviterByCode` | `DEPRECATED 2026-05-12: legacy read/compatibility only; no new binding policy.` |
| `GetSummary`, `ListRewards` | `DEPRECATED 2026-05-12: read-only legacy history only.` |
| `ApplyBaseRechargeRewards` | `DEPRECATED 2026-05-12: retired reward writer; S4 payment flows must not call this.` |
| admin rebind/manual/recompute methods | `DEPRECATED 2026-05-12: mutation disabled by default; requires new user-approved decision.` |

### 5.2 Database Layer Seal

Every S4 migration touching legacy invite must include this banner:

```sql
-- FROZEN 2026-05-12: traditional invite is retired.
-- DO NOT re-enable users.invite_code, users.invited_by_user_id, users.invite_bound_at,
-- invite_code_aliases, invite_relationship_events, invite_reward_records, or invite_admin_actions.
-- See docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md
```

Column/table comments are required in X2-A. Hard write rejection is required in X2-B unless a new user-approved decision explicitly waives it.

### 5.3 CI Static Seal: Baseline, Not Whole-File Allowlist

The CI seal must compare exact known references against a checked-in baseline. Whole-file allowlists are forbidden because they hide the most likely accidental re-enablements.

Baseline file proposal:

```text
# docs/superpowers/seals/legacy-invite-baseline.txt
# format: file:line:function_or_context:reason
backend/internal/service/invite_service.go:35:InviteService:deprecated service definition only
backend/internal/service/invite_service.go:119:GenerateUniqueInviteCode:deprecated, no new callers
backend/internal/service/invite_service.go:136:ResolveInviterByCode:legacy read compatibility only
backend/internal/service/invite_service.go:175:GetSummary:read-only legacy summary
backend/internal/service/invite_service.go:199:ListRewards:read-only legacy rewards
backend/internal/service/invite_service.go:210:ApplyBaseRechargeRewards:retired reward writer
backend/internal/service/admin_service_invite.go:52:RebindInviter:disabled after S4
backend/internal/service/admin_service_invite.go:114:CreateManualInviteGrant:disabled after S4
backend/internal/service/admin_service_invite.go:168:RecomputePreview:read-only only if kept
backend/internal/service/admin_service_invite.go:228:RecomputeExecute:disabled after S4
backend/internal/server/routes/user.go:106:invite routes:read-only endpoints only
backend/internal/server/routes/admin.go:619:admin invite routes:read-only or 410 only
backend/internal/service/wire.go:32:ProvideInviteService:legacy injection until removal
backend/cmd/server/wire_gen.go:96:wire generated:legacy injection until removal
```

CI scan scope must include at least:

```sh
rg -n -i 'invite_code|invited_by_user_id|invite_bound_at|InviteService|ApplyBaseRechargeRewards|GenerateUniqueInviteCode|/api/v1/invite|/admin/invites|invite_enabled|traditional_invite_enabled|enable_traditional_invite' backend frontend frontend-v2
```

Routes/wire files are mandatory scan inputs:

| File | Why mandatory |
|---|---|
| `backend/internal/server/routes/user.go` | Can expose or re-enable user legacy invite routes. |
| `backend/internal/server/routes/admin.go` | Can expose or re-enable admin mutation routes. |
| `backend/internal/service/wire.go` | Can keep or recreate legacy service injection. |
| `backend/cmd/server/wire_gen.go` | Generated injection can retain references that must be tracked. |

CI comparison rule:

```sh
# Pseudo flow, not final implementation.
rg -n -i 'invite_code|invited_by_user_id|invite_bound_at|InviteService|ApplyBaseRechargeRewards|GenerateUniqueInviteCode|/api/v1/invite|/admin/invites|invite_enabled|traditional_invite_enabled|enable_traditional_invite' backend frontend frontend-v2 \
  | normalize_to_file_line_func \
  | sort > /tmp/legacy-invite-current.txt

sort docs/superpowers/seals/legacy-invite-baseline.txt > /tmp/legacy-invite-baseline.txt
comm -13 /tmp/legacy-invite-baseline.txt /tmp/legacy-invite-current.txt > /tmp/legacy-invite-new.txt

test ! -s /tmp/legacy-invite-new.txt
```

Non-zero `rg` behavior must be handled explicitly: no matches is success only when the baseline is expected empty; scanner errors are failures.

Negative self-test requirement: CI must include a unit/script test that injects a fixture such as `backend/internal/service/not_in_baseline_fixture.go` containing a fake `InviteService.ApplyBaseRechargeRewards(...)` call outside the baseline. The test must assert the scanner fails and reports the extra reference.

### 5.4 Feature Flag Ban

No new setting, config, environment variable, database setting, or frontend flag may re-enable traditional invite without a new user-approved decision document.

Forbidden names include, but are not limited to:

| Forbidden pattern | Reason |
|---|---|
| `invite_enabled` | Ambiguous and likely to be mistaken for traditional invite resurrection. |
| `traditional_invite_enabled` | Directly reopens deprecated feature. |
| `legacy_invite_enabled` | Directly reopens deprecated feature. |
| `enable_traditional_invite` | Directly reopens deprecated feature. |
| Reusing `invitation_code_enabled` for traditional invite | `invitation_code_enabled` belongs to registration gate semantics, not invite-growth semantics. |

CI static scan must include negative grep for these names. A future need to support a temporary operational escape hatch must start a new decision document and get explicit user approval.

### 5.5 Test Layer Seal

S4 tests must include at least these red assertions:

```go
func TestLegacyInviteFrozen_NoNewUserInviteFields(t *testing.T) {
    user := createUserWithAffiliateCode(t, "AFF123")
    require.Empty(t, user.InviteCode)
    require.Nil(t, user.InvitedByUserID)
    require.Nil(t, user.InviteBoundAt)
}

func TestLegacyInviteFrozen_NoBaseRechargeRewardCall(t *testing.T) {
    svc := NewPaymentFlowWithSpyInviteService()
    svc.MarkOrderPaid(orderWithAffiliateEnabled())
    require.False(t, svc.InviteSpy.ApplyBaseRechargeRewardsCalled)
}

func TestInviteSummaryReadOnly_DoesNotGenerateCode(t *testing.T) {
    user := createUserWithoutInviteCode(t)
    callGET(t, "/api/v1/invite/summary", user)
    reloaded := loadUser(t, user.ID)
    require.Empty(t, reloaded.InviteCode)
}
```

### 5.6 Documentation Layer Seal

S4 implementation should add references to this decision in high-signal docs only:

| Doc/control plane | Required note |
|---|---|
| `MIGRATION_TODO.md` | Traditional invite is frozen; do not re-enable. |
| `CLAUDE.md` / `AGENTS.md` / `DEV_GUIDE.md` if present | Point agents to this decision before touching invite/affiliate/redeem code. |
| `coordination.brief.constraints` | Foreman-owned redline: traditional invite must not be reopened without user approval. |
| `docs/superpowers/contracts/API-CONTRACT.md` | Update in S4 per contract delta below; not changed by this T014 task. |

### 5.7 UI Copy Checklist

S4 UI work must rename ambiguous copy by meaning:

| Concept | Allowed labels |
|---|---|
| Registration gate | `registration invitation`, `注册准入码` |
| Affiliate | `affiliate code`, `promotion code`, `推广码`, `返利码` |
| Traditional invite history | `legacy invite code`, `历史邀请关系码`; only in archive/admin contexts |

Implementation inventory must include the frontend and frontend-v2 i18n/page files already cited in section 1, plus any route/page copy found by `rg -n -i '邀请码|invite|affiliate|推广|返利' frontend frontend-v2`.

## 6. Affiliate Continuity: S4 Entry Conditions

Before S4 disables traditional reward writes, these tests must pass in the target branch/environment:

| Gate | Required passing behavior |
|---|---|
| Signup with `aff_code` | Creates/binds `user_affiliates.inviter_id`; does not write `users.invited_by_user_id`, `users.invite_code`, or `users.invite_bound_at`. |
| Paid order/recharge with `affiliate_enabled=true` | Accrues affiliate ledger/quota through affiliate paths. |
| Paid order/recharge after S4 freeze | Does not call `InviteService.ApplyBaseRechargeRewards`. |
| User has both legacy invite history and affiliate inviter | Does not double-pay. Legacy records remain history; affiliate reward path is the only future reward path. |

If `affiliate_enabled=false` in a target environment, product must explicitly accept that there may be no invite/rebate reward path after traditional invite freeze. If that is not acceptable, S4 cannot proceed until affiliate is enabled and tested.

## 7. Implementation Route

This is sequence only. It is not permission to write code before S4.

| Step | Work | Exit condition |
|---|---|---|
| S4-1 | User/reviewer approves this v2 decision and S4 task scope. | Reviewer pass recorded; no open P0. |
| S4-2 | Add code-level deprecation banners/comments. | Static seal baseline includes only approved legacy references. |
| S4-3 | Update `API-CONTRACT.md` and contract tests according to the endpoint behavior matrix. | Contract tests prove keep/read-only/410 behavior. |
| S4-4 | Validate affiliate continuity gates. | Four S4 entry tests pass. |
| S4-5 | Remove/hide traditional invite UI entry points and clarify copy. | No UI route encourages new traditional invite use. |
| S4-6 | Deploy X2-A comment seal. | DB comments and migration banner are present. |
| S4-7 | Disable legacy mutation handlers and reward writes. | Admin mutation endpoints return `410 Gone`; payment flows do not call traditional reward writer. |
| S4-8 | Deploy X2-B hard freeze. | DB rejects `users` legacy field writes and legacy invite table writes. |
| S4-9 | Optional later X3 archive/drop, only after dry-run checks and user approval. | Admin/history can explain disputes from archive data without live dropped fields. |

## 8. Risks and Rollback

| Risk | Mitigation | Rollback / stop point |
|---|---|---|
| Breaking external HTTP contract | S4 first updates contract delta and tests; read-only endpoints keep stable success responses where chosen. | Stop before S4-3 if user rejects contract delta. |
| Accidentally disabling affiliate | Affiliate routes are explicitly `keep`; S4 entry tests protect aff signup and paid accrual. | Stop before S4-7 if affiliate tests fail. |
| Legacy history becomes unreadable | X2 keeps schema; X3 requires archive read paths before drops. | Stop before X3 drop; X2 rollback is trigger/comment removal. |
| Direct API callers still mutate legacy invite | Mutation endpoints get `410 Gone`; DB hard freeze rejects writes. | Stop before S4-8 if legitimate legacy writes are still observed. |
| Future agent reopens feature | Baseline CI seal, feature flag ban, code/DB/doc banners. | New user-approved decision required to reopen. |
| Irreversible data loss | X1/X3 drops are marked irreversible without backup/archive. | User can halt at every step before column/table drop. |

Irreversible points are: dropping live columns, dropping live legacy invite tables, deleting service/DTO code instead of deprecating it, and migrating/relabeling historical rewards into affiliate ledger. Those require separate explicit user approval.

## 9. T006 Contract Delta Recommendations

Do not edit `docs/superpowers/contracts/API-CONTRACT.md` in this T014 task. The following changes must be made in the S4 implementation task before code behavior changes:

| T006 area | Required delta in S4 |
|---|---|
| `GET /api/v1/user/aff` and subroutes | Keep active contract unchanged; clarify this is affiliate, not traditional invite. |
| `/api/v1/admin/affiliates/*` | Keep active contract unchanged; clarify this is affiliate, not traditional invite. |
| `GET /api/v1/invite/summary` | Change from active/generating semantics to read-only legacy history. Explicitly forbid ensuring/generating `users.invite_code`. |
| `GET /api/v1/invite/rewards` | Mark read-only legacy history; response schema may keep historical fields but no write/recompute side effects. |
| `GET /api/v1/admin/invites/stats` | Mark read-only archive/history. |
| `GET /api/v1/admin/invites/relationships` | Mark read-only archive/history. |
| `GET /api/v1/admin/invites/rewards` | Mark read-only archive/history. |
| `GET /api/v1/admin/invites/actions` | Mark read-only archive/history. |
| `POST /api/v1/admin/invites/rebind` | Change active behavioral contract to `410 Gone` deprecation response unless user re-approves. |
| `POST /api/v1/admin/invites/manual-grants` | Change active behavioral contract to `410 Gone` deprecation response unless user re-approves. |
| `POST /api/v1/admin/invites/recompute/execute` | Change active behavioral contract to `410 Gone` deprecation response unless user re-approves. |
| Contract tests | Add golden tests that `GET /api/v1/invite/summary` is read-only and admin legacy mutation endpoints return the stable deprecation body. |
