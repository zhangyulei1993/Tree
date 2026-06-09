<template>
  <PageShell>
    <section class="container my-page">
      <h1>我的家庭</h1>
      <div class="grid">
        <article v-for="family in myFamilies" :key="family.id" class="card item">
          <div>
            <h3>{{ family.name }}</h3>
            <p class="muted">成员：{{ family.memberName }} · 角色：{{ family.role }} · 状态：{{ family.status }}</p>
          </div>
          <RouterLink class="button secondary" :to="`/families/${family.id}/public`">查看</RouterLink>
        </article>
      </div>
      <section class="card item">
        <div>
          <h3>加入申请</h3>
          <p v-for="request in joinRequests" :key="request.id" class="muted">{{ request.familyName }}：{{ request.status }}，{{ request.reason }}</p>
        </div>
      </section>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import PageShell from '@/components/PageShell.vue'
import { joinRequests, myFamilies } from '@/mock/data'
</script>

<style scoped>
.my-page {
  padding: 32px 0;
}

.item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
}

h3 {
  margin: 0 0 8px;
}
</style>
