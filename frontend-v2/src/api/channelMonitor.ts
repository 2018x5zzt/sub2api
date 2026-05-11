import { apiClient } from './client'

export interface ChannelMonitorStatus {
  id: number
  name: string
  enabled: boolean
  provider: string
}

export async function listChannelMonitors(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/channel-monitors', { params })
  return data
}

export async function getChannelMonitorStatus(id: number | string) {
  const { data } = await apiClient.get<ChannelMonitorStatus>(`/channel-monitors/${id}/status`)
  return data
}

export const channelMonitorAPI = {
  listChannelMonitors,
  getChannelMonitorStatus
}

