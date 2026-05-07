import { apiClient } from '@/api/client'
import type { AdminUsageLog, PaginatedResponse, UsageQueryParams, UsageStatsResponse } from '@/types'

export interface SimpleUser {
  id: number
  email: string
  username: string
}

export interface SimpleApiKey {
  id: number
  name: string
  user_id: number
}

export async function listAdminUsage(params: UsageQueryParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<AdminUsageLog>>('/admin/usage/logs', {
    params
  })
  return data
}

export async function getAdminUsageStats(params: { start_date?: string; end_date?: string } = {}) {
  const { data } = await apiClient.get<UsageStatsResponse>('/admin/usage/stats', { params })
  return data
}

export async function searchUsers(keyword: string) {
  const { data } = await apiClient.get<SimpleUser[]>('/admin/usage/search-users', {
    params: { keyword }
  })
  return data
}

export async function searchApiKeys(userId?: number, keyword?: string) {
  const { data } = await apiClient.get<SimpleApiKey[]>('/admin/usage/search-api-keys', {
    params: { user_id: userId, keyword }
  })
  return data
}

export const adminUsageAPI = {
  listAdminUsage,
  getAdminUsageStats,
  searchUsers,
  searchApiKeys
}
