<template>
  <view class="archive-page affairs-page">
    <MiniBackHome />

    <view class="affairs-head archive-page-head">
      <view>
        <text class="archive-kicker">Family Affairs</text>
          <text class="archive-title">我的邀请与申请</text>
          <text class="archive-subtitle">处理收到的邀请，查看加入记录</text>
      </view>
      <view class="archive-seal">务</view>
    </view>

    <view v-if="loading && !authChecked" class="affairs-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在同步家庭事务..." />
    </view>

    <view v-else-if="loadError" class="affairs-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" size="sm" @click="loadCounts">重新加载</MiniButton>
    </view>

    <template v-else-if="authChecked">
      <view class="affairs-summary">
        <text v-if="pendingTotal" class="archive-chip affairs-alert">待处理 {{ pendingTotal }}</text>
        <text v-else class="archive-chip">暂无待处理事务</text>
      </view>

      <view class="affairs-directory archive-list">
        <view class="archive-row" @click="openReceivedInvitations">
          <view class="archive-row-main">
            <text class="archive-row-title">收到的家庭邀请</text>
            <text class="archive-row-desc">接受或拒绝家人发来的邀请</text>
          </view>
          <text v-if="pendingInvitationCount" class="affair-badge">{{ pendingInvitationCount }}</text>
          <text v-else class="archive-row-meta">邀请</text>
          <text class="archive-arrow">›</text>
        </view>

        <view class="archive-row" @click="openJoinHistory">
          <view class="archive-row-main">
            <text class="archive-row-title">历史加入申请</text>
            <text class="archive-row-desc">查看早期申请记录与审核结果</text>
          </view>
          <text v-if="pendingRequestCount" class="affair-badge">{{ pendingRequestCount }}</text>
          <text v-else class="archive-row-meta">记录</text>
          <text class="archive-arrow">›</text>
        </view>

        <view class="archive-row" @click="openMyFamilies">
          <view class="archive-row-main">
            <text class="archive-row-title">我的家庭</text>
            <text class="archive-row-desc">查看和管理已加入的家庭</text>
          </view>
          <text class="archive-row-meta">家庭</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listMyInvitations } from '@/api/invitations'
import { listMyJoinRequests } from '@/api/joinRequests'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Invitation, JoinRequest } from '@/types/api'

const session = useSessionStore()
const invitations = ref<Invitation[]>([])
const requests = ref<JoinRequest[]>([])
const loading = ref(false)
const authChecked = ref(false)
const loadError = ref('')
const navigating = ref(false)

const pendingInvitationCount = computed(() =>
  invitations.value.filter((item) => item.status === 'PENDING' && Date.parse(item.expiredAt) > Date.now()).length
)
const pendingRequestCount = computed(() =>
  requests.value.filter((item) => item.requestStatus === 'PENDING').length
)
const pendingTotal = computed(() => pendingInvitationCount.value + pendingRequestCount.value)

function resetPageData() {
  invitations.value = []
  requests.value = []
  loading.value = false
  authChecked.value = false
  loadError.value = ''
  navigating.value = false
}

async function loadCounts() {
  session.restoreSession()
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin('/pages/me/family-affairs')
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    const [invitationResult, requestResult] = await Promise.all([
      listMyInvitations(),
      listMyJoinRequests()
    ])
    invitations.value = invitationResult
    requests.value = requestResult
    authChecked.value = true
  } catch (error) {
    loadError.value = apiErrorMessage(error, '家庭事务加载失败。')
  } finally {
    loading.value = false
  }
}

function openReceivedInvitations() {
  if (navigating.value) return
  navigating.value = true
  uni.navigateTo({ url: '/pages/invite/my', complete: () => { navigating.value = false } })
}

function openJoinHistory() {
  if (navigating.value) return
  navigating.value = true
  uni.navigateTo({ url: '/pages/join/my', complete: () => { navigating.value = false } })
}

function openMyFamilies() {
  if (navigating.value) return
  navigating.value = true
  uni.switchTab({ url: '/pages/family/my', complete: () => { navigating.value = false } })
}

onShow(() => { void loadCounts() })
onUnload(resetPageData)
</script>

<style scoped>
.affairs-page { padding-top: 28rpx; }
.affairs-page :deep(.mini-back-home) { margin-bottom: 22rpx; }
.affairs-summary { display: flex; flex-wrap: wrap; gap: 10rpx; margin-bottom: 22rpx; }
.affairs-alert { border-color: rgba(168, 59, 45, 0.34); background: rgba(168, 59, 45, 0.1); color: var(--archive-cinnabar); }
.affairs-directory { margin-bottom: 24rpx; }
.affairs-state :deep(.mini-notice) { margin-bottom: 18rpx; }
</style>
