import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useSubscriptionProductStore } from '@/stores/subscriptionProducts'

const mockGetActive = vi.fn()

vi.mock('@/api/subscriptionProducts', () => ({
  default: {
    getActive: (...args: any[]) => mockGetActive(...args)
  }
}))

const fakeProducts = [
  {
    product_id: 101,
    subscription_id: 501,
    code: 'gpt_team',
    name: 'GPT Team',
    description: 'Shared GPT access',
    expires_at: '2026-05-25T12:00:00Z',
    status: 'active',
    daily_usage_usd: 4,
    weekly_usage_usd: 12,
    monthly_usage_usd: 17.5,
    daily_limit_usd: 10,
    weekly_limit_usd: 50,
    monthly_limit_usd: 100,
    daily_carryover_in_usd: 2,
    daily_carryover_remaining_usd: 1.25,
    groups: [
      { group_id: 201, group_name: 'gpt-4', debit_multiplier: 1.5, status: 'active', sort_order: 1 },
      { group_id: 202, group_name: 'gpt-4o', debit_multiplier: 1, status: 'active', sort_order: 2 }
    ]
  }
]

describe('useSubscriptionProductStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads subscription products into the store', async () => {
    mockGetActive.mockResolvedValue(fakeProducts)
    const store = useSubscriptionProductStore()

    const result = await store.fetchActive()

    expect(result).toEqual(fakeProducts)
    expect(store.items).toEqual(fakeProducts)
    expect(store.hasActiveProducts).toBe(true)
    expect(store.loading).toBe(false)
  })
})
