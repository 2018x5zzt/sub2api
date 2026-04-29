import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
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
        if (key === 'userSubscriptions.daysRemaining') return `${params?.days ?? 0} days remaining`
        if (key === 'userSubscriptions.last7DaysUsage') return 'Past 7 days usage'
        if (key === 'userSubscriptions.last30DaysUsage') return 'Past 30 days usage'
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
    vi.useRealTimers()
    vi.clearAllMocks()
    getMySubscriptions.mockResolvedValue([])
    mockProductStore.items = []
    mockProductStore.fetchActive = fetchActive
    fetchActive.mockResolvedValue([])
  })

  afterEach(() => {
    vi.useRealTimers()
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
    expect(text).not.toContain('$3.00 / $45.00')
    expect(text).toContain('Past 7 days usage')
    expect(text).toContain('Past 30 days usage')
    expect(text).toContain('$12.34 / Unlimited')
    expect(text).toContain('$56.78 / Unlimited')
  })

  it('hides legacy group subscriptions that are covered by an active product', async () => {
    getMySubscriptions.mockResolvedValue([
      {
        id: 11,
        user_id: 791,
        group_id: 21,
        status: 'active',
        daily_usage_usd: 1,
        weekly_usage_usd: 16,
        monthly_usage_usd: 16,
        daily_carryover_in_usd: 0,
        daily_effective_limit_usd: 45,
        daily_remaining_total_usd: 44,
        daily_remaining_carryover_usd: 0,
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
          type: 'subscription',
          daily_limit_usd: 45,
          weekly_limit_usd: 0,
          monthly_limit_usd: 0
        }
      }
    ])
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
    expect(text).toContain('GPT Daily 45')
    expect(text).not.toContain('Old group subscription')
  })

  it('does not show carryover for legacy group subscriptions', async () => {
    getMySubscriptions.mockResolvedValue([
      {
        id: 12,
        user_id: 791,
        group_id: 31,
        status: 'active',
        daily_usage_usd: 20,
        weekly_usage_usd: 20,
        monthly_usage_usd: 20,
        daily_carryover_in_usd: 15,
        daily_effective_limit_usd: 60,
        daily_remaining_total_usd: 40,
        daily_remaining_carryover_usd: 10,
        daily_window_start: null,
        weekly_window_start: null,
        monthly_window_start: null,
        created_at: '2026-04-14T01:05:05Z',
        updated_at: '2026-04-26T16:52:16Z',
        expires_at: '2026-05-25T16:52:16Z',
        group: {
          id: 31,
          name: 'Legacy Daily',
          description: 'Old group subscription',
          type: 'subscription',
          daily_limit_usd: 45,
          weekly_limit_usd: 0,
          monthly_limit_usd: 0
        }
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
    expect(text).toContain('Legacy Daily')
    expect(text).toContain('$20.00 / $45.00')
    expect(text).not.toContain('Yesterday carryover')
    expect(text).not.toContain('$20.00 / $60.00')
  })

  it('keeps fractional remaining days rounded up for product subscriptions', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-04-26T00:30:00Z'))
    mockProductStore.items = [
      {
        product_id: 101,
        subscription_id: 501,
        code: 'gpt_daily_45',
        name: 'GPT Daily 45',
        description: 'GPT subscription',
        expires_at: '2026-05-26T01:00:00Z',
        status: 'active',
        daily_usage_usd: 0,
        weekly_usage_usd: 0,
        monthly_usage_usd: 0,
        daily_limit_usd: 45,
        weekly_limit_usd: 0,
        monthly_limit_usd: 0,
        daily_carryover_in_usd: 0,
        daily_carryover_remaining_usd: 0,
        groups: []
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

    expect(wrapper.text()).toContain('31 days remaining')
  })
})
