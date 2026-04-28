import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'
import ProfileView from '@/views/user/ProfileView.vue'

const mocks = vi.hoisted(() => ({
  getPublicSettings: vi.fn()
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      id: 1,
      username: 'tester',
      email: 'tester@example.com',
      role: 'user',
      balance: 12.34,
      concurrency: 3,
      status: 'active',
      allowed_groups: null,
      created_at: '2026-04-13T08:00:00Z',
      updated_at: '2026-04-13T08:00:00Z'
    }
  })
}))

vi.mock('@/api', () => ({
  authAPI: {
    getPublicSettings: mocks.getPublicSettings
  }
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
  const wrapper = mount(ProfileView, {
    global: {
      plugins: [createTestI18n()],
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        ProfileInfoCard: true,
        ProfileEditForm: true,
        ProfilePasswordForm: true,
        ProfileTotpCard: true,
        Icon: true
      }
    }
  })

  await flushPromises()
  return wrapper
}

describe('ProfileView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.getPublicSettings.mockResolvedValue({ contact_info: '' })
  })

  it('在并发限制统计卡附近提示用户可通过邀请返利增加并发', async () => {
    const wrapper = await mountView()

    expect(wrapper.text()).toContain('并发限制')
    expect(wrapper.text()).toContain('点开邀请返利部分，可以增加自己的并发！')
  })
})
