import { apiClient } from '@/api/client'

export interface BackupS3Config {
  endpoint: string
  region: string
  bucket: string
  prefix: string
  access_key_id: string
  secret_access_key_configured?: boolean
  secret_access_key?: string
  use_path_style: boolean
}

export interface BackupScheduleConfig {
  enabled: boolean
  cron_expression: string
  retention_days: number
}

export interface BackupRecord {
  id: string
  size_bytes: number
  created_at: string
  status: 'completed' | 'failed' | 'in_progress'
  storage: 's3' | 'local'
  s3_key?: string
  error_message?: string
  encrypted: boolean
}

export interface CreateBackupRequest {
  password?: string
  description?: string
}

export interface TestS3Response {
  success: boolean
  message: string
  details?: string
}

export async function getS3Config() {
  const { data } = await apiClient.get<BackupS3Config>('/admin/backups/s3-config')
  return data
}

export async function updateS3Config(config: BackupS3Config) {
  const { data } = await apiClient.put<BackupS3Config>('/admin/backups/s3-config', config)
  return data
}

export async function testS3Connection(config: BackupS3Config) {
  const { data } = await apiClient.post<TestS3Response>('/admin/backups/s3-config/test', config)
  return data
}

export async function getSchedule() {
  const { data } = await apiClient.get<BackupScheduleConfig>('/admin/backups/schedule')
  return data
}

export async function updateSchedule(config: BackupScheduleConfig) {
  const { data } = await apiClient.put<BackupScheduleConfig>('/admin/backups/schedule', config)
  return data
}

export async function createBackup(req?: CreateBackupRequest) {
  const { data } = await apiClient.post<BackupRecord>('/admin/backups', req || {})
  return data
}

export async function listBackups() {
  const { data } = await apiClient.get<{ items: BackupRecord[] }>('/admin/backups')
  return data
}

export async function deleteBackup(id: string) {
  await apiClient.delete(`/admin/backups/${id}`)
}

export async function getBackupDownloadURL(id: string) {
  const { data } = await apiClient.get<{ url: string }>(`/admin/backups/${id}/download-url`)
  return data
}

export async function restoreBackup(id: string, password: string) {
  const { data } = await apiClient.post<BackupRecord>(`/admin/backups/${id}/restore`, { password })
  return data
}

export const adminBackupAPI = {
  getS3Config,
  updateS3Config,
  testS3Connection,
  getSchedule,
  updateSchedule,
  createBackup,
  listBackups,
  deleteBackup,
  getBackupDownloadURL,
  restoreBackup
}
