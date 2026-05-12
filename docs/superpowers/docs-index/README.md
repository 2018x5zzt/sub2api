# Superpowers Documentation Index

This index is the routing map for `docs/superpowers/`. If you add a new document, first read [`CONVENTIONS.md`](./CONVENTIONS.md) to decide the directory, filename, status line, and index entry.

Scope: this file indexes the active governance directories under `docs/superpowers/`: `audits/`, `contracts/`, `plans/`, `decisions/`, and this `docs-index/` area. Historical `specs/` files remain outside this active map until a follow-up sweep explicitly migrates or archives them.

## Audits

Use `audits/` for read-only inventories, parity matrices, gap analyses, and evidence collection. An audit should describe observed state and risk; it should not authorize implementation by itself. Typical lifecycle: `active` while feeding a plan, `superseded` after a newer audit replaces the evidence base, `archived` after the relevant implementation and verification are complete.

| Title | Document | Summary | Landed | Task | Status |
|---|---|---|---:|---|---|
| frontend-v2 功能 / 路由 / API 对齐矩阵 | [`audits/2026-05-12-frontend-v2-parity-matrix.md`](../audits/2026-05-12-frontend-v2-parity-matrix.md) | Maps old xlabapi Vue routes, frontend-v2 routes, API parity, missing entries, and P0/P1 migration risk. | 2026-05-12 | T001 | active |
| frontend-v2 Visual Style Gap Audit | [`audits/2026-05-12-frontend-v2-style-gap.md`](../audits/2026-05-12-frontend-v2-style-gap.md) | Audits Landing vs Console/Dashboard visual token usage, component duplication, layout gaps, and recommended style-unification path. | 2026-05-12 | T002 | active |
| frontend-v2 Parity Placeholder Inventory | [`audits/2026-05-12-placeholder-inventory.md`](../audits/2026-05-12-placeholder-inventory.md) | Lists all frontend-v2 `ParityPlaceholder` call sites and their route/API/action props for placeholder-reduction work. | 2026-05-12 | T002 | active |
| Semantic Dual Tracks Audit | [`audits/2026-05-12-semantic-dual-tracks.md`](../audits/2026-05-12-semantic-dual-tracks.md) | Documents backend semantic dual tracks for subscription, invite/affiliate, and multiplier concepts with code evidence. | 2026-05-12 | T003 | active |
| Test Coverage Inventory + P0 Contract Test Gaps | [`audits/2026-05-12-test-coverage-inventory.md`](../audits/2026-05-12-test-coverage-inventory.md) | Inventories existing backend/frontend tests and identifies P0 HTTP contract-test gaps before refactoring. | 2026-05-12 | T004 | active |
| Landing Style Direction (基于截图 + 新 logo) | [`audits/2026-05-12-landing-style-direction.md`](../audits/2026-05-12-landing-style-direction.md) | 基于用户三张截图 + 新 logo 反推 Landing 风格方向；含 Hero 字号路径决策、装饰强度量化清单、logo 三轨合一 S3 落地拆分。 | 2026-05-12 | T011 | active |

## Contracts

Use `contracts/` for durable interface facts: HTTP paths, request/response shapes, status/envelope behavior, contract-test inventories, and compatibility guarantees. A contract document is the tie-breaker for implementation when plans or audits disagree about API behavior. Typical lifecycle: `active` while governing current code, `superseded` only by a newer contract with the same scope, `archived` only after the covered interface is intentionally removed.

