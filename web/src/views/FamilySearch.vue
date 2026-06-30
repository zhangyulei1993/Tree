<template>
  <PageShell>
    <section class="container search-page">
      <div class="search-hero">
        <span class="eyebrow">公开家族</span>
        <h1>寻找你的家族</h1>
        <p>浏览已审核公开的家族主页、公开树和联系方式。未公开家庭不会出现在这里。</p>
      </div>

      <form class="card filters" @submit.prevent="runSearch">
        <input v-model.trim="keyword" class="field" placeholder="家族名称 / 姓氏 / 籍贯 / 地区" />
        <input v-model.trim="surname" class="field" placeholder="姓氏，例如：张" />
        <input v-model.trim="region" class="field" placeholder="地区，例如：山东" />
        <button class="button" :disabled="loading">{{ loading ? '搜索中...' : '搜索' }}</button>
      </form>

      <section v-if="loading && families.length === 0" class="card state-panel">正在加载公开家庭...</section>
      <section v-else-if="error" class="card state-panel error" role="alert">
        <strong>公开家庭加载失败</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="runSearch">重新加载</button>
      </section>
      <template v-else>
        <p class="muted">搜索结果：共 {{ total }} 个家庭</p>
        <section v-if="families.length === 0" class="card state-panel">
          {{ hasFilters ? '未找到匹配的公开家庭，请尝试其他关键词。' : '当前暂无公开家庭。' }}
        </section>
        <div v-else class="grid">
          <FamilyCard v-for="family in familyCards" :key="family.id" :family="family" />
        </div>
        <button v-if="hasMore" class="button secondary load-more" :disabled="loading" @click="loadMore">
          {{ loading ? '加载中...' : '加载更多' }}
        </button>
      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listPublicFamilies } from '@/api/families'
import FamilyCard from '@/components/FamilyCard.vue'
import PageShell from '@/components/PageShell.vue'
import type { PublicFamilyListItem } from '@/types/api'

const keyword = ref('')
const surname = ref('')
const region = ref('')
const families = ref<PublicFamilyListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const error = ref('')

const hasFilters = computed(() => Boolean(keyword.value || surname.value || region.value))
const hasMore = computed(() => families.value.length < total.value)
const familyCards = computed(() => families.value.map((family) => ({
  id: String(family.id),
  name: family.familyName,
  surname: family.familySurname,
  nativePlace: family.nativePlace || '未填写',
  regionText: family.regionText || '未填写',
  description: family.description || '该家庭暂未填写简介。',
  publicContact: family.publicContactVisible
    ? [family.publicContactName, family.publicContactNote].filter(Boolean).join(' / ')
    : '',
  publicStatus: 'APPROVED' as const
})))

async function fetchFamilies(append: boolean) {
  loading.value = true
  error.value = ''
  try {
    const result = await listPublicFamilies({
      keyword: keyword.value || undefined,
      familySurname: surname.value || undefined,
      regionText: region.value || undefined,
      page: page.value,
      pageSize
    })
    total.value = result.total
    families.value = append ? [...families.value, ...result.items] : result.items
  } catch (requestError) {
    if (!append) {
      families.value = []
      total.value = 0
    }
    error.value = apiErrorMessage(requestError, '无法加载公开家庭。')
  } finally {
    loading.value = false
  }
}

function runSearch() {
  page.value = 1
  void fetchFamilies(false)
}

function loadMore() {
  if (loading.value || !hasMore.value) return
  page.value += 1
  void fetchFamilies(true)
}

onMounted(runSearch)
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
  letter-spacing: 0;
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
  grid-template-columns: 1.5fr 160px 180px auto;
  gap: 12px;
  margin: 18px 0 20px;
  padding: 18px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: var(--shadow-soft);
}

.state-panel {
  display: grid;
  min-height: 180px;
  place-items: center;
  gap: 12px;
  padding: 28px;
  text-align: center;
}

.state-panel.error {
  color: var(--color-danger);
}

.load-more {
  display: block;
  margin: 22px auto 0;
}

.grid {
  gap: 18px;
}

@media (max-width: 820px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
