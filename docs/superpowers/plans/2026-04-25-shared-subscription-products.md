# Shared Subscription Products Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce product-level shared subscription entitlements that expose multiple real groups, keep API keys bound to one real group, and debit one shared product quota pool with per-group multipliers.

**Architecture:** Add additive product tables plus product-aware usage-log and redeem-code columns, then route all migrated subscription settlement through a new product settlement resolver and the existing deduplicated post-usage billing path. Keep legacy `user_subscriptions` intact for non-migrated groups and rollback, while introducing separate user/admin product APIs and frontend surfaces for shared-product users and operators.

**Tech Stack:** Go, Gin, ent, PostgreSQL SQL migrations, Google Wire, Vue 3, Pinia, Vitest, testify

---

## File Map

### Backend schema and generated models

- Create `backend/migrations/110_add_shared_subscription_products.sql` to add new product tables, indexes, constraints, and additive columns on `redeem_codes` and `usage_logs`.
- Create `backend/ent/schema/subscription_product.go` to define product metadata and authoritative limits.
- Create `backend/ent/schema/subscription_product_group.go` to define product-to-real-group bindings and multipliers.
- Create `backend/ent/schema/user_product_subscription.go` to define the authoritative shared quota owner.
- Create `backend/ent/schema/product_subscription_migration_source.go` to preserve backfill audit lineage.
- Modify `backend/ent/schema/redeem_code.go` and `backend/ent/schema/usage_log.go` to model new foreign keys and product billing fields.

### Backend repositories and services

- Create `backend/internal/repository/subscription_product_repo.go` for product CRUD, product-group bindings, and user product subscription queries.
- Create `backend/internal/repository/subscription_product_repo_integration_test.go` for schema-backed repository coverage.
- Create `backend/internal/service/subscription_product.go` and `backend/internal/service/user_product_subscription.go` for service-layer models.
- Create `backend/internal/service/subscription_product_errors.go` for reusable product-mode runtime errors and metadata builders.
- Create `backend/internal/service/subscription_product_port.go` for repository interfaces used by services and middleware.
- Create `backend/internal/service/subscription_product_service.go` for eligibility, visibility expansion, limit checks, and cache invalidation.
- Create `backend/internal/service/subscription_product_service_test.go` for product-settlement behavior.
- Create `backend/internal/service/subscription_product_errors_test.go` for stable reason-code and metadata coverage.
- Create `backend/internal/service/product_settlement.go` for the one-request authoritative settlement context.
- Create `backend/internal/service/product_settlement_test.go` for resolver edge cases.

### Backend runtime and API surfaces

- Modify `backend/internal/service/usage_billing.go`, `backend/internal/service/usage_log.go`, `backend/internal/repository/usage_billing_repo.go`, and `backend/internal/repository/usage_log_repo.go` to carry product billing state through the authoritative billing path.
- Modify `backend/internal/service/gateway_service.go` and `backend/internal/service/openai_gateway_service.go` to resolve product settlement context and dual-write product usage metadata.
- Modify `backend/internal/service/api_key_service.go`, `backend/internal/service/admin_service.go`, `backend/internal/service/billing_cache_service.go`, and `backend/internal/server/middleware/api_key_auth.go` to authorize product-settled groups and cache product ownership.
- Create `backend/internal/handler/subscription_product_handler.go` and `backend/internal/handler/admin/subscription_product_handler.go` for user/admin product endpoints.
- Create `backend/internal/handler/dto/subscription_product.go` for user/admin response mapping.
- Modify `backend/internal/handler/handler.go`, `backend/internal/handler/wire.go`, `backend/internal/server/routes/user.go`, and `backend/internal/server/routes/admin.go` to register new handlers and routes.

### Backend admin and migration support

- Modify `backend/internal/service/redeem_service.go`, `backend/internal/handler/admin/redeem_handler.go`, and `frontend/src/api/admin/redeem.ts` to support `product_id`-backed subscription codes.
- Modify `backend/internal/service/domain_constants.go`, `backend/internal/service/setting_service.go`, `backend/internal/handler/admin/setting_handler.go`, and `backend/internal/handler/dto/settings.go` to support default product subscriptions alongside legacy default group subscriptions.
- Create `backend/cmd/shared-subscription-products-backfill/main.go` for dry-run/apply backfill.
- Create `backend/cmd/shared-subscription-products-validate/main.go` for post-backfill validation queries and sample checks.

### Frontend

- Create `frontend/src/api/subscriptionProducts.ts` and `frontend/src/stores/subscriptionProducts.ts` for user-facing shared-product data.
- Create `frontend/src/api/admin/subscriptionProducts.ts` for admin product management.
- Modify `frontend/src/types/index.ts` to add shared-product response types, redeem payload support, and admin settings fields.
- Modify `frontend/src/views/user/SubscriptionsView.vue`, `frontend/src/components/common/SubscriptionProgressMini.vue`, `frontend/src/views/user/KeysView.vue`, and `frontend/src/components/common/GroupSelector.vue` to render product cards and product-expanded group choices.
- Create `frontend/src/views/admin/SubscriptionProductsView.vue` and modify `frontend/src/router/index.ts`, `frontend/src/api/admin/index.ts`, `frontend/src/views/admin/RedeemView.vue`, and `frontend/src/views/admin/SettingsView.vue` for admin product operations.

