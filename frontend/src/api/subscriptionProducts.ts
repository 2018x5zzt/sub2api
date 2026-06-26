/**
 * User-facing shared subscription product API.
 */

import { apiClient } from './client'
import type { ActiveSubscriptionProduct, SubscriptionProductSummary } from '@/types'

export async function getActive(): Promise<ActiveSubscriptionProduct[]> {
  const { data } = await apiClient.get<ActiveSubscriptionProduct[]>('/subscription-products/active')
  return data
}

export async function getSummary(): Promise<SubscriptionProductSummary> {
  const { data } = await apiClient.get<SubscriptionProductSummary>('/subscription-products/summary')
  return data
}

export async function getProgress(): Promise<SubscriptionProductSummary> {
  const { data } = await apiClient.get<SubscriptionProductSummary>('/subscription-products/progress')
  return data
}

export const subscriptionProductsAPI = {
  getActive,
  getSummary,
  getProgress
}

export default subscriptionProductsAPI
