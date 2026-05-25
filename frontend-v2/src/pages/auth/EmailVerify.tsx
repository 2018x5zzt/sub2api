import { useEffect, useRef, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, AlertCircle } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Button } from '@/components/ui/Button'
import { useAuthStore } from '@/stores/auth'
import { authAPI } from '@/api/auth'
import { toast } from '@/components/ui/Toast'
import { loadAffiliateReferralCode } from '@/utils/affiliateReferral'

interface RegisterData {
  email: string
  password: string
  turnstile_token?: string
  promo_code?: string
  invitation_code?: string
  aff_code?: string
}

const REGISTER_DATA_KEY = 'register_data'

export default function EmailVerifyPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const register = useAuthStore((s) => s.register)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const siteName = publicSettings?.site_name || 'Xlabapi'

  const [data, setData] = useState<RegisterData | null>(null)
  const [code, setCode] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [sending, setSending] = useState(false)
  const [codeSent, setCodeSent] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const [codeError, setCodeError] = useState<string | null>(null)

  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const initialSendRef = useRef(false)

  function startCountdown(seconds: number) {
    setCountdown(seconds)
    if (timerRef.current) clearInterval(timerRef.current)
    timerRef.current = setInterval(() => {
      setCountdown((c) => {
        if (c <= 1) {
          if (timerRef.current) {
            clearInterval(timerRef.current)
            timerRef.current = null
          }
          return 0
        }
        return c - 1
      })
    }, 1000)
  }

  async function sendCode(emailToUse: string, turnstileToken?: string) {
    setSending(true)
    setError(null)
    try {
      const resp = await authAPI.sendVerifyCode({
        email: emailToUse,
        turnstile_token: turnstileToken
      })
      setCodeSent(true)
      startCountdown(resp.countdown || 60)
    } catch (e) {
      setError((e as { message?: string })?.message || (t('auth.sendCodeFailed') as string))
    } finally {
      setSending(false)
    }
  }

  // Load registration data on mount and auto-send code once
  useEffect(() => {
    const raw = sessionStorage.getItem(REGISTER_DATA_KEY)
    if (raw) {
      try {
        const parsed = JSON.parse(raw) as RegisterData
        if (parsed.email && parsed.password) {
          setData(parsed)
          if (!initialSendRef.current) {
            initialSendRef.current = true
            sendCode(parsed.email, parsed.turnstile_token)
          }
        }
      } catch {
        // ignore
      }
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function onResend() {
    if (!data) return
    await sendCode(data.email)
  }

  function validate(): boolean {
    setCodeError(null)
    const trimmed = code.trim()
    if (!trimmed) {
      setCodeError(t('auth.codeRequired') as string)
      return false
    }
    if (!/^\d{6}$/.test(trimmed)) {
      setCodeError(t('auth.invalidCode') as string)
      return false
    }
    return true
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!data || !validate()) return
    setSubmitting(true)
    setError(null)
    try {
      const affCode = data.aff_code?.trim() || loadAffiliateReferralCode()
      await register({
        email: data.email,
        password: data.password,
        verify_code: code.trim(),
        turnstile_token: data.turnstile_token,
        promo_code: data.promo_code,
        invitation_code: data.invitation_code,
        ...(affCode ? { aff_code: affCode } : {})
      })
      sessionStorage.removeItem(REGISTER_DATA_KEY)
      toast.success(t('auth.accountCreatedSuccess', { siteName }) as string)
      navigate('/dashboard', { replace: true })
    } catch (err) {
      setError((err as { message?: string })?.message || (t('auth.verifyFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  function onBack() {
    sessionStorage.removeItem(REGISTER_DATA_KEY)
    navigate('/register')
  }

  if (!data) {
    return (
      <AuthLayout title={t('auth.verifyYourEmail') as string}>
        <div className="rounded-xl border border-signal-warn/30 bg-signal-warn/5 p-5 flex items-start gap-3">
          <AlertCircle className="h-5 w-5 text-signal-warn shrink-0 mt-0.5" />
          <div className="text-sm">
            <p className="font-medium text-ink-1">{t('auth.sessionExpired')}</p>
            <p className="mt-1 text-ink-2">{t('auth.sessionExpiredDesc')}</p>
          </div>
        </div>
        <button onClick={onBack} className="mt-5 inline-flex items-center gap-2 text-sm text-ink-3 hover:text-ink-1">
          <ArrowLeft className="h-4 w-4" />
          {t('auth.backToRegistration')}
        </button>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout
      title={t('auth.verifyYourEmail') as string}
      subtitle={`${t('auth.sendCodeDesc')} ${data.email}`}
    >
      <form onSubmit={onSubmit} className="space-y-5">
        <div>
          <label htmlFor="code" className="input-label text-center block">
            {t('auth.verificationCode')}
          </label>
          <input
            id="code"
            name="code"
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            required
            maxLength={6}
            disabled={submitting}
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
            placeholder="000000"
            className={`input py-3 text-center font-mono text-xl tracking-[0.5em] ${codeError ? 'input-error' : ''}`}
          />
          {codeError ? (
            <p className="input-error-text text-center">{codeError}</p>
          ) : (
            <p className="text-xs mt-1 text-ink-3 text-center">{t('auth.verificationCodeHint')}</p>
          )}
        </div>

        {codeSent && !error && (
          <div className="rounded-xl border border-signal-ok/30 bg-signal-ok/5 px-4 py-3 text-sm text-signal-ok">
            {t('auth.codeSentSuccess')}
          </div>
        )}

        {error && (
          <div className="rounded-xl border border-signal-err/30 bg-signal-err/5 px-4 py-3 text-sm text-signal-err">
            {error}
          </div>
        )}

        <Button type="submit" loading={submitting} disabled={!code} className="w-full" size="lg">
          {submitting ? t('auth.verifying') : t('auth.verifyAndCreate')}
        </Button>

        <div className="text-center">
          {countdown > 0 ? (
            <span className="text-sm text-ink-3">{t('auth.resendCountdown', { countdown })}</span>
          ) : (
            <button
              type="button"
              onClick={onResend}
              disabled={sending}
              className="link text-sm disabled:opacity-50"
            >
              {sending ? t('auth.sendingCode') : t('auth.resendCode')}
            </button>
          )}
        </div>

        <div className="text-center">
          <button
            type="button"
            onClick={onBack}
            className="inline-flex items-center gap-2 text-sm text-ink-3 hover:text-ink-1"
          >
            <ArrowLeft className="h-4 w-4" />
            {t('auth.backToRegistration')}
          </button>
        </div>
      </form>
    </AuthLayout>
  )
}