### Verification and docs

- Modify `backend/internal/repository/migrations_schema_integration_test.go`, `backend/internal/repository/usage_billing_repo_integration_test.go`, `backend/internal/service/openai_gateway_record_usage_test.go`, `backend/internal/service/gateway_record_usage_test.go`, `backend/internal/service/admin_service_apikey_test.go`, `frontend/src/stores/__tests__/subscriptions.spec.ts`, `frontend/src/components/common/__tests__/SubscriptionProgressMini.spec.ts`, `frontend/src/components/__tests__/ApiKeyCreate.spec.ts`, and `frontend/src/views/admin/__tests__/SettingsView.spec.ts` to lock the new behavior.
- Create `docs/superpowers/plans/2026-04-25-shared-subscription-products-rollout.md` as the operator-facing rollout and rollback runbook.

### Task 1: Add additive schema and ent models

**Files:**
- Create: `backend/migrations/110_add_shared_subscription_products.sql`
- Create: `backend/ent/schema/subscription_product.go`
- Create: `backend/ent/schema/subscription_product_group.go`
- Create: `backend/ent/schema/user_product_subscription.go`
- Create: `backend/ent/schema/product_subscription_migration_source.go`
- Modify: `backend/ent/schema/redeem_code.go`
- Modify: `backend/ent/schema/usage_log.go`
- Modify: `backend/internal/repository/migrations_schema_integration_test.go`
- Test: `backend/internal/repository/migrations_schema_integration_test.go`

- [ ] **Step 1: Write the failing schema tests**

Add focused migration assertions such as:

```go
func TestMigrationsSharedSubscriptionProductsSchema(t *testing.T) {
    requireColumn(t, db, "subscription_products", "code")
    requireColumn(t, db, "subscription_product_groups", "debit_multiplier")
    requireColumn(t, db, "user_product_subscriptions", "daily_carryover_remaining_usd")
    requireColumn(t, db, "product_subscription_migration_sources", "legacy_user_subscription_id")
    requireColumn(t, db, "redeem_codes", "product_id")
    requireColumn(t, db, "usage_logs", "product_subscription_id")
    requireColumn(t, db, "usage_logs", "product_debit_cost")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/repository -run TestMigrationsSharedSubscriptionProductsSchema -count=1`
Expected: FAIL because the product tables and new columns do not exist yet.

- [ ] **Step 3: Write minimal schema and ent implementation**

Add the migration and ent fields:

```sql
CREATE TABLE IF NOT EXISTS subscription_products (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    default_validity_days INT NOT NULL DEFAULT 30,
    daily_limit_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    weekly_limit_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    monthly_limit_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE redeem_codes
  ADD COLUMN IF NOT EXISTS product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL;

ALTER TABLE usage_logs
  ADD COLUMN IF NOT EXISTS product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS product_subscription_id BIGINT REFERENCES user_product_subscriptions(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS group_debit_multiplier DECIMAL(10,4),
  ADD COLUMN IF NOT EXISTS product_debit_cost DECIMAL(20,10);
```

```go
field.Int64("product_id").Optional().Nillable()
field.Float("group_debit_multiplier").
    Optional().
    Nillable().
    SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"})
field.Float("product_debit_cost").
    Optional().
    Nillable().
    SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"})
```

- [ ] **Step 4: Run generation and schema verification**

Run: `cd backend && go generate ./ent && go test ./internal/repository -run TestMigrationsSharedSubscriptionProductsSchema -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/110_add_shared_subscription_products.sql \
  backend/ent/schema/subscription_product.go \
  backend/ent/schema/subscription_product_group.go \
  backend/ent/schema/user_product_subscription.go \
  backend/ent/schema/product_subscription_migration_source.go \
  backend/ent/schema/redeem_code.go \
  backend/ent/schema/usage_log.go \
  backend/internal/repository/migrations_schema_integration_test.go \
  backend/ent
git commit -m "feat: add shared subscription product schema"
```

### Task 2: Add repositories and service-layer product models

**Files:**
- Create: `backend/internal/repository/subscription_product_repo.go`
- Create: `backend/internal/repository/subscription_product_repo_integration_test.go`
- Create: `backend/internal/service/subscription_product.go`
- Create: `backend/internal/service/user_product_subscription.go`
- Create: `backend/internal/service/subscription_product_errors.go`
- Create: `backend/internal/service/subscription_product_errors_test.go`
- Create: `backend/internal/service/subscription_product_port.go`
- Create: `backend/internal/service/subscription_product_service.go`
- Create: `backend/internal/service/subscription_product_service_test.go`
- Modify: `backend/internal/repository/wire.go`
- Modify: `backend/internal/service/wire.go`
- Test: `backend/internal/repository/subscription_product_repo_integration_test.go`
- Test: `backend/internal/service/subscription_product_service_test.go`

