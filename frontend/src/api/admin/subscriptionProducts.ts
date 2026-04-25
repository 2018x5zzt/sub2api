/**
 * Admin shared subscription product API.
 */

import { apiClient } from '../client'
import type {
  AdminSubscriptionProduct,
  AdminSubscriptionProductBinding,
  AdminUserProductSubscription,
  CreateSubscriptionProductRequest,
  ProductGroupBindingInput,
  UpdateSubscriptionProductRequest
} from '@/types'

export async function listProducts(): Promise<AdminSubscriptionProduct[]> {
  const { data } = await apiClient.get<AdminSubscriptionProduct[]>('/admin/subscription-products')
  return data
}

export async function createProduct(
  input: CreateSubscriptionProductRequest
): Promise<AdminSubscriptionProduct> {
  const { data } = await apiClient.post<AdminSubscriptionProduct>(
    '/admin/subscription-products',
    input
  )
  return data
}

export async function updateProduct(
  id: number,
  input: UpdateSubscriptionProductRequest
): Promise<AdminSubscriptionProduct> {
  const { data } = await apiClient.put<AdminSubscriptionProduct>(
    `/admin/subscription-products/${id}`,
    input
  )
  return data
}

export async function syncBindings(
  id: number,
  bindings: ProductGroupBindingInput[]
): Promise<AdminSubscriptionProductBinding[]> {
  const { data } = await apiClient.put<AdminSubscriptionProductBinding[]>(
    `/admin/subscription-products/${id}/bindings`,
    { bindings }
  )
  return data
}

export async function listSubscriptions(id: number): Promise<AdminUserProductSubscription[]> {
  const { data } = await apiClient.get<AdminUserProductSubscription[]>(
    `/admin/subscription-products/${id}/subscriptions`
  )
  return data
}

export const subscriptionProductsAPI = {
  listProducts,
  createProduct,
  updateProduct,
  syncBindings,
  listSubscriptions
}

export default subscriptionProductsAPI
