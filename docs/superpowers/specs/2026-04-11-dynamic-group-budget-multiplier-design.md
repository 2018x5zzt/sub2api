# Dynamic Group Budget Multiplier Design

## Summary

This design introduces a new pricing mode for groups: dynamic pricing.

When a user creates an API key and binds it to a dynamic pricing group, the key gets its own `budget_multiplier`. This value is a routing-budget control for that key inside that specific dynamic group. It is used to decide whether the request may continue and whether a candidate account may be selected. It is not the direct multiplier used for final billing.

The system first calculates the API key's recent 7-day weighted average multiplier:

```text
current_average_multiplier = recent_7d_actual_cost / recent_7d_standard_cost
```

If the current average is already above the configured budget, the request is rejected immediately. Otherwise, the router evaluates candidate accounts and only allows candidates whose predicted post-request average would remain within budget.

Final billing keeps the existing settlement model:

```text
ActualCost = TotalCost * effective_multiplier
effective_multiplier = resolved_group_multiplier * account_group.billing_multiplier
```

Where `resolved_group_multiplier` is the user-specific group multiplier when present, otherwise the group's default multiplier.

Claude's existing dynamic multiplier group is the first migration target. The same design must later support OpenAI and any other platform that adopts dynamic pricing groups.

## Problem

The current system already uses groups as the user-facing package layer. A group decides account pool access, billing behavior, and visibility. However, the current Claude dynamic multiplier setup lacks a user-facing budget control. Users cannot express what average multiplier they are willing to accept while still allowing the system to dynamically route within a pool.

Without this budget control:

- users cannot trade off price against availability in an explicit way
- dynamic routing feels opaque
- high-multiplier fallback decisions are hard to justify
- future rollout of dynamic pricing groups across platforms will not have a reusable product model

## Goals

- Add a reusable dynamic pricing group model that is not Claude-specific.
- Let each dynamic pricing group define its own default budget multiplier.
- Let each API key bound to a dynamic pricing group save its own `budget_multiplier`.
- Show `budget_multiplier` only when the selected group is a dynamic pricing group.
- Use a rolling 7-day, standard-cost-weighted average multiplier as the budget control rule.
- Keep the existing billing settlement path and avoid introducing a second billing formula for dynamic groups.
- Preserve budget guarantees by failing requests when the current rolling average is already over budget or when no budget-compatible candidate remains.
- Reuse the existing custom error response capability to explain budget-based failures.
- Migrate the existing Claude dynamic multiplier group with a default budget multiplier of `8.0`.

## Non-Goals

- Do not support one API key bound to multiple groups at the same time.
- Do not allow changing an existing API key's bound group after creation.
- Do not share budget state across dynamic groups.
- Do not silently exceed the key's configured budget multiplier for availability reasons.
- Do not build a general pricing strategy engine in this iteration.

## Domain Model

### Group

Groups gain a pricing mode:

- `fixed`
- `dynamic`

When a group uses `dynamic`, it also stores:

- `default_budget_multiplier`

This value is a template default for future API keys created under that group. It does not retroactively rewrite existing keys.

### API Key

API keys remain single-group credentials. An API key stores:

- `group_id`
- `budget_multiplier` when `group_id` points to a dynamic pricing group

`budget_multiplier` belongs to the API key inside its currently bound dynamic group. It is not a global user setting and it is not shared across different dynamic groups.

### Group Billing Multiplier Resolution

Dynamic pricing groups continue to use the existing group multiplier stack:

- group default rate multiplier
- optional user-specific group multiplier override
- account-group binding multiplier: `account_group.billing_multiplier`

The resolved billing multiplier for a selected account is:

```text
effective_multiplier = resolved_group_multiplier * account_group.billing_multiplier
```

`budget_multiplier` does not replace this stack. It only constrains routing against rolling budget policy.

### Scope Rules

- A budget multiplier only matters for API keys bound to dynamic pricing groups.
- Different dynamic groups have independent budget settings and independent rolling budget behavior.
- Deleting an API key deletes its budget configuration and makes its budget history irrelevant for future routing.

## User Experience

### Admin Group Creation and Editing

Group management adds a pricing mode selector:

- fixed pricing
- dynamic pricing

When the admin selects fixed pricing:

- keep the existing fixed multiplier behavior
- do not show dynamic budget configuration

When the admin selects dynamic pricing:

