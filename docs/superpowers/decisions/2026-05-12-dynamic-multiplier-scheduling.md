# Dynamic Multiplier Scheduling Decision

- Date: 2026-05-12
- Task: T018
- Status: active
- Owner: 后端开发实现者
- Scope: 动态倍率调度规则；固化用户确认的动态分组、预算倍率、账号级倍率乘算、7 天均费兜底规则，防后续 agent 改错。
- Supersedes: N/A

## Background

用户确认动态倍率功能当前方向是好的，本任务只补决策文档，作为后续代码注释、隔离测试、CI grep 保护的单一事实源。本文对齐 T003 的倍率三态结论：固定计费倍率、动态预算/账号选择倍率、产品订阅扣减倍率不能混用。

输入只读：`docs/superpowers/audits/2026-05-12-semantic-dual-tracks.md`、`docs/superpowers/docs-index/CONVENTIONS.md`、`backend/internal/service/` 调度/计费代码、`backend/ent/schema/` 分组/账号/API Key schema。

本文不改任何 `.go` 代码，不改既有文档正文，不新增测试代码。S4 保护机制只写成后续实施要求。

## Conclusion

动态分组的用户侧结算倍率规则钉死为：动态分组基础倍率恒为 1，实际用户侧倍率来自账号在该分组绑定上的 `account_groups.billing_multiplier`，默认 1；预算倍率只用于动态账号选择，不是直接落账倍率。

调度顺序钉死为：先在 `账号倍率 <= 预算倍率` 的账号中选倍率最大的；这些账号不可用时继续按低倍率账号从高到低重试；若只能使用高于预算倍率的账号，则必须满足过去 7 天窗口平均花费低于预算倍率，否则返回 `No available count`。

## Red Lines

- 不得把动态分组的 `groups.rate_multiplier` 重新接入用户侧扣费；动态分组用户侧 base multiplier 必须保持 1。
- 不得把 `budget_multiplier` 当成实际账单倍率写入 `usage_logs.rate_multiplier`。
- 不得删除或绕过 `compareDynamicPricingAccountPreference`、`isAccountWithinDynamicBudget`、`withDynamicPricingBudgetState` 的动态预算调度语义。
- 不得把 `accounts.rate_multiplier` 混成用户/API Key 扣费倍率；它是账号成本统计口径。
- 不得把产品订阅 `subscription_product_groups.debit_multiplier` 混入本文的动态账号调度规则。

## 1. Why This Design Is Being Frozen

T003 指出倍率存在三组语义：固定/用户侧计费倍率、动态预算/账号选择倍率、产品订阅扣减倍率。动态倍率的风险不是功能缺失，而是字段名都叫 multiplier，后续 agent 容易把预算倍率、账号成本倍率、账号-分组绑定倍率、产品扣减倍率互相替换。

本文的目标是把用户确认的四规则和当前代码证据绑定起来：后续如果代码、测试、API 文案与本文冲突，先以本文为准，除非用户新批一份 decision。

## 2. 四规则原文

原文：
> 1. 允许开一个分组，类型设计为"动态"

原文：
> 2. 动态类型的分组里边是允许设计一个预算倍率的

原文：
> 3. 要能支持不同的账号在同一个分组里按不同的倍率乘算。也就是在结算的路径下多乘一个数，这个数默认是 1

原文：
> 4. 在动态分组里，分组倍率是 1，然后不同账号设计了不同的倍率乘算，用户是可以设计一个预算倍率的。然后调度策略服从以下的规律：
>    - "优先使用小于等于预算倍率的账号里面，倍率最大的那一个；
>    - 当上述条件不可用，那么继续在低倍率账号里从高到低的重试；
>    - 如果上述账号都不可用，可以使用高倍率的账号，但是必须满足'在过去的 7 天窗口内平均花费低于预算倍率'才可用，否则返回 'No available count'"

## 3. Terms Pinned To T003 And Code

