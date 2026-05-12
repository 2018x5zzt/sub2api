# Dynamic Multiplier Isolation Test Plan

- Date: 2026-05-12
- Task: T019
- Status: draft
- Owner: 测试者
- Scope: Turn T018 dynamic multiplier scheduling rules into an executable isolated backend test checklist.
- Supersedes: N/A

## Background

T018 freezes the dynamic multiplier behavior because the feature is currently correct but easy to regress: several fields are named `multiplier`, while their semantics differ. T018 §6 already lists 12 required assertions and protected identifiers; this plan turns those assertions into isolated test cases that can be implemented later by backend owners.

This document is different from `docs/superpowers/contracts/CONTRACT-TEST-PLAN.md`: that plan protects HTTP/API contracts with golden tests, while this plan protects backend scheduling and billing logic inside `backend/internal/service`. These tests should fail when a later agent accidentally mixes dynamic budget, account-group billing multiplier, group fixed multiplier, account cost multiplier, or product debit multiplier.

Inputs read for this plan:

- `docs/superpowers/decisions/2026-05-12-dynamic-multiplier-scheduling.md`
- `docs/superpowers/audits/2026-05-12-semantic-dual-tracks.md`
- `docs/superpowers/contracts/CONTRACT-TEST-PLAN.md`
- `docs/superpowers/docs-index/CONVENTIONS.md`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/gateway_record_usage_test.go`
- `backend/internal/service/dynamic_pricing.go`

## Conclusion

Add 12 isolated backend tests, exactly matching T018 §6. The current suite already covers several positive paths, but the protections are not complete enough to block the common future mistakes: treating fixed groups as dynamic, ignoring API key budget override, allowing high-multiplier fallback on no history, confusing equality with over-budget, or bypassing runtime/sticky rechecks.

The P0 core is Rule 1 through Rule 4: dynamic group detection, budget resolution, billing multiplier write path, and scheduling order/admission. Boundary and sticky/runtime cases should follow as the second batch.

## Red Lines

- This document does not write actual Go tests.
- Do not modify existing `.go`, `.ts`, or existing `.md` files as part of T019.
- §2 contains exactly 12 test cases, one for each T018 §6 assertion, no more and no fewer.
- S4 implementation must treat T018 as the source of truth unless the user approves a new decision.

## 1. Why T018 §6 Needs Independent Tests

T018 §6 is currently a decision-level checklist. It prevents misunderstanding, but it does not stop a future merge by itself. The dynamic multiplier rules need executable isolation tests because the failure modes are subtle:

- A change can still compile while silently routing dynamic groups through fixed sorting.
- A refactor can pass HTTP contract tests while changing internal account selection order.
- A billing path can write a plausible `rate_multiplier` while using `budget_multiplier` instead of `account_groups.billing_multiplier`.
- A sticky session or runtime recheck can bypass the dynamic budget gate even if normal selection tests pass.

The existing tests are useful anchors:

| Existing test | Current coverage | Gap |
|---|---|---|
| `backend/internal/service/openai_gateway_service_test.go:769` | Budget-or-lower account preference chooses multiplier 6 when budget is 8 | Does not prove fixed groups skip dynamic sorting; does not test equality or low retry |
| `backend/internal/service/openai_gateway_service_test.go:816` | High account may be used when budget average allows | Name overlaps with later case; fixture uses low account full and high account available |
| `backend/internal/service/openai_gateway_service_test.go:865` | High account allowed when 7-day average is below budget | Good base for fallback allowed |
| `backend/internal/service/openai_gateway_service_test.go:914` | High account blocked when average equals budget | Good base for fallback blocked / boundary equality of average |
| `backend/internal/service/openai_gateway_record_usage_test.go:283` | OpenAI record usage writes account-group billing multiplier | Needs missing-binding default variant |
| `backend/internal/service/gateway_record_usage_test.go:195` | Generic gateway record usage writes account-group billing multiplier | Good cross-path guard |

The T007 contract plan remains relevant as style guidance: use stable names, explicit fixtures, and deterministic expectations. But these tests are not HTTP golden tests; they should be backend table-driven/unit tests under `backend/internal/service`.

## 2. Test Case Checklist

### 2.1 Rule 1 — Dynamic Group Uses Dynamic Sorting

- Test name: `TestDynamicScheduling_Rule1_DynamicGroupUsesDynamicSorting`
- Assertion intent: If this fails, `GroupPricingModeDynamic` is no longer the switch that activates dynamic budget-aware sorting.
- Data fixture design: Create group `10` with `PricingMode: GroupPricingModeDynamic`, `DefaultBudgetMultiplier: 8`; create API key `100` bound to group `10` with no per-key override; create accounts `A1=3x`, `A2=6x`, `A3=10x` via `dynamicOpenAIAccount`; set all loads to `0` and all accounts active/schedulable.
- Expected result: `SelectAccountWithLoadAwareness` returns `account_id == A2` because `6 <= 8` and is the highest multiplier in the first bucket.
- Reuse existing test: Extend existing `backend/internal/service/openai_gateway_service_test.go:769` by renaming or wrapping it with the protected identifier comment. Current test already covers the positive dynamic case.
- Priority: P0.

Pseudo:

```go
// pseudo only; protects GroupPricingModeDynamic and compareDynamicPricingAccountPreference
ctx := dynamicContext(groupID=10, budget=8)
repo := accounts(3, 6, 10)
selection, err := svc.SelectAccountWithLoadAwareness(ctx, &groupID, "", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(2), selection.Account.ID)
```

### 2.2 Rule 2 — Budget Resolution Priority

- Test name: `TestDynamicScheduling_Rule2_APIKeyBudgetOverridesGroupDefaultAndSystemDefault`
- Assertion intent: If this fails, `budget_multiplier` resolution no longer follows API key override -> group default -> `DefaultBudgetMultiplier`.
- Data fixture design: Use three subcases with group `10` dynamic and accounts `A1=6x`, `A2=9x`, `A3=12x`. Case A: API key budget `10`, group default `8`, expect `A2`. Case B: API key budget nil, group default `8`, expect `A1`. Case C: both nil, system default `8`, expect `A1`. Usage stats can be empty because only first-bucket accounts are expected.
- Expected result: selected account matches budget source: API key override picks highest `<=10`; group default/system default picks highest `<=8`.
- Reuse existing test: New table-driven test. Existing `openai_gateway_service_test.go:769` only proves explicit API key budget.
- Priority: P0.

Pseudo:

```go
// pseudo only; protects resolveDynamicBudgetMultiplier and DefaultBudgetMultiplier
for _, tc := range cases {
  ctx := dynamicContextWithBudgets(tc.apiKeyBudget, tc.groupDefaultBudget)
  selection, err := svc.SelectAccountWithLoadAwareness(ctx, &groupID, "", "gpt-4", nil)
  require.NoError(t, err)
  require.Equal(t, tc.expectedAccountID, selection.Account.ID)
}
```

### 2.3 Rule 3 — Dynamic Billing Uses Account-Group Multiplier

- Test name: `TestDynamicBilling_Rule3_UsageLogRateMultiplierComesFromAccountGroupBinding`
- Assertion intent: If this fails, dynamic billing is writing the wrong multiplier, usually `groups.rate_multiplier`, `api_keys.budget_multiplier`, or `accounts.rate_multiplier` instead of `account_groups.billing_multiplier`.
- Data fixture design: Dynamic group `21` with `RateMultiplier: 99` to catch accidental fixed multiplier use; API key budget `8`; account `3003` has `AccountGroups[{GroupID:21, BillingMultiplier:3}]` and `RateMultiplier: ptr(7)` to prove account cost multiplier is separate; record usage through both OpenAI and generic gateway paths.
- Expected result: `usage_logs.rate_multiplier == 3`; actual user cost equals standard cost times `3`; account cost multiplier is not used for user deduction. Missing binding variant should write `1`.
- Reuse existing test: Extend existing `backend/internal/service/openai_gateway_record_usage_test.go:283` and `backend/internal/service/gateway_record_usage_test.go:195`; add missing-binding default subcase.
- Priority: P0.

Pseudo:

```go
// pseudo only; protects resolveBillingMultiplierForUsage and GroupBillingMultiplier
err := svc.RecordUsage(ctx, inputWithDynamicGroupAndAccountGroupMultiplier(3))
require.NoError(t, err)
require.Equal(t, 3.0, usageRepo.lastLog.RateMultiplier)
require.InDelta(t, standardCost*3.0, usageRepo.lastLog.ActualCost, epsilon)
```

### 2.4 Rule 4 First Bucket — Highest Account At Or Below Budget

- Test name: `TestDynamicScheduling_Rule4_FirstBucketChoosesHighestAtOrBelowBudget`
- Assertion intent: If this fails, the primary dynamic scheduling rule is broken: budget-or-lower accounts must be sorted by highest multiplier first.
- Data fixture design: Dynamic group `10`, budget `8`, accounts `A1=3x`, `A2=6x`, `A3=10x`; all active, schedulable, model-compatible, load `0`, acquire succeeds.
- Expected result: scheduler returns `account_id == A2`; it must not choose lower `A1` by load/LRU/priority before dynamic preference.
- Reuse existing test: Same fixture as `backend/internal/service/openai_gateway_service_test.go:769`; keep as explicit Rule 4-named test or alias in comment.
- Priority: P0.

Pseudo:

```go
// pseudo only; protects compareDynamicPricingAccountPreference
ctx := dynamicContext(budget=8)
repo := accounts(multiplier(1,3), multiplier(2,6), multiplier(3,10))
selection, err := svc.SelectAccountWithLoadAwareness(ctx, &groupID, "", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(2), selection.Account.ID)
```

### 2.5 Rule 4 Low Retry — Unavailable Higher In-Budget Falls Back To Lower In-Budget

- Test name: `TestDynamicScheduling_Rule4_LowRetryFallsBackHighToLowWithinBudget`
- Assertion intent: If this fails, runtime unavailability no longer retries lower budget-or-equal accounts in descending multiplier order.
- Data fixture design: Dynamic group `10`, budget `8`, accounts `A1=3x`, `A2=6x`, `A3=10x`. Make `A2` unavailable via `stubConcurrencyCache.acquireResults[2] = false` or load/wait plan depending on target function; keep `A1` available; keep `A3` over budget and average blocked or not needed.
- Expected result: scheduler returns `account_id == A1`, not `A3`, and no `ErrDynamicPricingBudgetExceeded`.
- Reuse existing test: New test; can reuse `dynamicOpenAIAccount`, `stubConcurrencyCache.acquireResults`, and service setup from `openai_gateway_service_test.go:769`.
- Priority: P0.

Pseudo:

```go
// pseudo only; protects compareDynamicPricingAccountPreference and runtime retry order
cache := stubConcurrencyCache{acquireResults: map[int64]bool{2:false, 1:true}}
selection, err := svc.SelectAccountWithLoadAwareness(ctx, &groupID, "", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(1), selection.Account.ID)
```

### 2.6 Rule 4 High Fallback Allowed — Seven-Day Average Below Budget

- Test name: `TestDynamicScheduling_Rule4_HighFallbackAllowedWhenSevenDayAverageBelowBudget`
- Assertion intent: If this fails, over-budget accounts may be incorrectly blocked even when the 7-day average multiplier is below budget.
- Data fixture design: Dynamic group `10`, budget `8`, accounts `A1=6x`, `A2=10x`; make `A1` unavailable through load/acquire; set `stubOpenAIUsageLogRepo.stats` to `TotalCost=100`, `TotalActualCost=790`, `TotalRequests=10`, so average `7.9 < 8`.
- Expected result: scheduler returns `account_id == A2`.
- Reuse existing test: Existing `backend/internal/service/openai_gateway_service_test.go:865` already covers this; keep and rename/comment for Rule 4 high fallback allowed.
- Priority: P0.

Pseudo:

```go
// pseudo only; protects isAccountWithinDynamicBudget and dynamicBudgetWindow
usageStats := UsageStats{TotalCost:100, TotalActualCost:790, TotalRequests:10}
selection, err := svc.SelectAccountWithLoadAwareness(ctxWithStats(usageStats), &groupID, "", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(2), selection.Account.ID)
```

### 2.7 Rule 4 High Fallback Blocked — Seven-Day Average At Or Above Budget

- Test name: `TestDynamicScheduling_Rule4_HighFallbackBlockedWhenSevenDayAverageAtOrAboveBudget`
- Assertion intent: If this fails, over-budget accounts can leak through after budget is exhausted, causing uncontrolled cost drift.
- Data fixture design: Dynamic group `10`, budget `8`, accounts `A1=6x`, `A2=10x`; make `A1` unavailable; set usage stats `TotalCost=100`, `TotalActualCost=800` for equality and optionally `810` for above-budget table subcase.
- Expected result: scheduler returns `ErrDynamicPricingBudgetExceeded`; selection is nil; external message maps to `No available count` in higher layers.
- Reuse existing test: Existing `backend/internal/service/openai_gateway_service_test.go:914` covers equality. Extend table to include above-budget `8.1`.
- Priority: P0.

Pseudo:

```go
// pseudo only; protects ErrDynamicPricingBudgetExceeded and isAccountWithinDynamicBudget
usageStats := UsageStats{TotalCost:100, TotalActualCost:800, TotalRequests:10}
selection, err := svc.SelectAccountWithLoadAwareness(ctxWithStats(usageStats), &groupID, "", "gpt-4", nil)
require.ErrorIs(t, err, ErrDynamicPricingBudgetExceeded)
require.Nil(t, selection)
```

### 2.8 Boundary Equality — Multiplier Equal To Budget Is In First Bucket

- Test name: `TestDynamicScheduling_Boundary_AccountMultiplierEqualBudgetIsWithinBudget`
- Assertion intent: If this fails, equality was changed from allowed to blocked, contradicting T018 and `dynamicBudgetEpsilon` semantics.
- Data fixture design: Dynamic group `10`, budget `8`, accounts `A1=6x`, `A2=8x`, `A3=10x`; all active and available; no usage history needed.
- Expected result: scheduler returns `account_id == A2`; it must not treat `8x` as over-budget.
- Reuse existing test: New focused boundary test; can share the first-bucket table from §3.
- Priority: P1.

Pseudo:

```go
// pseudo only; protects dynamicBudgetEpsilon and <= budget behavior
repo := accounts(multiplier(1,6), multiplier(2,8), multiplier(3,10))
selection, err := svc.SelectAccountWithLoadAwareness(ctxBudget(8), &groupID, "", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(2), selection.Account.ID)
```

### 2.9 Boundary No History — High Fallback Is Blocked Without Seven-Day Cost Basis

- Test name: `TestDynamicScheduling_Boundary_NoHistoryBlocksHighFallback`
- Assertion intent: If this fails, the no-history behavior changed; current code blocks high fallback when `TotalCost == 0` because average cannot be computed.
- Data fixture design: Dynamic group `10`, budget `8`, only high account `A2=10x` available; no in-budget accounts, or make in-budget account unavailable; `stubOpenAIUsageLogRepo.stats` nil or `TotalCost=0`, `TotalActualCost=0`, `TotalRequests=0`.
- Expected result: scheduler returns `ErrDynamicPricingBudgetExceeded` and nil selection. If product later chooses to allow first high request, this test must be updated only after a new user-approved decision.
- Reuse existing test: New boundary test; use `stubOpenAIUsageLogRepo{}` default empty stats.
- Priority: P1.

Pseudo:

```go
// pseudo only; protects isAccountWithinDynamicBudget state.currentStandardCost <= 0 branch
svc.usageLogRepo = stubOpenAIUsageLogRepo{}
selection, err := svc.SelectAccountWithLoadAwareness(ctxBudget(8), &groupID, "", "gpt-4", nil)
require.ErrorIs(t, err, ErrDynamicPricingBudgetExceeded)
require.Nil(t, selection)
```

### 2.10 Boundary Order — Multiple High Accounts Choose Lowest Over-Budget First

- Test name: `TestDynamicScheduling_Boundary_HighFallbackChoosesLowestOverBudgetMultiplierFirst`
- Assertion intent: If this fails, the high-fallback ordering changed from cost-minimizing order to arbitrary priority/load/LRU order.
- Data fixture design: Dynamic group `10`, budget `8`, accounts `A1=10x`, `A2=12x`, `A3=20x`; all high accounts active; set usage stats average `7.5 < 8`; equal load and priority for all.
- Expected result: scheduler returns `account_id == A1` (10x), the lowest over-budget multiplier.
- Reuse existing test: New test; can be a table case with high fallback allowed.
- Priority: P1.

Pseudo:

```go
// pseudo only; protects compareDynamicPricingAccountPreference over-budget ordering
repo := accounts(multiplier(1,10), multiplier(2,12), multiplier(3,20))
selection, err := svc.SelectAccountWithLoadAwareness(ctxAvgBelowBudget(), &groupID, "", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(1), selection.Account.ID)
```

### 2.11 Runtime Recheck — Disabled Or Full Candidate Is Skipped

- Test name: `TestDynamicScheduling_RuntimeRecheck_SkipsUnavailableCandidateAndKeepsDynamicOrder`
- Assertion intent: If this fails, dynamic ordering is only applied before runtime checks and the final selected account can violate the retry order.
- Data fixture design: Dynamic group `10`, budget `8`, accounts `A1=3x`, `A2=6x`, `A3=10x`; make `A2` pass initial listing but fail runtime acquisition using `stubConcurrencyCache.acquireResults[2] = false`; make `A1` acquire successfully; keep `A3` high and blocked by average at budget.
- Expected result: scheduler skips `A2`, then returns `A1`; if all in-budget candidates fail and high is blocked, returns `ErrDynamicPricingBudgetExceeded`.
- Reuse existing test: New test. Related to existing load/acquire tests around `openai_gateway_service_test.go:642` and dynamic tests around `:769`.
- Priority: P1.

Pseudo:

```go
// pseudo only; protects runtime recheck path plus isAccountWithinDynamicBudget
cache := stubConcurrencyCache{acquireResults: map[int64]bool{2:false, 1:true}}
selection, err := svc.SelectAccountWithLoadAwareness(ctxBudget(8), &groupID, "", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(1), selection.Account.ID)
```

### 2.12 Sticky Session — Sticky Account Must Pass Dynamic Budget Admission

- Test name: `TestDynamicScheduling_StickySession_RechecksDynamicBudgetBeforeReusingStickyAccount`
- Assertion intent: If this fails, sticky sessions can bypass dynamic budget rules and keep routing to an over-budget account.
- Data fixture design: Dynamic group `10`, budget `8`, sticky cache maps `openai:sticky` to high account `A2=10x`; low account `A1=6x` is available; usage stats average `8.0 >= 8` so `A2` must be rejected; `stubGatewayCache.sessionBindings` starts with sticky binding.
- Expected result: scheduler does not return `A2`; it returns `A1` and updates/clears sticky binding. If no low account exists, expect `ErrDynamicPricingBudgetExceeded`.
- Reuse existing test: Extend sticky tests around `backend/internal/service/openai_gateway_service_test.go:537` and `:701`; no existing test combines sticky with dynamic budget admission.
- Priority: P1.

Pseudo:

```go
// pseudo only; protects sticky account recheck and isAccountWithinDynamicBudget
cache := &stubGatewayCache{sessionBindings: map[string]int64{"openai:sticky": 2}}
selection, err := svc.SelectAccountWithLoadAwareness(ctxBudgetAtAverage(), &groupID, "sticky", "gpt-4", nil)
require.NoError(t, err)
require.Equal(t, int64(1), selection.Account.ID)
require.NotEqual(t, int64(2), cache.sessionBindings["openai:sticky"])
```

## 3. Table-Driven Test Design Suggestions

### Cases That Should Be Table-Driven

The scheduling order and admission tests share the same small fixture shape: dynamic group, one API key, several accounts with different account-group billing multipliers, optional usage stats, optional disabled/acquire failure map.

Good table-driven candidates:

- Rule 2 budget resolution priority.
- Rule 4 first bucket.
- Rule 4 low retry.
- Rule 4 high fallback allowed.
- Rule 4 high fallback blocked.
- Boundary equality.
- Boundary no history.
- Boundary order.

Recommended case struct:

```go
// pseudo only
type dynamicSchedulingCase struct {
  name               string
  groupDefaultBudget *float64
  apiKeyBudget       *float64
  accountMultipliers map[int64]float64
  sevenDayTotalCost  float64
  sevenDayActualCost float64
  sevenDayRequests   int64
  disabledAccounts   map[int64]bool
  acquireResults     map[int64]bool
  stickyAccountID    *int64
  expectedAccountID  *int64
  expectedError      error
}
```

Field notes:

- `accountMultipliers`: maps account ID to `account_groups.billing_multiplier`.
- `groupDefaultBudget` and `apiKeyBudget`: isolate `resolveDynamicBudgetMultiplier` priority.
- `sevenDayTotalCost` / `sevenDayActualCost`: isolate current average multiplier (`TotalActualCost / TotalCost`).
- `disabledAccounts`: use when the candidate should be filtered before runtime acquisition.
- `acquireResults`: use when the candidate appears initially but fails runtime acquisition.
- `stickyAccountID`: use only for sticky-specific cases.
- `expectedError`: should use `ErrDynamicPricingBudgetExceeded` for dynamic budget-blocked no-account results.

### Cases That Should Stay Separate

Do not merge these into the main scheduling table unless the helper becomes too complex:

- Rule 1: fixed vs dynamic activation needs an explicit fixed-group control to prove fixed groups do not use dynamic sorting.
- Rule 3: billing write path lives in `RecordUsage`, not account selection; it should remain in record usage tests and run through both OpenAI and generic gateway paths.
- Runtime recheck: the important fixture is acquisition/runtime state, not just sort order.
- Sticky session: needs cache state and sticky binding assertions; merging it into general cases obscures the failure mode.

### Suggested Split

| Test file | Suggested tests |
|---|---|
| `backend/internal/service/openai_gateway_service_test.go` | Rule 1, Rule 2, Rule 4, boundary, runtime, sticky scheduling tests |
| `backend/internal/service/openai_gateway_record_usage_test.go` | OpenAI Rule 3 billing snapshot tests |
| `backend/internal/service/gateway_record_usage_test.go` | Generic gateway Rule 3 billing snapshot tests |

## 4. Fixture Reuse Checklist

### Existing Helpers / Mocks To Reuse

| Helper/mock | Location | Use |
|---|---|---|
| `dynamicOpenAIAccount(id, groupID, billingMultiplier)` | `backend/internal/service/openai_gateway_service_test.go:988` | Builds active OpenAI account with account-group billing multiplier |
| `stubOpenAIAccountRepo` | `backend/internal/service/openai_gateway_service_test.go:30` | Supplies candidate accounts and sticky lookup by ID |
| `stubOpenAIUsageLogRepo` | `backend/internal/service/openai_gateway_service_test.go:35` | Supplies 7-day `usagestats.UsageStats` for dynamic budget state |
| `stubConcurrencyCache` | `backend/internal/service/openai_gateway_service_test.go:96` | Simulates load, acquire success/failure, wait count |
| `stubGatewayCache` | `backend/internal/service/openai_gateway_service_test.go:379` | Simulates sticky session binding and updates |
| `newOpenAIRecordUsageServiceForTest` | `backend/internal/service/openai_gateway_record_usage_test.go` | Builds OpenAI record usage service with stub repos |
| `newGatewayRecordUsageServiceForTest` | `backend/internal/service/gateway_record_usage_test.go` | Builds generic gateway record usage service with stub repos |
| `openAIRecordUsageLogRepoStub` | record usage test files | Captures last usage log for multiplier assertions |
| `openAIRecordUsageUserRepoStub` | record usage test files | Captures deducted balance amount |

### Helpers Worth Adding During S4

These helpers are not required for T019, but would reduce duplicate fixture code when backend implements the tests:

```go
// pseudo only
func dynamicAPIKeyContext(groupID int64, apiKeyBudget, groupDefaultBudget *float64) context.Context
func buildDynamicAccounts(groupID int64, multipliers map[int64]float64) []Account
func buildUsageStats(avgMultiplier float64, totalCost float64) *usagestats.UsageStats
func buildAccountWithSevenDayAvg(multiplier float64, avgCost float64) Account
func ptrFloat(v float64) *float64
func expectSelectedAccount(t *testing.T, selection *AccountSelection, id int64)
```

Helper design constraints:

- Do not hide the exact multiplier values used by the rule under test.
- Prefer explicit `map[int64]float64{1:3, 2:6, 3:10}` over factory magic.
- Keep `usageStats.TotalActualCost / TotalCost` visually obvious, because that is the 7-day average gate.
- Add comments naming protected identifiers from T018 §6 near each helper or case.

## 5. CI Grep Linkage

T018 §6 proposes grep protection for these identifiers:

```text
GroupPricingModeDynamic
DefaultBudgetMultiplier
dynamicBudgetWindow
ErrDynamicPricingBudgetExceeded
resolveDynamicBudgetMultiplier
resolveBillingMultiplierForUsage
compareDynamicPricingAccountPreference
isAccountWithinDynamicBudget
withDynamicPricingBudgetState
GroupBillingMultiplier
EffectiveBillingMultiplier
BillingRateMultiplier
```

Each implemented test should include a short comment naming the protected identifier it guards. This makes grep failures actionable: if an identifier is renamed or removed, the owner can immediately see which behavior tests must be reviewed.

| Test | Protected identifiers to mention in test comment |
|---|---|
| Rule 1 | `GroupPricingModeDynamic`, `compareDynamicPricingAccountPreference` |
| Rule 2 | `resolveDynamicBudgetMultiplier`, `DefaultBudgetMultiplier` |
| Rule 3 | `resolveBillingMultiplierForUsage`, `GroupBillingMultiplier`, `EffectiveBillingMultiplier`, `BillingRateMultiplier` |
| Rule 4 first bucket | `compareDynamicPricingAccountPreference`, `GroupBillingMultiplier` |
| Rule 4 low retry | `compareDynamicPricingAccountPreference`, `isAccountWithinDynamicBudget` |
| Rule 4 high fallback allowed | `withDynamicPricingBudgetState`, `dynamicBudgetWindow`, `isAccountWithinDynamicBudget` |
| Rule 4 high fallback blocked | `ErrDynamicPricingBudgetExceeded`, `isAccountWithinDynamicBudget` |
| Boundary equality | `dynamicBudgetEpsilon`, `isAccountWithinDynamicBudget` |
| Boundary no history | `withDynamicPricingBudgetState`, `isAccountWithinDynamicBudget` |
| Boundary order | `compareDynamicPricingAccountPreference` |
| Runtime recheck | `isAccountWithinDynamicBudget`, `ErrDynamicPricingBudgetExceeded` |
| Sticky session | `isAccountWithinDynamicBudget`, `withDynamicPricingBudgetState` |

Suggested implementation comment pattern:

```go
// Protects T018 Rule 4: compareDynamicPricingAccountPreference must choose highest <= budget before high fallback.
```

CI should not only grep identifiers; it should also run these focused tests. Suggested focused command after S4 implementation:

```sh
go test ./backend/internal/service -run 'DynamicScheduling|DynamicBilling|DynamicPricing'
```

## 6. Landing Priority

### First Batch: P0 Core Rules

Implement first:

1. `TestDynamicScheduling_Rule1_DynamicGroupUsesDynamicSorting`
2. `TestDynamicScheduling_Rule2_APIKeyBudgetOverridesGroupDefaultAndSystemDefault`
3. `TestDynamicBilling_Rule3_UsageLogRateMultiplierComesFromAccountGroupBinding`
4. `TestDynamicScheduling_Rule4_FirstBucketChoosesHighestAtOrBelowBudget`
5. `TestDynamicScheduling_Rule4_LowRetryFallsBackHighToLowWithinBudget`
6. `TestDynamicScheduling_Rule4_HighFallbackAllowedWhenSevenDayAverageBelowBudget`
7. `TestDynamicScheduling_Rule4_HighFallbackBlockedWhenSevenDayAverageAtOrAboveBudget`

Reason: these cover the four user-approved rules directly. If any fail, dynamic multiplier may produce wrong bills or wrong account selection in production.

### Second Batch: P1 Boundaries

Implement second:

1. `TestDynamicScheduling_Boundary_AccountMultiplierEqualBudgetIsWithinBudget`
2. `TestDynamicScheduling_Boundary_NoHistoryBlocksHighFallback`
3. `TestDynamicScheduling_Boundary_HighFallbackChoosesLowestOverBudgetMultiplierFirst`

Reason: these lock current boundary behavior and prevent later “cleanup” from silently changing cost exposure.

### Third Batch: Existing-Coverage Extensions

Implement third:

1. `TestDynamicScheduling_RuntimeRecheck_SkipsUnavailableCandidateAndKeepsDynamicOrder`
2. `TestDynamicScheduling_StickySession_RechecksDynamicBudgetBeforeReusingStickyAccount`

Reason: runtime recheck and sticky paths have adjacent existing tests, but not enough dynamic-budget assertions. They are important, but slightly lower priority because normal selection and billing paths must be locked first.

## 7. Red Lines / Non-Goals

- This plan does not add actual tests.
- This plan does not modify existing tests.
- This plan does not change T018 behavior, even where T018 lists open questions.
- S4 implementation should not rewrite scheduling logic just to make tests easier.
- Do not use HTTP contract/golden fixtures for these cases; they are backend service isolation tests.
- Do not mix product subscription `debit_multiplier` into dynamic scheduling fixtures.
- Do not assert UI-visible `dynamic_budget_matched_multiplier` here; that belongs to API/DTO contract tests.

## 8. Feedback For Foreman

The requested 12 assertions were all found in T018 §6 and mapped one-to-one in §2. No extra or missing assertion was detected.

Observations for S4 implementers:

- Existing tests already cover Rule 4 first bucket, high fallback allowed, high fallback blocked, and dynamic billing write path.
- Missing tests are mainly isolation and boundary guards: Rule 1 fixed-vs-dynamic control, Rule 2 budget priority, low retry, equality, no-history behavior, high-fallback ordering, runtime recheck, and sticky recheck.
- T018 has an open question about no-history high fallback. This plan intentionally locks current code behavior: no 7-day cost basis blocks high fallback. If user changes that decision, update §2.9 before implementation.
