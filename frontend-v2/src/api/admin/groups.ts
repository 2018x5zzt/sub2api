import { apiClient } from '@/api/client'
import type {
  AdminGroup,
  CreateGroupRequest,
  GroupPlatform,
  PaginatedResponse,
  UpdateGroupRequest
} from '@/types'

export async function listGroups(
  page = 1,
  pageSize = 20,
  filters?: { platform?: GroupPlatform; search?: string; status?: string }
) {
  const { data } = await apiClient.get<PaginatedResponse<AdminGroup>>('/admin/groups', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function listAllGroups(platform?: GroupPlatform) {
  const { data } = await apiClient.get<AdminGroup[]>('/admin/groups/all', {
    params: platform ? { platform } : {}
  })
  return data
}

export async function createGroup(payload: CreateGroupRequest) {
  const { data } = await apiClient.post<AdminGroup>('/admin/groups', payload)
  return data
}

export async function updateGroup(id: number, payload: UpdateGroupRequest) {
  const { data } = await apiClient.put<AdminGroup>(`/admin/groups/${id}`, payload)
  return data
}

export async function toggleGroupStatus(id: number, status: 'active' | 'inactive') {
  return updateGroup(id, { status })
}

export async function deleteGroup(id: number) {
  await apiClient.delete(`/admin/groups/${id}`)
}

export const adminGroupsAPI = {
  listGroups,
  listAllGroups,
  createGroup,
  updateGroup,
  toggleGroupStatus,
  deleteGroup
}
