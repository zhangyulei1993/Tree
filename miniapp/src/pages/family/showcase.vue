<template>
  <view class="archive-page showcase-page">
    <view class="showcase-head archive-page-head">
      <view>
        <text class="archive-kicker">Public Archive</text>
        <text class="archive-title">展示家庭</text>
        <text class="archive-subtitle">浏览经平台审核的家庭简介与公开家庭树</text>
      </view>
      <view class="archive-seal">展</view>
    </view>

    <view class="showcase-notice archive-panel">
      <text class="showcase-notice-label">公开册页</text>
      <text class="showcase-notice-text">加入家庭请通过成员邀请。你可以通过家人分享的链接访问家庭主页。</text>
    </view>

    <view v-if="loading && items.length === 0" class="showcase-state archive-form-panel">
      <MiniEmptyState symbol="展" title="正在加载" description="正在读取展示家庭列表，请稍候。" />
    </view>

    <view v-else-if="errorMessage" class="showcase-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" class="retry-button" @click="reload">重新加载</MiniButton>
    </view>

    <template v-else>
      <view v-if="items.length > 0" class="showcase-list archive-list">
        <view class="showcase-section-head">
          <view>
            <text class="archive-kicker">Catalog</text>
            <text class="showcase-section-title">公开展示</text>
          </view>
          <text class="showcase-count">共 {{ total }} 个</text>
        </view>

        <view
          v-for="family in items"
          :key="family.id"
          class="archive-row showcase-row"
          @click="openFamily(family.id)"
        >
          <text class="archive-surname-stamp">{{ family.familySurname.slice(0, 1) }}</text>
          <view class="archive-row-main">
            <text class="archive-row-title">{{ family.familyName }}</text>
            <text class="archive-row-desc">{{ familyCardDesc(family) }}</text>
            <text v-if="familyRegionLabel(family)" class="showcase-region">{{ familyRegionLabel(family) }}</text>
          </view>
          <text class="archive-row-meta">公开</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>

      <view v-else class="showcase-state archive-form-panel">
        <MiniEmptyState
          symbol="展"
          title="暂无展示家庭"
          description="当前没有可浏览的展示家庭。你可以通过家人分享的链接访问家庭主页。"
        />
      </view>

      <view v-if="hasMore" class="showcase-load-more">
        <text
          class="archive-thin-button"
          :class="{ disabled: loadingMore }"
          @click="loadMore"
        >
          {{ loadingMore ? '加载中…' : '加载更多' }}
        </text>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listPublicFamilyShowcase } from '@/api/families'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
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
  return family.regionText || family.nativePlace || ''
}

function familyCardDesc(family: PublicFamilyShowcaseItem) {
  return family.description || '该家庭暂未填写公开简介。'
}

reload()
</script>

<style scoped>
.showcase-page {
  padding-top: 28rpx;
  padding-bottom: calc(180rpx + env(safe-area-inset-bottom));
}

.showcase-head {
  margin-bottom: 18rpx;
}

.showcase-notice {
  margin-bottom: 22rpx;
  padding: 22rpx 0;
}

.showcase-notice-label {
  display: block;
  margin-bottom: 8rpx;
  color: var(--archive-cinnabar);
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 4rpx;
}

.showcase-notice-text {
  display: block;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.65;
}

.showcase-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.retry-button {
  margin-top: 16rpx;
}

.showcase-section-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 8rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding: 0 0 16rpx;
}

.showcase-section-title {
  display: block;
  margin-top: 4rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 32rpx;
  font-weight: 700;
}

.showcase-count {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

.showcase-row {
  align-items: flex-start;
  padding-top: 22rpx;
  padding-bottom: 22rpx;
}

.showcase-row:active {
  opacity: 0.96;
}

.showcase-region {
  display: block;
  margin-top: 6rpx;
  color: var(--archive-blue);
  font-size: 21rpx;
  line-height: 1.4;
}

.showcase-load-more {
  display: flex;
  justify-content: center;
  margin-top: 24rpx;
}

.showcase-load-more .disabled {
  opacity: 0.55;
}

.showcase-page :deep(.mini-button) {
  border-radius: 0;
  box-shadow: none;
}
</style>
