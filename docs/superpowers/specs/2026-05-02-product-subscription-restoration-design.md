# Product Subscription Restoration Design

**Date:** 2026-05-02

## Goal

Restore the full xlabapi product-subscription model after the upstream migration: one purchased product unlocks multiple real groups, API keys still bind to one real group, all product groups debit one shared product quota pool, and daily carryover remains available for only the next day.

## Incident Scope

The current branch restored part of the runtime settlement path, but the product-subscription feature is incomplete. The user-visible subscription page still renders legacy group subscriptions, product-level user/admin APIs are missing, API key creation does not have a clean product-expanded group source, and redeem/default-grant flows still target only legacy group subscriptions.

## Recovery Strategy

Use the 2026-04-25/2026-04-26 xlabapi implementation as the behavioral baseline and restore it in small verified batches.

Batch 1 is the emergency user-facing repair:

- Add user-facing product subscription API responses.
- Render product subscription cards on the user subscription page.
- Ensure API key group choices include groups unlocked by active product subscriptions.
- Preserve product runtime settlement and daily carryover semantics.

Batch 2 restores operator surfaces:

- Admin product management endpoints and UI.
- Product subscription redeem codes.
- Default product subscription grants during registration.

Batch 3 hardens migration and verification:

- Replace hard-coded migration assumptions with safer reports where practical.
- Add validation tests and operational checks for shared pools, multipliers, and one-day carryover.

## Required Semantics

- A product owns shared daily, weekly, and monthly quota state in `user_product_subscriptions`.
- A product may expose multiple real groups through `subscription_product_groups`.
- Each real group has a `debit_multiplier`.
- API keys bind to real `group_id`, never to product id.
- Product usage writes `product_id`, `product_subscription_id`, `group_debit_multiplier`, and `product_debit_cost`.
- Product traffic must not also debit legacy `user_subscriptions`.
- Daily carryover is generated only from yesterday's unused fresh daily quota.
- Carryover is consumed before today's fresh quota.
- Carryover expires after one day and never rolls into the third day.

## First Acceptance Gate

The emergency batch is acceptable when:

- `GET /api/v1/subscription-products/active` returns active product subscriptions with groups, multipliers, limits, usage, and carryover fields.
- The user subscription page shows one product card with multiple real groups instead of only legacy group cards.
- Existing legacy group subscriptions still render.
- Product-mode usage billing tests pass.
- Carryover tests pass for product subscriptions.
