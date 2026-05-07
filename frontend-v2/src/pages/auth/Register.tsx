import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Mail, Lock, KeyRound, Gift } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { useAuthStore } from '@/stores/auth'
import { authAPI } from '@/api/auth'
import { toast } from '@/components/ui/Toast'

export default function RegisterPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const register = useAuthStore((s) => s.register)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const siteName = publicSettings?.site_name || 'Sub2API'

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [promoCode, setPromoCode] = useState('')
  const [invitationCode, setInvitationCode] = useState('')
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const emailVerifyEnabled = publicSettings?.email_verify_enabled
  const promoEnabled = publicSettings?.promo_code_enabled
  const invitationEnabled = publicSettings?.invitation_code_enabled

  async function sendCode() {
    if (!email) {
      setError(t('auth.emailRequired') as string)
      return
    }
    setSendingCode(true)
    setError(null)
    try {
      const resp = await authAPI.sendVerifyCode({ email })
      toast.success(t('auth.codeSentSuccess') as string)
      const seconds = resp.countdown || 60
      setCountdown(seconds)
      const timer = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) {
            clearInterval(timer)
            return 0
          }
          return c - 1
        })
      }, 1000)
    } catch (e) {
      setError((e as { message?: string })?.message || (t('auth.sendCodeFailed') as string))
    } finally {
      setSendingCode(false)
    }
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await register({
        email,
        password,
        verify_code: emailVerifyEnabled ? verifyCode : undefined,
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
    <AuthLayout title={t('auth.createAccount') as string} subtitle={t('auth.signUpToStart', { siteName }) as string}>
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

        {emailVerifyEnabled && (
          <div className="flex gap-2 items-end">
            <div className="flex-1">
              <Input
                name="verify_code"
                label={t('auth.verificationCode') as string}
                placeholder={t('auth.verificationCodeHint') as string}
                required
                value={verifyCode}
                onChange={(e) => setVerifyCode(e.target.value)}
                leftIcon={<KeyRound className="h-4 w-4" />}
                disabled={submitting}
              />
            </div>
            <Button
              type="button"
              variant="secondary"
              onClick={sendCode}
              disabled={sendingCode || countdown > 0 || submitting}
              loading={sendingCode}
            >
              {countdown > 0 ? `${countdown}s` : t('auth.resendCode')}
            </Button>
          </div>
        )}

        <Input
          name="password"
          type="password"
          label={t('auth.passwordLabel') as string}
          placeholder={t('auth.passwordPlaceholder') as string}
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
          {t('auth.createAccount')}
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
