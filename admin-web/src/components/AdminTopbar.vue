<template>
  <header class="topbar">
    <div>
      <strong>{{ route.meta.title || '管理后台' }}</strong>
      <span>Tree 家脉亲缘平台</span>
    </div>
    <div class="actions">
      <el-select v-if="apiMode === 'mock'" v-model="selectedRole" size="small" style="width: 160px" @change="onRoleChange">
        <el-option label="ROOT_ADMIN" value="ROOT_ADMIN" />
        <el-option label="SUPER_ADMIN" value="SUPER_ADMIN" />
        <el-option label="PLATFORM_ADMIN" value="PLATFORM_ADMIN" />
      </el-select>
      <span v-if="auth.admin" class="admin-name">{{ auth.admin.displayName || auth.admin.username }}</span>
      <RoleTag v-if="auth.admin" :role="auth.admin.role" />
      <el-button text :loading="loggingOut" @click="logout">退出</el-button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiMode, getApiErrorMessage } from '@/api/client'
import RoleTag from '@/components/RoleTag.vue'
import { useAuthStore, type AdminRole } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const selectedRole = ref<AdminRole>(auth.admin?.role || 'ROOT_ADMIN')
const loggingOut = ref(false)

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

async function logout() {
  if (loggingOut.value) return
  loggingOut.value = true
  try {
    await auth.logout()
  } catch (error) {
    ElMessage.error(`退出失败：${getApiErrorMessage(error)}`)
  } finally {
    await router.replace('/admin/login')
    loggingOut.value = false
  }
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

.admin-name {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 13px;
}
</style>
