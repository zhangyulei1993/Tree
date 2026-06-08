<template>
  <header class="topbar">
    <div>
      <strong>{{ route.meta.title || '管理后台' }}</strong>
      <span>Tree 家脉亲缘平台</span>
    </div>
    <div class="actions">
      <el-select v-model="selectedRole" size="small" style="width: 160px" @change="onRoleChange">
        <el-option label="ROOT_ADMIN" value="ROOT_ADMIN" />
        <el-option label="SUPER_ADMIN" value="SUPER_ADMIN" />
        <el-option label="PLATFORM_ADMIN" value="PLATFORM_ADMIN" />
      </el-select>
      <RoleTag v-if="auth.admin" :role="auth.admin.role" />
      <el-button text @click="logout">退出</el-button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import RoleTag from '@/components/RoleTag.vue'
import { useAuthStore, type AdminRole } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const selectedRole = ref<AdminRole>(auth.admin?.role || 'ROOT_ADMIN')

watch(
  () => auth.admin?.role,
  (role) => {
    if (role) selectedRole.value = role
  }
)

function onRoleChange(role: AdminRole) {
  auth.switchRole(role)
  if (route.meta.roles && !route.meta.roles.includes(role)) {
    router.push('/admin/no-permission')
  }
}

function logout() {
  auth.logout()
  router.push('/admin/login')
}
</script>

<style scoped>
.topbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 24px;
  border-bottom: 1px solid var(--color-border);
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(10px);
}

.topbar span {
  display: block;
  margin-top: 3px;
  color: var(--color-text-secondary);
  font-size: 12px;
}

.actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
