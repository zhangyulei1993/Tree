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
      <el-menu-item-group v-for="group in visibleGroups" :key="group.label" :title="group.label">
        <el-menu-item v-for="item in group.items" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu-item-group>
    </el-menu>
  </aside>
</template>

<script setup lang="ts">
import {
  CollectionTag,
  Connection,
    Document,
  Files,
  HomeFilled,
  House,
  FolderOpened,
  MessageBox,
  Operation,
  SetUp,
  User,
  UserFilled,
    Wallet
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

const menuGroups: { label: string; items: MenuItem[] }[] = [
  { label: '业务总览', items: [
    { path: '/admin/dashboard', label: '仪表盘', icon: HomeFilled },
    { path: '/admin/users', label: '用户管理', icon: User },
    { path: '/admin/families', label: '家庭管理', icon: House }
  ] },
  { label: '内容与运营', items: [
    { path: '/admin/content', label: '内容中心', icon: Files },
    { path: '/admin/share-configs', label: '分享配置', icon: Connection },
    { path: '/admin/showcase-families', label: '帝王世家配置', icon: CollectionTag, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/tool-configs', label: '常用工具配置', icon: SetUp, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] }
  ] },
  { label: '审核与风控', items: [
    { path: '/admin/public-applications', label: '优质家庭审核', icon: Document },
    { path: '/admin/quality-family-config', label: '优质家庭准入规则', icon: SetUp, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/founder-transfer-requests', label: '创始人转让审核', icon: Connection, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/dissolution-requests', label: '家庭解散审核', icon: MessageBox, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/recovery-center', label: '变更恢复中心', icon: FolderOpened, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/operation-logs', label: '操作日志', icon: Operation }
  ] },
  { label: '平台设置', items: [
    { path: '/admin/account-quota-configs', label: '账号权益配置', icon: SetUp, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/capacity-commerce', label: '容量商品与订单', icon: Wallet, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/backup-operation', label: '备份与存储', icon: FolderOpened, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] },
    { path: '/admin/admin-users', label: '后台管理员管理', icon: UserFilled, roles: ['ROOT_ADMIN', 'SUPER_ADMIN'] }
  ] }
]

const visibleGroups = computed(() => {
  const role = auth.admin?.role
  return menuGroups
    .map((group) => ({ ...group, items: group.items.filter((item) => !item.roles || (role && item.roles.includes(role))) }))
    .filter((group) => group.items.length > 0)
})
</script>

<style scoped>
.sidebar {
  min-height: 100vh;
  position: sticky;
  top: 0;
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  background:
    radial-gradient(circle at 0% 0%, rgba(59, 110, 168, 0.18), transparent 200px),
    linear-gradient(180deg, #15243a 0%, #122033 100%);
  color: #fff;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 72px;
  padding: 0 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}

.brand-mark {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border-radius: 12px;
  background: linear-gradient(135deg, #f2d39a 0%, var(--color-warm-gold) 100%);
  color: #17263e;
  font-weight: 700;
  box-shadow: 0 10px 24px rgba(200, 164, 93, 0.22);
}

.brand strong {
  letter-spacing: 0.01em;
}

.brand span {
  display: block;
  margin-top: 2px;
  color: rgba(255, 255, 255, 0.68);
  font-size: 12px;
}

.menu {
  padding: 16px 10px 20px;
  border-right: 0;
  background: transparent;
}

:deep(.el-menu-item-group__title) {
  padding: 14px 10px 7px;
  color: rgba(255, 255, 255, 0.44);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

:deep(.el-menu-item) {
  height: 46px;
  margin-bottom: 6px;
  border-radius: 14px;
  color: rgba(255, 255, 255, 0.78);
  transition:
    background 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}

:deep(.el-menu-item.is-active),
:deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  transform: translateX(2px);
}

:deep(.el-menu-item.is-active) {
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
}

@media (max-width: 960px) {
  .sidebar {
    min-height: auto;
    position: static;
  }
}
</style>
