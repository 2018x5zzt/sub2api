# Review: Deprecate Traditional Invite Decision

- Date: 2026-05-12
- Reviewer: 审查者
- Input decision: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md`
- Related inputs: `docs/superpowers/audits/2026-05-12-semantic-dual-tracks.md`, `docs/superpowers/contracts/API-CONTRACT.md`, `docs/superpowers/plans/2026-05-12-frontend-v2-systemic-alignment.md`

## Conclusion

结论：需修订后再进入 S4 实施。

方向本身成立：弃用 `users.invite_code` 传统 invite、保留 `user_affiliates.aff_code` affiliate，是和 T003/T009 一致的；X2 first、X3 later 的数据库路线也比直接 drop 更稳。

但当前决策文档还不能直接作为执行封印，因为三个点会让后续 agent 或 API 调用绕过封印：

1. CI/静态封印不是可达的强约束，allowlist 过宽且没有 baseline/diff 机制。
2. X2 数据库封印只覆盖 `users` 字段的 UPDATE 示例，没有覆盖 INSERT 和 legacy invite 相关表写入。
3. `/api/v1/invite/*` 与 `/api/v1/admin/invites/*` 的 HTTP 兼容策略没有钉死，现有 T006 仍把其中多个传统 invite 端点定义为 active contract。

建议判定：不阻断产品方向；阻断 S4 实施授权，直到下方 P0/P1 修订动作被写回决策或拆成前置任务。

## Findings

### P0: HTTP contract strategy for deprecated invite endpoints is underspecified

Evidence:

- Decision marks user/admin traditional invite APIs in deprecation scope: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:30` and `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:31`.
- Decision says handlers may become legacy/read-only or disabled: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:315` and `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:316`.
- T006 currently defines `GET /api/v1/invite/summary` as a contract that may ensure/generate a traditional invite code: `docs/superpowers/contracts/API-CONTRACT.md:166`.
- T006 currently defines admin invite mutation endpoints as behavioral contracts with write side effects: `docs/superpowers/contracts/API-CONTRACT.md:179`.

Risk:

If S4 hides UI and adds comments but leaves routes active, direct API callers or future agents can still revive traditional invite behavior. If S4 disables or changes these endpoints without amending T006, it violates the P0 external HTTP contract rule in T009.

Required revision:

- Add an explicit endpoint matrix for S4:
  - `/api/v1/user/aff` and `/api/v1/admin/affiliates/*`: unchanged, kept.
  - `GET /api/v1/invite/summary`: choose one exact behavior: keep 200 read-only without generating `invite_code`, or return a documented deprecation response. Do not leave “read-only or disabled” ambiguous.
  - `GET /api/v1/invite/rewards`: choose read-only history or deprecation response.
  - `GET /api/v1/admin/invites/{stats,relationships,rewards,actions}`: choose read-only archive behavior or deprecation response.
  - `POST /api/v1/admin/invites/{rebind,manual-grants,recompute/execute}`: must be disabled/gated by default after S4 freeze unless the user explicitly reopens them.
- Update `docs/superpowers/contracts/API-CONTRACT.md` in the same S4 task, or create a contract delta doc before implementation.
- Add contract tests that assert the chosen post-S4 behavior, especially that `GET /invite/summary` cannot create/write `users.invite_code`.

### P0: X2 database freeze does not actually freeze all legacy invite writes

Evidence:

- X2 comments preserve fields and say writes are blocked: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:108` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:126`.
- The hard trigger example only covers `UPDATE OF invite_code, invited_by_user_id, invite_bound_at ON users`: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:333` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:350`.
- Deprecated tables include `invite_code_aliases`, `invite_relationship_events`, `invite_reward_records`, and `invite_admin_actions`: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:18` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:21`.

Risk:

A future code path can still insert a new user with `invite_code` already populated, or insert/update legacy reward/action/alias rows. That bypasses the intended DB seal even if the user-column UPDATE trigger exists.

Required revision:

- For `users`, specify both INSERT and UPDATE behavior:
  - On INSERT, reject non-null `invite_code`, `invited_by_user_id`, or `invite_bound_at` after freeze.
  - On UPDATE, reject any change to those fields.
- Add read-only triggers or equivalent enforcement for `invite_code_aliases`, `invite_relationship_events`, `invite_reward_records`, and `invite_admin_actions` if X2 means those tables are retained only for history.
- If write triggers are too risky for first S4 batch, explicitly downgrade X2 to “comment-only visibility seal” and make the actual write-freeze a separate required exit criterion before hiding/removing app code.

### P0: CI/static seal is too broad to stop future accidental re-enablements

Evidence:

- Decision proposes a grep guard: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:360` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:385`.
- The guard allowlists whole files including `backend/internal/service/invite_service.go`, `backend/internal/handler/invite_handler.go`, and `backend/internal/handler/admin/invite_handler.go`: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:377` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:379`.

Risk:

Whole-file allowlists are exactly where a future agent would edit to revive traditional invite behavior. A new call inside `invite_service.go` or a new mutation branch inside `admin/invite_handler.go` would be invisible to the guard.

Required revision:

- Replace broad `-g '!file.go'` allowlists with a checked-in baseline file or a script that filters exact known lines/functions.
- Fail on `rg` errors separately; do not collapse every non-zero result into success.
- Include route registration and wire files in the guard:
  - `backend/internal/server/routes/user.go`
  - `backend/internal/server/routes/admin.go`
  - `backend/internal/service/wire.go`
  - `backend/cmd/server/wire_gen.go` if generated references remain
- Add a negative test fixture: intentionally add a fake `InviteService` call outside the baseline in CI script unit tests, and prove the script fails.

### P1: Feature-flag boundary is mostly safe, but the decision should explicitly ban a traditional invite enable flag

Evidence:

- Current settings expose `invitation_code_enabled` for registration gate and `affiliate_enabled` for affiliate; T003 says these are separate concepts.
- Decision keeps registration invitation out of scope: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:59`.
- Decision keeps affiliate: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:38` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:49`.

Risk:

I did not find a current “enable traditional invite” flag in the reviewed docs/code references. But without an explicit ban, a later agent may add `invite_enabled` or reuse `invitation_code_enabled` as a mistaken resurrection switch.

Required revision:

- Add a rule: no new setting/feature flag may enable `InviteService`, `users.invite_code`, `/api/v1/invite`, or `/api/v1/admin/invites` without a new user-approved decision.
- Add static checks for new setting names such as `invite_enabled`, `traditional_invite_enabled`, and route/UI labels that imply traditional invite is active.

### P1: Affiliate reward-chain continuity needs an explicit S4 gate

Evidence:

- Decision states affiliate remains the future system: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:42` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:48`.
- Decision states future traditional base rewards stop after S4 freeze and affiliate rebates remain active if affiliate is enabled: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:240`.
- T009 requires invite/affiliate conflict tests if both reward models do not collapse: `docs/superpowers/plans/2026-05-12-frontend-v2-systemic-alignment.md:78`.

Risk:

If `affiliate_enabled` is false in an environment, S4 freeze can intentionally result in no invite/rebate reward path. That may be acceptable, but it must be explicit. If not acceptable, S4 must assert affiliate is enabled and the payment fulfillment path still accrues rebate.

Required revision:

- Add S4 entry condition: before disabling traditional reward writes, verify one of these is true:
  - Product decision accepts “no invite/rebate reward when affiliate is disabled”; or
  - `affiliate_enabled` is enabled and tested in target environment.
- Add behavioral tests for:
  - signup with `aff_code` binds `user_affiliates.inviter_id` and does not write `users.invited_by_user_id`;
  - paid order/recharge accrues affiliate ledger/quota when affiliate is enabled;
  - paid order/recharge does not call `InviteService.ApplyBaseRechargeRewards` after freeze;
  - users with both legacy invite history and affiliate inviter are not double-paid.

### P1: X3 archive path is incomplete for readable history

Evidence:

- X3 archives only `users.invite_code`, `invited_by_user_id`, and `invite_bound_at` into `legacy_user_invites`: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:166` to `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:186`.
- Reward/action/alias table handling is left as “Either keep original ... or copy them”: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:183`.

Risk:

X3 cannot be considered a complete readable migration path until aliases, relationship events, reward records, and admin actions have a chosen archive/read-only strategy. Otherwise user binding history may survive while reward/audit history becomes split or lost.

Required revision:

- For X3, add a concrete archive table/list for every deprecated table, or explicitly choose “keep original invite_* tables read-only”.
- Add dry-run checks: row counts, null inviter counts, duplicate invite codes, alias coverage, reward/action/event counts, and a sample relationship trace from invitee to inviter to rewards.
- Make X3 exit criteria include “admin/history read path can still explain a legacy invite dispute without querying dropped live fields.”

### P2: UI terminology direction is correct but needs an implementation inventory

Evidence:

- Decision says marketing/UI “邀请码” should be renamed by meaning: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md:266`.
- T003 recommends exactly this split: registration invitation code, invite_code, affiliate code.
- T009 keeps the same boundary and says S4 should split registration gate, invite relation, and affiliate promotion code: `docs/superpowers/plans/2026-05-12-frontend-v2-systemic-alignment.md:161` to `docs/superpowers/plans/2026-05-12-frontend-v2-systemic-alignment.md:163`.

Risk:

Low relative to the API/DB seals, but visible UI copy can mislead future implementation. “Invite” in affiliate admin pages is still likely to be read as traditional invite unless the implementation task has a file-level inventory.

Required revision:

- Add a short S4 UI copy checklist listing frontend and frontend-v2 i18n/page files to rename.
- Require labels:
  - registration gate: “registration invitation / 注册准入码”;
  - affiliate: “affiliate code / 推广码 / 返利码”;
  - traditional invite: “legacy invite code / 历史邀请关系码” only in archive/admin contexts.

## Pass/Revise/Block Decision

- Product decision direction: pass.
- DB option recommendation X2 first: pass with revision.
- Five-layer seal as currently written: needs revision before implementation.
- S4 implementation authorization: block until P0 findings are resolved.

## Concrete Backend Actions

1. Add an S4 endpoint behavior matrix and update `API-CONTRACT.md` or a contract delta before touching routes.
2. Strengthen X2 DB freeze to cover INSERT/UPDATE on `users` fields and writes to legacy invite tables, or explicitly split comment-only vs hard-freeze phases.
3. Replace broad grep allowlists with an exact-baseline static seal script and include routes/wiring in the scan.
4. Add affiliate continuity tests before disabling traditional reward writes.
5. Complete X3 archive/read-only strategy for alias, relationship, reward, and admin-action tables before any drop/archive migration.
