import { createRouter, createWebHistory } from 'vue-router'
import Login from './views/Login.vue'
import AdminLayout from './layout/AdminLayout.vue'
import Dashboard from './views/Dashboard.vue'
import SiteConfig from './views/SiteConfig.vue'
import BannerManage from './views/BannerManage.vue'
import ArticleManage from './views/ArticleManage.vue'
import ContactManage from './views/ContactManage.vue'
import FamilyManage from './views/FamilyManage.vue'
import MemberManage from './views/MemberManage.vue'
import RelationshipManage from './views/RelationshipManage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: Login },
    {
      path: '/',
      component: AdminLayout,
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', component: Dashboard },
        { path: 'site', component: SiteConfig },
        { path: 'banner', component: BannerManage },
        { path: 'article', component: ArticleManage },
        { path: 'contact', component: ContactManage },
        { path: 'family', component: FamilyManage },
        { path: 'member', component: MemberManage },
        { path: 'relationship', component: RelationshipManage }
      ]
    }
  ]
})

router.beforeEach((to) => {
  const token = localStorage.getItem('admin_token')

  if (to.path !== '/login' && !token) {
    return '/login'
  }

  if (to.path === '/login' && token) {
    return '/dashboard'
  }

  return true
})

export default router