| User/product term | Code term | Decision |
|---|---|---|
| `group.type = 'dynamic'` | `groups.pricing_mode = 'dynamic'`; service constants `GroupPricingModeDynamic = "dynamic"` | Do not add a new `group.type` field. The group type marker is current `pricing_mode`. |
| Dynamic group base multiplier | `groups.rate_multiplier` exists, but `resolveBillingMultiplierForUsage` sets dynamic `BaseRateMultiplier = 1` | In dynamic groups, user-side group/base multiplier is 1. |
| `group.budget_multiplier` / budget multiplier | Group default is `groups.default_budget_multiplier`; API key override is `api_keys.budget_multiplier`; fallback is `DefaultBudgetMultiplier = 8.0` | Budget multiplier is for scheduling admission/sort target, not a direct billing multiplier. |
| Account-level multiplier inside a group | `account_groups.billing_multiplier`; service domain `AccountGroup.BillingMultiplier`; accessor `Account.GroupBillingMultiplier(groupID)` | This is the “结算路径下多乘一个数，默认 1”. |
| `account_cost_multiplier` | `accounts.rate_multiplier`; service accessor `Account.BillingRateMultiplier()`; usage snapshot `usage_logs.account_rate_multiplier` | Account cost/statistics multiplier. It does not change user/API Key deduction. |
| User-side final billing multiplier | `billingMultiplierResolution.EffectiveBillingMultiplier`; usage snapshot `usage_logs.rate_multiplier` | Dynamic: `1 * account.GroupBillingMultiplier(groupID)`. Fixed: group/user group rate times account-group billing adjustment. |
| Product subscription debit multiplier | `subscription_product_groups.debit_multiplier`; usage snapshots `group_debit_multiplier`, `product_debit_cost` | Out of scope for this dynamic scheduling decision. |

Code-verified cost formulas:

```text
standard_cost = upstream/model cost before user-facing multiplier

Dynamic group user billing:
user_effective_multiplier = 1 * account_groups.billing_multiplier
user_actual_cost = standard_cost * user_effective_multiplier
usage_logs.rate_multiplier = user_effective_multiplier

Account cost/statistics:
account_rate_multiplier = accounts.rate_multiplier defaulting to 1
usage_logs.account_rate_multiplier = account_rate_multiplier
account_cost_stat = standard_cost * account_rate_multiplier
```

The user phrase `final_cost = upstream_cost x group.billing_multiplier x account.cost_multiplier x group_account.multiplier` must be read against current code as two separate ledgers: user billing uses dynamic base 1 plus `account_groups.billing_multiplier`; account-cost reporting uses `accounts.rate_multiplier` separately.

## 4. Current Code Evidence

