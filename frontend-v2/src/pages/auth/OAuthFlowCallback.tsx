import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AlertCircle, ShieldCheck } from 'lucide-react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { toast } from '@/components/ui/Toast'
import { TurnstileWidget, type TurnstileWidgetHandle } from '@/components/TurnstileWidget'
import { apiClient } from '@/api/client'
import { authAPI, type OAuthAdoptionDecision, type OAuthTokenPayload, type PendingOAuthExchangeResponse } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const OAUTH_AFFILIATE_CODE_KEY = 'oauth_aff_code'
const AFFILIATE_REFERRAL_CODE_KEY = 'affiliate_referral_code'

type Provider = 'linuxdo' | 'oidc' | 'wechat'
type PendingAction = 'none' | 'choose_account_action' | 'create_account' | 'bind_login'

interface PendingOAuthCreatePayload {
  email: string
  password: string
  verify_code?: string
  invitation_code?: string
  promo_code?: string
}

type PendingCompletion = PendingOAuthExchangeResponse & {
  step?: string
  status?: string
  state?: string
  reason?: string
  code?: string
  pending_email?: string
  resolved_email?: string
  existing_account_email?: string
  email?: string
  suggested_email?: string
  provider_fallback?: string
  intent?: string
  adoption_required?: boolean
  suggested_display_name?: string
  suggested_avatar_url?: string
}

function parseFragmentParams(): URLSearchParams {
  if (typeof window === 'undefined') return new URLSearchParams()
  const raw = window.location.hash
  return new URLSearchParams(raw.startsWith('#') ? raw.slice(1) : raw)
}

function sanitizeRedirectPath(path: string | null | undefined): string {
  const value = (path || '').trim()
  if (!value) return '/dashboard'
  if (!value.startsWith('/')) return '/dashboard'
  if (value.startsWith('//') || value.includes('://')) return '/dashboard'
  if (value.includes('\n') || value.includes('\r')) return '/dashboard'
  return value
}

