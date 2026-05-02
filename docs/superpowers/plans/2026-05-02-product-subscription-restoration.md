# Product Subscription Restoration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore product subscriptions so users see product-level subscriptions, can use all product groups, and shared quota/carryover stays correct.

**Architecture:** Rebuild the product layer around the existing restored runtime primitives. Add focused repository/service/handler methods for active product subscriptions, then update the frontend to render product cards alongside legacy subscriptions.

**Tech Stack:** Go, Gin, ent SQL access, Vue 3, Pinia-compatible API modules, Vitest, testify

---

## Emergency Batch

### Task 1: User Product Subscription API

**Files:**
- Modify: `backend/internal/service/subscription_product.go`
- Modify: `backend/internal/repository/subscription_product_repo.go`
- Modify: `backend/internal/service/subscription_product_service.go`
- Create: `backend/internal/handler/dto/subscription_product.go`
- Create: `backend/internal/handler/subscription_product_handler.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`
- Modify: `backend/internal/server/routes/user.go`
- Test: `backend/internal/service/subscription_product_service_test.go`
- Test: `backend/internal/handler/subscription_product_handler_test.go`

- [ ] Add failing service test for one active product with two groups.
- [ ] Add failing handler test for `GET /api/v1/subscription-products/active`.
- [ ] Implement repository query grouped by `user_product_subscriptions`.
- [ ] Implement service method `ListActiveProducts`.
- [ ] Implement DTO mapper and handler.
- [ ] Register route.
- [ ] Run focused backend tests.

### Task 2: User Subscription Page Product Cards

**Files:**
- Create: `frontend/src/api/subscriptionProducts.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/views/user/SubscriptionsView.vue`
- Test: `frontend/src/views/user/__tests__/SubscriptionsView.spec.ts`

- [ ] Add failing frontend test for rendering one product card with two groups.
- [ ] Add product subscription API client and types.
- [ ] Load legacy subscriptions and product subscriptions together.
- [ ] Render product cards before legacy cards.
- [ ] Show product limits, carryover, and group multipliers.
- [ ] Run focused frontend tests and typecheck if available.

### Task 3: API Key Product Group Visibility

**Files:**
- Modify: `backend/internal/repository/subscription_product_repo.go`
- Modify: `backend/internal/service/subscription_product_service.go`
- Modify: API key/group option provider used by `frontend/src/views/user/KeysView.vue`
- Test: relevant backend API key/group tests

- [ ] Add failing test showing active product subscription exposes all product groups.
- [ ] Add product-expanded visible group method.
- [ ] Merge product groups with existing legacy visible groups.
- [ ] Run focused tests.

### Task 4: Carryover and Runtime Verification

**Files:**
- Modify only if tests reveal a defect:
  - `backend/internal/repository/subscription_usage_sql.go`
  - `backend/internal/service/subscription_product_service.go`
  - `backend/internal/service/product_settlement.go`

- [ ] Run existing product billing and carryover tests.
- [ ] Add missing one-day-only product carryover test if current tests do not cover third-day expiry.
- [ ] Fix only proven defects.

## Follow-up Batch

Restore admin product management, product redeem codes, and default product grants after the emergency user-facing batch is green.
