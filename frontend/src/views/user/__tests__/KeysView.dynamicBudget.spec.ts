import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import KeysView from '../KeysView.vue'

const {
  listKeys,
  updateKey,
  getDashboardApiKeysUsage,
  getAvailableGroups,
  getUserGroupRates,
  getPublicSettings,
  showSuccess,
  showError,
} = vi.hoisted(() => ({
  listKeys: vi.fn(),
  updateKey: vi.fn(),
  getDashboardApiKeysUsage: vi.fn(),
  getAvailableGroups: vi.fn(),
  getUserGroupRates: vi.fn(),
  getPublicSettings: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list: listKeys,
    update: updateKey,
  },
  usageAPI: {
    getDashboardApiKeysUsage,
  },
  userGroupsAPI: {
    getAvailable: getAvailableGroups,
    getUserGroupRates,
  },
  authAPI: {
    getPublicSettings,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess,
    showError,
  }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep: () => false,
    nextStep: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (!params) return key
        return Object.entries(params).reduce(
          (message, [name, value]) => message.replace(`{${name}}`, String(value)),
          key
        )
      },
    }),
  }
})

const fixedGroup = {
  id: 10,
  name: 'Fixed Pool',
  description: '',
  platform: 'openai',
  rate_multiplier: 1,
  pricing_mode: 'fixed',
  default_budget_multiplier: null,
  is_exclusive: false,
  status: 'active',
  subscription_type: 'standard',
}

const dynamicGroup = {
  id: 20,
  name: 'Dynamic Pool',
  description: '',
  platform: 'openai',
  rate_multiplier: 1,
  pricing_mode: 'dynamic',
  default_budget_multiplier: 12,
  is_exclusive: false,
  status: 'active',
  subscription_type: 'standard',
}

const apiKey = {
  id: 99,
  user_id: 7,
  key: 'sk-test-key',
  name: 'fixed key',
  group_id: fixedGroup.id,
  budget_multiplier: 6,
  status: 'active',
  ip_whitelist: [],
  ip_blacklist: [],
  last_used_at: null,
  quota: 0,
  quota_used: 0,
  expires_at: null,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  group: fixedGroup,
  rate_limit_5h: 0,
  rate_limit_1d: 0,
  rate_limit_7d: 0,
  usage_5h: 0,
  usage_1d: 0,
  usage_7d: 0,
  window_5h_start: null,
  window_1d_start: null,
  window_7d_start: null,
  reset_5h_at: null,
  reset_1d_at: null,
  reset_7d_at: null,
}

function mountKeysView() {
  return mount(KeysView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template:
            '<div><slot name="filters" /><slot name="actions" /><slot name="table" /><slot name="pagination" /></div>',
        },
        DataTable: {
          props: ['data'],
          template:
            '<div><div v-for="row in data" :key="row.id"><slot name="cell-group" :row="row" /><slot name="cell-actions" :row="row" /></div></div>',
        },
        BaseDialog: {
          props: ['show', 'title'],
          template:
            '<div v-if="show" role="dialog"><h2>{{ title }}</h2><slot /><slot name="footer" /></div>',
        },
        ConfirmDialog: true,
        Pagination: true,
        EmptyState: true,
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue', 'change'],
          methods: {
            onChange(event: Event) {
              const target = event.target as HTMLSelectElement
              const value = target.value === '' ? null : Number(target.value)
              const option = this.options.find((item: { value: number | null }) => item.value === value)
              this.$emit('update:modelValue', value)
              this.$emit('change', value, option)
            },
          },
          template:
            '<select :value="modelValue ?? \'\'" @change="onChange"><option v-for="option in options" :key="option.value ?? \'empty\'" :value="option.value">{{ option.label }}</option></select>',
        },
        SearchInput: true,
        Icon: true,
        UseKeyModal: true,
        EndpointPopover: true,
        GroupBadge: {
          props: ['name'],
          template: '<span>{{ name }}</span>',
        },
        GroupOptionItem: {
          props: ['name'],
          template: '<span>{{ name }}</span>',
        },
      },
    },
  })
}

describe('KeysView dynamic budget handling', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    vi.clearAllMocks()
    listKeys.mockResolvedValue({
      items: [{ ...apiKey, group: { ...fixedGroup } }],
      total: 1,
      pages: 1,
    })
    updateKey.mockResolvedValue({})
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    getAvailableGroups.mockResolvedValue([{ ...fixedGroup }, { ...dynamicGroup }])
    getUserGroupRates.mockResolvedValue({})
    getPublicSettings.mockResolvedValue({})
  })

  it('asks for a budget multiplier when changing a fixed key to a dynamic group', async () => {
    const wrapper = mountKeysView()
    await flushPromises()

    await wrapper.find('button[title="keys.clickToChangeGroup"]').trigger('click')
    await flushPromises()
    await Array.from(document.body.querySelectorAll('button'))
      .find((button) => button.textContent?.includes('Dynamic Pool'))!
      .click()
    await flushPromises()

    const budgetInput = document.body.querySelector<HTMLInputElement>(
      '[data-testid="dynamic-budget-input"]'
    )
    expect(budgetInput).not.toBeNull()
    expect(budgetInput?.value).toBe('12')
    expect(updateKey).not.toHaveBeenCalled()

    budgetInput!.value = '15'
    budgetInput!.dispatchEvent(new Event('input'))
    await flushPromises()
    document.body
      .querySelector<HTMLButtonElement>('[data-testid="dynamic-budget-confirm"]')!
      .click()
    await flushPromises()

    expect(updateKey).toHaveBeenCalledWith(99, {
      group_id: 20,
      budget_multiplier: 15,
    })
  })

  it('defaults the edit form budget to the target dynamic group default when switching from a fixed group', async () => {
    const wrapper = mountKeysView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('common.edit'))!.trigger('click')
    await flushPromises()
    await wrapper.find('form#key-form select').setValue('20')
    await flushPromises()

    const budgetInput = wrapper.get<HTMLInputElement>('[data-testid="key-form-budget-input"]')
    expect(budgetInput.element.value).toBe('12')

    await budgetInput.setValue('14')
    await wrapper.find('form#key-form').trigger('submit.prevent')
    await flushPromises()

    expect(updateKey).toHaveBeenCalledWith(99, expect.objectContaining({
      group_id: 20,
      budget_multiplier: 14,
    }))
  })
})
