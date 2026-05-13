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

export type WeChatOAuthMode = 'open' | 'mp'
export type WeChatOAuthUnavailableReason =
  | 'not_configured'
  | 'external_browser_required'
  | 'wechat_browser_required'
  | 'native_app_required'

export interface ResolvedWeChatOAuthStart {
  mode: WeChatOAuthMode | null
  unavailableReason: WeChatOAuthUnavailableReason | null
}

export function isWeChatWebOAuthEnabled(settings: PublicSettings | null | undefined): boolean {
  const legacyEnabled = settings?.wechat_oauth_enabled === true
  const hasExplicitCapabilities =
    typeof settings?.wechat_oauth_open_enabled === 'boolean' ||
    typeof settings?.wechat_oauth_mp_enabled === 'boolean'
  if (!hasExplicitCapabilities) return legacyEnabled
  return settings?.wechat_oauth_open_enabled === true || settings?.wechat_oauth_mp_enabled === true
}

export function resolveWeChatOAuthStart(
  settings: PublicSettings | null | undefined,
  userAgent?: string
): ResolvedWeChatOAuthStart {
  const ua = (userAgent ?? (typeof navigator !== 'undefined' ? navigator.userAgent : '') ?? '').trim()
  const isWeChatBrowser = /MicroMessenger/i.test(ua)
  const legacyEnabled = settings?.wechat_oauth_enabled === true
  const openEnabled = typeof settings?.wechat_oauth_open_enabled === 'boolean'
    ? settings.wechat_oauth_open_enabled
    : legacyEnabled
  const mpEnabled = typeof settings?.wechat_oauth_mp_enabled === 'boolean'
    ? settings.wechat_oauth_mp_enabled
    : legacyEnabled
  const mobileEnabled = settings?.wechat_oauth_mobile_enabled === true

  if (isWeChatBrowser) {
    if (mpEnabled) return { mode: 'mp', unavailableReason: null }
    if (openEnabled) return { mode: null, unavailableReason: 'external_browser_required' }
    return { mode: null, unavailableReason: mobileEnabled ? 'native_app_required' : 'not_configured' }
  }
  if (openEnabled) return { mode: 'open', unavailableReason: null }
  if (mpEnabled) return { mode: null, unavailableReason: 'wechat_browser_required' }
  return { mode: null, unavailableReason: mobileEnabled ? 'native_app_required' : 'not_configured' }
}

export async function sendVerifyCode(req: SendVerifyCodeRequest): Promise<SendVerifyCodeResponse> {
  const { data } = await apiClient.post<SendVerifyCodeResponse>('/auth/send-verify-code', req)
  return data
}

export interface SendPendingOAuthVerifyCodeRequest extends SendVerifyCodeRequest {
  pending_auth_token?: string
  pending_oauth_token?: string
}

export async function sendPendingOAuthVerifyCode(
  req: SendPendingOAuthVerifyCodeRequest
): Promise<SendVerifyCodeResponse> {
  const { data } = await apiClient.post<SendVerifyCodeResponse>('/auth/oauth/pending/send-verify-code', req)
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

export interface OAuthTokenPayload {
  access_token: string
  refresh_token?: string
  expires_in?: number
  token_type?: string
}

export interface PendingOAuthExchangeResponse extends Partial<OAuthTokenPayload> {
  auth_result?: string
  redirect?: string
  error?: string
  provider?: string
  requires_2fa?: boolean
  temp_token?: string
  user_email_masked?: string
}

export interface OAuthAdoptionDecision {
  adoptDisplayName?: boolean
  adoptAvatar?: boolean
}

function serializeOAuthAdoptionDecision(decision?: OAuthAdoptionDecision): Record<string, boolean> {
  const payload: Record<string, boolean> = {}
  if (typeof decision?.adoptDisplayName === 'boolean') payload.adopt_display_name = decision.adoptDisplayName
  if (typeof decision?.adoptAvatar === 'boolean') payload.adopt_avatar = decision.adoptAvatar
  return payload
}

export async function completeLinuxDoOAuthRegistration(
  invitationCode: string,
  decision?: OAuthAdoptionDecision,
  affiliateCode?: string
): Promise<PendingOAuthExchangeResponse> {
  return completeOAuthRegistration('linuxdo', invitationCode, decision, affiliateCode)
}

export async function completeOAuthRegistration(
  provider: 'linuxdo' | 'oidc' | 'wechat',
  invitationCode: string,
  decision?: OAuthAdoptionDecision,
  affiliateCode?: string
): Promise<PendingOAuthExchangeResponse> {
  const { data } = await apiClient.post<PendingOAuthExchangeResponse>(
    `/auth/oauth/${provider}/complete-registration`,
    {
      invitation_code: invitationCode,
      ...(affiliateCode?.trim() ? { aff_code: affiliateCode.trim() } : {}),
      ...serializeOAuthAdoptionDecision(decision)
    }
  )
  return data
}

export async function exchangePendingOAuthCompletion(
  decision?: OAuthAdoptionDecision
): Promise<PendingOAuthExchangeResponse> {
  const { data } = await apiClient.post<PendingOAuthExchangeResponse>(
    '/auth/oauth/pending/exchange',
    serializeOAuthAdoptionDecision(decision)
  )
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
  isWeChatWebOAuthEnabled,
  resolveWeChatOAuthStart,
  sendVerifyCode,
  sendPendingOAuthVerifyCode,
  forgotPassword,
  resetPassword,
  completeLinuxDoOAuthRegistration,
  completeOAuthRegistration,
  exchangePendingOAuthCompletion
}

export default authAPI
