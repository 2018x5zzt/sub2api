# Dynamic Group Budget Multiplier Design

## Summary

This design introduces a new pricing mode for groups: dynamic pricing.

When a user creates an API key and binds it to a dynamic pricing group, the key gets its own `budget_multiplier`. This value controls the key's acceptable rolling pricing budget inside that specific dynamic group. The system routes requests dynamically across different multiplier pools while trying to keep the key's recent 7-day cost-weighted average multiplier within the configured budget.

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
- Preserve price guarantees by failing requests when only over-budget routing remains available.
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

- `系统会按最近7天标准成本加权平均倍率，尽量控制在该预算倍率以内。低倍率号池不足时，请求可能失败。`

### User API Key Editing

- API keys may edit `budget_multiplier` after creation.
- API keys may not change `group_id` after creation.
- If the API key is bound to a dynamic pricing group, show `预算倍率`.
- If the API key is bound to a fixed pricing group, hide this field.

## Routing and Budget Logic

### Budget Definition

Budget control uses a rolling 7-day weighted average:

```text
recent_7d_average_multiplier = recent_7d_actual_billed_cost / recent_7d_standard_cost
```

Where:

- `actual_billed_cost` includes dynamic multiplier effects
- `standard_cost` is the baseline cost before dynamic multiplier uplift

This is a standard-cost-weighted multiplier average, not a request-count average and not a natural-week average.

### Selection Strategy

Routing for dynamic pricing groups follows these rules:

1. Prefer lower-multiplier candidates first.
2. If a higher-multiplier candidate is considered, evaluate whether selecting it remains compatible with the key's budget policy.
3. If the candidate is budget-compatible, allow it.
4. If it is not budget-compatible, keep searching lower-cost candidates.
5. If no budget-compatible candidates remain, fail the request.

This preserves the product promise:

- protect price
- do not protect success rate at all costs

### Failure Condition

A request fails when:

- the API key is bound to a dynamic pricing group
- budget-compatible routing candidates are exhausted
- remaining available candidates would break budget policy

The failure does not mean the whole upstream is unavailable. It means the currently available routes are outside the key's accepted budget range.

## Error Handling

Budget-based failures should return a dedicated custom error instead of a generic upstream error.

Suggested error type:

- `dynamic_budget_blocked`

Suggested response content should include:

- configured `budget_multiplier`
- current recent 7-day actual average multiplier
- explanation that budget-range candidates are unavailable
- action hint to raise the budget multiplier or retry later

Suggested message template:

```text
当前预算倍率为 8.0x，最近7天实际平均倍率为 7.9x。
当前预算范围内的可用号池不足，请求无法继续调度。
建议提高预算倍率后重试，或稍后再试。
```

This design assumes the existing custom error response capability remains the delivery mechanism. The budget failure should integrate with that system rather than introducing a separate response framework.

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

- if a dynamic-group API key is missing `budget_multiplier`
- temporarily read the group's `default_budget_multiplier`
- still plan to repair the data rather than relying on fallback long term

## Update and Lifecycle Rules

- API keys may change `budget_multiplier`.
- API keys may not change `group_id` after creation.
- To move to another group, the user must create a new key and delete the old one.
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
  - lower-multiplier preference
  - budget-compatible higher-multiplier selection
  - request rejection when only over-budget candidates remain
- budget failure response uses the expected custom error type and message fields

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

If low-multiplier pools are frequently exhausted, users will see more budget-blocked failures. This is expected under a "protect price, not success rate" promise, but it requires clear messaging and operator visibility.

### UX Risk

Users may confuse:

- fixed group multiplier
- dynamic group budget multiplier

The UI must keep these concepts visibly separate. Fixed pricing groups should continue to talk about fixed billing multiplier. Dynamic pricing groups should talk about `预算倍率`.

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
- fail requests when no budget-compatible candidate remains

This keeps the product model understandable, preserves budget guarantees, and leaves a clean path to future OpenAI dynamic pricing groups.