- show `default_budget_multiplier`
- allow decimal input
- recommended range: `3.0` to `50.0`
- recommended step: `0.1`

The existing Claude dynamic multiplier group is migrated to:

- pricing mode: `dynamic`
- default budget multiplier: `8.0`

### User API Key Creation

When the user creates an API key:

1. The user selects a group.
2. If the selected group is a dynamic pricing group, show the `预算倍率` field.
3. Pre-fill that field with the group's `default_budget_multiplier`.
4. Let the user keep the default or edit it before saving.
5. If the selected group is not dynamic, do not show this field.

Suggested label:

- `预算倍率`

Suggested helper text:

- `系统会按最近7天标准成本加权平均倍率控制选号范围。预算倍率用于预算约束，不直接等于最终扣费倍率。`

### User API Key Editing

- API keys may edit `budget_multiplier` after creation.
- API keys may not change `group_id` after creation.
- If the API key is bound to a dynamic pricing group, show `预算倍率`.
- If the API key is bound to a fixed pricing group, hide this field.

## Routing and Budget Logic

### Budget Multiplier Resolution

The effective routing budget is resolved in this order:

```text
api_key.budget_multiplier > group.default_budget_multiplier > 8.0
```

The hard-coded `8.0` fallback is only a defensive read fallback for incomplete data during rollout.

### Budget Window and Current Average

Budget control uses a rolling 7-day weighted average. The budget window is:

```text
window_start = max(now - 7d, api_key.created_at)
window_end = now
```

The system loads usage stats for that API key within this window and computes:

```text
current_average_multiplier = recent_7d_actual_cost / recent_7d_standard_cost
```

Where:

- `actual_cost` is the already billed cost recorded in usage logs
- `standard_cost` is the baseline cost before dynamic multiplier uplift
- both values come from the API key's own usage history, not shared group history

This is a standard-cost-weighted multiplier average, not a request-count average and not a natural-week average.

If `current_average_multiplier > budget_multiplier`, the request is rejected before account selection.

### Effective Billing Multiplier

Dynamic pricing does not introduce a separate billing formula. For the selected account:

```text
resolved_group_multiplier =
  user_specific_group_multiplier or group.default_rate_multiplier

effective_multiplier =
  resolved_group_multiplier * account_group.billing_multiplier

ActualCost = TotalCost * effective_multiplier
```

This applies to token billing, per-request billing, and image billing. The account's own quota accounting may still use its separate account-level billing multiplier, but that value does not participate in dynamic budget evaluation.

### Candidate Budget Check

Each candidate account is evaluated by its `effective_multiplier`.

If there is no usable 7-day history for the API key yet, or `recent_7d_standard_cost = 0`, the rule is:

```text
allow_candidate = effective_multiplier <= budget_multiplier
```

If there is existing history, the system predicts the next request using the recent average standard cost per request:

```text
estimated_next_standard_cost =
  recent_7d_standard_cost / recent_7d_request_count

predicted_average_multiplier =
  (recent_7d_actual_cost + estimated_next_standard_cost * effective_multiplier) /
  (recent_7d_standard_cost + estimated_next_standard_cost)

allow_candidate = predicted_average_multiplier <= budget_multiplier
```

This is intentionally not a token-level estimate of the current request. It is a rolling budget policy based on recent average standard cost per request.

### Selection Strategy

Routing for dynamic pricing groups follows these rules:

1. Budget-compatible candidates rank ahead of budget-incompatible candidates.
2. If two candidates are both budget-compatible, prefer the higher `effective_multiplier`.
3. If two candidates are both budget-incompatible, prefer the lower `effective_multiplier`.
4. Only candidates that pass the budget check may actually be selected for the request.
5. If no budget-compatible candidates remain, fail the request.

This is not a low-multiplier-first policy. The actual strategy is:

- stay within the configured rolling budget
- use the highest multiplier still allowed by that budget
- fail once only over-budget candidates remain

### Failure Condition

A request fails when:

- the API key is bound to a dynamic pricing group
- the current rolling average multiplier is already above `budget_multiplier`
- or budget-compatible routing candidates are exhausted and remaining available candidates would break budget policy

The failure does not mean the whole upstream is unavailable. It means the currently available routes are outside the key's accepted budget range.

## Error Handling

Budget-based failures should return the existing dedicated custom error instead of a generic upstream error.

Error code:

- `DYNAMIC_PRICING_BUDGET_EXCEEDED`

Current default message:

