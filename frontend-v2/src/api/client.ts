import axios, {
  type AxiosInstance,
  type AxiosError,
  type InternalAxiosRequestConfig,
  type AxiosResponse
} from 'axios'
import type { ApiResponse } from '@/types'
import { getLocale } from '@/i18n'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

function subscribeTokenRefresh(cb: (token: string) => void) {
  refreshSubscribers.push(cb)
}

function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((cb) => cb(token))
  refreshSubscribers = []
}

const getUserTimezone = (): string => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('auth_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    if (config.headers) {
      config.headers['Accept-Language'] = getLocale()
    }
    if (config.method === 'get') {
      if (!config.params) config.params = {}
      config.params.timezone = getUserTimezone()
    }
    return config
  },
  (error) => Promise.reject(error)
)

apiClient.interceptors.response.use(
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
  async (error: AxiosError<ApiResponse<unknown>>) => {
    if (error.code === 'ERR_CANCELED' || axios.isCancel(error)) {
      return Promise.reject(error)
    }

    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (error.response) {
      const { status, data } = error.response
      const url = String(error.config?.url || '')
      const apiData = (typeof data === 'object' && data !== null ? data : {}) as Record<string, any>

      if (status === 401 && !originalRequest._retry) {
        const refreshToken = localStorage.getItem('refresh_token')
        const isAuthEndpoint =
          url.includes('/auth/login') || url.includes('/auth/register') || url.includes('/auth/refresh')

        if (refreshToken && !isAuthEndpoint) {
          if (isRefreshing) {
            return new Promise((resolve, reject) => {
              subscribeTokenRefresh((newToken) => {
                if (newToken) {
                  originalRequest._retry = true
                  if (originalRequest.headers) {
                    originalRequest.headers.Authorization = `Bearer ${newToken}`
                  }
                  resolve(apiClient(originalRequest))
                } else {
                  reject({ status, code: apiData.code, message: apiData.message || error.message })
                }
              })
            })
          }

          originalRequest._retry = true
          isRefreshing = true

          try {
            const refreshResponse = await axios.post(
              `${API_BASE_URL}/auth/refresh`,
              { refresh_token: refreshToken },
              { headers: { 'Content-Type': 'application/json' } }
            )
            const refreshData = refreshResponse.data as ApiResponse<{
              access_token: string
              refresh_token: string
              expires_in: number
            }>
            if (refreshData.code === 0 && refreshData.data) {
              const { access_token, refresh_token: newRT, expires_in } = refreshData.data
              localStorage.setItem('auth_token', access_token)
              localStorage.setItem('refresh_token', newRT)
              localStorage.setItem('token_expires_at', String(Date.now() + expires_in * 1000))
              onTokenRefreshed(access_token)
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${access_token}`
              }
              isRefreshing = false
              return apiClient(originalRequest)
            }
            throw new Error('refresh failed')
          } catch (refreshErr) {
            onTokenRefreshed('')
            isRefreshing = false
            localStorage.removeItem('auth_token')
            localStorage.removeItem('refresh_token')
            localStorage.removeItem('auth_user')
            localStorage.removeItem('token_expires_at')
            sessionStorage.setItem('auth_expired', '1')
            if (!window.location.pathname.includes('/login')) {
              window.location.href = '/login'
            }
            return Promise.reject({
              status: 401,
              code: 'TOKEN_REFRESH_FAILED',
              message: 'Session expired. Please log in again.'
            })
          }
        }

        const hasToken = !!localStorage.getItem('auth_token')
        localStorage.removeItem('auth_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('auth_user')
        localStorage.removeItem('token_expires_at')
        if (hasToken && !isAuthEndpoint) {
          sessionStorage.setItem('auth_expired', '1')
        }
        if (!window.location.pathname.includes('/login')) {
          window.location.href = '/login'
        }
      }

      return Promise.reject({
        status,
        code: apiData.code,
        error: apiData.error,
        message: apiData.message || apiData.detail || error.message
      })
    }

    return Promise.reject({ status: 0, message: 'Network error. Please check your connection.' })
  }
)

export default apiClient
