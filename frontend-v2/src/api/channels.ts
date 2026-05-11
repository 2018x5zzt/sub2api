import { apiClient } from './client'

export interface AvailableChannel {
  id: number
  name: string
  platform: string
  status: string
  group_ids?: number[]
}

export interface AvailableChannelPricing {
  billing_mode?: string
  input_price?: number | null
  output_price?: number | null
  cache_write_price?: number | null
  cache_read_price?: number | null
  image_output_price?: number | null
  per_request_price?: number | null
}

export interface AvailableChannelModel {
  name: string
  platform: string
  pricing?: AvailableChannelPricing | null
}

export interface AvailableChannelGroup {
  id: number
  name: string
  platform: string
  subscription_type?: string
  rate_multiplier?: number
  is_exclusive?: boolean
  pricing_mode?: string
  default_budget_multiplier?: number | null
  dynamic_multiplier_min?: number | null
  dynamic_multiplier_max?: number | null
  dynamic_budget_multiplier?: number | null
  dynamic_budget_matched_multiplier?: number | null
}

export interface AvailableChannelSection {
  platform: string
  groups: AvailableChannelGroup[]
  supported_models: AvailableChannelModel[]
}

export interface AvailableChannelEntry {
  name: string
  description: string
  platforms: AvailableChannelSection[]
}

export async function listAvailableChannels() {
  const { data } = await apiClient.get<AvailableChannelEntry[]>('/channels/available')
  return data
}

export async function getUserGroupRates() {
  const { data } = await apiClient.get('/groups/rates')
  return data
}

export const channelsAPI = {
  listAvailableChannels,
  getUserGroupRates
}
