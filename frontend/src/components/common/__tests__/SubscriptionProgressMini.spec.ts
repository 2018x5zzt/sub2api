import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import SubscriptionProgressMini from '../SubscriptionProgressMini.vue'

const mockStore = {
  activeSubscriptions: [] as any[],
  hasActiveSubscriptions: false,
  fetchActiveSubscriptions: vi.fn().mockResolvedValue(undefined)
}

vi.mock('@/stores', () => ({
  useSubscriptionStore: () => mockStore
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
  })

  it('uses the effective daily limit and shows carryover guidance', async () => {
    mockStore.activeSubscriptions = [
      {
        id: 1,
        group_id: 88,
        status: 'active',
        daily_usage_usd: 50,
        weekly_usage_usd: 0,
        monthly_usage_usd: 0,
        daily_window_start: new Date().toISOString(),
        weekly_window_start: null,
        monthly_window_start: null,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        expires_at: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
        daily_carryover_in_usd: 15,
        daily_effective_limit_usd: 60,
        daily_remaining_total_usd: 10,
        daily_remaining_carryover_usd: 10,
        group: {
          id: 88,
          name: 'Pro',
          daily_limit_usd: 45,
          weekly_limit_usd: null,
          monthly_limit_usd: null
        }
      }
    ]
    mockStore.hasActiveSubscriptions = true

    const wrapper = mount(SubscriptionProgressMini, {
      global: {
        stubs: {
          Icon: true,
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })

    await wrapper.find('button').trigger('click')

    expect(wrapper.text()).toContain('$50.00/$60.00')
    expect(wrapper.text()).toContain('Today available $60.00 (includes carryover $15.00)')
    expect(wrapper.text()).toContain('Carryover is used first and expires at end of day.')
  })
})
