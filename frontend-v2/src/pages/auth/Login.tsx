import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Mail, Lock, Eye, EyeOff, ShieldCheck, ArrowLeft } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { useAuthStore } from '@/stores/auth'
import { authAPI, isTotp2FARequired } from '@/api/auth'
import { toast } from '@/components/ui/Toast'

interface TwoFAState {
  tempToken: string
  emailMasked?: string
}

export default function LoginPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const redirect = params.get('redirect') || '/dashboard'

  const login = useAuthStore((s) => s.login)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const linuxdoEnabled = publicSettings?.linuxdo_oauth_enabled
  const backendModeEnabled = publicSettings?.backend_mode_enabled

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [errors, setErrors] = useState<{ email?: string; password?: string; form?: string }>({})

  // 2FA state — when set, the form swaps to a 6-digit code prompt
  const [twoFA, setTwoFA] = useState<TwoFAState | null>(null)
  const [totpCode, setTotpCode] = useState('')
  const [totpError, setTotpError] = useState<string | null>(null)

  function validate(): boolean {
    const next: typeof errors = {}
    if (!email) next.email = t('auth.emailRequired') as string
    if (!password) next.password = t('auth.passwordRequired') as string
    setErrors(next)
    return Object.keys(next).length === 0
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!validate()) return
    setSubmitting(true)
    setErrors({})
    try {
      const resp = await login({ email, password })
      if (isTotp2FARequired(resp)) {
        if (!resp.temp_token) {
          setErrors({ form: t('auth.loginFailed') as string })
          return
        }
        setTwoFA({ tempToken: resp.temp_token, emailMasked: resp.user_email_masked })
        return
      }
      toast.success(t('auth.loginSuccess') as string)
      navigate(redirect, { replace: true })
    } catch (err) {
      const msg = (err as { message?: string })?.message || (t('auth.loginFailed') as string)
      setErrors({ form: msg })
    } finally {
      setSubmitting(false)
    }
  }

  async function on2FASubmit(e: FormEvent) {
    e.preventDefault()
    if (!twoFA) return
    setTotpError(null)
    const trimmed = totpCode.trim()
    if (!/^\d{6}$/.test(trimmed)) {
      setTotpError(t('auth.invalidCode') as string)
      return
    }
    setSubmitting(true)
    try {
      const resp = await authAPI.login2FA({ temp_token: twoFA.tempToken, totp_code: trimmed })
      // login2FA already persisted token to localStorage; sync the store
      useAuthStore.getState().setAuthFromResponse(resp)
      toast.success(t('auth.loginSuccess') as string)
      navigate(redirect, { replace: true })
    } catch (err) {
      setTotpError((err as { message?: string })?.message || (t('auth.verifyFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  function cancel2FA() {
    setTwoFA(null)
    setTotpCode('')
    setTotpError(null)
  }

  if (twoFA) {
    return (
      <AuthLayout
        title={t('auth.verifyYourEmail') as string}
        subtitle={
          twoFA.emailMasked
            ? `${t('auth.sendCodeDesc')} ${twoFA.emailMasked}`
            : (t('auth.signInToAccount') as string)
        }
      >
        <form onSubmit={on2FASubmit} className="space-y-5">
          <div className="rounded-xl border border-line-2 bg-bg-2 px-4 py-3 text-sm text-ink-2 flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-orange shrink-0" />
            Two-factor authentication required
          </div>

          <div>
            <label htmlFor="totp" className="input-label text-center block">
              {t('auth.verificationCode')}
            </label>
            <input
              id="totp"
              name="totp"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              autoFocus
              required
              maxLength={6}
              disabled={submitting}
              value={totpCode}
              onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ''))}
              placeholder="000000"
              className={`input py-3 text-center font-mono text-xl tracking-[0.5em] ${totpError ? 'input-error' : ''}`}
            />
            {totpError && <p className="input-error-text text-center">{totpError}</p>}
          </div>

          <Button type="submit" loading={submitting} disabled={!totpCode} className="w-full" size="lg">
            {t('auth.verifying') ? t('auth.continue') : t('auth.signIn')}
          </Button>

          <div className="text-center">
            <button
              type="button"
              onClick={cancel2FA}
              className="inline-flex items-center gap-2 text-sm text-ink-3 hover:text-ink-1"
              disabled={submitting}
            >
              <ArrowLeft className="h-4 w-4" />
              {t('common.back')}
            </button>
          </div>
        </form>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout title={t('auth.welcomeBack') as string} subtitle={t('auth.signInToAccount') as string}>
      <form onSubmit={onSubmit} className="space-y-5">
        {errors.form && (
          <div className="rounded-lg border border-signal-err/20 bg-signal-err/5 px-3 py-2 text-sm text-signal-err">
            {errors.form}
          </div>
        )}

        {linuxdoEnabled && !backendModeEnabled && (
          <>
            <a
              href="/api/v1/auth/oauth/linuxdo/login"
              className="btn btn-ghost w-full"
              aria-label="LinuxDo"
            >
              {t('auth.linuxdo.signIn')}
            </a>
            <div className="flex items-center gap-3 text-xs text-ink-3">
              <div className="flex-1 hr-fade" />
              <span>{t('auth.linuxdo.orContinue')}</span>
              <div className="flex-1 hr-fade" />
            </div>
          </>
        )}

        <Input
          name="email"
          type="email"
          label={t('auth.emailLabel') as string}
          placeholder={t('auth.emailPlaceholder') as string}
          autoFocus
          autoComplete="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          leftIcon={<Mail className="h-4 w-4" />}
          error={errors.email}
          disabled={submitting}
        />

        <Input
          name="password"
          type={showPassword ? 'text' : 'password'}
          label={t('auth.passwordLabel') as string}
          placeholder={t('auth.passwordPlaceholder') as string}
          autoComplete="current-password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          leftIcon={<Lock className="h-4 w-4" />}
          rightAdornment={
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              className="btn btn-ghost btn-icon"
            >
              {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          }
          error={errors.password}
          disabled={submitting}
        />

        {publicSettings?.password_reset_enabled && (
          <div className="text-right">
            <Link to="/forgot-password" className="link text-xs">
              {t('auth.forgotPassword')}
            </Link>
          </div>
        )}

        <Button type="submit" loading={submitting} className="w-full" size="lg">
          {t('auth.signIn')}
        </Button>

        {publicSettings?.registration_enabled && (
          <div className="text-center text-sm text-ink-2">
            {t('auth.dontHaveAccount')}{' '}
            <Link to="/register" className="link">
              {t('auth.createAccount')}
            </Link>
          </div>
        )}
      </form>
    </AuthLayout>
  )
}
