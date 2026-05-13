import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AlertCircle } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { apiClient } from '@/api/client'
import { authAPI, type OAuthTokenPayload, type PendingOAuthExchangeResponse } from '@/api/auth'
import { toast } from '@/components/ui/Toast'

const OAUTH_AFFILIATE_CODE_KEY = 'oauth_aff_code'

function parseFragmentParams(): URLSearchParams {
  if (typeof window === 'undefined') return new URLSearchParams()
  const raw = window.location.hash
  return new URLSearchParams(raw.startsWith('#') ? raw.slice(1) : raw)
}

function sanitizeRedirect(path: string | null | undefined): string {
  if (!path) return '/dashboard'
  if (!path.startsWith('/')) return '/dashboard'
  if (path.startsWith('//')) return '/dashboard'
  if (path.includes('://')) return '/dashboard'
  if (path.includes('\n') || path.includes('\r')) return '/dashboard'
  return path
}

function persistTokenContext(tokenData: OAuthTokenPayload) {
  if (tokenData.refresh_token) localStorage.setItem('refresh_token', tokenData.refresh_token)
  if (tokenData.expires_in) {
    localStorage.setItem('token_expires_at', String(Date.now() + tokenData.expires_in * 1000))
  }
}

function hasOAuthToken(data: PendingOAuthExchangeResponse): data is OAuthTokenPayload {
  return typeof data.access_token === 'string' && data.access_token.trim().length > 0
}

function loadOAuthAffiliateCode(): string {
  try {
    return sessionStorage.getItem(OAUTH_AFFILIATE_CODE_KEY)?.trim() || ''
  } catch {
    return ''
  }
}

export default function LinuxDoCallbackPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const setToken = useAuthStore((s) => s.setToken)

  const [processing, setProcessing] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [needsInvitation, setNeedsInvitation] = useState(false)
  const [invitationCode, setInvitationCode] = useState('')
  const [invitationError, setInvitationError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [redirectTo, setRedirectTo] = useState('/dashboard')
  const [legacyPendingToken, setLegacyPendingToken] = useState('')
  const ranRef = useRef(false)

  useEffect(() => {
    if (ranRef.current) return
    ranRef.current = true

    ;(async () => {
      const params = parseFragmentParams()
      const token = params.get('access_token') || ''
      const refreshToken = params.get('refresh_token') || ''
      const expiresInStr = params.get('expires_in') || ''
      const url = new URL(window.location.href)
      const queryRedirect = url.searchParams.get('redirect')
      const redirect = sanitizeRedirect(params.get('redirect') || queryRedirect)
      setRedirectTo(redirect)
      const errParam = params.get('error')
      const errDesc = params.get('error_description') || params.get('error_message') || ''

      if (errParam) {
        if (errParam === 'invitation_required') {
          const pt = params.get('pending_oauth_token') || ''
          if (!pt) {
            setError(t('auth.linuxdo.invalidPendingToken') as string)
            setProcessing(false)
            return
          }
          setLegacyPendingToken(pt)
          setNeedsInvitation(true)
          setProcessing(false)
          return
        }
        setError(errDesc || errParam)
        setProcessing(false)
        return
      }

      if (!token) {
        try {
          const completion = await authAPI.exchangePendingOAuthCompletion()
          const completionRedirect = sanitizeRedirect(completion.redirect || redirect)
          setRedirectTo(completionRedirect)

          if (completion.error === 'invitation_required') {
            setNeedsInvitation(true)
            setProcessing(false)
            return
          }

          if (!hasOAuthToken(completion)) {
            setError((completion.error || t('auth.linuxdo.callbackMissingToken')) as string)
            setProcessing(false)
            return
          }

          persistTokenContext(completion)
          await setToken(completion.access_token)
          toast.success(t('auth.loginSuccess') as string)
          navigate(completionRedirect, { replace: true })
        } catch (e) {
          setError((e as { message?: string })?.message || (t('auth.linuxdo.callbackMissingToken') as string))
          setProcessing(false)
        }
        return
      }

      try {
        const expiresIn = parseInt(expiresInStr, 10)
        persistTokenContext({
          access_token: token,
          refresh_token: refreshToken || undefined,
          expires_in: Number.isNaN(expiresIn) ? undefined : expiresIn
        })
        await setToken(token)
        toast.success(t('auth.loginSuccess') as string)
        navigate(redirect, { replace: true })
      } catch (e) {
        setError((e as { message?: string })?.message || (t('auth.loginFailed') as string))
        setProcessing(false)
      }
    })()
  }, [navigate, setToken, t])

  async function onSubmitInvitation() {
    setInvitationError(null)
    if (!invitationCode.trim()) return
    setSubmitting(true)
    try {
      const affCode = loadOAuthAffiliateCode()
      const tokenData = legacyPendingToken
        ? (await apiClient.post<PendingOAuthExchangeResponse>('/auth/oauth/linuxdo/complete-registration', {
            pending_oauth_token: legacyPendingToken,
            invitation_code: invitationCode.trim(),
            ...(affCode ? { aff_code: affCode } : {})
          })).data
        : await authAPI.completeLinuxDoOAuthRegistration(invitationCode.trim(), undefined, affCode || undefined)
      if (!hasOAuthToken(tokenData)) {
        throw new Error(tokenData.error || (t('auth.linuxdo.completeRegistrationFailed') as string))
      }
      persistTokenContext(tokenData)
      await setToken(tokenData.access_token)
      toast.success(t('auth.loginSuccess') as string)
      navigate(redirectTo, { replace: true })
    } catch (e) {
      setInvitationError(
        (e as { message?: string })?.message || (t('auth.linuxdo.completeRegistrationFailed') as string)
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout
      title={t('auth.linuxdo.callbackTitle') as string}
      subtitle={(processing
        ? t('auth.linuxdo.callbackProcessing')
        : t('auth.linuxdo.callbackHint')) as string}
    >
      {needsInvitation && (
        <div className="space-y-4">
          <p className="text-sm text-ink-2">{t('auth.linuxdo.invitationRequired')}</p>
          <Input
            name="invitation"
            placeholder={t('auth.invitationCodePlaceholder') as string}
            value={invitationCode}
            onChange={(e) => setInvitationCode(e.target.value)}
            disabled={submitting}
            onKeyDown={(e) => {
              if (e.key === 'Enter') onSubmitInvitation()
            }}
            error={invitationError ?? undefined}
          />
          <Button
            onClick={onSubmitInvitation}
            disabled={submitting || !invitationCode.trim()}
            loading={submitting}
            className="w-full"
          >
            {submitting ? t('auth.linuxdo.completing') : t('auth.linuxdo.completeRegistration')}
          </Button>
        </div>
      )}

      {error && (
        <div className="rounded-xl border border-signal-err/30 bg-signal-err/5 p-4 mt-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-signal-err shrink-0 mt-0.5" />
            <div className="space-y-3">
              <p className="text-sm text-signal-err">{error}</p>
              <Link to="/login" className="btn btn-primary">
                {t('auth.linuxdo.backToLogin')}
              </Link>
            </div>
          </div>
        </div>
      )}
    </AuthLayout>
  )
}
