<template>
  <PageShell>
    <section class="hero">
      <div class="container hero-inner">
        <div>
          <span class="eyebrow">家脉亲缘平台</span>
          <h1>让家族记忆有序保存，也方便宗亲重新找到彼此。</h1>
          <p>Tree 支持公开家庭展示、列表式家族树、邀请确认和加入申请。当前为 M15 静态原型。</p>
          <div class="actions">
            <RouterLink class="button" to="/families">搜索公开家庭</RouterLink>
            <RouterLink class="button secondary" to="/login">登录 / 注册</RouterLink>
          </div>
        </div>
        <div class="card hero-card">
          <strong>公开家族入口</strong>
          <p>查看公开主页、公开树和游客留言，不展示手机号、openid、unionid 或 token。</p>
          <RouterLink class="button ghost" to="/families/family_001/public">查看张氏家族</RouterLink>
        </div>
      </div>
    </section>
    <section class="container section">
      <h2>公开家庭推荐</h2>
      <div class="grid families">
        <FamilyCard v-for="family in publicFamilies.slice(0, 2)" :key="family.id" :family="family" />
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import FamilyCard from '@/components/FamilyCard.vue'
import PageShell from '@/components/PageShell.vue'
import { publicFamilies } from '@/mock/data'
</script>

<style scoped>
.hero {
  padding: 56px 0 32px;
}

.hero-inner {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(280px, 0.8fr);
  gap: 28px;
  align-items: center;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-weight: 700;
}

h1 {
  max-width: 760px;
  margin: 12px 0;
  font-size: clamp(32px, 5vw, 56px);
  line-height: 1.08;
}

p {
  color: var(--color-text-secondary);
  line-height: 1.8;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 24px;
}

.hero-card {
  padding: 24px;
}

.section {
  padding: 24px 0;
}

.families {
  grid-template-columns: repeat(2, 1fr);
}

@media (max-width: 860px) {
  .hero-inner,
  .families {
    grid-template-columns: 1fr;
  }
}
</style>
