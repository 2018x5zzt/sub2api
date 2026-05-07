import { apiClient } from '@/api/client'
import type {
  AssignSubscriptionRequest,
  ExtendSubscriptionRequest,
  PaginatedResponse,
  UserSubscription
} from '@/types'

export interface AdminSubscriptionFilters {
  user_id?: number
  group_id?: number
  status?: 'active' | 'expired' | 'revoked'
  search?: string
}

export async function listAdminSubscriptions(
  page = 1,
  pageSize = 25,
  filters?: AdminSubscriptionFilters
) {
  const { data } = await apiClient.get<PaginatedResponse<UserSubscription>>(
    '/admin/subscriptions',
    { params: { page, page_size: pageSize, ...filters } }
  )
  return data
}

export async function assignSubscription(payload: AssignSubscriptionRequest) {
  const { data } = await apiClient.post<UserSubscription>('/admin/subscriptions/assign', payload)
  return data
}

export async function extendSubscription(id: number, payload: ExtendSubscriptionRequest) {
  const { data } = await apiClient.post<UserSubscription>(
    `/admin/subscriptions/${id}/extend`,
    payload
  )
  return data
}

export async function revokeSubscription(id: number) {
  const { data } = await apiClient.post<{ message: string }>(`/admin/subscriptions/${id}/revoke`)
  return data
}

export async function resetSubscriptionQuota(id: number, kind: 'daily' | 'weekly' | 'monthly') {
  const { data } = await apiClient.post<UserSubscription>(
    `/admin/subscriptions/${id}/reset-quota`,
    { kind }
  )
  return data
}

export const adminSubscriptionsAPI = {
  listAdminSubscriptions,
  assignSubscription,
  extendSubscription,
  revokeSubscription,
  resetSubscriptionQuota
}
