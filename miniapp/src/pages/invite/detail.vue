<template>
  <view class="archive-page invite-detail-page">
    <MiniBackHome />

    <view v-if="loading" class="detail-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在读取邀请详情..." />
    </view>

    <view v-else-if="loadError && !invitation" class="detail-state archive-form-panel">
      <MiniEmptyState
        symbol="!"
        title="邀请加载失败"
        :description="loadError"
        action-text="重新加载"
        @action="loadInvitation"
      />
    </view>

    <template v-else-if="invitation">
      <view class="detail-head archive-page-head">
        <view>
          <text class="archive-kicker">Invitation Letter</text>
          <text class="archive-title">邀请确认</text>
          <text class="archive-subtitle">{{ invitationSubtitle }}</text>
        </view>
        <view class="archive-seal detail-seal">谱</view>
      </view>

      <view class="invite-letter" :class="{ pending: invitation.status === 'PENDING' }">
        <view class="invite-letter-head">
          <view class="invite-letter-copy">
            <text class="invite-letter-kicker">家庭邀请帖</text>
            <text class="invite-family-name">{{ invitation.familyName }}</text>
            <view class="invite-target-line">
              <text class="invite-target-label">{{ isPendingMemberInvitation ? '成员身份' : '邀请确认成员' }}</text>
              <text class="invite-target-tag archive-paper-tag">{{ inviteTargetText }}</text>
            </view>
            <text class="invite-target-hint">{{ inviteTargetHint }}</text>
          </view>
          <view class="invite-letter-spine archive-book-spine">
            <text>入</text>
            <text>谱</text>
            <text>帖</text>
          </view>
        </view>

        <view v-if="invitation.status === 'PENDING'" class="invite-pending-mark">
          <text class="invite-pending-seal">待确认</text>
          <text class="invite-pending-note">请在有效期内完成确认</text>
        </view>

        <view class="invite-fields archive-list">
          <view class="archive-row invite-field-row">
            <text class="invite-field-label">邀请发起人</text>
            <text class="invite-field-value">
              {{ invitation.inviterDisplayName || '家庭管理员' }}（{{ inviterRoleText(invitation.inviterRole) }}）
            </text>
          </view>
          <view class="archive-row invite-field-row">
            <text class="invite-field-label">邀请方式</text>
            <text class="invite-field-value">{{ inviteChannelText(invitation.inviteChannel) }}</text>
          </view>
          <view class="archive-row invite-field-row">
            <text class="invite-field-label">加入后角色</text>
            <text class="invite-field-value">{{ roleText(invitation.familyRoleAfterAccept) }}</text>
          </view>
          <view class="archive-row invite-field-row">
            <text class="invite-field-label">有效期至</text>
            <text class="invite-field-value">{{ formatDate(invitation.expiredAt) }}</text>
          </view>
        </view>
      </view>

      <view class="archive-form-panel invite-message-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">邀请人留言</text>
          <text class="archive-section-subtitle">发起人附带的说明文字</text>
        </view>
        <text class="invite-message">{{ invitation.inviteMessage || '邀请人未填写额外说明。' }}</text>
      </view>

      <view class="archive-form-panel invite-benefits-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">接受后你可以</text>
        </view>
        <view class="invite-benefits archive-list">
          <view class="archive-row invite-benefit-row">
            <text class="invite-benefit-text">{{ isPendingMemberInvitation ? '先加入家庭，成员身份之后由管理员确认' : `在「${invitation.familyName}」中确认自己的成员身份` }}</text>
          </view>
          <view class="archive-row invite-benefit-row">
            <text class="invite-benefit-text">查看家庭成员和家庭树关系</text>
          </view>
          <view class="archive-row invite-benefit-row">
            <text class="invite-benefit-text">在个人中心查看已加入的家庭</text>
          </view>
        </view>
      </view>

      <MiniNotice tone="security" title="接受前请确认">
        {{ acceptNoticeText }}
      </MiniNotice>

      <view v-if="invitation.status === 'PENDING'" class="archive-form-panel invite-action-panel">
        <template v-if="session.isLoggedIn && session.isProfileComplete">
          <view class="archive-section-head">
            <text class="archive-section-title">处理邀请</text>
            <text class="archive-section-subtitle">确认无误后可接受；如有疑问可先拒绝</text>
          </view>
          <textarea
            v-model.trim="rejectReason"
            class="tree-textarea invite-reject-input"
            maxlength="300"
            placeholder="拒绝原因（可选）"
          />
          <view class="invite-actions">
            <MiniButton
              size="sm"
              :disabled="Boolean(acting)"
              :loading="acting === 'accept'"
              @click="confirmAccept"
            >
              接受邀请
            </MiniButton>
            <MiniButton
              size="sm"
              variant="secondary"
              :disabled="Boolean(acting)"
              :loading="acting === 'reject'"
              @click="confirmReject"
            >
              拒绝邀请
            </MiniButton>
          </view>
        </template>
        <template v-else-if="session.isLoggedIn && !session.isProfileComplete">
          <MiniNotice tone="warm" title="需要完善资料">
            请先设置昵称，再处理邀请。
          </MiniNotice>
          <MiniButton size="sm" class="invite-action-single" @click="requireProfileComplete">去完善资料</MiniButton>
        </template>
        <template v-else>
          <MiniNotice tone="warm" title="需要登录">
            请先登录，再处理邀请。
          </MiniNotice>
          <MiniButton size="sm" class="invite-action-single" @click="requireLogin">去登录</MiniButton>
        </template>
      </view>

      <view v-else class="archive-form-panel invite-resolved-panel">
        <view class="invite-resolved-line">
          <text class="archive-status-tag" :class="resolvedStatusClass(invitation.status)">
            {{ invitationStatusText(invitation.status) }}
          </text>
          <text class="invite-resolved-text">
            该邀请当前状态为「{{ invitationStatusText(invitation.status) }}」，不能继续处理。
          </text>
        </view>
        <MiniButton
          v-if="invitation.status === 'ACCEPTED'"
          size="sm"
          variant="secondary"
          class="invite-action-single"
          @click="openMyFamilies"
        >
          进入我的家庭
        </MiniButton>
      </view>

      <text v-if="actionError" class="tree-field-error invite-error">{{ actionError }}</text>
      <text v-if="result" class="tree-field-success invite-result">{{ result }}</text>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { acceptInvitation, getInvitationDetail, rejectInvitation } from '@/api/invitations'
