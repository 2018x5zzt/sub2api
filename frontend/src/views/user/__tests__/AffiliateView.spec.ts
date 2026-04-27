import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'
import AffiliateView from '@/views/user/AffiliateView.vue'

const mocks = vi.hoisted(() => ({
  getAffiliateDetail: vi.fn(),
  transferAffiliateQuota: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  refreshUser: vi.fn().mockResolvedValue(undefined),
  copyToClipboard: vi.fn().mockResolvedValue(undefined)
}))

vi.mock('@/api/user', () => ({
  default: {
    getAffiliateDetail: mocks.getAffiliateDetail,
    transferAffiliateQuota: mocks.transferAffiliateQuota
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mocks.showError,
    showSuccess: mocks.showSuccess
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    refreshUser: mocks.refreshUser
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: mocks.copyToClipboard
  })
}))

const createTestI18n = () =>
  createI18n({
    legacy: false,
    locale: 'zh',
    messages: { zh, en },
    messageCompiler: (message: string) => (ctx: any) =>
      message.replace(/\{(\w+)\}/g, (_match, key) => String(ctx?.values?.[key] ?? `{${key}}`))
  })

const mountView = async () => {
  const wrapper = mount(AffiliateView, {
    global: {
      plugins: [createTestI18n()],
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true
      }
    }
  })

  await flushPromises()
  return wrapper
}

describe('AffiliateView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.getAffiliateDetail.mockResolvedValue({
      user_id: 7,
      aff_code: 'BUILDER',
      aff_count: 5,
      effective_invitee_count: 3,
      aff_quota: 162,
      aff_frozen_quota: 0,
      aff_history_quota: 200,
      effective_rebate_rate_percent: 8,
      invitees: []
    })
  })

  it('explains the invite rebate ladder, balance payout, SKU factors, and registration concurrency separately', async () => {
    const wrapper = await mountView()
    const text = wrapper.text()

    expect(text).toContain('邀请返利')
    expect(text).toContain('当上包工头，Token 不用愁')
    expect(text).toContain('最高 20%')
    expect(text).toContain('以余额形式到账')
    expect(text).toContain('有充值记录的邀请人数')
    expect(text).toContain('余额 / 日卡')
    expect(text).toContain('100%')
    expect(text).toContain('周卡')
    expect(text).toContain('60%')
    expect(text).toContain('月卡')
    expect(text).toContain('30%')
    expect(text).toContain('并发不够？邀请来凑！')
    expect(text).toContain('默认 3')
    expect(text).toContain('邀请小号注册')
  })
})
