# Upgrade v0.1.106 Incremental Absorption Design

## Goal

Absorb the effective changes from `upgrade-v0.1.106-merge` into `xlabapi` without doing a single large merge, while preserving the current `xlabapi` business semantics that have already been validated for invite, redeem, and legacy invitation-code retirement.

## Current State

`xlabapi` is now the only active mainline branch for this fork. The current branch already includes:

- invite-system consolidation and retirement of legacy one-time registration invitation codes
- bilateral invite reward settlement for commercial balance redeem codes
- frontend/backend/admin alignment for the current invite flow
- several smaller branch absorptions that were already merged into `xlabapi`
  - `enterprise-visible-groups`
  - `dynamic-group-budget-multiplier`
  - `smtp-outlook-starttls-20260412-151123`
  - `main` and the equivalent `release/20260409-redpacket-runtime-base`

`upgrade-v0.1.106-merge` is still unmerged and remains the only active non-archive branch with substantial unabsorbed work. It is large enough that a one-shot merge is not acceptable: it contains upstream `v0.1.106`, local customizations, and cross-cutting edits spanning generated ent code, gateway behavior, promo/redeem flows, pricing, admin UI, and user-facing frontend flows.

## Problem Statement

Directly merging `upgrade-v0.1.106-merge` into `xlabapi` creates a conflict surface that is too broad to reason about safely in one pass. The earlier dry-run already showed conflicts across:

- ent generated/runtime files
- gateway and OAuth transformation paths
- pricing and billing logic
- promo/redeem handlers and services
- frontend redeem and model-hub views

The main risk is not just technical merge difficulty. The larger risk is silently regressing the current validated `xlabapi` product behavior, especially:

- the invite system that replaced legacy registration invitation codes
- the current invite reward semantics and frontend copy
- the current redeem entry behavior and current admin/public setting contracts

The absorption strategy must therefore optimize for controlled risk exposure rather than raw merge speed.

## Non-Goals

This work does not aim to:

- preserve `upgrade-v0.1.106-merge` as an independently releasable long-lived branch
- merge `archive/*` branches into `xlabapi`
- re-open product decisions that were already locked on `xlabapi`
- redesign invite, redeem, or promo semantics unless a concrete incompatibility makes a minimal compatibility adjustment necessary
- force generated code cleanliness at every intermediate step if the batch-level behavior and tests are already correct

## Design Principles

### 1. `xlabapi` semantics win on business-critical conflicts

For invite, redeem, promo, and settings behavior, current `xlabapi` semantics are the base truth. When a conflict appears between the current mainline and `upgrade-v0.1.106-merge`, the merge resolution must start from the current `xlabapi` behavior and only layer in the still-useful incremental capability from the upgrade branch.

### 2. Incremental absorption over one-shot merge

The branch will be absorbed in multiple batches. Each batch must:

- have a narrow feature boundary
- be mergeable and testable on its own
- end with a commit and push to `origin/xlabapi`
- leave the branch deployable before the next batch starts

### 3. Generated code is allowed to lag temporarily, but only within a batch

`ent`, `wire`, and API contract/generated artifacts are not the primary product goal of the first four batches. They may be touched as compatibility work during those batches, but a dedicated final cleanup batch is required to reconcile generated and snapshot surfaces comprehensively.

### 4. Stop at the batch boundary on failure

If a batch cannot be validated cleanly, the process stops at that batch. Later batches are not allowed to proceed while earlier batches are unresolved.

## Recommended Absorption Architecture

The branch will be absorbed in five batches.

### Batch 1: OpenAI / Gateway / OAuth / Messages / Codex

This batch absorbs the highest-value and most independent gateway-facing work first.

Expected contents:

- OpenAI compatibility fixes
- messages path normalization fixes
- codex / compat transformation updates
- OAuth request/response handling improvements
- `prompt_cache_key` related compat behavior
- `gpt-5.4-xhigh` and related model normalization fixes
- `429` silent failover and adjacent gateway-path resilience improvements
- `file_upload` OAuth scope support
- `X-Claude-Code-Session-Id` compatibility work

This batch must explicitly avoid changing invite, promo, redeem, or current frontend invite semantics.

### Batch 2: Antigravity / 429 / Internal-500 Penalty / Billing-Pricing

This batch absorbs the retry, failover, and pricing logic that is still mostly backend-internal and less likely to collide with the current user-facing invite and redeem product decisions.

Expected contents:

