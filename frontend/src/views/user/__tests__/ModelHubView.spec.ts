import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ModelHubView from '../ModelHubView.vue'

const { getAvailable, getUserGroupRates, copyToClipboard } = vi.hoisted(() => ({
  getAvailable: vi.fn(),
  getUserGroupRates: vi.fn(),
  copyToClipboard: vi.fn(),
}))

const messages: Record<string, string> = {
  'modelHub.eyebrow': 'Models by Group',
  'modelHub.title': 'Model Hub',
  'modelHub.description': 'Browse models',
  'modelHub.searchLabel': 'Search',
  'modelHub.searchPlaceholder': 'Search',
  'modelHub.platformFilterLabel': 'Platform',
  'modelHub.groupFilterLabel': 'Group',
  'modelHub.allPlatforms': 'All Platforms',
  'modelHub.allGroups': 'All Groups',
  'modelHub.groupsLabel': 'Available Groups',
  'modelHub.uniqueModelsLabel': 'Unique Models',
  'modelHub.visibleModelsLabel': 'Visible Models',
  'modelHub.platformsLabel': 'Platforms',
  'modelHub.sourceDefault': 'Platform defaults',
  'modelHub.sourceMapping': 'Account mapping aggregate',
  'modelHub.sourceMixed': 'Defaults + account mappings',
  'modelHub.pricingComputedWithRate': 'Prices include {rate}x',
  'modelHub.rateShort': 'Rate',
  'modelHub.inputPriceShort': 'Input',
  'modelHub.outputPriceShort': 'Output',
  'modelHub.defaultPriceShort': 'Default',
  'modelHub.perMillionTokens': '/ 1M tokens',
  'modelHub.perRequest': '/ request',
  'modelHub.perImage': '/ image',
  'modelHub.pricingUnavailable': 'Pricing unavailable',
  'modelHub.copyVisible': 'Copy visible results',
  'modelHub.copyGroup': 'Copy group models',
  'modelHub.copiedModel': 'Model name copied',
  'modelHub.copiedGroup': 'Group model list copied',
  'modelHub.copiedVisible': 'Visible results copied',
  'modelHub.modelCount': '{count} models',
  'modelHub.clearFilters': 'Clear filters',
  'modelHub.emptyTitle': 'No models',
  'modelHub.emptyDescription': 'No groups or models match',
  'modelHub.noModelsInGroup': 'No models',
  'modelHub.loadFailedTitle': 'Failed to load model list',
  'modelHub.loadFailedDescription': 'Refresh and try again',
  'admin.groups.platforms.openai': 'OpenAI',
  'common.refresh': 'Refresh',
}

vi.mock('@/api/channels', () => ({
  default: {
    getAvailable,
  },
}))

vi.mock('@/api', () => ({
  userGroupsAPI: {
    getUserGroupRates,
  },
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        let message = messages[key] ?? key
        if (params) {
          for (const [name, value] of Object.entries(params)) {
            message = message.replace(`{${name}}`, String(value))
          }
        }
        return message
      },
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const GroupBadgeStub = {
  props: ['name', 'rateMultiplier', 'userRateMultiplier'],
  template: '<span>{{ name }} {{ userRateMultiplier ?? rateMultiplier }}x</span>',
}

describe('ModelHubView pricing display', () => {
  beforeEach(() => {
    getAvailable.mockReset()
    getUserGroupRates.mockReset()
    copyToClipboard.mockReset()
  })

  it('shows per-token API prices as rate-adjusted prices per 1M tokens', async () => {
    getAvailable.mockResolvedValue([
      {
        name: 'Primary',
        description: '',
        platforms: [
          {
            platform: 'openai',
            groups: [
              {
                id: 10,
                name: 'VIP',
                platform: 'openai',
                subscription_type: 'standard',
                rate_multiplier: 2,
                is_exclusive: false,
              },
            ],
            supported_models: [
              {
                name: 'gpt-demo',
                platform: 'openai',
                pricing: {
                  billing_mode: 'token',
                  input_price: 0.000003,
                  output_price: 0.000015,
                  cache_write_price: null,
                  cache_read_price: null,
                  image_output_price: null,
                  per_request_price: null,
                  intervals: [
                    {
                      min_tokens: 0,
                      max_tokens: 128000,
                      input_price: 0.000004,
                      output_price: 0.00002,
                      cache_write_price: null,
                      cache_read_price: null,
                      per_request_price: null,
                    },
                  ],
                },
              },
              {
                name: 'request-demo',
                platform: 'openai',
                pricing: {
                  billing_mode: 'per_request',
                  input_price: null,
                  output_price: null,
                  cache_write_price: null,
                  cache_read_price: null,
                  image_output_price: null,
                  per_request_price: 0.02,
                  intervals: [],
                },
              },
            ],
          },
        ],
      },
    ])
    getUserGroupRates.mockResolvedValue({})

    const wrapper = mount(ModelHubView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          GroupBadge: GroupBadgeStub,
          Icon: true,
          LoadingSpinner: true,
          ModelIcon: true,
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Rate 2x')
    expect(text).toContain('Input $6.00 / 1M tokens')
    expect(text).toContain('Output $30.00 / 1M tokens')
    expect(text).toContain('0-128K Input $8.00 / 1M tokens · Output $40.00 / 1M tokens')
    expect(text).toContain('Default $0.04 / request')
    expect(text).not.toContain('$0.000003 / 1M tokens')
  })
})
