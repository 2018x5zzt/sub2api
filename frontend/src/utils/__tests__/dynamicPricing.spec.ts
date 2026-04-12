import { describe, expect, it } from 'vitest'

import { DEFAULT_DYNAMIC_BUDGET_MULTIPLIER, isDynamicPricingGroup, resolveApiKeyBudgetMultiplier, resolveGroupBudgetMultiplier } from '../dynamicPricing'

describe('dynamicPricing utils', () => {
  it('detects dynamic pricing groups', () => {
    expect(isDynamicPricingGroup({ pricing_mode: 'dynamic', default_budget_multiplier: 8 })).toBe(true)
    expect(isDynamicPricingGroup({ pricing_mode: 'fixed', default_budget_multiplier: null })).toBe(false)
    expect(isDynamicPricingGroup(null)).toBe(false)
  })

  it('resolves group budget multiplier with fallback default', () => {
    expect(resolveGroupBudgetMultiplier({ pricing_mode: 'dynamic', default_budget_multiplier: 9.5 })).toBe(9.5)
    expect(resolveGroupBudgetMultiplier({ pricing_mode: 'dynamic', default_budget_multiplier: null })).toBe(DEFAULT_DYNAMIC_BUDGET_MULTIPLIER)
    expect(resolveGroupBudgetMultiplier({ pricing_mode: 'fixed', default_budget_multiplier: 12 })).toBeNull()
  })

  it('resolves api key budget multiplier from key first and group default second', () => {
    expect(resolveApiKeyBudgetMultiplier({
      budget_multiplier: 10,
      group: { pricing_mode: 'dynamic', default_budget_multiplier: 8 } as any
    })).toBe(10)

    expect(resolveApiKeyBudgetMultiplier({
      budget_multiplier: null,
      group: { pricing_mode: 'dynamic', default_budget_multiplier: null } as any
    })).toBe(DEFAULT_DYNAMIC_BUDGET_MULTIPLIER)

    expect(resolveApiKeyBudgetMultiplier({
      budget_multiplier: 10,
      group: { pricing_mode: 'fixed', default_budget_multiplier: null } as any
    })).toBeNull()
  })
})
