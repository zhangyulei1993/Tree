<template>
  <view class="tree-page affairs-page">
    <MiniBackHome />

    <MiniCard variant="hero" class="affairs-hero">
      <text class="hero-kicker">我的家庭事务</text>
      <text class="hero-title">邀请与加入申请</text>
      <text class="hero-desc">在一个页面处理收到的邀请，并查看自己提交的加入申请。</text>
      <view class="summary-row">
        <text class="summary-chip">待处理邀请 {{ pendingInvitationCount }}</text>
        <text class="summary-chip">待审核申请 {{ pendingRequestCount }}</text>
      </view>
    </MiniCard>

    <MiniCard class="directory-card directory-card--list">
      <MiniDirectoryTile
        title="我的家庭"
        desc="查看和管理已加入的家庭"
        icon-class="tree-symbol-home"
        @click="openMyFamilies"
      />
    </MiniCard>

    <view class="affairs-tabs">
      <button :class="{ active: activeTab === 'invitations' }" @click="activeTab = 'invitations'">
        收到的邀请
        <text v-if="pendingInvitationCount" class="tab-count">{{ pendingInvitationCount }}</text>
      </button>
      <button :class="{ active: activeTab === 'requests' }" @click="activeTab = 'requests'">
        我提交的申请
        <text v-if="pendingRequestCount" class="tab-count">{{ pendingRequestCount }}</text>
      </button>
    </view>

    <MiniCard v-if="!authChecked || (loading && invitations.length === 0 && requests.length === 0)">
      <MiniEmptyState symbol="…" title="正在加载" description="正在同步家庭事务..." />
    </MiniCard>

    <MiniCard v-else-if="loadError">
      <MiniEmptyState
        symbol="!"
        title="加载失败"
        :description="loadError"
        action-text="重新加载"
        @action="loadAffairs"
      />
    </MiniCard>

    <view v-else-if="activeTab === 'invitations'" key="invitations" class="tab-panel">
      <MiniNotice tone="security" title="核对成员身份">
        接受前请确认家庭名称、成员姓名和邀请发起人。
      </MiniNotice>
      <MiniCard v-if="invitations.length === 0">
        <MiniEmptyState
          symbol="邀"
          title="暂无收到的邀请"
          description="家人发出的成员绑定邀请会显示在这里。"
          action-text="寻找家庭"
          @action="openSearch"
        />
      </MiniCard>
      <view v-for="group in invitationGroups" v-else :key="group.key" class="affair-group">
        <MiniSectionHeader :title="group.title" :subtitle="group.subtitle" />
        <MiniCard v-for="item in group.items" :key="item.invitationId" variant="soft" class="affair-card invitation-card">
          <view class="item-head">
            <view>
              <text class="item-kicker">家庭邀请</text>
              <text class="item-name">{{ item.familyName }}</text>
            </view>
            <MiniStatusTag :status="displayInvitationStatus(item)" :label="invitationStatusText(displayInvitationStatus(item))" />
          </view>
          <view class="detail-grid">
            <text class="item-meta">成员身份：{{ item.targetMemberName }}</text>
            <text class="item-meta">发起人：{{ item.inviterDisplayName || '家庭管理员' }}</text>
            <text v-if="item.inviteMessage" class="item-meta full">邀请说明：{{ item.inviteMessage }}</text>
            <text class="item-meta full weak">有效期至：{{ formatDate(item.expiredAt) }}</text>
          </view>
          <view v-if="displayInvitationStatus(item) === 'PENDING'" class="action-grid">
            <MiniButton
              :loading="actingInvitationId === item.invitationId && invitationAction === 'accept'"
              :disabled="Boolean(actingInvitationId)"
              @click="confirmAccept(item)"
            >
              接受邀请
            </MiniButton>
            <MiniButton
              variant="secondary"
              :loading="actingInvitationId === item.invitationId && invitationAction === 'reject'"
              :disabled="Boolean(actingInvitationId)"
              @click="confirmReject(item)"
            >
              拒绝邀请
            </MiniButton>
          </view>
          <MiniButton
            v-else-if="item.status === 'ACCEPTED'"
            variant="secondary"
            size="sm"
            @click="openMyFamilies"
          >
            进入我的家庭
          </MiniButton>
        </MiniCard>
      </view>
    </view>

    <view v-else key="requests" class="tab-panel">
      <MiniNotice tone="security" title="申请进度">
        待审核申请可主动取消；审核通过后可直接进入“我的家庭”。
      </MiniNotice>
      <MiniCard v-if="requests.length === 0">
        <MiniEmptyState
          symbol="申"
          title="暂无加入申请"
          description="你在公开家庭主页提交的申请会显示在这里。"
          action-text="寻找家庭"
          @action="openSearch"
        />
      </MiniCard>
      <view v-for="group in requestGroups" v-else :key="group.key" class="affair-group">
        <MiniSectionHeader :title="group.title" :subtitle="group.subtitle" />
        <MiniCard v-for="item in group.items" :key="item.requestId" variant="soft" class="affair-card request-card">
          <view class="item-head">
            <view>
              <text class="item-kicker">加入申请</text>
              <text class="item-name">{{ item.familyName || '未知家庭' }}</text>
            </view>
            <MiniStatusTag :status="item.requestStatus" :label="joinRequestStatusText(item.requestStatus)" />
          </view>
          <view class="detail-grid">
            <text class="item-meta full">申请理由：{{ item.applicantMessage || '未填写' }}</text>
            <text v-if="item.handleComment" class="item-meta full">审核意见：{{ item.handleComment }}</text>
            <text class="item-meta weak">提交时间：{{ formatDate(item.createdAt) }}</text>
          </view>
          <MiniButton
            v-if="item.requestStatus === 'PENDING'"
            variant="danger"
            :loading="cancellingRequestId === item.requestId"
            :disabled="Boolean(cancellingRequestId)"
            @click="confirmCancelRequest(item)"
          >
            取消申请
          </MiniButton>
          <MiniButton
            v-else-if="item.requestStatus === 'APPROVED'"
            variant="secondary"
            @click="openMyFamilies"
          >
            进入我的家庭
          </MiniButton>
        </MiniCard>
      </view>
    </view>

    <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniDirectoryTile from '@/components/base/MiniDirectoryTile.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import {
  invitationStatusText,
  joinRequestStatusText
} from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import MiniStatusTag from '@/components/base/MiniStatusTag.vue'
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
  uni.navigateTo({
    url: '/pages/family/my',
    complete: () => {
      navigating.value = false
    }
  })
}

