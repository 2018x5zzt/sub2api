import { describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mount } from '@vue/test-utils'

import ImageStudioView from '../ImageStudioView.vue'

const { authorize } = vi.hoisted(() => ({
  authorize: vi.fn().mockResolvedValue({
    redirect_uri: 'https://ai.mikuapi.org/auth/xlab/callback?code=code-from-xlab&state=image-studio',
  }),
}))

vi.mock('@/api', () => ({
  xlabOAuthAPI: { authorize },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('ImageStudioView', () => {
  it('loads the new Miku studio through an xlab-issued OAuth callback URL', async () => {
    const wrapper = mount(ImageStudioView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    const iframe = wrapper.get('iframe')

    expect(authorize).toHaveBeenCalledWith({
      client_id: 'miku-app',
      redirect_uri: 'https://ai.mikuapi.org/auth/xlab/callback?layout=horizontal',
      response_type: 'code',
      scope: 'profile email',
      state: 'image-studio',
    })
    expect(iframe.attributes('src')).toBe('https://ai.mikuapi.org/auth/xlab/callback?code=code-from-xlab&state=image-studio')
    expect(wrapper.text()).toContain('新版')
    expect(wrapper.text()).toContain('旧版入口')
  })

  it('switches to the gpt image playground when the legacy entry is clicked', async () => {
    const wrapper = mount(ImageStudioView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()
    await wrapper.get('[data-testid="legacy-image-studio"]').trigger('click')

    expect(wrapper.get('iframe').attributes('src')).toBe('https://iframe.mikuapi.org/?codexCli=true')
    expect(wrapper.text()).toContain('旧版')
    expect(wrapper.text()).toContain('新版入口')

    await wrapper.get('[data-testid="new-image-studio"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('iframe').attributes('src')).toBe('https://ai.mikuapi.org/auth/xlab/callback?code=code-from-xlab&state=image-studio')
  })
})
