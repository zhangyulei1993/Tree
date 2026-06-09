import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('@/views/Home.vue') },
    { path: '/families', component: () => import('@/views/FamilySearch.vue') },
    { path: '/families/:familyId/public', component: () => import('@/views/PublicFamily.vue') },
    { path: '/families/:familyId/tree/public', component: () => import('@/views/PublicTree.vue') },
    { path: '/login', component: () => import('@/views/Login.vue') },
    { path: '/register', component: () => import('@/views/Register.vue') },
    { path: '/me', component: () => import('@/views/Profile.vue') },
    { path: '/me/families', component: () => import('@/views/MyFamilies.vue') },
    { path: '/invite/:inviteToken', component: () => import('@/views/InviteDetail.vue') },
    { path: '/families/:familyId/join', component: () => import('@/views/JoinApply.vue') }
  ],
  scrollBehavior: () => ({ top: 0 })
})

export default router
