import { apiClient } from './client'

export async function authorizeXlabOAuth(payload: Record<string, unknown>) {
  const { data } = await apiClient.post('/oauth/authorize', payload)
  return data
}

export const oauthAPI = {
  authorizeXlabOAuth
}

