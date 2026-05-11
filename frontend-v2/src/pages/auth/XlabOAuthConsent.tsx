import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Loader2, ShieldCheck, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { oauthAPI } from '@/api/oauth'

function queryPayload(params: URLSearchParams) {
  return {
    client_id: params.get('client_id') || '',
    redirect_uri: params.get('redirect_uri') || '',
    response_type: params.get('response_type') || '',
    scope: params.get('scope') || '',
    state: params.get('state') || ''
  }
}

export default function XlabOAuthConsentPage() {
  const [params] = useSearchParams()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const payload = queryPayload(params)

  async function approve() {
    setLoading(true)
    setError('')
    try {
      const result: any = await oauthAPI.authorizeXlabOAuth(payload)
      const url = result?.redirect_uri || result?.redirect_url || result?.url
      if (!url) throw new Error('Authorization response did not include a redirect URL')
      window.location.replace(url)
    } catch (e: any) {
      setError(e?.message || 'Unable to complete Xlab authorization')
    } finally {
      setLoading(false)
    }
  }

  function deny() {
    const redirect = payload.redirect_uri
    if (redirect) {
      const url = new URL(redirect)
      url.searchParams.set('error', 'access_denied')
      if (payload.state) url.searchParams.set('state', payload.state)
      window.location.replace(url.toString())
    } else {
      setError('Authorization denied')
    }
  }

  return (
    <main className="min-h-screen bg-bg-0 p-6 text-ink-1 flex items-center justify-center">
      <Card className="w-full max-w-md p-6 text-center">
        <div className="mx-auto h-12 w-12 rounded-lg bg-orange-soft text-orange flex items-center justify-center">
          {error ? <XCircle className="h-6 w-6 text-signal-err" /> : <ShieldCheck className="h-6 w-6" />}
        </div>
        <h1 className="mt-4 font-display text-2xl">{error ? 'Xlab authorization failed' : 'Connect to Miku'}</h1>
        <p className="mt-2 text-sm text-ink-3">
          {error || 'Use your current XlabAPI session to authorize this application.'}
        </p>
        <div className="mt-5 rounded-lg border border-line-2 bg-bg-2 p-4 text-left text-xs">
          <div className="flex justify-between gap-3"><span className="text-ink-3">Client</span><code className="font-mono">{payload.client_id || '-'}</code></div>
          <div className="mt-2 flex justify-between gap-3"><span className="text-ink-3">Scope</span><code className="font-mono">{payload.scope || '-'}</code></div>
        </div>
        <div className="mt-6 flex justify-center gap-2">
          <Button variant="accent" onClick={approve} loading={loading}>
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            Allow
          </Button>
          <Button variant="ghost" onClick={deny}>Deny</Button>
          {error && <Link to="/dashboard" className="btn btn-ghost">Console</Link>}
        </div>
      </Card>
    </main>
  )
}
