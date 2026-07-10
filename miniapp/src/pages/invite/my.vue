<template>
  <view class="archive-page invite-my-page">
    <MiniBackHome />

    <view class="invite-head archive-page-head">
      <view>
        <text class="archive-kicker">Family Invitations</text>
        <text class="archive-title">收到的家庭邀请</text>
        <text class="archive-subtitle">核对家庭名称与成员身份后，再决定是否接受邀请</text>
      </view>
      <view class="archive-seal">邀</view>
    </view>

    <view class="invite-summary">
      <text class="archive-chip">待处理 {{ pendingCount }}</text>
      <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadInvitations">
        刷新列表
      </MiniButton>
    </view>

    <template v-if="authChecked">
      <MiniNotice tone="security" title="核对邀请信息">
        邀请由家庭管理员发起，请核对家庭名称、成员身份和邀请发起人后再接受。
      </MiniNotice>

      <view v-if="loading && invitations.length === 0" class="invite-state archive-form-panel">
        <MiniEmptyState symbol="…" title="正在加载" description="正在同步收到的邀请..." />
      </view>

      <view v-else-if="loadError && invitations.length === 0" class="invite-state archive-form-panel">
        <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
        <MiniButton variant="secondary" size="sm" class="invite-action" @click="loadInvitations">重新加载</MiniButton>
      </view>

      <view v-else-if="invitations.length === 0" class="invite-state archive-form-panel">
        <MiniEmptyState
          symbol="邀"
          title="暂无邀请"
          description="请家人发送家庭邀请或公开家庭分享链接。"
        />
      </view>

      <template v-else>
        <view v-for="group in invitationGroups" :key="group.key" class="invite-group archive-form-panel">
          <view class="archive-section-head">
            <text class="archive-section-title">{{ group.title }}</text>
            <text class="archive-section-subtitle">{{ group.subtitle }}</text>
          </view>
          <view class="invite-list archive-list">
            <view v-for="item in group.items" :key="item.invitationId" class="invite-item">
              <view class="invite-item-head">
                <view class="archive-row-main">
                  <text class="archive-row-title">{{ item.familyName }}</text>
                  <text class="archive-row-desc">成员身份：{{ inviteTargetText(item) }}</text>
                </view>
                <text class="archive-status-tag" :class="invitationStatusClass(displayStatus(item))">
                  {{ invitationStatusText(displayStatus(item)) }}
                </text>
              </view>

              <view class="invite-meta">
                <text>邀请人：{{ item.inviterDisplayName || '家庭管理员' }}（{{ inviterRoleText(item.inviterRole) }}）</text>
                <text class="invite-meta-weak">有效期至：{{ formatDate(item.expiredAt) }}</text>
              </view>

              <view v-if="expandedInvitationId === item.invitationId" class="invite-detail">
                <text>邀请方式：{{ inviteChannelText(item.inviteChannel) }}</text>
                <text>加入后角色：{{ roleText(item.familyRoleAfterAccept) }}</text>
                <text v-if="item.inviteMessage">邀请说明：{{ item.inviteMessage }}</text>
              </view>

              <view class="invite-actions">
                <template v-if="displayStatus(item) === 'PENDING'">
                  <MiniButton
                    size="sm"
                    :disabled="actingId === item.invitationId"
                    :loading="actingId === item.invitationId && actingType === 'accept'"
                    @click="confirmAccept(item)"
                  >
                    接受
                  </MiniButton>
                  <MiniButton
                    size="sm"
                    variant="secondary"
                    :disabled="actingId === item.invitationId"
                    :loading="actingId === item.invitationId && actingType === 'reject'"
                    @click="confirmReject(item)"
                  >
                    拒绝
                  </MiniButton>
                </template>
                <MiniButton
                  v-else-if="item.status === 'ACCEPTED'"
                  size="sm"
                  variant="secondary"
                  @click="openMyFamilies"
                >
                  进入我的家庭
                </MiniButton>
                <MiniButton size="sm" variant="secondary" @click="openInvitationDetail(item)">
                  {{ expandedInvitationId === item.invitationId ? '收起详情' : '查看详情' }}
                </MiniButton>
              </view>
            </view>
          </view>
        </view>
      </template>

      <text v-if="actionError" class="tree-field-error invite-error">{{ actionError }}</text>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onHide, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { acceptInvitation, listMyInvitations, rejectInvitation } from '@/api/invitations'
import { apiErrorMessage } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText, roleText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Invitation } from '@/types/api'

const session = useSessionStore()
const invitations = ref<Invitation[]>([])
const loading = ref(false)
const loadError = ref('')
const actionError = ref('')
const actingId = ref<number | string | null>(null)
const actingType = ref<'' | 'accept' | 'reject'>('')
const authChecked = ref(false)
const expandedInvitationId = ref<number | string | null>(null)

const pendingCount = computed(() =>
  invitations.value.filter((item) => displayStatus(item) === 'PENDING').length
)

