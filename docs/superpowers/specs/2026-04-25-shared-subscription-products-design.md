# Shared Subscription Products Design

**Date:** 2026-04-25

**Goal:** Replace per-group subscription settlement with a reusable product-level shared subscription model so one purchased subscription product can expose multiple real groups, keep one API key bound to one real group, and debit one shared quota pool with per-group multipliers.

## Summary

The current subscription model is group-centric:

- a redeem code targets one `group_id`
- a user subscription is stored as `user_id + group_id`
- visibility, API key eligibility, and quota settlement all key off that one group

That model cannot support the product semantics now required:

- one purchased product such as `GPT 月卡`
- multiple real groups such as `plus/team 混池` and `pro 池`
- each real group keeps independent routing and API key binding
- all requests debit one shared product quota pool
- different real groups debit that pool with different multipliers such as `1.0x` and `1.5x`

This design introduces a new product-level subscription layer:

- `subscription_products` define the user-facing product and its shared quota window
- `subscription_product_groups` bind one product to many real groups with `debit_multiplier`
- `user_product_subscriptions` hold the user's real entitlement and shared usage state
- API keys remain bound to one real group
- runtime settlement resolves `group -> product -> user product subscription`

The first migration target is the existing GPT monthly subscription population. After migration, eligible users keep the same effective subscription, see one `GPT 月卡` product, and gain access to an additional real `pro` group that debits the same shared pool at `1.5x`.

The model is intentionally generic. It must later support products that span multiple GPT groups and multiple Claude groups in the same shared USD pool.

## Problem

The existing code uses groups as both:

- the real routing target
- the user-facing purchased package
- the quota ownership boundary

That coupling blocks the required product behavior.

Concrete examples:

- A subscription redeem code currently stores only one `group_id`, so one code can activate only one group.
- A user subscription is unique per `user_id + group_id`, so the system has no first-class concept of "one product exposing many groups".
- The frontend currently renders active subscriptions as a list of group subscriptions rather than one product with multiple usable groups.
- Runtime billing and eligibility are keyed by `user + group`, so a `plus/team` key and a `pro` key cannot share one authoritative pool.

The first operational request is intentionally narrow:

- existing GPT monthly users should keep their current subscription
- those users should newly see an eligible `pro` group
- `plus/team` usage should debit the shared GPT monthly pool at `1.0x`
- `pro` usage should debit the same shared GPT monthly pool at `1.5x`

The design must satisfy that first migration while avoiding a one-off hack. It needs to become the long-term product model for future cross-group and cross-platform products.

## Decision

Introduce a product-level shared subscription model and treat groups only as real routing targets.

Three approaches were considered:

### Option 1: patch the current per-group subscription model

Keep `user_subscriptions(user_id, group_id)` as the primary model and add more visibility rules plus a shared-pool side table.

Advantages:

- smaller initial schema diff
- can superficially preserve current query paths

Disadvantages:

- the primary table would still claim the entitlement belongs to one group even when settlement belongs to one product
- every read path would need exceptions to recover the true product semantics
- future cross-platform products would inherit the same ambiguity

### Option 2: extend `user_subscriptions` with `shared_pool_id`

Keep one row per group subscription but point multiple rows at one shared quota owner.

Advantages:

- easier short-term migration than a clean product model
- can reuse some current list and preload code

Disadvantages:

- still leaks group-centric semantics everywhere
- frontend and operational explanations remain confusing
- audit and rollback logic become more complex because one table mixes group entitlement and shared product ownership

### Option 3: add a first-class shared subscription product model

Introduce product entities, product-to-group bindings, and user product subscriptions. Keep real groups and API keys as they are.

Advantages:

- matches the target product language exactly
- cleanly separates routing from entitlement ownership
- scales to many groups and many platforms
- supports a user-facing "one product, many usable groups" frontend

Disadvantages:

- larger migration and API surface change
- requires careful rollout gating

### Chosen approach

Use **Option 3**.

