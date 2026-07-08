<template>
  <view class="archive-page affairs-page">
    <MiniBackHome />

    <view class="affairs-head archive-page-head">
      <view>
        <text class="archive-kicker">Family Affairs</text>
        <text class="archive-title">家庭事务</text>
        <text class="archive-subtitle">处理收到的邀请，并查看自己提交的加入申请</text>
      </view>
      <view class="archive-seal">务</view>
    </view>

    <view class="affairs-summary">
      <text class="archive-chip">待处理邀请 {{ pendingInvitationCount }}</text>
      <text class="archive-chip">待审核申请 {{ pendingRequestCount }}</text>
    </view>

    <view class="affairs-directory archive-list">
      <view class="archive-row" @click="openMyFamilies">
        <view class="archive-row-main">
          <text class="archive-row-title">我的家庭</text>
          <text class="archive-row-desc">查看和管理已加入的家庭</text>
        </view>
        <text class="archive-row-meta">家庭</text>
        <text class="archive-arrow">›</text>
      </view>
    </view>

    <view class="archive-segment-tabs">
      <view
        class="archive-segment-tab"
        :class="{ active: activeTab === 'invitations' }"
        @click="activeTab = 'invitations'"
      >
        <text>收到的邀请</text>
        <text v-if="pendingInvitationCount" class="tab-count">{{ pendingInvitationCount }}</text>
      </view>
      <view
        class="archive-segment-tab"
        :class="{ active: activeTab === 'requests' }"
        @click="activeTab = 'requests'"
      >
        <text>我提交的申请</text>
        <text v-if="pendingRequestCount" class="tab-count">{{ pendingRequestCount }}</text>
      </view>
    </view>

    <view v-if="!authChecked || (loading && invitations.length === 0 && requests.length === 0)" class="affairs-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在同步家庭事务..." />
    </view>

    <view v-else-if="loadError" class="affairs-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" class="affairs-action" @click="loadAffairs">重新加载</MiniButton>
    </view>

    <view v-else-if="activeTab === 'invitations'" key="invitations" class="tab-panel">
      <MiniNotice tone="security" title="核对成员身份">
        接受前请确认家庭名称、成员姓名和邀请发起人。
      </MiniNotice>

      <view v-if="invitations.length === 0" class="archive-form-panel">
        <MiniEmptyState
          symbol="邀"
          title="暂无收到的邀请"
          description="请家人发送家庭邀请或公开家庭分享链接。"
        />
      </view>

      <view v-for="group in invitationGroups" v-else :key="group.key" class="affairs-group archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">{{ group.title }}</text>
          <text class="archive-section-subtitle">{{ group.subtitle }}</text>
        </view>
        <view class="affairs-list archive-list">
          <view v-for="item in group.items" :key="item.invitationId" class="affair-item">
            <view class="affair-item-head">
              <view class="archive-row-main">
                <text class="archive-row-title">{{ item.familyName }}</text>
                <text class="archive-row-desc">成员身份：{{ item.targetMemberName }}</text>
              </view>
              <text
                class="archive-status-tag"
                :class="invitationStatusClass(displayInvitationStatus(item))"
              >
                {{ invitationStatusText(displayInvitationStatus(item)) }}
              </text>
            </view>
            <view class="affair-meta">
              <text>发起人：{{ item.inviterDisplayName || '家庭管理员' }}</text>
              <text v-if="item.inviteMessage">邀请说明：{{ item.inviteMessage }}</text>
              <text class="affair-meta-weak">有效期至：{{ formatDate(item.expiredAt) }}</text>
            </view>
            <view v-if="displayInvitationStatus(item) === 'PENDING'" class="affair-actions">
              <MiniButton
                size="sm"
                :loading="actingInvitationId === item.invitationId && invitationAction === 'accept'"
                :disabled="Boolean(actingInvitationId)"
                @click="confirmAccept(item)"
              >
                接受邀请
              </MiniButton>
              <MiniButton
                size="sm"
                variant="secondary"
                :loading="actingInvitationId === item.invitationId && invitationAction === 'reject'"
                :disabled="Boolean(actingInvitationId)"
                @click="confirmReject(item)"
              >
                拒绝邀请
              </MiniButton>
            </view>
            <view v-else-if="item.status === 'ACCEPTED'" class="affair-actions">
              <MiniButton size="sm" variant="secondary" @click="openMyFamilies">进入我的家庭</MiniButton>
            </view>
          </view>
        </view>
      </view>
    </view>

    <view v-else key="requests" class="tab-panel">
      <MiniNotice tone="security" title="申请进度">
        待审核申请可主动取消；审核通过后可直接进入“我的家庭”。
      </MiniNotice>

      <view v-if="requests.length === 0" class="archive-form-panel">
        <MiniEmptyState
          symbol="申"
          title="暂无加入申请"
          description="你提交的家庭加入申请会显示在这里。请家人发送家庭邀请。"
        />
      </view>

      <view v-for="group in requestGroups" v-else :key="group.key" class="affairs-group archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">{{ group.title }}</text>
          <text class="archive-section-subtitle">{{ group.subtitle }}</text>
        </view>
        <view class="affairs-list archive-list">
          <view v-for="item in group.items" :key="item.requestId" class="affair-item">
            <view class="affair-item-head">
              <view class="archive-row-main">
                <text class="archive-row-title">{{ item.familyName || '未知家庭' }}</text>
                <text class="archive-row-desc">申请理由：{{ item.applicantMessage || '未填写' }}</text>
              </view>
              <text class="archive-status-tag" :class="joinStatusClass(item.requestStatus)">
                {{ joinRequestStatusText(item.requestStatus) }}
              </text>
            </view>
            <view class="affair-meta">
              <text v-if="item.handleComment">审核意见：{{ item.handleComment }}</text>
              <text class="affair-meta-weak">提交时间：{{ formatDate(item.createdAt) }}</text>
            </view>
            <view v-if="item.requestStatus === 'PENDING'" class="affair-actions">
              <MiniButton
                size="sm"
                variant="secondary"
                :loading="cancellingRequestId === item.requestId"
                :disabled="Boolean(cancellingRequestId)"
                @click="confirmCancelRequest(item)"
              >
                取消申请
              </MiniButton>
            </view>
            <view v-else-if="item.requestStatus === 'APPROVED'" class="affair-actions">
              <MiniButton size="sm" variant="secondary" @click="openMyFamilies">进入我的家庭</MiniButton>
            </view>
          </view>
        </view>
      </view>
    </view>

    <text v-if="actionError" class="tree-field-error affairs-error">{{ actionError }}</text>
  </view>
