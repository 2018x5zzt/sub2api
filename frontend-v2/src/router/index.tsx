import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'
import type { ReactElement } from 'react'
import Landing from '@/pages/Landing'
import DocsPage from '@/pages/Docs'
import KeyUsagePage from '@/pages/KeyUsage'
import SetupPage from '@/pages/Setup'
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
import CustomPage from '@/pages/user/CustomPage'
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
import AdminSubscriptionProductConfigPage from '@/pages/admin/SubscriptionProductConfig'
import AdminBackupPage from '@/pages/admin/Backup'
import AdminOpsDashboardPage from '@/pages/admin/OpsDashboard'
import AdminChannelPricingPage from '@/pages/admin/ChannelPricing'
import AdminChannelMonitorPage from '@/pages/admin/ChannelMonitor'
import AdminProxiesPage from '@/pages/admin/Proxies'
import AdminPaymentDashboardPage from '@/pages/admin/PaymentDashboard'
import AdminPaymentOrdersPage from '@/pages/admin/PaymentOrders'
import AdminPaymentPlansPage from '@/pages/admin/PaymentPlans'
import AdminAffiliateRecordsPage from '@/pages/admin/AffiliateRecords'
import NotFoundPage from '@/pages/NotFound'
import { ConsoleLayout } from '@/components/layout/ConsoleLayout'
import { BackendModePublicGate, RequireAdmin, RequireAuth, RedirectIfAuthed } from './guards'

const publicGate = (element: ReactElement) => (
  <BackendModePublicGate>{element}</BackendModePublicGate>
)

const router = createBrowserRouter([
  // Public marketing
  { path: '/', element: publicGate(<Landing />) },
  { path: '/home', element: publicGate(<Navigate to="/" replace />) },
  { path: '/docs', element: publicGate(<DocsPage />) },
  { path: '/setup', element: publicGate(<SetupPage />) },
  { path: '/key-usage', element: publicGate(<KeyUsagePage />) },

  // Auth
  { path: '/login', element: <RedirectIfAuthed><LoginPage /></RedirectIfAuthed> },
  { path: '/register', element: publicGate(<RedirectIfAuthed><RegisterPage /></RedirectIfAuthed>) },
  { path: '/email-verify', element: publicGate(<EmailVerifyPage />) },
  { path: '/forgot-password', element: publicGate(<ForgotPasswordPage />) },
  { path: '/reset-password', element: publicGate(<ResetPasswordPage />) },
  { path: '/auth/callback', element: publicGate(<OAuthCallbackPage />) },
  { path: '/auth/linuxdo/callback', element: publicGate(<LinuxDoCallbackPage />) },
  { path: '/auth/wechat/callback', element: publicGate(<OAuthCallbackPage />) },
  { path: '/auth/wechat/payment/callback', element: publicGate(<OAuthCallbackPage />) },
  { path: '/auth/oidc/callback', element: publicGate(<OAuthCallbackPage />) },
  { path: '/payment/result', element: publicGate(<PaymentResultPage />) },
  { path: '/payment/stripe', element: publicGate(<StripePaymentPage />) },
  { path: '/payment/stripe-popup', element: publicGate(<StripePopupPage />) },

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
      { path: '/custom/:id', element: <CustomPage /> },
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
      { path: '/admin/subscription-product-config', element: <AdminSubscriptionProductConfigPage /> },
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
