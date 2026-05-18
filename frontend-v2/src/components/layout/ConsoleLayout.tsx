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
  Link as LinkIcon,
  BookOpen,
  ChevronsLeft,
  ChevronsRight
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { LocaleSwitcher } from './LocaleSwitcher'
import { cn } from '@/lib/cn'
import { resolveDocLink } from '@/lib/docs'
import type { CustomMenuItem, PublicSettings } from '@/types'

interface NavItem {
  to: string
  labelKey?: string
  label?: string
  Icon: typeof LayoutDashboard
  hideInSimpleMode?: boolean
  featureFlag?: (settings: PublicSettings | null) => boolean
}

const SIDEBAR_COLLAPSED_KEY = 'xlabapi:console-sidebar-collapsed'

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
  { to: '/admin/keys', labelKey: 'nav.apiKeys', Icon: KeyRound },
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

function userNavSections(items: NavItem[]) {
  const core = new Set(['/dashboard', '/keys', '/models', '/usage', '/available-channels', '/monitor', '/profile'])
  const wallet = new Set(['/subscriptions', '/purchase', '/orders', '/redeem', '/affiliate'])
  const coreItems = items.filter((item) => core.has(item.to))
  const walletItems = items.filter((item) => wallet.has(item.to))
  const otherItems = items.filter((item) => !core.has(item.to) && !wallet.has(item.to))

  return [
    { label: '核心功能', items: coreItems },
    { label: '钱包 & 活动', items: walletItems },
    { label: '扩展功能', items: otherItems }
  ].filter((section) => section.items.length > 0)
}

