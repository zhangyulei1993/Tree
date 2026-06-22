<template>
  <PageShell>
    <section class="container auth-page">
      <form class="card auth-card" @submit.prevent="login">
        <div class="auth-intro">
          <span class="eyebrow">账号登录</span>
          <h1>回到你的家庭空间</h1>
          <p>登录后可查看家庭、成员、邀请和私有家谱。</p>
          <div class="intro-points">
            <span>家庭资料私密保存</span>
            <span>邀请亲人共同维护</span>
            <span>公开展示需审核</span>
          </div>
        </div>
        <div class="auth-form">
          <h2>手机号登录</h2>
          <p class="muted">请输入本地测试账号或已注册手机号。</p>
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
            体验微信登录流程
          </button>
          <RouterLink to="/register">还没有账号？去注册</RouterLink>
        </div>
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
  min-height: 620px;
  place-items: center;
  padding: 48px 0;
}

.auth-card {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(360px, 0.72fr);
  width: min(920px, 100%);
  gap: 0;
  padding: 0;
  overflow: hidden;
}

.auth-intro {
  display: grid;
  align-content: center;
  padding: 42px;
  background:
    radial-gradient(circle at 88% 18%, rgba(47, 107, 87, 0.14), transparent 160px),
    var(--gradient-hero);
}

.auth-intro h1 {
  margin: 16px 0 14px;
  color: var(--color-primary);
  font-size: clamp(32px, 4vw, 48px);
  letter-spacing: -0.04em;
  line-height: 1.1;
}

.auth-intro p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.8;
}

.intro-points {
  display: grid;
  gap: 10px;
  margin-top: 28px;
}

.intro-points span {
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  padding: 12px 14px;
  color: var(--color-primary);
  font-weight: 800;
}

.auth-form {
  display: grid;
  align-content: center;
  gap: 16px;
  padding: 42px;
}

.auth-form h2 {
  margin: 0;
  color: var(--color-primary);
  font-size: 28px;
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

@media (max-width: 820px) {
  .auth-card {
    grid-template-columns: 1fr;
  }

  .auth-intro,
  .auth-form {
    padding: 28px;
  }
}
</style>
