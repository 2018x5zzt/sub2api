# Review: Deprecate Traditional Invite Decision v2

- Date: 2026-05-12
- Reviewer: 审查者
- Input decision: `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite-v2.md`
- Prior review: `docs/superpowers/reviews/2026-05-12-review-deprecate-invite.md`
- Scope: verify whether the six T013 findings are substantively addressed. No decision/code/API contract edits were made.

## Conclusion

结论：通过，带 1 条 S4 contract-delta 修订建议。

v2 实质解决 T013 的三 P0 + 三 P1：endpoint matrix、X2 hard freeze、CI baseline seal、feature-flag ban、affiliate continuity gates、X3 archive completeness都已补齐。可以把 v2 作为 active decision 进入后续文档治理和 S4 scope 拆分。

唯一需要保留到 S4 的修订点：v2 对 admin legacy mutation 的 `410 Gone` body 写成 raw `{ "error": "legacy_invite_deprecated", "message": ... }`。现有 `/api/v1/*` 契约总则使用标准错误 envelope `{ "code", "message", "reason?", "metadata?" }`。S4 修改 `API-CONTRACT.md` 和实现时，应把 410 body 对齐为标准 envelope，例如 `{"code": <nonzero>, "message": "Traditional invite mutations are retired", "reason": "legacy_invite_deprecated"}`，或在 contract delta 明确这是一个允许的 raw exception。建议前者。

该点不阻断 v2 active，因为 v2 §9 已明确 `API-CONTRACT.md` 留到 S4 修改；但它应成为 S4 contract delta 的验收项。

## Checklist

| T013 item | v2 status | Evidence | Residual action |
|---|---|---|---|
| P0-1 HTTP compatibility matrix | Pass | v2 §2 gives each listed endpoint exactly one behavior: affiliate routes `keep`, user/admin legacy GETs `read-only`, admin mutations `deprecation-response`; mutation response is `410 Gone`. | In S4, align 410 body with standard `/api/v1` error envelope or document raw exception. |
| P0-2 X2 database freeze | Pass | v2 §3 splits X2-A comment seal and X2-B hard freeze; X2-B rejects `users` INSERT with legacy fields, rejects UPDATE changes, and adds read-only triggers for `invite_code_aliases`, `invite_relationship_events`, `invite_reward_records`, `invite_admin_actions`. | None for decision. Implementation must adapt SQL to actual DB dialect. |
| P0-3 CI static seal baseline | Pass | v2 §5.3 forbids whole-file allowlists, defines `docs/superpowers/seals/legacy-invite-baseline.txt`, uses `comm -13`, includes routes/wire files, and requires a negative self-test fixture. | None for decision. Implementation should make normalizer deterministic. |
| P1-4 feature flag ban | Pass | v2 §5.4 bans new settings/config/env/frontend flags and explicitly includes `invite_enabled`, `traditional_invite_enabled`, `legacy_invite_enabled`, `enable_traditional_invite`, plus misuse of `invitation_code_enabled`; CI scan includes these names. | None. |
| P1-5 affiliate continuity | Pass | v2 §6 lists four S4 gates: aff_code signup binds affiliate without legacy writes, paid order accrues affiliate with `affiliate_enabled=true`, S4 payment does not call `ApplyBaseRechargeRewards`, and legacy+affiliate users are not double-paid. It also handles `affiliate_enabled=false` by requiring product acceptance or blocking S4. | None. |
| P1-6 X3 archive completeness | Pass | v2 §3 X3 archives all four legacy tables plus user fields, defines row-count/null/duplicate/alias/reward-action-event/sample-chain dry-run checks, and requires admin/history read paths to explain disputes without dropped live fields. | None. |

## Extra Checks

### T006 Contract Delta

v2 §9 is directionally correct and executable: keep affiliate contracts unchanged, convert legacy invite GETs to read-only, convert admin mutation POSTs to `410 Gone`, and add golden tests. This sequencing avoids changing behavior before the contract is updated.

Required S4 addition: specify the exact 410 body using the standard `/api/v1` error envelope, or explicitly record a raw-response exception. Without this, implementers may accidentally introduce a new envelope inconsistency while trying to avoid the old P0 drift.

### v1 Superseded Governance

v1 not being marked superseded is a documentation-governance defect, not a v2 blocker. It should be handled by the document organizer after this pass:

- Mark `docs/superpowers/decisions/2026-05-12-deprecate-traditional-invite.md` as superseded by v2.
- Mark v2 as the active decision in whatever index/control-plane list is used.
- Avoid deleting v1 because T013 links to it and v2 is a revision trace.

## Final Decision

- v2 active status: pass.
- S4 implementation: still requires normal user-approved task scope, contract update, tests, and staged migration work as v2 says.
- Blocking findings: none.
- Non-blocking S4 action: align `410 Gone` body with standard `/api/v1` error envelope in `API-CONTRACT.md` and implementation.
