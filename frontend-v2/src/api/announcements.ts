import { apiClient } from './client'
import type { UserAnnouncement } from '@/types'

export async function list(unreadOnly = false): Promise<UserAnnouncement[]> {
  const { data } = await apiClient.get<UserAnnouncement[]>('/announcements', {
    params: unreadOnly ? { unread_only: 1 } : {}
  })
  return data
}

export async function markRead(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/announcements/${id}/read`)
  return data
}

export const announcementsAPI = { list, markRead }
export default announcementsAPI
