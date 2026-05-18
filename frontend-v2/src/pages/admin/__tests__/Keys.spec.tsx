import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AdminKeysPage from '../Keys'

const { listUsers, listUserApiKeys, updateApiKeyGroup, listAllGroups } = vi.hoisted(() => ({
  listUsers: vi.fn(),
  listUserApiKeys: vi.fn(),
  updateApiKeyGroup: vi.fn(),
  listAllGroups: vi.fn()
}))

const toastSuccess = vi.fn()
const toastError = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: { listUsers }
}))

vi.mock('@/api/admin/keys', () => ({
  adminKeysAPI: { listUserApiKeys, updateApiKeyGroup }
}))

vi.mock('@/api/admin/groups', () => ({
  adminGroupsAPI: { listAllGroups }
}))

vi.mock('@/components/ui/Toast', () => ({
  toast: {
    success: (...args: any[]) => toastSuccess(...args),
    error: (...args: any[]) => toastError(...args),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/components/layout/ConsoleLayout', () => ({
  PageHeader: ({ title, description }: { title: any; description?: any }) => (
    <div>
      <h1>{title}</h1>
      {description && <p>{description}</p>}
    </div>
  )
}))

vi.mock('@/components/ui/Card', () => ({
  Card: ({ children }: { children: any }) => <div>{children}</div>
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, onClick, ...rest }: any) => <button onClick={onClick} {...rest}>{children}</button>
}))

vi.mock('@/components/ui/Input', () => ({
  Input: ({ label, ...rest }: any) => (
    <label>
      {label}
      <input {...rest} />
    </label>
  )
}))

vi.mock('@/components/ui/Modal', () => ({
  Modal: ({ open, title, children, footer }: any) =>
    open ? (
      <div role="dialog">
        <h2>{title}</h2>
        {children}
        {footer}
      </div>
    ) : null
}))

vi.mock('@/components/ui/Badge', () => ({
  Badge: ({ children }: { children: any }) => <span>{children}</span>
}))

vi.mock('@/components/ui/Table', () => ({
  Table: ({ children }: { children: any }) => <table>{children}</table>,
  THead: ({ children }: { children: any }) => <thead>{children}</thead>,
  TBody: ({ children }: { children: any }) => <tbody>{children}</tbody>,
  TR: ({ children }: { children: any }) => <tr>{children}</tr>,
  TH: ({ children }: { children: any }) => <th>{children}</th>,
  TD: ({ children }: { children: any }) => <td>{children}</td>
}))

vi.mock('@/components/ui/Skeleton', () => ({
  Skeleton: () => <div />
}))

const users = [
  {
    id: 1,
    email: 'alice@example.com',
    username: 'alice',
    role: 'user',
    status: 'active',
    balance: 3.5,
    concurrency: 1,
    allowed_groups: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    notes: '',
    sora_storage_quota_bytes: 0,
    sora_storage_used_bytes: 0
  }
]

const groups = [
  {
    id: 10,
    name: 'Group A',
    description: '',
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    default_budget_multiplier: null,
    pricing_mode: 'fixed',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    balance_fallback_group_id: null,
    allow_messages_dispatch: false,
    require_oauth_only: false,
    require_privacy_set: false,
    rpm_limit: 0,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  },
  {
    id: 20,
    name: 'Group B',
    description: '',
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: true,
    status: 'active',
    subscription_type: 'standard',
    default_budget_multiplier: null,
    pricing_mode: 'fixed',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    balance_fallback_group_id: null,
    allow_messages_dispatch: false,
    require_oauth_only: false,
    require_privacy_set: false,
    rpm_limit: 0,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  }
]

const apiKeys = [
  {
    id: 99,
    user_id: 1,
    key: 'sk-test-admin-key-xxxx',
    name: 'Key 99',
    group_id: 10,
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    group: groups[0],
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
    reset_7d_at: null
  }
]

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false }
    }
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <AdminKeysPage />
    </QueryClientProvider>
  )
}

describe('AdminKeysPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listUsers.mockResolvedValue({ items: users, total: 1, page: 1, page_size: 20, pages: 1 })
    listAllGroups.mockResolvedValue(groups)
    listUserApiKeys.mockResolvedValue({ items: apiKeys, total: 1, page: 1, page_size: 100, pages: 1 })
    updateApiKeyGroup.mockResolvedValue({
      api_key: { ...apiKeys[0], group_id: 20, group: groups[1] },
      auto_granted_group_access: true,
      granted_group_id: 20,
      granted_group_name: 'Group B'
    })
  })

  it('loads selected user keys and updates key group through admin api', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('alice@example.com')

    await user.click(screen.getByRole('button', { name: 'admin.users.apiKeys' }))

    const dialog = await screen.findByRole('dialog')
    expect(within(dialog).getByText('Key 99')).toBeTruthy()

    await within(dialog).findByRole('option', { name: 'Group B' })

    const groupSelect = within(dialog).getByLabelText('admin.users.group') as HTMLSelectElement
    fireEvent.change(groupSelect, { target: { value: '20' } })

    await waitFor(() => {
      expect(updateApiKeyGroup).toHaveBeenCalledWith(99, { group_id: 20 })
    })
    await waitFor(() => {
      expect(toastSuccess).toHaveBeenCalled()
    })
  })
})
