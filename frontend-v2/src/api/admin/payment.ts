import { apiClient } from '@/api/client'

export async function getPaymentConfig() {
  const { data } = await apiClient.get('/admin/payment/config')
  return data
}

export async function updatePaymentConfig(payload: Record<string, unknown>) {
  const { data } = await apiClient.put('/admin/payment/config', payload)
  return data
}

export async function getPaymentDashboard(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/payment/dashboard', { params })
  return data
}

export async function getPaymentOrders(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/payment/orders', { params })
  return data
}

export async function getPaymentOrder(id: number | string) {
  const { data } = await apiClient.get(`/admin/payment/orders/${id}`)
  return data
}

export async function cancelPaymentOrder(id: number | string) {
  const { data } = await apiClient.post(`/admin/payment/orders/${id}/cancel`)
  return data
}

export async function retryPaymentOrder(id: number | string) {
  const { data } = await apiClient.post(`/admin/payment/orders/${id}/retry`)
  return data
}

export async function refundPaymentOrder(id: number | string, payload: Record<string, unknown>) {
  const { data } = await apiClient.post(`/admin/payment/orders/${id}/refund`, payload)
  return data
}

export async function getPaymentProviders() {
  const { data } = await apiClient.get('/admin/payment/providers')
  return data
}

export async function getPaymentPlans() {
  const { data } = await apiClient.get('/admin/payment/plans')
  return data
}

export async function createPaymentPlan(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/admin/payment/plans', payload)
  return data
}

export async function updatePaymentPlan(id: number | string, payload: Record<string, unknown>) {
  const { data } = await apiClient.put(`/admin/payment/plans/${id}`, payload)
  return data
}

export async function deletePaymentPlan(id: number | string) {
  const { data } = await apiClient.delete(`/admin/payment/plans/${id}`)
  return data
}

export const adminPaymentAPI = {
  getPaymentConfig,
  updatePaymentConfig,
  getPaymentDashboard,
  getPaymentOrders,
  getPaymentOrder,
  cancelPaymentOrder,
  retryPaymentOrder,
  refundPaymentOrder,
  getPaymentProviders,
  getPaymentPlans,
  createPaymentPlan,
  updatePaymentPlan,
  deletePaymentPlan
}
