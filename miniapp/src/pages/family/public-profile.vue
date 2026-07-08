<template>
  <view class="archive-page public-profile-page">
    <MiniBackHome />

    <view v-if="loadingFamily" class="profile-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载公开家庭信息..." />
    </view>

    <view v-else-if="familyError" class="profile-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ familyError }}</MiniNotice>
      <MiniButton variant="secondary" class="retry-button" @click="loadFamily">重新加载</MiniButton>
    </view>

    <template v-else-if="family">
      <view class="profile-head archive-page-head">
        <view>
          <text class="archive-kicker">Public Folio</text>
          <text class="archive-title">公开族谱扉页</text>
          <text class="archive-subtitle">经平台审核后对外展示的家庭简介</text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="public-cover">
        <view class="public-cover-copy">
          <text class="public-surname">{{ family.familySurname }}氏</text>
          <text class="public-name">{{ family.familyName }}</text>
          <text class="public-desc">{{ family.description || '该家庭暂未填写公开简介。' }}</text>
          <view class="public-status-line">
            <text class="public-status-badge">已审核公开</text>
            <text v-if="family.nativePlace" class="archive-chip">{{ family.nativePlace }}</text>
            <text v-if="family.regionText" class="archive-chip">{{ family.regionText }}</text>
          </view>
        </view>
        <view class="public-spine archive-book-spine">
          <text>公</text>
          <text>开</text>
          <text>谱</text>
        </view>
      </view>

      <view class="public-dossier">
        <view v-for="item in dossierItems" :key="item.label" class="dossier-cell">
          <text class="dossier-label">{{ item.label }}</text>
          <text class="dossier-value">{{ item.value }}</text>
        </view>
      </view>

      <view class="public-directory archive-list">
        <view class="archive-row" @click="openPublicTree">
          <view class="archive-row-main">
            <text class="archive-row-title">查看公开家谱</text>
            <text class="archive-row-desc">浏览经脱敏处理的公开家谱结构</text>
          </view>
          <text class="archive-row-meta">家谱</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>

      <view class="public-actions">
        <MiniButton @click="openPublicTree">查看公开家谱</MiniButton>
        <!-- #ifdef MP-WEIXIN -->
        <button class="wechat-share-button" open-type="share">分享公开家庭</button>
        <!-- #endif -->
      </view>

      <view v-if="family.publicContactVisible && hasPublicContact" class="public-contact archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">联系方式</text>
          <text class="archive-section-subtitle">家庭选择对外公开的联络方式</text>
        </view>
        <view v-if="family.publicContactName" class="tree-info-row">
          <text class="tree-info-label">联系人</text>
          <text class="tree-info-value">{{ family.publicContactName }}</text>
        </view>
        <view v-if="family.publicContactPhone" class="tree-info-row">
          <text class="tree-info-label">联系电话</text>
          <text class="tree-info-value">{{ family.publicContactPhone }}</text>
        </view>
        <view v-if="family.publicContactWechat" class="tree-info-row">
          <text class="tree-info-label">微信</text>
          <text class="tree-info-value">{{ family.publicContactWechat }}</text>
        </view>
        <text class="contact-hint">请通过家庭成员分享的邀请联系。</text>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShareAppMessage, onShareTimeline } from '@dcloudio/uni-app'
import { computed, onMounted, onUnmounted, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPublicFamilyDetail } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import {
  buildPublicFamilySharePayload,
  buildPublicFamilyTimelinePayload
} from '@/features/share/wechatShare'
import type { PublicFamily } from '@/types/api'

const familyId = ref('')
const family = ref<PublicFamily | null>(null)
const loadingFamily = ref(false)
const familyError = ref('')
let requestVersion = 0

const hasPublicContact = computed(() => Boolean(
  family.value?.publicContactName ||
  family.value?.publicContactPhone ||
  family.value?.publicContactWechat
))

const dossierItems = computed(() => {
  if (!family.value) return []
  const items = [
    { label: '姓氏', value: `${family.value.familySurname}氏` },
    { label: '家庭名称', value: family.value.familyName },
    { label: '籍贯', value: family.value.nativePlace || '未公开' },
    { label: '地区', value: family.value.regionText || '未公开' }
  ]
  return items
})

