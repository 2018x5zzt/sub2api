import { apiClient } from './client'
import type { PromoCodeScene, RedeemCodeRequest } from '@/types'

export interface RedeemHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_at: string
  created_at: string
  notes?: string
  group_id?: number
  validity_days?: number
  group?: {
    id: number
    name: string
  }
}

export interface RedeemResult {
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
}

export async function redeem(code: string): Promise<RedeemResult> {
  const payload: RedeemCodeRequest = { code }
  const { data } = await apiClient.post<RedeemResult>('/redeem', payload)
  return data
}

export async function getHistory(): Promise<RedeemHistoryItem[]> {
  const { data } = await apiClient.get<RedeemHistoryItem[]>('/redeem/history')
  return data
}

export const redeemAPI = { redeem, getHistory }
export default redeemAPI
