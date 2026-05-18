import { apiClient } from '@/api/client'
import type { ApiKey, PaginatedResponse } from '@/types'

export interface AdminUpdateApiKeyGroupResult {
  api_key: ApiKey
  auto_granted_group_access: boolean
  granted_group_id?: number
  granted_group_name?: string
}

export interface AdminUpdateApiKeyGroupPayload {
  group_id: number | null
  reset_rate_limit_usage?: boolean
}

export async function listUserApiKeys(
  userId: number,
  page = 1,
  pageSize = 100,
  filters?: { sort_by?: string; sort_order?: 'asc' | 'desc' }
) {
  const { data } = await apiClient.get<PaginatedResponse<ApiKey>>(`/admin/users/${userId}/api-keys`, {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function updateApiKeyGroup(id: number, payload: AdminUpdateApiKeyGroupPayload) {
  const body = {
    ...payload,
    group_id: payload.group_id === null ? 0 : payload.group_id
  }
  const { data } = await apiClient.put<AdminUpdateApiKeyGroupResult>(`/admin/api-keys/${id}`, body)
  return data
}

export const adminKeysAPI = {
  listUserApiKeys,
  updateApiKeyGroup
}
