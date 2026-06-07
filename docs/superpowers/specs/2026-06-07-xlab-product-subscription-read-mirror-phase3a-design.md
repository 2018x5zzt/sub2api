# Xlab Product Subscription Read Mirror Phase 3A Design

## Background

Phase 1 established the first `xlab-backend` boundary and Phase 2 put it into production behind `/xapi/v1`. The production frontend now calls xlab product-subscription APIs through the same-origin xlab route, but the current `xlab-backend` implementation is still a thin authenticated proxy to core:

```text
frontend-v2 -> /xapi/v1/subscription-products/* -> xlab-backend -> core /api/v1/subscription-products/*
```

This is a useful routing boundary, but it does not yet make core replaceable. Core still owns product subscription data, payment fulfillment, redeem grants, admin product management, gateway authorization, and usage billing. Directly switching core to upstream latest would still risk existing paid users because upstream ranges after `v0.1.125` touch payment, subscription, redeem, quota, and gateway usage behavior.

Phase 3A moves only the read data source for product-subscription pages into `xlab-backend`. It deliberately keeps all product-subscription writers and runtime authorization in current core until later phases.

## Goals

1. Add a persistent xlab DB layer to `xlab-backend`.
2. Mirror core product-subscription read data into xlab-owned tables.
3. Serve `/xapi/v1/subscription-products/active`, `/summary`, and `/progress` primarily from xlab DB.
4. Preserve the existing frontend-v2 response contract exactly.
5. Keep core proxy fallback so read outages or stale mirror data can be rolled back without frontend redeploys.
6. Add sync health, stale-data detection, and diff/audit hooks before relying on xlab DB reads.
7. Avoid any change to payment fulfillment, redeem, admin product writes, gateway billing, API key authorization, or core migrations in this phase.

## Non-goals

- Do not move payment orders or payment callbacks to `xlab-backend`.
- Do not move redeem-code product grants to `xlab-backend`.
- Do not move admin product-subscription create/update/assign/reset/revoke APIs.
- Do not change core gateway authorization or usage billing.
- Do not remove or disable core `/api/v1/subscription-products/*` APIs.
- Do not switch core to upstream latest in this phase.
- Do not make xlab DB the source of truth for paid user entitlements yet.

## Current implementation summary

### `xlab-backend`

Current code is intentionally small:

- `xlab-backend/internal/config/config.go`
  - Loads `XLAB_SERVER_ADDR`, `CORE_API_BASE_URL`, and `XLAB_CORE_TIMEOUT_SECONDS`.
- `xlab-backend/internal/core/client.go`
  - Validates core JWT via `GET /user/profile`.
  - Proxies core GET requests and unwraps core envelopes.
- `xlab-backend/internal/httpapi/router.go`
  - Registers `/health` and three `/xapi/v1/subscription-products/*` routes.
- `xlab-backend/internal/httpapi/auth.go`
  - Requires bearer token and validates it through core.
  - Currently keeps only the token in request context.
- `xlab-backend/internal/httpapi/subscription_products.go`
  - Proxies `/active`, `/summary`, and `/progress` to core.

There is no xlab DB config, migration runner, repository, service layer, sync job, admin API, or product-subscription writer.

### Core product-subscription dependency

Core currently owns the product-subscription behavior through raw SQL tables and service/repository code:

- User routes in `backend/internal/server/routes/user.go`:
  - `GET /api/v1/subscription-products/active`
  - `GET /api/v1/subscription-products/summary`
  - `GET /api/v1/subscription-products/progress`
- Handler and DTOs:
  - `backend/internal/handler/subscription_product_handler.go`
  - `backend/internal/handler/dto/subscription_product.go`
- Service and repository:
  - `backend/internal/service/subscription_product.go`
  - `backend/internal/service/subscription_product_service.go`
  - `backend/internal/repository/subscription_product_repo.go`
- Raw SQL migrations:
  - `backend/migrations/140_restore_shared_subscription_products.sql`
  - `backend/migrations/141_converge_legacy_group_subscriptions_to_products.sql`
  - `backend/migrations/142_add_redeem_code_product_id.sql`
  - `backend/migrations/144_product_subscription_family_balance_fallback.sql`
  - `backend/migrations/145_product_subscription_explicit_fallback_family.sql`
  - `backend/migrations/147_product_subscription_family_gpt.sql`
  - `backend/migrations/150_add_subscription_plan_product_id.sql`
  - `backend/migrations/151_align_active_product_subscription_expiry_to_day_end.sql`

Core also couples product subscriptions into payment, redeem, API key availability, gateway billing, and usage-log persistence. Those paths remain in core for Phase 3A.

## Target architecture

