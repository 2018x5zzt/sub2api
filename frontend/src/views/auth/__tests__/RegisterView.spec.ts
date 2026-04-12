import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RegisterView from '@/views/auth/RegisterView.vue'

const authApiMocks = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  validateInvitationCodeMock: vi.fn().mockResolvedValue({
    valid: true
  })
}))

const pushMock = vi.fn()
const registerMock = vi.fn()
const routeMock = {
  query: {
    invite: 'hello123'
  }
}

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock
  }),
  useRoute: () => routeMock
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'zh' }
    })
  }
})

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    register: registerMock
  }),
  useAppStore: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn()
  })
}))

vi.mock('@/api/auth', () => ({
  getPublicSettings: authApiMocks.getPublicSettingsMock,
  validatePromoCode: vi.fn(),
  validateInvitationCode: authApiMocks.validateInvitationCodeMock
}))

describe('RegisterView', () => {
  beforeEach(() => {
    pushMock.mockReset()
    registerMock.mockReset()
    authApiMocks.getPublicSettingsMock.mockReset()
    authApiMocks.validateInvitationCodeMock.mockClear()
    routeMock.query = { invite: 'hello123' }
    authApiMocks.getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: false,
      promo_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      site_name: 'Sub2API',
      linuxdo_oauth_enabled: false,
      registration_email_suffix_whitelist: []
    })
  })

  it('locks invite code from invite-link query and shows locked helper copy', async () => {
    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /></div>' },
          LinuxDoOAuthSection: { template: '<div />' },
          TurnstileWidget: { template: '<div />' },
          Icon: { template: '<span />' },
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })

    await flushPromises()

    const inviteInput = wrapper.find('#invitation_code')
    expect(inviteInput.exists()).toBe(true)
    expect((inviteInput.element as HTMLInputElement).value).toBe('HELLO123')
    expect(inviteInput.attributes('readonly')).toBeDefined()
    expect(wrapper.text()).toContain('auth.invitationCodeLockedFromLink')
  })

  it('keeps invite code editable when no invite query is provided', async () => {
    routeMock.query = {}

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /></div>' },
          LinuxDoOAuthSection: { template: '<div />' },
          TurnstileWidget: { template: '<div />' },
          Icon: { template: '<span />' },
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })

    await flushPromises()
    const inviteInput = wrapper.find('#invitation_code')
    expect(inviteInput.attributes('readonly')).toBeUndefined()
    await inviteInput.setValue('FREECODE')
    expect((inviteInput.element as HTMLInputElement).value).toBe('FREECODE')
  })
})
