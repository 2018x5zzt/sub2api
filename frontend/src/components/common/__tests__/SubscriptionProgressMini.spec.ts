import { describe, it, expect, vi, beforeEach } from 'vitest'
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
    mockStore.fetchActiveSubscriptions.mockClear()
    mockStore.activeSubscriptions = []
    mockStore.hasActiveSubscriptions = false
    mockProductStore.fetchActive.mockClear()
    mockProductStore.items = []
    mockProductStore.hasActiveProducts = false
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
    expect(wrapper.text()).toContain('$17.50/$100.00')
    expect(wrapper.text()).toContain('Today available $12.00 (includes carryover $2.00)')
  })
})
