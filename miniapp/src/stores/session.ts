import { defineStore } from 'pinia'

import { bindWechatPhone as bindWechatPhoneApi, logoutUser, redeemAssistanceCode, wechatMiniLogin, wechatMiniPhoneLogin } from '@/api/auth'
import { fetchCapabilities } from '@/api/capabilities'
import { getMe, updateProfile, uploadAvatar } from '@/api/profile'
import {
  apiMode,
  pendingRouteKey,
  sessionTokenKey,
  sessionUserKey
} from '@/api/client'
import { mockUser, type UserState } from '@/mock/data'
import { isProfileComplete } from '@/features/session/profileComplete'
import { isMpWeixinPlatform, resolveWechatLoginUnsupportedMessage } from '@/features/session/wechatLogin'
import type { BindWechatPhoneInput, UserCapabilities, UserInfo } from '@/types/api'

function restoredUser(): UserInfo | null {
  const value = uni.getStorageSync(sessionUserKey)
  if (!value) return null
  try {
    return typeof value === 'string' ? JSON.parse(value) as UserInfo : value as UserInfo
  } catch {
    uni.removeStorageSync(sessionUserKey)
    return null
  }
}

function deriveState(token: string, user: UserInfo | null): UserState {
  if (!token || !user) return 'guest'
  return isProfileComplete(user) ? 'wechatActive' : 'wechatProfileIncomplete'
}

const initialUser = restoredUser()
const initialToken = (uni.getStorageSync(sessionTokenKey) as string) || ''

export const useSessionStore = defineStore('session', {
  state: () => ({
    state: deriveState(initialToken, initialUser) as UserState,
    user: initialUser as UserInfo | null,
    token: initialToken
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.user && state.state !== 'guest'),
    isProfileComplete: (state) => state.state === 'wechatActive',
    mode: () => apiMode
  },
  actions: {
    restoreSession() {
      const token = (uni.getStorageSync(sessionTokenKey) as string) || ''
      const user = restoredUser()
      this.token = token
      this.user = user
      this.state = deriveState(token, user)
    },
    persistSession(token: string, user: UserInfo) {
      this.token = token
      this.user = user
      this.state = deriveState(token, user)
      uni.setStorageSync(sessionTokenKey, token)
      uni.setStorageSync(sessionUserKey, JSON.stringify(user))
    },
    clearSession() {
      this.token = ''
      this.user = null
      this.state = 'guest'
      uni.removeStorageSync(sessionTokenKey)
      uni.removeStorageSync(sessionUserKey)
    },
    async refreshAuthContext() {
      const user = await this.refreshMe()
      let capabilities: UserCapabilities | null = null
      try {
        capabilities = await fetchCapabilities()
      } catch {
        capabilities = null
      }
      return { user, capabilities }
    },
    async bindWechatPhone(input: BindWechatPhoneInput) {
      const result = await bindWechatPhoneApi(input)
      this.persistSession(result.accessToken, result.user)
      return this.refreshAuthContext()
    },
    async loginWithWechat() {
      const unsupported = resolveWechatLoginUnsupportedMessage()
      if (unsupported) {
        throw new Error(unsupported)
      }
      if (!isMpWeixinPlatform()) {
        throw new Error('微信快捷登录仅支持微信小程序。')
      }
      const loginResult = await new Promise<UniApp.LoginRes>((resolve, reject) => {
        uni.login({
          provider: 'weixin',
          success: resolve,
          fail: (error) => {
            const message = typeof error?.errMsg === 'string' ? error.errMsg : ''
            if (/not support|不支持|only.*mini program|provider.*weixin/i.test(message)) {
              reject(new Error(resolveWechatLoginUnsupportedMessage('h5') || '微信快捷登录仅支持微信小程序。'))
              return
            }
            reject(error)
          }
        })
      })
      if (!loginResult.code) {
        throw new Error('微信登录失败，请稍后重试。')
      }
      const result = await wechatMiniLogin(loginResult.code)
      this.persistSession(result.accessToken, result.user)
      return result
    },
    async loginWithWechatPhone(phoneCode: string) {
      if (!isMpWeixinPlatform()) {
        throw new Error('微信手机号快捷登录仅支持微信小程序。')
      }
      if (!phoneCode) {
        throw new Error('手机号授权失败，请稍后重试。')
      }
      const loginResult = await new Promise<UniApp.LoginRes>((resolve, reject) => {
        uni.login({ provider: 'weixin', success: resolve, fail: reject })
      })
      if (!loginResult.code) {
        throw new Error('微信身份验证失败，请稍后重试。')
      }
      const result = await wechatMiniPhoneLogin(phoneCode, loginResult.code)
      this.persistSession(result.accessToken, result.user)
      return result
    },
    async loginWithAssistanceCode(phone: string, code: string) { const result = await redeemAssistanceCode(phone, code); this.persistSession(result.accessToken, result.user); return result },
    async refreshMe() {
      const user = await getMe()
      this.persistSession(this.token, user)
      return user
    },
    async saveProfile(input: { nickname?: string }) {
      const user = await updateProfile(input)
      this.persistSession(this.token, user)
      return user
    },
    async saveAvatar(filePath: string) {
      const user = await uploadAvatar(filePath)
      this.persistSession(this.token, user)
      return user
    },
    requireLogin(route: string) {
      this.restoreSession()
      if (this.isLoggedIn) return true
      if (route) uni.setStorageSync(pendingRouteKey, route)
      uni.reLaunch({ url: '/pages/auth/wechat-login' })
      return false
    },
    requireProfileComplete(route: string) {
      this.restoreSession()
      if (!this.isLoggedIn) {
        if (route) uni.setStorageSync(pendingRouteKey, route)
        uni.reLaunch({ url: '/pages/auth/wechat-login' })
        return false
      }
      if (!this.isProfileComplete) {
        if (route) uni.setStorageSync(pendingRouteKey, route)
        uni.redirectTo({ url: '/pages/me/profile?onboarding=1&required=1' })
        return false
      }
      return true
    },
    finishLogin(defaultRoute = '/pages/family/my') {
      const pending = uni.getStorageSync(pendingRouteKey) as string
      uni.removeStorageSync(pendingRouteKey)
      uni.reLaunch({ url: pending || defaultRoute })
    },
    routeAfterAuth() {
      // WeChat login is sufficient to enter the app. Collect profile details only
      // when a later workflow actually needs them.
      this.finishLogin()
    },
    finishProfile() {
      this.finishLogin()
    },
    mockWechatLogin(withProfile = false) {
      this.persistSession(mockUser.token, {
        id: 'mock_user',
        phone: null,
        phoneVerified: false,
        nickname: withProfile ? mockUser.nickname : null,
        status: mockUser.status,
      })
    },
    async logout() {
      try {
        await logoutUser()
      } finally {
        this.clearSession()
      }
    }
  }
})
