import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SubscriptionProductsView from '../SubscriptionProductsView.vue'

const {
  listProducts,
  listUserSubscriptions,
  getAllGroups,
} = vi.hoisted(() => ({
  listProducts: vi.fn(),
  listUserSubscriptions: vi.fn(),
  getAllGroups: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    subscriptionProducts: {
      list: listProducts,
      listUserSubscriptions,
      create: vi.fn(),
      update: vi.fn(),
      listBindings: vi.fn(),
      syncBindings: vi.fn(),
      listSubscriptions: vi.fn(),
      assign: vi.fn(),
    },
    groups: {
      getAll: getAllGroups,
    },
    usage: {
      searchUsers: vi.fn(),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, fallback?: string) => fallback ?? key,
    }),
  }
})

const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: `
    <div>
      <div data-test="columns">{{ columns.map(col => col.key).join(',') }}</div>
      <div data-test="row-count">{{ data.length }}</div>
      <div v-for="row in data" :key="row.id" data-test="row">
        <slot name="cell-user" :row="row" :value="row.user_email" />
        <slot name="cell-product" :row="row" :value="row.product_name" />
        <slot name="cell-daily_usage" :row="row" />
        <slot name="cell-carryover" :row="row" />
        <slot name="cell-fresh_daily_usage" :row="row" :value="row.fresh_daily_usage_usd" />
      </div>
    </div>
  `,
}

describe('admin SubscriptionProductsView', () => {
  beforeEach(() => {
    localStorage.clear()
    listProducts.mockReset()
    listUserSubscriptions.mockReset()
    getAllGroups.mockReset()

    listProducts.mockResolvedValue([
      {
        id: 1,
        code: 'gpt_daily_45',
        name: 'GPT 订阅每天45刀',
        description: '',
        status: 'active',
        default_validity_days: 30,
        daily_limit_usd: 45,
        weekly_limit_usd: 0,
        monthly_limit_usd: 0,
        sort_order: 1,
        created_at: '2026-05-01T00:00:00Z',
        updated_at: '2026-05-01T00:00:00Z',
      },
    ])
    listUserSubscriptions.mockResolvedValue({
      items: [
        {
          id: 66,
          user_id: 1907,
          user_email: 'user@example.com',
          user_username: 'user1907',
          product_id: 1,
          product_code: 'gpt_daily_45',
          product_name: 'GPT 订阅每天45刀',
          daily_limit_usd: 45,
          daily_usage_usd: 12.5,
          weekly_usage_usd: 25,
          monthly_usage_usd: 50,
          daily_carryover_in_usd: 8,
          daily_carryover_remaining_usd: 3,
          carryover_used_usd: 5,
          fresh_daily_usage_usd: 7.5,
          starts_at: '2026-05-01T00:00:00Z',
          expires_at: '2026-06-01T00:00:00Z',
          status: 'active',
          notes: 'admin note',
          daily_window_start: '2026-05-02T00:00:00Z',
          weekly_window_start: null,
          monthly_window_start: null,
          assigned_by: null,
          assigned_at: '2026-05-01T00:00:00Z',
          created_at: '2026-05-01T00:00:00Z',
          updated_at: '2026-05-02T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    getAllGroups.mockResolvedValue([])
  })

  it('defaults to the per-user product subscription usage table', async () => {
    const wrapper = mount(SubscriptionProductsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
          },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: true,
          EmptyState: true,
          Select: true,
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(listUserSubscriptions).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
    })
    expect(listProducts).toHaveBeenCalled()

    const columns = wrapper.get('[data-test="columns"]').text().split(',')
    expect(columns).toEqual([
      'user',
      'product',
      'status',
      'daily_usage',
      'carryover',
      'fresh_daily_usage',
      'period',
      'notes',
    ])
    expect(wrapper.get('[data-test="row-count"]').text()).toBe('1')
    expect(wrapper.text()).toContain('user@example.com')
    expect(wrapper.text()).toContain('GPT 订阅每天45刀')
    expect(wrapper.text()).toContain('$12.50')
    expect(wrapper.text()).toContain('$8.00')
    expect(wrapper.text()).toContain('$7.50')
  })
})
