<template>
  <view class="page">
    <view class="card">
      <text class="title">我的邀请</text>
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
      <text class="muted">暂无收到的站内邀请。</text>
    </view>
    <view v-else>
      <view v-for="item in invitations" :key="item.invitationId" class="card item-card">
        <view class="item-header">
          <text class="section-title">{{ item.familyName }}</text>
          <text class="tag">{{ item.status }}</text>
        </view>
        <text class="muted">目标成员：{{ item.targetMemberName }}</text>
        <text class="muted">邀请渠道：{{ item.inviteChannel }}</text>
        <text v-if="item.inviteMessage" class="muted">邀请说明：{{ item.inviteMessage }}</text>
        <text class="muted">有效期至：{{ formatDate(item.expiredAt) }}</text>
        <view v-if="item.status === 'PENDING'" class="actions">
          <button
            class="button"
            :disabled="actingId === item.invitationId"
            :loading="actingId === item.invitationId && actingType === 'accept'"
            @click="accept(item.invitationId)"
          >
            接受
          </button>
          <button
            class="button secondary"
            :disabled="actingId === item.invitationId"
            :loading="actingId === item.invitationId && actingType === 'reject'"
            @click="reject(item.invitationId)"
          >
            拒绝
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
