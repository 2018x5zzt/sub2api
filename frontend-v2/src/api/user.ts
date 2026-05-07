import { apiClient } from './client'
import type { User, ChangePasswordRequest } from '@/types'

export async function getProfile() {
  const { data } = await apiClient.get<User>('/user/profile')
  return data
}

export async function updateProfile(payload: { username?: string }) {
  const { data } = await apiClient.put<User>('/user', payload)
  return data
}

export async function changePassword(oldPassword: string, newPassword: string) {
  const payload: ChangePasswordRequest = { old_password: oldPassword, new_password: newPassword }
  const { data } = await apiClient.put<{ message: string }>('/user/password', payload)
  return data
}

export const userAPI = { getProfile, updateProfile, changePassword }