- [ ] **Step 1: Write the failing repository and service tests**

Add integration and unit coverage for:

```go
func TestSubscriptionProductRepository_GetActiveProductByGroupID(t *testing.T) {}
func TestSubscriptionProductRepository_ListVisibleGroupsByUserID(t *testing.T) {}
func TestSubscriptionProductService_GetActiveProductSubscription(t *testing.T) {}
func TestSubscriptionProductService_ListVisibleGroups(t *testing.T) {}
func TestSubscriptionProductError_ProductLimitExceededIncludesMetadata(t *testing.T) {}
func TestSubscriptionProductError_ProductBindingInactiveIncludesMetadata(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/repository -run 'TestSubscriptionProductRepository_' -count=1 && go test ./internal/service -run 'Test(SubscriptionProductService_|SubscriptionProductError_)' -count=1`
Expected: FAIL because no repository, service, or product-mode runtime error helpers exist.

- [ ] **Step 3: Write minimal repository and service implementation**

Create focused ports and models:

```go
type SubscriptionProductBinding struct {
    ProductID         int64
    ProductCode       string
    ProductName       string
    GroupID           int64
    GroupName         string
    DebitMultiplier   float64
    ProductStatus     string
    BindingStatus     string
}

type ProductSubscriptionRepository interface {
    GetActiveProductBindingByGroupID(ctx context.Context, groupID int64) (*SubscriptionProductBinding, error)
    GetActiveUserProductSubscription(ctx context.Context, userID, productID int64) (*UserProductSubscription, error)
    ListVisibleGroupsByUserID(ctx context.Context, userID int64) ([]Group, error)
}
```

```go
func (s *SubscriptionProductService) GetActiveProductSubscription(ctx context.Context, userID, groupID int64) (*ProductSettlementContext, error) {
    binding, err := s.repo.GetActiveProductBindingByGroupID(ctx, groupID)
    if err != nil {
        return nil, err
    }
    sub, err := s.repo.GetActiveUserProductSubscription(ctx, userID, binding.ProductID)
    if err != nil {
        return nil, err
    }
    return &ProductSettlementContext{Binding: binding, Subscription: sub}, nil
}
```

```go
func NewProductBindingInactiveError(binding *SubscriptionProductBinding) error {
    return infraerrors.New(http.StatusForbidden, "PRODUCT_GROUP_BINDING_INACTIVE", "product group binding is inactive").WithMetadata(map[string]string{
        "product_id":       strconv.FormatInt(binding.ProductID, 10),
        "product_name":     binding.ProductName,
        "group_id":         strconv.FormatInt(binding.GroupID, 10),
        "group_name":       binding.GroupName,
        "debit_multiplier": strconv.FormatFloat(binding.DebitMultiplier, 'f', -1, 64),
    })
}

func NewProductLimitExceededError(binding *SubscriptionProductBinding, remainingBudget float64) error {
    return infraerrors.New(http.StatusForbidden, "PRODUCT_LIMIT_EXCEEDED", "product subscription limit exceeded").WithMetadata(map[string]string{
        "product_id":       strconv.FormatInt(binding.ProductID, 10),
        "product_name":     binding.ProductName,
        "group_id":         strconv.FormatInt(binding.GroupID, 10),
        "group_name":       binding.GroupName,
        "debit_multiplier": strconv.FormatFloat(binding.DebitMultiplier, 'f', -1, 64),
        "remaining_budget": strconv.FormatFloat(remainingBudget, 'f', 6, 64),
    })
}
```

Use the same file to define `NewSubscriptionProductNotFoundError`, `NewSubscriptionProductInactiveError`, `NewProductSubscriptionInvalidError`, and `NewProductMigrationConflictError` so every product-mode failure returns a stable `reason` and metadata envelope through `response.ErrorFrom`.

- [ ] **Step 4: Run focused repository and service verification**

Run: `cd backend && go test ./internal/repository -run 'TestSubscriptionProductRepository_' -count=1 && go test ./internal/service -run 'Test(SubscriptionProductService_|SubscriptionProductError_)' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/subscription_product_repo.go \
  backend/internal/repository/subscription_product_repo_integration_test.go \
  backend/internal/service/subscription_product.go \
  backend/internal/service/user_product_subscription.go \
  backend/internal/service/subscription_product_errors.go \
  backend/internal/service/subscription_product_errors_test.go \
  backend/internal/service/subscription_product_port.go \
  backend/internal/service/subscription_product_service.go \
  backend/internal/service/subscription_product_service_test.go \
  backend/internal/repository/wire.go \
  backend/internal/service/wire.go
git commit -m "feat: add shared subscription product repositories"
```

### Task 3: Extend the authoritative billing path for product settlement

