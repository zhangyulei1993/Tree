<template>
  <view class="archive-page sent-page">
    <MiniBackHome />

    <view v-if="familyName" class="sent-head archive-page-head">
      <view>
        <text class="archive-kicker">Sent Invitations</text>
        <text class="archive-title">{{ pageTitle }}</text>
        <text class="archive-subtitle">{{ pageSubtitle }}</text>
      </view>
      <view class="archive-seal">{{ familySealLetter }}</view>
    </view>

    <view v-if="familyName" class="sent-context">
      <text class="archive-chip">{{ familyRoleLabel }}</text>
      <text class="context-back" @click="openFamilyOverview">返回详情</text>
    </view>

    <template v-if="shareResult">
      <view class="share-panel archive-form-panel">
        <view class="share-panel-head">
          <view class="archive-section-head share-section-head">
            <text class="archive-section-title">待发送邀请</text>
            <text class="archive-section-subtitle">
              {{ shareInvitationSubtitle }}
            </text>
          </view>
          <text class="archive-status-tag is-cinnabar">待接受</text>
        </view>

        <MiniNotice tone="security" title="新邀请已生成">
          对方点开后可先查看完整邀请信息，再决定是否接受。此前生成的邀请链接已失效。
        </MiniNotice>

        <view class="share-fields archive-list">
          <view class="archive-row share-field-row">
            <text class="share-field-label">邀请加入</text>
            <text class="share-field-value">{{ shareResult.invitation.familyName }}</text>
          </view>
          <view class="archive-row share-field-row">
            <text class="share-field-label">{{ isPendingMemberInvitation(shareResult.invitation) ? '成员身份' : '确认身份' }}</text>
            <text class="share-field-value">{{ inviteTargetText(shareResult.invitation) }}</text>
          </view>
          <view class="archive-row share-field-row">
            <text class="share-field-label">邀请方式</text>
            <text class="share-field-value">{{ inviteChannelText(shareResult.invitation.inviteChannel) }}</text>
          </view>
          <view class="archive-row share-field-row">
            <text class="share-field-label">链接状态</text>
            <text class="share-field-value share-link-active">新生成 · 待分享</text>
          </view>
          <view class="archive-row share-field-row">
            <text class="share-field-label">有效期至</text>
            <text class="share-field-value">{{ formatDate(shareResult.invitation.expiredAt) }}</text>
          </view>
        </view>

        <view v-if="shareResult.invitation.inviteMessage" class="share-message">
          <text class="share-message-label">给对方的说明</text>
          <text class="share-message-content">{{ shareResult.invitation.inviteMessage }}</text>
        </view>

        <view class="share-actions">
          <!-- #ifdef MP-WEIXIN -->
          <button class="wechat-share-button wechat-share-button--compact" open-type="share">发送给微信好友</button>
          <!-- #endif -->
          <MiniButton size="sm" variant="secondary" @click="copyInviteLink">复制邀请链接</MiniButton>
          <MiniButton size="sm" variant="ghost" @click="clearShareResult">返回邀请列表</MiniButton>
        </view>
      </view>
    </template>

    <template v-else>
      <view v-if="familyName && memberId" class="member-invite-panel archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">邀请本人绑定节点</text>
          <text class="archive-section-subtitle">对方接受后会加入家庭，并成为这个家庭树节点本人</text>
        </view>
        <view class="archive-list member-invite-summary">
          <view class="archive-row share-field-row">
            <text class="share-field-label">目标节点</text>
            <text class="share-field-value">{{ memberName || '未命名成员' }}</text>
          </view>
        </view>
        <MiniNotice tone="security" title="接受后的结果">
          对方会绑定到“{{ memberName || '该成员' }}”节点；不会新增其他成员节点。
        </MiniNotice>
        <textarea
          v-model.trim="memberInviteMessage"
          class="tree-textarea family-invite-message"
          maxlength="300"
          placeholder="给对方的说明（可选）"
        />
        <MiniButton
          size="sm"
          :loading="actingType === 'create-member'"
          :disabled="Boolean(actingId) || actingType === 'create-member'"
          @click="createMemberBindingInvitation"
        >
          生成节点绑定邀请
        </MiniButton>
      </view>

      <view v-if="familyName && !isNodeInviteMode" class="family-invite-panel archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">邀请加入家庭</text>
          <text class="archive-section-subtitle">适合确认对方属于本家庭，但暂不确认具体家庭树节点</text>
        </view>
        <MiniNotice tone="security" title="接受后的结果">
          对方会加入家庭；接受前不会生成成员，接受后会按身份标签新增一位待接入成员，后续再确认分支。
        </MiniNotice>
        <input
          v-model.trim="pendingMemberLabel"
          class="tree-input pending-member-label"
          maxlength="40"
          placeholder="待确认身份标签，如：二房长子（待确认）"
        />
        <textarea
          v-model.trim="familyInviteMessage"
          class="tree-textarea family-invite-message"
          maxlength="300"
          placeholder="给对方的说明（可选）"
        />
        <MiniNotice tone="security" title="身份待确认">
          对方接受前不会出现在成员表；接受后会以你填写的身份标签进入家庭，便于后续区分多个待确认成员。
        </MiniNotice>
        <MiniButton
          size="sm"
          :loading="actingType === 'create-family'"
          :disabled="Boolean(actingId) || actingType === 'create-family'"
          @click="createPendingFamilyInvitation"
        >
          生成加入家庭邀请
        </MiniButton>
      </view>

      <view v-if="loading && invitations.length === 0" class="sent-state archive-form-panel">
        <MiniEmptyState symbol="…" title="正在加载" description="正在加载发出的家庭邀请..." />
      </view>

      <view v-else-if="loadError && invitations.length === 0" class="sent-state archive-form-panel">
        <MiniEmptyState
          symbol="!"
          title="加载失败"
          :description="loadError"
          action-text="重新加载"
          @action="loadInvitations"
        />
      </view>

      <view v-else-if="invitations.length === 0" class="sent-state archive-form-panel">
        <MiniEmptyState
          symbol="邀"
          title="暂无发出的家庭邀请"
          description="此处用于邀请加入家庭；如需绑定指定家庭树节点，请前往成员表或点击未绑定节点。"
          action-text="前往成员表绑定节点"
          @action="openMembers"
        />
      </view>

      <template v-else>
        <view class="sent-summary">
          <text class="archive-chip">待对方处理 {{ pendingCount }}</text>
          <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadInvitations">
            刷新列表
          </MiniButton>
        </view>

        <view v-for="group in invitationGroups" :key="group.key" class="sent-group archive-form-panel">
          <view class="archive-section-head">
            <text class="archive-section-title">{{ group.title }}</text>
            <text class="archive-section-subtitle">{{ group.subtitle }}</text>
          </view>
          <view class="sent-list archive-list">
            <view
              v-for="item in group.items"
              :key="item.invitationId"
              class="sent-item"
              :class="{ actionable: canRegenerate(item) }"
            >
              <view class="sent-item-head">
                <view class="archive-row-main">
                  <text class="archive-row-title">{{ inviteTargetText(item) }}</text>
                  <text class="archive-row-desc">{{ item.inviteMessage || '未填写邀请说明' }}</text>
                </view>
                <text class="archive-status-tag" :class="invitationStatusClass(displayStatus(item))">
                  {{ invitationStatusText(displayStatus(item)) }}
                </text>
              </view>

              <view class="sent-meta">
                <text>邀请方式：{{ inviteChannelText(item.inviteChannel) }}</text>
                <text>链接状态：{{ inviteLinkStatusText(item) }}</text>
                <text class="sent-meta-weak">有效期至：{{ formatDate(item.expiredAt) }}</text>
                <text class="sent-meta-weak">创建时间：{{ formatDate(item.createdAt) }}</text>
              </view>

              <view v-if="canRegenerate(item)" class="sent-actions">
                <MiniButton
                  size="sm"
                  :loading="actingId === item.invitationId && actingType === 'regenerate'"
                  :disabled="Boolean(actingId)"
                  @click="confirmRegenerate(item)"
                >
                  重新生成
                </MiniButton>
                <MiniButton
                  v-if="item.status === 'PENDING' && !isExpired(item)"
                  size="sm"
                  variant="secondary"
                  :loading="actingId === item.invitationId && actingType === 'cancel'"
                  :disabled="Boolean(actingId)"
                  @click="confirmCancel(item)"
                >
                  取消
                </MiniButton>
              </view>
            </view>
          </view>
        </view>
      </template>

      <text v-if="operationError" class="tree-field-error sent-error">{{ operationError }}</text>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onHide, onLoad, onShareAppMessage, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import {
  cancelInvitation,
  createInvitation,
  createFamilyInvitation,
  listFamilyInvitations,
  regenerateInvitation
} from '@/api/invitations'
import { buildHomeSharePayload, buildInviteSharePayload } from '@/features/share/wechatShare'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { CreatedInvitation, Invitation } from '@/types/api'
import { optionalText, validateTextField, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const familyName = ref('')
const familySurname = ref('')
const familyRole = ref('')
const inviteMode = ref('')
const memberId = ref('')
const memberName = ref('')
const invitations = ref<Invitation[]>([])
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const actingId = ref<number | string | null>(null)
const actingType = ref<'' | 'cancel' | 'regenerate' | 'create-family' | 'create-member'>('')
const shareResult = ref<CreatedInvitation | null>(null)
const familyInviteMessage = ref('')
const memberInviteMessage = ref('')
const pendingMemberLabel = ref('')
const familyRoleLabel = computed(() => (familyRole.value === 'FOUNDER' ? '创建者' : '管理员'))
const familySealLetter = computed(() => familySurname.value.slice(0, 1) || familyName.value.slice(0, 1) || '邀')
const isNodeInviteMode = computed(() => inviteMode.value === 'node' && Boolean(memberId.value))
const pageTitle = computed(() => (isNodeInviteMode.value ? '邀请本人绑定节点' : '邀请加入家庭'))
const pageSubtitle = computed(() =>
  isNodeInviteMode.value
    ? `${familyName.value} · 接受后成为指定家庭树节点本人`
    : `${familyName.value} · 接受后先加入家庭，节点身份可后续确认`
)
const pendingCount = computed(() =>
  invitations.value.filter((item) => displayStatus(item) === 'PENDING').length
)
const shareInvitationSubtitle = computed(() => {
  const invitation = shareResult.value?.invitation
  if (!invitation) return ''
  return isPendingMemberInvitation(invitation)
    ? `对方将先加入家庭，身份标签：${inviteTargetText(invitation)}`
    : `对方将绑定“${invitation.targetMemberName}”节点`
})
const invitationGroups = computed(() => {
  const current = invitations.value.filter((item) => displayStatus(item) === 'PENDING')
  const history = invitations.value.filter((item) => displayStatus(item) !== 'PENDING')
  return [
    { key: 'current', title: '待对方处理', subtitle: '当前仍有效的邀请', items: current },
    { key: 'history', title: '历史邀请', subtitle: '已接受、拒绝、取消或过期的记录', items: history }
  ].filter((group) => group.items.length > 0)
})

function currentRoute() {
  const query = [`familyId=${encodeURIComponent(familyId.value)}`]
  if (inviteMode.value) query.push(`mode=${encodeURIComponent(inviteMode.value)}`)
  if (memberId.value) query.push(`memberId=${encodeURIComponent(memberId.value)}`)
  if (memberName.value) query.push(`memberName=${encodeURIComponent(memberName.value)}`)
  return `/pages/invite/sent?${query.join('&')}`
}

function invitationStatusClass(status: string) {
  if (status === 'PENDING') return 'is-cinnabar'
  if (status === 'ACCEPTED') return 'is-ink'
  return 'is-muted'
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

function inviteLinkStatusText(item: Invitation) {
  const status = displayStatus(item)
  if (status === 'PENDING') return '链接有效'
  if (status === 'EXPIRED') return '链接已失效'
  if (status === 'ACCEPTED') return '对方已接受'
  if (status === 'REJECTED') return '对方已拒绝'
  if (status === 'CANCELLED') return '链接已取消'
  return '不可分享'
}

function isPendingMemberInvitation(item: Invitation) {
  return item.inviteType === 'JOIN_FAMILY_PENDING_MEMBER'
}

function inviteTargetText(item: Invitation) {
  return isPendingMemberInvitation(item)
    ? item.pendingMemberLabel || item.targetMemberName || '待确认成员'
    : item.targetMemberName
}

function resetTransientUI() {
  shareResult.value = null
  operationError.value = ''
  actingId.value = null
  actingType.value = ''
}

function resetPageData() {
  loading.value = false
  loadError.value = ''
  familyName.value = ''
  familySurname.value = ''
  familyRole.value = ''
  inviteMode.value = ''
  memberId.value = ''
  memberName.value = ''
  invitations.value = []
  resetTransientUI()
}

async function loadInvitations() {
  session.restoreSession()
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
    return
  }
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(currentRoute())
    return
  }
  const isInitialLoad = invitations.value.length === 0 && !shareResult.value
  loading.value = isInitialLoad
  loadError.value = ''
  operationError.value = ''
  try {
    const [family, invitationRows] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyInvitations(familyId.value)
    ])
    familyName.value = family.familyName
    familySurname.value = family.familySurname
    familyRole.value = family.role
    invitations.value = invitationRows
  } catch (error) {
    loadError.value = apiErrorMessage(error, '发出的家庭邀请加载失败。')
  } finally {
    loading.value = false
  }
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