</template>

<script setup lang="ts">
import { onHide, onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { acceptInvitation, listMyInvitations, rejectInvitation } from '@/api/invitations'
import { cancelJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import {
  invitationStatusText,
  joinRequestStatusText
} from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Invitation, JoinRequest } from '@/types/api'

const session = useSessionStore()
const activeTab = ref<'invitations' | 'requests'>('invitations')
const invitations = ref<Invitation[]>([])
const requests = ref<JoinRequest[]>([])
const loading = ref(false)
const authChecked = ref(false)
const loadError = ref('')
const actionError = ref('')
const actingInvitationId = ref<number | string | null>(null)
const invitationAction = ref<'' | 'accept' | 'reject'>('')
const cancellingRequestId = ref<number | string | null>(null)
const navigating = ref(false)

const pendingInvitationCount = computed(() =>
  invitations.value.filter((item) => displayInvitationStatus(item) === 'PENDING').length
)
const pendingRequestCount = computed(() =>
  requests.value.filter((item) => item.requestStatus === 'PENDING').length
)
const invitationGroups = computed(() => {
  const pending = invitations.value.filter((item) => displayInvitationStatus(item) === 'PENDING')
  const history = invitations.value.filter((item) => displayInvitationStatus(item) !== 'PENDING')
  return [
    { key: 'pending', title: '待我处理', subtitle: '需要确认成员身份的邀请', items: pending },
    { key: 'history', title: '历史邀请', subtitle: '已接受、拒绝或失效的记录', items: history }
  ].filter((group) => group.items.length > 0)
})
const requestGroups = computed(() => {
  const pending = requests.value.filter((item) => item.requestStatus === 'PENDING')
  const history = requests.value.filter((item) => item.requestStatus !== 'PENDING')
  return [
    { key: 'pending', title: '待家庭审核', subtitle: '可在审核前取消申请', items: pending },
    { key: 'history', title: '历史申请', subtitle: '已通过、驳回或取消的记录', items: history }
  ].filter((group) => group.items.length > 0)
})

function invitationStatusClass(status: string) {
  if (status === 'PENDING') return 'is-cinnabar'
  if (status === 'ACCEPTED') return 'is-ink'
  return 'is-muted'
}

function joinStatusClass(status: string) {
  if (status === 'PENDING') return 'is-cinnabar'
  if (status === 'APPROVED') return 'is-ink'
  return 'is-muted'
}

function resetTransientUI() {
  actionError.value = ''
  actingInvitationId.value = null
  invitationAction.value = ''
  cancellingRequestId.value = null
}

function resetPageData() {
  authChecked.value = false
  loading.value = false
  loadError.value = ''
  invitations.value = []
  requests.value = []
  resetTransientUI()
}

async function loadAffairs() {
  if (!session.requireLogin(`/pages/me/family-affairs?tab=${activeTab.value}`)) {
    authChecked.value = false
    return
  }
  authChecked.value = true
  const isInitialLoad = invitations.value.length === 0 && requests.value.length === 0
  loading.value = isInitialLoad
  if (isInitialLoad) {
    loadError.value = ''
    actionError.value = ''
  }
  try {
    const [invitationResult, requestResult] = await Promise.all([
      listMyInvitations(),
      listMyJoinRequests()
    ])
    invitations.value = invitationResult
    requests.value = requestResult
  } catch (error) {
    if (isInitialLoad) {
      loadError.value = apiErrorMessage(error, '家庭事务加载失败。')
    }
  } finally {
    loading.value = false
  }
}

function displayInvitationStatus(item: Invitation) {
  if (item.status === 'PENDING' && new Date(item.expiredAt).getTime() <= Date.now()) return 'EXPIRED'
  return item.status
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN')
}

function confirmAccept(item: Invitation) {
  uni.showModal({
    title: '接受邀请',
    content: `确定接受加入「${item.familyName}」的邀请吗？`,
    success: (result) => {
      if (result.confirm) handleInvitation(item, 'accept')
    }
  })
}

function confirmReject(item: Invitation) {
  uni.showModal({
    title: '拒绝邀请',
    content: `确定拒绝加入「${item.familyName}」的邀请吗？`,
    success: (result) => {
      if (result.confirm) handleInvitation(item, 'reject')
    }
  })
}

async function handleInvitation(item: Invitation, action: 'accept' | 'reject') {
  actingInvitationId.value = item.invitationId
  invitationAction.value = action
  actionError.value = ''
  try {
    if (action === 'accept') await acceptInvitation(item.invitationId)
    else await rejectInvitation(item.invitationId)
    await loadAffairs()
  } catch (error) {
    actionError.value = apiErrorMessage(error, action === 'accept' ? '接受邀请失败。' : '拒绝邀请失败。')
  } finally {
    actingInvitationId.value = null
    invitationAction.value = ''
  }
}

function confirmCancelRequest(item: JoinRequest) {
  uni.showModal({
    title: '取消加入申请',
    content: `确定取消对「${item.familyName || '该家庭'}」的加入申请吗？`,
    success: (result) => {
      if (result.confirm) cancelRequest(item)
    }
  })
}

async function cancelRequest(item: JoinRequest) {
  cancellingRequestId.value = item.requestId
  actionError.value = ''
  try {
    await cancelJoinRequest(item.familyId, item.requestId, { cancelReason: '用户主动取消' })
    await loadAffairs()
  } catch (error) {
    actionError.value = apiErrorMessage(error, '取消加入申请失败。')
  } finally {
    cancellingRequestId.value = null
  }
}

function openMyFamilies() {
  if (navigating.value) return
  navigating.value = true
  uni.switchTab({
    url: '/pages/family/my',
    complete: () => {
      navigating.value = false
    }
  })
}

onLoad((options) => {
  activeTab.value = options?.tab === 'requests' ? 'requests' : 'invitations'
})
onShow(() => {
  session.restoreSession()
  loadAffairs()
})
onHide(resetTransientUI)
onUnload(resetPageData)
</script>

<style scoped>
.affairs-page {
  padding-top: 28rpx;
}

.affairs-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.affairs-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 22rpx;
}