| Title | Document | Summary | Landed | Task | Status |
|---|---|---|---:|---|---|
| API Contract | [`contracts/API-CONTRACT.md`](../contracts/API-CONTRACT.md) | Canonical P0 API contract covering auth, user, keys, usage, subscriptions, redeem, affiliate, payment, admin, setup, OAuth, gateway, and compatibility aliases. | 2026-05-12 | T006 | active |
| P0 Contract Test Plan | [`contracts/CONTRACT-TEST-PLAN.md`](../contracts/CONTRACT-TEST-PLAN.md) | Converts the P0 contract gaps into concrete Go/frontend test strategy, golden-file rules, and named contract test cases. | 2026-05-12 | T007 | active |
| Dynamic Multiplier Isolation Test Plan | [`contracts/DYNAMIC-MULTIPLIER-TEST-PLAN.md`](../contracts/DYNAMIC-MULTIPLIER-TEST-PLAN.md) | 将 T018 §6 的 12 条断言落到可执行的后端隔离测试规格清单：测试名 / fixture / 预期 / 优先级 / 与现有测试关系 / 三批落地优先级。 | 2026-05-12 | T019 | draft |

## Plans

Use `plans/` for staged implementation routes, sequencing, ownership boundaries, verification gates, and rollback shape. Plans are not the source of truth for existing API contracts; they consume audits/contracts and turn them into work packages. Typical lifecycle: `active` while implementation is pending or running, `superseded` when a newer plan explicitly replaces it, `archived` after all planned work is complete or abandoned.

| Title | Document | Summary | Landed | Task | Status |
|---|---|---|---:|---|---|
| Product Subscription Restoration Implementation Plan | [`plans/2026-05-02-product-subscription-restoration.md`](../plans/2026-05-02-product-subscription-restoration.md) | Implementation route for restoring user product subscription APIs, UI cards, API-key visibility, and runtime verification. | 2026-05-02 | N/A | archived |
| xlabapi Miku Iframe SSO Implementation Plan | [`plans/2026-05-03-xlabapi-miku-iframe-sso.md`](../plans/2026-05-03-xlabapi-miku-iframe-sso.md) | Multi-repo implementation plan for xlabapi-issued Miku iframe SSO, backend binding, frontend callback, embed entry, and deployment checks. | 2026-05-03 | N/A | archived |
| Product Subscription Closure Implementation Plan | [`plans/2026-05-04-product-subscription-closure.md`](../plans/2026-05-04-product-subscription-closure.md) | Closure plan for product-subscription migrations, redeem enforcement, product binding, quota settlement, admin operations, and verification. | 2026-05-04 | N/A | archived |
| User Subscription Balance Fallback UI Implementation Plan | [`plans/2026-05-06-user-subscription-balance-fallback-ui.md`](../plans/2026-05-06-user-subscription-balance-fallback-ui.md) | Plan for explicit-save UI behavior around user subscription balance fallback controls and final verification. | 2026-05-06 | N/A | archived |
| Frontend V2 XlabAPI Parity Implementation Plan | [`plans/2026-05-11-frontend-v2-xlabapi-parity.md`](../plans/2026-05-11-frontend-v2-xlabapi-parity.md) | Earlier frontend-v2 parity implementation plan for API path fixes, route/navigation parity, minimal API clients, and test deployment. | 2026-05-11 | N/A | superseded |
| frontend-v2 系统性整改总规划 v1 | [`plans/2026-05-12-frontend-v2-systemic-alignment.md`](../plans/2026-05-12-frontend-v2-systemic-alignment.md) | Current synthesized master plan for frontend-v2 systemic alignment across functional parity, visual unity, semantic cleanup, and test gates. | 2026-05-12 | T009 | active |

## Decisions

Use `decisions/` for explicit choices that close ambiguity: selected direction, rejected alternatives, compatibility or deprecation boundaries, sealing rules, rollback limits, and user-review status. A decision should be short enough to answer "what did we choose and why?" without rereading all audits. Typical lifecycle: `draft` or `active` while awaiting/after approval, `superseded` if a later decision changes it, `archived` only when the choice no longer affects active work.

