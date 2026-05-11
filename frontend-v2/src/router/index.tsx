import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'
import Landing from '@/pages/Landing'
import DocsPage from '@/pages/Docs'
import LoginPage from '@/pages/auth/Login'
import RegisterPage from '@/pages/auth/Register'
import ForgotPasswordPage from '@/pages/auth/ForgotPassword'
import ResetPasswordPage from '@/pages/auth/ResetPassword'
import EmailVerifyPage from '@/pages/auth/EmailVerify'
import OAuthCallbackPage from '@/pages/auth/OAuthCallback'
import LinuxDoCallbackPage from '@/pages/auth/LinuxDoCallback'
import KeysPage from '@/pages/user/Keys'
import UsagePage from '@/pages/user/Usage'
import ProfilePage from '@/pages/user/Profile'
import RedeemPage from '@/pages/user/Redeem'
import SubscriptionsPage from '@/pages/user/Subscriptions'
import ModelHubPage from '@/pages/user/ModelHub'
import UserDashboard from '@/pages/user/Dashboard'
import AffiliatePage from '@/pages/user/Affiliate'
import AvailableChannelsPage from '@/pages/user/AvailableChannels'
import ChannelStatusPage from '@/pages/user/ChannelStatus'
import PurchasePage from '@/pages/user/Purchase'
import OrdersPage from '@/pages/user/Orders'
import PaymentQRCodePage from '@/pages/user/PaymentQRCode'
import PaymentResultPage from '@/pages/user/PaymentResult'
import StripePaymentPage from '@/pages/user/StripePayment'
import StripePopupPage from '@/pages/user/StripePopup'
import ImageStudioPage from '@/pages/user/ImageStudio'
import XlabOAuthConsentPage from '@/pages/auth/XlabOAuthConsent'
import AdminDashboard from '@/pages/admin/Dashboard'
import AdminUsersPage from '@/pages/admin/Users'
import AdminGroupsPage from '@/pages/admin/Groups'
import AdminAccountsPage from '@/pages/admin/Accounts'
import AdminUsagePage from '@/pages/admin/Usage'
import AdminSettingsPage from '@/pages/admin/Settings'
import AdminAnnouncementsPage from '@/pages/admin/Announcements'
import AdminRedeemCodesPage from '@/pages/admin/RedeemCodes'
import AdminPromoCodesPage from '@/pages/admin/PromoCodes'
import AdminSubscriptionsPage from '@/pages/admin/Subscriptions'
import AdminBackupPage from '@/pages/admin/Backup'
import AdminOpsDashboardPage from '@/pages/admin/OpsDashboard'
import AdminChannelPricingPage from '@/pages/admin/ChannelPricing'
import AdminChannelMonitorPage from '@/pages/admin/ChannelMonitor'
import AdminProxiesPage from '@/pages/admin/Proxies'
import AdminPaymentDashboardPage from '@/pages/admin/PaymentDashboard'
import AdminPaymentOrdersPage from '@/pages/admin/PaymentOrders'
import AdminPaymentPlansPage from '@/pages/admin/PaymentPlans'
import AdminAffiliateRecordsPage from '@/pages/admin/AffiliateRecords'
import ParityPlaceholder from '@/pages/ParityPlaceholder'
import NotFoundPage from '@/pages/NotFound'
import { ConsoleLayout } from '@/components/layout/ConsoleLayout'
import { RequireAdmin, RequireAuth, RedirectIfAuthed } from './guards'