```text
                      read requests
Browser / frontend-v2 ───────────────▶ /xapi/v1
                                            │
                                            ▼
                                      xlab-backend
                                      │         │
                         fresh mirror │         │ fallback/stale/error
                                      ▼         ▼
                                  xlab DB     core /api/v1
                                      ▲
                                      │ sync job
                                      │
                          core product-subscription tables
```

Core remains source of truth for writes and runtime entitlements during Phase 3A. Xlab DB becomes a read mirror for frontend subscription display.

## Runtime modes

Add an explicit read-source mode:

```text
XLAB_SUBSCRIPTION_READ_SOURCE=core|hybrid|xlab
```

Recommended meanings:

- `core`: always use the existing core proxy. This is the default rollback mode.
- `hybrid`: use xlab DB when fresh and complete enough; fallback to core on stale data, missing sync state, empty user results with core data, or DB errors.
- `xlab`: prefer xlab DB. Fallback remains available unless explicitly disabled by a later phase.

Add stale threshold config:

```text
XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS=600
```

Ten minutes is a safe default for initial production validation. It can be tightened after sync behavior is observed.

## Data model

The xlab DB schema should mirror only the fields needed for the current frontend read contract and future reconciliation. It should not attempt to model every core table yet.

### `xlab_subscription_products`

Stores product definitions mirrored from core.

```text
core_product_id bigint primary key
code text not null
name text not null
description text not null default ''
status text not null
product_family text not null default 'gpt'
daily_limit_usd numeric(18,6)
weekly_limit_usd numeric(18,6)
monthly_limit_usd numeric(18,6)
daily_carryover_enabled boolean not null default false
daily_carryover_limit_usd numeric(18,6)
source_created_at timestamptz
source_updated_at timestamptz
synced_at timestamptz not null
```

### `xlab_subscription_product_groups`

Stores product-to-group bindings and denormalized group display fields.

```text
core_binding_id bigint primary key
core_product_id bigint not null
core_group_id bigint not null
group_name text not null
group_platform text
balance_fallback_group_id bigint
balance_fallback_group_name text
debit_multiplier numeric(18,6) not null default 1
status text not null
sort_order integer not null default 0
source_created_at timestamptz
source_updated_at timestamptz
synced_at timestamptz not null
```

### `xlab_user_product_subscriptions`

Stores user subscription read state.

```text
core_subscription_id bigint primary key
core_user_id bigint not null
core_product_id bigint not null
status text not null
started_at timestamptz
expires_at timestamptz
daily_usage_usd numeric(18,6) not null default 0
weekly_usage_usd numeric(18,6) not null default 0
monthly_usage_usd numeric(18,6) not null default 0
daily_limit_usd numeric(18,6)
weekly_limit_usd numeric(18,6)
monthly_limit_usd numeric(18,6)
daily_carryover_in_usd numeric(18,6) not null default 0
daily_carryover_remaining_usd numeric(18,6) not null default 0
source_created_at timestamptz
source_updated_at timestamptz
synced_at timestamptz not null
```

### `xlab_sync_state`

Stores sync health and stale-data decisions.

```text
source_name text primary key
last_success_at timestamptz
last_watermark text
last_error text
last_error_at timestamptz
row_count integer not null default 0
created_at timestamptz not null default now()
updated_at timestamptz not null default now()
```

Use `source_name='product_subscriptions'` for the initial sync job.

## Sync strategy

### Phase 3A initial sync

Use scheduled full snapshots first. Do not add CDC, webhooks, or core triggers in this phase.

Recommended interval:

```text
XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS=300
```

Full snapshot steps:

1. Connect to core DB using a read-only or least-privilege credential.
2. Read `subscription_products`.
3. Read `subscription_product_groups` joined with `groups` for display fields.
4. Read active and recently expired `user_product_subscriptions` needed for current users and summary views.
5. Upsert rows into xlab mirror tables in a transaction.
6. Mark missing core rows as inactive/revoked or delete them according to a deterministic policy.
7. Update `xlab_sync_state` with `last_success_at`, `row_count`, and any watermark.
8. Emit structured logs with row counts and duration.

### Later sync improvements

After initial production validation, consider incremental sync using `updated_at` watermarks or event logs. Do not build that into the first Phase 3A implementation unless full snapshot size proves unsafe.

## API behavior

### Authentication

Continue using core as auth source:

1. `xlab-backend` receives bearer JWT.
2. It calls core `GET /user/profile`.
3. It stores both token and authenticated core user in request context.
4. Subscription read services query by `core_user_id`.

The current middleware stores only the token. Phase 3A must preserve the validated user in context as well.

### `/active`

For `GET /xapi/v1/subscription-products/active`:

1. Validate token through core.
2. If mode is `core`, proxy to core.
3. If mode is `hybrid` or `xlab`, check sync freshness.
4. Query active rows for `core_user_id` from xlab DB.
5. Join mirrored products and groups.
6. Apply the same active/expiry and display normalization as core reads.
7. Return the same array shape currently consumed by frontend-v2.
8. Fallback to core if configured conditions apply.

