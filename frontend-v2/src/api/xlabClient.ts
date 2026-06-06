import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig
} from 'axios'
import type { ApiResponse } from '@/types'
import { getLocale } from '@/i18n'

const XLAB_API_BASE_URL = import.meta.env.VITE_XLAB_API_BASE_URL || '/xapi/v1'

export const xlabClient: AxiosInstance = axios.create({
  baseURL: XLAB_API_BASE_URL,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

xlabClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('auth_token')
  if (token && config.headers) config.headers.Authorization = `Bearer ${token}`
  if (config.headers) config.headers['Accept-Language'] = getLocale()
  return config
})

xlabClient.interceptors.response.use(
  (response: AxiosResponse) => {
    const apiResponse = response.data as ApiResponse<unknown>
    if (apiResponse && typeof apiResponse === 'object' && 'code' in apiResponse) {
      if (apiResponse.code === 0) {
        response.data = apiResponse.data
      } else {
        return Promise.reject({
          status: response.status,
          code: apiResponse.code,
          message: apiResponse.message || 'Unknown error'
        })
      }
    }
    return response
  },
  (error: AxiosError<ApiResponse<unknown>>) => {
    const data = error.response?.data
    return Promise.reject({
      status: error.response?.status || 0,
      code: data?.code,
      message: data?.message || error.message || 'Network error'
    })
  }
)

export default xlabClient