export function ConsoleLayout({ admin, children }: { admin?: boolean; children?: ReactNode }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const logout = useAuthStore((s) => s.logout)
  const runMode = useAuthStore((s) => s.runMode)
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const siteName = publicSettings?.site_name || 'XlabAPI'
  const siteLogo = publicSettings?.site_logo || '/logo.png'
  const docLink = resolveDocLink(publicSettings?.doc_url)

  const [mobileOpen, setMobileOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    if (typeof window === 'undefined') return false
    return window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true'
  })
  const userCustomNav = customNavItems(publicSettings?.custom_menu_items ?? [], 'user')
  const items = publicSettings?.backend_mode_enabled ? [] : visibleNavItems([...userNav, ...userCustomNav], publicSettings, runMode)
  const showAdminItems = isAdmin || admin
  const adminItems = visibleNavItems(adminNav, publicSettings, runMode)
  const collapsed = sidebarCollapsed && !mobileOpen
  const version = publicSettings?.version

  useEffect(() => {
    setMobileOpen(false)
  }, [location.pathname])

  useEffect(() => {
    window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(sidebarCollapsed))
  }, [sidebarCollapsed])

  async function onLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  function renderSectionLabel(label: string) {
    if (collapsed) {
      return <div className="mx-auto my-3 h-px w-8 bg-line-2" aria-hidden />
    }
    return (
      <div className="px-3 pb-2 pt-3 text-[11px] font-mono uppercase tracking-[0.14em] text-ink-3">
        {label}
      </div>
    )
  }

  function renderNavLink(item: NavItem) {
    const { to, Icon } = item
    const label = navLabel(item, t)

    return (
      <NavLink
        key={to}
        to={to}
        title={collapsed ? '展开侧边栏' : undefined}
        className={({ isActive }) =>
          cn(
            'flex rounded-xl text-sm transition-colors',
            collapsed ? 'h-11 items-center justify-center px-0' : 'items-center gap-2.5 px-3 py-2',
            isActive
              ? 'bg-orange-soft text-orange font-medium shadow-sm'
              : 'text-ink-2 hover:bg-bg-3'
          )
        }
        end={to === '/admin/dashboard'}
      >
        <Icon className={cn('shrink-0', collapsed ? 'h-5 w-5' : 'h-4 w-4')} />
        {!collapsed && <span className="truncate">{label}</span>}
      </NavLink>
    )
  }

  return (
    <div className="min-h-screen flex bg-bg-0">
      {/* Sidebar */}
      <aside
        className={cn(
          'fixed lg:static inset-y-0 left-0 z-40 w-60 transform transition-[width,transform] duration-200 lg:translate-x-0',
          'bg-bg-1 border-r border-line-2 flex flex-col',
          collapsed ? 'lg:w-[76px]' : 'lg:w-60',
          mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        )}
      >
        <Link
          to="/"
          className={cn(
            'h-16 flex items-center border-b border-line-2',
            collapsed ? 'justify-center px-0' : 'gap-3 px-5'
          )}
          title={collapsed ? '展开侧边栏' : undefined}
        >
          <div className="h-9 w-9 rounded-xl bg-bg-3 border border-line-2 flex items-center justify-center overflow-hidden">
            <img src={siteLogo} alt="" className="h-full w-full object-contain" onError={(e) => { e.currentTarget.style.display = 'none' }} />
          </div>
          {!collapsed && (
            <span className="min-w-0">
              <span className="block truncate font-semibold text-ink-1">{siteName}</span>
              {version && <span className="block truncate text-xs text-ink-3">v{version}</span>}
            </span>
          )}
        </Link>

        <nav className={cn('flex-1 overflow-y-auto py-4', collapsed ? 'px-2 space-y-1' : 'px-3 space-y-1')}>
          {userNavSections(items).map((section) => (
            <div key={section.label}>
              {renderSectionLabel(section.label)}
              <div className="space-y-1">
                {section.items.map(renderNavLink)}
              </div>
            </div>
          ))}

          {showAdminItems && (
            <div className={cn(collapsed ? 'mt-2' : 'mt-4 border-t border-line-2 pt-1')}>
              {renderSectionLabel(t('layout.adminBanner') as string)}
              <div className="space-y-0.5">
                {adminItems.map(renderNavLink)}
              </div>
            </div>
          )}
        </nav>

        <div className={cn('border-t border-line-2 py-3', collapsed ? 'px-2 space-y-2' : 'px-3 space-y-1')}>
          {!collapsed && (
            <div className="px-3 py-2">
              <div className="text-xs text-ink-3 truncate">{user?.email}</div>
              <div className="text-xs text-ink-3 mt-0.5">
                {t('common.balance')}: <span className="font-mono text-ink-2">${user?.balance?.toFixed(4) ?? '0.0000'}</span>
              </div>
            </div>
          )}
          <button
            onClick={onLogout}
            title={collapsed ? '展开侧边栏' : undefined}
            className={cn(
              'w-full flex rounded-xl text-sm text-ink-2 transition-colors hover:bg-bg-3',
              collapsed ? 'h-11 items-center justify-center px-0' : 'items-center gap-2 px-3 py-2'
            )}
          >
            <LogOut className="h-4 w-4" />
            {!collapsed && <span>{t('common.logout')}</span>}
          </button>
          <button
            type="button"
            onClick={() => setSidebarCollapsed((value) => !value)}
            className={cn(
              'w-full flex rounded-xl text-sm text-ink-2 transition-colors hover:bg-bg-3',
              collapsed ? 'h-11 items-center justify-center px-0' : 'items-center gap-2 px-3 py-2'
            )}
            aria-label={collapsed ? '展开侧边栏' : '收起侧边栏'}
            title={collapsed ? '展开侧边栏' : undefined}
          >
            {collapsed ? <ChevronsRight className="h-4 w-4" /> : <ChevronsLeft className="h-4 w-4" />}
            {!collapsed && <span>收起</span>}
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
            aria-label={collapsed ? '展开侧边栏' : '收起侧边栏'}
          >
            {mobileOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </button>
          <div className="flex-1" />
          <a
            href={docLink.href}
            target={docLink.target}
            rel={docLink.rel}
            className="mr-2 inline-flex h-10 items-center gap-2 rounded-full border border-line-2 bg-bg-1 px-3 text-sm text-ink-2 transition-colors hover:bg-bg-3 hover:text-ink-1 sm:px-4"
            aria-label={collapsed ? '展开侧边栏' : '收起侧边栏'}
          >
            <BookOpen className="h-4 w-4" />
            <span className="hidden sm:inline">{t('nav.docs')}</span>
          </a>
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
