import { defineStore } from 'pinia'

import { mockUser, type UserState } from '@/mock/data'

export const useSessionStore = defineStore('session', {
  state: () => ({
    state: 'guest' as UserState,
    user: mockUser
  }),
  getters: {
    isLoggedIn: (state) => state.state !== 'guest',
    isPhoneBound: (state) => state.state === 'phoneBoundActive'
  },
  actions: {
    mockWechatLogin() {
      this.state = 'wechatLoggedInPendingPhone'
    },
    mockBindPhone() {
      this.state = 'phoneBoundActive'
    },
    logout() {
      this.state = 'guest'
    }
  }
})
