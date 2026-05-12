import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ExternalLink, Link as LinkIcon } from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { buildEmbeddedUrl, detectTheme } from '@/lib/embedded-url'

function isValidEmbedUrl(url: string) {
  return url.startsWith('http://') || url.startsWith('https://')
}

export default function CustomPage() {
  const { t, i18n } = useTranslation()
  const params = useParams()
  const user = useAuthStore((s) => s.user)
  const token = useAuthStore((s) => s.token)
  const initialized = useAuthStore((s) => s.initialized)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const item = publicSettings?.custom_menu_items?.find((menuItem) => menuItem.id === params.id)

  const embeddedUrl = useMemo(() => {
    if (!item) return ''
    return buildEmbeddedUrl(item.url, user?.id, token, detectTheme(), i18n.language)
  }, [i18n.language, item, token, user?.id])

  if (!initialized || !publicSettings) {
    return (
      <Card className="mx-auto flex max-w-xl items-center justify-center gap-3 p-8 text-sm text-ink-2">
        <Spinner />
        {t('common.loading')}
      </Card>
    )
  }

  if (!item) {
    return (
      <Card className="mx-auto max-w-xl p-8 text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-bg-2 text-ink-3">
          <LinkIcon className="h-5 w-5" />
        </div>
        <h1 className="text-lg font-medium text-ink-1">{t('customPage.notFoundTitle')}</h1>
        <p className="mt-2 text-sm text-ink-2">{t('customPage.notFoundDesc')}</p>
      </Card>
    )
  }

  if (!isValidEmbedUrl(embeddedUrl)) {
    return (
      <Card className="mx-auto max-w-xl p-8 text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-bg-2 text-ink-3">
          <LinkIcon className="h-5 w-5" />
        </div>
        <h1 className="text-lg font-medium text-ink-1">{t('customPage.notConfiguredTitle')}</h1>
        <p className="mt-2 text-sm text-ink-2">{t('customPage.notConfiguredDesc')}</p>
      </Card>
    )
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={item.label}
        actions={(
          <a href={embeddedUrl} target="_blank" rel="noreferrer" className="btn btn-ghost btn-sm">
            {t('customPage.openInNewTab')}
            <ExternalLink className="h-3.5 w-3.5" />
          </a>
        )}
      />
      <Card className="overflow-hidden p-0">
        <iframe
          src={embeddedUrl}
          title={item.label}
          className="block h-[calc(100vh-220px)] min-h-[560px] w-full border-0 bg-bg-1"
          allowFullScreen
        />
      </Card>
    </div>
  )
}
