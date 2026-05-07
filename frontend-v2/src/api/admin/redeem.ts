import { apiClient } from '@/api/client'
import type {
  GenerateRedeemCodesRequest,
  PaginatedResponse,
  RedeemCode,
  RedeemCodeType
} from '@/types'

export interface RedeemListFilters {
  type?: RedeemCodeType
  status?: string
  search?: string
  group_id?: number
}

export async function listRedeemCodes(page = 1, pageSize = 25, filters?: RedeemListFilters) {
  const { data } = await apiClient.get<PaginatedResponse<RedeemCode>>('/admin/redeem-codes', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function generateRedeemCodes(payload: GenerateRedeemCodesRequest) {
  const { data } = await apiClient.post<{ codes: RedeemCode[]; count: number }>(
    '/admin/redeem-codes/generate',
    payload
  )
  return data
}

export async function expireRedeemCode(id: number) {
  const { data } = await apiClient.post<RedeemCode>(`/admin/redeem-codes/${id}/expire`)
  return data
}

export async function deleteRedeemCode(id: number) {
  await apiClient.delete(`/admin/redeem-codes/${id}`)
}

export async function batchDeleteRedeemCodes(ids: number[]) {
  const { data } = await apiClient.post<{ deleted: number }>('/admin/redeem-codes/batch-delete', {
    ids
  })
  return data
}

export const adminRedeemAPI = {
  listRedeemCodes,
  generateRedeemCodes,
  expireRedeemCode,
  deleteRedeemCode,
  batchDeleteRedeemCodes
}
