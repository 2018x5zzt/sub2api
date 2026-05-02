import { apiClient } from '../client'
import type {
  AdminSubscriptionProduct,
  AdminSubscriptionProductBinding,
  AdminUserProductSubscription,
  AssignProductSubscriptionRequest,
  CreateSubscriptionProductRequest,
  SyncSubscriptionProductBindingRequest,
  UpdateSubscriptionProductRequest
} from '@/types'

export async function list(): Promise<AdminSubscriptionProduct[]> {
  const { data } = await apiClient.get<AdminSubscriptionProduct[]>('/admin/subscription-products')
  return data
}

export async function create(payload: CreateSubscriptionProductRequest): Promise<AdminSubscriptionProduct> {
  const { data } = await apiClient.post<AdminSubscriptionProduct>('/admin/subscription-products', payload)
  return data
}

export async function update(id: number, payload: UpdateSubscriptionProductRequest): Promise<AdminSubscriptionProduct> {
  const { data } = await apiClient.put<AdminSubscriptionProduct>(`/admin/subscription-products/${id}`, payload)
  return data
}

export async function listBindings(id: number): Promise<AdminSubscriptionProductBinding[]> {
  const { data } = await apiClient.get<AdminSubscriptionProductBinding[]>(`/admin/subscription-products/${id}/bindings`)
  return data
}

export async function syncBindings(
  id: number,
  bindings: SyncSubscriptionProductBindingRequest[]
): Promise<AdminSubscriptionProductBinding[]> {
  const { data } = await apiClient.put<AdminSubscriptionProductBinding[]>(`/admin/subscription-products/${id}/bindings`, {
    bindings
  })
  return data
}

export async function listSubscriptions(id: number): Promise<AdminUserProductSubscription[]> {
  const { data } = await apiClient.get<AdminUserProductSubscription[]>(`/admin/subscription-products/${id}/subscriptions`)
  return data
}

export async function assign(
  id: number,
  payload: AssignProductSubscriptionRequest
): Promise<{ subscription: AdminUserProductSubscription; reused: boolean }> {
  const { data } = await apiClient.post<{ subscription: AdminUserProductSubscription; reused: boolean }>(
    `/admin/subscription-products/${id}/assign`,
    payload
  )
  return data
}

export const subscriptionProductsAPI = {
  list,
  create,
  update,
  listBindings,
  syncBindings,
  listSubscriptions,
  assign
}

export default subscriptionProductsAPI
