import { apiClient } from '@/api/client'

export async function listAdminChannelMonitors(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/channel-monitors', { params })
  return data
}

export async function getAdminChannelMonitor(id: number | string) {
  const { data } = await apiClient.get(`/admin/channel-monitors/${id}`)
  return data
}

export async function createAdminChannelMonitor(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/admin/channel-monitors', payload)
  return data
}

export async function updateAdminChannelMonitor(id: number | string, payload: Record<string, unknown>) {
  const { data } = await apiClient.put(`/admin/channel-monitors/${id}`, payload)
  return data
}

export async function deleteAdminChannelMonitor(id: number | string) {
  const { data } = await apiClient.delete(`/admin/channel-monitors/${id}`)
  return data
}

export async function runAdminChannelMonitor(id: number | string) {
  const { data } = await apiClient.post(`/admin/channel-monitors/${id}/run`)
  return data
}

export async function listAdminChannelMonitorHistory(id: number | string, params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get(`/admin/channel-monitors/${id}/history`, { params })
  return data
}

export const adminChannelMonitorAPI = {
  listAdminChannelMonitors,
  getAdminChannelMonitor,
  createAdminChannelMonitor,
  updateAdminChannelMonitor,
  deleteAdminChannelMonitor,
  runAdminChannelMonitor,
  listAdminChannelMonitorHistory
}
