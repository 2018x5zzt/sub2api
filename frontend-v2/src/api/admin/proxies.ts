import { apiClient } from '@/api/client'

export async function listAdminProxies(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/proxies', { params })
  return data
}

export async function listAllAdminProxies(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/proxies/all', { params })
  return data
}

export async function getAdminProxy(id: number | string) {
  const { data } = await apiClient.get(`/admin/proxies/${id}`)
  return data
}

export async function createAdminProxy(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/admin/proxies', payload)
  return data
}

export async function updateAdminProxy(id: number | string, payload: Record<string, unknown>) {
  const { data } = await apiClient.put(`/admin/proxies/${id}`, payload)
  return data
}

export async function deleteAdminProxy(id: number | string) {
  const { data } = await apiClient.delete(`/admin/proxies/${id}`)
  return data
}

export async function testAdminProxy(id: number | string) {
  const { data } = await apiClient.post(`/admin/proxies/${id}/test`)
  return data
}

export async function qualityCheckAdminProxy(id: number | string) {
  const { data } = await apiClient.post(`/admin/proxies/${id}/quality-check`)
  return data
}

export async function getAdminProxyStats(id: number | string) {
  const { data } = await apiClient.get(`/admin/proxies/${id}/stats`)
  return data
}

export async function listAdminProxyAccounts(id: number | string) {
  const { data } = await apiClient.get(`/admin/proxies/${id}/accounts`)
  return data
}

export const adminProxiesAPI = {
  listAdminProxies,
  listAllAdminProxies,
  getAdminProxy,
  createAdminProxy,
  updateAdminProxy,
  deleteAdminProxy,
  testAdminProxy,
  qualityCheckAdminProxy,
  getAdminProxyStats,
  listAdminProxyAccounts
}
