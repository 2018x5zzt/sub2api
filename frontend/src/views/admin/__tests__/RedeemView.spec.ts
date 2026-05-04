import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import RedeemView from '../RedeemView.vue'

const { redeemList, listProducts } = vi.hoisted(() => ({
  redeemList: vi.fn(),
  listProducts: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    redeem: {
      list: redeemList,
      generate: vi.fn(),
      exportCodes: vi.fn(),
      delete: vi.fn(),
      batchDelete: vi.fn(),
    },
    subscriptionProducts: {
      list: listProducts,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
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
      t: (key: string) => key,
    }),
  }
})

describe('admin RedeemView', () => {
  it('uses selected product default validity days when generating product subscription codes', async () => {
    redeemList.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
    listProducts.mockResolvedValue([
      {
        id: 88,
        code: 'gpt_weekly_45',
        name: 'GPT weekly',
        description: '',
        status: 'active',
        product_family: 'gpt_shared',
        default_validity_days: 7,
        daily_limit_usd: 45,
        weekly_limit_usd: 315,
        monthly_limit_usd: 0,
        sort_order: 1,
        created_at: '2026-05-01T00:00:00Z',
        updated_at: '2026-05-01T00:00:00Z',
      },
    ])

    const wrapper = mount(RedeemView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
          },
          DataTable: true,
          Pagination: true,
          ConfirmDialog: true,
          Select: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    const vm = wrapper.vm as unknown as {
      generateForm: { type: string; product_id: number | null; validity_days: number }
    }
    vm.generateForm.type = 'subscription'
    await nextTick()
    vm.generateForm.product_id = 88
    await nextTick()

    expect(vm.generateForm.validity_days).toBe(7)
  })
})
