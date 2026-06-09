<template>
  <PageShell>
    <section class="container profile-page">
      <h1>个人中心</h1>
      <div class="card profile-card">
        <dl>
          <div><dt>昵称</dt><dd>{{ session.user.nickname }}</dd></div>
          <div><dt>手机号</dt><dd>{{ session.user.maskedPhone }}</dd></div>
          <div><dt>登录状态</dt><dd>{{ session.state }}</dd></div>
          <div><dt>账号状态</dt><dd>{{ session.user.status }}</dd></div>
        </dl>
        <div class="actions">
          <button class="button secondary" @click="session.mockBindPhone()">模拟绑定手机号</button>
          <RouterLink class="button" to="/me/families">我的家庭</RouterLink>
          <button class="button ghost" @click="session.logout()">退出 mock 登录</button>
        </div>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
</script>

<style scoped>
.profile-page {
  padding: 32px 0;
}

.profile-card {
  padding: 22px;
}

dl {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

dt {
  color: var(--color-text-secondary);
}

dd {
  margin: 4px 0 0;
  font-weight: 700;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

@media (max-width: 760px) {
  dl {
    grid-template-columns: 1fr;
  }
}
</style>
