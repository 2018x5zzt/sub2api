import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

describe('UseKeyModal', () => {
  const mountModal = (props: {
    baseUrl: string
    platform: 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'sora'
    allowMessagesDispatch?: boolean
  }) => mount(UseKeyModal, {
    props: {
      show: true,
      apiKey: 'sk-test',
      ...props
    },
    global: {
      stubs: {
        BaseDialog: {
          template: '<div><slot /><slot name="footer" /></div>'
        },
        Icon: {
          template: '<span />'
        }
      }
    }
  })

  it('renders updated GPT-5.4 mini/nano names in OpenCode config', async () => {
    const wrapper = mountModal({
      baseUrl: 'https://example.com/v1',
      platform: 'openai'
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Nano"')
  })

  it('uses the site root without /v1 for Anthropic use-key configs', async () => {
    const wrapper = mountModal({
      baseUrl: 'https://example.com/v1',
      platform: 'anthropic'
    })

    const codeBlocks = wrapper.findAll('pre code').map((block) => block.text())
    expect(codeBlocks.join('\n')).toContain('ANTHROPIC_BASE_URL="https://example.com"')
    expect(codeBlocks.join('\n')).not.toContain('ANTHROPIC_BASE_URL="https://example.com/v1"')

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.text()).toContain('"baseURL": "https://example.com"')
    expect(codeBlock.text()).not.toContain('"baseURL": "https://example.com/v1"')
  })

  it('adds /v1 for OpenAI use-key configs when the public endpoint is the site root', async () => {
    const wrapper = mountModal({
      baseUrl: 'https://example.com',
      platform: 'openai'
    })

    expect(wrapper.find('pre code').text()).toContain('base_url = "https://example.com/v1"')

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.text()).toContain('"baseURL": "https://example.com/v1"')
  })
})
