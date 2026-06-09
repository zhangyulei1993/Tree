<template>
  <PageShell>
    <section class="container auth-page">
      <form class="card auth-card" @submit.prevent="login">
        <h1>手机号登录</h1>
        <p class="muted">mock 登录不会请求真实后端，也不会写入真实 token。</p>
        <input class="field" placeholder="138****0000" />
        <input class="field" type="password" placeholder="密码 UI 示例" />
        <button class="button">登录</button>
        <button class="button secondary" type="button" @click="session.mockWechatLogin()">模拟微信登录未绑手机号</button>
        <RouterLink to="/register">还没有账号？去注册</RouterLink>
      </form>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const router = useRouter()

function login() {
  session.mockPhoneLogin()
  router.push('/me')
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
  gap: 14px;
  padding: 24px;
}
</style>
