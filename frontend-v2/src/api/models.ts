import { apiClient } from './client'
import type { Group, GroupModelCatalog } from '@/types'

/** List groups visible to the current user. */
export async function getUserGroups(): Promise<Group[]> {
  const { data } = await apiClient.get<Group[]>('/groups')
  return data
}

/** Model catalog for a single group. */
export async function getGroupModelCatalog(groupId: number): Promise<GroupModelCatalog> {
  const { data } = await apiClient.get<GroupModelCatalog>(`/groups/${groupId}/models`)
  return data
}

export const modelsAPI = { getUserGroups, getGroupModelCatalog }
export default modelsAPI
