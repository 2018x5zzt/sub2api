import { apiClient } from '@/api/client'
import type {
  CreatePromoCodeRequest,
  PaginatedResponse,
  PromoCode,
  PromoCodeScene,
  PromoCodeUsage,
  UpdatePromoCodeRequest
} from '@/types'

export interface PromoListFilters {
  scene?: PromoCodeScene
  status?: string
  search?: string
}

export async function listPromoCodes(page = 1, pageSize = 25, filters?: PromoListFilters) {
  const { data } = await apiClient.get<PaginatedResponse<PromoCode>>('/admin/promo-codes', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function createPromoCode(payload: CreatePromoCodeRequest) {
  const { data } = await apiClient.post<PromoCode>('/admin/promo-codes', payload)
  return data
}

export async function updatePromoCode(id: number, payload: UpdatePromoCodeRequest) {
  const { data } = await apiClient.put<PromoCode>(`/admin/promo-codes/${id}`, payload)
  return data
}

export async function deletePromoCode(id: number) {
  await apiClient.delete(`/admin/promo-codes/${id}`)
}

export async function getPromoUsages(id: number, page = 1, pageSize = 20) {
  const { data } = await apiClient.get<PaginatedResponse<PromoCodeUsage>>(
    `/admin/promo-codes/${id}/usages`,
    { params: { page, page_size: pageSize } }
  )
  return data
}

export const adminPromoAPI = {
  listPromoCodes,
  createPromoCode,
  updatePromoCode,
  deletePromoCode,
  getPromoUsages
}
