# Shared Subscription Products Rollout Runbook

## Purpose

This runbook describes how to roll out the first shared subscription product migration for the GPT monthly population, validate the cutover, and recover safely if product runtime must be disabled.

The first product in scope is:

- product code: `gpt_monthly`
- product name: `GPT 月卡`
- real groups:
  - `plus/team` mixed pool at `1.0x`
  - `pro` pool at `1.5x`

## Required Preconditions

Before starting, confirm all of the following are true:

- the backend build includes migration `110_add_shared_subscription_products.sql`
- the backend build includes product-aware billing, API key authorization, redeem, and admin product APIs
- the frontend build includes shared product cards and product-expanded key group selection
- the backfill and validation tools exist:
  - `backend/cmd/shared-subscription-products-backfill/main.go`
  - `backend/cmd/shared-subscription-products-validate/main.go`
- old pods can be drained completely before activation
- you have:
  - an admin API token
  - DB access for validation queries
  - a list of sample migrated users to spot-check after cutover

Recommended environment variables:

```bash
export API_BASE="https://<your-host>/api/v1"
export ADMIN_TOKEN="<admin-token>"
export PRODUCT_CODE="gpt_monthly"
export PRODUCT_NAME="GPT 月卡"
export PLUS_TEAM_GROUP_ID="88"
export PRO_GROUP_ID="89"
export MIGRATION_BATCH="2026-04-25-gpt-monthly"
export PRODUCT_ID=""
```

If your real group IDs differ from `88` and `89`, replace them before running any command.

## Build Verification Checklist

Run these checks against the exact backend and frontend revisions planned for rollout before Phase 0.

Backend:

```bash
cd backend
go test -tags=integration ./internal/repository -run 'Test(MigrationsSharedSubscriptionProductsSchema|SubscriptionProductRepository_|UsageBillingRepositoryApply_ProductSubscriptionCostAdvancesProductWindows)' -count=1
go test -tags=unit ./internal/service -run 'Test(AdminService_AdminUpdateAPIKeyGroupID_ProductSettledGroupAllowsActiveProductSubscription|APIKeyService_Create_ProductSettledGroupRequiresProductSubscription|RegisterUser_AppliesDefaultProductSubscriptions|OpenAIGatewayRecordUsage_ProductSettlement|SettingService_UpdateSettings_PersistsDefaultProductSubscriptions|RedeemService_Redeem_ProductSubscriptionCodeCreatesUserProductSubscription)' -count=1
go test -tags=unit ./internal/server/middleware -run 'TestAPIKeyAuth_ProductSettledGroup(LoadsProductContext|ReturnsStructuredRuntimeError)' -count=1
go test ./internal/handler -run 'TestSubscriptionProductHandler_' -count=1
go test ./internal/handler/admin -run 'Test(AdminSubscriptionProductHandler_|AdminRedeemHandler_Generate_ProductIDXorGroupID)' -count=1
go test ./cmd/shared-subscription-products-backfill -run 'Test(BuildBackfillReport_|ApplyBackfill_IdempotentWhenMigrationBatchReruns)' -count=1
```

The integration repository check uses testcontainers and requires local Docker access. Expected result: all backend commands print `ok`.

Frontend:

```bash
cd frontend
npm run test:run -- src/stores/__tests__/subscriptionProducts.spec.ts src/components/common/__tests__/SubscriptionProgressMini.spec.ts src/components/__tests__/ApiKeyCreate.spec.ts src/views/admin/__tests__/SubscriptionProductsView.spec.ts src/views/admin/__tests__/SettingsView.spec.ts
npm run typecheck
```

Expected result: Vitest reports all listed files passing and `vue-tsc --noEmit` exits successfully.

## Phase 0: Preflight Checks

1. Confirm the additive schema is live.

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_name IN (
  'subscription_products',
  'subscription_product_groups',
  'user_product_subscriptions',
  'product_subscription_migration_sources'
)
ORDER BY table_name;
```

Expected: all four tables exist.

2. Confirm the additive columns are live.

```sql
SELECT column_name
FROM information_schema.columns
WHERE table_name = 'usage_logs'
  AND column_name IN ('product_id', 'product_subscription_id', 'group_debit_multiplier', 'product_debit_cost')
ORDER BY column_name;
```

Expected: all four columns exist.

3. Confirm there is no pre-existing active-product conflict on the target groups.

```sql
SELECT sp.code AS product_code, spg.group_id, spg.status
FROM subscription_product_groups spg
JOIN subscription_products sp ON sp.id = spg.product_id
WHERE spg.deleted_at IS NULL
  AND sp.deleted_at IS NULL
  AND spg.group_id IN (88, 89)
  AND sp.status = 'active';
