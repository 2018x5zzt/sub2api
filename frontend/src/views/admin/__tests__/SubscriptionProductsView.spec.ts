import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SubscriptionProductsView from '../SubscriptionProductsView.vue'

const {
  listProducts,
  createProduct,
  syncBindings,
  getAllGroups,
  showError,
  showSuccess
} = vi.hoisted(() => ({
  listProducts: vi.fn(),
  createProduct: vi.fn(),
  syncBindings: vi.fn(),
  getAllGroups: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    subscriptionProducts: {
      listProducts,
      createProduct,
      syncBindings
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const fakeProduct = {
  id: 101,
  code: 'gpt_monthly',
  name: 'GPT 月卡',
  description: 'GPT shared monthly subscription',
  status: 'draft',
  default_validity_days: 30,
  daily_limit_usd: 10,
  weekly_limit_usd: 50,
  monthly_limit_usd: 100,
  sort_order: 10,
  created_at: '2026-04-25T00:00:00Z',
  updated_at: '2026-04-25T00:00:00Z'
}

describe('SubscriptionProductsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listProducts.mockResolvedValue([fakeProduct])
    createProduct.mockResolvedValue({ ...fakeProduct, id: 102, code: 'claude_monthly' })
    syncBindings.mockResolvedValue([
      {
        product_id: 101,
        group_id: 88,
        group_name: 'plus/team',
        debit_multiplier: 1.5,
        status: 'active',
        sort_order: 10,
        created_at: '2026-04-25T00:00:00Z',
        updated_at: '2026-04-25T00:00:00Z'
      }
    ])
    getAllGroups.mockResolvedValue([
      {
        id: 88,
        name: 'plus/team',
        description: 'OpenAI plus team',
        platform: 'openai',
        subscription_type: 'subscription',
        status: 'active',
        rate_multiplier: 1,
        sort_order: 10
      }
    ])
    showSuccess.mockResolvedValue(undefined)
  })

  it('creates and lists subscription products', async () => {
    const wrapper = mount(SubscriptionProductsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          Select: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('GPT 月卡')

    const setupState = (wrapper.vm as any).$?.setupState
    Object.assign(setupState.productForm, {
      code: 'claude_monthly',
      name: 'Claude 月卡',
      description: 'Claude shared monthly subscription',
      status: 'draft',
      default_validity_days: 30,
      monthly_limit_usd: 100,
      sort_order: 20
    })

    await setupState.createProduct()

    expect(createProduct).toHaveBeenCalledWith(
      expect.objectContaining({
        code: 'claude_monthly',
        name: 'Claude 月卡',
        monthly_limit_usd: 100
      })
    )
    expect(showSuccess).toHaveBeenCalled()
  })

  it('edits product-group multipliers', async () => {
    const wrapper = mount(SubscriptionProductsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          Select: true
        }
      }
    })

    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.selectedProduct = fakeProduct
    setupState.bindingDrafts = [
      {
        group_id: 88,
        debit_multiplier: 1.5,
        status: 'active',
        sort_order: 10
      }
    ]

    await setupState.saveBindings()

    expect(syncBindings).toHaveBeenCalledWith(101, [
      {
        group_id: 88,
        debit_multiplier: 1.5,
        status: 'active',
        sort_order: 10
      }
    ])
  })
})