function readLegacyFragmentLogin(params: URLSearchParams): OAuthTokenPayload | null {
  const accessToken = params.get('access_token')?.trim() || ''
  if (!accessToken) return null
  const expiresIn = Number.parseInt(params.get('expires_in')?.trim() || '', 10)
  return {
    access_token: accessToken,
    refresh_token: params.get('refresh_token')?.trim() || undefined,
    expires_in: Number.isFinite(expiresIn) && expiresIn > 0 ? expiresIn : undefined,
    token_type: params.get('token_type')?.trim() || undefined
  }
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

function isBindCompletion(data: PendingOAuthExchangeResponse): boolean {
  return !hasOAuthToken(data) && !!data.redirect && !data.error
}

function normalizedPendingState(value: string | null | undefined): string {
  return value?.trim().toLowerCase() || ''
}

function extractPendingAccountEmail(completion: PendingCompletion): string {
  return (
    completion.pending_email ||
    completion.existing_account_email ||
    completion.resolved_email ||
    completion.suggested_email ||
    completion.email ||
    ''
  ).trim()
}

function resolvePendingAccountAction(completion: PendingCompletion): PendingAction {
  const raw = normalizedPendingState(
    completion.step || completion.status || completion.state || completion.error || completion.intent
  )
  if (
    ['choice', 'choose_account_action_required', 'choose_account_action', 'choose_account', 'choose', 'account_action_required', 'account_choice_required'].includes(raw)
  ) {
    return 'choose_account_action'
  }
  if (['email_required', 'create_account_required', 'create_account'].includes(raw)) return 'create_account'
  if (
    ['bind_login_required', 'bind_login', 'existing_account', 'existing_account_binding_required', 'existing_account_required', 'adopt_existing_user_by_email'].includes(raw)
  ) {
    return 'bind_login'
  }
  return 'none'
}

function isCreateAccountRecoveryError(error: unknown): boolean {
  const data = (error as {
    response?: {
      data?: {
        reason?: string
        error?: string
        code?: string
        step?: string
        intent?: string
      }
    }
  }).response?.data
  const states = [data?.reason, data?.error, data?.code, data?.step, data?.intent]
    .map((value) => value?.trim().toLowerCase())
    .filter((value): value is string => Boolean(value))

  return states.includes('email_exists') ||
    states.includes('bind_login_required') ||
    states.includes('bind_login') ||
    states.includes('adopt_existing_user_by_email') ||
    states.includes('existing_account_required') ||
    states.includes('existing_account_binding_required')
}

function getRequestErrorMessage(error: unknown, fallback: string): string {
  const err = error as { message?: string; response?: { data?: { detail?: string; message?: string } } }
  return err.response?.data?.detail || err.response?.data?.message || err.message || fallback
}

function providerLabel(provider: Provider, configuredName?: string) {
  if (provider === 'linuxdo') return 'Linux.do'
  if (provider === 'wechat') return 'WeChat'
  return configuredName?.trim() || 'OIDC'
}

function resolveWeChatMode(
  settings: ReturnType<typeof useAuthStore.getState>['publicSettings']
): 'open' | 'mp' {
  if (settings?.wechat_oauth_mp_enabled === true && settings?.wechat_oauth_open_enabled !== true) return 'mp'
  return 'open'
}

function loadOAuthAffiliateCode(): string {
  try {
    return sessionStorage.getItem(OAUTH_AFFILIATE_CODE_KEY)?.trim() || ''
  } catch {
    return ''
  }
}

function loadAffiliateReferralCode(): string {
  try {
    const raw = localStorage.getItem(AFFILIATE_REFERRAL_CODE_KEY)
    if (!raw) return ''
    const parsed = JSON.parse(raw) as { code?: string; expiresAt?: number }
    if (!parsed.code || (parsed.expiresAt ?? 0) <= Date.now()) return ''
    return parsed.code.trim()
  } catch {
    return ''
  }
}

function oauthAffiliateCode() {
  return loadOAuthAffiliateCode() || loadAffiliateReferralCode()
}

export default function OAuthFlowCallbackPage({ provider }: { provider: Provider }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const setToken = useAuthStore((s) => s.setToken)
  const setPendingAuthSession = useAuthStore((s) => s.setPendingAuthSession)
  const clearPendingAuthSession = useAuthStore((s) => s.clearPendingAuthSession)
  const authed = useAuthStore((s) => s.isAuthenticated())
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const providerName = providerLabel(provider, publicSettings?.oidc_oauth_provider_name)

  const [processing, setProcessing] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [redirectTo, setRedirectTo] = useState('/dashboard')
  const [needsInvitation, setNeedsInvitation] = useState(false)
  const [pendingAction, setPendingAction] = useState<PendingAction>('none')
  const [pendingEmail, setPendingEmail] = useState('')
  const [existingAccountEmail, setExistingAccountEmail] = useState('')
  const [legacyPendingToken, setLegacyPendingToken] = useState('')
  const [invitationCode, setInvitationCode] = useState('')
  const [createEmail, setCreateEmail] = useState('')
  const [createPassword, setCreatePassword] = useState('')
  const [createVerifyCode, setCreateVerifyCode] = useState('')
  const [createInvitationCode, setCreateInvitationCode] = useState('')
  const [createAffiliateCode, setCreateAffiliateCode] = useState(oauthAffiliateCode)
  const [sendingCode, setSendingCode] = useState(false)
  const [sendCodeMessage, setSendCodeMessage] = useState('')
  const [sendCodeCountdown, setSendCodeCountdown] = useState(0)
  const [turnstileToken, setTurnstileToken] = useState('')
  const [bindEmail, setBindEmail] = useState('')
  const [bindPassword, setBindPassword] = useState('')
  const [needsAdoptionConfirmation, setNeedsAdoptionConfirmation] = useState(false)
  const [suggestedDisplayName, setSuggestedDisplayName] = useState('')
  const [suggestedAvatarUrl, setSuggestedAvatarUrl] = useState('')
  const [adoptDisplayName, setAdoptDisplayName] = useState(true)
  const [adoptAvatar, setAdoptAvatar] = useState(true)
  const [totpTempToken, setTotpTempToken] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [totpUserEmailMasked, setTotpUserEmailMasked] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const ranRef = useRef(false)
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const turnstileRef = useRef<TurnstileWidgetHandle | null>(null)

  const needsCreateAccount = pendingAction === 'create_account'
  const needsChooser = pendingAction === 'choose_account_action'
  const needsBindLogin = pendingAction === 'bind_login'
  const needsTotpChallenge = !!totpTempToken
  const turnstileRequired = publicSettings?.turnstile_enabled === true && !!publicSettings?.turnstile_site_key
  const providerKey = provider
  const adoptionDecision: OAuthAdoptionDecision = {
    adoptDisplayName,
    adoptAvatar
  }

  function clearSendCodeCountdown() {
    if (countdownRef.current) {
      clearInterval(countdownRef.current)
      countdownRef.current = null
    }
  }

  function startSendCodeCountdown(seconds?: number) {
    clearSendCodeCountdown()
    setSendCodeCountdown(Math.max(0, seconds || 0))
    if (!seconds || seconds <= 0) return
    countdownRef.current = setInterval(() => {
      setSendCodeCountdown((current) => {
        if (current <= 1) {
          clearSendCodeCountdown()
          return 0
        }
        return current - 1
      })
    }, 1000)
  }

  function resetTurnstile() {
    setTurnstileToken('')
    turnstileRef.current?.reset()
  }

  function buildWeChatStartUrl(intent: 'bind_current_user' | 'adopt_existing_user_by_email', redirect = redirectTo) {
    const base = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/$/, '')
    const params = new URLSearchParams({
      mode: resolveWeChatMode(publicSettings),
      redirect,
      intent
    })
    return `${base}/auth/oauth/wechat/start?${params.toString()}`
  }

  function buildWeChatExistingAccountResumePath(email?: string, redirect = redirectTo) {
    const params = new URLSearchParams({
      wechat_bind_existing: '1',
      redirect,
      mode: resolveWeChatMode(publicSettings)
    })
    const trimmed = email?.trim()
    if (trimmed) params.set('email', trimmed)
    return `/auth/wechat/callback?${params.toString()}`
  }

  function persistPending(
    redirect?: string,
    token = legacyPendingToken,
    options?: {
      adoptionRequired?: boolean
      displayName?: string
      avatarUrl?: string
    }
  ) {
    setPendingAuthSession({
      token,
      token_field: 'pending_oauth_token',
      provider,
      redirect: sanitizeRedirectPath(redirect || redirectTo),
      adoption_required: options?.adoptionRequired ?? needsAdoptionConfirmation,
      suggested_display_name: (options?.displayName ?? suggestedDisplayName) || undefined,
      suggested_avatar_url: (options?.avatarUrl ?? suggestedAvatarUrl) || undefined
    })
  }

  async function finalizeCompletion(completion: PendingOAuthExchangeResponse, redirect: string) {
    if (isBindCompletion(completion)) {
      clearPendingAuthSession()
      toast.success(t('profile.authBindings.bindSuccess') as string)
      navigate(sanitizeRedirectPath(completion.redirect || '/profile'), { replace: true })
      return
    }
    if (!hasOAuthToken(completion)) {
      throw new Error(completion.error || (t('auth.loginFailed') as string))
    }
    persistTokenContext(completion)
    await setToken(completion.access_token)
    clearPendingAuthSession()
    toast.success(t('auth.loginSuccess') as string)
    navigate(redirect, { replace: true })
  }

  function applyInteractiveState(completion: PendingCompletion, redirect: string): boolean {
    setSuggestedDisplayName(completion.suggested_display_name || '')
    setSuggestedAvatarUrl(completion.suggested_avatar_url || '')
    if (completion.error === 'invitation_required') {
      setNeedsInvitation(true)
      setPendingAction('none')
      setNeedsAdoptionConfirmation(false)
      setTotpTempToken('')
      persistPending(redirect, legacyPendingToken, {
        adoptionRequired: completion.adoption_required === true,
        displayName: completion.suggested_display_name,
        avatarUrl: completion.suggested_avatar_url
      })
      return true
    }
    if (completion.requires_2fa === true && completion.temp_token) {
      setTotpTempToken(completion.temp_token)
      setTotpCode('')
      setTotpUserEmailMasked(completion.user_email_masked || '')
      setNeedsInvitation(false)
      setNeedsAdoptionConfirmation(false)
      setPendingAction('none')
      persistPending(redirect, legacyPendingToken, {
        adoptionRequired: completion.adoption_required === true,
        displayName: completion.suggested_display_name,
        avatarUrl: completion.suggested_avatar_url
      })
      return true
    }
    const action = resolvePendingAccountAction(completion)
    if (action !== 'none') {
      const email = extractPendingAccountEmail(completion)
      setPendingEmail(email)
      setCreateEmail(email)
      setBindEmail(email)
      setBindPassword('')
      setNeedsInvitation(false)
      setNeedsAdoptionConfirmation(false)
      setPendingAction(action)
      setTotpTempToken('')
      persistPending(redirect, legacyPendingToken, {
        adoptionRequired: completion.adoption_required === true,
        displayName: completion.suggested_display_name,
        avatarUrl: completion.suggested_avatar_url
      })
      return true
    }
    if (
      completion.adoption_required === true &&
      (completion.suggested_display_name || completion.suggested_avatar_url)
    ) {
      setNeedsAdoptionConfirmation(true)
      setNeedsInvitation(false)
      setPendingAction('none')
      setTotpTempToken('')
      persistPending(redirect, legacyPendingToken, {
        adoptionRequired: true,
        displayName: completion.suggested_display_name,
        avatarUrl: completion.suggested_avatar_url
      })
      return true
    }
    if (completion.auth_result === 'pending_session') {
      persistPending(redirect, legacyPendingToken, { adoptionRequired: false })
      return true
    }
    return false
  }

  async function finalizePendingAccountResponse(completion: PendingCompletion) {
    const redirect = sanitizeRedirectPath(completion.redirect || redirectTo)
    setRedirectTo(redirect)
    if (applyInteractiveState(completion, redirect)) {
      setProcessing(false)
      return
    }
    await finalizeCompletion(completion, redirect)
  }

  useEffect(() => {
    if (ranRef.current) return
    ranRef.current = true

    ;(async () => {
      const params = parseFragmentParams()
      const redirect = sanitizeRedirectPath(params.get('redirect') || searchParams.get('redirect') || '/dashboard')
      const legacyLogin = readLegacyFragmentLogin(params)
      const legacyPendingToken = params.get('pending_oauth_token')?.trim() || ''
      const errParam = params.get('error')
      const errDesc = params.get('error_description') || params.get('error_message') || ''
      const queryEmail = searchParams.get('email')?.trim() || ''
      setRedirectTo(redirect)
      if (queryEmail) {
        setExistingAccountEmail(queryEmail)
        setBindEmail(queryEmail)
        setPendingEmail(queryEmail)
      }

      try {
        if (provider === 'wechat' && searchParams.get('wechat_bind_existing') === '1') {
          if (authed) {
            await apiClient.post('/auth/oauth/bind-token')
            window.location.href = buildWeChatStartUrl('bind_current_user', redirect)
            return
          }
          const resumePath = buildWeChatExistingAccountResumePath(queryEmail, redirect)
          const loginParams = new URLSearchParams({ redirect: resumePath })
          if (queryEmail) loginParams.set('email', queryEmail)
          navigate(`/login?${loginParams.toString()}`, { replace: true })
          return
        }

        if (legacyLogin) {
          persistTokenContext(legacyLogin)
          await setToken(legacyLogin.access_token)
          clearPendingAuthSession()
          toast.success(t('auth.loginSuccess') as string)
          navigate(redirect, { replace: true })
          return
        }

        if (errParam === 'invitation_required' && legacyPendingToken) {
          setLegacyPendingToken(legacyPendingToken)
          setNeedsInvitation(true)
          setProcessing(false)
          persistPending(redirect, legacyPendingToken)
          return
        }
        if (errParam) {
          setError(errDesc || errParam)
          setProcessing(false)
          return
        }

        const completion = await authAPI.exchangePendingOAuthCompletion()
        const completionRedirect = sanitizeRedirectPath(completion.redirect || redirect)
        setRedirectTo(completionRedirect)
        if (applyInteractiveState(completion, completionRedirect)) {
          setProcessing(false)
          return
        }
        await finalizeCompletion(completion, completionRedirect)
      } catch (e) {
        clearPendingAuthSession()
        setError(getRequestErrorMessage(e, t('auth.loginFailed') as string))
        setProcessing(false)
      }
    })()
    return () => clearSendCodeCountdown()
  }, [authed, clearPendingAuthSession, navigate, provider, searchParams, setToken, t])

  async function onInvitationSubmit(e: FormEvent) {
    e.preventDefault()
    const code = invitationCode.trim()
    if (!code) return
    setSubmitting(true)
    setFormError(null)
    try {
      const completion = legacyPendingToken
        ? (await apiClient.post<PendingCompletion>(`/auth/oauth/${provider}/complete-registration`, {
            pending_oauth_token: legacyPendingToken,
            invitation_code: code,
            ...(oauthAffiliateCode() ? { aff_code: oauthAffiliateCode() } : {}),
            adopt_display_name: adoptDisplayName,
            adopt_avatar: adoptAvatar
          })).data
        : await authAPI.completeOAuthRegistration(provider, code, adoptionDecision, oauthAffiliateCode() || undefined)
      await finalizePendingAccountResponse(completion)
    } catch (err) {
      setFormError(getRequestErrorMessage(err, t('auth.registrationFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  async function onCreateAccountSubmit(e: FormEvent) {
    e.preventDefault()
    const payload: PendingOAuthCreatePayload & Record<string, boolean | string | undefined> = {
      email: createEmail.trim(),
      password: createPassword,
      verify_code: createVerifyCode.trim() || undefined,
      invitation_code: createInvitationCode.trim() || undefined,
      aff_code: createAffiliateCode.trim() || oauthAffiliateCode() || undefined,
      adopt_display_name: adoptDisplayName,
      adopt_avatar: adoptAvatar
    }
    if (!payload.email || !payload.password) return
    setSubmitting(true)
    setFormError(null)
    try {
      const { data } = await apiClient.post<PendingCompletion>('/auth/oauth/pending/create-account', payload)
      await finalizePendingAccountResponse(data)
    } catch (err) {
      if (isCreateAccountRecoveryError(err)) {
        setPendingEmail(createEmail.trim())
        setBindEmail(createEmail.trim())
        setBindPassword('')
        setPendingAction('bind_login')
        return
      }
      setFormError(getRequestErrorMessage(err, t('auth.registrationFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  async function onSendCreateVerifyCode() {
    const email = createEmail.trim()
    if (!email) return
    if (turnstileRequired && !turnstileToken) {
      setFormError(t('auth.completeVerification') as string)
      return
    }
    setSendingCode(true)
    setFormError(null)
    setSendCodeMessage('')
    try {
      const resp = await authAPI.sendPendingOAuthVerifyCode({
        email,
        pending_oauth_token: legacyPendingToken || undefined,
        turnstile_token: turnstileRequired ? turnstileToken : undefined
      })
      setSendCodeMessage(resp.message || (t('auth.codeSentSuccess') as string))
      startSendCodeCountdown(resp.countdown || 60)
      if (turnstileRequired) resetTurnstile()
    } catch (err) {
      setFormError(getRequestErrorMessage(err, t('auth.sendCodeFailed') as string))
    } finally {
      setSendingCode(false)
    }
  }

  async function onContinueAdoption() {
    setSubmitting(true)
    setFormError(null)
    try {
      const completion = await authAPI.exchangePendingOAuthCompletion(adoptionDecision)
      await finalizePendingAccountResponse(completion)
    } catch (err) {
      setFormError(getRequestErrorMessage(err, t('auth.loginFailed') as string))
      setNeedsAdoptionConfirmation(false)
    } finally {
      setSubmitting(false)
    }
  }

  async function onBindCurrentAccount() {
    setSubmitting(true)
    setFormError(null)
    try {
      await apiClient.post('/auth/oauth/bind-token')
      if (provider === 'wechat') {
        window.location.href = buildWeChatStartUrl('bind_current_user')
        return
      }
      const completion = await authAPI.exchangePendingOAuthCompletion(adoptionDecision)
      await finalizePendingAccountResponse(completion)
    } catch (err) {
      setFormError(getRequestErrorMessage(err, t('auth.loginFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  async function onBindExistingFromInvitation() {
    if (authed) {
      await onBindCurrentAccount()
      return
    }
    const resumePath = buildWeChatExistingAccountResumePath(existingAccountEmail)
    const params = new URLSearchParams({ redirect: resumePath })
    if (existingAccountEmail.trim()) params.set('email', existingAccountEmail.trim())
    navigate(`/login?${params.toString()}`, { replace: true })
  }

  async function onBindLoginSubmit(e: FormEvent) {
    e.preventDefault()
    const email = bindEmail.trim()
    if (!email || !bindPassword) return
    setSubmitting(true)
    setFormError(null)
    try {
      const { data } = await apiClient.post<PendingCompletion>('/auth/oauth/pending/bind-login', {
        email,
        password: bindPassword,
        adopt_display_name: adoptDisplayName,
        adopt_avatar: adoptAvatar
      })
      await finalizePendingAccountResponse(data)
    } catch (err) {
      setFormError(getRequestErrorMessage(err, t('auth.loginFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  async function onTotpSubmit(e: FormEvent) {
    e.preventDefault()
    const code = totpCode.trim()
    if (!totpTempToken || !/^\d{6}$/.test(code)) return
    setSubmitting(true)
    setFormError(null)
    try {
      const completion = await authAPI.login2FA({ temp_token: totpTempToken, totp_code: code })
      await finalizeCompletion(completion, redirectTo)
    } catch (err) {
      setFormError(getRequestErrorMessage(err, t('auth.loginFailed') as string))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout
      title={t(`auth.${providerKey}.callbackTitle`, { providerName }) as string}
      subtitle={(processing
        ? t(`auth.${providerKey}.callbackProcessing`, { providerName })
        : t(`auth.${providerKey}.callbackHint`, { providerName })) as string}
    >
      {processing && <div className="h-8 w-8 mx-auto animate-spin rounded-full border-4 border-orange border-t-transparent" />}

      {(formError || error) && (
        <div className="rounded-xl border border-signal-err/30 bg-signal-err/5 p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-signal-err shrink-0 mt-0.5" />
            <div className="space-y-3">
              <p className="text-sm text-signal-err">{formError || error}</p>
              {error && <Link to="/login" className="btn btn-primary">{t('auth.backToLogin')}</Link>}
            </div>
          </div>
        </div>
      )}

      {needsInvitation && !processing && (
        <form onSubmit={onInvitationSubmit} className="space-y-4">
          <p className="text-sm text-ink-2">{t(`auth.${providerKey}.invitationRequired`, { providerName })}</p>
          <Input name="invitation" value={invitationCode} onChange={(e) => setInvitationCode(e.target.value)} placeholder={t('auth.invitationCodePlaceholder') as string} disabled={submitting} />
          <Button type="submit" className="w-full" loading={submitting} disabled={!invitationCode.trim()}>{t('auth.continue')}</Button>
          {provider === 'wechat' && (
            <div className="rounded-xl border border-line-2 bg-bg-2 p-4 space-y-3">
              <div className="space-y-1">
                <p className="text-sm font-medium text-ink-1">{t('auth.alreadyHaveAccount')}</p>
                <p className="text-xs text-ink-3">
                  {authed
                    ? t('auth.oauthFlow.bindCurrentAccountDescription', { providerName })
                    : t('auth.oauthFlow.signInThenBindDescription', { providerName })}
                </p>
              </div>
              {!authed && (
                <Input
                  name="existing_account_email"
                  type="email"
                  value={existingAccountEmail}
                  onChange={(e) => setExistingAccountEmail(e.target.value)}
                  placeholder={t('auth.emailPlaceholder') as string}
                  disabled={submitting}
                />
              )}
              <Button type="button" variant="ghost" className="w-full" loading={submitting} onClick={onBindExistingFromInvitation}>
                {authed ? t('auth.oauthFlow.bindCurrentAccount') : t('auth.signIn')}
              </Button>
            </div>
          )}
        </form>
      )}

      {needsChooser && !processing && (
        <div className="space-y-4">
          <p className="text-sm text-ink-2">{pendingEmail ? t('auth.oauthFlow.suggestedEmail', { email: pendingEmail }) : t('auth.oauthFlow.chooseAccountActionHint')}</p>
          {authed && (
            <Button type="button" className="w-full" onClick={onBindCurrentAccount} loading={submitting}>
              {t('auth.oauthFlow.bindCurrentAccount')}
            </Button>
          )}
          <Button type="button" className="w-full" onClick={() => setPendingAction('bind_login')}>{t('auth.oauthFlow.bindExistingAccount')}</Button>
          <Button type="button" variant="ghost" className="w-full" onClick={() => setPendingAction('create_account')}>{t('auth.oauthFlow.createNewAccount')}</Button>
        </div>
      )}

      {needsAdoptionConfirmation && !processing && (
        <div className="space-y-4">
          <div className="rounded-xl border border-line-2 bg-bg-2 p-4 space-y-3">
            <p className="text-sm font-medium text-ink-1">{t('auth.oauthFlow.profileDetailsTitle', { providerName })}</p>
            <p className="text-xs text-ink-3">{t('auth.oauthFlow.profileDetailsDescription', { providerName })}</p>
            {suggestedDisplayName && (
              <label className="flex items-start gap-3 rounded-lg border border-line-2 bg-bg-1 p-3 text-sm">
                <input type="checkbox" className="mt-1" checked={adoptDisplayName} onChange={(e) => setAdoptDisplayName(e.target.checked)} disabled={submitting} />
                <span>
                  <span className="block font-medium text-ink-1">{t('auth.oauthFlow.useDisplayName')}</span>
                  <span className="block text-ink-3">{suggestedDisplayName}</span>
                </span>
              </label>
            )}
            {suggestedAvatarUrl && (
              <label className="flex items-start gap-3 rounded-lg border border-line-2 bg-bg-1 p-3 text-sm">
                <input type="checkbox" className="mt-1" checked={adoptAvatar} onChange={(e) => setAdoptAvatar(e.target.checked)} disabled={submitting} />
                <img src={suggestedAvatarUrl} alt="" className="h-10 w-10 rounded-full border border-line-2 object-cover" />
                <span>
                  <span className="block font-medium text-ink-1">{t('auth.oauthFlow.useAvatar')}</span>
                  <span className="block break-all text-ink-3">{suggestedAvatarUrl}</span>
                </span>
              </label>
            )}
          </div>
          <Button type="button" className="w-full" onClick={onContinueAdoption} loading={submitting}>
            {t('auth.continue')}
          </Button>
        </div>
      )}

      {needsCreateAccount && !processing && (
        <form onSubmit={onCreateAccountSubmit} className="space-y-4">
          <p className="text-sm text-ink-2">{t('auth.oauthFlow.createAccountHint')}</p>
          <Input name="email" type="email" value={createEmail} onChange={(e) => setCreateEmail(e.target.value)} placeholder={t('auth.emailPlaceholder') as string} disabled={submitting} required />
          <Input name="password" type="password" value={createPassword} onChange={(e) => setCreatePassword(e.target.value)} placeholder={t('auth.passwordPlaceholder') as string} disabled={submitting} required />
          {turnstileRequired && (
            <TurnstileWidget
              ref={turnstileRef}
              siteKey={publicSettings.turnstile_site_key}
              onVerify={(token) => {
                setTurnstileToken(token)
                setFormError(null)
              }}
              onExpire={() => {
                setTurnstileToken('')
                setFormError(t('auth.turnstileExpired') as string)
              }}
              onError={() => {
                setTurnstileToken('')
                setFormError(t('auth.turnstileFailed') as string)
              }}
            />
          )}
          <div className="space-y-2">
            <Input name="verify_code" value={createVerifyCode} onChange={(e) => setCreateVerifyCode(e.target.value)} placeholder={t('auth.verificationCode') as string} disabled={submitting} />
            <Button
              type="button"
              variant="ghost"
              className="w-full"
              onClick={onSendCreateVerifyCode}
              loading={sendingCode}
              disabled={!createEmail.trim() || submitting || sendCodeCountdown > 0 || (turnstileRequired && !turnstileToken)}
            >
              {sendCodeCountdown > 0 ? t('auth.resendCountdown', { countdown: sendCodeCountdown }) : t('auth.sendVerificationCode')}
            </Button>
            {sendCodeMessage && <p className="text-xs text-signal-ok">{sendCodeMessage}</p>}
          </div>
          <Input name="invitation_code" value={createInvitationCode} onChange={(e) => setCreateInvitationCode(e.target.value)} placeholder={t('auth.invitationCodePlaceholder') as string} disabled={submitting} />
          <Input name="aff_code" value={createAffiliateCode} onChange={(e) => setCreateAffiliateCode(e.target.value)} placeholder={t('auth.oauthFlow.affiliateCodePlaceholder') as string} disabled={submitting} />
          <Button type="submit" className="w-full" loading={submitting} disabled={!createEmail.trim() || !createPassword}>{t('auth.createAccount')}</Button>
          <Button type="button" variant="ghost" className="w-full" onClick={() => setPendingAction('bind_login')} disabled={submitting}>{t('auth.oauthFlow.bindExistingAccount')}</Button>
        </form>
      )}

      {needsBindLogin && !processing && (
        <form onSubmit={onBindLoginSubmit} className="space-y-4">
          <p className="text-sm text-ink-2">{t('auth.oauthFlow.bindLoginHint', { providerName })}</p>
          <Input name="email" type="email" value={bindEmail} onChange={(e) => setBindEmail(e.target.value)} placeholder={t('auth.emailPlaceholder') as string} disabled={submitting} required />
          <Input name="password" type="password" value={bindPassword} onChange={(e) => setBindPassword(e.target.value)} placeholder={t('auth.passwordPlaceholder') as string} disabled={submitting} required />
          <Button type="submit" className="w-full" loading={submitting} disabled={!bindEmail.trim() || !bindPassword}>{t('auth.oauthFlow.logInAndBind')}</Button>
          <Button type="button" variant="ghost" className="w-full" onClick={() => setPendingAction('create_account')} disabled={submitting}>{t('auth.oauthFlow.createNewAccount')}</Button>
        </form>
      )}

      {needsTotpChallenge && !processing && (
        <form onSubmit={onTotpSubmit} className="space-y-4">
          <div className="rounded-xl border border-line-2 bg-bg-2 px-4 py-3 text-sm text-ink-2 flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-orange shrink-0" />
            {t('auth.oauthFlow.totpHint', { providerName, account: totpUserEmailMasked || t('auth.oauthFlow.yourAccount') })}
          </div>
          <Input name="totp" inputMode="numeric" maxLength={6} value={totpCode} onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ''))} placeholder="123456" disabled={submitting} required />
          <Button type="submit" className="w-full" loading={submitting} disabled={totpCode.trim().length !== 6}>{t('auth.oauthFlow.verifyAndContinue')}</Button>
        </form>
      )}
    </AuthLayout>
  )
}