const router = createBrowserRouter([
  // Public marketing
  { path: '/', element: <Landing /> },
  { path: '/home', element: <Navigate to="/" replace /> },
  { path: '/docs', element: <DocsPage /> },
  { path: '/setup', element: <ParityPlaceholder standalone title="Setup Wizard" legacyPath="/setup" endpoints={['GET /setup/status']} actions={[{ label: 'Login', to: '/login' }]} /> },
  { path: '/key-usage', element: <ParityPlaceholder standalone title="API Key Usage" legacyPath="/key-usage" endpoints={['GET /usage/dashboard/stats', 'GET /usage/dashboard/models']} actions={[{ label: 'Console', to: '/dashboard' }]} /> },

  // Auth
  { path: '/login', element: <RedirectIfAuthed><LoginPage /></RedirectIfAuthed> },
  { path: '/register', element: <RedirectIfAuthed><RegisterPage /></RedirectIfAuthed> },
  { path: '/email-verify', element: <EmailVerifyPage /> },
  { path: '/forgot-password', element: <ForgotPasswordPage /> },
  { path: '/reset-password', element: <ResetPasswordPage /> },
  { path: '/auth/callback', element: <OAuthCallbackPage /> },
  { path: '/auth/linuxdo/callback', element: <LinuxDoCallbackPage /> },
  { path: '/auth/wechat/callback', element: <OAuthCallbackPage /> },
  { path: '/auth/wechat/payment/callback', element: <OAuthCallbackPage /> },
  { path: '/auth/oidc/callback', element: <OAuthCallbackPage /> },
  { path: '/payment/result', element: <PaymentResultPage /> },
  { path: '/payment/stripe', element: <StripePaymentPage /> },
  { path: '/payment/stripe-popup', element: <StripePopupPage /> },

  {
    element: <RequireAuth><ConsoleLayout /></RequireAuth>,
    children: [
      { path: '/dashboard', element: <UserDashboard /> },
      { path: '/keys', element: <KeysPage /> },
      { path: '/usage', element: <UsagePage /> },
      { path: '/models', element: <ModelHubPage /> },
      { path: '/subscriptions', element: <SubscriptionsPage /> },
      { path: '/redeem', element: <RedeemPage /> },
      { path: '/profile', element: <ProfilePage /> },
      { path: '/invite', element: <Navigate to="/affiliate" replace /> },
      { path: '/affiliate', element: <AffiliatePage /> },
      { path: '/available-channels', element: <AvailableChannelsPage /> },
      { path: '/monitor', element: <ChannelStatusPage /> },
      { path: '/image-studio', element: <ImageStudioPage /> },
      { path: '/purchase', element: <PurchasePage /> },
      { path: '/orders', element: <OrdersPage /> },
      { path: '/payment/qrcode', element: <PaymentQRCodePage /> },
      { path: '/custom/:id', element: <ParityPlaceholder title="Custom Page" legacyPath="/custom/:id" endpoints={['GET /settings/public']} /> },
      { path: '/oauth/consent', element: <XlabOAuthConsentPage /> }
    ]
  },

  {
    element: <RequireAdmin><ConsoleLayout admin /></RequireAdmin>,
    children: [
      { path: '/admin', element: <Navigate to="/admin/dashboard" replace /> },
      { path: '/admin/dashboard', element: <AdminDashboard /> },
      { path: '/admin/users', element: <AdminUsersPage /> },
      { path: '/admin/groups', element: <AdminGroupsPage /> },
      { path: '/admin/accounts', element: <AdminAccountsPage /> },
      { path: '/admin/usage', element: <AdminUsagePage /> },
      { path: '/admin/announcements', element: <AdminAnnouncementsPage /> },
      { path: '/admin/redeem', element: <AdminRedeemCodesPage /> },
      { path: '/admin/promo-codes', element: <AdminPromoCodesPage /> },
      { path: '/admin/subscriptions', element: <AdminSubscriptionsPage /> },
      { path: '/admin/subscription-products', element: <Navigate to="/admin/subscriptions" replace /> },
      { path: '/admin/subscription-product-config', element: <ParityPlaceholder title="Subscription Product Config" legacyPath="/admin/subscription-product-config" endpoints={['GET /admin/subscription-products', 'GET /admin/product-subscriptions']} /> },
      { path: '/admin/backup', element: <AdminBackupPage /> },
      { path: '/admin/settings', element: <AdminSettingsPage /> },
      { path: '/admin/ops', element: <AdminOpsDashboardPage /> },
      { path: '/admin/channels', element: <Navigate to="/admin/channels/pricing" replace /> },
      { path: '/admin/channels/pricing', element: <AdminChannelPricingPage /> },
      { path: '/admin/channels/monitor', element: <AdminChannelMonitorPage /> },
      { path: '/admin/proxies', element: <AdminProxiesPage /> },
      { path: '/admin/invites', element: <Navigate to="/admin/users" replace /> },
      { path: '/admin/affiliates', element: <Navigate to="/admin/affiliates/invites" replace /> },
      { path: '/admin/affiliates/invites', element: <AdminAffiliateRecordsPage type="invites" /> },
      { path: '/admin/affiliates/rebates', element: <AdminAffiliateRecordsPage type="rebates" /> },
      { path: '/admin/affiliates/transfers', element: <AdminAffiliateRecordsPage type="transfers" /> },
      { path: '/admin/orders/dashboard', element: <AdminPaymentDashboardPage /> },
      { path: '/admin/orders', element: <AdminPaymentOrdersPage /> },
      { path: '/admin/orders/plans', element: <AdminPaymentPlansPage /> }
    ]
  },

  { path: '*', element: <NotFoundPage /> }
])

export function AppRouter() {
  return <RouterProvider router={router} />
}
