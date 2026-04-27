import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import GroupBadge from '../GroupBadge.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('GroupBadge', () => {
  it('shows rate for subscription groups when alwaysShowRate is enabled', () => {
    const wrapper = mount(GroupBadge, {
      props: {
        name: 'subscription-openai',
        platform: 'openai',
        subscriptionType: 'subscription',
        rateMultiplier: 1.5,
        alwaysShowRate: true
      },
      global: {
        stubs: {
          PlatformIcon: { template: '<span />' }
        }
      }
    })

    expect(wrapper.text()).toContain('1.5x')
  })
})
