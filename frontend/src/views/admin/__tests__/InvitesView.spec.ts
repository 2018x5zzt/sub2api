import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import InvitesView from '@/views/admin/InvitesView.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  const messages: Record<string, string> = {
    'admin.invites.riskPanelTitle': 'High-risk invite operations',
    'admin.invites.riskPanelBody': 'Historical rewards are not rewritten automatically',
    'admin.invites.preview': 'Preview recompute',
    'admin.invites.executeRecompute': 'Execute recompute'
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key
    })
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

vi.mock('@/api/admin/invites', () => ({
  default: {
    getStats: vi.fn().mockResolvedValue({
      total_invited_users: 3,
      qualified_reward_users_total: 2,
      base_rewards_total: 15,
      manual_grants_total: 5,
      recompute_adjustments_total: -2
    }),
    listRelationships: vi.fn().mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    }),
    listRewards: vi.fn().mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    }),
    listActions: vi.fn().mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    }),
    previewRecompute: vi.fn().mockResolvedValue({
      scope_hash: 'hash',
      qualifying_event_count: 1,
      deltas: [
        {
          inviter_user_id: 11,
          invitee_user_id: 7,
          reward_target_user_id: 7,
          reward_role: 'invitee',
          current_amount: 0,
          expected_amount: 5,
          delta_amount: 5
        }
      ]
    }),
    executeRecompute: vi.fn().mockResolvedValue({ message: 'ok' }),
    rebindInviter: vi.fn().mockResolvedValue({ message: 'ok' }),
    createManualGrant: vi.fn().mockResolvedValue({ message: 'ok' })
  }
}))

const TablePageLayoutStub = {
  template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
}

describe('InvitesView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders the always-visible warning panel', async () => {
    const wrapper = mount(InvitesView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: TablePageLayoutStub,
          DataTable: true,
          Pagination: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Historical rewards are not rewritten automatically')
  })

  it('requires recompute preview before execute', async () => {
    const wrapper = mount(InvitesView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: TablePageLayoutStub,
          DataTable: true,
          Pagination: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.find('[data-test="execute-recompute"]').attributes('disabled')).toBeDefined()
  })
})
