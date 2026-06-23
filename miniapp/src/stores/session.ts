import { defineStore } from 'pinia'

import { bindPhone, loginPhone, logoutUser, registerPhone, wechatMiniLogin } from '@/api/auth'
import { getMe, updateProfile, uploadAvatar } from '@/api/profile'
import {
  apiMode,
  pendingRouteKey,
  sessionTokenKey,
  sessionUserKey
} from '@/api/client'
import { mockUser, type UserState } from '@/mock/data'
import type { BindPhoneInput, LoginPhoneInput, RegisterPhoneInput, UserInfo } from '@/types/api'

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

const initialUser = restoredUser()
const initialToken = (uni.getStorageSync(sessionTokenKey) as string) || ''

export const useSessionStore = defineStore('session', {
  state: () => ({
    state: (initialToken && initialUser
      ? initialUser.phoneVerified ? 'phoneBoundActive' : 'wechatLoggedInPendingPhone'
      : 'guest') as UserState,
    user: initialUser as UserInfo | null,
    token: initialToken
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.user && state.state !== 'guest'),
    isPhoneBound: (state) => Boolean(state.user?.phoneVerified && state.state === 'phoneBoundActive'),
    mode: () => apiMode
  },
  actions: {
    restoreSession() {
      const token = (uni.getStorageSync(sessionTokenKey) as string) || ''
      const user = restoredUser()
      this.token = token
      this.user = user
      this.state = token && user
        ? user.phoneVerified ? 'phoneBoundActive' : 'wechatLoggedInPendingPhone'
        : 'guest'
    },
    persistSession(token: string, user: UserInfo) {
      this.token = token
      this.user = user
      this.state = user.phoneVerified ? 'phoneBoundActive' : 'wechatLoggedInPendingPhone'
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
    async login(input: LoginPhoneInput) {
      const result = await loginPhone(input)
      this.persistSession(result.accessToken, result.user)
    },
    async loginWithWechat() {
      const loginResult = await new Promise<UniApp.LoginRes>((resolve, reject) => {
        uni.login({
          provider: 'weixin',
          success: resolve,
          fail: reject
        })
      })
      if (!loginResult.code) {
        throw new Error('微信登录失败，请稍后重试。')
      }
      const result = await wechatMiniLogin(loginResult.code)
      this.persistSession(result.accessToken, result.user)
      return result
    },
    async bindPhone(input: BindPhoneInput) {
      const result = await bindPhone(input)
      this.persistSession(result.accessToken, result.user)
    },
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
    async register(input: RegisterPhoneInput) {
      const result = await registerPhone(input)
      this.persistSession(result.accessToken, result.user)
    },
    requireLogin(route: string) {
      this.restoreSession()
      if (this.isLoggedIn) return true
      if (route) uni.setStorageSync(pendingRouteKey, route)
      uni.reLaunch({ url: '/pages/auth/wechat-login' })
      return false
    },
    requirePhoneBound(route: string) {
      this.restoreSession()
      if (!this.isLoggedIn) {
        if (route) uni.setStorageSync(pendingRouteKey, route)
        uni.reLaunch({ url: '/pages/auth/wechat-login' })
        return false
      }
      if (!this.isPhoneBound) {
        if (route) uni.setStorageSync(pendingRouteKey, route)
        uni.navigateTo({ url: '/pages/auth/bind-phone' })
        return false
      }
      return true
    },
    finishLogin(defaultRoute = '/pages/family/my') {
      const pending = uni.getStorageSync(pendingRouteKey) as string
      uni.removeStorageSync(pendingRouteKey)
      uni.reLaunch({ url: pending || defaultRoute })
    },
    finishBind() {
      this.finishLogin()
    },
    mockWechatLogin() {
      this.persistSession(mockUser.token, {
        id: 'mock_user',
        phone: null,
        phoneVerified: false,
        nickname: mockUser.nickname,
        status: mockUser.status,
        passwordSet: false
      })
    },
    mockBindPhone() {
      if (!this.user) return
      this.persistSession(this.token || mockUser.token, {
        ...this.user,
        phone: mockUser.maskedPhone,
        phoneVerified: true,
        passwordSet: true
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
