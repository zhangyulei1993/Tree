import { defineStore } from 'pinia'

import { loginPhone, logoutUser, registerPhone } from '@/api/auth'
import { apiMode, sessionTokenKey, sessionUserKey } from '@/api/client'
import { currentUser, type UserState } from '@/mock/data'
import type { LoginPhoneInput, RegisterPhoneInput, UserInfo } from '@/types/api'

function restoredUser(): UserInfo | null {
  const raw = sessionStorage.getItem(sessionUserKey)
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserInfo
  } catch {
    sessionStorage.removeItem(sessionUserKey)
    return null
  }
}

const initialUser = restoredUser()
const initialToken = sessionStorage.getItem(sessionTokenKey) || ''

export const useSessionStore = defineStore('session', {
  state: () => ({
    state: (initialToken && initialUser ? 'phoneBoundActive' : 'guest') as UserState,
    user: initialUser as UserInfo | null,
    token: initialToken
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.user && state.state !== 'guest'),
    isPhoneBound: (state) => Boolean(state.user?.phoneVerified && state.state === 'phoneBoundActive'),
    mode: () => apiMode
  },
  actions: {
    persistSession(token: string, user: UserInfo) {
      this.token = token
      this.user = user
      this.state = user.phoneVerified ? 'phoneBoundActive' : 'wechatLoggedInPendingPhone'
      sessionStorage.setItem(sessionTokenKey, token)
      sessionStorage.setItem(sessionUserKey, JSON.stringify(user))
    },
    clearSession() {
      this.token = ''
      this.user = null
      this.state = 'guest'
      sessionStorage.removeItem(sessionTokenKey)
      sessionStorage.removeItem(sessionUserKey)
    },
    async login(input: LoginPhoneInput) {
      const result = await loginPhone(input)
      this.persistSession(result.accessToken, result.user)
    },
    async register(input: RegisterPhoneInput) {
      const result = await registerPhone(input)
      this.persistSession(result.accessToken, result.user)
    },
    setState(next: UserState) {
      this.state = next
    },
    mockPhoneLogin() {
      this.persistSession(currentUser.token, {
        id: currentUser.id,
        phone: currentUser.maskedPhone,
        phoneVerified: currentUser.phoneVerified,
        nickname: currentUser.nickname,
        status: currentUser.status
      })
    },
    mockWechatLogin() {
      this.state = 'wechatLoggedInPendingPhone'
    },
    mockBindPhone() {
      if (this.user) {
        this.user.phoneVerified = true
      }
      this.state = 'phoneBoundActive'
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
