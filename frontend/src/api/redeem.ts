/**
 * Redeem code API endpoints
 * Handles redeem code redemption for users
 */

import { apiClient } from './client'
import type { BenefitLeaderboard, PromoCodeScene, RedeemCodeRequest } from '@/types'

export interface RedeemHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_at: string
  created_at: string
  // Notes from admin for admin_balance/admin_concurrency types
  notes?: string
  // Subscription-specific fields
  group_id?: number
  validity_days?: number
  group?: {
    id: number
    name: string
  }
}

/**
 * Redeem a code
 * @param code - Redeem code string
 * @returns Redemption result with updated balance or concurrency
 */
export async function redeem(code: string): Promise<{
  message: string
  type: string
  value: number
  fixed_value?: number
  random_value?: number
  total_value?: number
  scene?: PromoCodeScene
  success_message?: string
  leaderboard_enabled?: boolean
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
}> {
  const payload: RedeemCodeRequest = { code }

  const { data } = await apiClient.post<{
    message: string
    type: string
    value: number
    fixed_value?: number
    random_value?: number
    total_value?: number
    scene?: PromoCodeScene
    success_message?: string
    leaderboard_enabled?: boolean
    new_balance?: number
    new_concurrency?: number
    group_name?: string
    validity_days?: number
  }>('/redeem', payload)

  return data
}

export async function getBenefitLeaderboard(code: string): Promise<BenefitLeaderboard> {
  const payload: RedeemCodeRequest = { code }
  const { data } = await apiClient.post<BenefitLeaderboard>('/redeem/benefit-leaderboard', payload)
  return data
}

/**
 * Get user's redemption history
 * @returns List of redeemed codes
 */
export async function getHistory(): Promise<RedeemHistoryItem[]> {
  const { data } = await apiClient.get<RedeemHistoryItem[]>('/redeem/history')
  return data
}

export const redeemAPI = {
  redeem,
  getHistory,
  getBenefitLeaderboard
}

export default redeemAPI
