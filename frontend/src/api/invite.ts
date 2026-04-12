import { apiClient } from './client'
import type { BasePaginationResponse, InviteRewardRecord, InviteSummary } from '@/types'

async function getSummary(): Promise<InviteSummary> {
  const { data } = await apiClient.get<InviteSummary>('/invite/summary')
  return data
}

async function listRewards(
  page = 1,
  pageSize = 20
): Promise<BasePaginationResponse<InviteRewardRecord>> {
  const { data } = await apiClient.get<BasePaginationResponse<InviteRewardRecord>>(
    '/invite/rewards',
    {
      params: { page, page_size: pageSize }
    }
  )
  return data
}

export const inviteAPI = { getSummary, listRewards }

export default inviteAPI