- `当前预算倍率下没有可用账号，可调高预算倍率后重试`

This error is used for both cases:

- the current rolling average is already over budget
- the current request cannot find any budget-compatible candidate

The budget failure integrates with the existing custom error response capability rather than introducing a separate response framework. Structured details such as current average or configured budget can be added later if the error-response layer is expanded, but they are not required for the initial design.

## Data Migration

### Existing Claude Dynamic Group

Migrate the current Claude dynamic multiplier group as follows:

- set pricing mode to `dynamic`
- set `default_budget_multiplier = 8.0`

### Existing API Keys Bound to That Group

For API keys currently bound to that Claude dynamic pricing group:

- set `budget_multiplier = 8.0`

Only migrate API keys currently bound to that group. Do not backfill API keys that were historically bound to it but no longer are.

### Safe Read Fallback

During rollout, backend reads for dynamic pricing group keys may use a defensive fallback:

- if a dynamic-group API key is missing `budget_multiplier`, read `group.default_budget_multiplier`
- if the group default is also missing, fall back to hard-coded `8.0`
- still plan to repair the data rather than relying on fallback long term

## Update and Lifecycle Rules

- API keys may change `budget_multiplier`.
- API keys may not change `group_id` after creation.
- To move to another group, the user must create a new key and delete the old one.
- When creating an API key for a dynamic pricing group, omit `budget_multiplier` only to accept the group's current default.
- Changing a group's `default_budget_multiplier` only affects future API keys.
- Existing API keys keep their own saved `budget_multiplier`.

This avoids cross-group budget leakage and avoids carrying rolling budget semantics from one dynamic pricing group to another.

## Testing

### Backend

- group schema and validation for `fixed` vs `dynamic`
- validation of `default_budget_multiplier` range and decimal handling
- API key create path stores `budget_multiplier` only for dynamic groups
- API key update path allows `budget_multiplier` edits but rejects `group_id` changes
- migration/backfill for the existing Claude dynamic group and bound API keys
- routing tests for:
  - request rejection when the current rolling average already exceeds budget
  - budget compatibility check with no history
  - predicted-average compatibility check with existing history
  - budget-compatible higher-multiplier preference
  - lower-multiplier preference only when all remaining candidates are already over budget
  - request rejection when only over-budget candidates remain
- billing tests for `ActualCost = TotalCost * effective_multiplier`
- billing tests that `budget_multiplier` affects routing only and does not directly replace the billing multiplier
- budget failure response uses the expected custom error code and message

### Frontend

- group create/edit form shows dynamic pricing fields only when appropriate
- API key create form shows `预算倍率` only for dynamic groups
- API key create form pre-fills from group default
- API key edit form allows budget edits and prevents group change
- normal groups do not show budget multiplier UI

### Migration Verification

- migrated Claude dynamic group has pricing mode `dynamic`
- migrated Claude dynamic group has default budget multiplier `8.0`
- currently bound keys under that group have `budget_multiplier = 8.0`
- non-dynamic-group keys remain unchanged

## Risks

### Product Risk

Because routing prefers the highest multiplier still allowed by budget, a key may consume budget headroom faster than users expect if they assume the system always prefers the cheapest account. This is acceptable, but it must be clearly documented.

### UX Risk

Users may confuse:

- fixed group multiplier
- dynamic group budget multiplier
- final billed multiplier on the selected account

The UI must keep these concepts visibly separate. Fixed pricing groups should continue to talk about fixed billing multiplier. Dynamic pricing groups should talk about `预算倍率` as a routing budget, not as the final billing multiplier.

### Expansion Risk

Claude is only the first migration target. The implementation must avoid platform-specific branches so OpenAI and later dynamic pricing groups can reuse the same domain model.

## Recommendation

Implement dynamic pricing as a group capability, not a Claude-specific feature and not a global API key budget system.

The first rollout should:

- migrate the existing Claude dynamic group
- add dynamic pricing mode and default budget multiplier to groups
- add per-key `budget_multiplier`
- prevent API key group switching
- expose `预算倍率` only for dynamic pricing groups
- reject requests when the current rolling average is already over budget
- otherwise select the highest-multiplier candidate that still keeps the rolling average within budget
- fail requests when no budget-compatible candidate remains

This keeps the product model aligned with the existing billing pipeline, makes the routing rule explicit, preserves budget guarantees, and leaves a clean path to future OpenAI dynamic pricing groups.
