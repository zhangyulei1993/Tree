<template>
  <view class="archive-page sent-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />

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
      <text class="context-back" @click="openFamilyOverview">{{ returnLabel }}</text>
    </view>

    <MiniNotice v-if="guideMode && familyName" tone="security" title="最后一步：完成邀请">
      {{ showPhoneBindGate ? '绑定手机号后即可发起邀请。' : '生成邀请后即可完成引导。' }}
    </MiniNotice>

    <view v-if="activeMemberInvitation" class="share-panel archive-form-panel">
      <view class="share-panel-head">
        <view class="archive-section-head share-section-head">
          <text class="archive-section-title">当前节点已有邀请</text>
          <text class="archive-section-subtitle">
            {{ shareInvitationSubtitle }} · 至 {{ formatDate(activeMemberInvitation.expiredAt) }}
          </text>
        </view>
        <text class="archive-status-tag is-cinnabar">待接受</text>
      </view>

      <MiniNotice v-if="!canShareCurrentInvitation" tone="security" title="邀请仍有效，但本机没有原链接">
        邀请记录未丢失。出于安全考虑，链接只保存在首次生成的设备；如需从当前设备发送，请重新生成。重新生成后旧链接会失效。
      </MiniNotice>

      <view class="share-actions">
        <!-- #ifdef MP-WEIXIN -->
        <button v-if="canShareCurrentInvitation" class="wechat-share-button wechat-share-button--compact" open-type="share">发送给微信好友</button>
        <!-- #endif -->
        <MiniButton
          size="sm"
          :loading="actingId === activeMemberInvitation.invitationId && actingType === 'regenerate'"
          :disabled="Boolean(actingId)"
          @click="confirmRegenerate(activeMemberInvitation)"
        >
          {{ canShareCurrentInvitation ? '重新生成邀请' : '重新生成并发送' }}
        </MiniButton>
        <MiniButton
          size="sm"
          variant="secondary"
          :loading="actingId === activeMemberInvitation.invitationId && actingType === 'cancel'"
          :disabled="Boolean(actingId)"
          @click="confirmCancel(activeMemberInvitation)"
        >
          取消邀请
        </MiniButton>
        <MiniButton size="sm" variant="ghost" @click="toggleInvitationHistory">
          {{ showHistory ? '收起邀请记录' : `查看邀请记录（${invitations.length}）` }}
        </MiniButton>
      </view>
    </view>

    <view v-else-if="familyName && memberId" class="member-invite-panel archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">生成节点邀请</text>
        <text class="archive-section-subtitle">对方接受后绑定该节点</text>
      </view>
      <view v-if="showPhoneBindGate" class="invite-phone-bind">
        <MiniNotice tone="security" title="需要先绑定手机号">
          绑定手机号后即可生成邀请。
        </MiniNotice>
        <!-- #ifdef MP-WEIXIN -->
        <button
          class="wechat-phone-bind-button"
          open-type="getPhoneNumber"
          :disabled="bindingWechatPhone"
          @getphonenumber="bindWithWechatPhone"
        >
          {{ bindingWechatPhone ? '绑定中...' : '绑定微信手机号' }}
        </button>
        <!-- #endif -->
      </view>
      <view class="archive-list member-invite-summary">
        <view class="archive-row share-field-row">
          <text class="share-field-label">目标节点</text>
          <text class="share-field-value">{{ memberName || '未命名成员' }}</text>
        </view>
      </view>
      <textarea
        v-model.trim="memberInviteMessage"
        class="tree-textarea family-invite-message"
        maxlength="300"
        placeholder="给对方的说明（可选）"
      />
      <MiniButton
        size="sm"
        :loading="actingType === 'create-member'"
        :disabled="showPhoneBindGate || Boolean(actingId) || actingType === 'create-member'"
        @click="createMemberBindingInvitation"
      >
        生成邀请
      </MiniButton>
      <MiniButton v-if="invitations.length > 0" size="sm" variant="ghost" class="history-trigger" @click="toggleInvitationHistory">
        {{ showHistory ? '收起邀请记录' : `查看邀请记录（${invitations.length}）` }}
      </MiniButton>
    </view>

    <view v-if="familyName && !memberId" class="archive-segment-tabs invitation-tabs">
      <view class="archive-segment-tab" :class="{ active: inviteTab === 'INVITABLE' }" @tap="inviteTab = 'INVITABLE'">可邀请</view>
      <view class="archive-segment-tab" :class="{ active: inviteTab === 'PENDING', 'has-pending': pendingCount > 0 }" @tap="inviteTab = 'PENDING'">待处理 {{ pendingCount ? `(${pendingCount})` : '' }}</view>
      <view class="archive-segment-tab" :class="{ active: inviteTab === 'HISTORY' }" @tap="inviteTab = 'HISTORY'">历史记录</view>
    </view>

    <view v-if="!memberId && inviteTab === 'INVITABLE' && availableInviteMembers.length" class="member-picker archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">选择成员发起邀请</text>
        <text class="archive-section-subtitle">仅展示尚未绑定账号的成员节点</text>
      </view>
      <view class="archive-list">
        <view
          v-for="member in availableInviteMembers"
          :key="member.memberId"
          class="archive-row"
          @tap="openMemberInvitation(member.memberId, member.name)"
        >
          <view class="archive-row-main">
            <text class="archive-row-title">{{ member.name || '未命名成员' }}</text>
            <text class="archive-row-desc">{{ member.gender === 'MALE' ? '男性' : member.gender === 'FEMALE' ? '女性' : '性别待确认' }} · 发起本人邀请</text>
          </view>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </view>

    <view v-if="!memberId && inviteTab === 'INVITABLE' && !loading && !availableInviteMembers.length" class="sent-state archive-form-panel">
      <MiniEmptyState symbol="邀" title="暂无可邀请成员" description="成员均已绑定或已有待处理邀请。" />
    </view>

    <template v-if="memberId ? showHistory : inviteTab !== 'INVITABLE'">
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

      <view v-else-if="visibleInvitationGroups.length === 0" class="sent-state archive-form-panel">
        <MiniEmptyState
          symbol="邀"
          :title="inviteTab === 'PENDING' ? '暂无待处理邀请' : '暂无历史邀请记录'"
          :description="inviteTab === 'PENDING' ? '当前没有等待对方接受的邀请。' : '邀请被接受、拒绝、取消或失效后会显示在这里。'"
        />
      </view>

      <template v-else>
        <view class="sent-summary">
          <text class="archive-chip">待对方处理 {{ pendingCount }}</text>
          <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadInvitations">
            刷新列表
          </MiniButton>
        </view>

        <view v-for="group in visibleInvitationGroups" :key="group.key" class="sent-group archive-form-panel">
          <view class="archive-section-head">
            <text class="archive-section-title">{{ group.title }}</text>
            <text class="archive-section-subtitle">{{ group.subtitle }}</text>
          </view>
          <view class="sent-list archive-list">
            <view
              v-for="item in group.items"
              :key="item.invitationId"
              class="sent-item"
              :class="{ actionable: canRegenerate(item), reopenable: displayStatus(item) === 'PENDING' }"
              @tap="openInvitationForSending(item)"
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
                <text class="sent-meta-weak">至 {{ formatDate(item.expiredAt) }}</text>
                <text class="sent-meta-weak">创建于 {{ formatDate(item.createdAt) }}</text>
              </view>

              <view v-if="canRegenerate(item) && displayStatus(item) !== 'PENDING'" class="sent-actions">
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
    </template>

    <text v-if="operationError" class="tree-field-error sent-error">{{ operationError }}</text>

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
  listFamilyInvitations,
  regenerateInvitation
} from '@/api/invitations'
import { listFamilyMembers } from '@/api/members'
import { buildHomeSharePayload, buildInviteSharePayload, loadShareConfigs } from '@/features/share/wechatShare'
import { requestSubscribeMessages } from '@/features/notifications/subscribe'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { useSessionStore } from '@/stores/session'
import type { CreatedInvitation, FamilyMember, Invitation } from '@/types/api'
import { optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const returnTo = ref('')
const familyName = ref('')
const familySurname = ref('')
const familyRole = ref('')
const inviteMode = ref('')
const memberId = ref('')
const memberName = ref('')
const guideMode = ref(false)
const invitations = ref<Invitation[]>([])
const familyMembers = ref<FamilyMember[]>([])
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const actingId = ref<number | string | null>(null)
const actingType = ref<'' | 'cancel' | 'regenerate' | 'create-member'>('')
const bindingWechatPhone = ref(false)
const shareResult = ref<CreatedInvitation | null>(null)
const memberInviteMessage = ref('')
const showHistory = ref(false)
const inviteTab = ref<'INVITABLE' | 'PENDING' | 'HISTORY'>('INVITABLE')
const familyRoleLabel = computed(() => (familyRole.value === 'FOUNDER' ? '创建者' : '管理员'))
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : returnTo.value === 'health'
    ? `/pages/family/health?familyId=${encodeURIComponent(familyId.value)}`
    : `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : returnTo.value === 'health' ? '返回家庭健康' : '返回家庭详情')
const familySealLetter = computed(() => familySurname.value.slice(0, 1) || familyName.value.slice(0, 1) || '邀')
const pageTitle = computed(() => '邀请本人绑定节点')
const pageSubtitle = computed(() =>
  `${familyName.value} · 接受后成为指定家庭树节点本人`
)
const pendingCount = computed(() =>
  invitations.value.filter((item) => displayStatus(item) === 'PENDING').length
)
const showPhoneBindGate = computed(() => session.isLoggedIn && !session.user?.phoneVerified)
const pendingInviteMemberIds = computed(() => new Set(
  invitations.value
    .filter((item) => displayStatus(item) === 'PENDING')
    .map((item) => String(item.targetMemberId))
))
const availableInviteMembers = computed(() => familyMembers.value.filter((member) =>
  member.status === 'ACTIVE'
    && !member.boundUserId
    && member.userBindingPolicy !== 'NOT_REQUIRED'
    && !pendingInviteMemberIds.value.has(String(member.memberId))
))
const activeMemberInvitation = computed(() => {
  const freshInvitation = shareResult.value?.invitation
  if (freshInvitation) {
    const latest = invitations.value.find((item) => item.invitationId === freshInvitation.invitationId)
    if (!latest || displayStatus(latest) === 'PENDING') return latest || freshInvitation
  }
  return invitations.value.find((item) =>
    String(item.targetMemberId) === memberId.value && displayStatus(item) === 'PENDING'
  ) || null
})
const canShareCurrentInvitation = computed(() =>
  Boolean(
    shareResult.value?.inviteToken
    && activeMemberInvitation.value
    && shareResult.value.invitation.invitationId === activeMemberInvitation.value.invitationId
  )
)
const shareInvitationSubtitle = computed(() => {
  const invitation = activeMemberInvitation.value
  if (!invitation) return ''
  return `对方将绑定“${invitation.targetMemberName}”节点`
})
const invitationGroups = computed(() => {
  const current = invitations.value.filter((item) => displayStatus(item) === 'PENDING')
  const history = invitations.value.filter((item) => displayStatus(item) !== 'PENDING')
  return [
    { key: 'current', title: '待对方处理', subtitle: '当前仍有效的邀请', items: current },
    { key: 'history', title: '历史邀请', subtitle: '已接受、拒绝、取消或过期的记录', items: history }
  ].filter((group) => group.items.length > 0)
})
const visibleInvitationGroups = computed(() => {
  if (memberId.value) return invitationGroups.value
  if (inviteTab.value === 'PENDING') return invitationGroups.value.filter((group) => group.key === 'current')
  if (inviteTab.value === 'HISTORY') return invitationGroups.value.filter((group) => group.key === 'history')
  return []
})

loadShareConfigs()

function currentRoute() {
  const query = [`familyId=${encodeURIComponent(familyId.value)}`]
  if (inviteMode.value) query.push(`mode=${encodeURIComponent(inviteMode.value)}`)
  if (memberId.value) query.push(`memberId=${encodeURIComponent(memberId.value)}`)
  if (memberName.value) query.push(`memberName=${encodeURIComponent(memberName.value)}`)
  if (guideMode.value) query.push('guide=1')
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
  if (status === 'PENDING') {
    return hasStoredInviteToken(item) ? '可继续发送' : '邀请有效，需重新生成链接'
  }
  if (status === 'EXPIRED') return '链接已失效'
  if (status === 'ACCEPTED') return '对方已接受'
  if (status === 'REJECTED') return '对方已拒绝'
  if (status === 'CANCELLED') return '链接已取消'
  return '不可分享'
}

function hasStoredInviteToken(item: Invitation) {
  const stored = uni.getStorageSync(inviteShareStorageKey(item.invitationId)) as Partial<StoredInviteShareLink> | undefined
  return typeof stored?.inviteToken === 'string' && stored.inviteToken.length > 0 && stored.expiredAt === item.expiredAt
}

function inviteTargetText(item: Invitation) {
  return item.targetMemberName
}

function resetTransientUI() {
  shareResult.value = null
  showHistory.value = false
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
  familyMembers.value = []
  inviteTab.value = 'INVITABLE'
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
    const [family, invitationRows, members] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyInvitations(familyId.value),
      listFamilyMembers(familyId.value)
    ])
    familyName.value = family.familyName
    familySurname.value = family.familySurname
    familyRole.value = family.role
    invitations.value = invitationRows
    familyMembers.value = members
    restoreStoredInviteShareLink(invitationRows)
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

function openMemberInvitation(targetMemberId: number | string, targetMemberName: string) {
  uni.navigateTo({
    url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}&mode=node&memberId=${encodeURIComponent(String(targetMemberId))}&memberName=${encodeURIComponent(targetMemberName)}${['tree', 'health'].includes(returnTo.value) ? `&returnTo=${encodeURIComponent(returnTo.value)}` : ''}`
  })
}