| Evidence | File:line | What it proves |
|---|---|---|
| Group pricing mode schema | `backend/ent/schema/group.go:53` | `pricing_mode` exists and defaults to `fixed`. |
| Dynamic group default budget schema | `backend/ent/schema/group.go:57` | `default_budget_multiplier` exists for dynamic pricing groups. |
| API key budget schema | `backend/ent/schema/api_key.go:51` | `api_keys.budget_multiplier` is the per-key budget multiplier. |
| Account-group billing multiplier schema | `backend/ent/schema/account_group.go:35` | `account_groups.billing_multiplier` exists and defaults to `1.0`. |
| Account cost multiplier schema | `backend/ent/schema/account.go:107` | `accounts.rate_multiplier` is account-side billing/cost multiplier. |
| Service constants | `backend/internal/service/dynamic_pricing.go:13` | `fixed`, `dynamic`, default budget `8.0`, 7-day window, and `No available count` error are centralized. |
| Dynamic budget resolution | `backend/internal/service/dynamic_pricing.go:110` | Budget priority is API key budget, then group default budget, then `DefaultBudgetMultiplier`. |
| Dynamic user-side billing multiplier | `backend/internal/service/dynamic_pricing.go:123` | Dynamic groups set base to 1 and multiply `account.GroupBillingMultiplier(apiKey.GroupID)`. |
| 7-day budget state window | `backend/internal/service/dynamic_pricing.go:177` | Current code builds a budget state using `dynamicBudgetWindow = 7 * 24 * time.Hour`. |
| 7-day average multiplier calculation | `backend/internal/service/dynamic_pricing.go:217` | Current code uses `TotalActualCost / TotalCost` from usage stats when `TotalCost > 0`. |
| Dynamic account preference sort | `backend/internal/service/dynamic_pricing.go:247` | Budget-or-lower accounts sort by highest multiplier; over-budget accounts sort by lower multiplier. |
| Dynamic budget admission | `backend/internal/service/dynamic_pricing.go:288` | Over-budget accounts are allowed only when current average multiplier is below budget. |
| OpenAI scheduling entry | `backend/internal/service/openai_gateway_service.go:1607` | `selectAccountWithLoadAwareness` wraps scheduling context with dynamic budget state. |
| Candidate filtering for dynamic budget | `backend/internal/service/openai_gateway_service.go:1729` | Candidates failing dynamic budget are skipped and can trigger budget-exceeded error. |
| Dynamic budget exceeded return | `backend/internal/service/openai_gateway_service.go:1737`; `backend/internal/service/openai_gateway_service.go:1852`; `backend/internal/service/openai_gateway_service.go:1887` | If dynamic budget blocks all candidates, scheduler returns `ErrDynamicPricingBudgetExceeded`. |
| Account comparison hook | `backend/internal/service/openai_gateway_service.go:1531`; `backend/internal/service/openai_gateway_service.go:1538`; `backend/internal/service/openai_gateway_service.go:1572` | Dynamic multiplier preference is applied before priority/load/LRU comparisons. |
| Sticky account recheck | `backend/internal/service/openai_gateway_service.go:1411`; `backend/internal/service/openai_gateway_service.go:1682` | Sticky account is rejected if dynamic budget says it is not schedulable. |
| Generic gateway billing path | `backend/internal/service/gateway_service.go:8423` | Generic `GatewayService.RecordUsage` resolves dynamic final multiplier before calculating cost. |
| OpenAI billing path | `backend/internal/service/openai_gateway_service.go:5292` | `OpenAIGatewayService.RecordUsage` uses the same dynamic multiplier resolver. |
| Usage log multiplier snapshot | `backend/internal/service/gateway_service.go:8674`; `backend/internal/service/openai_gateway_service.go:5369` | Final user-side multiplier is written to `usage_logs.rate_multiplier`. |
| Account multiplier snapshot | `backend/internal/service/gateway_service.go:8457`; `backend/internal/service/openai_gateway_service.go:5331` | Account-side cost multiplier is separately snapped from `Account.BillingRateMultiplier()`. |
| Actual cost multiplication | `backend/internal/service/billing_service.go:570` | Cost breakdown actual cost is standard cost times `RateMultiplier`. |
| Dynamic scheduling regression tests | `backend/internal/service/openai_gateway_service_test.go:769`; `backend/internal/service/openai_gateway_service_test.go:816`; `backend/internal/service/openai_gateway_service_test.go:865`; `backend/internal/service/openai_gateway_service_test.go:914` | Existing tests cover budget-or-lower preference and high-multiplier 7-day average gate. |
| Dynamic billing regression test | `backend/internal/service/openai_gateway_record_usage_test.go:283`; `backend/internal/service/gateway_record_usage_test.go:195` | Existing tests prove dynamic billing uses account-group billing multiplier as user-side multiplier. |
| Available channel display fields | `backend/internal/service/channel_available.go:125` | Display fields summarize dynamic min/max/budget/matched multiplier; they are not billing truth. |

## 5. Scheduling Algorithm Pseudocode And Boundary Cases

### Pseudocode

```text
input:
  group, api_key, candidate_accounts, requested_model, session/exclusion/load state

if group.pricing_mode != "dynamic":
  use normal fixed scheduling

budget = api_key.budget_multiplier
if budget is null:
  budget = group.default_budget_multiplier
if budget is null:
  budget = DefaultBudgetMultiplier

for each account:
  account_multiplier = account.GroupBillingMultiplier(group_id)
  if not configured or invalid:
    account_multiplier = 1

eligible_low_or_equal = accounts where account_multiplier <= budget
sort eligible_low_or_equal by account_multiplier desc

for account in eligible_low_or_equal:
  if account is active, schedulable, supports requested model, not excluded,
     not channel-restricted, not over concurrency, and DB runtime recheck still passes:
    return account

remaining_low_or_equal = lower multiplier accounts already ordered high to low
for account in remaining_low_or_equal:
  retry when earlier higher accounts are unavailable during runtime recheck/acquire
  if runtime state becomes available:
    return account

high_multiplier_accounts = accounts where account_multiplier > budget
sort high_multiplier_accounts by account_multiplier asc

for account in high_multiplier_accounts:
  if seven_day_average_multiplier < budget and account is otherwise usable:
    return account

return ErrDynamicPricingBudgetExceeded with message "No available count"
```

