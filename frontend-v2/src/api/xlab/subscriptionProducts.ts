import { xlabClient } from '../xlabClient'
import type { ActiveSubscriptionProduct, SubscriptionProductSummary } from '@/types'

export async function getActive(): Promise<ActiveSubscriptionProduct[]> {
  const { data } = await xlabClient.get<ActiveSubscriptionProduct[]>('/subscription-products/active')
  return data
}

export async function getSummary(): Promise<SubscriptionProductSummary> {
  const { data } = await xlabClient.get<SubscriptionProductSummary>('/subscription-products/summary')
  return data
}

export async function getProgress(): Promise<SubscriptionProductSummary> {
  const { data } = await xlabClient.get<SubscriptionProductSummary>('/subscription-products/progress')
  return data
}

export const xlabSubscriptionProductsAPI = { getActive, getSummary, getProgress }
