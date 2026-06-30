<template>
  <view class="tree-page search-page">
    <MiniCard class="search-panel-card">
      <view class="search-head">
        <text class="search-kicker">公开家庭</text>
        <text class="search-title">查找可公开浏览的家庭主页</text>
        <text class="search-desc">按家庭名称、姓氏、地区或籍贯筛选。</text>
      </view>

      <text class="tree-field-label">关键词</text>
      <input
        v-model="keyword"
        class="tree-input"
        placeholder="家庭名称 / 姓氏 / 地区 / 籍贯"
        confirm-type="search"
        @confirm="runSearch"
      />

      <text class="tree-field-label">姓氏筛选</text>
      <input v-model="familySurname" class="tree-input" placeholder="例如：张" />

      <text class="tree-field-label">地区筛选</text>
      <input v-model="regionText" class="tree-input" placeholder="例如：山东" />

      <view class="action-row">
        <MiniButton class="action-btn" @click="runSearch">搜索</MiniButton>
        <MiniButton variant="secondary" class="action-btn" @click="refreshList">刷新</MiniButton>
      </view>
    </MiniCard>

    <view v-if="loading && items.length === 0" class="state-wrap">
      <MiniEmptyState symbol="寻" title="正在加载" description="正在读取公开家庭列表，请稍候。" />
    </view>

    <view v-else-if="errorMessage" class="state-wrap">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" class="retry-btn" @click="refreshList">重新加载</MiniButton>
    </view>

    <template v-else>
      <view class="result-meta">
        <text class="result-count">共 {{ total }} 个公开家庭</text>
      </view>

      <MiniCard v-if="items.length === 0">
        <MiniEmptyState
          symbol="寻"
          :title="hasActiveFilters ? '未找到匹配家庭' : '暂无公开家庭'"
          :description="hasActiveFilters ? '请尝试其他关键词，或通过邀请链接访问家庭。' : '当前没有可浏览的公开家庭。'"
        />
      </MiniCard>

      <FamilyMiniCard
        v-for="family in items"
        :key="family.id"
        :name="family.familyName"
        :surname="family.familySurname"
        :region="familyRegionLabel(family)"
        :desc="familyCardDesc(family)"
        @click="openFamily(family.id)"
      />

      <view v-if="items.length > 0 && hasMore" class="load-more-wrap">
        <MiniButton
          variant="secondary"
          :loading="loadingMore"
          @click="loadMore"
        >
          加载更多
        </MiniButton>
      </view>

      <view v-else-if="items.length > 0" class="end-hint">
        <text class="end-text">已显示全部结果</text>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { listPublicFamilies } from '@/api/families'
import { apiErrorMessage } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyMiniCard from '@/components/family/FamilyMiniCard.vue'
import type { PublicFamilyListItem } from '@/types/api'

const keyword = ref('')
const familySurname = ref('')
const regionText = ref('')
const items = ref<PublicFamilyListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const loadingMore = ref(false)
const errorMessage = ref('')

const hasActiveFilters = computed(
  () => Boolean(keyword.value.trim() || familySurname.value.trim() || regionText.value.trim())
)
const hasMore = computed(() => items.value.length < total.value)

function currentQuery() {
  return {
    keyword: keyword.value.trim() || undefined,
    familySurname: familySurname.value.trim() || undefined,
    regionText: regionText.value.trim() || undefined,
    page: page.value,
    pageSize
  }
}

async function fetchPage(append: boolean) {
  if (append) {
    loadingMore.value = true
  } else {
    loading.value = true
    errorMessage.value = ''
  }

  try {
    const result = await listPublicFamilies(currentQuery())
    total.value = result.total
    items.value = append ? [...items.value, ...result.items] : result.items
  } catch (error) {
    if (!append) {
      items.value = []
      total.value = 0
    }
    errorMessage.value = apiErrorMessage(error, '公开家庭列表加载失败。')
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function runSearch() {
  page.value = 1
  fetchPage(false)
}

function refreshList() {
  page.value = 1
  fetchPage(false)
}

function loadMore() {
  if (loading.value || loadingMore.value || !hasMore.value) return
  page.value += 1
  fetchPage(true)
}

function familyRegionLabel(family: PublicFamilyListItem) {
  return [family.nativePlace, family.regionText].filter(Boolean).join(' · ') || undefined
}

function familyCardDesc(family: PublicFamilyListItem) {
  const parts: string[] = []
  if (family.description) parts.push(family.description)
  if (family.publicContactVisible) {
    const contactBits = [family.publicContactName, family.publicContactNote].filter(Boolean)
    if (contactBits.length > 0) parts.push(`联系：${contactBits.join(' · ')}`)
  }
  return parts.join('\n') || undefined
}

function openFamily(familyId: number | string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(String(familyId))}`
  })
}

onMounted(() => {
  refreshList()
})
</script>

<style scoped>
.search-page {
  padding-bottom: 32rpx;
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}
.search-panel-card {
  border-color: rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 100% 0%, rgba(216, 175, 104, 0.16), transparent 200rpx),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(244, 250, 247, 0.94) 100%);
  box-shadow: 0 18rpx 44rpx rgba(24, 54, 83, 0.08);
}
.search-head {
  margin-bottom: 24rpx;
}
.search-kicker,
.search-title,
.search-desc {
  display: block;
}
.search-kicker {
  color: var(--tree-green);
  font-size: 21rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}
.search-title {
  margin-top: 8rpx;
  color: var(--tree-text-primary);
  font-size: 36rpx;
  font-weight: 800;
  line-height: 1.32;
}
.search-desc {
  margin-top: 8rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.55;
}
.action-row {
  display: flex;
  gap: 16rpx;
  margin-top: 20rpx;
}
.action-btn {
  flex: 1;
}
.state-wrap {
  margin-top: 16rpx;
}
.retry-btn {
  margin-top: 20rpx;
}
.result-meta {
  padding: 8rpx 8rpx 14rpx;
}
.result-count {
  display: inline-flex;
  border-radius: 999rpx;
  background: rgba(24, 54, 83, 0.06);
  color: var(--tree-primary);
  padding: 8rpx 18rpx;
  font-size: 23rpx;
  font-weight: 700;
}
.load-more-wrap {
  padding: 8rpx 0 24rpx;
}
.end-hint {
  padding: 8rpx 0 24rpx;
  text-align: center;
}
.end-text {
  color: var(--tree-text-weak);
  font-size: 22rpx;
}
</style>
