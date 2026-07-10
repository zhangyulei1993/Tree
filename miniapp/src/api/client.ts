import type { ApiResponse } from '@/types/api'

const environment = import.meta.env as ImportMetaEnv & {
  VITE_API_MODE?: string
  VITE_API_BASE_URL?: string
}

export const apiMode = environment.VITE_API_MODE === 'real' ? 'real' : 'mock'
export const isRealApiMode = apiMode === 'real'

const QUOTA_EXCEEDED_CODES = new Set([41502, 41503, 41504])

type QuotaApiError = Error & {
  code?: number
  data?: {
    trustTier?: unknown
  }
}

type SystemInfo = {
  uniPlatform?: string
}

type UniRuntime = {
  getSystemInfoSync?: () => SystemInfo
}

function isAbsoluteHttpUrl(value: string) {
  return /^https?:\/\//i.test(value)
}

function isLocalhostUrl(value: string) {
  return /^(https?:\/\/)?(127\.0\.0\.1|localhost)([:/]|$)/i.test(value)
}

function isQuotaApiError(error: unknown): error is QuotaApiError {
  return error instanceof Error && 'code' in error
}

function trustTierFromError(error: QuotaApiError) {
  const trustTier = error.data?.trustTier
  return typeof trustTier === 'string' ? trustTier : ''
}

function quotaReachedMessage(trustTier: string) {
  if (trustTier === 'WECHAT_ONLY') {
    return '当前扩展能力尚未开放'
  }
  return '已达到当前账号权益上限'
}

function parseQuotaErrorMessage(error: unknown, fallback: string) {
  if (isQuotaApiError(error)) {
    if (error.code && QUOTA_EXCEEDED_CODES.has(error.code)) {
      return quotaReachedMessage(trustTierFromError(error))
    }
    return error.message || fallback
  }
  if (!(error instanceof Error)) {
    return fallback
  }
  return error.message || fallback
}

function readUniRuntime(): UniRuntime | null {
  try {
    if (typeof uni !== 'undefined' && uni && typeof uni.getSystemInfoSync === 'function') {
      return uni as UniRuntime
    }
  } catch {
    // uni may be unavailable outside uni-app runtimes.
  }
  return null
}

function resolveUniPlatform(): string {
  const runtime = readUniRuntime()
  if (!runtime?.getSystemInfoSync) {
    return ''
  }
  try {
    const info = runtime.getSystemInfoSync()
    const platform = typeof info?.uniPlatform === 'string' ? info.uniPlatform.trim() : ''
    return platform
  } catch {
    return ''
  }
}

function isMpWeixinPlatform(platform = resolveUniPlatform()): boolean {
  return platform === 'mp-weixin'
}

function resolveApiBaseURL() {
  const configured = (environment.VITE_API_BASE_URL || '').trim()

  if (!isRealApiMode) {
    return configured || '/api'
  }

  // 微信小程序 real 模式必须显式注入完整 HTTPS 地址，不能回落到 /api 或本地地址。
  if (isMpWeixinPlatform()) {
    return configured
  }

  // H5 dev/build 默认走 Vite proxy：/api -> 本地后端。
  return configured || '/api'
}

export function getApiBaseURL(): string {
  return resolveApiBaseURL()
}

export function getMpWeixinApiConfigError(): string | null {
  if (!isRealApiMode || !isMpWeixinPlatform()) {
    return null
  }

  const base = getApiBaseURL().trim()
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

export const sessionTokenKey = 'tree_miniapp_user_token'
export const sessionUserKey = 'tree_miniapp_user'
export const pendingRouteKey = 'tree_miniapp_pending_route'

export class ApiError extends Error {
  code?: number
  statusCode?: number
  data?: Record<string, unknown>

  constructor(
    message: string,
    options: { code?: number; statusCode?: number; data?: Record<string, unknown> } = {}
  ) {
    super(message)
    this.name = 'ApiError'
    this.code = options.code
    this.statusCode = options.statusCode
    this.data = options.data
  }
}

interface RequestOptions<T> {
  method?: UniApp.RequestOptions['method']
  data?: T
  public?: boolean
  headers?: Record<string, string>
}

export function apiUrl(path: string) {
  const base = getApiBaseURL().replace(/\/+$/, '')
  const suffix = path.startsWith('/') ? path : `/${path}`
  return `${base}${suffix}`
}

export function resolveAssetUrl(path?: string | null) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  if (path.startsWith('/api/')) {
    const origin = getApiBaseURL().replace(/\/api\/?$/, '')
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
  if (route && route !== '/pages/auth/wechat-login') {
    uni.setStorageSync(pendingRouteKey, route)
  }
  uni.reLaunch({ url: '/pages/auth/wechat-login' })
}

function resolveNetworkErrorMessage(errMsg: string) {
  const mpConfigError = getMpWeixinApiConfigError()
  if (mpConfigError) {
    return mpConfigError
  }
  if (
    isMpWeixinPlatform()
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
  const mpConfigError = getMpWeixinApiConfigError()
  if (mpConfigError) {
    return Promise.reject(new ApiError(mpConfigError))
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
            statusCode: response.statusCode,
            data: payload?.data
          }))
          return
        }
        if (response.statusCode < 200 || response.statusCode >= 300) {
          reject(new ApiError(payload?.message || `请求失败（HTTP ${response.statusCode}）`, {
            code: payload?.code,
            statusCode: response.statusCode,
            data: payload?.data
          }))
          return
        }
        if (!payload || payload.code !== 0) {
          reject(new ApiError(payload?.message || '请求失败', {
            code: payload?.code,
            statusCode: response.statusCode,
            data: payload?.data
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
  return parseQuotaErrorMessage(error, fallback)
}
