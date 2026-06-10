import type { AdminRole } from '@/stores/auth'
import { useAuthStore } from '@/stores/auth'
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    requiresAuth?: boolean
    roles?: AdminRole[]
  }
}

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/admin/dashboard'
  },
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/Login.vue'),
    meta: { title: '后台登录' }
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/Dashboard.vue'), meta: { title: '仪表盘' } },
      { path: 'users', name: 'Users', component: () => import('@/views/Users.vue'), meta: { title: '用户管理' } },
      { path: 'users/:userId', name: 'UserDetail', component: () => import('@/views/UserDetail.vue'), meta: { title: '用户详情' } },
      { path: 'families', name: 'Families', component: () => import('@/views/Families.vue'), meta: { title: '家庭管理' } },
      { path: 'families/:familyId', name: 'FamilyDetail', component: () => import('@/views/FamilyDetail.vue'), meta: { title: '家庭详情' } },
      { path: 'families/:familyId/members', name: 'FamilyMembers', component: () => import('@/views/FamilyMembers.vue'), meta: { title: '家庭成员管理' } },
      { path: 'public-applications', name: 'PublicApplications', component: () => import('@/views/PublicApplications.vue'), meta: { title: '公开申请审核' } },
      { path: 'visitor-messages', name: 'VisitorMessages', component: () => import('@/views/VisitorMessages.vue'), meta: { title: '游客留言审核' } },
      {
        path: 'founder-transfer-requests',
        name: 'FounderTransferRequests',
        component: () => import('@/views/FounderTransferRequests.vue'),
        meta: { title: '创始人转让审核', roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] }
      },
      {
        path: 'dissolution-requests',
        name: 'DissolutionRequests',
        component: () => import('@/views/DissolutionRequests.vue'),
        meta: { title: '家庭解散审核', roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] }
      },
      {
        path: 'admin-users',
        name: 'AdminUsers',
        component: () => import('@/views/AdminUsers.vue'),
        meta: { title: '后台管理员管理', roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] }
      },
      { path: 'operation-logs', name: 'OperationLogs', component: () => import('@/views/OperationLogs.vue'), meta: { title: '操作日志' } },
      { path: 'no-permission', name: 'NoPermission', component: () => import('@/views/NoPermission.vue'), meta: { title: '无权限' } }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/admin/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.init()

  if (to.path === '/admin/login' && auth.isLoggedIn) {
    const redirect = typeof to.query.redirect === 'string' ? to.query.redirect : '/admin/dashboard'
    return redirect.startsWith('/admin/') ? redirect : '/admin/dashboard'
  }

  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return {
      path: '/admin/login',
      query: { redirect: to.fullPath }
    }
  }

  if (to.meta.roles && !auth.hasRole(to.meta.roles)) {
    return '/admin/no-permission'
  }

  return true
})

export default router
