import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import SubscriptionProgressMini from '../SubscriptionProgressMini.vue'

const mockStore = {
  activeSubscriptions: [] as any[],
  hasActiveSubscriptions: false,
  fetchActiveSubscriptions: vi.fn().mockResolvedValue(undefined)
}

const mockProductStore = {
  items: [] as any[],
  hasActiveProducts: false,
  fetchActive: vi.fn().mockResolvedValue(undefined)
}

vi.mock('@/stores', () => ({
  useSubscriptionStore: () => mockStore,
  useSubscriptionProductStore: () => mockProductStore
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, any>) => {
      if (key === 'subscriptionProgress.viewDetails') return 'View subscription details'
      if (key === 'subscriptionProgress.title') return 'My Subscriptions'
      if (key === 'subscriptionProgress.activeCount') return `${params?.count ?? 0} active subscriptions`
      if (key === 'subscriptionProgress.daily') return 'Daily'
      if (key === 'subscriptionProgress.weekly') return 'Weekly'
      if (key === 'subscriptionProgress.monthly') return 'Monthly'
      if (key === 'subscriptionProgress.last7DaysShort') return '7d'
      if (key === 'subscriptionProgress.last30DaysShort') return '30d'
      if (key === 'subscriptionProgress.viewAll') return 'View all subscriptions'
      if (key === 'subscriptionProgress.daysRemaining') return `${params?.days ?? 0} days left`
      if (key === 'subscriptionProgress.todayAvailable') {
        return `Today available $${params?.total} (includes carryover $${params?.carryover})`
      }
      if (key === 'subscriptionProgress.carryoverRule') {
        return 'Carryover is used first and expires at end of day.'
      }
      return key
    }
  })
}))

describe('SubscriptionProgressMini', () => {
  beforeEach(() => {
    vi.useRealTimers()
    mockStore.fetchActiveSubscriptions.mockClear()
    mockStore.activeSubscriptions = []
    mockStore.hasActiveSubscriptions = false
    mockProductStore.fetchActive.mockClear()
    mockProductStore.items = []
    mockProductStore.hasActiveProducts = false
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders one product card with multiple groups', async () => {
    mockProductStore.items = [
      {
        product_id: 101,
        subscription_id: 501,
        code: 'gpt_team',
        name: 'GPT Team',
        description: 'Shared GPT access',
        expires_at: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
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
    mockProductStore.hasActiveProducts = true

    const wrapper = mount(SubscriptionProgressMini, {
      global: {
        stubs: {
          Icon: true,
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })

    await wrapper.find('button').trigger('click')

    expect(wrapper.text()).toContain('GPT Team')
    expect(wrapper.text()).toContain('gpt-4')
    expect(wrapper.text()).toContain('gpt-4o')
    expect(wrapper.text()).toContain('$4.00/$10.00')
    expect(wrapper.text()).not.toContain('$4.00/$12.00')
    expect(wrapper.text()).toContain('$17.50/$100.00')
    expect(wrapper.text()).toContain('Today available $12.00 (includes carryover $2.00)')
  })

  it('does not render legacy subscriptions covered by active products', async () => {
    mockStore.activeSubscriptions = [
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
        expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
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
    ]
    mockStore.hasActiveSubscriptions = true
    mockProductStore.items = [
      {
        product_id: 101,
        subscription_id: 501,
        code: 'gpt_daily_45',
        name: 'GPT Daily 45',
        description: 'GPT subscription',
        expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
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
          { group_id: 21, group_name: 'Legacy Plus Pool', debit_multiplier: 1, status: 'active', sort_order: 1 }
        ]
      }
    ]
    mockProductStore.hasActiveProducts = true

    const wrapper = mount(SubscriptionProgressMini, {
      global: {
        stubs: {
          Icon: true,
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })

    await wrapper.find('button').trigger('click')

    const text = wrapper.text()
    expect(text).toContain('GPT Daily 45')
    expect(text).toContain('$12.34/∞')
    expect(text).toContain('$56.78/∞')
    expect(text).not.toContain('Old group subscription')
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
    mockProductStore.hasActiveProducts = true

    const wrapper = mount(SubscriptionProgressMini, {
      global: {
        stubs: {
          Icon: true,
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })

    await wrapper.find('button').trigger('click')

    expect(wrapper.text()).toContain('31 days left')
  })
})
