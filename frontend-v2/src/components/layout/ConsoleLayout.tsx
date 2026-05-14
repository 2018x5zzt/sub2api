import { useState, useEffect, type ReactNode } from 'react'
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
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
  Megaphone,
  Boxes,
  Layers,
  BadgeCheck,
  Server,
  Settings,
  Ticket,
  Database,
  CreditCard,
  RadioTower,
  Activity,
  HandCoins,
  Image,
  ShoppingCart,
  Network,
  LineChart,
  ReceiptText,
  Link as LinkIcon
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { LocaleSwitcher } from './LocaleSwitcher'
import { cn } from '@/lib/cn'
import type { CustomMenuItem, PublicSettings } from '@/types'

interface NavItem {
  to: string
  labelKey?: string
  label?: string
  Icon: typeof LayoutDashboard
  hideInSimpleMode?: boolean
  featureFlag?: (settings: PublicSettings | null) => boolean
}

const userNav: NavItem[] = [
  { to: '/dashboard', labelKey: 'nav.dashboard', Icon: LayoutDashboard },
  { to: '/keys', labelKey: 'nav.apiKeys', Icon: KeyRound },
  { to: '/models', labelKey: 'nav.modelHub', Icon: Layers },
  { to: '/usage', labelKey: 'nav.usage', Icon: BarChart3, hideInSimpleMode: true },
  { to: '/available-channels', labelKey: 'nav.availableChannels', Icon: Network, hideInSimpleMode: true, featureFlag: (s) => s?.available_channels_enabled === true },
  { to: '/monitor', labelKey: 'nav.channelStatus', Icon: RadioTower, featureFlag: (s) => s?.channel_monitor_enabled !== false },
  { to: '/subscriptions', labelKey: 'nav.mySubscriptions', Icon: BadgeCheck, hideInSimpleMode: true },
  { to: '/purchase', labelKey: 'nav.buySubscription', Icon: CreditCard, hideInSimpleMode: true, featureFlag: (s) => s?.payment_enabled !== false },
  { to: '/orders', labelKey: 'nav.myOrders', Icon: ReceiptText, hideInSimpleMode: true, featureFlag: (s) => s?.payment_enabled !== false },
  { to: '/redeem', labelKey: 'nav.redeem', Icon: Gift, hideInSimpleMode: true },
  { to: '/affiliate', labelKey: 'nav.affiliate', Icon: HandCoins, hideInSimpleMode: true, featureFlag: (s) => s?.affiliate_enabled === true },
  { to: '/profile', labelKey: 'nav.profile', Icon: UserCircle },
  { to: '/image-studio', labelKey: 'nav.imageStudio', Icon: Image }
]

const adminNav: NavItem[] = [
  { to: '/admin/dashboard', labelKey: 'nav.dashboard', Icon: LayoutDashboard },
  { to: '/admin/ops', labelKey: 'nav.ops', Icon: Activity },
  { to: '/admin/users', labelKey: 'nav.users', Icon: UserCircle },
  { to: '/admin/groups', labelKey: 'nav.groups', Icon: Boxes },
  { to: '/admin/channels/pricing', labelKey: 'nav.channelPricing', Icon: LineChart },
  { to: '/admin/channels/monitor', labelKey: 'nav.channelMonitor', Icon: RadioTower },
  { to: '/admin/accounts', labelKey: 'nav.accounts', Icon: Server },
  { to: '/admin/subscriptions', labelKey: 'nav.subscriptions', Icon: BadgeCheck },
  { to: '/admin/subscription-product-config', labelKey: 'nav.subscriptionProductConfig', Icon: CreditCard },
  { to: '/admin/usage', labelKey: 'nav.usage', Icon: BarChart3 },
  { to: '/admin/redeem', labelKey: 'nav.redeemCodes', Icon: Gift },
  { to: '/admin/promo-codes', labelKey: 'nav.promoCodes', Icon: Ticket },
  { to: '/admin/announcements', labelKey: 'nav.announcements', Icon: Megaphone },
  { to: '/admin/proxies', labelKey: 'nav.proxies', Icon: Server },
  { to: '/admin/affiliates/invites', labelKey: 'nav.affiliateInviteRecords', Icon: HandCoins, hideInSimpleMode: true, featureFlag: (s) => s?.affiliate_enabled === true },
  { to: '/admin/affiliates/rebates', labelKey: 'nav.affiliateRebateRecords', Icon: ReceiptText, hideInSimpleMode: true, featureFlag: (s) => s?.affiliate_enabled === true },
  { to: '/admin/affiliates/transfers', labelKey: 'nav.affiliateTransferRecords', Icon: CreditCard, hideInSimpleMode: true, featureFlag: (s) => s?.affiliate_enabled === true },
  { to: '/admin/orders/dashboard', labelKey: 'nav.paymentDashboard', Icon: ShoppingCart, hideInSimpleMode: true, featureFlag: (s) => s?.payment_enabled !== false },
  { to: '/admin/orders', labelKey: 'nav.orderManagement', Icon: ReceiptText, hideInSimpleMode: true, featureFlag: (s) => s?.payment_enabled !== false },
  { to: '/admin/orders/plans', labelKey: 'nav.paymentPlans', Icon: CreditCard, hideInSimpleMode: true, featureFlag: (s) => s?.payment_enabled !== false },
  { to: '/admin/backup', labelKey: 'common.backup', Icon: Database },
  { to: '/admin/settings', labelKey: 'nav.settings', Icon: Settings }
]

