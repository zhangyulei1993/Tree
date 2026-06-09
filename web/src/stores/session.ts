import { defineStore } from 'pinia'

import { currentUser, type UserState } from '@/mock/data'

export const useSessionStore = defineStore('session', {
  state: () => ({
    state: 'guest' as UserState,
    user: currentUser
  }),
  getters: {
    isLoggedIn: (state) => state.state !== 'guest',
    isPhoneBound: (state) => state.state === 'phoneBoundActive'
  },
  actions: {
    setState(next: UserState) {
      this.state = next
    },
    mockPhoneLogin() {
      this.state = 'phoneBoundActive'
    },
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
