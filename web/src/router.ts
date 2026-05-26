import { createRouter, createWebHistory } from 'vue-router'
import WebLayout from './layout/WebLayout.vue'
import Home from './views/Home.vue'
import About from './views/About.vue'
import Stories from './views/Stories.vue'
import StoryDetail from './views/StoryDetail.vue'
import Contact from './views/Contact.vue'
import FamilyTree from './views/FamilyTree.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: WebLayout,
      children: [
        { path: '', component: Home },
        { path: 'about', component: About },
        { path: 'stories', component: Stories },
        { path: 'story/:id', component: StoryDetail },
        { path: 'family-tree', component: FamilyTree },
        { path: 'contact', component: Contact }
      ]
    }
  ],
  scrollBehavior() {
    return { top: 0 }
  }
})

export default router