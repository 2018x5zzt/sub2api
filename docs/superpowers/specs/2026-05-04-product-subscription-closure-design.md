# Product Subscription Closure Design

Date: 2026-05-04

## Goal

Make the xlabapi product subscription module closed-loop across database, backend services, gateway enforcement, admin tools, user settings, and frontend workflows.

The target behavior is strict product-subscription billing with explicit user-controlled balance fallback. Product subscriptions must not silently fall back to legacy subscriptions, implicit product guesses, or admin-chosen balance groups.

## Decisions

1. Product subscription redeem codes must carry `product_id`. If a subscription redeem code has no `product_id`, redemption fails with a configuration error.
2. API keys bound to subscription groups use the new product-subscription path. If no active product subscription can be resolved, the request fails.
3. Legacy `user_subscriptions` remain supported only for legacy groups and legacy flows. They are not used as an implicit fallback for failed product-subscription resolution.
4. Balance fallback is user-explicit and default off. A user must enable it, choose an authorized standard balance group, and set a positive cumulative limit.
5. Admins may manage a user's balance fallback settings, including enabled state, limit, used amount, and selected fallback group.
6. Daily product windows and daily carryover use Beijing time, UTC+8, with reset at 00:00 Beijing time.
7. Weekly and monthly product windows are rolling windows from the subscription instance window start: 7 days and 30 days. They are not natural calendar week/month windows.
8. Product quota may not go negative. If product quota is insufficient and balance fallback is disabled, the request is rejected before upstream work when the system can determine the shortage.
9. Balance billing may take a user's balance negative for the current settled request, but future requests are blocked while the balance is negative.
10. Multiple `product_family` values for the same user and real group require explicit selection. API keys bind an optional product family. If a key does not bind one and multiple families match, the request fails with a clear ambiguity error.
11. Product bindings may only target active subscription-type groups. Backend validation rejects standard groups.
12. Admin product-subscription list actions must have backend support: adjust, reset quota, revoke.

## Data Model

Add fields:

- `users.subscription_balance_fallback_group_id`: nullable FK-like value pointing to a standard group selected by the user or admin.
- `api_keys.subscription_product_family`: nullable string. When set, product-subscription resolution is limited to this family.

Retain existing fields:

- `users.subscription_balance_fallback_enabled`
- `users.subscription_balance_fallback_limit_usd`
- `users.subscription_balance_fallback_used_usd`
- `subscription_products.product_family`
- `groups.balance_fallback_group_id`

`groups.balance_fallback_group_id` remains as legacy configuration data but will not be used for new user-controlled fallback decisions.

## Product Subscription Resolution

Given a user and API key group:

1. If the API key group is not subscription type, use standard balance billing.
2. If the API key group is subscription type, resolve active product subscriptions that bind that real group.
3. If the API key has `subscription_product_family`, only consider that family.
4. If the API key has no family and exactly one family matches, use that family.
5. If the API key has no family and multiple families match, reject the request as ambiguous.
6. Inside a family, select and later split-debit by:
   - `subscription_products.sort_order ASC`
   - `user_product_subscriptions.starts_at ASC`
   - `user_product_subscriptions.id ASC`
7. If no active product is available, reject. Do not check legacy subscription state.
8. If product quota is insufficient, attempt user-controlled balance fallback only when the shortage is a quota shortage and the user enabled fallback.

## Balance Fallback

Fallback eligibility:

1. The request started from a subscription group.
2. Product quota is exhausted or insufficient.
3. `users.subscription_balance_fallback_enabled = true`.
4. `users.subscription_balance_fallback_limit_usd > users.subscription_balance_fallback_used_usd`.
5. `users.subscription_balance_fallback_group_id` is set.
6. The selected fallback group is active, standard, and authorized for this user under the same rules as normal standard API key group visibility.
7. The user balance is not already negative before the request.

Fallback behavior:

1. The request runtime group changes to the user's fallback balance group.
2. The usage log records balance billing, not product-subscription billing.
3. Settlement deducts balance and increments `subscription_balance_fallback_used_usd`.
4. Settlement rejects if the fallback cumulative limit would be exceeded.
5. Settlement may make `users.balance` negative. A negative balance blocks future requests until recharged.

