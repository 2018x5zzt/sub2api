import { apiClient } from './client'
import type { UserSubscription, SubscriptionProgress } from '@/types'

export interface SubscriptionSummary {
  active_count: number
  subscriptions: Array<{
    id: number
    group_name: string
    status: string
    daily_progress: number | null
    weekly_progress: number | null
    monthly_progress: number | null
    expires_at: string | null
    days_remaining: number | null
  }>
}

export async function getMySubscriptions(): Promise<UserSubscription[]> {
  const { data } = await apiClient.get<UserSubscription[]>('/subscriptions')
  return data
}

export async function getActiveSubscriptions(): Promise<UserSubscription[]> {
  const { data } = await apiClient.get<UserSubscription[]>('/subscriptions/active')
  return data
}

export async function getSubscriptionsProgress(): Promise<SubscriptionProgress[]> {
  const { data } = await apiClient.get<SubscriptionProgress[]>('/subscriptions/progress')
  return data
}

export async function getSubscriptionSummary(): Promise<SubscriptionSummary> {
  const { data } = await apiClient.get<SubscriptionSummary>('/subscriptions/summary')
  return data
}

export const subscriptionsAPI = {
  getMySubscriptions,
  getActiveSubscriptions,
  getSubscriptionsProgress,
  getSubscriptionSummary
}
export default subscriptionsAPI