This is the only approach that solves the first GPT migration cleanly and still leaves the codebase with a reusable product model for future GPT + Claude shared products.

## Goals

- Add a reusable user-facing subscription product model independent of any one group.
- Allow one subscription product to expose multiple real groups.
- Keep one API key bound to one real group.
- Debit one shared product quota pool using the selected real group's `debit_multiplier`.
- Support products that may later span multiple platforms.
- Migrate existing GPT monthly users into the new model without changing their effective entitlement window.
- Show migrated users one `GPT 月卡` product in the frontend instead of multiple independent group subscriptions.
- Preserve additive-only schema rollout and low-risk rollback.
- Produce enough documentation and rollout structure that work can continue safely after an interrupted session.

## Non-Goals

- Do not support one API key bound to multiple real groups at once.
- Do not silently reroute an existing API key to another group if its configured group is removed from a product.
- Do not support one real group belonging to multiple active subscription products in the first iteration.
- Do not delete or hard-migrate away legacy `user_subscriptions` during the first rollout.
- Do not backfill historical usage logs with product references for old traffic before cutover.
- Do not redesign unrelated balance-mode group behavior.

## Product Semantics

The product semantics for the new model are fixed as follows:

- A user purchases or redeems a **subscription product**, not a group.
- A subscription product has one shared quota pool whose accounting unit is **standard-cost USD**.
- A subscription product may expose multiple real groups.
- A request first resolves the API key's real bound group.
- The runtime then resolves which product exposes that group for that user.
- The system computes the request's standard USD cost and multiplies it by the product-group `debit_multiplier`.
- The multiplied amount is debited from the one shared product pool.

Example for the first rollout:

- Product: `GPT 月卡`
- Product group bindings:
  - `plus/team 混池` -> `1.0x`
  - `pro 池` -> `1.5x`
- If a request routed through `plus/team` has standard cost `$2`, the product pool is debited `$2`.
- If a request routed through `pro` has standard cost `$2`, the same product pool is debited `$3`.

This product-level pool unit remains USD even when products later span multiple platforms. Platform differences affect only the standard cost calculation before the multiplier is applied.

## Current-State Constraints

The design must coexist with the current implementation constraints:

- `groups` are the real routing units and remain the entity bound by API keys.
- `api_keys` currently store a single `group_id`; that stays unchanged.
- `redeem_codes` currently support `group_id` for subscription-type codes and must remain backward compatible during transition.
- `user_subscriptions` currently own window state such as daily, weekly, monthly usage and daily carryover.
- runtime eligibility and billing are currently keyed by `user + group`
- the frontend subscription store and UI currently assume active subscriptions are a flat list of group subscriptions

The new model must isolate change by introducing a parallel authoritative path for shared products instead of trying to coerce legacy tables into incompatible semantics.

## Domain Model

### Real Group

`groups` continue to mean:

- the real routing and account-pool target
- the real group the API key binds to
- the real group displayed on each API key record

For groups that belong to an active shared subscription product:

- they remain real groups
- they may keep `subscription_type = subscription` for compatibility
- their legacy `daily_limit_usd`, `weekly_limit_usd`, and `monthly_limit_usd` stop being the authoritative quota source
- their settlement source becomes the owning shared product

### Subscription Product

A subscription product is the user-facing entitlement that owns the shared quota pool.

Examples:

- `GPT 月卡`
- future `Claude + GPT 通用月卡`

A product defines:

- name and user-facing description
- status
- default validity days
- shared quota windows and limits
- ordering for display

### Product-Group Binding

A product-group binding connects one product to one real group and defines how requests through that group debit the shared pool.

The binding defines:

- owning `product_id`
- real `group_id`
- `debit_multiplier`
- status
- display order within the product

### User Product Subscription

This is the user's authoritative entitlement record in the new model.

It defines:

- which user owns which product
- validity window and status
- shared daily, weekly, monthly usage
- shared daily carryover state
- notes and assignment metadata

This record is the only quota owner in product mode.

