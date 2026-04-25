import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SubscriptionsView from '../SubscriptionsView.vue'

const { getMySubscriptions, showError, fetchActive, mockProductStore } = vi.hoisted(() => ({
  getMySubscriptions: vi.fn(),
  showError: vi.fn(),
  fetchActive: vi.fn(),
  mockProductStore: {
    items: [] as any[],
    fetchActive: vi.fn()
  }
}))

vi.mock('@/api/subscriptions', () => ({
  default: {
    getMySubscriptions
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError
  })
}))

vi.mock('@/stores/subscriptionProducts', () => ({
  useSubscriptionProductStore: () => mockProductStore
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, any>) => {
        if (key === 'userSubscriptions.status.active') return 'Active'
        if (key === 'userSubscriptions.expires') return 'Expires'
        if (key === 'userSubscriptions.daily') return 'Daily'
        if (key === 'userSubscriptions.weekly') return 'Weekly'
        if (key === 'userSubscriptions.monthly') return 'Monthly'
        if (key === 'userSubscriptions.unlimited') return 'Unlimited'
        if (key === 'userSubscriptions.dailyQuotaBreakdown') {
          return `Yesterday carryover $${params?.carryover} + today $${params?.today} = today available $${params?.total}`
        }
        return key
      }
    })
  }
})

describe('SubscriptionsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getMySubscriptions.mockResolvedValue([])
    mockProductStore.items = []
    mockProductStore.fetchActive = fetchActive
    fetchActive.mockResolvedValue([])
  })

  it('shows product daily carryover and unlimited weekly or monthly usage', async () => {
    mockProductStore.items = [
      {
        product_id: 101,
        subscription_id: 501,
        code: 'gpt_daily_45',
        name: 'GPT Daily 45',
        description: 'GPT subscription',
        expires_at: null,
        status: 'active',
        daily_usage_usd: 3,
        weekly_usage_usd: 12.34,
        monthly_usage_usd: 56.78,
        daily_limit_usd: 45,
        weekly_limit_usd: 0,
        monthly_limit_usd: 0,
        daily_carryover_in_usd: 2,
        daily_carryover_remaining_usd: 1.25,
        groups: [
          {
            group_id: 21,
            group_name: 'plus/team mixed pool',
            debit_multiplier: 1,
            status: 'active',
            sort_order: 1
          },
          {
            group_id: 33,
            group_name: 'pro pool',
            debit_multiplier: 1.5,
            status: 'active',
            sort_order: 2
          }
        ]
      }
    ]

    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true
        }
      }
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('$3.00 / $47.00')
    expect(text).toContain('Yesterday carryover $2.00 + today $45.00 = today available $47.00')
    expect(text).toContain('$12.34 / Unlimited')
    expect(text).toContain('$56.78 / Unlimited')
  })
})
