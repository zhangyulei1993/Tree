import { defineStore } from 'pinia'

import { getAdminMe, login as loginAdmin, logout as logoutAdmin } from '@/api/adminAuth'
import { ADMIN_INFO_KEY, ADMIN_TOKEN_KEY, apiMode } from '@/api/client'
import type { AdminInfo, AdminRole } from '@/types/api'

export type { AdminRole } from '@/types/api'

const roleName: Record<AdminRole, string> = {
  ROOT_ADMIN: 'Root 管理员',
  SUPER_ADMIN: '超级管理员',
  PLATFORM_ADMIN: '平台管理员'
}

function storedAdmin(): AdminInfo | null {
  const value = sessionStorage.getItem(ADMIN_INFO_KEY)
  if (!value) return null
  try {
    return JSON.parse(value) as AdminInfo
  } catch {
    sessionStorage.removeItem(ADMIN_INFO_KEY)
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    admin: storedAdmin() as AdminInfo | null,
    token: sessionStorage.getItem(ADMIN_TOKEN_KEY) || '',
    initialized: false,
    initializing: false
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.admin && state.token),
    roleLabel: (state) => (state.admin ? roleName[state.admin.role] : ''),
    hasRole: (state) => (roles: AdminRole[]) => Boolean(state.admin && roles.includes(state.admin.role))
  },
  actions: {
    persistSession(token: string, admin: AdminInfo) {
      this.token = token
      this.admin = admin
      sessionStorage.setItem(ADMIN_TOKEN_KEY, token)
      sessionStorage.setItem(ADMIN_INFO_KEY, JSON.stringify(admin))
    },
    clearSession() {
      this.admin = null
      this.token = ''
      sessionStorage.removeItem(ADMIN_TOKEN_KEY)
      sessionStorage.removeItem(ADMIN_INFO_KEY)
      sessionStorage.removeItem('tree_admin_mock_token')
    },
    async login(username: string, password: string) {
      const result = await loginAdmin(username, password)
      this.persistSession(result.accessToken, result.admin)
      this.initialized = true
      return result.admin
    },
    async init() {
      if (this.initialized || this.initializing) return
      this.initializing = true
      try {
        if (apiMode === 'real' && this.token) {
          const admin = await getAdminMe()
          this.persistSession(this.token, admin)
        } else if (apiMode !== 'real' && this.token && !this.admin) {
          this.admin = storedAdmin()
        }
      } catch {
        this.clearSession()
      } finally {
        this.initialized = true
        this.initializing = false
      }
    },
    mockLogin(role: AdminRole) {
      if (apiMode === 'real') return
      const admin: AdminInfo = {
        id: role === 'ROOT_ADMIN' ? 1 : role === 'SUPER_ADMIN' ? 2 : 3,
        username: `${role.toLowerCase()}_demo`,
        displayName: roleName[role],
        role,
        status: 'ACTIVE'
      }
      this.persistSession('example_admin_token', admin)
      this.initialized = true
    },
    switchRole(role: AdminRole) {
      if (apiMode === 'real') return
      if (!this.admin) {
        this.mockLogin(role)
        return
      }
      this.admin = {
        ...this.admin,
        role,
        displayName: roleName[role],
        username: `${role.toLowerCase()}_demo`
      }
    },
    async logout() {
      try {
        if (apiMode === 'real' && this.token) {
          await logoutAdmin()
        }
      } finally {
        this.clearSession()
        this.initialized = true
      }
    }
  }
})
