import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Mail, Lock, Gift } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/components/ui/Toast'

const REGISTER_DATA_KEY = 'register_data'

export default function RegisterPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const register = useAuthStore((s) => s.register)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const siteName = publicSettings?.site_name || 'Sub2API'

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [promoCode, setPromoCode] = useState('')
  const [invitationCode, setInvitationCode] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const emailVerifyEnabled = publicSettings?.email_verify_enabled
  const promoEnabled = publicSettings?.promo_code_enabled
  const invitationEnabled = publicSettings?.invitation_code_enabled

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    // When email verification is required, stash the form payload and route to
    // the verification page (matches the legacy two-page flow).
    if (emailVerifyEnabled) {
      sessionStorage.setItem(
        REGISTER_DATA_KEY,
        JSON.stringify({
          email,
          password,
          promo_code: promoEnabled && promoCode ? promoCode : undefined,
          invitation_code: invitationEnabled && invitationCode ? invitationCode : undefined
        })
      )
      navigate('/email-verify')
      return
    }

    setSubmitting(true)
    try {
      await register({
        email,
        password,
        promo_code: promoEnabled && promoCode ? promoCode : undefined,
        invitation_code: invitationEnabled && invitationCode ? invitationCode : undefined
      })
      toast.success(t('auth.accountCreatedSuccess', { siteName }) as string)
      navigate('/dashboard', { replace: true })
    } catch (err) {
      setError((err as { message?: string })?.message || (t('auth.registrationFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout
      title={t('auth.createAccount') as string}
      subtitle={t('auth.signUpToStart', { siteName }) as string}
    >
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
          autoComplete="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          leftIcon={<Mail className="h-4 w-4" />}
          disabled={submitting}
        />

        <Input
          name="password"
          type="password"
          label={t('auth.passwordLabel') as string}
          placeholder={t('auth.createPasswordPlaceholder') as string}
          hint={t('auth.passwordHint') as string}
          autoComplete="new-password"
          required
          minLength={6}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          leftIcon={<Lock className="h-4 w-4" />}
          disabled={submitting}
        />

        {promoEnabled && (
          <Input
            name="promo_code"
            label={t('auth.promoCodeLabel') as string}
            placeholder={t('auth.promoCodePlaceholder') as string}
            value={promoCode}
            onChange={(e) => setPromoCode(e.target.value)}
            leftIcon={<Gift className="h-4 w-4" />}
            disabled={submitting}
          />
        )}

        {invitationEnabled && (
          <Input
            name="invitation_code"
            label={t('auth.invitationCodeLabel') as string}
            placeholder={t('auth.invitationCodePlaceholder') as string}
            value={invitationCode}
            onChange={(e) => setInvitationCode(e.target.value)}
            disabled={submitting}
          />
        )}

        <Button type="submit" loading={submitting} className="w-full" size="lg">
          {emailVerifyEnabled ? t('auth.continue') : t('auth.createAccount')}
        </Button>

        <div className="text-center text-sm text-ink-2">
          {t('auth.alreadyHaveAccount')}{' '}
          <Link to="/login" className="link">
            {t('auth.signIn')}
          </Link>
        </div>
      </form>
    </AuthLayout>
  )
}