const invitationGroups = computed(() => {
  const pending = invitations.value.filter((item) => displayStatus(item) === 'PENDING')
  const history = invitations.value.filter((item) => displayStatus(item) !== 'PENDING')
  return [
    { key: 'pending', title: '待我处理', subtitle: '需要确认家庭与成员身份的邀请', items: pending },
    { key: 'history', title: '历史记录', subtitle: '已接受、拒绝或失效的邀请', items: history }
  ].filter((group) => group.items.length > 0)
})

function invitationStatusClass(status: string) {
  if (status === 'PENDING') return 'is-cinnabar'
  if (status === 'ACCEPTED') return 'is-ink'
  return 'is-muted'
}

function resetTransientUI() {
  actionError.value = ''
  actingId.value = null
  actingType.value = ''
  expandedInvitationId.value = null
}

function resetPageData() {
  authChecked.value = false
  loading.value = false
  loadError.value = ''
  invitations.value = []
  resetTransientUI()
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

function displayStatus(item: Invitation) {
  if (item.status === 'PENDING' && new Date(item.expiredAt).getTime() <= Date.now()) return 'EXPIRED'
  return item.status
}

function isPendingMemberInvitation(item: Invitation) {
  return item.inviteType === 'JOIN_FAMILY_PENDING_MEMBER'
}

function inviteTargetText(item: Invitation) {
  return isPendingMemberInvitation(item)
    ? item.pendingMemberLabel || item.targetMemberName || '身份待确认'
    : item.targetMemberName
}

function inviteChannelText(channel: string) {
  switch (channel) {
    case 'SHARE_LINK':
      return '链接邀请'
    case 'IN_APP':
      return '站内邀请'
    default:
      return '其他方式'
  }
}

function inviterRoleText(role?: string) {
  return role === 'FAMILY_FOUNDER' ? '家庭创建者' : '家庭管理员'
}

function openMyFamilies() {
  uni.switchTab({ url: '/pages/family/my' })
}

function openInvitationDetail(item: Invitation) {
  const inviteToken = (item as Invitation & { inviteToken?: string }).inviteToken
  if (inviteToken) {
    uni.navigateTo({ url: `/pages/invite/detail?inviteToken=${encodeURIComponent(inviteToken)}` })
    return
  }
  expandedInvitationId.value =
    expandedInvitationId.value === item.invitationId ? null : item.invitationId
}

async function loadInvitations() {
  session.restoreSession()
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin('/pages/invite/my')
    return
  }
  authChecked.value = true
  const isInitialLoad = invitations.value.length === 0
  loading.value = isInitialLoad
  loadError.value = ''
  actionError.value = ''
  try {
    invitations.value = await listMyInvitations()
  } catch (error) {
    loadError.value = apiErrorMessage(error, '邀请列表加载失败。')
  } finally {
    loading.value = false
  }
}

function confirmAccept(item: Invitation) {
  uni.showModal({
    title: '接受邀请',
    content: `确定接受加入「${item.familyName}」的邀请吗？`,
    success: (result) => {
      if (result.confirm) accept(item.invitationId)
    }
  })
}

function confirmReject(item: Invitation) {
  uni.showModal({
    title: '拒绝邀请',
    content: `确定拒绝加入「${item.familyName}」的邀请吗？`,
    success: (result) => {
      if (result.confirm) reject(item.invitationId)
    }
  })
}

async function accept(invitationId: number | string) {
  actingId.value = invitationId
  actingType.value = 'accept'
  actionError.value = ''
  try {
    await acceptInvitation(invitationId)
    await loadInvitations()
  } catch (error) {
    actionError.value = apiErrorMessage(error, '接受邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

async function reject(invitationId: number | string) {
  actingId.value = invitationId
  actingType.value = 'reject'
  actionError.value = ''
  try {
    await rejectInvitation(invitationId)
    await loadInvitations()
  } catch (error) {
    actionError.value = apiErrorMessage(error, '拒绝邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

onShow(loadInvitations)
onHide(resetTransientUI)
onUnload(resetPageData)
</script>

<style scoped>
.invite-my-page {
  padding-top: 28rpx;
}

.invite-my-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.invite-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 22rpx;
}

.invite-my-page :deep(.mini-notice) {
  margin-bottom: 20rpx;
}

.invite-group {
  margin-bottom: 24rpx;
}

.invite-list {
  margin-top: 8rpx;
}

.invite-item {
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.invite-item:last-child {
  border-bottom: 0;
}

.invite-item-head {
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

.invite-meta,
.invite-detail {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.invite-detail {
  margin-top: 10rpx;
  padding-top: 10rpx;
  border-top: 1rpx dashed rgba(44, 36, 22, 0.12);
}

.invite-meta-weak {
  opacity: 0.88;
}

.invite-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.invite-action {
  margin-top: 16rpx;
  align-self: flex-start;
}

.invite-error {
  display: block;
  margin-top: 16rpx;
}
</style>
