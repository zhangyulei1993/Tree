<template>
  <view class="page">
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载邀请详情...</text>
    </view>
    <view v-else-if="loadError && !invitation" class="card state-card">
      <text class="title">邀请确认</text>
      <text class="error">{{ loadError }}</text>
      <button class="button secondary" @click="loadInvitation">重新加载</button>
    </view>
    <view v-else-if="invitation" class="card">
      <text class="title">邀请确认</text>
      <view class="info-row">
        <text class="label">家庭</text>
        <text>{{ invitation.familyName }}</text>
      </view>
      <view class="info-row">
        <text class="label">邀请成员</text>
        <text>{{ invitation.targetMemberName }}</text>
      </view>
      <view class="info-row">
        <text class="label">邀请方式</text>
        <text>{{ inviteChannelText(invitation.inviteChannel) }}</text>
      </view>
      <view class="info-row">
        <text class="label">加入后角色</text>
        <text>{{ roleText(invitation.familyRoleAfterAccept) }}</text>
      </view>
      <view class="info-row">
        <text class="label">邀请状态</text>
        <text class="tag">{{ inviteStatusText(invitation.status) }}</text>
      </view>
      <view class="info-row">
        <text class="label">有效期至</text>
        <text>{{ formatDate(invitation.expiredAt) }}</text>
      </view>
      <text v-if="invitation.inviteMessage" class="notice">邀请说明：{{ invitation.inviteMessage }}</text>

      <view v-if="invitation.status === 'PENDING'" class="actions">
        <template v-if="session.isLoggedIn && session.isPhoneBound">
          <textarea
            v-model.trim="rejectReason"
            class="textarea"
            maxlength="300"
            placeholder="拒绝原因（可选）"
          />
          <button class="button" :disabled="Boolean(acting)" :loading="acting === 'accept'" @click="confirmAccept">
            接受邀请
          </button>
          <button
            class="button secondary"
            :disabled="Boolean(acting)"
            :loading="acting === 'reject'"
            @click="confirmReject"
          >
            拒绝邀请
          </button>
        </template>
        <template v-else>
          <text class="notice">请先使用手机号登录或注册，再处理邀请。</text>
          <button class="button" @click="requireLogin">手机号登录</button>
        </template>
      </view>
      <view v-else>
        <text class="notice">该邀请当前状态为「{{ inviteStatusText(invitation.status) }}」，不能继续处理。</text>
      </view>
      <text v-if="actionError" class="error">{{ actionError }}</text>
      <text v-if="result" class="success">{{ result }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { acceptInvitation, getInvitationDetail, rejectInvitation } from '@/api/invitations'
import { apiErrorMessage } from '@/api/client'
import { useSessionStore } from '@/stores/session'
import type { Invitation } from '@/types/api'

const session = useSessionStore()
const inviteToken = ref('')
const invitation = ref<Invitation | null>(null)
const rejectReason = ref('')
const loading = ref(false)
const acting = ref<'' | 'accept' | 'reject'>('')
const loadError = ref('')
const actionError = ref('')
const result = ref('')

function currentRoute() {
  return `/pages/invite/detail?inviteToken=${encodeURIComponent(inviteToken.value)}`
}

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

function requireLogin() {
  session.requireLogin(currentRoute())
}

async function loadInvitation() {
  if (!inviteToken.value) {
    loadError.value = '邀请信息缺失。'
    return
  }
  loading.value = true
  loadError.value = ''
  actionError.value = ''
  result.value = ''
  try {
    invitation.value = await getInvitationDetail(inviteToken.value)
  } catch (error) {
    invitation.value = null
    loadError.value = apiErrorMessage(error, '邀请详情加载失败。')
  } finally {
    loading.value = false
  }
}

function confirmAccept() {
  if (!invitation.value) return
  uni.showModal({
    title: '接受邀请',
    content: `确定接受加入「${invitation.value.familyName}」的邀请吗？`,
    success: (modalResult) => {
      if (modalResult.confirm) accept()
    }
  })
}

function confirmReject() {
  if (!invitation.value) return
  uni.showModal({
    title: '拒绝邀请',
    content: `确定拒绝加入「${invitation.value.familyName}」的邀请吗？`,
    success: (modalResult) => {
      if (modalResult.confirm) reject()
    }
  })
}

async function accept() {
  if (!invitation.value) return
  if (!session.requireLogin(currentRoute())) return
  acting.value = 'accept'
  actionError.value = ''
  result.value = ''
  try {
    invitation.value = await acceptInvitation(invitation.value.invitationId)
    result.value = '邀请已接受。'
  } catch (error) {
    actionError.value = apiErrorMessage(error, '接受邀请失败。')
  } finally {
    acting.value = ''
  }
}

async function reject() {
  if (!invitation.value) return
  if (!session.requireLogin(currentRoute())) return
  acting.value = 'reject'
  actionError.value = ''
  result.value = ''
  try {
    invitation.value = await rejectInvitation(invitation.value.invitationId, {
      reason: rejectReason.value || undefined
    })
    result.value = '邀请已拒绝。'
  } catch (error) {
    actionError.value = apiErrorMessage(error, '拒绝邀请失败。')
  } finally {
    acting.value = ''
  }
}

onLoad((options) => {
  inviteToken.value = String(options?.inviteToken || '')
  session.restoreSession()
  loadInvitation()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  padding: 14rpx 0;
  border-top: 1rpx solid #e5e0d6;
  font-size: 26rpx;
}

.label {
  flex-shrink: 0;
  color: #6b7280;
}

.info-row text:last-child {
  text-align: right;
}

.actions {
  margin-top: 24rpx;
}

.notice,
.error,
.success {
  display: block;
  margin-top: 18rpx;
  font-size: 24rpx;
}

.notice {
  color: #6b7280;
}

.error {
  color: #c0392b;
}

.success {
  color: #2f6b57;
  font-weight: 600;
}

.button[disabled] {
  opacity: 0.55;
}
</style>
