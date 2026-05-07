import { apiClient } from './client'

export interface XlabOAuthAuthorizeRequest {
  client_id: string
  redirect_uri: string
  response_type: string
  scope?: string
  state?: string
}

export interface XlabOAuthAuthorizeResponse {
  redirect_uri: string
}

export const xlabOAuthAPI = {
  async authorize(request: XlabOAuthAuthorizeRequest): Promise<XlabOAuthAuthorizeResponse> {
    const { data } = await apiClient.post<XlabOAuthAuthorizeResponse>('/oauth/authorize', request)
    return data
  },
}
