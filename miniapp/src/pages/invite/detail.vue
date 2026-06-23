<template>
  <view class="tree-page">
    <MiniBackHome />
    <MiniSectionHeader title="邀请确认" subtitle="请核对邀请信息后再决定是否加入家庭。" />

    <MiniCard v-if="loading">
      <view class="state-block">
        <text class="tree-muted">正在加载邀请详情...</text>
      </view>
    </MiniCard>

    <MiniCard v-else-if="loadError && !invitation">
      <MiniEmptyState
        symbol="!"
        title="邀请加载失败"
        :description="loadError"
        action-text="重新加载"
        @action="loadInvitation"
      />
    </MiniCard>

    <template v-else-if="invitation">
      <MiniNotice tone="security" title="安全提示">
        邀请由家庭管理员发起，请确认家庭名称与邀请成员后再操作。我们不会向你索要密码或验证码。
      </MiniNotice>

      <MiniCard variant="soft">
        <view class="trust-card-head">
          <text class="trust-family-name">{{ invitation.familyName }}</text>
          <MiniStatusTag
            :status="invitation.status"
            :label="invitationStatusText(invitation.status)"
          />
        </view>
        <text class="tree-muted trust-hint">邀请你以「{{ invitation.targetMemberName }}」身份加入此家庭</text>

        <view class="tree-info-row">
          <text class="tree-info-label">邀请方式</text>
          <text class="tree-info-value">{{ inviteChannelText(invitation.inviteChannel) }}</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">加入后角色</text>
          <text class="tree-info-value">{{ roleText(invitation.familyRoleAfterAccept) }}</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">有效期至</text>
          <text class="tree-info-value">{{ formatDate(invitation.expiredAt) }}</text>
        </view>
      </MiniCard>

      <MiniCard v-if="invitation.inviteMessage">
        <text class="detail-block-title">邀请说明</text>
        <text class="tree-muted">{{ invitation.inviteMessage }}</text>
      </MiniCard>

      <MiniCard v-if="invitation.status === 'PENDING'">
        <template v-if="session.isLoggedIn && session.isPhoneBound">
          <text class="detail-block-title">处理邀请</text>
          <textarea
            v-model.trim="rejectReason"
            class="tree-textarea"
            maxlength="300"
            placeholder="拒绝原因（可选）"
          />
          <MiniButton
            :disabled="Boolean(acting)"
            :loading="acting === 'accept'"
            @click="confirmAccept"
          >
            接受邀请
          </MiniButton>
          <MiniButton
            variant="secondary"
            :disabled="Boolean(acting)"
            :loading="acting === 'reject'"
            @click="confirmReject"
          >
            拒绝邀请
          </MiniButton>
        </template>
        <template v-else-if="session.isLoggedIn && !session.isPhoneBound">
          <MiniNotice tone="warm" title="需要绑定手机号">
            请先绑定手机号并设置登录密码，再处理邀请。
          </MiniNotice>
          <MiniButton @click="requirePhoneBound">去绑定手机号</MiniButton>
        </template>
        <template v-else>
          <MiniNotice tone="warm" title="需要登录">
            请先登录，再处理邀请。
          </MiniNotice>
          <MiniButton @click="requireLogin">微信登录</MiniButton>
          <MiniButton variant="secondary" class="btn-top" @click="goPhoneLogin">手机号登录</MiniButton>
        </template>
      </MiniCard>

      <MiniCard v-else>
        <MiniNotice tone="info">
          该邀请当前状态为「{{ invitationStatusText(invitation.status) }}」，不能继续处理。
        </MiniNotice>
      </MiniCard>

      <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
      <text v-if="result" class="tree-field-success">{{ result }}</text>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { acceptInvitation, getInvitationDetail, rejectInvitation } from '@/api/invitations'
import { apiErrorMessage, pendingRouteKey } from '@/api/client'
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

function requireLogin() {
  session.requireLogin(currentRoute())
}

function requirePhoneBound() {
  session.requirePhoneBound(currentRoute())
}

function goPhoneLogin() {
  uni.setStorageSync(pendingRouteKey, currentRoute())
  uni.navigateTo({ url: '/pages/auth/phone-login' })
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
  if (!session.requirePhoneBound(currentRoute())) return
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
  if (!session.requirePhoneBound(currentRoute())) return
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
.state-block {
  padding: 32rpx 0;
  text-align: center;
}

.trust-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.trust-family-name {
  flex: 1;
  color: var(--tree-text-primary);
  font-size: 34rpx;
  font-weight: 600;
  line-height: 1.4;
}

.trust-hint {
  display: block;
  margin-bottom: 8rpx;
  line-height: 1.7;
}

.detail-block-title {
  display: block;
  margin-bottom: 12rpx;
  color: var(--tree-text);
  font-size: 28rpx;
  font-weight: 600;
}
</style>