- Antigravity internal `500` progressive penalty
- privacy/tier/plan-type related Antigravity or OAuth improvements
- billing changes that intentionally preserve the user-requested model for charging
- rate-limit persistence and retry behavior related to gateway account switching
- effective model pricing logic that belongs to service-side pricing decisions

This batch must still avoid changing the current promo/redeem/invite user flow.

### Batch 3: Admin Account Management / TLS Profile / Bulk Ops / Keys-ModelHub Frontend

This batch absorbs the admin tooling and frontend management layer changes that are substantial but operationally separable.

Expected contents:

- TLS fingerprint profile management
- account bulk operations and account-management improvements
- keys UI enhancements and related API surface changes
- model hub display/pricing presentation changes
- backend/admin DTO and handler changes needed to support these screens

This batch may touch frontend screens broadly, but it still must not redefine the current redeem or invite semantics.

### Batch 4: Promo / Redeem / Benefit Code / Lucky Leaderboard

This is the most sensitive batch and must remain isolated.

Expected contents:

- promo and benefit-code improvements from the upgrade branch
- redeem flow improvements that are still compatible with current `xlabapi`
- benefit-code leaderboard behavior and related UI/handler/service updates

Conflict rule:

- current `xlabapi` invite and legacy invitation-code retirement behavior stays intact
- current `xlabapi` register/invite-link behavior stays intact
- current `xlabapi` bilateral invite-reward semantics stay intact
- only non-conflicting redeem/promo enhancements are absorbed

### Batch 5: Generated-Code / Ent / Wire / Contract / Migration Cleanup

This batch is technical reconciliation, not new product behavior.

Expected contents:

- `ent` generated/runtime cleanup
- `wire` regeneration and DI cleanup
- API contract snapshot updates
- migration or schema cleanup needed to align the combined result
- remaining generated-file drift removal

The success criterion for this batch is a stable and internally coherent mainline after the previous feature batches are already absorbed.

## Merge Rules

### Branch Handling

- `xlabapi` remains the target branch for every batch
- `upgrade-v0.1.106-merge` is never merged as a single final one-shot merge
- `archive/main-before-runtime-redpacket-20260409` and `archive/origin-main-before-runtime-redpacket-20260409` are never absorbed

### Commit and Push Discipline

Each batch ends with:

1. one or more merge/conflict-resolution commits scoped to that batch
2. a final follow-up fix commit if tests or snapshots require it
3. push to `origin/xlabapi`

The branch must not accumulate multiple unresolved batches locally before pushing.

### Conflict Priority

When a file conflict occurs, resolve in this order:

1. preserve current `xlabapi` product semantics
2. preserve the upgrade batch's targeted capability
3. preserve testability and type/API consistency
4. defer broad generated cleanup to Batch 5 when possible

## Verification Strategy

Every batch requires three layers of verification.

### Layer 1: Batch-Local Focused Verification

Run only the smallest test slice that proves the batch's intended behavior. Examples include:

- gateway/OAuth unit tests for Batch 1
- billing/pricing unit tests for Batch 2
- admin/frontend targeted tests for Batch 3
- redeem/promo targeted tests for Batch 4

### Layer 2: Mainline Regression Verification

After the focused tests pass, run the relevant backend unit suites that cover the touched subsystem and any current `xlabapi` behavior that could have been affected.

### Layer 3: Frontend and Build Verification

When frontend files are touched, run:

- the targeted frontend tests for the affected views/composables
- `npm run build`

The build may still emit existing non-blocking chunk warnings, but no new build failures are allowed.

## Deployment Boundary

The deployment-side override `GATEWAY_UPSTREAM_RESPONSE_READ_MAX_BYTES=16777216` is a deployment-repository concern, not a source-code branch-absorption concern. It is already active in the running container configuration, but it does not count as a code-branch absorption step inside `sub2api-src`.

This means the upgrade absorption process should treat deployment override changes separately from source-branch integration work.

## Success Criteria

The absorption effort is complete only when all of the following are true:

- the effective capabilities from `upgrade-v0.1.106-merge` have been absorbed into `xlabapi` in batch form
- current `xlabapi` invite, redeem, and legacy-invitation retirement semantics remain intact unless an explicitly approved compatibility change supersedes them
- `origin/xlabapi` has been updated after every completed batch
- all remaining unmerged branches are archive-only branches
- the final `xlabapi` state passes the agreed backend, frontend, and build verification gates

