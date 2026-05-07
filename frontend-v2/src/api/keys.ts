import { apiClient } from './client'
import type { ApiKey, PaginatedResponse, CreateApiKeyRequest, UpdateApiKeyRequest } from '@/types'

export async function listKeys(
  page = 1,
  pageSize = 20,
  filters?: { search?: string; status?: string; group_id?: number | string }
) {
  const { data } = await apiClient.get<PaginatedResponse<ApiKey>>('/keys', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function createKey(payload: CreateApiKeyRequest) {
  const { data } = await apiClient.post<ApiKey>('/keys', payload)
  return data
}

export async function updateKey(id: number, payload: UpdateApiKeyRequest) {
  const { data } = await apiClient.put<ApiKey>(`/keys/${id}`, payload)
  return data
}

export async function deleteKey(id: number) {
  await apiClient.delete(`/keys/${id}`)
}

export const keysAPI = { listKeys, createKey, updateKey, deleteKey }
