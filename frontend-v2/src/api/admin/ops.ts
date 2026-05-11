import { apiClient } from '@/api/client'

export async function getOpsSnapshot() {
  const { data } = await apiClient.get('/admin/ops/dashboard/snapshot-v2')
  return data
}

export async function getOpsOverview(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/ops/dashboard/overview', { params })
  return data
}

export async function getOpsRealtimeTraffic(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/ops/realtime-traffic', { params })
  return data
}

export async function getOpsConcurrency(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/ops/concurrency', { params })
  return data
}

export async function getOpsAccountAvailability(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/ops/account-availability', { params })
  return data
}

export async function listOpsSystemLogs(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/ops/system-logs', { params })
  return data
}

export async function listOpsRequestErrors(params: Record<string, unknown> = {}) {
  const { data } = await apiClient.get('/admin/ops/request-errors', { params })
  return data
}

export async function getOpsRequestError(id: number | string) {
  const { data } = await apiClient.get(`/admin/ops/request-errors/${id}`)
  return data
}

export async function retryOpsRequestErrorClient(id: number | string) {
  const { data } = await apiClient.post(`/admin/ops/request-errors/${id}/retry-client`)
  return data
}

export async function resolveOpsRequestError(id: number | string, resolved: boolean) {
  const { data } = await apiClient.put(`/admin/ops/request-errors/${id}/resolve`, { resolved })
  return data
}

export const opsAPI = {
  getOpsSnapshot,
  getOpsOverview,
  getOpsRealtimeTraffic,
  getOpsConcurrency,
  getOpsAccountAvailability,
  listOpsSystemLogs,
  listOpsRequestErrors,
  getOpsRequestError,
  retryOpsRequestErrorClient,
  resolveOpsRequestError
}
