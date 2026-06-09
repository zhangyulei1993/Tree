<template>
  <PageShell>
    <section class="container auth-page">
      <form class="card auth-card" @submit.prevent="login">
        <div>
          <span class="eyebrow">{{ session.mode === 'real' ? '真实 API' : 'Mock 模式' }}</span>
          <h1>手机号登录</h1>
          <p class="muted">
            {{ session.mode === 'real' ? '使用本地 Tree API 验证账号。' : 'Mock 模式不会请求后端。' }}
          </p>
        </div>
        <label>
          <span>手机号</span>
          <input v-model.trim="phone" class="field" inputmode="numeric" autocomplete="username" placeholder="请输入手机号" />
        </label>
        <label>
          <span>密码</span>
          <input v-model="password" class="field" type="password" autocomplete="current-password" placeholder="请输入密码" />
        </label>
        <p v-if="error" class="feedback error" role="alert">{{ error }}</p>
        <button class="button" :disabled="submitting">
          {{ submitting ? '登录中...' : '登录' }}
        </button>
        <button
          v-if="session.mode === 'mock'"
          class="button secondary"
          type="button"
          @click="mockWechatLogin"
        >
          模拟微信登录未绑手机号
        </button>
        <RouterLink to="/register">还没有账号？去注册</RouterLink>
      </form>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const route = useRoute()
const router = useRouter()
const phone = ref('')
const password = ref('')
const submitting = ref(false)
const error = ref('')

async function login() {
  error.value = ''
  if (!/^1[3-9]\d{9}$/.test(phone.value)) {
    error.value = '请输入正确的手机号。'
    return
  }
  if (password.value.length < 6) {
    error.value = '密码至少需要 6 位。'
    return
  }

  submitting.value = true
  try {
    await session.login({ phone: phone.value, password: password.value, clientType: 'H5_WEB' })
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/me'
    await router.push(redirect)
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '登录失败。')
  } finally {
    submitting.value = false
  }
}

function mockWechatLogin() {
  session.mockWechatLogin()
}
</script>

<style scoped>
.auth-page {
  display: grid;
  min-height: 560px;
  place-items: center;
}

.auth-card {
  display: grid;
  width: min(420px, 100%);
  gap: 16px;
  padding: 24px;
}

.auth-card h1 {
  margin: 8px 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

label {
  display: grid;
  gap: 7px;
  font-size: 14px;
  font-weight: 600;
}

.feedback {
  margin: 0;
  font-weight: 600;
}

.error {
  color: var(--color-danger);
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>
