import { useTranslation } from 'react-i18next'
import { ArrowLeft, ExternalLink } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { Card } from '@/components/ui/Card'
import { DEFAULT_DOC_URL, resolveDocLink } from '@/lib/docs'

export default function DocsPage() {
  const { t } = useTranslation()
  const docUrl = useAuthStore((s) => s.publicSettings?.doc_url)
  const apiBaseUrl = useAuthStore((s) => s.publicSettings?.api_base_url)
  const docLink = resolveDocLink(docUrl)
  const displayDocUrl = docUrl?.trim() || DEFAULT_DOC_URL

  return (
    <main className="min-h-screen bg-bg-0 text-ink-1 p-6 sm:p-10">
      <div className="mx-auto max-w-3xl">
        <div className="mb-6">
          <Link to="/" className="mb-5 inline-flex items-center gap-2 text-sm text-ink-3 transition-colors hover:text-ink-1">
            <ArrowLeft className="h-4 w-4" />
            返回首页
          </Link>
          <h1 className="font-display text-3xl leading-tight">{t('docs.title')}</h1>
          <p className="mt-2 text-sm text-ink-2">{t('docs.description')}</p>
        </div>

        <Card className="p-5 space-y-5">
          {displayDocUrl ? (
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 className="text-sm font-medium text-ink-1">{t('docs.externalTitle')}</h2>
                <p className="mt-1 text-sm text-ink-2 break-all">{displayDocUrl}</p>
              </div>
              <a href={docLink.href} target={docLink.target} rel={docLink.rel} className="btn btn-primary">
                {t('docs.openDocs')}
                {docLink.external && <ExternalLink className="h-4 w-4" />}
              </a>
            </div>
          ) : (
            <div>
              <h2 className="text-sm font-medium text-ink-1">{t('docs.inlineTitle')}</h2>
              <p className="mt-1 text-sm text-ink-2">{t('docs.inlineDescription')}</p>
            </div>
          )}

          <div className="border-t border-line-2 pt-4 grid gap-3 sm:grid-cols-2">
            <div className="rounded-lg border border-line-2 bg-bg-2 p-4">
              <div className="text-xs uppercase tracking-wider text-ink-3">{t('docs.apiBase')}</div>
              <code className="mt-2 block text-sm text-ink-1 break-all">{apiBaseUrl || '/v1'}</code>
            </div>
            <div className="rounded-lg border border-line-2 bg-bg-2 p-4">
              <div className="text-xs uppercase tracking-wider text-ink-3">{t('docs.console')}</div>
              <code className="mt-2 block text-sm text-ink-1">/dashboard</code>
            </div>
          </div>
        </Card>
      </div>
    </main>
  )
}
