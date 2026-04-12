import { apiClient } from '../client'
import type {
  AdminInviteAction,
  AdminInviteRebindRequest,
  AdminInviteRelationshipRow,
  AdminInviteRecomputeExecuteRequest,
  AdminInviteRecomputePreview,
  AdminInviteRecomputePreviewRequest,
  AdminInviteRewardRow,
  AdminInviteStats,
  AdminManualInviteGrantRequest,
  BasePaginationResponse
} from '@/types'

type RelationshipFilters = {
  search?: string
  inviter_user_id?: number
  invitee_user_id?: number
}

type RewardFilters = {
  search?: string
  reward_type?: string
  target_user_id?: number
}

type ActionFilters = {
  action_type?: string
  target_user_id?: number
  operator_user_id?: number
}

async function getStats(): Promise<AdminInviteStats> {
  const { data } = await apiClient.get<AdminInviteStats>('/admin/invites/stats')
  return data
}

async function listRelationships(
  page = 1,
  pageSize = 20,
  filters: RelationshipFilters = {}
): Promise<BasePaginationResponse<AdminInviteRelationshipRow>> {
  const { data } = await apiClient.get<BasePaginationResponse<AdminInviteRelationshipRow>>(
    '/admin/invites/relationships',
    {
      params: {
        page,
        page_size: pageSize,
        ...filters
      }
    }
  )
  return data
}

async function listRewards(
  page = 1,
  pageSize = 20,
  filters: RewardFilters = {}
): Promise<BasePaginationResponse<AdminInviteRewardRow>> {
  const { data } = await apiClient.get<BasePaginationResponse<AdminInviteRewardRow>>(
    '/admin/invites/rewards',
    {
      params: {
        page,
        page_size: pageSize,
        ...filters
      }
    }
  )
  return data
}

async function listActions(
  page = 1,
  pageSize = 20,
  filters: ActionFilters = {}
): Promise<BasePaginationResponse<AdminInviteAction>> {
  const { data } = await apiClient.get<BasePaginationResponse<AdminInviteAction>>(
    '/admin/invites/actions',
    {
      params: {
        page,
        page_size: pageSize,
        ...filters
      }
    }
  )
  return data
}

async function rebindInviter(payload: AdminInviteRebindRequest): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/invites/rebind', payload)
  return data
}

async function createManualGrant(
  payload: AdminManualInviteGrantRequest
): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(
    '/admin/invites/manual-grants',
    payload
  )
  return data
}

async function previewRecompute(
  payload: AdminInviteRecomputePreviewRequest
): Promise<AdminInviteRecomputePreview> {
  const { data } = await apiClient.post<AdminInviteRecomputePreview>(
    '/admin/invites/recompute/preview',
    payload
  )
  return data
}

async function executeRecompute(
  payload: AdminInviteRecomputeExecuteRequest
): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(
    '/admin/invites/recompute/execute',
    payload
  )
  return data
}

export const adminInvitesAPI = {
  getStats,
  listRelationships,
  listRewards,
  listActions,
  rebindInviter,
  createManualGrant,
  previewRecompute,
  executeRecompute
}

export default adminInvitesAPI