## Data Model

### `subscription_products`

Create a new table with soft delete and standard timestamps.

Suggested fields:

- `id`
- `code`
- `name`
- `description`
- `status`
  - `draft`
  - `active`
  - `disabled`
- `default_validity_days`
- `daily_limit_usd`
- `weekly_limit_usd`
- `monthly_limit_usd`
- `sort_order`
- `created_at`
- `updated_at`
- `deleted_at`

Rules:

- `code` is stable and unique among non-deleted rows
- `draft` products may be configured and backfilled without affecting runtime
- only `active` products participate in visibility, API key eligibility, or billing

### `subscription_product_groups`

Create a new table with soft delete and standard timestamps.

Suggested fields:

- `id`
- `product_id`
- `group_id`
- `debit_multiplier`
- `status`
  - `active`
  - `inactive`
- `sort_order`
- `created_at`
- `updated_at`
- `deleted_at`

Rules:

- active uniqueness on `group_id` among non-deleted rows
  - one real group may belong to only one active product in this iteration
- active uniqueness on `product_id + group_id`
- `debit_multiplier` must be greater than `0`

### `user_product_subscriptions`

Create a new table with soft delete and standard timestamps.

Suggested fields:

- `id`
- `user_id`
- `product_id`
- `starts_at`
- `expires_at`
- `status`
  - `active`
  - `expired`
  - `suspended`
- `daily_window_start`
- `weekly_window_start`
- `monthly_window_start`
- `daily_usage_usd`
- `weekly_usage_usd`
- `monthly_usage_usd`
- `daily_carryover_in_usd`
- `daily_carryover_remaining_usd`
- `assigned_by`
- `assigned_at`
- `notes`
- `created_at`
- `updated_at`
- `deleted_at`

Rules:

- partial unique index on `user_id + product_id` where `deleted_at IS NULL`
- product-level windows and carryover semantics mirror the current user subscription model

### `product_subscription_migration_sources`

Create a migration audit table for traceability and rollback support.

Suggested fields:

- `id`
- `product_subscription_id`
- `legacy_user_subscription_id`
- `migration_batch`
- `legacy_group_id`
- `legacy_status`
- `legacy_starts_at`
- `legacy_expires_at`
- `legacy_daily_usage_usd`
- `legacy_weekly_usage_usd`
- `legacy_monthly_usage_usd`
- `created_at`

Rules:

- one legacy source row per migrated legacy subscription
- this table is append-only operational audit data

### `redeem_codes` extension

Extend `redeem_codes` with:

- `product_id` nullable foreign key

Rules:

- new shared-product subscription codes should use `product_id`
- legacy group subscription codes may continue to use `group_id`
- a subscription redeem code may target **exactly one** of:
  - `group_id`
  - `product_id`

### `usage_logs` extension

Extend `usage_logs` with:

- `product_id` nullable foreign key
- `product_subscription_id` nullable foreign key
- `group_debit_multiplier` decimal nullable
- `product_debit_cost` decimal nullable

Rules:

- legacy rows may leave these fields null
- all new shared-product subscription traffic must populate them
- `group_id` remains populated with the real bound group
- legacy `subscription_id` remains the legacy `user_subscriptions` foreign key only
- shared-product traffic must leave legacy `subscription_id` null and use `product_subscription_id` as the authoritative entitlement reference

## Legacy Coexistence Rules

The first rollout is additive-only.

Legacy tables remain:

- `user_subscriptions`
- `groups`
- `api_keys`
- legacy `redeem_codes.group_id`

The coexistence rules are strict:

- A real group can settle through **legacy group subscriptions** or **shared product subscriptions**, never both at the same time.
- Backfilled products may remain in `draft` status while legacy runtime stays authoritative.
- When a product becomes `active`, any real groups bound to that product are treated as product-settled groups and must no longer use legacy `user_subscriptions` for runtime settlement.
- Legacy rows are retained for audit and rollback but stop being the active settlement source for migrated groups after cutover.

