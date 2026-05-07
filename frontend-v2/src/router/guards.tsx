import { Navigate, useLocation } from 'react-router-dom'
import type { ReactElement } from 'react'
import { useAuthStore } from '@/stores/auth'
import { FullPageSpinner } from '@/components/ui/Spinner'

export function RequireAuth({ children }: { children: ReactElement }) {
  const location = useLocation()
  const initialized = useAuthStore((s) => s.initialized)
  const authed = useAuthStore((s) => s.isAuthenticated())
  if (!initialized) return <FullPageSpinner />
  if (!authed) {
    return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname)}`} replace />
  }
  return children
}

export function RequireAdmin({ children }: { children: ReactElement }) {
  const location = useLocation()
  const initialized = useAuthStore((s) => s.initialized)
  const authed = useAuthStore((s) => s.isAuthenticated())
  const isAdmin = useAuthStore((s) => s.isAdmin())
  if (!initialized) return <FullPageSpinner />
  if (!authed) {
    return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname)}`} replace />
  }
  if (!isAdmin) return <Navigate to="/dashboard" replace />
  return children
}

export function RedirectIfAuthed({ children }: { children: ReactElement }) {
  const initialized = useAuthStore((s) => s.initialized)
  const authed = useAuthStore((s) => s.isAuthenticated())
  if (!initialized) return <FullPageSpinner />
  if (authed) return <Navigate to="/dashboard" replace />
  return children
}
