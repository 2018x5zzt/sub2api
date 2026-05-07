import { apiClient } from '@/api/client'
import type {
  Announcement,
  AnnouncementStatus,
  CreateAnnouncementRequest,
  PaginatedResponse,
  UpdateAnnouncementRequest,
  AnnouncementUserReadStatus
} from '@/types'

export async function listAnnouncements(
  page = 1,
  pageSize = 20,
  filters?: { status?: AnnouncementStatus; search?: string }
) {
  const { data } = await apiClient.get<PaginatedResponse<Announcement>>('/admin/announcements', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function getAnnouncement(id: number) {
  const { data } = await apiClient.get<Announcement>(`/admin/announcements/${id}`)
  return data
}

export async function createAnnouncement(payload: CreateAnnouncementRequest) {
  const { data } = await apiClient.post<Announcement>('/admin/announcements', payload)
  return data
}

export async function updateAnnouncement(id: number, payload: UpdateAnnouncementRequest) {
  const { data } = await apiClient.put<Announcement>(`/admin/announcements/${id}`, payload)
  return data
}

export async function deleteAnnouncement(id: number) {
  await apiClient.delete(`/admin/announcements/${id}`)
}

export async function getAnnouncementReadStatus(id: number) {
  const { data } = await apiClient.get<AnnouncementUserReadStatus[]>(
    `/admin/announcements/${id}/read-status`
  )
  return data
}

export const adminAnnouncementsAPI = {
  listAnnouncements,
  getAnnouncement,
  createAnnouncement,
  updateAnnouncement,
  deleteAnnouncement,
  getAnnouncementReadStatus
}