## Runtime Design

### Group Visibility

User-visible subscription groups must no longer come from `user_subscriptions(user_id, group_id)` for migrated products.

New logic:

1. Load all active `user_product_subscriptions` for the user.
2. Expand each product to its active `subscription_product_groups`.
3. The expanded real groups become the visible subscription-eligible groups for API key creation.

Consequences:

- a migrated GPT monthly user automatically sees both `plus/team` and `pro`
- no extra legacy group subscription rows are needed just to expose `pro`

### API Key Binding

API keys remain single-group credentials.

All API key group-binding surfaces must use the same settlement-source rules, including:

- user API key create flows
- admin API key create flows
- admin API key rebind or change-group flows
- any batch import or sync path that writes `api_keys.group_id`

Binding validation changes:

- if the target group is a normal balance-mode group, preserve existing logic
- if the target group belongs to an active shared subscription product:
  - resolve the owning product
  - require an active `user_product_subscription`
  - do not require a legacy `user_subscription`

The group shown on the API key record remains the real bound group.

### Eligibility Check

The current runtime path checks active entitlement by `user + group`.

Product-mode runtime must instead:

1. resolve the API key's real `group_id`
2. determine whether the group belongs to an active shared subscription product
3. if yes, resolve the user's active `user_product_subscription`
4. validate status, expiry, and shared product quota windows
5. if no, fall back to the legacy or balance path

Authoritative limits in product mode come from `subscription_products`, not group-level limit columns.

### Shared Pool Debit

For product mode:

1. compute the request's standard USD cost using the existing pricing path before any shared-product debit multiplier
2. load the bound group's `debit_multiplier` from `subscription_product_groups`
3. calculate:

```text
product_debit_cost = standard_total_cost * debit_multiplier
```

4. increment the owning `user_product_subscription` windows by `product_debit_cost`

Examples:

- `plus/team` standard cost `$2` with multiplier `1.0` -> debit `$2`
- `pro` standard cost `$2` with multiplier `1.5` -> debit `$3`

### Billing Contract

The shared-product model must extend the **authoritative post-usage billing path**, not bypass it.

Required runtime contract:

- each product-settled request resolves one authoritative settlement context containing:
  - real `group_id`
  - owning `product_id`
  - owning `user_product_subscription.id`
  - `debit_multiplier`
- middleware and gateway code must treat that product settlement context as the one quota owner for the request
- product mode must not settle quota merely by writing `usage_logs`; the deduplicated post-usage billing apply path remains authoritative

Required billing-command semantics:

- the authoritative post-usage billing command or equivalent write path must carry:
  - `group_id`
  - `product_id`
  - `product_subscription_id`
  - `group_debit_multiplier`
  - `standard_total_cost`
  - `product_debit_cost`
- `standard_total_cost` means the request cost before the shared-product debit multiplier is applied
- `product_debit_cost` means `standard_total_cost * group_debit_multiplier`
- product quota windows are incremented by `product_debit_cost`, not by legacy group subscription usage and not by `usage_logs.actual_cost`

Coexistence with existing billing dimensions:

- in product mode, the shared product pool is the only subscription quota owner
- legacy `user_subscriptions` must not be incremented for a product-settled request
- legacy `usage_logs.subscription_id` must remain null for a product-settled request
- existing non-product sinks that remain keyed by actual billed request cost, such as API key quota or API key rate-limit usage, may continue to use the existing `actual_cost` semantics in the first rollout
- the implementation must therefore distinguish:
  - product pool debit amount: `product_debit_cost`
  - request billed amount for existing non-product sinks: current `actual_cost`

### Usage Log Semantics

Every shared-product request must record:

- `group_id` as the real bound group
- `product_id` as the owning shared product
- `product_subscription_id` as the user's authoritative entitlement row
- `group_debit_multiplier` as the applied multiplier
- `product_debit_cost` as the amount debited from the product pool

This preserves the answer to both questions:

