import { apiClient } from '@/api/client'
import type { PaginatedResponse } from '@/types'

export interface AdminSubscriptionProduct {
  id: number
  code: string
  name: string
  description?: string
  status: 'draft' | 'active' | 'disabled' | string
  product_family?: string
  default_validity_days?: number
  daily_limit_usd?: number
  weekly_limit_usd?: number
  monthly_limit_usd?: number
  sort_order?: number
  created_at?: string
  updated_at?: string
}

export interface AdminSubscriptionProductBinding {
  product_id: number
  group_id: number
  group_name?: string
  debit_multiplier?: number
  status?: 'active' | 'inactive' | string
  sort_order?: number
}

export interface AdminUserProductSubscription {
  id: number
  user_id: number
  product_id: number
  starts_at?: string
  expires_at?: string
  status: 'active' | 'expired' | 'revoked' | string
  daily_usage_usd?: number
  weekly_usage_usd?: number
  monthly_usage_usd?: number
  notes?: string
}

export interface AdminProductSubscriptionListItem extends AdminUserProductSubscription {
  user_email?: string
  user_username?: string
  product_code?: string
  product_name?: string
  daily_limit_usd?: number
  weekly_limit_usd?: number
  monthly_limit_usd?: number
}

export interface SubscriptionProductPayload {
  code?: string
  name?: string
  description?: string
  status?: string
  product_family?: string
  default_validity_days?: number
  daily_limit_usd?: number
  weekly_limit_usd?: number
  monthly_limit_usd?: number
  sort_order?: number
}

export interface SyncSubscriptionProductBindingRequest {
  group_id: number
  debit_multiplier?: number
  status?: string
  sort_order?: number
}

export interface ListUserProductSubscriptionsParams {
  page?: number
  page_size?: number
  search?: string
  product_id?: number | null
  user_id?: number | null
  status?: string | null
  sort_by?: 'expires_at' | 'created_at' | 'daily_usage_usd'
  sort_order?: 'asc' | 'desc'
}

export async function listProducts() {
  const { data } = await apiClient.get<AdminSubscriptionProduct[]>('/admin/subscription-products')
  return data
}

export async function listUserSubscriptions(params: ListUserProductSubscriptionsParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<AdminProductSubscriptionListItem>>(
    '/admin/product-subscriptions',
    { params }
  )
  return data
}

export async function createProduct(payload: SubscriptionProductPayload) {
  const { data } = await apiClient.post<AdminSubscriptionProduct>('/admin/subscription-products', payload)
  return data
}

export async function updateProduct(id: number, payload: SubscriptionProductPayload) {
  const { data } = await apiClient.put<AdminSubscriptionProduct>(`/admin/subscription-products/${id}`, payload)
  return data
}

export async function listBindings(id: number) {
  const { data } = await apiClient.get<AdminSubscriptionProductBinding[]>(`/admin/subscription-products/${id}/bindings`)
  return data
}

export async function syncBindings(id: number, bindings: SyncSubscriptionProductBindingRequest[]) {
  const { data } = await apiClient.put<AdminSubscriptionProductBinding[]>(`/admin/subscription-products/${id}/bindings`, {
    bindings
  })
  return data
}

export async function listProductSubscriptions(id: number) {
  const { data } = await apiClient.get<AdminUserProductSubscription[]>(`/admin/subscription-products/${id}/subscriptions`)
  return data
}

export const adminSubscriptionProductsAPI = {
  listProducts,
  listUserSubscriptions,
  createProduct,
  updateProduct,
  listBindings,
  syncBindings,
  listProductSubscriptions
}
