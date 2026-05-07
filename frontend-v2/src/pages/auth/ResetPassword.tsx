import { useMemo, useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Mail, Lock, Eye, EyeOff, AlertCircle, CheckCircle2 } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { authAPI } from '@/api/auth'
import { toast } from '@/components/ui/Toast'

export default function ResetPasswordPage() {
  const { t } = useTranslation()
  const [params] = useSearchParams()
  const email = params.get('email') || ''
  const token = params.get('token') || ''
  const invalidLink = !email || !token

  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [success, setSuccess] = useState(false)
  const [errors, setErrors] = useState<{ password?: string; confirmPassword?: string; form?: string }>({})

  const validate = useMemo(
    () => () => {
      const next: typeof errors = {}
      if (!password) next.password = t('auth.passwordRequired') as string
      else if (password.length < 6) next.password = t('auth.passwordMinLength') as string
      if (!confirmPassword) next.confirmPassword = t('auth.confirmPasswordRequired') as string
      else if (password !== confirmPassword) next.confirmPassword = t('auth.passwordsDoNotMatch') as string
      setErrors(next)
      return Object.keys(next).length === 0
    },
    [password, confirmPassword, t]
  )

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!validate()) return
    setSubmitting(true)
    setErrors({})
    try {
      await authAPI.resetPassword({ email, token, new_password: password })
      setSuccess(true)
      toast.success(t('auth.passwordResetSuccess') as string)
    } catch (err) {
      const e = err as { code?: string; message?: string }
      const msg =
        e.code === 'INVALID_RESET_TOKEN'
          ? (t('auth.invalidOrExpiredToken') as string)
          : e.message || (t('auth.resetPasswordFailed') as string)
      setErrors({ form: msg })
    } finally {
      setSubmitting(false)
    }
  }

  if (invalidLink) {
    return (
      <AuthLayout title={t('auth.invalidResetLink') as string}>
        <div className="space-y-5 text-center">
          <div className="rounded-xl border border-signal-err/30 bg-signal-err/5 p-5">
            <div className="flex flex-col items-center gap-3">
              <AlertCircle className="h-10 w-10 text-signal-err" />
              <p className="text-sm text-ink-2">{t('auth.invalidResetLinkHint')}</p>
            </div>
          </div>
          <Link to="/forgot-password" className="link inline-flex items-center gap-2 text-sm">
            {t('auth.requestNewResetLink')}
          </Link>
        </div>
      </AuthLayout>
    )
  }

  if (success) {
    return (
      <AuthLayout title={t('auth.passwordResetSuccess') as string}>
        <div className="space-y-5 text-center">
          <div className="rounded-xl border border-signal-ok/30 bg-signal-ok/5 p-5">
            <div className="flex flex-col items-center gap-3">
              <CheckCircle2 className="h-10 w-10 text-signal-ok" />
              <p className="text-sm text-ink-2">{t('auth.passwordResetSuccessHint')}</p>
            </div>
          </div>
          <Link to="/login" className="btn btn-primary inline-flex">
            {t('auth.signIn')}
          </Link>
        </div>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout
      title={t('auth.resetPasswordTitle') as string}
      subtitle={t('auth.resetPasswordHint') as string}
    >
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
          value={email}
          readOnly
          disabled
          leftIcon={<Mail className="h-4 w-4" />}
        />

        <Input
          name="password"
          type={showPassword ? 'text' : 'password'}
          label={t('auth.newPassword') as string}
          placeholder={t('auth.newPasswordPlaceholder') as string}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="new-password"
          required
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

        <Input
          name="confirmPassword"
          type={showConfirm ? 'text' : 'password'}
          label={t('auth.confirmPassword') as string}
          placeholder={t('auth.confirmPasswordPlaceholder') as string}
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          autoComplete="new-password"
          required
          leftIcon={<Lock className="h-4 w-4" />}
          rightAdornment={
            <button
              type="button"
              onClick={() => setShowConfirm((v) => !v)}
              className="btn btn-ghost btn-icon"
            >
              {showConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          }
          error={errors.confirmPassword}
          disabled={submitting}
        />

        <Button type="submit" loading={submitting} className="w-full" size="lg">
          {submitting ? t('auth.resettingPassword') : t('auth.resetPassword')}
        </Button>

        <div className="text-center text-sm text-ink-2">
          {t('auth.rememberedPassword')}{' '}
          <Link to="/login" className="link">
            {t('auth.signIn')}
          </Link>
        </div>
      </form>
    </AuthLayout>
  )
}
