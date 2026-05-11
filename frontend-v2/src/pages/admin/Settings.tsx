import { useEffect, useState, type ReactNode, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Save, Loader2, Mail, MailCheck, Send } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { adminSettingsAPI, type SystemSettings, type UpdateSettingsRequest } from '@/api/admin/settings'
import { toast } from '@/components/ui/Toast'
import { cn } from '@/lib/cn'

type Tab = 'general' | 'registration' | 'smtp' | 'turnstile' | 'oauth' | 'fallback'

interface TabDef {
  key: Tab
  label: string
  description?: string
}

const TABS: TabDef[] = [
  { key: 'general', label: 'General / OEM' },
  { key: 'registration', label: 'Registration' },
  { key: 'smtp', label: 'SMTP' },
  { key: 'turnstile', label: 'Turnstile' },
  { key: 'oauth', label: 'OAuth' },
  { key: 'fallback', label: 'Fallback / Routing' }
]

interface FieldRowProps {
  label: ReactNode
  hint?: ReactNode
  children: ReactNode
}

function FieldRow({ label, hint, children }: FieldRowProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-[260px_1fr] gap-4 md:gap-8 items-start py-4 border-b border-line-1 last:border-b-0">
      <div>
        <div className="text-sm text-ink-1 font-medium">{label}</div>
        {hint && <div className="text-xs text-ink-3 mt-1 leading-relaxed">{hint}</div>}
      </div>
      <div>{children}</div>
    </div>
  )
}

function Toggle({
  checked,
  onChange,
  disabled
}: {
  checked: boolean
  onChange: (v: boolean) => void
  disabled?: boolean
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={cn(
        'relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors',
        'border border-line-2 disabled:opacity-50 disabled:cursor-not-allowed',
        checked ? 'bg-orange' : 'bg-bg-3'
      )}
    >
      <span
        className={cn(
          'inline-block h-4 w-4 rounded-full bg-white transition-transform',
          checked ? 'translate-x-[22px]' : 'translate-x-1'
        )}
      />
    </button>
  )
}

function ConfiguredPill({ on }: { on: boolean }) {
  return on ? (
    <span className="inline-flex items-center gap-1.5 text-xs text-signal-ok font-mono">
      <span className="w-1.5 h-1.5 rounded-full bg-signal-ok" />
      configured
    </span>
  ) : (
    <span className="inline-flex items-center gap-1.5 text-xs text-ink-3 font-mono">
      <span className="w-1.5 h-1.5 rounded-full bg-ink-4" />
      not set
    </span>
  )
}

function SecretInput({
  configured,
  value,
  onChange,
  placeholder,
  disabled
}: {
  configured: boolean
  value: string
  onChange: (v: string) => void
  placeholder?: string
  disabled?: boolean
}) {
  const [show, setShow] = useState(false)
  return (
    <div className="space-y-1.5">
      <div className="flex items-center gap-3">
        <Input
          name="secret"
          type={show ? 'text' : 'password'}
          placeholder={placeholder ?? (configured ? 'Leave blank to keep existing' : 'Enter secret')}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className="font-mono"
          rightAdornment={
            <button
              type="button"
              onClick={() => setShow((v) => !v)}
              className="btn btn-ghost btn-icon btn-sm"
              tabIndex={-1}
            >
              {show ? 'hide' : 'show'}
            </button>
          }
        />
      </div>
      <ConfiguredPill on={configured} />
    </div>
  )
}

interface State {
  draft: UpdateSettingsRequest
  smtp_password: string
  turnstile_secret_key: string
  linuxdo_connect_client_secret: string
}

function emptyState(): State {
  return {
    draft: {},
    smtp_password: '',
    turnstile_secret_key: '',
    linuxdo_connect_client_secret: ''
  }
}

