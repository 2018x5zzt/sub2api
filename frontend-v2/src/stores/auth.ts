import { create } from 'zustand'
import { authAPI, isTotp2FARequired, type LoginResponse } from '@/api/auth'
import type { User, LoginRequest, RegisterRequest, AuthResponse, PublicSettings } from '@/types'

const AUTH_TOKEN_KEY = 'auth_token'
const AUTH_USER_KEY = 'auth_user'
const REFRESH_TOKEN_KEY = 'refresh_token'

interface AuthState {
  user: User | null
  token: string | null
  publicSettings: PublicSettings | null
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
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: null,
  publicSettings: null,
  initialized: false,
  runMode: 'standard',

  isAuthenticated: () => !!get().token && !!get().user,
  isAdmin: () => get().user?.role === 'admin',

  init: async () => {
    const savedToken = localStorage.getItem(AUTH_TOKEN_KEY)
    const savedUser = localStorage.getItem(AUTH_USER_KEY)
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
        set({ token: null, user: null })
      }
      return null
    }
  }
}))
