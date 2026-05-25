// @vitest-environment jsdom

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import RegisterPage from '../Register'

const registerMock = vi.fn()
const navigateMock = vi.fn()

type PublicSettingsShape = {
  site_name: string
  email_verify_enabled: boolean
  promo_code_enabled: boolean
  invitation_code_enabled: boolean
}

const currentSettings: PublicSettingsShape = {
  site_name: 'Xlabapi',
  email_verify_enabled: false,
  promo_code_enabled: false,
  invitation_code_enabled: false
}

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/components/layout/AuthLayout', () => ({
  AuthLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}))

vi.mock('@/components/ui/Input', () => ({
  Input: ({ label, ...rest }: { label: string } & Record<string, unknown>) => (
    <label>
      {label}
      <input {...rest} />
    </label>
  )
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, loading, ...rest }: { children: React.ReactNode; loading?: boolean } & Record<string, unknown>) => (
    <button {...rest}>{loading ? 'Loading' : children}</button>
  )
}))

vi.mock('@/components/ui/Toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: (selector: (state: {
    register: typeof registerMock
    publicSettings: PublicSettingsShape
  }) => unknown) => selector({
    register: registerMock,
    publicSettings: currentSettings
  })
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => navigateMock
  }
})

function renderPage(path = '/register?aff=AFF_FROM_QUERY') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/register" element={<RegisterPage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('RegisterPage affiliate code forwarding', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    sessionStorage.clear()
    localStorage.clear()
    currentSettings.site_name = 'Xlabapi'
    currentSettings.email_verify_enabled = false
    currentSettings.promo_code_enabled = false
    currentSettings.invitation_code_enabled = false
    registerMock.mockResolvedValue({ id: 1 })
  })

  it('sends aff_code from query on direct registration submit', async () => {
    const user = userEvent.setup()
    renderPage('/register?aff=AFF_FROM_QUERY')

    await user.type(screen.getByLabelText('auth.emailLabel'), 'new-user@example.com')
    await user.type(screen.getByLabelText('auth.passwordLabel'), 'password123')
    await user.click(screen.getByRole('button', { name: 'auth.createAccount' }))

    await waitFor(() => {
      expect(registerMock).toHaveBeenCalledTimes(1)
    })
    expect(registerMock).toHaveBeenCalledWith(expect.objectContaining({
      email: 'new-user@example.com',
      password: 'password123',
      aff_code: 'AFF_FROM_QUERY'
    }))
  })

  it('stores aff_code in register_data when email verification flow is enabled', async () => {
    const user = userEvent.setup()
    currentSettings.email_verify_enabled = true
    renderPage('/register?aff=AFF_VERIFY_FLOW')

    await user.type(screen.getByLabelText('auth.emailLabel'), 'verify-user@example.com')
    await user.type(screen.getByLabelText('auth.passwordLabel'), 'password123')
    await user.click(screen.getByRole('button', { name: 'auth.continue' }))

    const raw = sessionStorage.getItem('register_data')
    expect(raw).toBeTruthy()
    const payload = JSON.parse(raw || '{}') as Record<string, unknown>
    expect(payload.aff_code).toBe('AFF_VERIFY_FLOW')
    expect(registerMock).not.toHaveBeenCalled()
  })
})
