import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsErrorDetailModal from '../OpsErrorDetailModal.vue'
import OpsErrorLogTable from '../OpsErrorLogTable.vue'

const mockGetRequestErrorDetail = vi.fn()
const mockGetUpstreamErrorDetail = vi.fn()
const mockListRequestErrorUpstreamErrors = vi.fn()
const mockShowError = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getRequestErrorDetail: (...args: any[]) => mockGetRequestErrorDetail(...args),
    getUpstreamErrorDetail: (...args: any[]) => mockGetUpstreamErrorDetail(...args),
    listRequestErrorUpstreamErrors: (...args: any[]) => mockListRequestErrorUpstreamErrors(...args),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: mockShowError,
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, any>) => {
        if (key === 'admin.ops.errorDetail.titleWithId' && params?.id) {
          return `detail-${params.id}`
        }
        return key
      },
    }),
  }
})

const TooltipStub = defineComponent({
  name: 'ElTooltipStub',
  template: '<div><slot /></div>',
})

const PaginationStub = defineComponent({
  name: 'Pagination',
  template: '<div class="pagination-stub" />',
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' },
  },
  emits: ['close'],
  template: '<div class="base-dialog"><slot /></div>',
})

const IconStub = defineComponent({
  name: 'Icon',
  template: '<span class="icon-stub" />',
})

function makeErrorDetail(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    created_at: '2026-04-13T00:00:00Z',
    phase: 'request',
    type: 'request',
    error_owner: 'client',
    error_source: 'client_request',
    severity: 'error',
    status_code: 400,
    platform: 'openai',
    model: 'fallback-model',
    requested_model: 'gpt-4.1',
    upstream_model: 'gpt-4.1-mini',
    is_retryable: false,
    retry_count: 0,
    resolved: false,
    client_request_id: 'client-1',
    request_id: 'req-1',
    message: 'bad request',
    user_id: 7,
    user_email: 'user@example.com',
    account_id: 8,
    account_name: 'acct',
    group_id: 9,
    group_name: 'grp',
    error_body: '{"error":"bad request"}',
    user_agent: 'vitest',
    request_body: '{}',
    request_body_truncated: false,
    is_business_limited: false,
    request_type: 1,
    inbound_endpoint: '/v1/chat/completions',
    upstream_endpoint: 'https://api.openai.com/v1/chat/completions',
    ...overrides,
  }
}

describe('Ops error displays', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListRequestErrorUpstreamErrors.mockResolvedValue({ items: [] })
    mockGetUpstreamErrorDetail.mockResolvedValue(makeErrorDetail())
    mockGetRequestErrorDetail.mockResolvedValue(makeErrorDetail())
  })

  it('shows requested to upstream model mapping in the error log table', () => {
    const wrapper = mount(OpsErrorLogTable, {
      props: {
        rows: [makeErrorDetail()],
        total: 1,
        loading: false,
        page: 1,
        pageSize: 10,
      },
      global: {
        stubs: {
          Pagination: PaginationStub,
          'el-tooltip': TooltipStub,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('gpt-4.1')
    expect(text).toContain('gpt-4.1-mini')
  })

  it('shows upstream model details in the error detail modal summary', async () => {
    const wrapper = mount(OpsErrorDetailModal, {
      props: {
        show: true,
        errorId: 1,
        errorType: 'request',
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: IconStub,
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('gpt-4.1')
    expect(text).toContain('gpt-4.1-mini')
  })
})
