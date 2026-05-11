import { apiClient } from '@/api/client'

export async function listAdminChannels(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/channels', { params })
  return data
}

export async function getAdminChannel(id: number | string) {
  const { data } = await apiClient.get(`/admin/channels/${id}`)
  return data
}

export async function createAdminChannel(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/admin/channels', payload)
  return data
}

export async function updateAdminChannel(id: number | string, payload: Record<string, unknown>) {
  const { data } = await apiClient.put(`/admin/channels/${id}`, payload)
  return data
}

export async function deleteAdminChannel(id: number | string) {
  const { data } = await apiClient.delete(`/admin/channels/${id}`)
  return data
}

export async function listChannelPricing(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/channels/model-pricing', { params })
  return data
}

export const adminChannelsAPI = {
  listAdminChannels,
  getAdminChannel,
  createAdminChannel,
  updateAdminChannel,
  deleteAdminChannel,
  listChannelPricing
}
