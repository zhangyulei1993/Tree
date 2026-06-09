<template>
  <article class="card family-card">
    <div>
      <span class="status">{{ family.publicStatus === 'APPROVED' ? '已公开' : '未公开' }}</span>
      <h3>{{ family.name }}</h3>
      <p>{{ family.description }}</p>
    </div>
    <dl>
      <div><dt>姓氏</dt><dd>{{ family.surname }}</dd></div>
      <div><dt>籍贯</dt><dd>{{ family.nativePlace }}</dd></div>
      <div><dt>地区</dt><dd>{{ family.regionText }}</dd></div>
      <div><dt>创始人</dt><dd>{{ family.founderName }}</dd></div>
    </dl>
    <p class="muted">{{ family.publicContact || '该家庭未设置公开联系方式，如需联系请通过平台协助。' }}</p>
    <div class="actions">
      <RouterLink class="button" :to="`/families/${family.id}/public`">查看公开主页</RouterLink>
      <RouterLink class="button secondary" :to="`/families/${family.id}/tree/public`">查看公开树</RouterLink>
      <RouterLink class="button ghost" :to="`/families/${family.id}/join`">申请加入</RouterLink>
    </div>
  </article>
</template>

<script setup lang="ts">
import type { PublicFamily } from '@/mock/data'

defineProps<{ family: PublicFamily }>()
</script>

<style scoped>
.family-card {
  display: grid;
  gap: 14px;
  padding: 18px;
}

h3 {
  margin: 8px 0;
  font-size: 20px;
}

p {
  margin: 0;
  line-height: 1.7;
}

.status {
  color: var(--color-success);
  font-size: 13px;
  font-weight: 700;
}

dl {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin: 0;
}

dt {
  color: var(--color-text-secondary);
  font-size: 12px;
}

dd {
  margin: 4px 0 0;
  font-weight: 600;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

@media (max-width: 760px) {
  dl {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