The pseudocode intentionally mirrors the four user rules. It does not introduce optimization, probabilistic weighting, or alternative fallback ordering.

### Boundary Cases

| Case | Required behavior |
|---|---|
| No account is active/schedulable/model-compatible | Return no available account. In dynamic budget-blocked cases, return `ErrDynamicPricingBudgetExceeded` / `No available count`; otherwise normal no-account error. |
| All account multipliers are `> budget`, and 7-day average is `>= budget` | Return `No available count`. |
| All account multipliers are `> budget`, and 7-day average is `< budget` | A high-multiplier account may be selected; current comparator orders over-budget accounts by lower multiplier first. |
| Account multiplier exactly equals budget | Treat as within budget and include in the first bucket; equality is allowed with epsilon. |
| 7-day window has no usage or `TotalCost == 0` | Current code does not calculate an average and blocks over-budget accounts because `state.currentStandardCost <= 0` returns false for high multiplier admission. This is code evidence, not a new product answer. |
| Multiple budget-or-lower accounts are usable | Choose the highest multiplier among `<= budget`; ties fall through to priority/load/LRU depending on path. |
| Multiple high-multiplier accounts satisfy the 7-day average gate | Current comparator orders over-budget accounts by lower multiplier first, then priority/load/LRU. |
| Budget multiplier not set or null | Resolve to API key budget if present, else group default, else `DefaultBudgetMultiplier = 8.0`; it does not degrade to arbitrary max multiplier selection. |
| Account disabled / maintenance / temporarily unschedulable during scheduling | Runtime recheck skips it and continues through the ordered candidates; if all viable candidates fail, return the appropriate no-available error. |
| Sticky session points to a high-multiplier account | Sticky hit must still pass `isAccountSchedulableForDynamicPricing`; otherwise sticky is cleared/skipped. |
| Account-group binding missing or invalid multiplier | `Account.GroupBillingMultiplier` / `AccountGroup.EffectiveBillingMultiplier` defaults to 1. |

## 6. S4 Protection Mechanism

### Code Comment Banner

When S4 touches dynamic scheduling, add this banner above these entry points: `backend/internal/service/dynamic_pricing.go:247`, `backend/internal/service/dynamic_pricing.go:288`, `backend/internal/service/openai_gateway_service.go:1607`.

```go
// DO NOT MODIFY dynamic multiplier scheduling without user approval.
// Source of truth: docs/superpowers/decisions/2026-05-12-dynamic-multiplier-scheduling.md
// Rules: dynamic group base multiplier is 1; account_groups.billing_multiplier is the per-account user multiplier;
// budget_multiplier is scheduling-only; high-multiplier fallback requires 7-day average below budget.
```

### Isolation Test Checklist For T019

T019 should write test cases, not this document. Minimum required assertions:

| Test class | Required assertion |
|---|---|
| Rule 1 | A group with `PricingMode = GroupPricingModeDynamic` is treated as dynamic; fixed groups do not use dynamic budget sorting. |
| Rule 2 | API key budget overrides group default; group default overrides `DefaultBudgetMultiplier`. |
| Rule 3 | Dynamic billing writes `usage_logs.rate_multiplier = account_groups.billing_multiplier`; missing binding defaults to 1. |
| Rule 4 first bucket | With multipliers `[3, 6, 10]` and budget `8`, scheduler chooses multiplier `6`. |
| Rule 4 low retry | If multiplier `6` account is unavailable, scheduler retries lower budget-or-equal accounts from high to low. |
| Rule 4 high fallback allowed | If only `> budget` accounts are available and 7-day average is below budget, scheduler may choose high multiplier. |
| Rule 4 high fallback blocked | If only `> budget` accounts are available and 7-day average is at/above budget, scheduler returns `No available count`. |
| Boundary equality | Multiplier exactly equal to budget is in the first bucket. |
| Boundary no history | No 7-day usage / `TotalCost == 0` behavior is locked to current code unless user changes decision. |
| Boundary order | Multiple high-multiplier accounts satisfying average gate choose the lowest over-budget multiplier first. |
| Runtime recheck | Disabled/maintenance/rate-limited account is skipped and next ordered candidate is tried. |
| Sticky session | Sticky account must still pass dynamic budget admission. |

