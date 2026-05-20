import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import KeysPage from '../Keys'

const { listKeys, createKey, updateKey, deleteKey, getUserGroups, getUserGroupRates, getDashboardApiKeysUsage } = vi.hoisted(() => ({
  listKeys: vi.fn(),
  createKey: vi.fn(),
  updateKey: vi.fn(),
  deleteKey: vi.fn(),
  getUserGroups: vi.fn(),
  getUserGroupRates: vi.fn(),
  getDashboardApiKeysUsage: vi.fn()
}))

vi.mock('@/api/keys', () => ({
  keysAPI: { listKeys, createKey, updateKey, deleteKey }
}))

vi.mock('@/api/models', () => ({
  modelsAPI: { getUserGroups, getUserGroupRates }
}))

vi.mock('@/api/usage', () => ({
  usageAPI: { getDashboardApiKeysUsage }
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/components/layout/ConsoleLayout', () => ({
  PageHeader: ({ title, description, actions }: { title: any; description?: any; actions?: any }) => (
    <div>
      <h1>{title}</h1>
      {description && <p>{description}</p>}
      {actions}
    </div>
  )
}))

vi.mock('@/components/ui/Card', () => ({
  Card: ({ children }: { children: any }) => <div>{children}</div>
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, onClick, loading, ...rest }: any) => (
    <button onClick={onClick} {...rest}>{loading ? 'Loading' : children}</button>
  )
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
  Modal: ({ open, title, children, footer, onClose }: any) =>
    open ? (
      <div role="dialog">
        <h2>{title}</h2>
        <button onClick={onClose}>close</button>
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

vi.mock('@/components/ui/Toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

vi.mock('@/i18n', () => ({
  default: {
    t: (key: string, params?: Record<string, unknown>) => {
      if (!params) return key
      return Object.entries(params).reduce((message, [name, value]) => message.replace(`{${name}}`, String(value)), key)
    }
  }
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { email: 'user@example.com', balance: 0 },
    runMode: 'standard',
    publicSettings: null,
    isAdmin: () => false,
    logout: vi.fn()
  })
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => vi.fn()
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

const dynamicGroup = {
  ...fixedGroup,
  id: 20,
  name: 'Dynamic Pool',
  pricing_mode: 'dynamic',
  default_budget_multiplier: 12
}

const keyItem = {
  id: 99,
  user_id: 7,
  key: 'sk-test-key',
  name: 'fixed key',
  group_id: fixedGroup.id,
  status: 'active' as const,
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
  reset_7d_at: null
}

function selectOption(select: HTMLSelectElement, value: string) {
  fireEvent.change(select, { target: { value } })
}

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false }
    }
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <KeysPage />
    </QueryClientProvider>
  )
}

