import { useState, useEffect, type ReactNode } from 'react'
import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  LayoutDashboard,
  KeyRound,
  BarChart3,
  Gift,
  UserCircle,
  Menu,
  X,
  LogOut,
  Shield,
  Megaphone,
  Boxes,
  Layers,
  BadgeCheck
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { LocaleSwitcher } from './LocaleSwitcher'
import { cn } from '@/lib/cn'

interface NavItem {
  to: string
  labelKey: string
  Icon: typeof LayoutDashboard
  adminOnly?: boolean
}

const userNav: NavItem[] = [
  { to: '/dashboard', labelKey: 'nav.dashboard', Icon: LayoutDashboard },
  { to: '/keys', labelKey: 'nav.apiKeys', Icon: KeyRound },
  { to: '/models', labelKey: 'nav.modelHub', Icon: Layers },
  { to: '/usage', labelKey: 'nav.usage', Icon: BarChart3 },
  { to: '/subscriptions', labelKey: 'nav.mySubscriptions', Icon: BadgeCheck },
  { to: '/redeem', labelKey: 'nav.redeem', Icon: Gift },
  { to: '/profile', labelKey: 'nav.profile', Icon: UserCircle }
]

const adminNav: NavItem[] = [
  { to: '/admin', labelKey: 'nav.dashboard', Icon: LayoutDashboard },
  { to: '/admin/users', labelKey: 'nav.users', Icon: UserCircle },
  { to: '/admin/announcements', labelKey: 'nav.announcements', Icon: Megaphone },
  { to: '/admin/groups', labelKey: 'nav.groups', Icon: Boxes }
]

export function ConsoleLayout({ admin, children }: { admin?: boolean; children?: ReactNode }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const logout = useAuthStore((s) => s.logout)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const siteName = publicSettings?.site_name || 'Sub2API'
  const siteLogo = publicSettings?.site_logo || '/logo.png'

  const [mobileOpen, setMobileOpen] = useState(false)
  const items = admin ? adminNav : userNav

  useEffect(() => {
    setMobileOpen(false)
  }, [location.pathname])

  async function onLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-screen flex bg-bg-0">
      {/* Sidebar */}
      <aside
        className={cn(
          'fixed lg:static inset-y-0 left-0 z-40 w-60 transform transition-transform lg:translate-x-0',
          'bg-bg-1 border-r border-line-2 flex flex-col',
          mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        )}
      >
        <Link to="/" className="px-5 h-16 flex items-center gap-3 border-b border-line-2">
          <div className="h-8 w-8 rounded-lg bg-bg-3 border border-line-2 flex items-center justify-center overflow-hidden">
            <img src={siteLogo} alt="" className="h-full w-full object-contain" onError={(e) => { e.currentTarget.style.display = 'none' }} />
          </div>
          <span className="font-medium text-ink-1 truncate">{siteName}</span>
        </Link>

        <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-0.5">
          {items.map(({ to, labelKey, Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/admin'}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors',
                  isActive
                    ? 'bg-orange-soft text-orange font-medium'
                    : 'text-ink-2 hover:bg-bg-3'
                )
              }
            >
              <Icon className="h-4 w-4 shrink-0" />
              <span>{t(labelKey)}</span>
            </NavLink>
          ))}

          {isAdmin && !admin && (
            <NavLink
              to="/admin"
              className="mt-3 flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-ink-2 hover:bg-bg-3 border border-line-2"
            >
              <Shield className="h-4 w-4 shrink-0" />
              <span>Admin Console</span>
            </NavLink>
          )}
          {admin && (
            <NavLink
              to="/dashboard"
              className="mt-3 flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-ink-2 hover:bg-bg-3 border border-line-2"
            >
              <LayoutDashboard className="h-4 w-4 shrink-0" />
              <span>User Console</span>
            </NavLink>
          )}
        </nav>

        <div className="px-3 py-3 border-t border-line-2 space-y-1">
          <div className="px-3 py-2">
            <div className="text-xs text-ink-3 truncate">{user?.email}</div>
            <div className="text-xs text-ink-3 mt-0.5">
              {t('common.balance')}: <span className="font-mono text-ink-2">${user?.balance?.toFixed(4) ?? '0.0000'}</span>
            </div>
          </div>
          <button onClick={onLogout} className="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-ink-2 hover:bg-bg-3">
            <LogOut className="h-4 w-4" />
            <span>{t('nav.logout')}</span>
          </button>
        </div>
      </aside>

      {mobileOpen && (
        <div className="lg:hidden fixed inset-0 z-30 bg-bg-0/30" onClick={() => setMobileOpen(false)} aria-hidden />
      )}

      {/* Main content */}
      <div className="flex-1 min-w-0 flex flex-col">
        <header className="h-16 px-5 flex items-center justify-between border-b border-line-2 bg-bg-1">
          <button
            className="lg:hidden btn btn-ghost btn-icon"
            onClick={() => setMobileOpen((v) => !v)}
            aria-label="Toggle menu"
          >
            {mobileOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </button>
          <div className="flex-1" />
          <LocaleSwitcher compact />
        </header>
        <main className="flex-1 p-5 sm:p-8 overflow-y-auto">
          {children ?? <Outlet />}
        </main>
      </div>
    </div>
  )
}

export function PageHeader({ title, description, actions }: { title: ReactNode; description?: ReactNode; actions?: ReactNode }) {
  return (
    <div className="flex items-start justify-between gap-4 mb-6">
      <div className="min-w-0">
        <h1 className="font-display text-3xl text-ink-1 leading-tight">{title}</h1>
        {description && <p className="mt-1.5 text-sm text-ink-2">{description}</p>}
      </div>
      {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
    </div>
  )
}