export default function AdminSettingsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [tab, setTab] = useState<Tab>('general')
  const [state, setState] = useState<State>(emptyState)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-settings'],
    queryFn: () => adminSettingsAPI.getSettings()
  })

  // Reset draft whenever fresh server data arrives
  useEffect(() => {
    setState(emptyState())
  }, [data])

  // Effective view = server values overlaid with the in-flight draft
  const view: SystemSettings | null = data ? { ...data, ...state.draft } : null

  function set<K extends keyof SystemSettings>(key: K, value: SystemSettings[K]) {
    setState((s) => ({ ...s, draft: { ...s.draft, [key]: value } }))
  }

  const saveMut = useMutation({
    mutationFn: (payload: UpdateSettingsRequest) => adminSettingsAPI.updateSettings(payload),
    onSuccess: (fresh) => {
      qc.setQueryData(['admin-settings'], fresh)
      toast.success(t('common.success') as string)
      setState(emptyState())
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const testSmtp = useMutation({
    mutationFn: () => {
      if (!view) throw new Error('Not loaded')
      return adminSettingsAPI.testSmtpConnection({
        host: view.smtp_host,
        port: view.smtp_port,
        username: view.smtp_username || undefined,
        password: state.smtp_password || undefined,
        use_tls: view.smtp_use_tls
      })
    },
    onSuccess: (r) => toast.success(r.message || 'Connected'),
    onError: (e: { message?: string }) => toast.error(e?.message || 'SMTP test failed')
  })

  function onSave(e: FormEvent) {
    e.preventDefault()
    const payload: UpdateSettingsRequest = { ...state.draft }
    if (state.smtp_password) payload.smtp_password = state.smtp_password
    if (state.turnstile_secret_key) payload.turnstile_secret_key = state.turnstile_secret_key
    if (state.linuxdo_connect_client_secret)
      payload.linuxdo_connect_client_secret = state.linuxdo_connect_client_secret
    saveMut.mutate(payload)
  }

  if (isLoading || !view) {
    return (
      <>
        <PageHeader title="Settings" description="System configuration" />
        <Card className="p-6">
          <Skeleton className="h-4 w-48 mb-4" />
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-10 mb-2" />
          ))}
        </Card>
      </>
    )
  }

  const dirty = Object.keys(state.draft).length > 0 || state.smtp_password || state.turnstile_secret_key || state.linuxdo_connect_client_secret

  return (
    <form onSubmit={onSave}>
      <PageHeader
        title="Settings"
        description="System configuration. Sora S3, beta policies, and ops monitoring keys are deferred to Phase 3."
        actions={
          <Button
            type="submit"
            variant="accent"
            disabled={!dirty}
            loading={saveMut.isPending}
          >
            <Save className="h-3.5 w-3.5" />
            {t('common.save')}
          </Button>
        }
      />

      <div className="flex flex-wrap gap-1 mb-4 border-b border-line-2 -mb-px">
        {TABS.map((tabDef) => (
          <button
            key={tabDef.key}
            type="button"
            onClick={() => setTab(tabDef.key)}
            className={cn(
              'px-4 py-2.5 text-sm transition-colors -mb-px border-b-2',
              tab === tabDef.key
                ? 'text-orange border-orange'
                : 'text-ink-3 border-transparent hover:text-ink-2'
            )}
          >
            {tabDef.label}
          </button>
        ))}
      </div>

      <Card className="p-6">
        {tab === 'general' && (
          <div>
            <FieldRow label="Site name">
              <Input
                name="site_name"
                value={view.site_name}
                onChange={(e) => set('site_name', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Site subtitle">
              <Input
                name="site_subtitle"
                value={view.site_subtitle}
                onChange={(e) => set('site_subtitle', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Logo URL" hint="Public URL or absolute path">
              <Input
                name="site_logo"
                value={view.site_logo}
                onChange={(e) => set('site_logo', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="API base URL" hint="Shown to users on the keys page">
              <Input
                name="api_base_url"
                value={view.api_base_url}
                onChange={(e) => set('api_base_url', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Frontend URL" hint="Used to build links in outgoing emails">
              <Input
                name="frontend_url"
                value={view.frontend_url}
                onChange={(e) => set('frontend_url', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Doc URL">
              <Input
                name="doc_url"
                value={view.doc_url}
                onChange={(e) => set('doc_url', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Contact info" hint="Shown on the redeem page when codes look bad">
              <Input
                name="contact_info"
                value={view.contact_info}
                onChange={(e) => set('contact_info', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Hide CCS import" hint="Hide the &quot;Import to CCS&quot; button on the keys page">
              <Toggle
                checked={view.hide_ccs_import_button}
                onChange={(v) => set('hide_ccs_import_button', v)}
              />
            </FieldRow>
            <FieldRow label="Purchase subscription" hint="Surface a buy/recharge entry in the user nav">
              <div className="space-y-3">
                <Toggle
                  checked={view.purchase_subscription_enabled}
                  onChange={(v) => set('purchase_subscription_enabled', v)}
                />
                <Input
                  name="purchase_subscription_url"
                  placeholder="https://shop.example.com"
                  value={view.purchase_subscription_url}
                  onChange={(e) => set('purchase_subscription_url', e.target.value)}
                  disabled={!view.purchase_subscription_enabled}
                />
              </div>
            </FieldRow>
            <FieldRow label="Backend mode" hint="Hides marketing/upsell surfaces — for internal-only deployments">
              <Toggle
                checked={view.backend_mode_enabled}
                onChange={(v) => set('backend_mode_enabled', v)}
              />
            </FieldRow>
          </div>
        )}

        {tab === 'registration' && (
          <div>
            <FieldRow label="Allow registration">
              <Toggle
                checked={view.registration_enabled}
                onChange={(v) => set('registration_enabled', v)}
              />
            </FieldRow>
            <FieldRow label="Email verification">
              <Toggle
                checked={view.email_verify_enabled}
                onChange={(v) => set('email_verify_enabled', v)}
              />
            </FieldRow>
            <FieldRow label="Password reset">
              <Toggle
                checked={view.password_reset_enabled}
                onChange={(v) => set('password_reset_enabled', v)}
              />
            </FieldRow>
            <FieldRow label="Promo codes">
              <Toggle
                checked={view.promo_code_enabled}
                onChange={(v) => set('promo_code_enabled', v)}
              />
            </FieldRow>
            <FieldRow label="Invitation codes">
              <Toggle
                checked={view.invitation_code_enabled}
                onChange={(v) => set('invitation_code_enabled', v)}
              />
            </FieldRow>
            <FieldRow label="TOTP (2FA)" hint="Encryption key must be configured at startup">
              <div className="space-y-2">
                <Toggle
                  checked={view.totp_enabled}
                  onChange={(v) => set('totp_enabled', v)}
                  disabled={!view.totp_encryption_key_configured}
                />
                <ConfiguredPill on={view.totp_encryption_key_configured} />
              </div>
            </FieldRow>
            <FieldRow
              label="Email suffix whitelist"
              hint="One per line. Empty = allow all. Example: example.com"
            >
              <textarea
                rows={4}
                value={(view.registration_email_suffix_whitelist || []).join('\n')}
                onChange={(e) =>
                  set(
                    'registration_email_suffix_whitelist',
                    e.target.value
                      .split('\n')
                      .map((s) => s.trim())
                      .filter(Boolean)
                  )
                }
                className="input font-mono text-xs"
                style={{ height: 'auto' }}
              />
            </FieldRow>
            <FieldRow label="Default balance ($)" hint="Granted on registration">
              <Input
                name="default_balance"
                type="number"
                step="0.01"
                min="0"
                value={String(view.default_balance ?? 0)}
                onChange={(e) => set('default_balance', Number(e.target.value) || 0)}
              />
            </FieldRow>
            <FieldRow label="Default concurrency">
              <Input
                name="default_concurrency"
                type="number"
                min="0"
                value={String(view.default_concurrency ?? 0)}
                onChange={(e) => set('default_concurrency', Number(e.target.value) || 0)}
              />
            </FieldRow>
          </div>
        )}

        {tab === 'smtp' && (
          <div>
            <FieldRow label="SMTP host">
              <Input
                name="smtp_host"
                value={view.smtp_host}
                onChange={(e) => set('smtp_host', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="SMTP port">
              <Input
                name="smtp_port"
                type="number"
                value={String(view.smtp_port ?? 0)}
                onChange={(e) => set('smtp_port', Number(e.target.value) || 0)}
              />
            </FieldRow>
            <FieldRow label="Use TLS">
              <Toggle checked={view.smtp_use_tls} onChange={(v) => set('smtp_use_tls', v)} />
            </FieldRow>
            <FieldRow label="Username">
              <Input
                name="smtp_username"
                value={view.smtp_username}
                onChange={(e) => set('smtp_username', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Password">
              <SecretInput
                configured={view.smtp_password_configured}
                value={state.smtp_password}
                onChange={(v) => setState((s) => ({ ...s, smtp_password: v }))}
              />
            </FieldRow>
            <FieldRow label="From address">
              <Input
                name="smtp_from_email"
                type="email"
                leftIcon={<Mail className="h-4 w-4" />}
                value={view.smtp_from_email}
                onChange={(e) => set('smtp_from_email', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="From name">
              <Input
                name="smtp_from_name"
                value={view.smtp_from_name}
                onChange={(e) => set('smtp_from_name', e.target.value)}
              />
            </FieldRow>
            <FieldRow label="Test connection" hint="Tries the host:port combo with the credentials above">
              <Button
                type="button"
                variant="ghost"
                onClick={() => testSmtp.mutate()}
                disabled={!view.smtp_host || testSmtp.isPending}
              >
                {testSmtp.isPending ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <MailCheck className="h-3.5 w-3.5" />
                )}
                Test
              </Button>
            </FieldRow>
          </div>
        )}

        {tab === 'turnstile' && (
          <div>
            <FieldRow label="Enable Turnstile">
              <Toggle
                checked={view.turnstile_enabled}
                onChange={(v) => set('turnstile_enabled', v)}
              />
            </FieldRow>
            <FieldRow label="Site key">
              <Input
                name="turnstile_site_key"
                value={view.turnstile_site_key}
                onChange={(e) => set('turnstile_site_key', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
            <FieldRow label="Secret key">
              <SecretInput
                configured={view.turnstile_secret_key_configured}
                value={state.turnstile_secret_key}
                onChange={(v) => setState((s) => ({ ...s, turnstile_secret_key: v }))}
              />
            </FieldRow>
          </div>
        )}

        {tab === 'oauth' && (
          <div>
            <FieldRow label="Enable LinuxDo Connect">
              <Toggle
                checked={view.linuxdo_connect_enabled}
                onChange={(v) => set('linuxdo_connect_enabled', v)}
              />
            </FieldRow>
            <FieldRow label="Client ID">
              <Input
                name="linuxdo_connect_client_id"
                value={view.linuxdo_connect_client_id}
                onChange={(e) => set('linuxdo_connect_client_id', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
            <FieldRow label="Client secret">
              <SecretInput
                configured={view.linuxdo_connect_client_secret_configured}
                value={state.linuxdo_connect_client_secret}
                onChange={(v) =>
                  setState((s) => ({ ...s, linuxdo_connect_client_secret: v }))
                }
              />
            </FieldRow>
            <FieldRow
              label="Redirect URL"
              hint="Must match the callback configured at the OAuth provider"
            >
              <Input
                name="linuxdo_connect_redirect_url"
                value={view.linuxdo_connect_redirect_url}
                onChange={(e) => set('linuxdo_connect_redirect_url', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
          </div>
        )}

        {tab === 'fallback' && (
          <div>
            <FieldRow label="Enable model fallback" hint="Reroute requests to a default model when the requested one isn't available">
              <Toggle
                checked={view.enable_model_fallback}
                onChange={(v) => set('enable_model_fallback', v)}
              />
            </FieldRow>
            <FieldRow label="Anthropic fallback">
              <Input
                name="fallback_model_anthropic"
                placeholder="claude-3-5-haiku-latest"
                value={view.fallback_model_anthropic}
                onChange={(e) => set('fallback_model_anthropic', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
            <FieldRow label="OpenAI fallback">
              <Input
                name="fallback_model_openai"
                placeholder="gpt-4o-mini"
                value={view.fallback_model_openai}
                onChange={(e) => set('fallback_model_openai', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
            <FieldRow label="Gemini fallback">
              <Input
                name="fallback_model_gemini"
                placeholder="gemini-2.0-flash"
                value={view.fallback_model_gemini}
                onChange={(e) => set('fallback_model_gemini', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
            <FieldRow label="Antigravity fallback">
              <Input
                name="fallback_model_antigravity"
                value={view.fallback_model_antigravity}
                onChange={(e) => set('fallback_model_antigravity', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
            <FieldRow label="Allow ungrouped key scheduling" hint="Let API keys without a group ID schedule against any non-exclusive group">
              <Toggle
                checked={view.allow_ungrouped_key_scheduling}
                onChange={(v) => set('allow_ungrouped_key_scheduling', v)}
              />
            </FieldRow>
            <FieldRow label="Min Claude Code version" hint="Reject requests from older clients">
              <Input
                name="min_claude_code_version"
                placeholder="0.0.0"
                value={view.min_claude_code_version}
                onChange={(e) => set('min_claude_code_version', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
            <FieldRow label="Max Claude Code version" hint="Reject requests from newer clients">
              <Input
                name="max_claude_code_version"
                placeholder="leave blank for no upper bound"
                value={view.max_claude_code_version}
                onChange={(e) => set('max_claude_code_version', e.target.value)}
                className="font-mono"
              />
            </FieldRow>
          </div>
        )}
      </Card>

      {dirty && (
        <div className="fixed bottom-4 right-4 z-30 card shadow-elevated p-3 px-4 flex items-center gap-3">
          <span className="text-sm text-ink-2">You have unsaved changes</span>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setState(emptyState())}
          >
            {t('common.reset')}
          </Button>
          <Button type="submit" variant="accent" size="sm" loading={saveMut.isPending}>
            <Send className="h-3 w-3" />
            {t('common.save')}
          </Button>
        </div>
      )}
    </form>
  )
}