```

Expected: zero rows before this rollout.

## Phase 1: Deploy Compatible Backend With Runtime Still Inactive

1. Deploy the backend version that understands shared subscription products.
2. Do not create or activate the product yet.
3. Confirm the service is healthy.

```bash
curl -fsS "${API_BASE%/api/v1}/health"
```

Expected: HTTP `200`.

4. Confirm legacy GPT monthly traffic still works normally before product activation.

Spot-check:

- one existing GPT monthly API key request succeeds
- the request still writes legacy `usage_logs.subscription_id`
- no `product_id` fields are present yet on new traffic

## Phase 2: Bootstrap the Draft Product

1. Create the draft product.

```bash
curl -fsS -X POST "$API_BASE/admin/subscription-products" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "gpt_monthly",
    "name": "GPT 月卡",
    "description": "GPT monthly shared subscription",
    "status": "draft",
    "default_validity_days": 30,
    "monthly_limit_usd": 100,
    "sort_order": 10
  }'
```

2. Resolve the created product ID and export it for the remaining API calls.

```sql
SELECT id, code, status
FROM subscription_products
WHERE code = 'gpt_monthly'
  AND deleted_at IS NULL;
```

Expected: exactly one row. Then set:

```bash
export PRODUCT_ID="<id from the query above>"
```

3. Bind the real groups while the product remains draft.

```bash
curl -fsS -X PUT "$API_BASE/admin/subscription-products/$PRODUCT_ID/bindings" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"bindings\": [
      {\"group_id\": ${PLUS_TEAM_GROUP_ID}, \"debit_multiplier\": 1.0, \"status\": \"active\", \"sort_order\": 10},
      {\"group_id\": ${PRO_GROUP_ID}, \"debit_multiplier\": 1.5, \"status\": \"active\", \"sort_order\": 20}
    ]
  }"
```

4. Confirm the product exists but is still runtime-inactive.

```sql
SELECT id, code, status
FROM subscription_products
WHERE code = 'gpt_monthly'
  AND deleted_at IS NULL;
```

Expected: one row with `status = 'draft'`.

## Phase 3: Count Product-Only Keys Before Cutover

This rollout introduces `pro` as a product-only group for migrated GPT monthly users. Rollback for those keys is fail-closed, not legacy fallback.

Before cutover, record the count of keys already bound to the future product-only group:

```sql
SELECT COUNT(*) AS api_key_count
FROM api_keys
WHERE deleted_at IS NULL
  AND group_id = 89;
```

If your actual `pro` group ID is not `89`, replace the literal before recording the count.

Expected: typically `0` before activation. Save this number in the rollout ticket.

## Phase 4: Dry-Run Backfill

Run the backfill in dry-run mode first.

```bash
cd /root/sub2api-src/backend && \
go run ./cmd/shared-subscription-products-backfill \
  --dry-run \
  --product-code "$PRODUCT_CODE" \
  --source-group-ids "${PLUS_TEAM_GROUP_ID}" \
  --migration-batch "$MIGRATION_BATCH"
```

Expected dry-run output:

- total candidate legacy subscriptions
- skipped rows with explicit reasons
- duplicate or ambiguous users
- sample rows showing legacy subscription -> user product subscription mapping

Do not continue until:

- skipped rows are understood
- ambiguous users are explicitly handled or excluded
- sample windows, usage, and carryover values match the source legacy rows

## Phase 5: Apply Backfill

Once dry-run output is accepted, run the apply step.

```bash
cd /root/sub2api-src/backend && \
go run ./cmd/shared-subscription-products-backfill \
  --product-code "$PRODUCT_CODE" \
  --source-group-ids "${PLUS_TEAM_GROUP_ID}" \
  --migration-batch "$MIGRATION_BATCH"
```

Then run validation:

```bash
cd /root/sub2api-src/backend && \
go run ./cmd/shared-subscription-products-validate \
  --product-code "$PRODUCT_CODE" \
  --source-group-ids "${PLUS_TEAM_GROUP_ID}" \
  --migration-batch "$MIGRATION_BATCH"
```

Required validation queries:

```sql
SELECT COUNT(*) AS migrated_rows
FROM user_product_subscriptions ups
JOIN subscription_products sp ON sp.id = ups.product_id
WHERE sp.code = 'gpt_monthly'
  AND ups.deleted_at IS NULL;
```

```sql
SELECT COUNT(*) AS audit_rows
FROM product_subscription_migration_sources psms
JOIN user_product_subscriptions ups ON ups.id = psms.product_subscription_id
JOIN subscription_products sp ON sp.id = ups.product_id
WHERE sp.code = 'gpt_monthly';
```

Expected: counts match the applied backfill report.

Capture sample migrated rows for post-cutover verification:

```sql
SELECT
  ups.id AS product_subscription_id,
  ups.user_id,
  ups.product_id
FROM user_product_subscriptions ups
JOIN subscription_products sp ON sp.id = ups.product_id
WHERE sp.code = 'gpt_monthly'
  AND ups.deleted_at IS NULL