describe('KeysPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listKeys.mockResolvedValue({ items: [keyItem], total: 1, pages: 1 })
    getDashboardApiKeysUsage.mockResolvedValue({
      stats: {
        [keyItem.id]: {
          api_key_id: keyItem.id,
          today_actual_cost: 0,
          total_actual_cost: 0,
          today_requests: 0,
          today_tokens: 0,
          total_requests: 0,
          total_tokens: 0
        }
      }
    })
    createKey.mockResolvedValue({ key: 'sk-new' })
    updateKey.mockResolvedValue({})
    deleteKey.mockResolvedValue(undefined)
    getUserGroups.mockResolvedValue([fixedGroup, dynamicGroup])
    getUserGroupRates.mockResolvedValue({})
  })

  it('renders group selection and enforces budget when binding a dynamic group', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('fixed key')
    await user.click(screen.getByRole('button', { name: /keys.createKey/i }))

    const dialog = await screen.findByRole('dialog')
    const nameInput = within(dialog).getByLabelText('keys.nameLabel') as HTMLInputElement
    await user.type(nameInput, 'new key')

    const groupSelect = within(dialog).getByLabelText('keys.groupLabel') as HTMLSelectElement
    selectOption(groupSelect, String(dynamicGroup.id))

    const budgetInput = within(dialog).getByLabelText('keys.budgetMultiplierLabel') as HTMLInputElement
    expect(budgetInput.value).toBe('12')

    await user.clear(budgetInput)
    await user.click(within(dialog).getByRole('button', { name: 'common.create' }))
    expect(createKey).not.toHaveBeenCalled()

    await user.click(within(dialog).getByRole('button', { name: 'common.create' }))
    expect(createKey).not.toHaveBeenCalled()

    await user.type(budgetInput, '18')
    await user.click(within(dialog).getByRole('button', { name: 'common.create' }))
    expect(createKey).toHaveBeenCalledWith(expect.objectContaining({
      name: 'new key',
      group_id: 20,
      budget_multiplier: 18
    }))
  })

  it('loads and renders per-key usage stats', async () => {
    getDashboardApiKeysUsage.mockResolvedValueOnce({
      stats: {
        [keyItem.id]: {
          api_key_id: keyItem.id,
          today_actual_cost: 1.2345,
          total_actual_cost: 9.8765,
          today_requests: 3,
          today_tokens: 300,
          total_requests: 12,
          total_tokens: 4567
        }
      }
    })

    renderPage()

    await screen.findByText('fixed key')
    expect(getDashboardApiKeysUsage).toHaveBeenCalledWith([keyItem.id], expect.anything())
    expect(await screen.findByText('$1.2345')).toBeTruthy()
    expect(await screen.findByText('$9.8765')).toBeTruthy()
    expect(await screen.findByText(/300 Tok · 3 Req/)).toBeTruthy()
    expect(await screen.findByText(/4,567 Tok · 12 Req/)).toBeTruthy()
  })

  it('updates existing key group and dynamic budget in edit modal', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('fixed key')
    await user.click(screen.getByRole('button', { name: /keys.editKey/i }))

    const dialog = await screen.findByRole('dialog')
    const nameInput = within(dialog).getByLabelText('keys.nameLabel') as HTMLInputElement
    expect(nameInput.value).toBe('fixed key')

    const groupSelect = within(dialog).getByLabelText('keys.groupLabel') as HTMLSelectElement
    selectOption(groupSelect, String(dynamicGroup.id))

    const budgetInput = within(dialog).getByLabelText('keys.budgetMultiplierLabel') as HTMLInputElement
    await user.clear(budgetInput)
    await user.type(budgetInput, '14')

    await user.click(within(dialog).getByRole('button', { name: 'common.update' }))
    expect(updateKey).toHaveBeenCalledWith(99, expect.objectContaining({
      name: 'fixed key',
      group_id: 20,
      budget_multiplier: 14
    }))
  })

  it('submits full edit payload fields when quota/rate-limit/expiration are enabled', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('fixed key')
    await user.click(screen.getByRole('button', { name: /keys.editKey/i }))

    const dialog = await screen.findByRole('dialog')

    const enableRateLimit = within(dialog).getByLabelText('keys.rateLimitSection') as HTMLInputElement
    if (!enableRateLimit.checked) {
      await user.click(enableRateLimit)
    }
    await user.type(within(dialog).getByLabelText('keys.rateLimit5h'), '5')
    await user.type(within(dialog).getByLabelText('keys.rateLimit1d'), '10')
    await user.type(within(dialog).getByLabelText('keys.rateLimit7d'), '20')

    const quotaInput = within(dialog).getByLabelText('keys.quotaAmount') as HTMLInputElement
    await user.type(quotaInput, '30')

    const enableExpiration = within(dialog).getByLabelText('keys.expiration') as HTMLInputElement
    if (!enableExpiration.checked) {
      await user.click(enableExpiration)
    }
    const expirationInput = within(dialog).getByLabelText('keys.expirationDate') as HTMLInputElement
    await user.clear(expirationInput)
    await user.type(expirationInput, '2030-01-02T03:04')

    await user.click(within(dialog).getByRole('button', { name: 'common.update' }))

    expect(updateKey).toHaveBeenCalledTimes(1)
    const [id, payload] = updateKey.mock.calls[0] as [number, Record<string, unknown>]
    expect(id).toBe(99)
    expect(payload).toEqual(expect.objectContaining({
      name: 'fixed key',
      group_id: fixedGroup.id,
      quota: 30,
      rate_limit_5h: 5,
      rate_limit_1d: 10,
      rate_limit_7d: 20,
      status: 'active'
    }))
    expect(typeof payload.expires_at).toBe('string')
    expect(String(payload.expires_at)).toContain('2030-01-02T')
  })

  it('opens edit modal when ip restriction lists are null in API response', async () => {
    const user = userEvent.setup()
    listKeys.mockResolvedValueOnce({
      items: [
        {
          ...keyItem,
          id: 100,
          name: 'null-ip key',
          ip_whitelist: null,
          ip_blacklist: null
        }
      ],
      total: 1,
      pages: 1
    })

    renderPage()

    await screen.findByText('null-ip key')
    await user.click(screen.getByRole('button', { name: /keys.editKey/i }))

    const dialog = await screen.findByRole('dialog')
    expect(dialog).toBeTruthy()
    const nameInput = within(dialog).getByLabelText('keys.nameLabel') as HTMLInputElement
    expect(nameInput.value).toBe('null-ip key')
  })
})
