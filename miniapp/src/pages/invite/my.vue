<template>
  <view class="tree-page">
    <MiniBackHome />
    <template v-if="authChecked">
      <MiniSectionHeader title="我的邀请" subtitle="查看并处理收到的家庭邀请。" />

      <MiniCard>
        <MiniNotice tone="security">
          邀请由家庭管理员发起，请核对家庭名称与成员身份后再接受。
        </MiniNotice>
        <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadInvitations">
          刷新列表
        </MiniButton>
        <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
      </MiniCard>

      <MiniCard v-if="loading">
        <view class="state-block">
          <text class="tree-muted">正在加载邀请...</text>
        </view>
      </MiniCard>

      <MiniCard v-else-if="loadError">
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
          action-text="寻找家族"
          @action="openSearch"
        />
      </MiniCard>

      <template v-else>
        <MiniCard v-for="item in invitations" :key="item.invitationId" variant="soft">
          <view class="item-head">
            <text class="item-name">{{ item.familyName }}</text>
            <MiniStatusTag :status="item.status" :label="invitationStatusText(item.status)" />
          </view>
          <text class="tree-muted item-meta">邀请成员：{{ item.targetMemberName }}</text>
          <text class="tree-muted item-meta">邀请方式：{{ inviteChannelText(item.inviteChannel) }}</text>
          <text class="tree-muted item-meta">加入后角色：{{ roleText(item.familyRoleAfterAccept) }}</text>
          <text v-if="item.inviteMessage" class="tree-muted item-meta">邀请说明：{{ item.inviteMessage }}</text>
          <text class="tree-weak item-meta">有效期至：{{ formatDate(item.expiredAt) }}</text>
          <view v-if="item.status === 'PENDING'" class="item-actions">
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
        </MiniCard>
      </template>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onHide, onShow, onUnload } from '@dcloudio/uni-app'
import { ref } from 'vue'

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

function resetAuthView() {
  authChecked.value = false
  loading.value = false
  loadError.value = ''
  actionError.value = ''
  invitations.value = []
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
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

function openSearch() {
  uni.switchTab({ url: '/pages/family/search' })
}

async function loadInvitations() {
  authChecked.value = false
  invitations.value = []
  if (!session.requireLogin('/pages/invite/my')) return
  authChecked.value = true
  loading.value = true
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

onShow(() => {
  session.restoreSession()
  loadInvitations()
})
onHide(resetAuthView)
onUnload(resetAuthView)
</script>

<style scoped>
.state-block {
  padding: 32rpx 0;
  text-align: center;
}

.item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.item-name {
  flex: 1;
  color: #1f2937;
  font-size: 30rpx;
  font-weight: 700;
}

.item-meta {
  display: block;
  margin-top: 6rpx;
}

.item-actions {
  margin-top: 20rpx;
}
</style>
