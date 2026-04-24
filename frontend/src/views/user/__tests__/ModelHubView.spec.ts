import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ModelHubView from '../ModelHubView.vue'

const { getModels, copyToClipboard } = vi.hoisted(() => ({
  getModels: vi.fn(),
  copyToClipboard: vi.fn(),
}))

const messages: Record<string, string> = {
  'common.refresh': 'Refresh',
  'modelHub.title': 'Model Hub',
  'modelHub.description': 'Browse models',
  'modelHub.eyebrow': 'Models',
  'modelHub.copyVisible': 'Copy visible',
  'modelHub.groupsLabel': 'Groups',
  'modelHub.uniqueModelsLabel': 'Unique models',
  'modelHub.visibleModelsLabel': 'Visible models',
  'modelHub.platformsLabel': 'Platforms',
  'modelHub.searchLabel': 'Search',
  'modelHub.searchPlaceholder': 'Search models',
  'modelHub.platformFilterLabel': 'Platforms',
  'modelHub.groupFilterLabel': 'Groups',
  'modelHub.allPlatforms': 'All platforms',
  'modelHub.allGroups': 'All groups',
  'modelHub.clearFilters': 'Clear filters',
  'modelHub.modelCount': '{count} models',
  'modelHub.sourceDefault': 'Default',
  'modelHub.pricingComputedWithRate': 'Prices include {rate}x',
  'modelHub.copyGroup': 'Copy group',
  'modelHub.noModelsInGroup': 'No models',
  'modelHub.pricingUnavailable': 'Unavailable',
  'modelHub.rateShort': 'Rate',
  'modelHub.defaultPriceShort': 'Default',
  'modelHub.perRequest': '/ request',
  'modelHub.emptyTitle': 'Empty',
  'modelHub.emptyDescription': 'No models',
}

vi.mock('@/api', () => ({
  userGroupsAPI: {
    getModels,
  },
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('@/utils/format', () => ({
  formatCurrency: (value: number) => `$${value.toFixed(2)}`,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        const template = messages[key] ?? key
        return template.replace(/\{(\w+)\}/g, (_, name) => String(params?.[name] ?? ''))
      },
    }),
  }
})

describe('ModelHubView', () => {
  beforeEach(() => {
    getModels.mockReset()
    copyToClipboard.mockReset()
    copyToClipboard.mockResolvedValue(true)
  })

  it('shows rate and per-request pricing for image models', async () => {
    getModels.mockResolvedValue([
      {
        group: {
          id: 1,
          name: 'OpenAI Image',
          platform: 'openai',
          description: '',
          subscription_type: 'standard',
          rate_multiplier: 1.2,
        },
        source: 'default',
        effective_rate_multiplier: 1.2,
        user_rate_multiplier: null,
        models: [
          {
            id: 'gpt-image-1',
            display_name: 'GPT Image 1',
            pricing: {
              currency: 'USD',
              billing_mode: 'image',
              default_price_per_request: 0.12,
              request_tiers: [
                {
                  tier_label: '1K',
                  price_per_request: 0.06,
                },
              ],
            },
          },
        ],
      },
    ] as any)

    const wrapper = mount(ModelHubView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          EmptyState: { template: '<div><slot name="action" /></div>' },
          GroupBadge: { template: '<div />' },
          LoadingSpinner: { template: '<div />' },
          ModelIcon: { template: '<div />' },
          Icon: { template: '<div />' },
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Rate 1.2x')
    expect(text).toContain('Default $0.12 / request')
    expect(text).toContain('1K $0.06 / request')
  })
})