ORDER BY ups.id
LIMIT 5;
```

Save at least one `user_id` and its matching `product_subscription_id` from this query for Phase 9 spot checks.

## Phase 6: Drain Old Pods Before Activation

Do not activate product runtime while old pods are still serving traffic.

Checklist:

- deployment rollout is complete
- old backend pods are fully drained
- only the product-aware backend revision is serving requests

If you cannot prove that old pods are gone, stop here.

## Phase 7: Activate Product Runtime

1. Activate the product.

```bash
curl -fsS -X PUT "$API_BASE/admin/subscription-products/$PRODUCT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "active"
  }'
```

2. Confirm the product is active.

```sql
SELECT code, status
FROM subscription_products
WHERE code = 'gpt_monthly'
  AND deleted_at IS NULL;
```

Expected: one row with `status = 'active'`.

3. Confirm no target group is still settling through legacy runtime in code or config.

Operationally, this means:

- migrated users should resolve `user_product_subscriptions`
- new product traffic should write `usage_logs.product_id`
- new product traffic should leave `usage_logs.subscription_id` null

## Phase 8: Deploy Frontend Switch

Deploy the frontend revision that:

- renders product cards instead of fake per-group GPT subscriptions
- shows `plus/team · 1.0x` and `pro · 1.5x` under `GPT 月卡`
- uses product-expanded group choices in API key creation

Do not treat the rollout as complete until both backend and frontend are live.

## Phase 9: Post-Cutover Verification

### User-visible checks

For at least three sample migrated users, verify all of the following:

- `GET /api/v1/subscription-products/active` returns one `GPT 月卡`
- the response includes both `plus/team` and `pro`
- the user can create a new key bound to `pro`
- the user subscription page shows one product card, not multiple fake GPT group cards

### Billing and logging checks

Run one `plus/team` request and one `pro` request using separate keys owned by the same migrated user, then query:

```sql
SELECT
  created_at,
  group_id,
  product_id,
  product_subscription_id,
  group_debit_multiplier,
  total_cost,
  actual_cost,
  product_debit_cost,
  subscription_id
FROM usage_logs
WHERE user_id = <sample_user_id>
ORDER BY created_at DESC
LIMIT 10;
```

Replace `<sample_user_id>` with one migrated `user_id` captured in Phase 5.

Expected:

- `group_id` matches the real bound group
- `product_id` and `product_subscription_id` are non-null
- `subscription_id` is null for product-settled rows
- `plus/team` rows show `group_debit_multiplier = 1.0`
- `pro` rows show `group_debit_multiplier = 1.5`
- `product_debit_cost = total_cost * group_debit_multiplier`

### Quota-owner checks

```sql
SELECT
  monthly_usage_usd,
  weekly_usage_usd,
  daily_usage_usd
FROM user_product_subscriptions
WHERE id = <sample_product_subscription_id>;
```

Replace `<sample_product_subscription_id>` with the matching Phase 5 sample row.

Expected:

- usage increments after each request
- the increment amount follows `product_debit_cost`
- the matching legacy `user_subscriptions` row does not continue advancing for the same request

## Emergency Rollback

Use rollback only if the new runtime must be disabled urgently.

### Safe rollback actions

1. Disable product runtime by deactivating the product.

```bash
curl -fsS -X PUT "$API_BASE/admin/subscription-products/$PRODUCT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "disabled"
  }'
```

2. Confirm the product is disabled.

```sql
SELECT code, status
FROM subscription_products
WHERE code = 'gpt_monthly';
```

3. For migrated legacy-source groups, verify requests now resolve through preserved legacy settlement again.

### Important rollback limitation

`pro` is a product-only group in this rollout. Once product runtime is disabled:

- keys bound to `pro` must fail closed
- they do not fall back to legacy `user_subscriptions`
- they remain unusable until product runtime is restored or an operator remaps them to a supported group

### Product-only key impact query

```sql
SELECT id, user_id, name
FROM api_keys
WHERE deleted_at IS NULL
  AND group_id = 89
ORDER BY id;
```

If your actual `pro` group ID is not `89`, replace the literal before running the query.

Use this output for:

- user communication
- temporary support handling
- operator remap planning if required

## Residual Risks

- Product activation before old pod drain can create split-brain settlement behavior.
- Shared-product runtime is still optimistic at request start: authorization happens before post-usage write completion.
- Product-only groups such as `pro` intentionally fail closed during rollback.
- Process-local L1 caches can lag briefly after invalidation; verification should allow short convergence time after cutover writes.

## Completion Criteria

Treat the rollout as complete only when all of the following are true:

- the product is active
- migrated sample users see `GPT 月卡`
- migrated sample users can create `pro` keys
- `plus/team` and `pro` requests both debit the same `user_product_subscription`
- `usage_logs.subscription_id` remains null for product-settled rows
- no unexpected `PRODUCT_*` or settlement-conflict errors appear after cutover
