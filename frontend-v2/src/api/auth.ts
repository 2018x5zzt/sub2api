import { apiClient } from './client'
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  CurrentUserResponse,
  SendVerifyCodeRequest,
  SendVerifyCodeResponse,
  PublicSettings,
  TotpLoginResponse,
  TotpLogin2FARequest
} from '@/types'

export type LoginResponse = AuthResponse | TotpLoginResponse

export function isTotp2FARequired(r: LoginResponse): r is TotpLoginResponse {
  return 'requires_2fa' in r && r.requires_2fa === true
}

export function setAuthToken(t: string) { localStorage.setItem('auth_token', t) }
export function setRefreshToken(t: string) { localStorage.setItem('refresh_token', t) }
export function setTokenExpiresAt(s: number) { localStorage.setItem('token_expires_at', String(Date.now() + s * 1000)) }
export function getAuthToken() { return localStorage.getItem('auth_token') }
export function getRefreshToken() { return localStorage.getItem('refresh_token') }
export function clearAuthToken() {
  localStorage.removeItem('auth_token')
  localStorage.removeItem('refresh_token')
  localStorage.removeItem('auth_user')
  localStorage.removeItem('token_expires_at')
}

export async function login(credentials: LoginRequest): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>('/auth/login', credentials)
  if (!isTotp2FARequired(data)) {
    setAuthToken(data.access_token)
    if (data.refresh_token) setRefreshToken(data.refresh_token)
    if (data.expires_in) setTokenExpiresAt(data.expires_in)
    localStorage.setItem('auth_user', JSON.stringify(data.user))
  }
  return data
}

export async function login2FA(req: TotpLogin2FARequest): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/login/2fa', req)
  setAuthToken(data.access_token)
  if (data.refresh_token) setRefreshToken(data.refresh_token)
  if (data.expires_in) setTokenExpiresAt(data.expires_in)
  localStorage.setItem('auth_user', JSON.stringify(data.user))
  return data
}

export async function register(userData: RegisterRequest): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/register', userData)
  setAuthToken(data.access_token)
  if (data.refresh_token) setRefreshToken(data.refresh_token)
  if (data.expires_in) setTokenExpiresAt(data.expires_in)
  localStorage.setItem('auth_user', JSON.stringify(data.user))
  return data
}

export async function getCurrentUser() {
  const { data } = await apiClient.get<CurrentUserResponse>('/auth/me')
  return data
}

export async function logout(): Promise<void> {
  const rt = getRefreshToken()
  if (rt) {
    try {
      await apiClient.post('/auth/logout', { refresh_token: rt })
    } catch {
      // ignore
    }
  }
  clearAuthToken()
}

export async function getPublicSettings(): Promise<PublicSettings> {
  const { data } = await apiClient.get<PublicSettings>('/settings/public')
  return data
}

export async function sendVerifyCode(req: SendVerifyCodeRequest): Promise<SendVerifyCodeResponse> {
  const { data } = await apiClient.post<SendVerifyCodeResponse>('/auth/send-verify-code', req)
  return data
}

export interface ForgotPasswordRequest {
  email: string
  turnstile_token?: string
}

export async function forgotPassword(req: ForgotPasswordRequest): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/auth/forgot-password', req)
  return data
}

export interface ResetPasswordRequest {
  email: string
  token: string
  new_password: string
}

export async function resetPassword(req: ResetPasswordRequest): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/auth/reset-password', req)
  return data
}

export const authAPI = {
  login,
  login2FA,
  isTotp2FARequired,
  register,
  getCurrentUser,
  logout,
  setAuthToken,
  setRefreshToken,
  setTokenExpiresAt,
  getAuthToken,
  getRefreshToken,
  clearAuthToken,
  getPublicSettings,
  sendVerifyCode,
  forgotPassword,
  resetPassword
}

export default authAPI