- where did the request actually route
- which shared subscription product paid for it

### Caching

Current subscription read caches are keyed by `user + group`.

The new model requires:

- product subscription read cache keyed by `user + product`
- product-group binding cache keyed by `group`
- product visibility expansion cache keyed by `user`

Invalidation rules:

- any write to `user_product_subscriptions` invalidates `user + product` and the user's visibility cache
- any write to `subscription_product_groups` invalidates `group` binding cache and visibility caches for users who hold the affected product
- post-billing writes invalidate the same product subscription read state used by future eligibility checks

## API and Frontend Design

### User-Facing Subscription API

Keep legacy `/api/v1/subscriptions/*` endpoints for compatibility during transition.

Add a new product-oriented surface:

- `GET /api/v1/subscription-products/active`
- `GET /api/v1/subscription-products/summary`
- `GET /api/v1/subscription-products/progress`

Suggested response shape for active products:

```json
[
  {
    "product_id": 101,
    "code": "gpt_monthly",
    "name": "GPT 月卡",
    "description": "GPT monthly shared subscription",
    "status": "active",
    "starts_at": "2026-04-01T00:00:00Z",
    "expires_at": "2026-05-01T00:00:00Z",
    "daily_usage_usd": 0,
    "weekly_usage_usd": 0,
    "monthly_usage_usd": 42.5,
    "daily_limit_usd": null,
    "weekly_limit_usd": null,
    "monthly_limit_usd": 100,
    "daily_carryover_in_usd": 0,
    "daily_carryover_remaining_usd": 0,
    "groups": [
      {
        "group_id": 88,
        "group_name": "plus-team-mixed",
        "platform": "openai",
        "debit_multiplier": 1.0,
        "sort_order": 10
      },
      {
        "group_id": 89,
        "group_name": "pro",
        "platform": "openai",
        "debit_multiplier": 1.5,
        "sort_order": 20
      }
    ]
  }
]
```

### Frontend Subscription Display

The frontend should render one product card, not one card per real group, for users who hold shared products.

Each product card shows:

- product name
- shared quota progress
- expiration
- visible real groups and each group's debit multiplier

For the first GPT migration:

- users see one `GPT 月卡`
- users see group list:
  - `plus/team 混池 · 1.0x`
  - `pro 池 · 1.5x`

### API Key UI

The API key creation flow continues to present real group choices.

Eligibility source:

- normal groups: existing logic
- product-settled groups: user's active product subscriptions expanded to groups

The API key list continues to display the real bound group only.

### Admin Product Management

Add a distinct admin subscription product management surface rather than overloading `groups` pages.

Required operations:

- create product
- edit product
- activate or disable product
- bind and unbind real groups
- set `debit_multiplier` per real group
- inspect active user product subscriptions

The first rollout needs enough UI and API to configure:

- `GPT 月卡`
- `plus/team` binding at `1.0x`
- `pro` binding at `1.5x`

## Redeem and Assignment Design

### Product Subscription Redeem Codes

New shared-product subscription codes target `product_id`.

Redeem behavior:

- validate the target product exists and is active
- create or extend `user_product_subscription`
- do not create legacy `user_subscriptions`

### Legacy Compatibility

Legacy group subscription codes targeting `group_id` remain supported during transition.

The admin API and UI must enforce:

- a subscription code may target one of `product_id` or `group_id`
- the request is invalid if both are set or both are missing

### Default Product Assignment

If the system continues to support default subscription grants for new users, the setting model must evolve from default group subscriptions to default product subscriptions for shared-product cases.

Legacy group defaults may remain temporarily for non-migrated products.

## Migration Design

### Migration Principles

- additive-only schema changes first
- idempotent data migration scripts
- explicit dry-run mode
- no destructive rewrite of legacy rows in the first rollout
- no mixed old/new settlement on migrated groups after cutover

### First GPT Migration

The first migration target is existing GPT monthly users.

The operational sequence:

