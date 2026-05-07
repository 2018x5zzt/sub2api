import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'
import Landing from '@/pages/Landing'
import Console from '@/pages/Console'
import LoginPage from '@/pages/auth/Login'
import RegisterPage from '@/pages/auth/Register'
import ForgotPasswordPage from '@/pages/auth/ForgotPassword'
import KeysPage from '@/pages/user/Keys'
import UsagePage from '@/pages/user/Usage'
import ProfilePage from '@/pages/user/Profile'
import RedeemPage from '@/pages/user/Redeem'
import AdminDashboard from '@/pages/admin/Dashboard'
import AdminUsersPage from '@/pages/admin/Users'
import { PlaceholderPage } from '@/pages/admin/Placeholder'
import NotFoundPage from '@/pages/NotFound'
import { ConsoleLayout } from '@/components/layout/ConsoleLayout'
import { RequireAdmin, RequireAuth, RedirectIfAuthed } from './guards'

const router = createBrowserRouter([
  // Public marketing — Landing v4 - Plato (verbatim port of variant-d.jsx)
  { path: '/', element: <Landing /> },
  { path: '/home', element: <Navigate to="/" replace /> },

  // Auth
  { path: '/login', element: <RedirectIfAuthed><LoginPage /></RedirectIfAuthed> },
  { path: '/register', element: <RedirectIfAuthed><RegisterPage /></RedirectIfAuthed> },
  { path: '/forgot-password', element: <ForgotPasswordPage /> },

  // Console v4 - Plato (verbatim port of console-v4.jsx). Self-contained: brings
  // its own NavBar, so it does not nest inside ConsoleLayout.
  { path: '/dashboard', element: <RequireAuth><Console /></RequireAuth> },

  // Inner user pages still use the existing console shell with sidebar.
  {
    element: <RequireAuth><ConsoleLayout /></RequireAuth>,
    children: [
      { path: '/keys', element: <KeysPage /> },
      { path: '/usage', element: <UsagePage /> },
      { path: '/redeem', element: <RedeemPage /> },
      { path: '/profile', element: <ProfilePage /> }
    ]
  },

  {
    element: <RequireAdmin><ConsoleLayout admin /></RequireAdmin>,
    children: [
      { path: '/admin', element: <AdminDashboard /> },
      { path: '/admin/users', element: <AdminUsersPage /> },
      { path: '/admin/groups', element: <PlaceholderPage title="Groups" /> },
      { path: '/admin/accounts', element: <PlaceholderPage title="Accounts" /> },
      { path: '/admin/announcements', element: <PlaceholderPage title="Announcements" /> },
      { path: '/admin/usage', element: <PlaceholderPage title="Usage" /> },
      { path: '/admin/redeem', element: <PlaceholderPage title="Redeem Codes" /> },
      { path: '/admin/settings', element: <PlaceholderPage title="Settings" /> }
    ]
  },

  { path: '*', element: <NotFoundPage /> }
])

export function AppRouter() {
  return <RouterProvider router={router} />
}
