# Product Subscription Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the xlabapi product-subscription billing loop with strict product redeem codes, explicit product-family selection, user-selected balance fallback, Beijing daily windows, rolling subscription windows, and complete admin operations.

**Architecture:** Add two nullable data fields, then route all product-subscription decisions through service/repository APIs that understand product family and user-selected fallback groups. Keep legacy subscriptions available for legacy flows, but remove implicit fallback from failed product-subscription resolution. Update frontend forms to match backend invariants.

**Tech Stack:** Go backend with Gin, ent-generated schema plus SQL migrations, repository/service layering, Vue 3 frontend, Vitest, Go unit/integration tests.

---

### Task 1: Migrations And Data Shapes

**Files:**
- Create: `backend/migrations/145_product_subscription_explicit_fallback_family.sql`
- Modify: `backend/ent/schema/user.go`
- Modify: `backend/ent/schema/api_key.go`
- Modify: generated ent files after `go generate ./ent`
- Modify: `backend/internal/service/user.go`
- Modify: `backend/internal/service/api_key.go`
- Modify: `backend/internal/handler/dto/types.go`
- Modify: `frontend/src/types/index.ts`

- [ ] Add nullable `users.subscription_balance_fallback_group_id`.
- [ ] Add nullable `api_keys.subscription_product_family`.
- [ ] Regenerate ent code.
- [ ] Expose the fields in service and DTO types.

### Task 2: Redeem Code Product Enforcement

**Files:**
- Modify: `backend/internal/service/redeem_product_subscription_test.go`
- Modify: `backend/internal/service/redeem_service.go`
- Modify: `backend/internal/service/admin_service_redeem_product_test.go`
- Modify: `backend/internal/service/admin_service.go`
- Modify: `backend/internal/handler/admin/redeem_handler.go`
- Modify: `frontend/src/views/admin/RedeemView.vue`

- [ ] Write failing tests for product redeem code redemption without `product_id`.
- [ ] Write failing tests for admin product redeem code generation without `product_id`.
- [ ] Reject product subscription redeem codes without `product_id`.
- [ ] Reject admin generation requests for product subscription codes without product selection.
- [ ] Update admin redeem UI validation so product subscription type requires a product.

### Task 3: Product Binding Validation

**Files:**
- Modify: `backend/internal/repository/subscription_product_repo_integration_test.go`
- Modify: `backend/internal/repository/subscription_product_repo.go`
- Modify: `backend/internal/service/subscription_product_errors.go`
- Modify: `frontend/src/views/admin/SubscriptionProductConfigView.vue`

- [ ] Write failing integration test showing a standard group cannot be bound to a product.
- [ ] Add repository validation for group type before syncing product bindings.
- [ ] Surface a clear error to the admin UI.
- [ ] Filter binding picker to subscription groups where frontend data is available.

### Task 4: Product Family Resolution

**Files:**
- Modify: `backend/internal/repository/subscription_product_repo_integration_test.go`
- Modify: `backend/internal/repository/subscription_product_repo.go`
- Modify: `backend/internal/service/subscription_product_service.go`
- Modify: `backend/internal/service/subscription_product.go`
- Modify: `backend/internal/service/api_key_service.go`
- Modify: `backend/internal/handler/api_key_handler.go`
- Modify: `backend/internal/handler/dto/mappers.go`
- Modify: `frontend/src/views/user/ApiKeysView.vue` or current key management component
- Modify: `frontend/src/api/apiKeys.ts`
- Modify: `frontend/src/types/index.ts`

- [ ] Write failing tests for ambiguous multiple product families.
- [ ] Write failing tests for explicit API key product family.
- [ ] Change active product lookup to accept optional family.
- [ ] Return ambiguity when multiple families match and no family is specified.
- [ ] Add API key create/edit support for `subscription_product_family`.
- [ ] Add frontend family selection when a selected subscription group has multiple families.

### Task 5: User-Selected Balance Fallback

**Files:**
- Modify: `backend/internal/server/middleware/api_key_auth_subscription_quota_test.go`
- Modify: `backend/internal/server/middleware/api_key_auth.go`
- Modify: `backend/internal/service/user_service_test.go`
- Modify: `backend/internal/service/user_service.go`
- Modify: `backend/internal/repository/user_repo.go`
- Modify: `backend/internal/service/api_key_service_available_groups_test.go`
- Modify: `frontend/src/views/user/SubscriptionsView.vue`
- Modify: `frontend/src/api/user.ts`
- Modify: `frontend/src/types/index.ts`

