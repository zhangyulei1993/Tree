import axios, { AxiosError, type AxiosResponse } from 'axios'

import type { ApiResponse } from '@/types/api'

export const ADMIN_TOKEN_KEY = 'tree_admin_token'
export const ADMIN_INFO_KEY = 'tree_admin_info'
export const apiMode = import.meta.env.VITE_API_MODE === 'real' ? 'real' : 'mock'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10000
})

apiClient.interceptors.request.use((config) => {
  const token = sessionStorage.getItem(ADMIN_TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiResponse<null>>) => {
    if (error.response?.status === 401) {
      sessionStorage.removeItem(ADMIN_TOKEN_KEY)
      sessionStorage.removeItem(ADMIN_INFO_KEY)
      if (window.location.pathname !== '/admin/login') {
        const redirect = `${window.location.pathname}${window.location.search}`
        window.location.assign(`/admin/login?redirect=${encodeURIComponent(redirect)}`)
      }
    }
    return Promise.reject(error)
  }
)

export function unwrapData<T>(response: AxiosResponse<ApiResponse<T>>): T {
  if (response.data.code !== 0) {
    throw new Error(response.data.message || '请求失败')
  }
  return response.data.data
}

export function getApiErrorMessage(error: unknown): string {
  if (axios.isAxiosError<ApiResponse<null>>(error)) {
    return error.response?.data?.message || error.message || '请求失败'
  }
  return error instanceof Error ? error.message : '请求失败'
}
