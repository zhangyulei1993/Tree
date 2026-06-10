<template>
  <main class="login-page">
    <section class="login-panel surface">
      <div class="login-brand">
        <div class="brand-mark">T</div>
        <h1>Tree 管理后台</h1>
        <p>{{ apiMode === 'real' ? '使用管理员账号登录本地联调环境。' : '静态原型使用 mock 登录。' }}</p>
      </div>
      <el-form label-position="top" @submit.prevent="login">
        <el-form-item v-if="apiMode === 'mock'" label="管理员角色">
          <el-select v-model="role">
            <el-option label="ROOT_ADMIN" value="ROOT_ADMIN" />
            <el-option label="SUPER_ADMIN" value="SUPER_ADMIN" />
            <el-option label="PLATFORM_ADMIN" value="PLATFORM_ADMIN" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model.trim="username" :disabled="submitting" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="password"
            type="password"
            :disabled="submitting"
            autocomplete="current-password"
            show-password
            @keyup.enter="login"
          />
        </el-form-item>
        <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />
        <el-button
          native-type="submit"
          type="primary"
          size="large"
          class="login-button"
          :loading="submitting"
          :disabled="apiMode === 'real' && (!username || !password)"
        >
          登录
        </el-button>
      </el-form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiMode, getApiErrorMessage } from '@/api/client'
import { useAuthStore, type AdminRole } from '@/stores/auth'

const role = ref<AdminRole>('ROOT_ADMIN')
const username = ref('')
const password = ref('')
const submitting = ref(false)
const errorMessage = ref('')
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

async function login() {
  if (submitting.value) return
  errorMessage.value = ''
  submitting.value = true
  try {
    if (apiMode === 'real') {
      await auth.login(username.value, password.value)
    } else {
      auth.mockLogin(role.value)
    }
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/admin/dashboard'
    await router.replace(redirect.startsWith('/admin/') ? redirect : '/admin/dashboard')
  } catch (error) {
    errorMessage.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
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
  margin-top: 16px;
}
</style>
