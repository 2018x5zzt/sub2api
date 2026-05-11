import { apiClient } from './client'

export async function getAffiliateSummary() {
  const { data } = await apiClient.get('/user/aff')
  return data
}

export async function transferAffiliateQuota(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/user/aff/transfer', payload)
  return data
}

export const affiliateAPI = {
  getAffiliateSummary,
  transferAffiliateQuota
}