Existing tests to preserve and extend: `backend/internal/service/openai_gateway_service_test.go:769`, `backend/internal/service/openai_gateway_service_test.go:816`, `backend/internal/service/openai_gateway_service_test.go:865`, `backend/internal/service/openai_gateway_service_test.go:914`, `backend/internal/service/openai_gateway_record_usage_test.go:283`, `backend/internal/service/gateway_record_usage_test.go:195`.

### CI Grep Protection

CI should fail or require explicit approval if these identifiers are removed, renamed, or bypassed without updating this decision and tests:

```sh
rg -n 'GroupPricingModeDynamic|DefaultBudgetMultiplier|dynamicBudgetWindow|ErrDynamicPricingBudgetExceeded|resolveDynamicBudgetMultiplier|resolveBillingMultiplierForUsage|compareDynamicPricingAccountPreference|isAccountWithinDynamicBudget|withDynamicPricingBudgetState|GroupBillingMultiplier|EffectiveBillingMultiplier|BillingRateMultiplier' backend/internal/service backend/ent/schema
```

Mandatory protected files:

| File | Protected facts |
|---|---|
| `backend/internal/service/dynamic_pricing.go` | Budget resolution, 7-day window, account preference comparator, dynamic admission gate, `No available count`. |
| `backend/internal/service/openai_gateway_service.go` | Dynamic budget state injection and account selection ordering. |
| `backend/internal/service/gateway_service.go` | Generic record usage multiplier resolution and usage log snapshots. |
| `backend/internal/service/openai_gateway_service.go` | OpenAI record usage multiplier resolution and usage log snapshots. |
| `backend/internal/service/account.go` | `GroupBillingMultiplier` and `BillingRateMultiplier` separation. |
| `backend/internal/service/account_group.go` | Account-group billing multiplier defaulting to 1. |
| `backend/ent/schema/group.go` | `pricing_mode` and `default_budget_multiplier`. |
| `backend/ent/schema/account_group.go` | `billing_multiplier`. |
| `backend/ent/schema/api_key.go` | `budget_multiplier`. |
| `backend/ent/schema/account.go` | `rate_multiplier`. |

### Documentation Linkage

This decision is the single source of truth for dynamic multiplier scheduling. If future code or tests conflict with it, fix the code/tests or create a new user-approved decision. Do not silently change behavior because a field name seems more intuitive.

## 7. Non-Goals

- 不讨论 fixed 分组倍率调度；固定分组仍归固定倍率逻辑。
- 不讨论产品订阅倍率 `product_debit_multiplier` / `subscription_product_groups.debit_multiplier`；见 T003 产品订阅扣减倍率语义。
- 不讨论分组切换策略、fallback group 选择、渠道限制策略；这些归产品/调度其他决策。
- 不讨论 UI 重命名或 API 契约改动；本文只固化后端动态倍率调度语义。
- 不把当前 `pricing_mode` 改名成 `group.type`，不新增 `group_account_multiplier` 表。

## 8. Open Questions For User

- “7 天窗口”是自然日窗口还是 rolling 144 小时？当前代码是 `7 * 24 * time.Hour` rolling window，且窗口起点会被 API key 创建时间截断。
- “7 天均费”的分母应是调用次数、标准成本 `TotalCost`、时间跨度内小时数，还是账号维度成本？当前代码是 `TotalActualCost / TotalCost`，并且按 API key usage stats 取数，不是逐账号平均。
- 高倍率兜底里的“过去 7 天窗口内平均花费低于预算倍率”是否必须改成“被选账号自身的 7 天平均”，还是沿用当前 API key 维度平均？本文不擅自改行为。
- `No available count` 对外 HTTP status 是否统一为 `429` + `{code: "no_available_account", ...}`？当前内部错误是 `infraerrors.TooManyRequests("DYNAMIC_PRICING_BUDGET_EXCEEDED", "No available count")`，具体 HTTP envelope 留给契约任务确认。
- 7 天窗口内没有历史用量时，高倍率账号是否应允许首单通过？当前代码阻止高倍率兜底，是否保留需用户拍板。
