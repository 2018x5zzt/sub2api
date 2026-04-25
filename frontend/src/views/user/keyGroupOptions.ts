import type {
  ActiveSubscriptionProduct,
  Group,
  GroupPlatform,
  SubscriptionType
} from '@/types'

export interface ApiKeyGroupOption {
  [key: string]: unknown
  value: number
  label: string
  description: string | null
  rate?: number
  userRate: number | null
  subscriptionType: SubscriptionType
  platform: GroupPlatform
  sourceProductName?: string
  productId?: number
  debitMultiplier?: number
}

export function buildApiKeyGroupOptions(
  groups: Group[],
  userGroupRates: Record<number, number>,
  products: ActiveSubscriptionProduct[] = []
): ApiKeyGroupOption[] {
  const options = new Map<number, ApiKeyGroupOption>()

  for (const group of groups) {
    options.set(group.id, {
      value: group.id,
      label: group.name,
      description: group.description,
      rate: group.rate_multiplier,
      userRate: userGroupRates[group.id] ?? null,
      subscriptionType: group.subscription_type,
      platform: group.platform
    })
  }

  for (const product of products) {
    const sortedGroups = [...(product.groups || [])].sort(
      (a, b) => (a.sort_order || 0) - (b.sort_order || 0)
    )

    for (const productGroup of sortedGroups) {
      const groupID = productGroup.group_id
      const existing = options.get(groupID)
      const multiplier = productGroup.debit_multiplier
      const productDescription = `${product.name} · ${multiplier}x`

      options.set(groupID, {
        value: groupID,
        label: existing?.label || productGroup.group_name,
        description: existing?.description || productDescription,
        rate: existing?.rate,
        userRate: existing?.userRate ?? null,
        subscriptionType: existing?.subscriptionType || 'subscription',
        platform: productGroup.platform || existing?.platform || 'openai',
        sourceProductName: product.name,
        productId: product.product_id,
        debitMultiplier: multiplier
      })
    }
  }

  return Array.from(options.values())
}
