<template>
  <PageShell>
    <section class="container auth-page">
      <form class="card auth-card" @submit.prevent="submit">
        <h1>注册账号</h1>
        <p class="muted">验证码仅为 UI 占位，不展示真实验证码。</p>
        <input class="field" placeholder="139****0001" />
        <input class="field" placeholder="验证码" />
        <input class="field" placeholder="昵称" />
        <button class="button">完成 mock 注册</button>
      </form>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'

const router = useRouter()
const session = useSessionStore()

function submit() {
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
