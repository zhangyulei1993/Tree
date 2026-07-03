<template>
  <view class="tree-page invite-detail-page">
    <MiniBackHome />

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
      <MiniCard variant="hero" class="invitation-hero">
        <view class="trust-card-head">
          <view>
            <text class="invitation-kicker">来自家庭</text>
            <text class="trust-family-name">{{ invitation.familyName }}</text>
          </view>
          <MiniStatusTag
            :status="invitation.status"
            :label="invitationStatusText(invitation.status)"
          />
        </view>
        <view class="identity-confirmation">
          <text class="identity-label">邀请你确认的成员身份</text>
          <text class="identity-name">{{ invitation.targetMemberName }}</text>
          <text class="tree-muted identity-hint">接受后，你的账号将与该家谱成员节点绑定。</text>
        </view>

        <view class="tree-info-row">
          <text class="tree-info-label">邀请发起人</text>
          <text class="tree-info-value">
            {{ invitation.inviterDisplayName || '家庭管理员' }}（{{ inviterRoleText(invitation.inviterRole) }}）
          </text>
        </view>
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

      <MiniCard>
        <text class="detail-block-title">接受后你可以</text>
        <view class="benefit-list">
          <view class="benefit-item"><text class="benefit-dot" /><text>在「{{ invitation.familyName }}」中确认自己的成员身份</text></view>
          <view class="benefit-item"><text class="benefit-dot" /><text>查看家庭成员和家谱关系</text></view>
          <view class="benefit-item"><text class="benefit-dot" /><text>在个人中心查看已加入的家庭</text></view>
        </view>
      </MiniCard>

      <MiniCard>
        <text class="detail-block-title">邀请人留言</text>
        <text class="tree-muted invitation-message">
          {{ invitation.inviteMessage || '邀请人未填写额外说明。' }}
        </text>
      </MiniCard>

      <MiniNotice tone="security" title="接受前请确认">
        请确认上方姓名就是你在家谱中的身份。如果信息不符，请先拒绝邀请并联系家庭管理员。平台不会向你索要密码或验证码。
      </MiniNotice>

      <MiniCard v-if="invitation.status === 'PENDING'">
        <template v-if="session.isLoggedIn && session.isProfileComplete">
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
        <template v-else-if="session.isLoggedIn && !session.isProfileComplete">
          <MiniNotice tone="warm" title="需要完善资料">
            请先设置昵称，再处理邀请。
          </MiniNotice>
          <MiniButton @click="requireProfileComplete">去完善资料</MiniButton>
        </template>
        <template v-else>
          <MiniNotice tone="warm" title="需要登录">
            请先登录，再处理邀请。
          </MiniNotice>
          <MiniButton @click="requireLogin">微信登录</MiniButton>
        </template>
      </MiniCard>

      <MiniCard v-else>
        <MiniNotice tone="info">
          该邀请当前状态为「{{ invitationStatusText(invitation.status) }}」，不能继续处理。
        </MiniNotice>
        <MiniButton
          v-if="invitation.status === 'ACCEPTED'"
          class="btn-top"
          @click="openMyFamilies"
        >
          进入我的家庭
        </MiniButton>
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
import { apiErrorMessage } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText, roleText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
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

function inviterRoleText(role?: string) {
  return role === 'FAMILY_FOUNDER' ? '家庭创建者' : '家庭管理员'
}

function requireLogin() {
  session.requireLogin(currentRoute())
}

function requireProfileComplete() {
  session.requireProfileComplete(currentRoute())
}

function openMyFamilies() {
  uni.navigateTo({ url: '/pages/family/my' })
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
  if (!session.requireProfileComplete(currentRoute())) return
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
  if (!session.requireProfileComplete(currentRoute())) return
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
.invite-detail-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

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
  color: #fff;
  font-size: 38rpx;
  font-weight: 800;
  line-height: 1.4;
}

.invitation-hero {
  margin-bottom: 26rpx;
}

.invitation-kicker {
  display: block;
  margin-bottom: 6rpx;
  color: rgba(248, 231, 194, 0.92);
  font-size: 22rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}

.identity-confirmation {
  margin: 18rpx 0 8rpx;
  padding: 22rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.16);
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.12);
}

.identity-label,
.identity-name,
.identity-hint {
  display: block;
}

.identity-label {
  color: rgba(255, 255, 255, 0.72);
  font-size: 22rpx;
}

.identity-name {
  margin: 8rpx 0;
  color: #fff;
  font-size: 42rpx;
  font-weight: 800;
}

.identity-hint {
  color: rgba(255, 255, 255, 0.70);
}

.benefit-list {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
}

.benefit-item {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  border: 1rpx solid rgba(226, 232, 240, 0.82);
  border-radius: 20rpx;
  background: rgba(248, 250, 252, 0.86);
  color: var(--tree-text-secondary);
  padding: 16rpx 18rpx;
  font-size: 26rpx;
  line-height: 1.6;
}

.benefit-dot {
  flex: 0 0 auto;
  width: 10rpx;
  height: 10rpx;
  margin-top: 15rpx;
  border-radius: 50%;
  background: var(--tree-green, #2f6b57);
}

.invitation-message {
  display: block;
  line-height: 1.7;
}

.trust-hint {
  display: block;
  margin-bottom: 8rpx;
  line-height: 1.7;
}

.detail-block-title {
  display: block;
  margin-bottom: 16rpx;
  color: var(--tree-text);
  font-size: 31rpx;
  font-weight: 800;
}
</style>
