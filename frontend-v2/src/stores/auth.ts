import { create } from 'zustand'
import { authAPI, isTotp2FARequired, type LoginResponse } from '@/api/auth'
import type { User, LoginRequest, RegisterRequest, AuthResponse, PublicSettings } from '@/types'

const AUTH_TOKEN_KEY = 'auth_token'
const AUTH_USER_KEY = 'auth_user'
const REFRESH_TOKEN_KEY = 'refresh_token'
const TOKEN_EXPIRES_AT_KEY = 'token_expires_at'
const PENDING_AUTH_SESSION_KEY = 'pending_auth_session'

type PendingAuthTokenField = 'pending_auth_token' | 'pending_oauth_token'

export interface PendingAuthSessionSummary {
  token: string
  token_field: PendingAuthTokenField
  provider: string
  redirect?: string
  adoption_required?: boolean
  suggested_display_name?: string
  suggested_avatar_url?: string
}

function normalizePendingAuthTokenField(value: unknown): PendingAuthTokenField {
  return value === 'pending_oauth_token' ? 'pending_oauth_token' : 'pending_auth_token'
}

function getPersistedPendingAuthSession(): PendingAuthSessionSummary | null {
  const raw = localStorage.getItem(PENDING_AUTH_SESSION_KEY)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as Partial<PendingAuthSessionSummary> | null
    const provider = typeof parsed?.provider === 'string' ? parsed.provider.trim() : ''
    if (!provider) {
      localStorage.removeItem(PENDING_AUTH_SESSION_KEY)
      return null
    }
    return {
      token: typeof parsed?.token === 'string' ? parsed.token : '',
      token_field: normalizePendingAuthTokenField(parsed?.token_field),
      provider,
      redirect: typeof parsed?.redirect === 'string' ? parsed.redirect : undefined,
      adoption_required: typeof parsed?.adoption_required === 'boolean' ? parsed.adoption_required : undefined,
      suggested_display_name: typeof parsed?.suggested_display_name === 'string' ? parsed.suggested_display_name : undefined,
      suggested_avatar_url: typeof parsed?.suggested_avatar_url === 'string' ? parsed.suggested_avatar_url : undefined
    }
  } catch {
    localStorage.removeItem(PENDING_AUTH_SESSION_KEY)
    return null
  }
}

interface AuthState {
  user: User | null
  token: string | null
  publicSettings: PublicSettings | null
  pendingAuthSession: PendingAuthSessionSummary | null
  initialized: boolean
  runMode: 'standard' | 'simple'
  isAuthenticated: () => boolean
  isAdmin: () => boolean
  init: () => Promise<void>
  loadPublicSettings: () => Promise<void>
  login: (req: LoginRequest) => Promise<LoginResponse>
  register: (req: RegisterRequest) => Promise<User>
  logout: () => Promise<void>
  refreshUser: () => Promise<User | null>
  setAuthFromResponse: (resp: AuthResponse) => void
  /** Set auth state from an externally-issued token (e.g. OAuth callback). */
  setToken: (accessToken: string) => Promise<User | null>
  setPendingAuthSession: (session: PendingAuthSessionSummary | null) => void
  clearPendingAuthSession: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: null,
  publicSettings: null,
  pendingAuthSession: null,
  initialized: false,
  runMode: 'standard',

  isAuthenticated: () => !!get().token && !!get().user,
  isAdmin: () => get().user?.role === 'admin',

  init: async () => {
    const savedToken = localStorage.getItem(AUTH_TOKEN_KEY)
    const savedUser = localStorage.getItem(AUTH_USER_KEY)
    set({ pendingAuthSession: getPersistedPendingAuthSession() })
    if (savedToken && savedUser) {
      try {
        set({ token: savedToken, user: JSON.parse(savedUser) })
        get().refreshUser().catch(() => {})
      } catch {
        get().logout()
      }
    }
    set({ initialized: true })
  },

  loadPublicSettings: async () => {
    try {
      const ps = await authAPI.getPublicSettings()
      set({ publicSettings: ps })
    } catch (e) {
      console.warn('Failed to load public settings', e)
    }
  },

  setAuthFromResponse: (resp) => {
    if (resp.refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, resp.refresh_token)
    const { run_mode, ...userData } = resp.user
    localStorage.setItem(AUTH_TOKEN_KEY, resp.access_token)
    localStorage.setItem(AUTH_USER_KEY, JSON.stringify(userData))
    set({ token: resp.access_token, user: userData, runMode: run_mode || 'standard' })
  },

  login: async (req) => {
    const resp = await authAPI.login(req)
    if (!isTotp2FARequired(resp)) {
      get().setAuthFromResponse(resp)
    }
    return resp
  },

  register: async (req) => {
    const resp = await authAPI.register(req)
    get().setAuthFromResponse(resp)
    return resp.user
  },

  logout: async () => {
    try {
      await authAPI.logout()
    } catch {
      // ignore
    }
    authAPI.clearAuthToken()
    localStorage.removeItem(TOKEN_EXPIRES_AT_KEY)
    set({ token: null, user: null })
  },

  refreshUser: async () => {
    if (!get().token) return null
    try {
      const u = await authAPI.getCurrentUser()
      const { run_mode, ...userData } = u
      localStorage.setItem(AUTH_USER_KEY, JSON.stringify(userData))
      set({ user: userData, runMode: run_mode || 'standard' })
      return userData
    } catch (e) {
      const status = (e as { status?: number }).status
      if (status === 401) {
        authAPI.clearAuthToken()
        localStorage.removeItem(TOKEN_EXPIRES_AT_KEY)
        set({ token: null, user: null })
      }
      return null
    }
  },

  setToken: async (accessToken) => {
    localStorage.setItem(AUTH_TOKEN_KEY, accessToken)
    set({ token: accessToken })
    return get().refreshUser()
  },

  setPendingAuthSession: (session) => {
    if (session) {
      localStorage.setItem(PENDING_AUTH_SESSION_KEY, JSON.stringify(session))
      set({ pendingAuthSession: session })
      return
    }
    localStorage.removeItem(PENDING_AUTH_SESSION_KEY)
    set({ pendingAuthSession: null })
  },

  clearPendingAuthSession: () => {
    localStorage.removeItem(PENDING_AUTH_SESSION_KEY)
    set({ pendingAuthSession: null })
  }
}))
