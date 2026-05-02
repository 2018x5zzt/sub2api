import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SubscriptionsView from '../SubscriptionsView.vue'

const { getMySubscriptions, getActiveProducts, showError } = vi.hoisted(() => ({
  getMySubscriptions: vi.fn(),
  getActiveProducts: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/subscriptions', () => ({
  default: {
    getMySubscriptions
  }
}))

vi.mock('@/api/subscriptionProducts', () => ({
  default: {
    getActive: getActiveProducts
  },
  subscriptionProductsAPI: {
    getActive: getActiveProducts
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError
  })
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, any>) => {
        if (key === 'userSubscriptions.status.active') return 'Active'
        if (key === 'userSubscriptions.expires') return 'Expires'
        if (key === 'userSubscriptions.noExpiration') return 'No expiration'
        if (key === 'userSubscriptions.daily') return 'Daily'
        if (key === 'userSubscriptions.weekly') return 'Weekly'
        if (key === 'userSubscriptions.monthly') return 'Monthly'
        if (key === 'userSubscriptions.daysRemaining') return `${params?.days ?? 0} days remaining`
        if (key === 'userSubscriptions.dailyQuotaBreakdown') {
          return `Yesterday carryover $${params?.carryover} + today $${params?.today} = today available $${params?.total}`
        }
        if (key === 'userSubscriptions.visibleGroups') return 'Included groups'
        if (key === 'userSubscriptions.groupMultiplier') return `${params?.multiplier}x`
        if (key === 'userSubscriptions.unlimited') return 'Unlimited'
        if (key === 'userSubscriptions.unlimitedDesc') return 'No usage limits'
        if (key === 'common.today') return 'today'
        if (key === 'common.tomorrow') return 'tomorrow'
        return key
      }
    })
  }
})

describe('SubscriptionsView product subscriptions', () => {
  beforeEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
    getMySubscriptions.mockResolvedValue([])
    getActiveProducts.mockResolvedValue([])
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders active product subscription groups and one-day daily carryover', async () => {
    getActiveProducts.mockResolvedValue([
      {
        product_id: 101,
        subscription_id: 501,
        code: 'gpt_daily_45',
        name: 'GPT Daily 45',
        description: 'Shared GPT product',
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
    ])

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
    expect(text).toContain('GPT Daily 45')
    expect(text).toContain('Shared GPT product')
    expect(text).toContain('$3.00 / $47.00')
    expect(text).toContain('Yesterday carryover $2.00 + today $45.00 = today available $47.00')
    expect(text).toContain('plus/team mixed pool')
    expect(text).toContain('pro pool')
    expect(text).toContain('1.5x')
  })

  it('hides legacy group subscriptions covered by an active product', async () => {
    getMySubscriptions.mockResolvedValue([
      {
        id: 11,
        user_id: 791,
        group_id: 21,
        status: 'active',
        daily_usage_usd: 1,
        weekly_usage_usd: 16,
        monthly_usage_usd: 16,
        daily_window_start: null,
        weekly_window_start: null,
        monthly_window_start: null,
        created_at: '2026-04-14T01:05:05Z',
        updated_at: '2026-04-26T16:52:16Z',
        expires_at: '2026-05-25T16:52:16Z',
        group: {
          id: 21,
          name: 'Legacy Plus Pool',
          description: 'Old group subscription',
          platform: 'openai',
          rate_multiplier: 1,
          is_exclusive: false,
          status: 'active',
          subscription_type: 'subscription',
          daily_limit_usd: 45,
          weekly_limit_usd: 0,
          monthly_limit_usd: 0
        }
      }
    ])
    getActiveProducts.mockResolvedValue([
      {
        product_id: 101,
        subscription_id: 501,
        code: 'gpt_daily_45',
        name: 'GPT Daily 45',
        description: 'Shared GPT product',
        expires_at: null,
        status: 'active',
        daily_usage_usd: 3,
        weekly_usage_usd: 0,
        monthly_usage_usd: 0,
        daily_limit_usd: 45,
        weekly_limit_usd: 0,
        monthly_limit_usd: 0,
        daily_carryover_in_usd: 0,
        daily_carryover_remaining_usd: 0,
        groups: [
          {
            group_id: 21,
            group_name: 'Legacy Plus Pool',
            debit_multiplier: 1,
            status: 'active',
            sort_order: 1
          }
        ]
      }
    ])

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
    expect(text).toContain('GPT Daily 45')
    expect(text).not.toContain('Old group subscription')
  })
})
