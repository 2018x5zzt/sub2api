import { apiClient } from '@/api/client'
import type {
  Account,
  AccountPlatform,
  PaginatedResponse,
  UpdateAccountRequest
} from '@/types'

export async function listAccounts(
  page = 1,
  pageSize = 20,
  filters?: { platform?: AccountPlatform; status?: string; search?: string; group_id?: number }
) {
  const { data } = await apiClient.get<PaginatedResponse<Account>>('/admin/accounts', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function getAccount(id: number) {
  const { data } = await apiClient.get<Account>(`/admin/accounts/${id}`)
  return data
}

export async function updateAccount(id: number, payload: UpdateAccountRequest) {
  const { data } = await apiClient.put<Account>(`/admin/accounts/${id}`, payload)
  return data
}

export async function deleteAccount(id: number) {
  await apiClient.delete(`/admin/accounts/${id}`)
}

export async function toggleAccountStatus(id: number, status: 'active' | 'inactive') {
  const { data } = await apiClient.put<Account>(`/admin/accounts/${id}/status`, { status })
  return data
}

export async function setAccountSchedulable(id: number, schedulable: boolean) {
  const { data } = await apiClient.put<Account>(`/admin/accounts/${id}/schedulable`, { schedulable })
  return data
}

export async function clearAccountError(id: number) {
  const { data } = await apiClient.post<Account>(`/admin/accounts/${id}/clear-error`)
  return data
}

export async function clearAccountRateLimit(id: number) {
  const { data } = await apiClient.post<Account>(`/admin/accounts/${id}/clear-rate-limit`)
  return data
}

export const adminAccountsAPI = {
  listAccounts,
  getAccount,
  updateAccount,
  deleteAccount,
  toggleAccountStatus,
  setAccountSchedulable,
  clearAccountError,
  clearAccountRateLimit
}
