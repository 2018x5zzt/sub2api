import { Link } from 'react-router-dom'
import type { ReactNode } from 'react'
import { useAuthStore } from '@/stores/auth'
import { LocaleSwitcher } from './LocaleSwitcher'

export function AuthLayout({ children, title, subtitle }: { children: ReactNode; title?: string; subtitle?: string }) {
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const siteName = publicSettings?.site_name || 'Sub2API'
  const siteLogo = publicSettings?.site_logo || '/logo.png'

  return (
    <div className="min-h-screen flex flex-col bg-bg-0">
      <header className="px-6 py-4">
        <div className="mx-auto max-w-6xl flex items-center justify-between">
          <Link to="/" className="flex items-center gap-3">
            <div className="h-9 w-9 rounded-lg bg-bg-1 border border-line-2 flex items-center justify-center overflow-hidden">
              <img src={siteLogo} alt="" className="h-full w-full object-contain" onError={(e) => { e.currentTarget.style.display = 'none' }} />
            </div>
            <span className="font-medium text-ink-1">{siteName}</span>
          </Link>
          <LocaleSwitcher compact />
        </div>
      </header>

      <main className="flex-1 flex items-center justify-center px-4 py-8">
        <div className="w-full max-w-md">
          {title && (
            <div className="text-center mb-8">
              <h1 className="font-display text-display-md text-ink-1">{title}</h1>
              {subtitle && <p className="mt-2 text-sm text-ink-2">{subtitle}</p>}
            </div>
          )}
          <div className="card p-8">{children}</div>
        </div>
      </main>
    </div>
  )
}
