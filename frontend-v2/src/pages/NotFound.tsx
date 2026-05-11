import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function NotFoundPage() {
  const { t } = useTranslation()
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-bg-0 p-6">
      <div className="font-mono text-xs text-ink-3 mb-4">404</div>
      <h1 className="font-display text-display-md text-ink-1 mb-3">{t('notFound.title')}</h1>
      <p className="text-ink-2 mb-8 text-center max-w-md">{t('notFound.description')}</p>
      <Link to="/" className="btn btn-primary">
        {t('home.getStarted')}
      </Link>
    </div>
  )
}