### `/summary` and `/progress`

Core currently treats progress similarly to summary. Phase 3A should preserve that behavior unless a later product requirement changes it.

Summary must include:

```text
active_count
total_monthly_usage_usd
total_monthly_limit_usd
products
```

### Fallback conditions

Fallback to core proxy in `hybrid` mode when:

- xlab DB connection fails.
- `xlab_sync_state` is missing.
- `last_success_at` is older than `XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS`.
- xlab DB query fails.
- xlab DB returns no active rows for a user and core returns active rows.
- response contract mapping fails before writing a response.

In `xlab` mode, fallback should still be available initially. A future phase may add a separate setting to disable fallback after confidence is high.

## Response contract

The frontend-v2 app expects the current fields from `frontend-v2/src/types/index.ts`:

```text
ActiveSubscriptionProduct
SubscriptionProductGroup
SubscriptionProductSummary
```

Phase 3A must not rename fields, remove fields, or change numeric units. Keep USD values in the same units and precision semantics as core responses.

The xlab API envelope must remain compatible with `frontend-v2/src/api/xlabClient.ts`, which unwraps `{ code: 0, data }`.

## Observability and audit

Add logs and optional diagnostic endpoints or CLI commands for:

- Latest sync success time.
- Latest sync error.
- Product count.
- Binding count.
- Active subscription count.
- Fallback count by reason.
- Per-user diff between xlab DB response and core proxy response for a supplied user ID or bearer token.

Any diagnostic endpoint must be admin-only or disabled by default if exposed over HTTP.

## Rollout plan

1. Deploy new `xlab-backend` with DB config present but `XLAB_SUBSCRIPTION_READ_SOURCE=core`.
2. Run migrations for xlab DB.
3. Start sync job and observe row counts.
4. Run diff/audit against core responses for selected users.
5. Switch to `hybrid` mode.
6. Watch fallback counts and subscription page behavior.
7. If stable, optionally switch to `xlab` mode while keeping fallback enabled.
8. Leave frontend-v2 on `/xapi/v1`; no frontend redeploy is required for read-source mode changes.

## Rollback plan

Fast rollback is runtime config only:

```text
XLAB_SUBSCRIPTION_READ_SOURCE=core
```

Then restart `xlab-backend`.

If a frontend-level rollback is also needed, rebuild the embedded frontend with:

```bash
VITE_XLAB_API_BASE_URL=/api/v1
```

No database rollback is required because Phase 3A mirror tables are not source of truth and do not write back to core.

## Testing requirements

### `xlab-backend`

- Config tests for DB and read-source env vars.
- Auth middleware test proving the validated core user is stored in context.
- Repository tests for active products by user.
- Repository tests for summary and progress response data.
- Expired subscription tests.
- Group binding ordering tests.
- Carryover field tests.
- Fallback tests for DB unavailable, stale sync, missing sync state, and core unavailable.
- Contract tests comparing mirrored response shape with core fixture response.

### Core regression tests

Run existing product-subscription tests that protect current paid-user behavior:

- `backend/internal/handler/subscription_product_handler_test.go`
- `backend/internal/repository/subscription_product_repo_integration_test.go`
- `backend/internal/service/subscription_product_service_test.go`
- `backend/internal/service/payment_subscription_product_fulfillment_test.go`
- `backend/internal/service/redeem_product_subscription_test.go`
- `backend/internal/service/api_key_service_available_groups_test.go`
- `backend/internal/service/gateway_service_subscription_billing_test.go`
- `backend/internal/repository/usage_billing_repo_integration_test.go`
- `backend/internal/repository/usage_log_repo_integration_test.go`
- `backend/migrations/auth_identity_payment_migrations_regression_test.go`

### Frontend-v2

- Existing xlab adapter tests.
- `npm --prefix frontend-v2 run typecheck`.
- `VITE_XLAB_API_BASE_URL=/xapi/v1 npm --prefix frontend-v2 run build`.

## Safety gates before production `hybrid`

- `xlab-backend` tests pass.
- Core product-subscription regression tests pass.
- Frontend-v2 typecheck and build pass.
- Xlab DB sync has at least one successful full snapshot.
- Diff/audit for selected active subscription users matches core on the fields consumed by frontend-v2.
- `/xapi/v1/subscription-products/active` works with fallback disabled in a staging or controlled local environment.

## How this helps upstream migration

After Phase 3A, frontend product-subscription reads no longer depend directly on core read APIs. This reduces one part of the future upstream upgrade risk, but it does not yet make core replaceable. Later phases still need to move admin writes, payment fulfillment, redeem grants, entitlement projection, and gateway authorization semantics before core can safely converge to upstream latest.
