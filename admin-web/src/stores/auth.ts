import { defineStore } from 'pinia'

export type AdminRole = 'ROOT_ADMIN' | 'SUPER_ADMIN' | 'PLATFORM_ADMIN'

export interface MockAdmin {
  id: string
  username: string
  displayName: string
  role: AdminRole
}

const roleName: Record<AdminRole, string> = {
  ROOT_ADMIN: 'Root 管理员',
  SUPER_ADMIN: '超级管理员',
  PLATFORM_ADMIN: '平台管理员'
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    admin: null as MockAdmin | null,
    token: sessionStorage.getItem('tree_admin_mock_token') || ''
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.admin && state.token),
    roleLabel: (state) => (state.admin ? roleName[state.admin.role] : '')
  },
  actions: {
    mockLogin(role: AdminRole) {
      this.admin = {
        id: `admin_${role.toLowerCase()}`,
        username: `${role.toLowerCase()}_demo`,
        displayName: roleName[role],
        role
      }
      this.token = 'example_admin_token'
      sessionStorage.setItem('tree_admin_mock_token', this.token)
    },
    switchRole(role: AdminRole) {
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
    logout() {
      this.admin = null
      this.token = ''
      sessionStorage.removeItem('tree_admin_mock_token')
    }
  }
})