1. Create a draft product `GPT 月卡`.
2. Bind the target real groups:
   - GPT `plus/team` mixed group at `1.0x`
   - GPT `pro` group at `1.5x`
3. Identify legacy GPT monthly `user_subscriptions` by configured source group IDs.
4. Backfill one `user_product_subscription` per targeted user/product.
5. Copy authoritative entitlement state from the legacy row:
   - status
   - starts_at
   - expires_at
   - current window starts
   - current daily, weekly, monthly usage
   - current daily carryover state
6. Record the mapping in `product_subscription_migration_sources`.
7. Validate migrated samples and aggregate counts.
8. Drain old pods.
9. Activate the product and enable product runtime for its groups.

### Migration Conflict Rules

Migration must stop or skip rows that violate assumptions:

- one user has multiple active legacy GPT monthly subscriptions that would map to one product without a deterministic merge rule
- target real group is already bound to another active product
- the backfill would create duplicate active `user_product_subscription` rows

Skipped rows must appear in the migration report. The script must not silently merge or overwrite ambiguous entitlements.

### Historical Usage

Historical `usage_logs` are not backfilled with new `product_id` references in the first iteration.

Authoritative active-window usage is carried over directly from the legacy `user_subscriptions` record. New traffic after cutover populates product fields on new usage logs.

## Error Model

Add explicit product-mode runtime errors:

- `SUBSCRIPTION_PRODUCT_NOT_FOUND`
- `SUBSCRIPTION_PRODUCT_INACTIVE`
- `PRODUCT_SUBSCRIPTION_INVALID`
- `PRODUCT_LIMIT_EXCEEDED`
- `PRODUCT_GROUP_BINDING_INACTIVE`
- `PRODUCT_MIGRATION_CONFLICT`

Required response metadata where available:

- `product_id`
- `product_name`
- `group_id`
- `group_name`
- `debit_multiplier`
- `remaining_budget`

These errors must not silently degrade to legacy group subscription checks or balance mode.

## Safety Rules

The following rules are mandatory and not optional implementation details:

- One real group may belong to at most one active product.
- One active user product subscription may exist per `user_id + product_id`.
- One API key binds one real group.
- Shared-product groups must not silently fall back to another group if their binding is removed.
- Shared-product runtime must not read both legacy `user_subscriptions` and new `user_product_subscriptions` as independent quota owners for the same request.
- Product mode uses product limits as the only authoritative quota source.
- Shared pool debits use standard USD cost multiplied by the product-group `debit_multiplier`.
- The runtime must fail closed on ambiguous settlement source conflicts.

## Testing Requirements

### Schema and Migration Tests

- new tables, indexes, partial unique constraints, and foreign keys exist
- migration scripts support dry-run
- migration scripts are idempotent when rerun
- conflicting source data is reported rather than silently merged

### Repository and Service Tests

- query active user product subscriptions
- expand product subscriptions into visible real groups
- validate API key eligibility through product ownership
- debit shared pool correctly at `1.0x` and `1.5x`
- debit two different API keys bound to different real groups against one shared product pool
- carryover and window resets operate at product subscription level
- product-settled billing increments `user_product_subscriptions` and does not increment legacy `user_subscriptions`
- product-settled usage logs populate `product_subscription_id` and leave legacy `subscription_id` null
- idempotent post-usage billing uses the product subscription as the one authoritative subscription quota owner

### Handler and Contract Tests

- new product endpoints return product-level response shapes
- API key create/update authorizes product-settled groups correctly
- shared-product runtime errors return stable codes and metadata
- admin redeem endpoints correctly validate `product_id` versus `group_id`
- admin API key rebind flows authorize product-settled groups using product ownership rules

### Rollout Verification

For the first GPT migration:

- sample users keep expected `starts_at` and `expires_at`
- sample users keep expected current usage values after backfill
- sample users newly see the `pro` real group
- `plus/team` traffic debits at `1.0x`
- `pro` traffic debits at `1.5x`
- post-cutover product subscription usage matches expected deltas in logs and admin views

