import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Mail } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { authAPI } from '@/api/auth'

export default function ForgotPasswordPage() {
  const { t } = useTranslation()
  const [email, setEmail] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [done, setDone] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await authAPI.forgotPassword({ email })
      setDone(true)
    } catch (err) {
      setError((err as { message?: string })?.message || (t('auth.sendResetLinkFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout title={t('auth.forgotPasswordTitle') as string} subtitle={t('auth.forgotPasswordHint') as string}>
      {done ? (
        <div className="space-y-4 text-center">
          <h3 className="font-medium text-ink-1">{t('auth.resetEmailSent')}</h3>
          <p className="text-sm text-ink-2">{t('auth.resetEmailSentHint')}</p>
          <Link to="/login" className="btn btn-ghost">
            {t('auth.backToLogin')}
          </Link>
        </div>
      ) : (
        <form onSubmit={onSubmit} className="space-y-5">
          {error && (
            <div className="rounded-lg border border-signal-err/20 bg-signal-err/5 px-3 py-2 text-sm text-signal-err">
              {error}
            </div>
          )}
          <Input
            name="email"
            type="email"
            label={t('auth.emailLabel') as string}
            placeholder={t('auth.emailPlaceholder') as string}
            autoFocus
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            leftIcon={<Mail className="h-4 w-4" />}
          />
          <Button type="submit" loading={submitting} className="w-full" size="lg">
            {t('auth.sendResetLink')}
          </Button>
          <div className="text-center text-sm">
            <Link to="/login" className="link">
              {t('auth.backToLogin')}
            </Link>
          </div>
        </form>
      )}
    </AuthLayout>
  )
}