## Redeem Code Behavior

Admin product-subscription redeem code generation must require a product:

1. Frontend disables product-subscription code generation until a product is selected.
2. Backend rejects product-subscription code creation without `product_id`.
3. Redemption rejects product-subscription codes missing `product_id` even if legacy `group_id` is present.
4. No product is guessed from group bindings.

Balance and concurrency redeem codes are unaffected.

## Admin Operations

Product-subscription list supports:

1. Adjust:
   - update `expires_at`
   - update `status`
   - update `notes`
   - no per-user daily-limit override
2. Reset quota:
   - reset daily usage and Beijing daily window
   - reset weekly usage and rolling weekly window
   - reset monthly usage and rolling monthly window
   - clear daily carryover when daily reset is selected
3. Revoke:
   - set status to revoked
   - keep usage history
4. Manage fallback for a user:
   - enable/disable
   - set cumulative limit
   - set used amount or reset to zero
   - set fallback group

## Frontend

User subscription page:

1. Balance fallback card is default off.
2. When enabled, user must pick a standard authorized fallback group.
3. User must set a positive cumulative limit.
4. Show used and remaining fallback budget.
5. Show negative balance note when balance is below zero.
6. Product cards label daily windows as Beijing-time day windows and weekly/monthly windows as rolling windows.

API key create/edit:

1. When selecting a subscription group with one available product family, bind that family automatically.
2. When multiple families are available, require family selection.
3. Standard groups do not show family selection.

Admin subscription pages:

1. Product-subscription list action buttons call implemented backend routes.
2. Adjust dialog removes `daily_limit_usd`.
3. Admin user or subscription detail controls expose fallback management.

## Error Semantics

Use clear errors:

- Missing product ID on product redeem code: `PRODUCT_REDEEM_CODE_INVALID`
- No product subscription for subscription group: `SUBSCRIPTION_NOT_FOUND`
- Ambiguous product family: `PRODUCT_FAMILY_REQUIRED`
- Product quota insufficient and fallback disabled: `SUBSCRIPTION_BALANCE_FALLBACK_REQUIRED`
- Fallback group missing: `SUBSCRIPTION_BALANCE_FALLBACK_GROUP_REQUIRED`
- Fallback group unauthorized: `SUBSCRIPTION_BALANCE_FALLBACK_GROUP_FORBIDDEN`
- Fallback cumulative limit exceeded: existing fallback limit error
- Negative balance before request: existing insufficient balance error or a balance-specific error

## Tests

Backend tests must cover:

1. Product redeem code without `product_id` is rejected.
2. Product redeem code generation without `product_id` is rejected.
3. Product bindings reject standard groups.
4. Subscription group resolution does not fall back to legacy subscription.
5. Multiple families require explicit API key family.
6. API key family restricts product selection.
7. Same-family subscriptions split debit in order.
8. Product quota shortage with fallback disabled rejects.
9. Product quota shortage with fallback enabled switches to selected balance group.
10. Unauthorized fallback group rejects.
11. Balance fallback can make balance negative but blocks the next request.
12. Fallback cumulative limit is enforced.
13. Beijing daily window reset and carryover.
14. Rolling 30-day monthly window.
15. Admin adjust, reset, and revoke routes.

Frontend tests must cover:

1. Fallback requires enabled, group, and positive limit.
2. Fallback group options only include standard authorized groups.
3. Product redeem code generation requires product selection.
4. Admin product-subscription buttons call backend routes.
5. Adjust dialog does not expose per-user daily limit.
6. API key product-family selection appears only when needed.

## Migration Notes

Migration should be additive:

1. Add nullable `users.subscription_balance_fallback_group_id`.
2. Add nullable `api_keys.subscription_product_family`.
3. Do not delete `groups.balance_fallback_group_id`.
4. Do not auto-delete existing product bindings unless a later manual cleanup is approved.
5. Existing users keep fallback disabled unless they already enabled it; users with enabled fallback must select a group before fallback works.