import { apiErrorMessage } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText, roleText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Invitation } from '@/types/api'
import { optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const inviteToken = ref('')
const invitation = ref<Invitation | null>(null)
const rejectReason = ref('')
const loading = ref(false)
const acting = ref<'' | 'accept' | 'reject'>('')
const loadError = ref('')
const actionError = ref('')
const result = ref('')
const isPendingMemberInvitation = computed(() => invitation.value?.inviteType === 'JOIN_FAMILY_PENDING_MEMBER')
const invitationSubtitle = computed(() =>
  isPendingMemberInvitation.value ? '确认是否加入该家庭，成员身份可稍后再确认' : '请核对家庭与成员身份，再决定是否家庭树绑定'
)
const inviteTargetText = computed(() =>
  isPendingMemberInvitation.value
    ? invitation.value?.pendingMemberLabel || invitation.value?.targetMemberName || '身份待确认'
    : invitation.value?.targetMemberName || ''
)
const inviteTargetHint = computed(() =>
  isPendingMemberInvitation.value
    ? '接受后，你会先加入该家庭；管理员稍后可确认你的成员节点。'
    : '接受后，你的账号将与该家庭树成员节点绑定。'
)
const acceptNoticeText = computed(() =>
  isPendingMemberInvitation.value
    ? '请确认这是你要加入的家庭。成员身份暂不确定时，可以先接受邀请，之后再由家庭管理员确认。平台不会向你索要密码或验证码。'
    : '请确认上方姓名就是你在家庭树中的身份。如果信息不符，请先拒绝邀请并联系家庭管理员。平台不会向你索要密码或验证码。'
)

function currentRoute() {
  return `/pages/invite/detail?inviteToken=${encodeURIComponent(inviteToken.value)}`
}

function resolvedStatusClass(status: string) {
  if (status === 'ACCEPTED') return 'is-ink'
  return 'is-muted'
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
  uni.switchTab({ url: '/pages/family/my' })
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
  const validationMessage = validateRejectReason()
  if (validationMessage) {
    actionError.value = validationMessage
    return
  }
  uni.showModal({
    title: '拒绝邀请',
    content: `确定拒绝加入「${invitation.value.familyName}」的邀请吗？`,
    success: (modalResult) => {
      if (modalResult.confirm) reject()
    }
  })
}

function validateRejectReason() {
  return validateTextFields([
    { value: rejectReason.value, label: '拒绝原因', kind: 'multiLine', maxLength: 200 }
  ])
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
  const validationMessage = validateRejectReason()
  if (validationMessage) {
    actionError.value = validationMessage
    return
  }
  acting.value = 'reject'
  actionError.value = ''
  result.value = ''
  try {
    invitation.value = await rejectInvitation(invitation.value.invitationId, {
      reason: optionalText(rejectReason.value)
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
  padding-top: 28rpx;
}

.invite-detail-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.detail-seal {
  flex-shrink: 0;
}

.invite-letter {
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.42), transparent 38%),
    rgba(255, 252, 245, 0.52);
}

.invite-letter.pending {
  box-shadow: inset 0 0 0 1rpx rgba(168, 59, 45, 0.08);
}

.invite-letter-head {
  display: flex;
  align-items: stretch;
  gap: 18rpx;
  padding: 24rpx 0 18rpx;
}

.invite-letter-copy {
  flex: 1;
  min-width: 0;
  padding-left: 2rpx;
}

.invite-letter-kicker {
  display: block;
  color: var(--archive-cinnabar);
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 4rpx;
}

.invite-family-name {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 42rpx;
  font-weight: 700;
  line-height: 1.3;
}

.invite-target-line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12rpx;
  margin-top: 18rpx;
}

.invite-target-label {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

.invite-target-tag {
  min-height: 52rpx;
  padding: 0 18rpx;
  font-size: 28rpx;
}

.invite-target-hint {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.invite-letter-spine {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  width: 72rpx;
  color: rgba(255, 255, 255, 0.92);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 24rpx;
  font-weight: 700;
  letter-spacing: 4rpx;
}

.invite-pending-mark {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin: 0 0 8rpx;
  padding: 0 0 18rpx;
  border-bottom: 1rpx dashed rgba(168, 59, 45, 0.22);
}

.invite-pending-seal {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 96rpx;
  min-height: 96rpx;
  border: 3rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 1.1;
  text-align: center;
  transform: rotate(-8deg);
}

.invite-pending-note {
  flex: 1;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.55;
}

.invite-fields {
  padding-bottom: 8rpx;
}

.invite-field-row {
  align-items: flex-start;
}

.invite-field-label {
  flex-shrink: 0;
  width: 168rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.5;
}

.invite-field-value {
  flex: 1;
  min-width: 0;
  color: var(--archive-ink);
  font-size: 24rpx;
  line-height: 1.55;
  text-align: right;
}

.invite-message-panel,
.invite-benefits-panel,
.invite-action-panel,
.invite-resolved-panel {
  margin-bottom: 24rpx;
}

.invite-message {
  display: block;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.75;
}

.invite-benefit-row {
  min-height: 72rpx;
}

.invite-benefit-text {
  flex: 1;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.55;
}

.invite-detail-page :deep(.mini-notice) {
  margin-bottom: 20rpx;
}

.invite-reject-input {
  margin-bottom: 16rpx;
}

.invite-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
}

.invite-action-single {
  margin-top: 16rpx;
  align-self: flex-start;
}

.invite-resolved-line {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  padding-bottom: 4rpx;
  border-bottom: 1rpx solid var(--archive-line);
}

.archive-status-tag {
  align-self: flex-start;
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx solid var(--archive-line);
  padding: 0 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  line-height: 1.2;
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

.invite-resolved-text {
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.6;
}

.invite-error,
.invite-result {
  display: block;
  margin-top: 8rpx;
}
</style>
