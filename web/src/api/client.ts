import axios, { AxiosError } from 'axios'

import type { ApiResponse } from '@/types/api'

const environment = (import.meta as ImportMeta & {
  env: {
    VITE_API_MODE?: string
    VITE_API_BASE_URL?: string
  }
}).env

export const apiMode = environment.VITE_API_MODE === 'real' ? 'real' : 'mock'
export const isRealApiMode = apiMode === 'real'

export const sessionTokenKey = 'tree_web_user_token'
export const sessionUserKey = 'tree_web_user'

export const apiClient = axios.create({
  baseURL: environment.VITE_API_BASE_URL || '/api',
  timeout: 10000
})

apiClient.interceptors.request.use((config) => {
  const token = sessionStorage.getItem(sessionTokenKey)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiResponse<unknown>>) => {
    if (error.response?.status === 401) {
      sessionStorage.removeItem(sessionTokenKey)
      sessionStorage.removeItem(sessionUserKey)
      if (window.location.pathname !== '/login') {
        const redirect = `${window.location.pathname}${window.location.search}`
        window.location.assign(`/login?redirect=${encodeURIComponent(redirect)}`)
      }
    }
    return Promise.reject(error)
  }
)

export function apiErrorMessage(error: unknown, fallback = '请求失败，请稍后重试。') {
  if (axios.isAxiosError<ApiResponse<unknown>>(error)) {
    return error.response?.data?.message || error.message || fallback
  }
  return error instanceof Error ? error.message : fallback
}
