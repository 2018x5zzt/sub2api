// @vitest-environment jsdom

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import EmailVerifyPage from '../EmailVerify'

const registerMock = vi.fn()
const sendVerifyCodeMock = vi.fn()
const navigateMock = vi.fn()

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/components/layout/AuthLayout', () => ({
  AuthLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
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

vi.mock('@/api/auth', () => ({
  authAPI: {
    sendVerifyCode: (...args: unknown[]) => sendVerifyCodeMock(...args)
  }
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: (selector: (state: {
    register: typeof registerMock
    publicSettings: { site_name: string } | null
  }) => unknown) => selector({
    register: registerMock,
    publicSettings: { site_name: 'Xlabapi' }
  })
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => navigateMock
  }
})

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/email-verify']}>
      <Routes>
        <Route path="/email-verify" element={<EmailVerifyPage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('EmailVerifyPage affiliate code forwarding', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    sessionStorage.clear()
    localStorage.clear()
    sendVerifyCodeMock.mockResolvedValue({ countdown: 60 })
    registerMock.mockResolvedValue({ id: 1 })
    sessionStorage.setItem('register_data', JSON.stringify({
      email: 'verify-user@example.com',
      password: 'password123',
      turnstile_token: 'turnstile-token',
      invitation_code: 'INVITE123',
      aff_code: 'AFF_FROM_REGISTER_CACHE'
    }))
  })

  it('submits verify-code registration with cached aff_code', async () => {
    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(sendVerifyCodeMock).toHaveBeenCalledTimes(1)
    })

    const codeInput = await screen.findByLabelText('auth.verificationCode')
    await user.type(codeInput, '123456')
    await user.click(screen.getByRole('button', { name: 'auth.verifyAndCreate' }))

    await waitFor(() => {
      expect(registerMock).toHaveBeenCalledTimes(1)
    })
    expect(registerMock).toHaveBeenCalledWith(expect.objectContaining({
      email: 'verify-user@example.com',
      password: 'password123',
      verify_code: '123456',
      invitation_code: 'INVITE123',
      aff_code: 'AFF_FROM_REGISTER_CACHE'
    }))
  })
})