**Files:**
- Create: `backend/internal/service/product_settlement.go`
- Create: `backend/internal/service/product_settlement_test.go`
- Modify: `backend/internal/service/usage_billing.go`
- Modify: `backend/internal/service/usage_log.go`
- Modify: `backend/internal/repository/usage_billing_repo.go`
- Modify: `backend/internal/repository/usage_billing_repo_integration_test.go`
- Modify: `backend/internal/repository/usage_log_repo.go`
- Modify: `backend/internal/repository/usage_log_repo_request_type_test.go`
- Modify: `backend/internal/service/gateway_service.go`
- Modify: `backend/internal/service/openai_gateway_service.go`
- Modify: `backend/internal/service/gateway_record_usage_test.go`
- Modify: `backend/internal/service/openai_gateway_record_usage_test.go`
- Test: `backend/internal/repository/usage_billing_repo_integration_test.go`
- Test: `backend/internal/service/openai_gateway_record_usage_test.go`

- [ ] **Step 1: Write the failing billing tests**

Add assertions for product-settled traffic:

```go
func TestUsageBillingRepositoryApply_ProductSubscriptionCostAdvancesProductWindows(t *testing.T) {}
func TestOpenAIGatewayRecordUsage_ProductSettlementLeavesLegacySubscriptionIDNil(t *testing.T) {}
func TestGatewayRecordUsage_ProductSettlementUsesProductDebitCost(t *testing.T) {}
func TestGatewayRecordUsage_ProductSettlementSharesQuotaAcrossTwoRealGroups(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/repository -run TestUsageBillingRepositoryApply_ProductSubscriptionCostAdvancesProductWindows -count=1 && go test ./internal/service -run 'TestOpenAIGatewayRecordUsage_ProductSettlement|TestGatewayRecordUsage_ProductSettlement' -count=1`
Expected: FAIL because billing commands and usage logs do not carry product-settlement fields yet.

- [ ] **Step 3: Write minimal product billing implementation**

Extend the billing command and usage log model:

```go
type UsageBillingCommand struct {
    // existing fields...
    ProductID             *int64
    ProductSubscriptionID *int64
    GroupID               *int64
    GroupDebitMultiplier  *float64
    StandardTotalCost     float64
    ProductDebitCost      float64
}
```

```go
if p.ProductSettlement != nil && p.ProductSettlement.Subscription != nil && p.Cost.TotalCost > 0 {
    cmd.ProductID = &p.ProductSettlement.Binding.ProductID
    cmd.ProductSubscriptionID = &p.ProductSettlement.Subscription.ID
    cmd.GroupID = &p.ProductSettlement.Binding.GroupID
    cmd.GroupDebitMultiplier = &p.ProductSettlement.Binding.DebitMultiplier
    cmd.StandardTotalCost = p.Cost.TotalCost
    cmd.ProductDebitCost = p.Cost.TotalCost * p.ProductSettlement.Binding.DebitMultiplier
}
```

```go
if cmd.ProductDebitCost > 0 && cmd.ProductSubscriptionID != nil {
    if err := incrementUsageBillingProductSubscription(ctx, tx, *cmd.ProductSubscriptionID, cmd.ProductDebitCost); err != nil {
        return err
    }
}
```

- [ ] **Step 4: Run focused billing verification**

Run: `cd backend && go test ./internal/repository -run 'TestUsageBillingRepositoryApply_(ProductSubscriptionCostAdvancesProductWindows|SubscriptionCostAdvancesDailyWindowAndConsumesCarryover)' -count=1 && go test ./internal/service -run 'Test(OpenAIGatewayRecordUsage_ProductSettlement|GatewayRecordUsage_ProductSettlement|GatewayRecordUsage_ProductSettlementSharesQuotaAcrossTwoRealGroups)' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/product_settlement.go \
  backend/internal/service/product_settlement_test.go \
  backend/internal/service/usage_billing.go \
  backend/internal/service/usage_log.go \
  backend/internal/repository/usage_billing_repo.go \
  backend/internal/repository/usage_billing_repo_integration_test.go \
  backend/internal/repository/usage_log_repo.go \
  backend/internal/repository/usage_log_repo_request_type_test.go \
  backend/internal/service/gateway_service.go \
  backend/internal/service/openai_gateway_service.go \
  backend/internal/service/gateway_record_usage_test.go \
  backend/internal/service/openai_gateway_record_usage_test.go
git commit -m "feat: settle shared subscription products through usage billing"
```

### Task 4: Authorize product-settled groups in key binding and auth middleware

**Files:**
- Modify: `backend/internal/service/api_key_service.go`
- Modify: `backend/internal/service/admin_service.go`
- Modify: `backend/internal/service/billing_cache_service.go`
- Modify: `backend/internal/server/middleware/api_key_auth.go`
- Modify: `backend/internal/server/middleware/api_key_auth_test.go`
- Modify: `backend/internal/service/admin_service_apikey_test.go`
- Modify: `backend/internal/service/billing_cache_service_subscription_test.go`
- Modify: `backend/internal/service/subscription_product_service.go`
- Test: `backend/internal/service/admin_service_apikey_test.go`
- Test: `backend/internal/server/middleware/api_key_auth_test.go`

- [ ] **Step 1: Write the failing authorization tests**

Add tests that cover user create, middleware auth, and admin rebind:

