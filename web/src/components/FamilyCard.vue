<template>
  <article class="card family-card">
    <div class="card-head">
      <div class="seal">{{ family.surname.slice(0, 1) }}</div>
      <div>
        <span class="status-pill">{{ family.publicStatus === 'APPROVED' ? '已公开' : '未公开' }}</span>
        <h3>{{ family.name }}</h3>
        <p>{{ family.description }}</p>
      </div>
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
  position: relative;
  gap: 16px;
  min-height: 100%;
  padding: 24px;
  overflow: hidden;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.family-card:hover {
  border-color: rgba(47, 107, 87, 0.24);
  box-shadow: var(--shadow-soft);
  transform: translateY(-2px);
}

.family-card::after {
  content: '';
  position: absolute;
  right: -38px;
  top: -42px;
  width: 140px;
  height: 140px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(47, 107, 87, 0.09), transparent 66%);
  pointer-events: none;
}

.card-head {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.seal {
  display: grid;
  flex: 0 0 58px;
  width: 58px;
  height: 58px;
  place-items: center;
  border-radius: 20px;
  background: linear-gradient(145deg, #eef7f2 0%, #f8f3e8 100%);
  color: var(--color-primary);
  font-size: 24px;
  font-weight: 900;
}

h3 {
  margin: 10px 0 8px;
  color: var(--color-primary);
  font-size: 22px;
  letter-spacing: -0.01em;
}

p {
  margin: 0;
  line-height: 1.7;
}

dl {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  border: 1px solid rgba(31, 58, 95, 0.05);
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(247, 250, 248, 0.9) 0%, rgba(251, 250, 246, 0.9) 100%);
  padding: 14px;
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
  align-items: center;
}

@media (max-width: 760px) {
  .card-head {
    align-items: stretch;
    flex-direction: column;
  }

  dl {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
