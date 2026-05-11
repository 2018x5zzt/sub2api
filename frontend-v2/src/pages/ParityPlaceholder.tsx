import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ArrowRight, ExternalLink } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'

interface ParityPlaceholderProps {
  title: string
  description?: string
  legacyPath?: string
  endpoints?: string[]
  actions?: Array<{ label: string; to: string }>
  standalone?: boolean
}

export default function ParityPlaceholder({
  title,
  description,
  legacyPath,
  endpoints = [],
  actions = [],
  standalone = false
}: ParityPlaceholderProps) {
  const { t } = useTranslation()

  const content = (
    <>
      <PageHeader
        title={title}
        description={description || (t('parity.description') as string)}
        actions={actions.map((action) => (
          <Link key={action.to} to={action.to} className="btn btn-ghost btn-sm">
            {action.label}
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        ))}
      />

      <Card className="p-5 space-y-4">
        <div className="flex items-start gap-3">
          <div className="h-9 w-9 rounded-lg bg-orange-soft text-orange flex items-center justify-center shrink-0">
            <ExternalLink className="h-4 w-4" />
          </div>
          <div className="min-w-0">
            <h2 className="text-sm font-medium text-ink-1">{t('parity.entryRestored')}</h2>
            <p className="mt-1 text-sm text-ink-2 leading-6">{t('parity.entryRestoredDescription')}</p>
            {legacyPath && (
              <p className="mt-2 text-xs text-ink-3">
                {t('parity.legacyPath')}: <code className="font-mono text-ink-2">{legacyPath}</code>
              </p>
            )}
          </div>
        </div>

        {endpoints.length > 0 && (
          <div className="border-t border-line-2 pt-4">
            <div className="text-xs font-medium uppercase tracking-wider text-ink-3 mb-2">
              {t('parity.expectedApiSurface')}
            </div>
            <div className="flex flex-wrap gap-2">
              {endpoints.map((endpoint) => (
                <code key={endpoint} className="rounded-md border border-line-2 bg-bg-3 px-2 py-1 text-xs text-ink-2">
                  {endpoint}
                </code>
              ))}
            </div>
          </div>
        )}
      </Card>
    </>
  )

  if (!standalone) return content

  return (
    <main className="min-h-screen bg-bg-0 p-6 text-ink-1 sm:p-10">
      <div className="mx-auto max-w-4xl">
        {content}
      </div>
    </main>
  )
}
