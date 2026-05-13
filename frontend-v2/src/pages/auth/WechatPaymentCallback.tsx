import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'

function parseFragmentParams(): URLSearchParams {
  if (typeof window === 'undefined') return new URLSearchParams()
  const raw = window.location.hash
  return new URLSearchParams(raw.startsWith('#') ? raw.slice(1) : raw)
}

function normalizeRedirectPath(path: string | null | undefined): string {
  const value = (path || '').trim()
  if (!value) return '/purchase'
  if (!value.startsWith('/')) return '/purchase'
  if (value.startsWith('//') || value.includes('://')) return '/purchase'
  if (value === '/payment') return '/purchase'
  if (value.startsWith('/payment?')) return `/purchase${value.slice('/payment'.length)}`
  return value
}

function appendQueryParam(query: URLSearchParams, key: string, value: string | null) {
  if (value) query.set(key, value)
}

export default function WechatPaymentCallbackPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [error, setError] = useState('')

  useEffect(() => {
    const fragment = parseFragmentParams()
    const readParam = (key: string) => fragment.get(key) || searchParams.get(key) || ''
    const err = readParam('error') || readParam('err_msg') || readParam('errmsg')
    const errDesc = readParam('error_description') || readParam('message')
    if (err) {
      setError(errDesc || err)
      return
    }

    const resumeToken = readParam('wechat_resume_token')
    const openid = readParam('openid')
    if (!resumeToken && !openid) {
      setError(t('auth.wechatPayment.callbackMissingResumeToken') as string)
      return
    }

    const redirectURL = new URL(normalizeRedirectPath(readParam('redirect')), window.location.origin)
    const query = new URLSearchParams(redirectURL.search)
    query.set('wechat_resume', '1')

    if (resumeToken) {
      query.set('wechat_resume_token', resumeToken)
    } else {
      query.set('openid', openid)
      appendQueryParam(query, 'state', readParam('state'))
      appendQueryParam(query, 'scope', readParam('scope'))
      appendQueryParam(query, 'payment_type', readParam('payment_type'))
      appendQueryParam(query, 'amount', readParam('amount'))
      appendQueryParam(query, 'order_type', readParam('order_type'))
      appendQueryParam(query, 'plan_id', readParam('plan_id'))
    }

    navigate(`${redirectURL.pathname}?${query.toString()}`, { replace: true })
  }, [navigate, searchParams, t])

  return (
    <div className="min-h-screen bg-bg-0 px-4 py-10">
      <div className="mx-auto max-w-2xl">
        <Card className="p-6">
          <h1 className="text-lg font-medium text-ink-1">{t('auth.wechatPayment.callbackTitle')}</h1>
          <p className="mt-2 text-sm text-ink-2">{error || t('auth.wechatPayment.callbackProcessing')}</p>
          {!error ? (
            <div className="mt-6 flex items-center justify-center py-10">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-orange border-t-transparent" />
            </div>
          ) : (
            <div className="mt-6 rounded-lg border border-line-2 bg-bg-2 p-4">
              <p className="text-sm text-ink-2">{error}</p>
              <Button type="button" className="mt-4" onClick={() => navigate('/purchase', { replace: true })}>
                {t('auth.wechatPayment.backToPayment')}
              </Button>
            </div>
          )}
        </Card>
      </div>
    </div>
  )
}