```go
func TestAPIKeyService_Create_ProductSettledGroupRequiresProductSubscription(t *testing.T) {}
func TestAdminService_AdminUpdateAPIKeyGroupID_ProductSettledGroupAllowsActiveProductSubscription(t *testing.T) {}
func TestAPIKeyAuth_ProductSettledGroupLoadsProductContext(t *testing.T) {}
func TestAPIKeyAuth_ProductSettledGroupReturnsStructuredRuntimeError(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/service -run 'Test(APIKeyService_Create_ProductSettledGroupRequiresProductSubscription|AdminService_AdminUpdateAPIKeyGroupID_ProductSettledGroupAllowsActiveProductSubscription)' -count=1 && go test ./internal/server/middleware -run 'TestAPIKeyAuth_ProductSettledGroup(LoadsProductContext|ReturnsStructuredRuntimeError)' -count=1`
Expected: FAIL because all binding paths still require legacy `user_subscriptions` and the middleware does not yet return stable product-mode reasons and metadata.

- [ ] **Step 3: Write minimal authorization implementation**

Introduce one shared eligibility helper:

```go
func (s *APIKeyService) canUserBindGroup(ctx context.Context, user *User, group *Group) bool {
    if !group.IsSubscriptionType() {
        return user.CanBindGroup(group.ID, group.IsExclusive)
    }
    if productCtx, err := s.subscriptionProductService.GetActiveProductSubscription(ctx, user.ID, group.ID); err == nil && productCtx != nil {
        return true
    }
    _, err := s.userSubRepo.GetActiveByUserIDAndGroupID(ctx, user.ID, group.ID)
    return err == nil
}
```

```go
if productCtx != nil {
    c.Set(string(ContextKeyProductSubscription), productCtx.Subscription)
    c.Set(string(ContextKeySubscriptionProduct), productCtx.Binding)
}
```

```go
binding := productCtx.Binding
if binding.BindingStatus != "active" {
    return NewProductBindingInactiveError(binding)
}
```

Ensure every service path that writes `api_keys.group_id` uses the same helper, including the normal user create flow and admin rebind or import-style paths already routed through `admin_service.go`.

- [ ] **Step 4: Run focused auth verification**

Run: `cd backend && go test ./internal/service -run 'Test(APIKeyService_Create_ProductSettledGroupRequiresProductSubscription|AdminService_AdminUpdateAPIKeyGroupID_ProductSettledGroupAllowsActiveProductSubscription)' -count=1 && go test ./internal/server/middleware -run 'TestAPIKeyAuth_ProductSettledGroup(LoadsProductContext|ReturnsStructuredRuntimeError)' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/api_key_service.go \
  backend/internal/service/admin_service.go \
  backend/internal/service/billing_cache_service.go \
  backend/internal/server/middleware/api_key_auth.go \
  backend/internal/server/middleware/api_key_auth_test.go \
  backend/internal/service/admin_service_apikey_test.go \
  backend/internal/service/billing_cache_service_subscription_test.go \
  backend/internal/service/subscription_product_service.go
git commit -m "feat: authorize api keys with product subscriptions"
```

### Task 5: Support product redeem codes, default assignment, and admin product service operations

**Files:**
- Create: `backend/internal/service/subscription_product_admin.go`
- Create: `backend/internal/service/subscription_product_admin_test.go`
- Modify: `backend/internal/service/redeem_service.go`
- Modify: `backend/internal/service/domain_constants.go`
- Modify: `backend/internal/service/setting_service.go`
- Modify: `backend/internal/service/setting_service_update_test.go`
- Modify: `backend/internal/service/auth_service_register_test.go`
- Modify: `backend/internal/handler/admin/redeem_handler.go`
- Modify: `backend/internal/handler/admin/redeem_handler_test.go`
- Modify: `backend/internal/handler/admin/setting_handler.go`
- Modify: `backend/internal/handler/dto/settings.go`
- Test: `backend/internal/service/subscription_product_admin_test.go`
- Test: `backend/internal/handler/admin/redeem_handler_test.go`

- [ ] **Step 1: Write the failing admin and redeem tests**

Add test coverage such as:

```go
func TestRedeemService_Redeem_ProductSubscriptionCodeCreatesUserProductSubscription(t *testing.T) {}
func TestAdminRedeemHandler_Generate_ProductIDXorGroupID(t *testing.T) {}
func TestSettingService_UpdateSettings_PersistsDefaultProductSubscriptions(t *testing.T) {}
func TestRegisterUser_AppliesDefaultProductSubscriptions(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/service -run 'Test(RedeemService_Redeem_ProductSubscriptionCodeCreatesUserProductSubscription|SettingService_UpdateSettings_PersistsDefaultProductSubscriptions|RegisterUser_AppliesDefaultProductSubscriptions)' -count=1 && go test ./internal/handler/admin -run TestAdminRedeemHandler_Generate_ProductIDXorGroupID -count=1`
Expected: FAIL because redeem, settings, and registration only understand legacy `group_id` subscription grants.

- [ ] **Step 3: Write minimal product admin implementation**

Add XOR validation and product default settings:

