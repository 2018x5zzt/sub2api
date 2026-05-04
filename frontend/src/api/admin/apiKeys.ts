/**
 * Admin API Keys API endpoints
 * Handles API key management for administrators
 */

import { apiClient } from '../client'
import type { ApiKey } from '@/types'

export interface UpdateApiKeyGroupResult {
  api_key: ApiKey
  auto_granted_group_access: boolean
  granted_group_id?: number
  granted_group_name?: string
}

export interface UpdateAdminApiKeyGroupPayload {
  group_id: number | null
  reset_rate_limit_usage?: boolean
}

/**
 * Update an API key's group binding
 * @param id - API Key ID
 * @param payload - Group ID (0 to unbind, positive to bind, null/undefined to skip) plus optional reset metadata
 * @returns Updated API key with auto-grant info
 */
export async function updateApiKeyGroup(
  id: number,
  payload: number | null | UpdateAdminApiKeyGroupPayload
): Promise<UpdateApiKeyGroupResult> {
  const body: UpdateAdminApiKeyGroupPayload = payload !== null && typeof payload === 'object'
    ? payload
    : { group_id: payload }
  const { data } = await apiClient.put<UpdateApiKeyGroupResult>(`/admin/api-keys/${id}`, {
    ...body,
    group_id: body.group_id === null ? 0 : body.group_id
  })
  return data
}

export const apiKeysAPI = {
  updateApiKeyGroup
}

export default apiKeysAPI