function customMenuItemPath(item: CustomMenuItem): string {
  return item.url.startsWith('/') ? item.url : `/custom/${item.id}`
}

function customNavItems(items: CustomMenuItem[], visibility: CustomMenuItem['visibility']): NavItem[] {
  return items
    .filter((item) => item.visibility === visibility)
    .sort((a, b) => a.sort_order - b.sort_order)
    .map((item) => ({
      to: customMenuItemPath(item),
      label: item.label,
      Icon: LinkIcon
    }))
}

function visibleNavItems(items: NavItem[], settings: PublicSettings | null, runMode: 'standard' | 'simple') {
  return items.filter((item) => {
    if (runMode === 'simple' && item.hideInSimpleMode) return false
    if (item.featureFlag && item.featureFlag(settings) === false) return false
    return true
  })
}

function navLabel(item: NavItem, t: (key: string) => string) {
  return item.label ?? (item.labelKey ? t(item.labelKey) : '')
}

export function ConsoleLayout({ admin, children }: { admin?: boolean; children?: ReactNode }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const logout = useAuthStore((s) => s.logout)
  const runMode = useAuthStore((s) => s.runMode)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const siteName = publicSettings?.site_name || 'XlabAPI'
  const siteLogo = publicSettings?.site_logo || '/logo.png'

  const [mobileOpen, setMobileOpen] = useState(false)
  const userCustomNav = customNavItems(publicSettings?.custom_menu_items ?? [], 'user')
  const items = publicSettings?.backend_mode_enabled ? [] : visibleNavItems([...userNav, ...userCustomNav], publicSettings, runMode)
  const showAdminItems = isAdmin || admin
  const adminItems = visibleNavItems(adminNav, publicSettings, runMode)

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
          {items.map((item) => {
            const { to, Icon } = item
            return (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors',
                  isActive
                    ? 'bg-orange-soft text-orange font-medium'
                    : 'text-ink-2 hover:bg-bg-3'
                )
              }
              end={to === '/admin/dashboard'}
            >
              <Icon className="h-4 w-4 shrink-0" />
              <span>{navLabel(item, t)}</span>
            </NavLink>
            )
          })}

          {showAdminItems && (
            <div className="mt-4 border-t border-line-2 pt-3">
              <div className="px-3 pb-2 text-[11px] font-mono uppercase tracking-[0.14em] text-ink-3">
                {t('layout.adminBanner')}
              </div>
              <div className="space-y-0.5">
                {adminItems.map((item) => {
                  const { to, Icon } = item
                  return (
                  <NavLink
                    key={to}
                    to={to}
                    className={({ isActive }) =>
                      cn(
                        'flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors',
                        isActive
                          ? 'bg-orange-soft text-orange font-medium'
                          : 'text-ink-2 hover:bg-bg-3'
                      )
                    }
                    end={to === '/admin/dashboard'}
                  >
                    <Icon className="h-4 w-4 shrink-0" />
                    <span>{navLabel(item, t)}</span>
                  </NavLink>
                  )
                })}
              </div>
            </div>
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
            aria-label={t('accessibility.toggleMenu') as string}
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
        <h1 className="font-display text-display-sm text-ink-1 leading-tight">{title}</h1>
        {description && <p className="mt-1.5 text-sm text-ink-2">{description}</p>}
      </div>
      {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
    </div>
  )
}
