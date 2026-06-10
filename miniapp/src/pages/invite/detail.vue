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
      <text class="muted">家庭：{{ invitation.familyName }}</text>
      <text class="muted">目标成员：{{ invitation.targetMemberName }}</text>
      <text class="muted">邀请渠道：{{ invitation.inviteChannel }}</text>
      <text class="muted">接受后角色：{{ invitation.familyRoleAfterAccept }}</text>
      <text class="muted">状态：{{ invitation.status }}</text>
      <text class="muted">有效期至：{{ formatDate(invitation.expiredAt) }}</text>
      <text v-if="invitation.inviteMessage" class="notice">{{ invitation.inviteMessage }}</text>

      <view v-if="invitation.status === 'PENDING'" class="actions">
        <template v-if="session.isLoggedIn && session.isPhoneBound">
          <textarea
            v-model.trim="rejectReason"
            class="textarea"
            maxlength="300"
            placeholder="拒绝原因（可选）"
          />
          <button class="button" :disabled="Boolean(acting)" :loading="acting === 'accept'" @click="accept">
            接受绑定
          </button>
          <button
            class="button secondary"
            :disabled="Boolean(acting)"
            :loading="acting === 'reject'"
            @click="reject"
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
        <text class="notice">该邀请当前状态为 {{ invitation.status }}，不能继续处理。</text>
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

function requireLogin() {
  session.requireLogin(currentRoute())
}

async function loadInvitation() {
  if (!inviteToken.value) {
    loadError.value = '缺少邀请参数。'
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