- [ ] Write failing middleware tests for fallback disabled, selected fallback group, and unauthorized fallback group.
- [ ] Add user profile update support for fallback group ID.
- [ ] Resolve fallback group from user settings, not group mapping.
- [ ] Check standard group authorization before fallback.
- [ ] Block fallback if user balance is already negative.
- [ ] Update user subscription UI to require group and positive limit when enabling.

### Task 6: Product Quota And Balance Settlement

**Files:**
- Modify: `backend/internal/repository/usage_billing_repo_integration_test.go`
- Modify: `backend/internal/repository/usage_billing_repo.go`
- Modify: `backend/internal/repository/subscription_usage_sql.go`
- Modify: `backend/internal/service/gateway_service_subscription_billing_test.go`
- Modify: `backend/internal/service/gateway_service.go`

- [ ] Write failing tests for product quota not going negative.
- [ ] Write failing tests for fallback balance going negative but next request being blocked.
- [ ] Ensure product split debit rejects insufficient product quota.
- [ ] Keep balance fallback cumulative limit strict.
- [ ] Allow balance deduction to go negative for settled fallback and standard balance requests.
- [ ] Ensure pre-request auth blocks negative-balance users.

### Task 7: Beijing Daily Windows And Rolling Weekly/Monthly Windows

**Files:**
- Modify: `backend/internal/service/subscription_product_service_test.go`
- Modify: `backend/internal/repository/usage_billing_repo_integration_test.go`
- Modify: `backend/internal/repository/subscription_usage_sql.go`
- Modify: `backend/internal/service/subscription_product_service.go`
- Modify: `frontend/src/views/user/SubscriptionsView.vue`
- Modify: `frontend/src/views/admin/SubscriptionProductsView.vue`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`

- [ ] Write failing tests for Beijing 00:00 daily reset and carryover.
- [ ] Write failing tests for rolling 7-day weekly and 30-day monthly reset from instance window start.
- [ ] Use a Beijing fixed zone helper for daily reset windows.
- [ ] Keep weekly/monthly windows rolling from existing window starts.
- [ ] Update frontend copy to state Beijing daily reset and rolling 7/30-day windows.

### Task 8: Admin Product Subscription Operations

**Files:**
- Modify: `backend/internal/handler/admin/subscription_product_handler.go`
- Modify: `backend/internal/server/routes/admin.go`
- Modify: `backend/internal/service/subscription_product_service.go`
- Modify: `backend/internal/repository/subscription_product_repo.go`
- Modify: `backend/internal/handler/subscription_product_handler_test.go`
- Modify: `frontend/src/api/admin/subscriptionProducts.ts`
- Modify: `frontend/src/views/admin/SubscriptionProductsView.vue`

- [ ] Write failing handler/service tests for adjust, reset quota, and revoke.
- [ ] Implement repository update methods.
- [ ] Implement service methods with validation.
- [ ] Register routes.
- [ ] Remove `daily_limit_usd` from adjust request and dialog.
- [ ] Confirm buttons call implemented endpoints.

### Task 9: Admin User Fallback Management

**Files:**
- Modify: `backend/internal/handler/admin/user_handler.go`
- Modify: `backend/internal/service/admin_service.go`
- Modify: `backend/internal/service/admin_service_list_users_test.go`
- Modify: `frontend/src/views/admin/UsersView.vue` or current admin user editor
- Modify: `frontend/src/api/admin/users.ts`
- Modify: `frontend/src/types/index.ts`

- [ ] Write failing tests for admin setting fallback enabled, limit, used amount, and group.
- [ ] Add backend admin update support.
- [ ] Add frontend controls in the existing admin user edit surface.
- [ ] Allow admin reset of fallback used amount.

### Task 10: Documentation And Verification

**Files:**
- Modify: `docs/PRODUCT_SUBSCRIPTIONS_CN.md`
- Modify: `docs/superpowers/specs/2026-05-04-product-subscription-closure-design.md` if behavior changes during implementation

- [ ] Update long-form product subscription docs.
- [ ] Run targeted backend tests for changed packages.
- [ ] Run frontend typecheck and relevant Vitest tests.
- [ ] Run `git status --short`.
- [ ] Record any residual risks in the final response.
