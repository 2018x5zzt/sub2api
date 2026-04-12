import type { ApiKey, Group } from '@/types'

export const DEFAULT_DYNAMIC_BUDGET_MULTIPLIER = 8

type DynamicPricingGroupLike = Pick<Group, 'pricing_mode' | 'default_budget_multiplier'> | null | undefined
type DynamicPricingApiKeyLike = Pick<ApiKey, 'budget_multiplier' | 'group'> | null | undefined

export function isDynamicPricingGroup(group: DynamicPricingGroupLike): boolean {
  return group?.pricing_mode === 'dynamic'
}

export function resolveGroupBudgetMultiplier(group: DynamicPricingGroupLike): number | null {
  if (!isDynamicPricingGroup(group)) {
    return null
  }
  return group?.default_budget_multiplier ?? DEFAULT_DYNAMIC_BUDGET_MULTIPLIER
}

export function resolveApiKeyBudgetMultiplier(apiKey: DynamicPricingApiKeyLike): number | null {
  if (!apiKey || !isDynamicPricingGroup(apiKey.group)) {
    return null
  }
  return apiKey.budget_multiplier ?? resolveGroupBudgetMultiplier(apiKey.group)
}