.affairs-directory {
  margin-bottom: 24rpx;
}

.archive-segment-tab {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
}

.tab-count {
  min-width: 28rpx;
  border: 1rpx solid rgba(168, 59, 45, 0.28);
  border-radius: 999rpx;
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
  padding: 0 8rpx;
  font-size: 18rpx;
  line-height: 28rpx;
  text-align: center;
}

.archive-segment-tab.active .tab-count {
  border-color: rgba(168, 59, 45, 0.42);
  background: rgba(168, 59, 45, 0.14);
}

.affairs-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.affairs-page :deep(.mini-notice) {
  margin-bottom: 20rpx;
}

.affairs-group {
  margin-bottom: 24rpx;
}

.affairs-list {
  margin-top: 8rpx;
}

.affair-item {
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.affair-item:last-child {
  border-bottom: 0;
}

.affair-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.archive-status-tag {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx solid var(--archive-line);
  padding: 0 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  line-height: 1.2;
}

.archive-status-tag.is-cinnabar {
  border-color: rgba(168, 59, 45, 0.28);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
}

.archive-status-tag.is-ink {
  border-color: rgba(22, 51, 83, 0.22);
  background: rgba(22, 51, 83, 0.08);
  color: var(--archive-blue);
}

.archive-status-tag.is-muted {
  background: rgba(255, 248, 234, 0.58);
  color: var(--archive-ink-soft);
}

.affair-meta {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.affair-meta-weak {
  color: var(--archive-ink-soft);
  opacity: 0.88;
}

.affair-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.affairs-action {
  margin-top: 16rpx;
  align-self: flex-start;
}

.affairs-error {
  display: block;
  margin-top: 16rpx;
}
</style>