function openPublicTree() {
  uni.navigateTo({
    url: `/pages/family/public-tree?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function resetPageState() {
  family.value = null
  familyError.value = ''
  loadingFamily.value = false
}

function familyIdFromCurrentHash() {
  if (typeof window === 'undefined') return ''
  const [, query = ''] = window.location.hash.split('?')
  return new URLSearchParams(query).get('familyId') || ''
}

function isCurrentProfileRoute() {
  if (typeof window === 'undefined') return false
  return window.location.hash.split('?')[0] === '#/pages/family/public-profile'
}

function reloadForFamilyId(nextFamilyId: string) {
  requestVersion += 1
  familyId.value = nextFamilyId.trim()
  resetPageState()

  if (!familyId.value) {
    familyError.value = '家庭公开信息暂不可访问。'
    return
  }

  loadFamily()
}

function handleHashChange() {
  if (!isCurrentProfileRoute()) return
  const nextFamilyId = familyIdFromCurrentHash()
  if (nextFamilyId !== familyId.value) {
    reloadForFamilyId(nextFamilyId)
  }
}

async function loadFamily() {
  if (!familyId.value) {
    family.value = null
    familyError.value = '家庭公开信息暂不可访问。'
    return
  }
  const version = requestVersion
  const currentFamilyId = familyId.value
  family.value = null
  loadingFamily.value = true
  familyError.value = ''
  try {
    const result = await getPublicFamilyDetail(currentFamilyId)
    if (version === requestVersion && currentFamilyId === familyId.value) {
      family.value = result
    }
  } catch (error) {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      family.value = null
      familyError.value = apiErrorMessage(error, '家庭公开信息暂不可访问。')
    }
  } finally {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      loadingFamily.value = false
    }
  }
}

onLoad((options) => {
  reloadForFamilyId(String(options?.familyId || ''))
})

onShareAppMessage(() => buildPublicFamilySharePayload({
  id: family.value?.id || familyId.value,
  familyName: family.value?.familyName || '公开家庭'
}))

onShareTimeline(() => buildPublicFamilyTimelinePayload({
  id: family.value?.id || familyId.value,
  familyName: family.value?.familyName || '公开家庭'
}))

onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener('hashchange', handleHashChange)
  }
})

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('hashchange', handleHashChange)
  }
})
</script>

<style scoped>
.public-profile-page {
  padding-top: 28rpx;
}

.public-profile-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.profile-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.retry-button {
  margin-top: 16rpx;
}

.profile-head {
  margin-bottom: 24rpx;
}

.public-cover {
  display: flex;
  min-height: 240rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.public-cover-copy {
  flex: 1;
  min-width: 0;
  padding: 28rpx 24rpx 28rpx 0;
}

.public-surname {
  display: block;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  font-weight: 650;
  letter-spacing: 3rpx;
}

.public-name {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 42rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  line-height: 1.28;
}

.public-desc {
  display: block;
  margin-top: 14rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.68;
}

.public-status-line {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 18rpx;
}

.public-status-badge {
  display: inline-flex;
  border: 1rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  padding: 5rpx 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 1rpx;
}

.public-spine {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  width: 88rpx;
  flex-shrink: 0;
}

.public-spine text {
  color: rgba(255, 255, 255, 0.92);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 800;
}

.public-dossier {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border-top: 1rpx solid var(--archive-line);
  border-left: 1rpx solid var(--archive-line);
  margin-bottom: 24rpx;
}

.dossier-cell {
  min-height: 108rpx;
  border-right: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 20rpx;
  box-sizing: border-box;
}

.dossier-label,
.dossier-value {
  display: block;
}

.dossier-label {
  color: var(--archive-cinnabar);
  font-size: 21rpx;
  line-height: 1.2;
}

.dossier-value {
  margin-top: 10rpx;
  color: var(--archive-ink);
  font-size: 27rpx;
  font-weight: 700;
  line-height: 1.35;
}

.public-directory {
  margin-bottom: 20rpx;
}

.public-actions {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-bottom: 24rpx;
}

.public-contact {
  margin-bottom: 16rpx;
}

.contact-hint {
  display: block;
  margin-top: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.6;
}

.wechat-share-button {
  margin: 0;
  border: 1rpx solid var(--archive-cinnabar);
  border-radius: 0;
  background: transparent;
  color: var(--archive-cinnabar);
  font-size: 26rpx;
  line-height: 2.2;
}

.wechat-share-button::after {
  border: 0;
}

.public-profile-page :deep(.mini-button) {
  border-radius: 0;
  box-shadow: none;
}
</style>