function openFamilyOverview() {
  uni.redirectTo({ url: returnUrl.value })
}

function openInvitationForSending(item: Invitation) {
  if (displayStatus(item) !== 'PENDING') return
  if (String(item.targetMemberId) === memberId.value) {
    showHistory.value = false
    return
  }
  uni.navigateTo({
    url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}&mode=node&memberId=${encodeURIComponent(String(item.targetMemberId))}&memberName=${encodeURIComponent(item.targetMemberName)}${['tree', 'health'].includes(returnTo.value) ? `&returnTo=${encodeURIComponent(returnTo.value)}` : ''}`
  })
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
    removeStoredInviteShareLink(item.invitationId)
    if (shareResult.value?.invitation.invitationId === item.invitationId) {
      shareResult.value = null
    }
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '取消邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

function confirmRegenerate(item: Invitation) {
  if (!ensurePhoneBoundForInvitation()) return
  uni.showModal({
    title: '重新生成邀请',
    content: '重新生成后，旧邀请链接将立即失效。是否继续？',
    success: (result) => {
      if (result.confirm) regenerate(item)
    }
  })
}

async function createMemberBindingInvitation() {
  if (!memberId.value) return
  if (!ensurePhoneBoundForInvitation()) return
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
    await requestSubscribeMessages(['member', 'memberChange'])
    shareResult.value = await createInvitation(familyId.value, memberId.value, {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: memberInviteMessage.value.trim() || undefined,
      familyRoleAfterAccept: 'MEMBER'
    })
    persistInviteShareLink(shareResult.value)
    memberInviteMessage.value = ''
    markNewUserGuideMilestone('INVITE_GUIDE_OPENED')
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '生成节点邀请失败。')
  } finally {
    actingType.value = ''
  }
}

function ensurePhoneBoundForInvitation() {
  if (session.user?.phoneVerified) return true
  operationError.value = '请先点击上方“绑定微信手机号”，完成后再生成邀请。'
  return false
}

type PhoneNumberEvent = {
  detail?: {
    code?: string
    errMsg?: string
  }
}

type StoredInviteShareLink = {
  inviteToken: string
  expiredAt: string
}

function inviteShareStorageKey(invitationID: number | string) {
  return `tree:invite-share-link:${session.user?.id || 'guest'}:${familyId.value}:${invitationID}`
}

function persistInviteShareLink(result: CreatedInvitation) {
  if (!result.inviteToken) return
  uni.setStorageSync(inviteShareStorageKey(result.invitation.invitationId), {
    inviteToken: result.inviteToken,
    expiredAt: result.invitation.expiredAt
  } satisfies StoredInviteShareLink)
}

function removeStoredInviteShareLink(invitationID: number | string) {
  uni.removeStorageSync(inviteShareStorageKey(invitationID))
}

function restoreStoredInviteShareLink(rows: Invitation[]) {
  const targetRows = rows.filter((item) => String(item.targetMemberId) === memberId.value)
  const active = targetRows.find((item) => displayStatus(item) === 'PENDING')
  for (const item of targetRows) {
    if (item !== active) removeStoredInviteShareLink(item.invitationId)
  }
  if (!active) {
    shareResult.value = null
    return
  }
  const stored = uni.getStorageSync(inviteShareStorageKey(active.invitationId)) as Partial<StoredInviteShareLink> | undefined
  if (typeof stored?.inviteToken !== 'string' || !stored.inviteToken || stored.expiredAt !== active.expiredAt) {
    if (shareResult.value?.invitation.invitationId !== active.invitationId) shareResult.value = null
    return
  }
  shareResult.value = { invitation: active, inviteToken: stored.inviteToken }
}

async function bindWithWechatPhone(event: PhoneNumberEvent) {
  const phoneCode = event.detail?.code || ''
  if (!phoneCode) {
    const errMsg = event.detail?.errMsg || ''
    operationError.value = /deny|cancel|拒绝|取消/i.test(errMsg)
      ? '你已取消手机号授权。'
      : '手机号授权失败，请稍后重试。'
    return
  }
  bindingWechatPhone.value = true
  operationError.value = ''
  try {
    await session.bindWechatPhone({ phoneCode })
    await session.refreshMe().catch(() => undefined)
    uni.showToast({ title: '绑定成功', icon: 'success' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '手机号绑定失败，请稍后重试。')
  } finally {
    bindingWechatPhone.value = false
  }
}

async function regenerate(item: Invitation) {
  actingId.value = item.invitationId
  actingType.value = 'regenerate'
  operationError.value = ''
  try {
    await requestSubscribeMessages(['member', 'memberChange'])
    shareResult.value = await regenerateInvitation(item.invitationId)
    persistInviteShareLink(shareResult.value)
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '重新生成邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

function toggleInvitationHistory() {
  showHistory.value = !showHistory.value
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
  returnTo.value = String(options?.returnTo || '')
  guideMode.value = String(options?.guide || '') === '1'
  session.restoreSession()
  if (session.isLoggedIn && familyId.value) {
    loading.value = true
  }
})
onShow(() => {
  void loadShareConfigs()
  loadInvitations()
})
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

.invitation-tabs {
  margin-bottom: 20rpx;
}

.invitation-tabs .archive-segment-tab.has-pending {
  color: var(--archive-cinnabar);
  font-weight: 700;
}

.sent-group,
.share-panel,
.member-invite-panel,
.member-picker,
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

.invite-phone-bind {
  margin: 0 0 18rpx;
}

.invite-phone-bind :deep(.mini-notice) {
  margin-bottom: 14rpx;
}

.wechat-phone-bind-button {
  display: flex;
  min-height: 82rpx;
  align-items: center;
  justify-content: center;
  width: 100%;
  margin: 0;
  border: 1rpx solid rgba(31, 58, 95, 0.22);
  border-radius: var(--tree-radius-md, 16rpx);
  background: rgba(255, 252, 245, 0.92);
  color: var(--archive-blue);
  font-size: 27rpx;
  font-weight: 700;
  line-height: 1.35;
}

.wechat-phone-bind-button::after {
  border: 0;
}

.wechat-phone-bind-button[disabled] {
  opacity: 0.58;
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

.sent-item.reopenable {
  cursor: pointer;
}

.sent-item.reopenable:active {
  background: rgba(22, 51, 83, 0.05);
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
  font-size: 22rpx;
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

.sent-meta > text {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  word-break: break-all;
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

.share-hero {
  display: flex;
  align-items: center;
  gap: 20rpx;
  margin: -4rpx 0 24rpx;
  padding-bottom: 22rpx;
  border-bottom: 1rpx solid var(--archive-line);
}

.share-hero-seal {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 76rpx;
  height: 76rpx;
  border: 2rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 36rpx;
  font-weight: 700;
  transform: rotate(-5deg);
}

.share-hero-copy {
  min-width: 0;
}

.share-hero-kicker {
  display: block;
  color: var(--archive-cinnabar);
  font-size: 18rpx;
  font-weight: 650;
  letter-spacing: 3rpx;
  line-height: 1.3;
}

.share-hero-title {
  display: block;
  margin-top: 5rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 34rpx;
  font-weight: 700;
  line-height: 1.35;
}

.share-hero-desc {
  display: block;
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.5;
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