## Rollout Design

The rollout must be explicit and interrupt-safe.

### Phase 1: additive schema

- deploy new tables and columns only
- no runtime behavior changes yet

### Phase 2: compatible backend

- deploy backend code that understands the new schema
- keep product runtime inactive while products are still `draft`
- legacy group-settled subscriptions remain authoritative

### Phase 3: product bootstrap

- create draft products and draft product-group bindings
- do not expose them to runtime yet

### Phase 4: dry-run migration

- run dry-run backfill for the target GPT monthly population
- save the output report
- inspect conflict rows and sample rows before continuing

### Phase 5: apply migration

- execute the real backfill
- persist migration-source audit rows
- run validation queries and sample checks

### Phase 6: cutover preparation

- ensure all old pods are drained before product-settled groups become active
- this is required because old pods do not understand the new quota owner

### Phase 7: activate product runtime

- activate the migrated products
- enable product eligibility and product billing for their groups

### Phase 8: frontend switch

- deploy frontend that renders shared products and their group lists
- ensure API key creation uses product-expanded eligible groups

### Phase 9: post-cutover verification

- confirm sample users see `GPT 月卡`
- confirm sample users can create `pro` keys
- confirm usage debits the product pool at configured multipliers
- monitor errors for settlement-source conflicts and missing product bindings

## Rollback Design

The first iteration uses **logical rollback**, not destructive data rollback.

Rollback rules:

- do not delete legacy `user_subscriptions`
- do not delete new product data during emergency rollback
- disable product runtime by deactivating the product or the runtime gate
- for migrated groups that still have a preserved legacy settlement source, runtime may fall back to legacy settlement once product runtime is disabled
- for product-only groups introduced by the shared-product model and lacking preserved legacy entitlements, runtime must fail closed once product runtime is disabled
- existing API keys bound to product-only groups become unusable during rollback until product runtime is restored or an operator remaps those keys to a supported group
- the rollout runbook must therefore include:
  - identifying product-only groups in the affected product
  - counting existing API keys bound to those groups before cutover
  - operator steps for communication, optional key remap, or temporary group restriction during rollback

This rollback is safe because legacy rows remain intact for the migrated legacy-source groups, old/new settlement are never both active at once for the same group, and product-only groups explicitly fail closed instead of silently degrading to another settlement source.

The rollout does not require a reverse migration of user data during the first emergency response path.

## File Boundary for Implementation

The implementation will span:

- new ent schemas and SQL migrations for product tables and log columns
- repository layer for product subscriptions and product-group bindings
- subscription service replacements or extensions for product-mode lookup and usage increment
- billing eligibility and gateway post-billing paths
- redeem service and admin redeem APIs
- API key visibility and binding checks
- new user-facing subscription product handlers and DTOs
- frontend subscription stores and components
- admin product management UI and API integration
- migration scripts, validation scripts, and rollout notes

The implementation plan must split these into independently reviewable tasks and include a separate rollout runbook.

## First-Rollout Acceptance Criteria

The first GPT monthly rollout is successful only if all of the following are true:

- existing GPT monthly users are migrated to one `GPT 月卡` product subscription
- migrated users keep their effective validity window
- migrated users newly see the `pro` real group
- users still create API keys against one real group at a time
- `plus/team` usage debits the shared GPT monthly pool at `1.0x`
- `pro` usage debits the same pool at `1.5x`
- frontend shows one product card instead of multiple fake subscriptions for migrated GPT monthly users
- rollback can disable product runtime without destructive reverse migration

## Future Extension Boundary

This design intentionally leaves room for future products that span multiple platforms, including GPT and Claude in the same shared pool.

The reusable assumptions are:

- product quota unit is standard-cost USD
- real groups remain routing targets
- the debit multiplier is defined on the product-group binding, not the product subscription itself
- user ownership is product-centric, not group-centric

Those constraints are sufficient for future cross-platform products without revisiting the core settlement model introduced here.
