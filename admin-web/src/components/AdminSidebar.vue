<template>
  <aside class="sidebar">
    <div class="brand">
      <div class="brand-mark">T</div>
      <div>
        <strong>Tree</strong>
        <span>管理后台</span>
      </div>
    </div>
    <el-menu :default-active="$route.path" router class="menu">
      <el-menu-item v-for="item in visibleMenus" :key="item.path" :index="item.path">
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </el-menu-item>
    </el-menu>
  </aside>
</template>

<script setup lang="ts">
import {
  ChatDotRound,
  Connection,
  Document,
  Files,
  HomeFilled,
  House,
  Key,
  MessageBox,
  Operation,
  SetUp,
  User,
  UserFilled
} from '@element-plus/icons-vue'
import { computed } from 'vue'

import { useAuthStore, type AdminRole } from '@/stores/auth'

interface MenuItem {
  path: string
  label: string
  icon: unknown
  roles?: AdminRole[]
}

const auth = useAuthStore()

const menus: MenuItem[] = [
  { path: '/admin/dashboard', label: '仪表盘', icon: HomeFilled },
  { path: '/admin/users', label: '用户管理', icon: User },
  { path: '/admin/families', label: '家庭管理', icon: House },
  { path: '/admin/public-applications', label: '公开申请审核', icon: Document },
  { path: '/admin/visitor-messages', label: '游客留言审核', icon: ChatDotRound },
  { path: '/admin/content', label: '内容中心', icon: Files },
  { path: '/admin/founder-transfer-requests', label: '创始人转让审核', icon: Connection, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
  { path: '/admin/dissolution-requests', label: '家庭解散审核', icon: MessageBox, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
  { path: '/admin/admin-users', label: '后台管理员管理', icon: UserFilled, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
  { path: '/admin/operation-logs', label: '操作日志', icon: Operation },
  { path: '/admin/families/family_001/members', label: '成员管理样例', icon: SetUp },
  { path: '/admin/families/family_001', label: '家庭详情样例', icon: Key }
]

const visibleMenus = computed(() => {
  const role = auth.admin?.role
  return menus.filter((item) => !item.roles || (role && item.roles.includes(role)))
})
</script>

<style scoped>
.sidebar {
  min-height: 100vh;
  border-right: 1px solid var(--color-border);
  background: #17263e;
  color: #fff;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 64px;
  padding: 0 18px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}

.brand-mark {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  border-radius: 8px;
  background: var(--color-warm-gold);
  color: #17263e;
  font-weight: 700;
}

.brand span {
  display: block;
  margin-top: 2px;
  color: rgba(255, 255, 255, 0.68);
  font-size: 12px;
}

.menu {
  border-right: 0;
  background: transparent;
}

:deep(.el-menu-item) {
  color: rgba(255, 255, 255, 0.78);
}

:deep(.el-menu-item.is-active),
:deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}
</style>
