import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function NotFoundPage() {
  const { t } = useTranslation()
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-bg-0 p-6">
      <div className="font-mono text-xs text-ink-3 mb-4">404</div>
      <h1 className="font-display text-display-md text-ink-1 mb-3">Page not found</h1>
      <p className="text-ink-2 mb-8 text-center max-w-md">
        The page you’re looking for doesn’t exist or has been moved.
      </p>
      <Link to="/" className="btn btn-primary">
        {t('home.goToDashboard')}
      </Link>
    </div>
  )
}
