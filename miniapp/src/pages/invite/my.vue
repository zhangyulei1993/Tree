<template>
  <view class="tree-page invite-page">
    <MiniBackHome />
    <template v-if="authChecked">
      <MiniCard variant="hero" class="invite-hero">
        <MiniNotice tone="security">
          邀请由家庭管理员发起，请核对家庭名称与成员身份后再接受。
        </MiniNotice>
        <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadInvitations">
          刷新列表
        </MiniButton>
        <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
      </MiniCard>

      <MiniCard v-if="loading && invitations.length === 0">
        <view class="state-block">
          <text class="tree-muted">正在加载邀请...</text>
        </view>
      </MiniCard>

      <MiniCard v-else-if="loadError && invitations.length === 0">
        <MiniEmptyState
          symbol="!"
          title="加载失败"
          :description="loadError"
          action-text="重新加载"
          @action="loadInvitations"
        />
      </MiniCard>

      <MiniCard v-else-if="invitations.length === 0">
        <MiniEmptyState
          symbol="邀"
          title="暂无邀请"
          description="收到家庭邀请后，会在这里显示。你可以请家人发送邀请，或在公开家庭主页提交加入申请。"
          action-text="寻找家庭"
          @action="openSearch"
        />
      </MiniCard>

      <template v-else>
        <view v-for="group in invitationGroups" :key="group.key" class="invitation-group">
          <MiniSectionHeader :title="group.title" :subtitle="group.subtitle" />
          <MiniCard v-for="item in group.items" :key="item.invitationId" variant="soft" class="invitation-card">
            <view class="item-head">
              <view>
                <text class="item-kicker">家庭邀请</text>
                <text class="item-name">{{ item.familyName }}</text>
              </view>
              <MiniStatusTag :status="displayStatus(item)" :label="invitationStatusText(displayStatus(item))" />
            </view>
            <view class="invite-detail-grid">
              <text class="item-meta">邀请成员：{{ item.targetMemberName }}</text>
              <text class="item-meta">
                发起人：{{ item.inviterDisplayName || '家庭管理员' }}（{{ inviterRoleText(item.inviterRole) }}）
              </text>
              <text class="item-meta">邀请方式：{{ inviteChannelText(item.inviteChannel) }}</text>
              <text class="item-meta">加入后角色：{{ roleText(item.familyRoleAfterAccept) }}</text>
              <text v-if="item.inviteMessage" class="item-meta full">邀请说明：{{ item.inviteMessage }}</text>
              <text class="item-meta full weak">有效期至：{{ formatDate(item.expiredAt) }}</text>
            </view>
            <view v-if="displayStatus(item) === 'PENDING'" class="item-actions">
              <MiniButton
                :disabled="actingId === item.invitationId"
                :loading="actingId === item.invitationId && actingType === 'accept'"
                @click="confirmAccept(item)"
              >
                接受邀请
              </MiniButton>
              <MiniButton
                variant="secondary"
                :disabled="actingId === item.invitationId"
                :loading="actingId === item.invitationId && actingType === 'reject'"
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
      </template>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText, roleText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import MiniStatusTag from '@/components/base/MiniStatusTag.vue'
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
const invitationGroups = computed(() => {
  const pending = invitations.value.filter((item) => displayStatus(item) === 'PENDING')
  const history = invitations.value.filter((item) => displayStatus(item) !== 'PENDING')
  return [
    { key: 'pending', title: '待我处理', subtitle: '需要确认成员身份的邀请', items: pending },
    { key: 'history', title: '历史记录', subtitle: '已接受、拒绝或失效的邀请', items: history }
  ].filter((group) => group.items.length > 0)
})

function resetTransientUI() {
  actionError.value = ''
  actingId.value = null
  actingType.value = ''
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

function openSearch() {
  uni.switchTab({ url: '/pages/family/search' })
}

function openMyFamilies() {
  uni.navigateTo({ url: '/pages/family/my' })
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
.invite-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.invite-hero {
  margin-bottom: 26rpx;
}

.invite-hero :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.state-block {
  padding: 32rpx 0;
  text-align: center;
}

.invitation-group {
  margin-bottom: 26rpx;
}

.item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.item-kicker {
  display: block;
  margin-bottom: 6rpx;
  color: var(--tree-green);
  font-size: 21rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}

.item-name {
  flex: 1;
  color: var(--tree-text-primary);
  font-size: 34rpx;
  font-weight: 800;
}

.item-meta {
  display: block;
  color: var(--tree-text-secondary);
  font-size: 23rpx;
  line-height: 1.55;
}

.item-meta.weak {
  color: var(--tree-text-weak);
}

.invite-detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10rpx 16rpx;
  border-radius: 22rpx;
  background: rgba(248, 250, 252, 0.82);
  padding: 18rpx;
}

.item-meta.full {
  grid-column: 1 / -1;
}

.item-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14rpx;
  margin-top: 20rpx;
}

.invitation-card {
  position: relative;
  border-color: rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 100% 0%, rgba(216, 175, 104, 0.12), transparent 160rpx),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(250, 246, 238, 0.94) 100%);
  box-shadow: 0 16rpx 40rpx rgba(24, 54, 83, 0.07);
}

.invitation-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 26rpx;
  bottom: 26rpx;
  width: 7rpx;
  border-radius: 0 999rpx 999rpx 0;
  background: linear-gradient(180deg, var(--tree-gold) 0%, var(--tree-green) 100%);
}
</style>