function openSearch() {
  uni.switchTab({ url: '/pages/family/search' })
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
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.affairs-hero {
  margin-bottom: 22rpx;
}

.hero-kicker,
.hero-title,
.hero-desc {
  display: block;
}

.hero-kicker {
  color: rgba(248, 231, 194, 0.92);
  font-size: 21rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}

.hero-title {
  margin-top: 8rpx;
  color: #fff;
  font-size: 38rpx;
  font-weight: 800;
}

.hero-desc {
  margin-top: 10rpx;
  color: rgba(255, 255, 255, 0.72);
  font-size: 24rpx;
  line-height: 1.6;
}

.summary-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 20rpx;
}

.summary-chip {
  border: 1rpx solid rgba(255, 255, 255, 0.18);
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  padding: 7rpx 14rpx;
  font-size: 21rpx;
}

.affairs-tabs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10rpx;
  margin-bottom: 22rpx;
  border-radius: 26rpx;
  background: rgba(255, 255, 255, 0.82);
  padding: 8rpx;
  box-shadow: 0 10rpx 28rpx rgba(24, 54, 83, 0.06);
}

.affairs-tabs button {
  min-height: 72rpx;
  margin: 0;
  border: 0;
  border-radius: 20rpx;
  background: transparent;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  transition: transform 180ms ease-out, background-color 180ms ease-out;
}

.affairs-tabs button:active {
  background-color: rgba(24, 54, 83, 0.05);
  transform: translateY(1rpx) scale(0.99);
}

.affairs-tabs button::after {
  border: 0;
}

.affairs-tabs button.active {
  background: linear-gradient(135deg, var(--tree-primary) 0%, var(--tree-green) 100%);
  color: #fff;
  font-weight: 800;
}

.affairs-tabs button.active:active {
  background: linear-gradient(135deg, var(--tree-primary) 0%, var(--tree-green) 100%);
  transform: translateY(1rpx) scale(0.99);
}

.tab-count {
  margin-left: 6rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.18);
  padding: 2rpx 8rpx;
  font-size: 18rpx;
}

.affairs-page > :deep(.mini-notice) {
  margin-bottom: 20rpx;
}

.affair-group {
  margin-bottom: 26rpx;
}

.affair-card {
  position: relative;
  overflow: hidden;
}

.affair-card::before {
  content: '';
  position: absolute;
  top: 24rpx;
  bottom: 24rpx;
  left: 0;
  width: 7rpx;
  border-radius: 0 999rpx 999rpx 0;
}

.invitation-card::before {
  background: linear-gradient(180deg, var(--tree-gold) 0%, var(--tree-green) 100%);
}

.request-card::before {
  background: linear-gradient(180deg, var(--tree-green) 0%, var(--tree-primary) 100%);
}

.item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.item-kicker,
.item-name,
.item-meta {
  display: block;
}

.item-kicker {
  color: var(--tree-green);
  font-size: 21rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}

.item-name {
  margin-top: 5rpx;
  color: var(--tree-text-primary);
  font-size: 32rpx;
  font-weight: 800;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10rpx 16rpx;
  border-radius: 20rpx;
  background: rgba(248, 250, 252, 0.84);
  padding: 18rpx;
}

.item-meta {
  color: var(--tree-text-secondary);
  font-size: 23rpx;
  line-height: 1.55;
}

.item-meta.full {
  grid-column: 1 / -1;
}

.item-meta.weak {
  color: var(--tree-text-weak);
}

.action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12rpx;
  margin-top: 18rpx;
}

.affair-card > :deep(.mini-button) {
  margin-top: 18rpx;
}

.directory-card {
  margin-bottom: 16rpx;
}

.directory-card--list {
  padding-top: 8rpx;
  padding-bottom: 8rpx;
}

.tab-panel {
  animation: tab-panel-in 200ms ease-out;
}

@keyframes tab-panel-in {
  from {
    opacity: 0;
    transform: translateY(8rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.action-grid :deep(.mini-button) {
  margin-top: 0;
}
</style>
