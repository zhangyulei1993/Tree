<template>
  <view class="tree-page showcase-page">
    <MiniCard class="showcase-panel-card">
      <view class="showcase-head">
        <text class="showcase-kicker">展示家庭</text>
        <text class="showcase-title">已审核公开的家庭主页</text>
        <text class="showcase-desc">浏览经平台审核的家庭简介与公开家谱，加入家庭请通过成员邀请。</text>
      </view>
    </MiniCard>

    <MiniCard v-if="loading && items.length === 0">
      <MiniEmptyState symbol="展" title="正在加载" description="正在读取展示家庭列表，请稍候。" />
    </MiniCard>

    <MiniCard v-else-if="errorMessage">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="reload">重新加载</MiniButton>
    </MiniCard>

    <template v-else>
      <view v-if="items.length > 0" class="result-panel">
        <text class="result-count">共 {{ total }} 个展示家庭</text>
        <FamilyMiniCard
          v-for="family in items"
          :key="family.id"
          :name="family.familyName"
          :surname="family.familySurname"
          :region="familyRegionLabel(family)"
          :desc="familyCardDesc(family)"
          @click="openFamily(family.id)"
        />
      </view>

      <MiniCard v-else>
        <MiniEmptyState
          symbol="展"
          title="暂无展示家庭"
          description="当前没有可浏览的展示家庭。你可以通过家人分享的链接访问家庭主页。"
        />
      </MiniCard>

      <MiniButton
        v-if="hasMore"
        variant="secondary"
        class="load-more-btn"
        :loading="loadingMore"
        :disabled="loadingMore"
        @click="loadMore"
      >
        加载更多
      </MiniButton>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listPublicFamilyShowcase } from '@/api/families'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyMiniCard from '@/components/family/FamilyMiniCard.vue'
import type { PublicFamilyShowcaseItem } from '@/types/api'

const items = ref<PublicFamilyShowcaseItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const loadingMore = ref(false)
const errorMessage = ref('')

const hasMore = computed(() => items.value.length < total.value)

async function fetchPage(nextPage: number, append: boolean) {
  if (append) {
    loadingMore.value = true
  } else {
    loading.value = true
    errorMessage.value = ''
  }
  try {
    const result = await listPublicFamilyShowcase({ page: nextPage, pageSize })
    total.value = result.total
    page.value = result.page
    items.value = append ? [...items.value, ...result.items] : result.items
  } catch (error) {
    if (!append) {
      items.value = []
      total.value = 0
    }
    errorMessage.value = apiErrorMessage(error, '展示家庭列表加载失败。')
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function reload() {
  page.value = 1
  fetchPage(1, false)
}

function loadMore() {
  if (!hasMore.value || loadingMore.value) return
  fetchPage(page.value + 1, true)
}

function openFamily(familyId: number | string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(String(familyId))}`
  })
}

function familyRegionLabel(family: PublicFamilyShowcaseItem) {
  return family.regionText || family.nativePlace || undefined
}

function familyCardDesc(family: PublicFamilyShowcaseItem) {
  return family.description || '该家庭暂未填写公开简介。'
}

reload()
</script>

<style scoped>
.showcase-page {
  min-height: auto;
  padding-bottom: calc(180rpx + env(safe-area-inset-bottom));
}

.showcase-panel-card {
  margin-bottom: 20rpx;
}

.showcase-head {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.showcase-kicker,
.showcase-title,
.showcase-desc {
  display: block;
}

.showcase-kicker {
  color: var(--tree-green);
  font-size: 22rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.showcase-title {
  color: var(--tree-text-primary);
  font-size: 34rpx;
  font-weight: 800;
}

.showcase-desc {
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.7;
}

.result-panel {
  margin-bottom: 18rpx;
}

.result-count {
  display: block;
  margin: 0 8rpx 16rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}

.load-more-btn {
  margin-top: 8rpx;
}
</style>
