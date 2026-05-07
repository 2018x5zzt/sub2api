import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Mail, Lock, Eye, EyeOff } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { useAuthStore } from '@/stores/auth'
import { isTotp2FARequired } from '@/api/auth'
import { toast } from '@/components/ui/Toast'

export default function LoginPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const redirect = params.get('redirect') || '/dashboard'

  const login = useAuthStore((s) => s.login)
  const publicSettings = useAuthStore((s) => s.publicSettings)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [errors, setErrors] = useState<{ email?: string; password?: string; form?: string }>({})

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
        // TODO Phase 2: 2FA flow page
        setErrors({ form: 'Two-factor authentication required (not yet migrated to v2)' })
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

  return (
    <AuthLayout title={t('auth.welcomeBack') as string} subtitle={t('auth.signInToAccount') as string}>
      <form onSubmit={onSubmit} className="space-y-5">
        {errors.form && (
          <div className="rounded-lg border border-signal-err/20 bg-signal-err/5 px-3 py-2 text-sm text-signal-err">
            {errors.form}
          </div>
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
