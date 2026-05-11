import { apiClient } from '@/api/client'

export async function listAffiliateInvites(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/affiliates/invites', { params })
  return data
}

export async function listAffiliateRebates(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/affiliates/rebates', { params })
  return data
}

export async function listAffiliateTransfers(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/affiliates/transfers', { params })
  return data
}

export async function getAffiliateUserOverview(userId: number | string) {
  const { data } = await apiClient.get(`/admin/affiliates/users/${userId}/overview`)
  return data
}

export const adminAffiliateAPI = {
  listAffiliateInvites,
  listAffiliateRebates,
  listAffiliateTransfers,
  getAffiliateUserOverview
}
