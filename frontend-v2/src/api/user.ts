import { apiClient } from './client'
import type { User, ChangePasswordRequest } from '@/types'

export interface UpdateProfilePayload {
  username?: string
  subscription_balance_fallback_enabled?: boolean
  subscription_balance_fallback_limit_usd?: number
  subscription_balance_fallback_group_id?: number | null
}

export async function getProfile() {
  const { data } = await apiClient.get<User>('/user/profile')
  return data
}

export async function updateProfile(payload: UpdateProfilePayload) {
  const { data } = await apiClient.put<User>('/user', payload)
  return data
}

export async function changePassword(oldPassword: string, newPassword: string) {
  const payload: ChangePasswordRequest = { old_password: oldPassword, new_password: newPassword }
  const { data } = await apiClient.put<{ message: string }>('/user/password', payload)
  return data
}

export const userAPI = { getProfile, updateProfile, changePassword }
