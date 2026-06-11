<template>
  <view class="page">
    <view class="card">
      <text class="title">我的邀请</text>
      <text class="muted">查看并处理收到的家庭邀请。</text>
      <button class="button secondary" :disabled="loading" :loading="loading" @click="loadInvitations">
        刷新
      </button>
      <text v-if="actionError" class="error">{{ actionError }}</text>
    </view>

    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载邀请...</text>
    </view>
    <view v-else-if="loadError" class="card state-card">
      <text class="error">{{ loadError }}</text>
      <button class="button secondary" @click="loadInvitations">重新加载</button>
    </view>
    <view v-else-if="invitations.length === 0" class="card state-card">
      <text class="section-title">暂无邀请</text>
      <text class="muted">收到家庭邀请后，会在这里显示。</text>
    </view>
    <view v-else>
      <view v-for="item in invitations" :key="item.invitationId" class="card item-card">
        <view class="item-header">
          <text class="section-title">{{ item.familyName }}</text>
          <text class="tag">{{ inviteStatusText(item.status) }}</text>
        </view>
        <text class="muted">邀请成员：{{ item.targetMemberName }}</text>
        <text class="muted">邀请方式：{{ inviteChannelText(item.inviteChannel) }}</text>
        <text class="muted">加入后角色：{{ roleText(item.familyRoleAfterAccept) }}</text>
        <text v-if="item.inviteMessage" class="muted">邀请说明：{{ item.inviteMessage }}</text>
        <text class="muted">有效期至：{{ formatDate(item.expiredAt) }}</text>
        <view v-if="item.status === 'PENDING'" class="actions">
          <button
            class="button"
            :disabled="actingId === item.invitationId"
            :loading="actingId === item.invitationId && actingType === 'accept'"
            @click="confirmAccept(item)"
          >
            接受邀请
          </button>
          <button
            class="button secondary"
            :disabled="actingId === item.invitationId"
            :loading="actingId === item.invitationId && actingType === 'reject'"
            @click="confirmReject(item)"
          >
            拒绝邀请
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { acceptInvitation, listMyInvitations, rejectInvitation } from '@/api/invitations'
import { apiErrorMessage } from '@/api/client'
import { useSessionStore } from '@/stores/session'
import type { Invitation } from '@/types/api'

const session = useSessionStore()
const invitations = ref<Invitation[]>([])
const loading = ref(false)
const loadError = ref('')
const actionError = ref('')
const actingId = ref<number | string | null>(null)
const actingType = ref<'' | 'accept' | 'reject'>('')

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

function inviteStatusText(status: string) {
  switch (status) {
    case 'PENDING':
      return '待处理'
    case 'ACCEPTED':
      return '已接受'
    case 'REJECTED':
      return '已拒绝'
    case 'EXPIRED':
      return '已过期'
    case 'CANCELLED':
      return '已取消'
    default:
      return '未知状态'
  }
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

function roleText(role: string) {
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    case 'ROOT_ADMIN':
      return '超级管理员'
    default:
      return '未知角色'
  }
}

async function loadInvitations() {
  if (!session.requireLogin('/pages/invite/my')) return
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

onLoad(() => {
  session.restoreSession()
  loadInvitations()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.item-card {
  margin-top: 20rpx;
}

.item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.actions {
  margin-top: 20rpx;
}

.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.button[disabled] {
  opacity: 0.55;
}
</style>
