<template>
  <main class="login-page">
    <section class="login-panel surface">
      <div class="login-brand">
        <div class="brand-mark">T</div>
        <h1>Tree 管理后台</h1>
        <p>静态原型使用 mock 登录，不连接真实后端。</p>
      </div>
      <el-form label-position="top">
        <el-form-item label="管理员角色">
          <el-select v-model="role">
            <el-option label="ROOT_ADMIN" value="ROOT_ADMIN" />
            <el-option label="SUPER_ADMIN" value="SUPER_ADMIN" />
            <el-option label="PLATFORM_ADMIN" value="PLATFORM_ADMIN" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input model-value="mock_admin" disabled />
        </el-form-item>
        <el-form-item label="密码">
          <el-input model-value="mock_password" disabled show-password />
        </el-form-item>
        <el-button type="primary" size="large" class="login-button" @click="login">进入后台</el-button>
      </el-form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore, type AdminRole } from '@/stores/auth'

const role = ref<AdminRole>('ROOT_ADMIN')
const auth = useAuthStore()
const router = useRouter()

function login() {
  auth.mockLogin(role.value)
  router.push('/admin/dashboard')
}
</script>

<style scoped>
.login-page {
  display: grid;
  min-height: 100vh;
  place-items: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #eef2f7 100%);
}

.login-panel {
  width: min(420px, calc(100vw - 32px));
  padding: 28px;
}

.login-brand {
  margin-bottom: 24px;
  text-align: center;
}

.brand-mark {
  display: grid;
  width: 42px;
  height: 42px;
  margin: 0 auto 14px;
  place-items: center;
  border-radius: 8px;
  background: var(--color-primary);
  color: #fff;
  font-weight: 700;
}

h1 {
  margin: 0;
  font-size: 24px;
}

p {
  margin: 8px 0 0;
  color: var(--color-text-secondary);
}

.login-button {
  width: 100%;
}
</style>
