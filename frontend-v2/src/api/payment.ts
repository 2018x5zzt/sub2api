import { apiClient } from './client'

export interface PaymentPlan {
  id: number
  name: string
  description?: string
  price: number
  original_price?: number | null
  rate_multiplier?: number | null
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  group_platform?: string | null
}

export interface PaymentChannel {
  id: number
  name: string
  provider: string
  enabled: boolean
}

export interface PaymentOrder {
  id: number
  status: string
  amount: number
  payment_type?: string
  pay_url?: string
  qr_code?: string
  expires_at?: string
}

export async function getPaymentConfig() {
  const { data } = await apiClient.get('/payment/config')
  return data
}

export async function getPaymentPlans() {
  const { data } = await apiClient.get<PaymentPlan[]>('/payment/plans')
  return data
}

export async function getPaymentChannels() {
  const { data } = await apiClient.get<PaymentChannel[]>('/payment/channels')
  return data
}

export async function getPaymentCheckoutInfo() {
  const { data } = await apiClient.get('/payment/checkout-info')
  return data
}

export async function getPaymentLimits() {
  const { data } = await apiClient.get('/payment/limits')
  return data
}

export async function createPaymentOrder(payload: Record<string, unknown>) {
  const { data } = await apiClient.post<PaymentOrder>('/payment/orders', payload)
  return data
}

export async function verifyPublicOrder(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/payment/public/orders/verify', payload)
  return data
}

export async function resolvePublicOrder(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/payment/public/orders/resolve', payload)
  return data
}

export async function getMyOrders(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/payment/orders/my', { params })
  return data
}

export async function getOrder(id: number | string) {
  const { data } = await apiClient.get<PaymentOrder>(`/payment/orders/${id}`)
  return data
}

export async function cancelOrder(id: number | string) {
  const { data } = await apiClient.post(`/payment/orders/${id}/cancel`)
  return data
}

export async function requestOrderRefund(id: number | string, payload: Record<string, unknown>) {
  const { data } = await apiClient.post(`/payment/orders/${id}/refund-request`, payload)
  return data
}

export async function getRefundEligibleProviders() {
  const { data } = await apiClient.get('/payment/orders/refund-eligible-providers')
  return data
}

export const paymentAPI = {
  getPaymentConfig,
  getPaymentPlans,
  getPaymentChannels,
  getPaymentCheckoutInfo,
  getPaymentLimits,
  createPaymentOrder,
  verifyPublicOrder,
  resolvePublicOrder,
  getMyOrders,
  getOrder,
  cancelOrder,
  requestOrderRefund,
  getRefundEligibleProviders
}