```go
type GenerateRedeemCodesRequest struct {
    Count        int    `json:"count"`
    Type         string `json:"type"`
    GroupID      *int64 `json:"group_id"`
    ProductID    *int64 `json:"product_id"`
    ValidityDays int    `json:"validity_days"`
}

if req.Type == "subscription" && ((req.GroupID == nil) == (req.ProductID == nil)) {
    response.BadRequest(c, "exactly one of group_id or product_id is required for subscription type")
    return
}
```

```go
const SettingKeyDefaultSubscriptionProducts = "default_subscription_products"

type DefaultSubscriptionProductSetting struct {
    ProductID    int64 `json:"product_id"`
    ValidityDays int   `json:"validity_days"`
}
```

- [ ] **Step 4: Run focused redeem and settings verification**

Run: `cd backend && go test ./internal/service -run 'Test(RedeemService_Redeem_ProductSubscriptionCodeCreatesUserProductSubscription|SettingService_UpdateSettings_PersistsDefaultProductSubscriptions|RegisterUser_AppliesDefaultProductSubscriptions)' -count=1 && go test ./internal/handler/admin -run 'TestAdminRedeemHandler_Generate_ProductIDXorGroupID' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/subscription_product_admin.go \
  backend/internal/service/subscription_product_admin_test.go \
  backend/internal/service/redeem_service.go \
  backend/internal/service/domain_constants.go \
  backend/internal/service/setting_service.go \
  backend/internal/service/setting_service_update_test.go \
  backend/internal/service/auth_service_register_test.go \
  backend/internal/handler/admin/redeem_handler.go \
  backend/internal/handler/admin/redeem_handler_test.go \
  backend/internal/handler/admin/setting_handler.go \
  backend/internal/handler/dto/settings.go
git commit -m "feat: support product grants and defaults"
```

### Task 6: Add user and admin HTTP product endpoints and wire them

**Files:**
- Create: `backend/internal/handler/dto/subscription_product.go`
- Create: `backend/internal/handler/subscription_product_handler.go`
- Create: `backend/internal/handler/subscription_product_handler_test.go`
- Create: `backend/internal/handler/admin/subscription_product_handler.go`
- Create: `backend/internal/handler/admin/subscription_product_handler_test.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/server/routes/admin.go`
- Modify: `backend/cmd/server/wire.go`
- Test: `backend/internal/handler/subscription_product_handler_test.go`
- Test: `backend/internal/handler/admin/subscription_product_handler_test.go`

- [ ] **Step 1: Write the failing handler tests**

Add contract tests for:

```go
func TestSubscriptionProductHandler_GetActive(t *testing.T) {}
func TestSubscriptionProductHandler_GetSummary(t *testing.T) {}
func TestAdminSubscriptionProductHandler_Create(t *testing.T) {}
func TestAdminSubscriptionProductHandler_UpdateBindings(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/handler -run 'TestSubscriptionProductHandler_' -count=1 && go test ./internal/handler/admin -run 'TestAdminSubscriptionProductHandler_' -count=1`
Expected: FAIL because no product handlers or routes exist.

- [ ] **Step 3: Write minimal handler and route implementation**

Register the new surfaces:

```go
products := authenticated.Group("/subscription-products")
{
    products.GET("/active", h.SubscriptionProduct.GetActive)
    products.GET("/summary", h.SubscriptionProduct.GetSummary)
    products.GET("/progress", h.SubscriptionProduct.GetProgress)
}
```

```go
products := admin.Group("/subscription-products")
{
    products.GET("", h.Admin.SubscriptionProduct.List)
    products.POST("", h.Admin.SubscriptionProduct.Create)
    products.PUT("/:id", h.Admin.SubscriptionProduct.Update)
    products.PUT("/:id/bindings", h.Admin.SubscriptionProduct.SyncBindings)
    products.GET("/:id/subscriptions", h.Admin.SubscriptionProduct.ListSubscriptions)
}
```

- [ ] **Step 4: Run focused handler verification**

Run: `cd backend && go test ./internal/handler -run 'TestSubscriptionProductHandler_' -count=1 && go test ./internal/handler/admin -run 'TestAdminSubscriptionProductHandler_' -count=1 && go generate ./cmd/server`
Expected: PASS and regenerated wire output without missing providers.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/dto/subscription_product.go \
  backend/internal/handler/subscription_product_handler.go \
  backend/internal/handler/subscription_product_handler_test.go \
  backend/internal/handler/admin/subscription_product_handler.go \
  backend/internal/handler/admin/subscription_product_handler_test.go \
  backend/internal/handler/handler.go \
  backend/internal/handler/wire.go \
  backend/internal/server/routes/user.go \
  backend/internal/server/routes/admin.go \
  backend/cmd/server/wire.go \
  backend/cmd/server/wire_gen.go
