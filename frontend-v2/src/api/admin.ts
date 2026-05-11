import { apiClient } from './client'
import type { AdminUser, PaginatedResponse, DashboardStats, UpdateUserRequest } from '@/types'

export async function getDashboard() {
  const { data } = await apiClient.get<DashboardStats>('/admin/dashboard')
  return data
}

export async function listUsers(page = 1, pageSize = 20, filters?: { search?: string; role?: string; status?: string }) {
  const { data } = await apiClient.get<PaginatedResponse<AdminUser>>('/admin/users', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function updateUser(id: number, payload: UpdateUserRequest) {
  const { data } = await apiClient.put<AdminUser>(`/admin/users/${id}`, payload)
  return data
}

export async function deleteUser(id: number) {
  await apiClient.delete(`/admin/users/${id}`)
}

export const adminAPI = { getDashboard, listUsers, updateUser, deleteUser }
