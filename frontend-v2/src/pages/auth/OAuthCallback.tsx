import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Copy } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { toast } from '@/components/ui/Toast'

/**
 * Generic OAuth landing page — useful when an admin authorization flow needs
 * the user to copy code/state values back into another window.
 */
export default function OAuthCallbackPage() {
  const { t } = useTranslation()
  const [params] = useSearchParams()
  const code = params.get('code') || ''
  const state = params.get('state') || ''
  const error = params.get('error') || params.get('error_description') || ''
  const fullUrl = useMemo(() => (typeof window !== 'undefined' ? window.location.href : ''), [])

  function copy(value: string) {
    if (!value) return
    navigator.clipboard.writeText(value).then(
      () => toast.success(t('common.copiedToClipboard') as string),
      () => toast.error(t('common.copyFailed') as string)
    )
  }

  const Row = ({ label, value }: { label: string; value: string }) => (
    <div>
      <label className="input-label">{label}</label>
      <div className="flex gap-2">
        <input value={value} readOnly className="input flex-1 font-mono text-xs" />
        <Button type="button" variant="ghost" disabled={!value} onClick={() => copy(value)}>
          <Copy className="h-3.5 w-3.5" />
          {t('keys.copyToClipboard')}
        </Button>
      </div>
    </div>
  )

  return (
    <div className="min-h-screen bg-bg-0 px-4 py-10">
      <div className="mx-auto max-w-2xl">
        <Card className="p-6">
          <h1 className="text-lg font-medium text-ink-1">OAuth Callback</h1>
          <p className="mt-2 text-sm text-ink-3">
            Copy the <code className="font-mono">code</code> (and{' '}
            <code className="font-mono">state</code> if needed) back to the admin authorization
            flow.
          </p>

          <div className="mt-6 space-y-4">
            <Row label={t('auth.oauth.code')} value={code} />
            <Row label={t('auth.oauth.state')} value={state} />
            <Row label={t('auth.oauth.fullUrl')} value={fullUrl} />

            {error && (
              <div className="rounded-lg border border-signal-err/30 bg-signal-err/5 p-3 text-sm text-signal-err">
                {error}
              </div>
            )}
          </div>
        </Card>
      </div>
    </div>
  )
}