function isExpired(item: Invitation) {
  return item.status === 'PENDING' && new Date(item.expiredAt).getTime() <= Date.now()
}

function displayStatus(item: Invitation) {
  return isExpired(item) ? 'EXPIRED' : item.status
}

function canRegenerate(item: Invitation) {
  return item.inviteChannel === 'SHARE_LINK'
    && (item.status === 'PENDING' || item.status === 'EXPIRED')
}

function openMembers() {
  uni.navigateTo({ url: `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}` })
}

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function confirmCancel(item: Invitation) {
  uni.showModal({
    title: '取消邀请',
    content: `确定取消发给「${inviteTargetText(item)}」的邀请吗？`,
    success: (result) => {
      if (result.confirm) cancel(item)
    }
  })
}

async function cancel(item: Invitation) {
  actingId.value = item.invitationId
  actingType.value = 'cancel'
  operationError.value = ''
  try {
    await cancelInvitation(item.invitationId, { reason: '小程序已发邀请页面取消' })
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '取消邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

function confirmRegenerate(item: Invitation) {
  uni.showModal({
    title: '重新生成邀请',
    content: '重新生成后，旧邀请链接将立即失效。是否继续？',
    success: (result) => {
      if (result.confirm) regenerate(item)
    }
  })
}

async function createPendingFamilyInvitation() {
  const labelError = validateTextField({
    label: '待确认身份标签',
    value: pendingMemberLabel.value,
    kind: 'name',
    required: true,
    maxLength: 40
  })
  const validationMessage = labelError || validateTextFields([
    { value: familyInviteMessage.value, label: '邀请说明', kind: 'multiLine', maxLength: 300 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  actingType.value = 'create-family'
  operationError.value = ''
  try {
    shareResult.value = await createFamilyInvitation(familyId.value, {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: familyInviteMessage.value.trim() || undefined,
      pendingMemberLabel: pendingMemberLabel.value.trim(),
      familyRoleAfterAccept: 'MEMBER'
    })
    familyInviteMessage.value = ''
    pendingMemberLabel.value = ''
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '生成家庭邀请失败。')
  } finally {
    actingType.value = ''
  }
}

async function createMemberBindingInvitation() {
  if (!memberId.value) return
  const validationMessage = validateTextFields([
    { value: memberInviteMessage.value, label: '邀请说明', kind: 'multiLine', maxLength: 300 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  actingType.value = 'create-member'
  operationError.value = ''
  try {
    shareResult.value = await createInvitation(familyId.value, memberId.value, {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: memberInviteMessage.value.trim() || undefined,
      familyRoleAfterAccept: 'MEMBER'
    })
    memberInviteMessage.value = ''
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '生成节点邀请失败。')
  } finally {
    actingType.value = ''
  }
}

async function regenerate(item: Invitation) {
  actingId.value = item.invitationId
  actingType.value = 'regenerate'
  operationError.value = ''
  try {
    shareResult.value = await regenerateInvitation(item.invitationId)
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '重新生成邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

function copyInviteLink() {
  const token = shareResult.value?.inviteToken
  if (!token) return
  uni.setClipboardData({
    data: `https://tree.bigbigboy.cn/invite/${encodeURIComponent(token)}`,
    success: () => uni.showToast({ title: '邀请链接已复制', icon: 'success' })
  })
}

function clearShareResult() {
  shareResult.value = null
}

function safeDecodeRouteParam(value: unknown) {
  const text = String(value || '')
  if (!text) return ''
  try {
    return decodeURIComponent(text)
  } catch {
    return text
  }
}

onShareAppMessage(() => {
  const invitation = shareResult.value?.invitation
  if (shareResult.value && invitation) {
    return buildInviteSharePayload({
      inviteToken: shareResult.value.inviteToken,
      familyName: invitation.familyName,
      targetMemberName: inviteTargetText(invitation),
      inviteType: invitation.inviteType
    })
  }
  return buildHomeSharePayload()
})

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  inviteMode.value = String(options?.mode || '')
  memberId.value = String(options?.memberId || '')
  memberName.value = safeDecodeRouteParam(options?.memberName)
  session.restoreSession()
  if (session.isLoggedIn && familyId.value) {
    loading.value = true
  }
})
onShow(loadInvitations)
onHide(resetTransientUI)
onUnload(resetPageData)
</script>

<style scoped>
.sent-page {
  padding-top: 28rpx;
}

.sent-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.sent-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 22rpx;
}

.context-back {
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.4;
}

.sent-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 22rpx;
}

.sent-group,
.share-panel,
.member-invite-panel,
.family-invite-panel,
.sent-state {
  margin-bottom: 24rpx;
}

.pending-member-label,
.family-invite-message {
  margin-bottom: 14rpx;
}

.member-invite-summary {
  margin: 8rpx 0 14rpx;
}

.family-invite-panel :deep(.mini-notice) {
  margin-bottom: 16rpx;
}

.sent-list {
  margin-top: 8rpx;
}

.sent-item {
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.sent-item:last-child {
  border-bottom: 0;
}

.sent-item.actionable {
  background: linear-gradient(90deg, rgba(168, 59, 45, 0.03), transparent 42%);
}

.sent-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.archive-status-tag {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx solid var(--archive-line);
  padding: 0 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  line-height: 1.2;
}

.archive-status-tag.is-cinnabar {
  border-color: rgba(168, 59, 45, 0.28);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
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

.sent-meta {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.sent-meta-weak {
  opacity: 0.88;
}

.sent-actions,
.share-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.share-panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.share-section-head {
  flex: 1;
  min-width: 0;
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: 0;
}

.share-panel :deep(.mini-notice) {
  margin: 18rpx 0;
}

.share-fields {
  margin-top: 8rpx;
}

.share-field-row {
  align-items: flex-start;
}

.share-field-label {
  flex-shrink: 0;
  width: 168rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.5;
}

.share-field-value {
  flex: 1;
  min-width: 0;
  color: var(--archive-ink);
  font-size: 24rpx;
  line-height: 1.55;
  text-align: right;
}

.share-link-active {
  color: var(--archive-cinnabar);
}

.share-message {
  margin-top: 16rpx;
  padding-top: 14rpx;
  border-top: 1rpx dashed var(--archive-line);
}

.share-message-label,
.share-message-content {
  display: block;
}

.share-message-label {
  margin-bottom: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

.share-message-content {
  color: var(--archive-ink);
  font-size: 24rpx;
  line-height: 1.65;
}

.sent-error {
  display: block;
  margin-top: 8rpx;
}
</style>
