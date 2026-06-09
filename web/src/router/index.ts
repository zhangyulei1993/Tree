import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { useSessionStore } from '@/stores/session'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
  }
}

const routes: RouteRecordRaw[] = [
  { path: '/', component: () => import('@/views/Home.vue') },
  { path: '/families', component: () => import('@/views/FamilySearch.vue') },
  { path: '/families/:familyId/public', component: () => import('@/views/PublicFamily.vue') },
  { path: '/families/:familyId/tree/public', component: () => import('@/views/PublicTree.vue') },
  { path: '/login', component: () => import('@/views/Login.vue') },
  { path: '/register', component: () => import('@/views/Register.vue') },
  { path: '/me', component: () => import('@/views/Profile.vue'), meta: { requiresAuth: true } },
  { path: '/me/families', component: () => import('@/views/MyFamilies.vue'), meta: { requiresAuth: true } },
  { path: '/invite/:inviteToken', component: () => import('@/views/InviteDetail.vue') },
  { path: '/families/:familyId/join', component: () => import('@/views/JoinApply.vue') }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
})

router.beforeEach((to) => {
  const session = useSessionStore()
  if (to.meta.requiresAuth && !session.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if ((to.path === '/login' || to.path === '/register') && session.isLoggedIn) {
    return '/me'
  }
  return true
})

export default router
