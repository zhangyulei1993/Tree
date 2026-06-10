import type { ApiResponse } from '@/types/api'

const environment = import.meta.env as Record<string, string | undefined>

export const apiMode = environment.VITE_API_MODE === 'real' ? 'real' : 'mock'
export const isRealApiMode = apiMode === 'real'
export const apiBaseURL = environment.VITE_API_BASE_URL || '/api'

export const sessionTokenKey = 'tree_miniapp_user_token'
export const sessionUserKey = 'tree_miniapp_user'
export const pendingRouteKey = 'tree_miniapp_pending_route'

export class ApiError extends Error {
  code?: number
  statusCode?: number

  constructor(message: string, options: { code?: number; statusCode?: number } = {}) {
    super(message)
    this.name = 'ApiError'
    this.code = options.code
    this.statusCode = options.statusCode
  }
}

interface RequestOptions<T> {
  method?: UniApp.RequestOptions['method']
  data?: T
  public?: boolean
  headers?: Record<string, string>
}

function requestURL(path: string) {
  const base = apiBaseURL.replace(/\/+$/, '')
  const suffix = path.startsWith('/') ? path : `/${path}`
  return `${base}${suffix}`
}

function currentRoute() {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1] as (UniApp.Page & { options?: Record<string, string> }) | undefined
  if (!page?.route) return ''
  const query = Object.entries(page.options || {})
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
    .join('&')
  return `/${page.route}${query ? `?${query}` : ''}`
}

export function clearStoredSession() {
  uni.removeStorageSync(sessionTokenKey)
  uni.removeStorageSync(sessionUserKey)
}

function handleUnauthorized(skipRedirect: boolean) {
  clearStoredSession()
  if (skipRedirect) return

  const route = currentRoute()
  if (route && route !== '/pages/auth/phone-login') {
    uni.setStorageSync(pendingRouteKey, route)
  }
  uni.reLaunch({ url: '/pages/auth/phone-login' })
}

export function request<TResponse, TData = unknown>(
  path: string,
  options: RequestOptions<TData> = {}
): Promise<TResponse> {
  const token = uni.getStorageSync(sessionTokenKey) as string
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  return new Promise((resolve, reject) => {
    uni.request<ApiResponse<TResponse>>({
      url: requestURL(path),
      method: options.method || 'GET',
      data: options.data,
      header: headers,
      success(response) {
        const payload = response.data
        if (response.statusCode === 401) {
          handleUnauthorized(Boolean(options.public))
          reject(new ApiError(payload?.message || '登录状态已失效', {
            code: payload?.code,
            statusCode: response.statusCode
          }))
          return
        }
        if (response.statusCode < 200 || response.statusCode >= 300) {
          reject(new ApiError(payload?.message || `请求失败（HTTP ${response.statusCode}）`, {
            code: payload?.code,
            statusCode: response.statusCode
          }))
          return
        }
        if (!payload || payload.code !== 0) {
          reject(new ApiError(payload?.message || '请求失败', {
            code: payload?.code,
            statusCode: response.statusCode
          }))
          return
        }
        resolve(payload.data)
      },
      fail(error) {
        reject(new ApiError(error.errMsg || '网络请求失败'))
      }
    })
  })
}

export function apiErrorMessage(error: unknown, fallback = '请求失败，请稍后重试。') {
  return error instanceof Error && error.message ? error.message : fallback
}
