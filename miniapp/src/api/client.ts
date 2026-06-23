import type { ApiResponse } from '@/types/api'

const environment = import.meta.env as ImportMetaEnv & {
  VITE_API_MODE?: string
  VITE_API_BASE_URL?: string
  UNI_PLATFORM?: string
}

const uniPlatform = environment.UNI_PLATFORM || ''
const isMpWeixinPlatform = uniPlatform === 'mp-weixin'

export const apiMode = environment.VITE_API_MODE === 'real' ? 'real' : 'mock'
export const isRealApiMode = apiMode === 'real'

function isAbsoluteHttpUrl(value: string) {
  return /^https?:\/\//i.test(value)
}

function isLocalhostUrl(value: string) {
  return /^(https?:\/\/)?(127\.0\.0\.1|localhost)([:/]|$)/i.test(value)
}

function resolveApiBaseURL() {
  const configured = (environment.VITE_API_BASE_URL || '').trim()

  if (!isRealApiMode) {
    return configured || '/api'
  }

  // 微信小程序 real 模式必须显式注入完整 HTTPS 地址，不能回落到 /api 或本地地址。
  if (isMpWeixinPlatform) {
    return configured
  }

  // H5 dev/build 默认走 Vite proxy：/api -> 本地后端。
  return configured || '/api'
}

export const apiBaseURL = resolveApiBaseURL()

export function getMpWeixinApiConfigError(): string | null {
  if (!isRealApiMode || !isMpWeixinPlatform) {
    return null
  }

  const base = apiBaseURL.trim()
  if (!base) {
    return '小程序未配置 API 地址。请使用 pnpm build:mp-weixin:staging 重新构建，或设置 VITE_API_BASE_URL=https://tapi.bigbigboy.cn/api。'
  }
  if (!isAbsoluteHttpUrl(base)) {
    return `小程序 real 模式不能使用相对地址「${base}」。请使用完整 HTTPS 地址（如 https://tapi.bigbigboy.cn/api）重新构建。`
  }
  if (isLocalhostUrl(base)) {
    return '小程序不能访问 localhost 或 127.0.0.1。请使用 https://tapi.bigbigboy.cn/api 重新构建。'
  }
  if (!/^https:\/\//i.test(base)) {
    return '微信小程序 request 合法域名要求 HTTPS。请将 VITE_API_BASE_URL 设为 https://tapi.bigbigboy.cn/api 后重新构建。'
  }
  return null
}

export const mpWeixinApiConfigError = getMpWeixinApiConfigError()

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

export function apiUrl(path: string) {
  const base = apiBaseURL.replace(/\/+$/, '')
  const suffix = path.startsWith('/') ? path : `/${path}`
  return `${base}${suffix}`
}

export function resolveAssetUrl(path?: string | null) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  if (path.startsWith('/api/')) {
    const origin = apiBaseURL.replace(/\/api\/?$/, '')
    return `${origin}${path}`
  }
  return apiUrl(path)
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

function resolveNetworkErrorMessage(errMsg: string) {
  if (mpWeixinApiConfigError) {
    return mpWeixinApiConfigError
  }
  if (
    isMpWeixinPlatform
    && isRealApiMode
    && /ERR_CONNECTION_REFUSED|request:fail|cronet_error_code:-102/i.test(errMsg)
  ) {
    return '无法连接 API 服务器。请确认已使用 staging 地址构建（https://tapi.bigbigboy.cn/api），并在微信公众平台配置 request 合法域名 tapi.bigbigboy.cn。'
  }
  return errMsg || '网络请求失败'
}

export function request<TResponse, TData = unknown>(
  path: string,
  options: RequestOptions<TData> = {}
): Promise<TResponse> {
  if (mpWeixinApiConfigError) {
    return Promise.reject(new ApiError(mpWeixinApiConfigError))
  }

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
      url: apiUrl(path),
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
        reject(new ApiError(resolveNetworkErrorMessage(error.errMsg || '')))
      }
    })
  })
}

export function apiErrorMessage(error: unknown, fallback = '请求失败，请稍后重试。') {
  return error instanceof Error && error.message ? error.message : fallback
}
