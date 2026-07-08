<template>
  <view class="archive-page manage-center-page">
    <MiniBackHome />

    <MiniFamilyPageSkeleton v-if="loading && !family" variant="manage" />

    <view v-else-if="errorMessage && !family" class="manage-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </view>

    <template v-else-if="family">
      <view class="manage-head archive-page-head">
        <view>
          <text class="archive-kicker">Family Admin</text>
          <text class="archive-title">家庭管理</text>
          <text class="archive-subtitle">
            {{ family.familyName }} · 申请、邀请与公开展示权限
          </text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="manage-context">
        <text class="archive-chip">{{ roleText(family.role) }}</text>
        <text class="context-back" @click="openFamilyDetail">返回详情</text>
      </view>

      <view v-if="!canManageFamily" class="manage-state archive-form-panel">
        <MiniNotice tone="warm" title="无管理权限">只有家庭创建者或家庭管理员可以进入家庭管理。</MiniNotice>
        <MiniButton variant="secondary" @click="openFamilyDetail">返回家庭详情</MiniButton>
      </view>

      <view v-else class="manage-directory archive-list">
        <view class="archive-row" @click="openJoinRequests">
          <view class="archive-row-main">
            <text class="archive-row-title">收到的加入申请</text>
            <text class="archive-row-desc">查看并处理申请加入该家庭的记录</text>
          </view>
          <text class="archive-row-meta">申请</text>
          <text class="archive-arrow">›</text>
        </view>
        <view class="archive-row" @click="openSentInvitations">
          <view class="archive-row-main">
            <text class="archive-row-title">发出的成员邀请</text>
            <text class="archive-row-desc">查看、取消或重新生成成员邀请</text>
          </view>
          <text class="archive-row-meta">邀请</text>
          <text class="archive-arrow">›</text>
        </view>
        <view class="archive-row" @click="openSettings">
          <view class="archive-row-main">
            <text class="archive-row-title">公开展示与权限</text>
            <text class="archive-row-desc">公开信息、管理员与高风险操作</text>
          </view>
          <text class="archive-row-meta">权限</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

async function loadFamily() {
  session.restoreSession()
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/manage-center?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = !family.value
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    family.value = await getFamilyDetail(familyId.value)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭管理加载失败。')
  } finally {
    loading.value = false
  }
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  family.value = null
}

function openFamilyDetail() {
  if (getCurrentPages().length > 1) {
    uni.navigateBack({ delta: 1 })
    return
  }
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function openJoinRequests() {
  uni.navigateTo({ url: `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}` })
}

function openSentInvitations() {
  uni.navigateTo({ url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}` })
}

function openSettings() {
  uni.navigateTo({ url: `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}` })
}

function roleText(role: string) {
  return ({ FOUNDER: '创建者', FAMILY_ADMIN: '家庭管理员', MEMBER: '普通成员' } as Record<string, string>)[role] || '成员'
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
})
onShow(loadFamily)
onUnload(resetPageData)
</script>

<style scoped>
.manage-center-page {
  padding-top: 28rpx;
}

.manage-center-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.manage-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.manage-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 24rpx;
}

.context-back {
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.4;
}

.context-back:active {
  color: var(--archive-cinnabar);
}

.manage-directory {
  margin-top: 4rpx;
}
</style>
