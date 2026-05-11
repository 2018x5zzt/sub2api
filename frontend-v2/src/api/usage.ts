import { apiClient } from './client'
import type { UsageLog, PaginatedResponse, UsageQueryParams, ModelStat, TrendDataPoint } from '@/types'

export interface UserDashboardStats {
  total_api_keys: number
  active_api_keys: number
  total_requests: number
  total_input_tokens: number
  total_output_tokens: number
  total_cache_creation_tokens: number
  total_cache_read_tokens: number
  total_tokens: number
  total_cost: number
  total_actual_cost: number
  today_requests: number
  today_input_tokens: number
  today_output_tokens: number
  today_cache_creation_tokens: number
  today_cache_read_tokens: number
  today_tokens: number
  today_cost: number
  today_actual_cost: number
  average_duration_ms: number
  rpm: number
  tpm: number
}

export async function getUserDashboard() {
  const { data } = await apiClient.get<UserDashboardStats>('/usage/dashboard')
  return data
}

export async function listUsage(params: UsageQueryParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<UsageLog>>('/usage/logs', { params })
  return data
}

export interface TrendResponse {
  trend: TrendDataPoint[]
  start_date: string
  end_date: string
  granularity: string
}

export async function getUserTrend(params: { start_date?: string; end_date?: string; granularity?: 'day' | 'hour' } = {}) {
  const { data } = await apiClient.get<TrendResponse>('/usage/trend', { params })
  return data
}

export async function getUserModelStats(params: { start_date?: string; end_date?: string } = {}) {
  const { data } = await apiClient.get<{ models: ModelStat[] }>('/usage/models', { params })
  return data
}

export const usageAPI = { getUserDashboard, listUsage, getUserTrend, getUserModelStats }
