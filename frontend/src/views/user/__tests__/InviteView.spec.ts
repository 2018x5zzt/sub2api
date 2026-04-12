import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import InviteView from '@/views/user/InviteView.vue'
import zh from '@/i18n/locales/zh'

vi.mock('@/api/invite', () => ({
  inviteAPI: {
    getSummary: vi.fn().mockResolvedValue({
      invite_code: 'HELLO123',
      invite_link: 'https://example.com/register?invite=HELLO123',
      invited_users_total: 4,
      invitees_recharge_total: 300,
      base_rewards_total: 9
    }),
    listRewards: vi.fn().mockResolvedValue({
      items: [
        {
          reward_role: 'invitee',
          reward_type: 'base_invite_reward',
          reward_amount: 3,
          created_at: '2026-04-11T08:00:00Z'
        }
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
  }
}))

describe('InviteView', () => {
  it('renders bilateral reward copy without exposing a fixed percentage', async () => {
    const i18n = createI18n({
      legacy: false,
      locale: 'zh',
      messages: { zh, en },
      // Vitest uses the runtime-only vue-i18n build, so tests provide a tiny compiler for plain strings.
      messageCompiler: (message: string) => (ctx: any) =>
        message.replace(/\{(\w+)\}/g, (_match, key) => String(ctx?.values?.[key] ?? `{${key}}`))
    })

    const wrapper = mount(InviteView, {
      global: {
        plugins: [i18n],
        stubs: { AppLayout: { template: '<div><slot /></div>' } }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('HELLO123')
    expect(wrapper.text()).toContain(zh.invite.description)
    expect(wrapper.text()).not.toMatch(/\d+(\.\d+)?%/)
  })

  it('keeps the english invite copy free of percentages', async () => {
    const i18n = createI18n({
      legacy: false,
      locale: 'en',
      messages: { zh, en },
      // Vitest uses the runtime-only vue-i18n build, so tests provide a tiny compiler for plain strings.
      messageCompiler: (message: string) => (ctx: any) =>
        message.replace(/\{(\w+)\}/g, (_match, key) => String(ctx?.values?.[key] ?? `{${key}}`))
    })

    const wrapper = mount(InviteView, {
      global: {
        plugins: [i18n],
        stubs: { AppLayout: { template: '<div><slot /></div>' } }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain(en.invite.description)
    expect(wrapper.text()).not.toMatch(/\d+(\.\d+)?%/)
  })
})
