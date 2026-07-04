<template>
  <view class="tree-page public-profile-page">
    <MiniBackHome />
    <MiniCard v-if="loadingFamily">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载公开家庭信息..." />
    </MiniCard>

    <MiniCard v-else-if="familyError">
      <MiniSectionHeader title="公开家庭" />
      <MiniNotice tone="warm" title="加载失败">{{ familyError }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </MiniCard>

    <template v-else-if="family">
      <MiniCard variant="hero" class="public-hero tree-pedigree-watermark">
        <text class="tree-public-eyebrow">公开家庭主页</text>
        <view class="tree-dossier-head">
          <view class="tree-dossier-seal">{{ family.familySurname.slice(0, 1) }}</view>
          <view class="tree-dossier-copy">
            <text class="public-surname">{{ family.familySurname }}氏</text>
            <text class="tree-page-title">{{ family.familyName }}</text>
          </view>
        </view>
        <text class="public-desc">{{ family.description || '该家庭暂未填写公开简介。' }}</text>
        <view class="tree-archive-ribbon">
          <text v-if="family.nativePlace" class="tree-archive-chip">{{ family.nativePlace }}</text>
          <text v-if="family.regionText" class="tree-archive-chip green">{{ family.regionText }}</text>
        </view>
        <view class="hero-actions">
          <MiniButton @click="openPublicTree">查看公开家谱</MiniButton>
        </view>
      </MiniCard>

      <view v-if="family.publicContactVisible && hasPublicContact" class="tree-space contact-space">
        <view class="tree-space-head">
          <text class="tree-space-title">联系方式</text>
          <text class="tree-space-subtitle">家庭选择对外公开的联络方式</text>
        </view>
        <view class="tree-space-body tree-contact-card">
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
          <view class="info-block">
            <text class="tree-muted contact-hint">请通过家庭成员分享的邀请联系。</text>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, onMounted, onUnmounted, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPublicFamilyDetail } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
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
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.public-hero {
  margin-bottom: 26rpx;
}

.public-hero :deep(.tree-page-title),
.public-hero .tree-page-title {
  color: #fff;
}

.public-surname {
  display: block;
  margin-bottom: 8rpx;
  color: rgba(248, 231, 194, 0.92);
  font-size: 24rpx;
  font-weight: 600;
  letter-spacing: 2rpx;
}

.public-desc {
  display: block;
  margin-top: 12rpx;
  color: rgba(255, 255, 255, 0.76);
  font-size: 26rpx;
  line-height: 1.7;
}

.hero-actions {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14rpx;
  margin-top: 22rpx;
}

.info-block {
  padding: 14rpx 0;
  border-top: 1rpx solid var(--tree-border-subtle);
}
</style>
