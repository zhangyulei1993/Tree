<template>
  <PageShell>
    <section class="container search-page">
      <div class="search-hero">
        <span class="eyebrow">公开家族</span>
        <h1>寻找你的家族</h1>
        <p>浏览已审核公开的家族主页、公开树和联系方式。未公开家庭不会出现在这里。</p>
      </div>
      <div class="card filters">
        <input v-model="keyword" class="field" placeholder="请输入家族名称 / 姓氏 / 籍贯 / 地区" />
        <select v-model="surname" class="field"><option value="">全部姓氏</option><option>张</option><option>李</option><option>王</option></select>
        <select v-model="region" class="field"><option value="">全部地区</option><option>山东</option><option>浙江</option><option>河南</option></select>
        <button class="button">搜索</button>
      </div>
      <p class="muted">搜索结果：共 {{ filtered.length }} 个家庭</p>
      <div class="grid">
        <FamilyCard v-for="family in filtered" :key="family.id" :family="family" />
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import FamilyCard from '@/components/FamilyCard.vue'
import PageShell from '@/components/PageShell.vue'
import { publicFamilies } from '@/mock/data'

const keyword = ref('')
const surname = ref('')
const region = ref('')
const filtered = computed(() => publicFamilies.filter((family) =>
  (!keyword.value || JSON.stringify(family).includes(keyword.value)) &&
  (!surname.value || family.surname === surname.value) &&
  (!region.value || family.regionText.includes(region.value))
))
</script>

<style scoped>
.search-page {
  padding: 54px 0 32px;
}

.search-hero {
  max-width: 760px;
  margin-bottom: 22px;
}

.search-hero h1 {
  margin: 14px 0 10px;
  color: var(--color-primary);
  font-size: clamp(34px, 5vw, 54px);
  letter-spacing: -0.04em;
  line-height: 1.08;
}

.search-hero p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 17px;
  line-height: 1.75;
}

.filters {
  display: grid;
  grid-template-columns: 1.5fr 160px 160px auto;
  gap: 12px;
  margin: 18px 0 20px;
  padding: 18px;
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.08), transparent 170px),
    rgba(255, 255, 255, 0.94);
}

@media (max-width: 820px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
