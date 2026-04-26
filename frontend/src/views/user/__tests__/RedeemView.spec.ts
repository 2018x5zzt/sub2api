import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'
import RedeemView from '@/views/user/RedeemView.vue'

const mocks = vi.hoisted(() => ({
  routerPush: vi.fn(),
  refreshUser: vi.fn().mockResolvedValue(undefined),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  showWarning: vi.fn(),
  fetchActiveSubscriptions: vi.fn().mockResolvedValue(undefined),
  redeem: vi.fn(),
  getHistory: vi.fn(),
  getBenefitLeaderboard: vi.fn(),
  getPublicSettings: vi.fn()
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mocks.routerPush
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      id: 1,
      username: 'tester',
      email: 'tester@example.com',
      role: 'user',
      balance: 12.34,
      concurrency: 2,
      status: 'active',
      allowed_groups: null,
      created_at: '2026-04-13T08:00:00Z',
      updated_at: '2026-04-13T08:00:00Z'
    },
    refreshUser: mocks.refreshUser
  })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mocks.showError,
    showSuccess: mocks.showSuccess,
    showWarning: mocks.showWarning
  })
}))

vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({
    fetchActiveSubscriptions: mocks.fetchActiveSubscriptions
  })
}))

vi.mock('@/api', () => ({
  redeemAPI: {
    redeem: mocks.redeem,
    getHistory: mocks.getHistory,
    getBenefitLeaderboard: mocks.getBenefitLeaderboard
  },
  authAPI: {
    getPublicSettings: mocks.getPublicSettings
  }
}))

const BaseDialogStub = {
  props: {
    show: {
      type: Boolean,
      default: false
    },
    title: {
      type: String,
      default: ''
    }
  },
  template: `
    <section v-if="show">
      <h2>{{ title }}</h2>
      <slot />
      <slot name="footer" />
    </section>
  `
}

const createTestI18n = () =>
  createI18n({
    legacy: false,
    locale: 'zh',
    messages: { zh, en },
    messageCompiler: (message: string) => (ctx: any) =>
      message.replace(/\{(\w+)\}/g, (_match, key) => String(ctx?.values?.[key] ?? `{${key}}`))
  })

const mountView = async () => {
  const wrapper = mount(RedeemView, {
    global: {
      plugins: [createTestI18n()],
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: BaseDialogStub,
        Icon: true
      }
    }
  })

  await flushPromises()
  return wrapper
}

describe('RedeemView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.getHistory.mockResolvedValue([])
    mocks.getPublicSettings.mockResolvedValue({ contact_info: '' })
  })

  it('does not expose a page-level lucky leaderboard entry before redeem flow finishes', async () => {
    const wrapper = await mountView()

    expect(wrapper.text()).not.toContain(zh.redeem.viewLuckyLeaderboard)
  })

  it('keeps leaderboard access inside the repeat benefit dialog for already redeemed codes', async () => {
    mocks.redeem.mockRejectedValueOnce({
      response: {
        data: {
          reason: 'PROMO_CODE_ALREADY_USED',
          detail: 'already used'
        }
      }
    })
    mocks.getBenefitLeaderboard.mockResolvedValueOnce({
      code: 'BENEFIT2026',
      fixed_value: 5,
      random_pool_value: 20,
      random_remaining_value: 8,
      max_uses: 10,
      used_count: 4,
      entries: []
    })

    const wrapper = await mountView()

    await wrapper.get('#code').setValue('BENEFIT2026')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(mocks.getBenefitLeaderboard).toHaveBeenCalledWith('BENEFIT2026')
    expect(wrapper.text()).toContain(zh.redeem.repeatRedeemDialogTitle)
    expect(wrapper.text()).toContain(zh.redeem.viewLuckyLeaderboard)
  })

  it('prompts for username instead of opening leaderboard access when repeat redeem still lacks username', async () => {
    mocks.redeem.mockRejectedValueOnce({
      response: {
        data: {
          reason: 'PROMO_CODE_ALREADY_USED',
          detail: 'already used'
        }
      }
    })
    mocks.getBenefitLeaderboard.mockRejectedValueOnce({
      response: {
        data: {
          reason: 'PROMO_CODE_USERNAME_REQUIRED',
          detail: 'username required'
        }
      }
    })

    const wrapper = await mountView()

    await wrapper.get('#code').setValue('BENEFIT2026')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain(zh.redeem.usernameRequiredTitle)
    expect(wrapper.text()).not.toContain(zh.redeem.repeatRedeemDialogTitle)
  })

  it('tells users to create a subscription-group API key after redeeming a subscription', async () => {
    mocks.redeem.mockResolvedValueOnce({
      message: 'Code redeemed successfully',
      type: 'subscription',
      value: 30,
      group_name: '【订阅】plus/team混合池',
      validity_days: 30
    })

    const wrapper = await mountView()

    await wrapper.get('#code').setValue('SUB2026')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(mocks.showSuccess).toHaveBeenCalledWith(
      '订阅成功！去生成一个新的订阅分组 API Key 吧！'
    )
    expect(wrapper.text()).toContain('订阅成功！去生成一个新的订阅分组 API Key 吧！')
  })
})
