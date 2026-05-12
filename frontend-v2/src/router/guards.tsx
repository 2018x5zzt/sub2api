import { Navigate, useLocation } from 'react-router-dom'
import type { ReactElement } from 'react'
import { useAuthStore } from '@/stores/auth'
import { FullPageSpinner } from '@/components/ui/Spinner'

const BACKEND_MODE_ALLOWED_PUBLIC_PATHS = ['/login', '/key-usage', '/setup', '/payment/result']
const BACKEND_MODE_CALLBACK_PATHS = [
  '/auth/callback',
  '/auth/linuxdo/callback',
  '/auth/oidc/callback',
  '/auth/wechat/callback',
  '/auth/wechat/payment/callback'
]
const SIMPLE_MODE_RESTRICTED_PATHS = [
  '/admin/groups',
  '/admin/subscriptions',
  '/admin/subscription-products',
  '/admin/subscription-product-config',
  '/admin/redeem',
  '/subscriptions',
  '/redeem'
]
const PAYMENT_PATHS = [
  '/purchase',
  '/orders',
  '/payment/qrcode',
  '/payment/stripe',
  '/payment/stripe-popup',
  '/admin/orders/dashboard',
  '/admin/orders',
  '/admin/orders/plans'
]

function isAllowedBackendModePublicPath(path: string) {
  return BACKEND_MODE_ALLOWED_PUBLIC_PATHS.some((allowed) => path === allowed || path.startsWith(`${allowed}/`)) ||
    BACKEND_MODE_CALLBACK_PATHS.includes(path)
}

function isRestrictedInSimpleMode(path: string) {
  return SIMPLE_MODE_RESTRICTED_PATHS.some((restricted) => path === restricted || path.startsWith(`${restricted}/`))
}

function requiresPayment(path: string) {
  return PAYMENT_PATHS.some((paymentPath) => path === paymentPath || path.startsWith(`${paymentPath}/`))
}

function dashboardFor(isAdmin: boolean) {
  return isAdmin ? '/admin/dashboard' : '/dashboard'
}

function loginRedirect(pathname: string) {
  return `/login?redirect=${encodeURIComponent(pathname)}`
}

export function BackendModePublicGate({ children }: { children: ReactElement }) {
  const location = useLocation()
  const initialized = useAuthStore((s) => s.initialized)
  const authed = useAuthStore((s) => s.isAuthenticated())
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const publicSettings = useAuthStore((s) => s.publicSettings)

  if (!initialized) return <FullPageSpinner />
  if (publicSettings?.backend_mode_enabled && !authed && !isAllowedBackendModePublicPath(location.pathname)) {
    return <Navigate to="/login" replace />
  }
  if (publicSettings?.backend_mode_enabled && authed && !isAdmin && !isAllowedBackendModePublicPath(location.pathname)) {
    return <Navigate to="/login" replace />
  }
  return children
}

export function RequireAuth({ children }: { children: ReactElement }) {
  const location = useLocation()
  const initialized = useAuthStore((s) => s.initialized)
  const authed = useAuthStore((s) => s.isAuthenticated())
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const runMode = useAuthStore((s) => s.runMode)
  const publicSettings = useAuthStore((s) => s.publicSettings)

  if (!initialized) return <FullPageSpinner />
  if (!authed) {
    return <Navigate to={loginRedirect(location.pathname)} replace />
  }
  if (publicSettings?.backend_mode_enabled && !isAdmin && !isAllowedBackendModePublicPath(location.pathname)) {
    return <Navigate to="/login" replace />
  }
  if (runMode === 'simple' && isRestrictedInSimpleMode(location.pathname)) {
    return <Navigate to={dashboardFor(isAdmin)} replace />
  }
  if (requiresPayment(location.pathname) && publicSettings?.payment_enabled === false) {
    return <Navigate to={dashboardFor(isAdmin)} replace />
  }
  return children
}

export function RequireAdmin({ children }: { children: ReactElement }) {
  const location = useLocation()
  const initialized = useAuthStore((s) => s.initialized)
  const authed = useAuthStore((s) => s.isAuthenticated())
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const runMode = useAuthStore((s) => s.runMode)
  const publicSettings = useAuthStore((s) => s.publicSettings)

  if (!initialized) return <FullPageSpinner />
  if (!authed) {
    return <Navigate to={loginRedirect(location.pathname)} replace />
  }
  if (!isAdmin) return <Navigate to="/dashboard" replace />
  if (runMode === 'simple' && isRestrictedInSimpleMode(location.pathname)) {
    return <Navigate to="/admin/dashboard" replace />
  }
  if (requiresPayment(location.pathname) && publicSettings?.payment_enabled === false) {
    return <Navigate to="/admin/dashboard" replace />
  }
  return children
}

export function RedirectIfAuthed({ children }: { children: ReactElement }) {
  const initialized = useAuthStore((s) => s.initialized)
  const authed = useAuthStore((s) => s.isAuthenticated())
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const publicSettings = useAuthStore((s) => s.publicSettings)

  if (!initialized) return <FullPageSpinner />
  if (authed) {
    if (publicSettings?.backend_mode_enabled && !isAdmin) return children
    return <Navigate to={dashboardFor(isAdmin)} replace />
  }
  return children
}
