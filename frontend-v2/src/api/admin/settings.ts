import { apiClient } from '@/api/client'
import type { CustomMenuItem } from '@/types'

/** Default subscription rule applied at user registration. */
export interface DefaultSubscriptionSetting {
  group_id: number
  validity_days?: number | null
}

/** Subset of SystemSettings the Phase 2D admin view exposes. */
export interface SystemSettings {
  // Registration
  registration_enabled: boolean
  email_verify_enabled: boolean
  registration_email_suffix_whitelist: string[]
  promo_code_enabled: boolean
  password_reset_enabled: boolean
  frontend_url: string
  invitation_code_enabled: boolean
  totp_enabled: boolean
  totp_encryption_key_configured: boolean

  // Defaults
  default_balance: number
  default_concurrency: number
  default_subscriptions: DefaultSubscriptionSetting[]

  // OEM
  site_name: string
  site_logo: string
  site_subtitle: string
  api_base_url: string
  contact_info: string
  doc_url: string
  home_content: string
  hide_ccs_import_button: boolean
  purchase_subscription_enabled: boolean
  purchase_subscription_url: string
  sora_client_enabled: boolean
  backend_mode_enabled: boolean
  custom_menu_items: CustomMenuItem[]

  // SMTP
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password_configured: boolean
  smtp_from_email: string
  smtp_from_name: string
  smtp_use_tls: boolean

  // Turnstile
  turnstile_enabled: boolean
  turnstile_site_key: string
  turnstile_secret_key_configured: boolean

  // LinuxDo OAuth
  linuxdo_connect_enabled: boolean
  linuxdo_connect_client_id: string
  linuxdo_connect_client_secret_configured: boolean
  linuxdo_connect_redirect_url: string

  // Model fallback
  enable_model_fallback: boolean
  fallback_model_anthropic: string
  fallback_model_openai: string
  fallback_model_gemini: string
  fallback_model_antigravity: string

  // Identity patch
  enable_identity_patch: boolean
  identity_patch_prompt: string

  // Ops Monitoring
  ops_monitoring_enabled: boolean
  ops_realtime_monitoring_enabled: boolean

  // Claude Code version pinning
  min_claude_code_version: string
  max_claude_code_version: string

  // Group isolation
  allow_ungrouped_key_scheduling: boolean
}

/** Fields that may be sent to PUT /admin/settings. Secret fields take a string
 *  to set/replace; backend echoes only `_configured` booleans on read. */
export interface UpdateSettingsRequest extends Partial<Omit<SystemSettings, 'smtp_password_configured' | 'turnstile_secret_key_configured' | 'linuxdo_connect_client_secret_configured' | 'totp_encryption_key_configured'>> {
  smtp_password?: string
  turnstile_secret_key?: string
  linuxdo_connect_client_secret?: string
}

export async function getSettings(): Promise<SystemSettings> {
  const { data } = await apiClient.get<SystemSettings>('/admin/settings')
  return data
}

export async function updateSettings(payload: UpdateSettingsRequest): Promise<SystemSettings> {
  const { data } = await apiClient.put<SystemSettings>('/admin/settings', payload)
  return data
}

export interface TestSmtpRequest {
  host: string
  port: number
  username?: string
  password?: string
  use_tls: boolean
}

export async function testSmtpConnection(config: TestSmtpRequest) {
  const { data } = await apiClient.post<{ message: string }>('/admin/settings/test-smtp', config)
  return data
}

export interface SendTestEmailRequest extends TestSmtpRequest {
  from_email: string
  from_name?: string
  to_email: string
}

export async function sendTestEmail(req: SendTestEmailRequest) {
  const { data } = await apiClient.post<{ message: string }>('/admin/settings/send-test-email', req)
  return data
}

export const adminSettingsAPI = {
  getSettings,
  updateSettings,
  testSmtpConnection,
  sendTestEmail
}