| Title | Document | Summary | Landed | Task | Status |
|---|---|---|---:|---|---|
| Decision Draft: Deprecate Traditional Invite | [`decisions/2026-05-12-deprecate-traditional-invite.md`](../decisions/2026-05-12-deprecate-traditional-invite.md) | Draft decision to deprecate traditional `users.invite_code` invite growth while preserving affiliate rebate semantics and sealing re-enable paths. Superseded by v2. | 2026-05-12 | T011 | superseded |
| Decision: Deprecate Traditional Invite (v2) | [`decisions/2026-05-12-deprecate-traditional-invite-v2.md`](../decisions/2026-05-12-deprecate-traditional-invite-v2.md) | v2 修订版，含 T013 三 P0 + 三 P1 修订：S4 endpoint 矩阵、X2 双阶段冻结、CI baseline 封印、affiliate 连续性 gate、X3 完整归档。 | 2026-05-12 | T014 | active |
| Decision: Dynamic Multiplier Scheduling | [`decisions/2026-05-12-dynamic-multiplier-scheduling.md`](../decisions/2026-05-12-dynamic-multiplier-scheduling.md) | 固化动态分组 + 预算倍率 + 账号级倍率乘算 + 7 天均费兜底的四规则；含术语对齐 T003、25+ 处 file:line 代码证据、伪代码 + 12 条边界用例、S4 保护机制四件套（banner/测试/CI grep/protected files）。 | 2026-05-12 | T018 | active |

## Reviews

Use `reviews/` for explicit review findings on audits, contracts, plans, or decisions. Reviews should say what was checked, what failed or passed, and whether a newer review supersedes the result. Typical lifecycle: `active` while it is the latest review for the target document, `superseded by <newer review>` when a follow-up review replaces it, `archived` after the reviewed work is no longer active.

| Title | Document | Summary | Landed | Task | Status |
|---|---|---|---:|---|---|
| Review: Deprecate Traditional Invite Decision | [`reviews/2026-05-12-review-deprecate-invite.md`](../reviews/2026-05-12-review-deprecate-invite.md) | Initial review of the traditional invite deprecation decision; superseded by v2 review. | 2026-05-12 | T013 | superseded by v2 review |
| Review: Deprecate Traditional Invite Decision v2 | [`reviews/2026-05-12-review-deprecate-invite-v2.md`](../reviews/2026-05-12-review-deprecate-invite-v2.md) | Final v2 review confirming the T013 P0/P1 findings were addressed, with one S4 contract-delta recommendation. | 2026-05-12 | T015 | active |

## Docs Index

Use `docs-index/` only for governance of the documentation set itself: the map you are reading and the rules for future documents.

| Title | Document | Summary | Landed | Task | Status |
|---|---|---|---:|---|---|
| Superpowers Documentation Index | [`docs-index/README.md`](./README.md) | Directory-level map for audits, contracts, plans, decisions, and docs-index governance documents. | 2026-05-12 | T005 | active |
| Superpowers Documentation Conventions | [`docs-index/CONVENTIONS.md`](./CONVENTIONS.md) | Naming, placement, frontmatter, lifecycle, single-source-of-truth, and task-linking conventions for future docs. | 2026-05-12 | T005 | active |

## Current Tie-Breakers

- API method/path/status/envelope/field facts: `contracts/API-CONTRACT.md` wins over audits and plans.
- Contract-test scope and ordering: `contracts/CONTRACT-TEST-PLAN.md` wins over older test notes in audits.
- Frontend-v2 implementation sequencing: `plans/2026-05-12-frontend-v2-systemic-alignment.md` wins over `plans/2026-05-11-frontend-v2-xlabapi-parity.md` where they differ.
- Evidence for current gaps: the 2026-05-12 audit set is the active evidence base until replaced by newer audits.
- Traditional invite deprecation: `decisions/2026-05-12-deprecate-traditional-invite-v2.md` is v2 active; S4 实施仍需独立任务批准.
- 动态倍率调度：`decisions/2026-05-12-dynamic-multiplier-scheduling.md` is the single source of truth; `contracts/DYNAMIC-MULTIPLIER-TEST-PLAN.md` is the test counterpart. 代码与之冲突应改回代码，除非新决策替换。
