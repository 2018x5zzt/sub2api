import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import XlabOAuthConsentView from '../XlabOAuthConsentView.vue'

const { authorize, routeQuery, replaceLocation } = vi.hoisted(() => ({
  authorize: vi.fn(),
  routeQuery: {
    client_id: 'miku-client',
    redirect_uri: 'https://ai.mikuapi.org/auth/xlab/callback',
    response_type: 'code',
    scope: 'profile email',
    state: 'state-from-miku',
  } as Record<string, string>,
  replaceLocation: vi.fn(),
}))

vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
  useRoute: () => ({ query: routeQuery }),
}))

vi.mock('@/api', () => ({
  xlabOAuthAPI: { authorize },
}))

describe('XlabOAuthConsentView', () => {
  beforeEach(() => {
    authorize.mockReset()
    replaceLocation.mockReset()
    routeQuery.client_id = 'miku-client'
    routeQuery.redirect_uri = 'https://ai.mikuapi.org/auth/xlab/callback'
    routeQuery.response_type = 'code'
    routeQuery.scope = 'profile email'
    routeQuery.state = 'state-from-miku'
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { replace: replaceLocation },
    })
  })

  it('authorizes Miku and redirects to the returned callback URL', async () => {
    authorize.mockResolvedValue({
      redirect_uri: 'https://ai.mikuapi.org/auth/xlab/callback?code=code-from-xlab&state=state-from-miku',
    })

    mount(XlabOAuthConsentView)
    await flushPromises()

    expect(authorize).toHaveBeenCalledWith({
      client_id: 'miku-client',
      redirect_uri: 'https://ai.mikuapi.org/auth/xlab/callback',
      response_type: 'code',
      scope: 'profile email',
      state: 'state-from-miku',
    })
    expect(replaceLocation).toHaveBeenCalledWith(
      'https://ai.mikuapi.org/auth/xlab/callback?code=code-from-xlab&state=state-from-miku',
    )
  })

  it('shows an error when the authorization request cannot be completed', async () => {
    authorize.mockRejectedValue({ message: 'invalid redirect_uri' })

    const wrapper = mount(XlabOAuthConsentView)
    await flushPromises()

    expect(wrapper.text()).toContain('invalid redirect_uri')
    expect(replaceLocation).not.toHaveBeenCalled()
  })
})
