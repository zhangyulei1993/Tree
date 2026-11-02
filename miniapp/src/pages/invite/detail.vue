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
          <text class="archive-kicker">Family Invitation</text>
          <text class="archive-title">{{ invitation.status === 'PENDING' ? '确认加入家庭' : '邀请状态' }}</text>
          <text class="archive-subtitle">{{ invitationSubtitle }}</text>
        </view>
        <view class="archive-seal detail-seal">谱</view>
      </view>

      <view class="invite-letter" :class="{ pending: invitation.status === 'PENDING' }">
        <view class="invite-letter-head">
          <view class="invite-letter-copy">
            <text class="invite-family-name">{{ invitation.familyName }}</text>
            <view class="invite-target-line">
              <text class="invite-target-label">绑定为</text>
              <text class="invite-target-tag archive-paper-tag">{{ inviteTargetText }}</text>
            </view>
            <text v-if="invitation.status === 'PENDING'" class="invite-target-hint">接受后将绑定该节点。</text>
          </view>
        </view>

        <view class="invite-fields archive-list">
          <view class="archive-row invite-field-row">
            <text class="invite-field-label">邀请人</text>
            <text class="invite-field-value">
              {{ invitation.inviterDisplayName || '家庭管理员' }}（{{ inviterRoleText(invitation.inviterRole) }}）
            </text>
          </view>
          <view class="archive-row invite-field-row">
            <text class="invite-field-label">有效期至</text>
            <text class="invite-field-value">{{ formatDate(invitation.expiredAt) }}</text>
          </view>
        </view>
      </view>

      <view v-if="invitation.status === 'PENDING'" class="archive-form-panel invite-action-panel invite-action-panel-primary">
        <template v-if="session.isLoggedIn && session.isProfileComplete">
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
          <textarea
            v-if="showRejectReason"
            v-model.trim="rejectReason"
            class="tree-textarea invite-reject-input"
            maxlength="300"
            placeholder="拒绝原因（可选）"
          />
        </template>
        <template v-else-if="session.isLoggedIn">
          <MiniButton size="sm" class="invite-action-single" @click="requireProfileComplete">完善昵称后继续</MiniButton>
        </template>
        <template v-else>
          <MiniButton size="sm" class="invite-action-single" @click="requireLogin">登录后继续</MiniButton>
        </template>
      </view>

      <view v-if="invitation.inviteMessage" class="archive-form-panel invite-message-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">邀请人留言</text>
        </view>
        <text class="invite-message">{{ invitation.inviteMessage }}</text>
      </view>

      <MiniNotice v-if="invitation.status === 'PENDING'" tone="security" title="接受前请确认">
        {{ acceptNoticeText }}
      </MiniNotice>

      <view v-else class="archive-form-panel invite-resolved-panel">
        <view class="invite-resolved-line">
          <text class="archive-status-tag" :class="resolvedStatusClass(invitation.status)">
            {{ invitationStatusText(invitation.status) }}
          </text>
          <text class="invite-resolved-text">{{ resolvedMessage }}</text>
        </view>
        <MiniButton
          v-if="acceptedByCurrentUser"
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
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { acceptInvitation, getInvitationDetail, rejectInvitation } from '@/api/invitations'
import { apiErrorMessage } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Invitation } from '@/types/api'
import { optionalText, validateTextFields } from '@/utils/inputValidation'
import { readQueryParam, safeDecodeURIComponent } from '@/utils/queryString'
import { buildInvitationDetailSharePayload, loadShareConfigs } from '@/features/share/wechatShare'

const session = useSessionStore()
const inviteToken = ref('')
const invitation = ref<Invitation | null>(null)
const rejectReason = ref('')
const showRejectReason = ref(false)
const loading = ref(false)
const acting = ref<'' | 'accept' | 'reject'>('')
const loadError = ref('')
const actionError = ref('')
const result = ref('')
const invitationSubtitle = computed(() => invitation.value?.status === 'PENDING' ? '核对身份后处理' : '该邀请已无法使用')
const inviteTargetText = computed(() => invitation.value?.targetMemberName || '')
const acceptNoticeText = computed(() => '请确认上方姓名为你的身份；不符请拒绝并联系邀请人。')
const acceptedByCurrentUser = computed(() =>
  invitation.value?.status === 'ACCEPTED' &&
  Boolean(session.isLoggedIn && invitation.value.acceptedByUserId && session.user?.id) &&
  String(invitation.value.acceptedByUserId) === String(session.user?.id)
)
const resolvedMessage = computed(() => {
  if (!invitation.value) return ''
  if (invitation.value.status === 'ACCEPTED') {
    const name = invitation.value.acceptedByDisplayName?.trim()
    const recipient = name ? `「${name}」` : '其他账号'
    return acceptedByCurrentUser.value
      ? '你已接受该邀请，可在“我的家庭”查看。'
      : `这份邀请已由${recipient}接受，不能再次使用。请联系家庭管理员获取新的邀请。`
  }
  return `该邀请已${invitationStatusText(invitation.value.status)}，不能继续处理。请联系家庭管理员获取新的邀请。`
})

function currentRoute() {
  return `/pages/invite/detail?inviteToken=${encodeURIComponent(inviteToken.value)}`
}

function parseInviteToken(options?: Record<string, unknown>) {
  const direct = options?.inviteToken || options?.token
  if (direct) return String(direct)
  const scene = options?.scene ? safeDecodeURIComponent(String(options.scene)) : ''
  if (!scene) return ''
  if (!scene.includes('=')) return scene
  return readQueryParam(scene, ['inviteToken', 'token']) || scene
}

function resolvedStatusClass(status: string) {
  if (status === 'ACCEPTED') return 'is-ink'
  return 'is-muted'
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
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
  showRejectReason.value = false
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
  if (!showRejectReason.value) {
    showRejectReason.value = true
    actionError.value = ''
    return
  }
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
  inviteToken.value = parseInviteToken(options)
  session.restoreSession()
  loadInvitation()
  void loadShareConfigs()
})

onShareAppMessage(() => invitation.value
  ? buildInvitationDetailSharePayload({ inviteToken: inviteToken.value, familyName: invitation.value.familyName, status: invitation.value.status })
  : { title: '家庭邀请', path: currentRoute(), imageUrl: '' })
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
  margin-bottom: 16rpx;
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
  padding: 20rpx 0 14rpx;
}

.invite-letter-copy {
  flex: 1;
  min-width: 0;
  padding-left: 2rpx;
}

.invite-family-name {
  display: block;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 38rpx;
  font-weight: 700;
  line-height: 1.3;
}

.invite-target-line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12rpx;
  margin-top: 12rpx;
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
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.invite-fields {
  padding-bottom: 0;
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
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  word-break: break-all;
}

.invite-message-panel,
.invite-action-panel,
.invite-resolved-panel {
  margin-bottom: 16rpx;
}

.invite-message {
  display: block;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.75;
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 4;
  word-break: break-all;
}

.invite-detail-page :deep(.mini-notice) {
  margin-bottom: 16rpx;
}

.invite-reject-input {
  margin-top: 16rpx;
}

.invite-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10rpx;
}

.invite-actions :deep(.mini-button) {
  width: 100%;
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