git commit -m "feat: expose subscription product api surfaces"
```

### Task 7: Switch user-facing frontend flows to product subscriptions

**Files:**
- Create: `frontend/src/api/subscriptionProducts.ts`
- Create: `frontend/src/stores/subscriptionProducts.ts`
- Modify: `frontend/src/stores/index.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/views/user/SubscriptionsView.vue`
- Modify: `frontend/src/components/common/SubscriptionProgressMini.vue`
- Modify: `frontend/src/views/user/KeysView.vue`
- Modify: `frontend/src/components/common/GroupSelector.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/components/__tests__/ApiKeyCreate.spec.ts`
- Modify: `frontend/src/components/common/__tests__/SubscriptionProgressMini.spec.ts`
- Create: `frontend/src/stores/__tests__/subscriptionProducts.spec.ts`
- Test: `frontend/src/stores/__tests__/subscriptionProducts.spec.ts`
- Test: `frontend/src/components/common/__tests__/SubscriptionProgressMini.spec.ts`

- [ ] **Step 1: Write the failing frontend tests**

Add tests that assert:

```ts
it('renders one product card with multiple groups', async () => {})
it('offers product-settled groups in api key creation', async () => {})
it('loads subscription products into the store', async () => {})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npm run test:run -- src/stores/__tests__/subscriptionProducts.spec.ts src/components/common/__tests__/SubscriptionProgressMini.spec.ts src/components/__tests__/ApiKeyCreate.spec.ts`
Expected: FAIL because the product API, store, and UI do not exist yet.

- [ ] **Step 3: Write minimal product frontend implementation**

Introduce product types and store actions:

```ts
export interface ActiveSubscriptionProduct {
  product_id: number
  code: string
  name: string
  expires_at: string | null
  monthly_usage_usd: number
  monthly_limit_usd: number | null
  daily_carryover_in_usd: number
  daily_carryover_remaining_usd: number
  groups: Array<{
    group_id: number
    group_name: string
    debit_multiplier: number
  }>
}
```

```ts
export const useSubscriptionProductStore = defineStore('subscriptionProducts', () => {
  const items = ref<ActiveSubscriptionProduct[]>([])
  async function fetchActive() {
    items.value = await subscriptionProductsAPI.getActive()
  }
  return { items, fetchActive }
})
```

- [ ] **Step 4: Run focused frontend verification**

Run: `cd frontend && npm run test:run -- src/stores/__tests__/subscriptionProducts.spec.ts src/components/common/__tests__/SubscriptionProgressMini.spec.ts src/components/__tests__/ApiKeyCreate.spec.ts && npm run typecheck`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/subscriptionProducts.ts \
  frontend/src/stores/subscriptionProducts.ts \
  frontend/src/stores/index.ts \
  frontend/src/types/index.ts \
  frontend/src/views/user/SubscriptionsView.vue \
  frontend/src/components/common/SubscriptionProgressMini.vue \
  frontend/src/views/user/KeysView.vue \
  frontend/src/components/common/GroupSelector.vue \
  frontend/src/router/index.ts \
  frontend/src/components/__tests__/ApiKeyCreate.spec.ts \
  frontend/src/components/common/__tests__/SubscriptionProgressMini.spec.ts \
  frontend/src/stores/__tests__/subscriptionProducts.spec.ts
git commit -m "feat: add shared subscription product user ui"
```

### Task 8: Add admin product management UI plus settings and redeem updates

**Files:**
- Create: `frontend/src/api/admin/subscriptionProducts.ts`
- Create: `frontend/src/views/admin/SubscriptionProductsView.vue`
- Modify: `frontend/src/api/admin/index.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/views/admin/RedeemView.vue`
- Modify: `frontend/src/views/admin/SettingsView.vue`
- Modify: `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- Create: `frontend/src/views/admin/__tests__/SubscriptionProductsView.spec.ts`
- Test: `frontend/src/views/admin/__tests__/SubscriptionProductsView.spec.ts`

- [ ] **Step 1: Write the failing admin frontend tests**

Add coverage for:

```ts
it('creates and lists subscription products', async () => {})
it('edits product-group multipliers', async () => {})
it('persists default product subscriptions from settings', async () => {})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npm run test:run -- src/views/admin/__tests__/SubscriptionProductsView.spec.ts src/views/admin/__tests__/SettingsView.spec.ts`
Expected: FAIL because no admin product UI or settings fields exist.

- [ ] **Step 3: Write minimal admin frontend implementation**

Create focused admin APIs:

```ts
export async function listProducts() {
  const { data } = await apiClient.get('/admin/subscription-products')
  return data
}

export async function syncBindings(id: number, bindings: ProductGroupBindingInput[]) {
  const { data } = await apiClient.put(`/admin/subscription-products/${id}/bindings`, { bindings })
  return data
}
```

```ts
form.default_subscription_products = Array.isArray(settings.default_subscription_products)
  ? settings.default_subscription_products
  : []
```

- [ ] **Step 4: Run focused admin frontend verification**

Run: `cd frontend && npm run test:run -- src/views/admin/__tests__/SubscriptionProductsView.spec.ts src/views/admin/__tests__/SettingsView.spec.ts && npm run typecheck`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/admin/subscriptionProducts.ts \
  frontend/src/views/admin/SubscriptionProductsView.vue \
  frontend/src/api/admin/index.ts \
  frontend/src/types/index.ts \
  frontend/src/router/index.ts \
  frontend/src/views/admin/RedeemView.vue \
  frontend/src/views/admin/SettingsView.vue \
  frontend/src/views/admin/__tests__/SettingsView.spec.ts \
  frontend/src/views/admin/__tests__/SubscriptionProductsView.spec.ts
git commit -m "feat: add shared subscription product admin ui"
```

### Task 9: Add migration tools, validation tooling, and rollout artifacts

**Files:**
- Create: `backend/cmd/shared-subscription-products-backfill/main.go`
- Create: `backend/cmd/shared-subscription-products-backfill/main_test.go`
- Create: `backend/cmd/shared-subscription-products-validate/main.go`
- Create: `docs/superpowers/plans/2026-04-25-shared-subscription-products-rollout.md`
- Test: `backend/cmd/shared-subscription-products-backfill/main_test.go`

- [ ] **Step 1: Write the failing migration-tool tests or dry-run contract checks**

Add lightweight command-level checks such as:

```go
func TestBuildBackfillReport_SkipsDuplicateActiveProductSubscriptions(t *testing.T) {}
func TestBuildBackfillReport_RequiresDeterministicLegacySource(t *testing.T) {}
func TestApplyBackfill_IdempotentWhenMigrationBatchReruns(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./cmd/shared-subscription-products-backfill -run 'Test(BuildBackfillReport_|ApplyBackfill_IdempotentWhenMigrationBatchReruns)' -count=1`
Expected: FAIL because the backfill and validation tools do not exist yet.

- [ ] **Step 3: Write minimal tooling and rollout doc implementation**

Implement dry-run/apply flags and validation queries:

```go
type BackfillOptions struct {
    DryRun         bool
    ProductCode    string
    SourceGroupIDs []int64
    MigrationBatch string
}

func main() {
    // load config, connect DB, render report, optionally apply transactionally
}
```

```sql
SELECT ups.user_id, ups.product_id, COUNT(*)
FROM user_product_subscriptions ups
WHERE ups.deleted_at IS NULL
GROUP BY ups.user_id, ups.product_id
HAVING COUNT(*) > 1;
```

- [ ] **Step 4: Run tool verification**

Run: `cd backend && go test ./cmd/shared-subscription-products-backfill -run 'Test(BuildBackfillReport_|ApplyBackfill_IdempotentWhenMigrationBatchReruns)' -count=1 && go run ./cmd/shared-subscription-products-backfill --dry-run --product-code gpt_monthly --source-group-ids 88 --migration-batch local-check`
Expected: PASS for tests; dry-run prints counts, conflicts, and sample rows without writing data.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/shared-subscription-products-backfill/main.go \
  backend/cmd/shared-subscription-products-backfill/main_test.go \
  backend/cmd/shared-subscription-products-validate/main.go \
  docs/superpowers/plans/2026-04-25-shared-subscription-products-rollout.md
git commit -m "docs: add shared subscription product rollout tooling"
```

### Task 10: Run full-stack verification before rollout

**Files:**
- Modify: `docs/superpowers/plans/2026-04-25-shared-subscription-products-rollout.md`
- Test: `backend/internal/repository/usage_billing_repo_integration_test.go`
- Test: `backend/internal/service/openai_gateway_record_usage_test.go`
- Test: `backend/internal/service/admin_service_apikey_test.go`
- Test: `frontend/src/components/common/__tests__/SubscriptionProgressMini.spec.ts`

- [ ] **Step 1: Assemble the verification command set**

Use a single checklist in the rollout doc:

```bash
cd backend && go test ./internal/repository -run 'Test(SubscriptionProductRepository_|UsageBillingRepositoryApply_ProductSubscriptionCostAdvancesProductWindows)' -count=1
cd backend && go test ./internal/service -run 'Test(OpenAIGatewayRecordUsage_ProductSettlement|APIKeyService_Create_ProductSettledGroupRequiresProductSubscription|AdminService_AdminUpdateAPIKeyGroupID_ProductSettledGroupAllowsActiveProductSubscription)' -count=1
cd backend && go test ./internal/server/middleware -run 'TestAPIKeyAuth_ProductSettledGroup(LoadsProductContext|ReturnsStructuredRuntimeError)' -count=1
cd backend && go test ./internal/handler -run 'TestSubscriptionProductHandler_' -count=1
cd frontend && npm run test:run -- src/stores/__tests__/subscriptionProducts.spec.ts src/components/common/__tests__/SubscriptionProgressMini.spec.ts src/components/__tests__/ApiKeyCreate.spec.ts src/views/admin/__tests__/SubscriptionProductsView.spec.ts
cd frontend && npm run typecheck
```

- [ ] **Step 2: Run the backend verification set**

Run the three backend commands above.
Expected: PASS

- [ ] **Step 3: Run the frontend verification set**

Run the two frontend commands above.
Expected: PASS

- [ ] **Step 4: Update the rollout doc with the exact passing command set**

Paste the commands and required green checks into `docs/superpowers/plans/2026-04-25-shared-subscription-products-rollout.md`.

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/plans/2026-04-25-shared-subscription-products-rollout.md
git commit -m "docs: finalize shared subscription product verification checklist"
```
