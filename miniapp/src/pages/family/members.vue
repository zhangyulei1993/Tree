<template>
  <view class="tree-page">
    <MiniBackHome />
    <FamilyContextHeader
      v-if="family"
      :family-name="family.familyName"
      section="成员名册与账号绑定"
      subtitle="查看成员档案、账号绑定和邀请状态"
      :role-label="roleText(family.role)"
      @back="openFamilyOverview"
    />
    <view v-else class="tree-tool-banner members-banner">
      <view class="banner-copy">
        <text class="tree-tool-banner-title">成员名册与账号绑定</text>
        <text class="tree-tool-banner-desc">查看成员基本信息与账号绑定</text>
      </view>
      <view class="tree-pedigree-mark" aria-hidden="true">
        <view class="node node-root" />
        <view class="line-v" />
        <view class="line-l" />
        <view class="line-r" />
        <view class="node node-branch node-left" />
        <view class="node node-branch node-right" />
        <view class="trunk" />
      </view>
    </view>

    <MiniCard v-if="errorMessage && members.length === 0">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadMembers">重新加载</MiniButton>
    </MiniCard>

    <template v-else>
      <MiniCard v-if="loading && members.length === 0">
        <MiniEmptyState symbol="…" title="正在加载" description="正在加载成员..." />
      </MiniCard>

      <MiniCard v-else-if="members.length === 0">
        <MiniEmptyState
          symbol="员"
          title="暂无成员"
          description="当前家庭还没有可展示的成员。"
        />
      </MiniCard>

      <view v-else class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">成员名册</text>
          <text class="tree-space-subtitle">共 {{ members.length }} 位成员</text>
        </view>
        <view class="tree-space-body section-pad">
          <view v-for="member in members" :key="member.memberId" class="member-entry">
            <view
              class="member-node"
              :class="{
                actionable: hasNodeAction(member),
                selected: isActionMember(member)
              }"
              @tap="goEditMember(member)"
            >
              <MemberMiniCard
                :name="member.name"
                :age-text="formatMemberAge(member)"
                :gender-text="formatMemberGender(member.gender)"
                :living-text="formatMemberLiving(member.isAlive)"
                :binding-text="formatMemberBindingNeed(member)"
                :show-binding="true"
                :status-label="member.status !== 'ACTIVE' ? memberStatusText(member.status) : undefined"
                :status-tone="memberStatusTone(member.status)"
                :role-label="roleText(member.boundFamilyRole)"
              />
              <view v-if="canEditMember(member)" class="member-edit-cue" @tap.stop="goEditMember(member)">
                <text>编辑</text>
                <text class="cue-arrow">›</text>
              </view>
              <view
                v-if="hasNodeAction(member)"
                class="inline-action-bar"
                :class="{ single: !hasMoreAction(member) }"
              >
                <view v-if="hasMoreAction(member)" class="inline-action more" @tap.stop="toggleActionMember(member)">
                  <text class="inline-action-main">
                    {{ inlineActionMain(member) }}
                  </text>
                  <text class="inline-action-sub">
                    {{ inlineActionSub(member) }}
                  </text>
                  <text class="inline-action-arrow">›</text>
                </view>
              </view>
            </view>
            <view
              v-if="isActionMember(member) && hasNodeAction(member)"
              class="member-actions"
            >
              <view class="member-actions-head">
                <text class="member-actions-title">本节点操作</text>
                <text class="member-actions-target">{{ member.name }} · 成员资料与账号绑定</text>
              </view>
              <MiniNotice v-if="actionError" tone="warm" title="操作失败">
                {{ actionError }}
              </MiniNotice>
              <view v-if="canManageFamily" class="node-action-grid">
                <MiniButton variant="secondary" size="sm" @click="goEditMember(member)">
                  编辑 {{ member.name }} 资料
                </MiniButton>
                <MiniButton
                  v-if="canDeleteForMember(member)"
                  variant="secondary"
                  size="sm"
                  :loading="deletingMemberId === String(member.memberId)"
                  :disabled="deletingMemberId === String(member.memberId)"
                  @click="confirmDeleteMember(member)"
                >
                  删除成员节点
                </MiniButton>
              </view>
              <MiniNotice
                v-if="memberInvitation(member)"
                :tone="displayInvitationStatus(memberInvitation(member)!) === 'EXPIRED' ? 'warm' : 'security'"
                :title="displayInvitationStatus(memberInvitation(member)!) === 'EXPIRED' ? '该节点邀请已过期' : '该节点已有待处理邀请'"
              >
                {{
                  displayInvitationStatus(memberInvitation(member)!) === 'EXPIRED'
                    ? `请前往“发出的成员邀请”重新生成 ${member.name} 的绑定链接。`
                    : `${member.name} 这个成员节点已有待处理邀请，无需重复创建。`
                }}
              </MiniNotice>
              <MiniButton
                v-if="canShowInviteForMember(member)"
                variant="secondary"
                size="sm"
                @click="openInvitePanel(member)"
              >
                邀请 {{ member.name }} 本人绑定此节点
              </MiniButton>
              <MiniButton
                v-else-if="memberInvitation(member)"
                variant="secondary"
                size="sm"
                @click="openSentInvitations"
              >
                查看 {{ member.name }} 的邀请记录
              </MiniButton>
              <MiniButton
                v-if="canUnbindForMember(member)"
                variant="secondary"
                size="sm"
                :loading="unbindingMemberId === String(member.memberId)"
                :disabled="unbindingMemberId === String(member.memberId)"
                @click="confirmUnbindMember(member)"
              >
                解除账号绑定
              </MiniButton>
            </view>
            <MiniCard
              v-if="isInvitePanelMember(member)"
              :id="`invite-panel-${member.memberId}`"
              variant="soft"
              class="invite-panel"
            >
              <view class="invite-head">
                <view>
                  <text class="invite-title">邀请 {{ member.name }} 绑定账号</text>
                  <text class="tree-muted invite-desc">邀请有效期为 7 天，对方接受后将绑定到该成员节点。</text>
                </view>
                <text class="invite-close" @tap.stop="closeInvitePanel">×</text>
              </view>

              <template v-if="!createdInvitation">
                <textarea
                  v-model.trim="inviteMessage"
                  class="tree-textarea"
                  maxlength="300"
                  placeholder="邀请说明（可选）"
                />
                <text v-if="inviteError" class="tree-field-error">{{ inviteError }}</text>
                <MiniButton
                  :loading="inviteSubmitting"
                  :disabled="inviteSubmitting"
                  @click="createShareInvitation"
                >
                  生成邀请
                </MiniButton>
              </template>

              <template v-else>
                <MiniNotice tone="security" title="邀请已生成">
                  请立即发送给本人。邀请链接只在当前页面保留，不会写入本地存储。
                </MiniNotice>
                <!-- #ifdef MP-WEIXIN -->
                <button class="wechat-share-button" open-type="share">发送给微信好友</button>
                <!-- #endif -->
                <MiniButton variant="secondary" @click="copyInviteLink">复制邀请链接</MiniButton>
                <MiniButton
                  variant="secondary"
                  :loading="inviteCancelling"
                  :disabled="inviteCancelling"
                  @click="confirmCancelInvitation"
                >
                  取消本次邀请
                </MiniButton>
                <text v-if="inviteError" class="tree-field-error">{{ inviteError }}</text>
              </template>
            </MiniCard>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onHide, onLoad, onShareAppMessage, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, nextTick, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import {
  cancelInvitation,
  createInvitation,
  listFamilyInvitations
} from '@/api/invitations'
import { deleteFamilyMember, listFamilyMembers, unbindFamilyMemberUser } from '@/api/members'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { statusTagTone } from '@/components/base/formatStatus'
import FamilyContextHeader from '@/components/family/FamilyContextHeader.vue'
import MemberMiniCard from '@/components/family/MemberMiniCard.vue'
import { useSessionStore } from '@/stores/session'
import type {
  CreatedInvitation,
  FamilyDetail,
  FamilyMember,
  Invitation
} from '@/types/api'
import {
  canDeleteMember,
  canEditMember as canEditMemberAction,
  canShowInviteButton,
  canUnbindMember,
  hasMoreAction as hasMoreMemberAction,
  hasNodeAction as hasMemberNodeAction,
  invitationHeadline,
  memberActionSummary
} from '@/features/family/memberListActions'
import {
  formatMemberAge,
  formatMemberBindingNeed,
  formatMemberGender,
  formatMemberLiving
} from '@/utils/memberFormat'

const session = useSessionStore()
session.restoreSession()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const invitations = ref<Invitation[]>([])
const loading = ref(session.isLoggedIn)
const errorMessage = ref('')
const selectedActionMemberId = ref('')
const selectedInviteMember = ref<FamilyMember | null>(null)
const createdInvitation = ref<CreatedInvitation | null>(null)
const inviteMessage = ref('')
const inviteError = ref('')
const inviteSubmitting = ref(false)
const inviteCancelling = ref(false)
const actionError = ref('')
const deletingMemberId = ref('')
const unbindingMemberId = ref('')

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

function resetTransientUI() {
  closeInvitePanel()
  selectedActionMemberId.value = ''
  actionError.value = ''
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  family.value = null
  members.value = []
  invitations.value = []
  resetTransientUI()
}

async function loadMembers() {
  session.restoreSession()
  if (!familyId.value) {
    loading.value = false
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = members.value.length === 0
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    const familyResult = await getFamilyDetail(familyId.value)
    const canManage = familyResult.role === 'FOUNDER' || familyResult.role === 'FAMILY_ADMIN'
    const [memberResult, invitationResult] = await Promise.all([
      listFamilyMembers(familyId.value),
      canManage ? listFamilyInvitations(familyId.value) : Promise.resolve([])
    ])
    family.value = familyResult
    invitations.value = invitationResult
    members.value = sortMembersForBinding(memberResult)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '成员列表加载失败。')
  } finally {
    loading.value = false
  }
}

function memberActionContext(member: FamilyMember) {
  const invitation = memberInvitation(member)
  return {
    canManageFamily: canManageFamily.value,
    invitation,
    invitationExpired: invitation ? displayInvitationStatus(invitation) === 'EXPIRED' : false
  }
}

function canEditMember(member: FamilyMember) {
  return canEditMemberAction(member, canManageFamily.value)
}

function canUnbindForMember(member: FamilyMember) {
  return canUnbindMember(member, canManageFamily.value)
}

function canDeleteForMember(member: FamilyMember) {
  return canDeleteMember(member, canManageFamily.value)
}

function canShowInviteForMember(member: FamilyMember) {
  return canShowInviteButton(member, memberActionContext(member))
}

function hasBindingAction(member: FamilyMember) {
  const ctx = memberActionContext(member)
  return canShowInviteButton(member, ctx) || Boolean(ctx.invitation)
}

function hasNodeAction(member: FamilyMember) {
  return hasMemberNodeAction(member, memberActionContext(member))
}

function hasMoreAction(member: FamilyMember) {
  return hasMoreMemberAction(member, memberActionContext(member))
}

function inlineActionMain(member: FamilyMember) {
  const ctx = memberActionContext(member)
  const headline = invitationHeadline(ctx)
  if (headline) return headline
  if (isActionMember(member)) return '收起本节点操作'
  return '更多节点操作'
}

function inlineActionSub(member: FamilyMember) {
  const ctx = memberActionContext(member)
  const summary = memberActionSummary(member, ctx)
  if (summary) return summary
  return `${member.name} 暂无可用管理操作`
}

function isExpiredInvitation(item: Invitation) {
  return item.status === 'PENDING' && new Date(item.expiredAt).getTime() <= Date.now()
}

function displayInvitationStatus(item: Invitation) {
  return isExpiredInvitation(item) ? 'EXPIRED' : item.status
}

function sortMembersForBinding(list: FamilyMember[]) {
  return [...list].sort((left, right) => {
    const leftRank = hasBindingAction(left) ? 0 : 1
    const rightRank = hasBindingAction(right) ? 0 : 1
    return leftRank - rightRank
  })
}

function isActionMember(member: FamilyMember) {
  return String(selectedActionMemberId.value) === String(member.memberId)
}

function toggleActionMember(member: FamilyMember) {
  if (!hasNodeAction(member)) return
  const memberId = String(member.memberId)
  if (selectedActionMemberId.value === memberId) {
    closeInvitePanel()
    selectedActionMemberId.value = ''
    actionError.value = ''
    return
  }
  closeInvitePanel()
  selectedActionMemberId.value = memberId
  actionError.value = ''
}

function memberInvitation(member: FamilyMember) {
  return invitations.value.find((item) =>
    String(item.targetMemberId) === String(member.memberId)
    && (displayInvitationStatus(item) === 'PENDING' || displayInvitationStatus(item) === 'EXPIRED')
  )
}

function invitationSummary(item: Invitation) {
  return displayInvitationStatus(item) === 'EXPIRED' ? '邀请已过期' : '邀请待接受'
}

function confirmUnbindMember(member: FamilyMember) {
  uni.showModal({
    title: '解除账号绑定',
    content: `确定解除「${member.name}」的账号绑定吗？解除后该账号将不能再以此节点身份进入当前家庭；成员节点与家谱关系不会删除。`,
    success: (result) => {
      if (result.confirm) unbindMemberAccount(member)
    }
  })
}

async function unbindMemberAccount(member: FamilyMember) {
  unbindingMemberId.value = String(member.memberId)
  actionError.value = ''
  try {
    await unbindFamilyMemberUser(familyId.value, member.memberId, '小程序成员列表解除绑定')
    uni.showToast({ title: '已解除绑定', icon: 'success' })
    selectedActionMemberId.value = ''
    await loadMembers()
  } catch (error) {
    actionError.value = apiErrorMessage(error, '解除绑定失败。')
  } finally {
    unbindingMemberId.value = ''
  }
}

function openSentInvitations() {
  uni.navigateTo({ url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}` })
}

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function goEditMember(member: FamilyMember) {
  if (!canEditMember(member)) return
  uni.navigateTo({
    url: `/pages/family/member-edit?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(String(member.memberId))}`
  })
}

function openInvitePanel(member: FamilyMember) {
  selectedActionMemberId.value = String(member.memberId)
  selectedInviteMember.value = member
  createdInvitation.value = null
  inviteMessage.value = ''
  inviteError.value = ''
  nextTick(() => {
    uni.pageScrollTo({
      selector: `#invite-panel-${member.memberId}`,
      duration: 250
    })
  })
}

function isInvitePanelMember(member: FamilyMember) {
  return String(selectedInviteMember.value?.memberId || '') === String(member.memberId)
}

function closeInvitePanel() {
  selectedInviteMember.value = null
  createdInvitation.value = null
  inviteMessage.value = ''
  inviteError.value = ''
}

function confirmDeleteMember(member: FamilyMember) {
  uni.showModal({
    title: '删除成员节点',
    content: `确定删除“${member.name}”吗？有亲属关系、创建者或管理员身份时，系统会拒绝删除。`,
    success: (result) => {
      if (result.confirm) removeMember(member)
    }
  })
}

async function removeMember(member: FamilyMember) {
  deletingMemberId.value = String(member.memberId)
  actionError.value = ''
  try {
    await deleteFamilyMember(familyId.value, member.memberId, '小程序成员列表删除')
    uni.showToast({ title: '成员已删除', icon: 'success' })
    selectedActionMemberId.value = ''
    await loadMembers()
  } catch (error) {
    actionError.value = apiErrorMessage(error, '删除成员失败。')
  } finally {
    deletingMemberId.value = ''
  }
}

async function createShareInvitation() {
  const member = selectedInviteMember.value
  if (!member || inviteSubmitting.value) return
  inviteSubmitting.value = true
  inviteError.value = ''
  try {
    const result = await createInvitation(familyId.value, member.memberId, {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: inviteMessage.value || undefined,
      familyRoleAfterAccept: 'MEMBER'
    })
    createdInvitation.value = result
    invitations.value = [
      result.invitation,
      ...invitations.value.filter(
        (item) => String(item.invitationId) !== String(result.invitation.invitationId)
      )
    ]
  } catch (error) {
    inviteError.value = apiErrorMessage(error, '创建邀请失败。')
  } finally {
    inviteSubmitting.value = false
  }
}

function invitePath() {
  const token = createdInvitation.value?.inviteToken
  return token ? `/pages/invite/detail?inviteToken=${encodeURIComponent(token)}` : '/pages/home/index'
}

function copyInviteLink() {
  const token = createdInvitation.value?.inviteToken
  if (!token) return
  uni.setClipboardData({
    data: `https://tree.bigbigboy.cn/invite/${encodeURIComponent(token)}`,
    success: () => uni.showToast({ title: '邀请链接已复制', icon: 'success' })
  })
}

function confirmCancelInvitation() {
  if (!createdInvitation.value) return
  uni.showModal({
    title: '取消邀请',
    content: '取消后，已经发送的邀请链接将无法继续使用。',
    success: (result) => {
      if (result.confirm) cancelCurrentInvitation()
    }
  })
}

async function cancelCurrentInvitation() {
  const invitationId = createdInvitation.value?.invitation.invitationId
  if (!invitationId || inviteCancelling.value) return
  inviteCancelling.value = true
  inviteError.value = ''
  try {
    await cancelInvitation(invitationId, { reason: '小程序取消分享邀请' })
    closeInvitePanel()
    uni.showToast({ title: '邀请已取消', icon: 'success' })
    await loadMembers()
  } catch (error) {
    inviteError.value = apiErrorMessage(error, '取消邀请失败。')
  } finally {
    inviteCancelling.value = false
  }
}

function memberStatusText(status: string) {
  switch (status) {
    case 'ACTIVE':
      return '正常'
    case 'DELETED':
      return '已删除'
    case 'DISABLED':
      return '已停用'
    case 'PENDING':
      return '待审核'
    case 'APPROVED':
      return '已通过'
    case 'REJECTED':
      return '已拒绝'
    default:
      return '未知状态'
  }
}

function memberStatusTone(status: string) {
  return statusTagTone(status)
}

function roleText(role?: string | null) {
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    default:
      return role ? '未知角色' : ''
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  session.restoreSession()
  if (session.isLoggedIn && familyId.value) {
    loading.value = true
  }
})
onShareAppMessage(() => ({
  title: selectedInviteMember.value
    ? `${family.value?.familyName || '家庭'} 邀请你确认「${selectedInviteMember.value.name}」身份并加入家谱`
    : 'Tree 家脉亲缘',
  path: invitePath(),
  imageUrl: '/static/share/family-invitation.jpg'
}))
onShow(loadMembers)
onHide(resetTransientUI)
onUnload(resetPageData)
</script>

<style scoped>
.members-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 24rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.18);
  border-radius: 36rpx;
  background:
    radial-gradient(circle at 90% 12%, rgba(216, 175, 104, 0.22), transparent 220rpx),
    linear-gradient(135deg, #17304c 0%, #245653 100%);
  padding: 32rpx;
  color: #fff;
  box-shadow: 0 24rpx 60rpx rgba(24, 54, 83, 0.18);
}

.banner-copy {
  flex: 1;
  min-width: 0;
}

.section-pad {
  padding: 8rpx 24rpx 16rpx;
}

.tree-page :deep(.mini-card.soft) {
  margin-bottom: 16rpx;
}

.member-entry {
  margin-bottom: 18rpx;
}

.member-node {
  position: relative;
  border-radius: 26rpx;
  transition: all 0.2s ease;
}

.member-node.actionable {
  position: relative;
  border: 2rpx solid rgba(184, 149, 90, 0.34);
  background:
    radial-gradient(circle at 100% 0%, rgba(216, 175, 104, 0.16), transparent 150rpx),
    linear-gradient(180deg, rgba(255, 250, 240, 0.92) 0%, rgba(255, 255, 255, 0.98) 100%);
  box-shadow: 0 14rpx 34rpx rgba(184, 149, 90, 0.10);
}

.member-node.selected {
  border-color: rgba(47, 107, 87, 0.42);
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.14), transparent 160rpx),
    linear-gradient(180deg, rgba(243, 250, 247, 0.98) 0%, #ffffff 100%);
  box-shadow: 0 16rpx 42rpx rgba(47, 107, 87, 0.12);
}

.member-node :deep(.member-mini-card) {
  margin-bottom: 0;
}

.member-node.actionable :deep(.member-mini-card) {
  border-color: transparent;
  background: transparent;
  box-shadow: none;
  padding-right: 92rpx;
}

.member-edit-cue {
  position: absolute;
  top: 18rpx;
  right: 18rpx;
  z-index: 2;
  display: inline-flex;
  align-items: center;
  gap: 4rpx;
  min-height: 44rpx;
  border: 1rpx solid rgba(47, 107, 87, 0.18);
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.92);
  color: var(--tree-green, #2f6b57);
  padding: 0 14rpx;
  font-size: 22rpx;
  font-weight: 700;
  box-shadow: 0 8rpx 20rpx rgba(24, 54, 83, 0.07);
}

.member-edit-cue:active {
  transform: translateY(1rpx);
  background: var(--tree-green-light, #e8f3ee);
}

.cue-arrow {
  font-size: 24rpx;
  line-height: 1;
}

.inline-action-bar {
  display: block;
  padding: 0 18rpx 12rpx;
}

.inline-action-bar.single {
  grid-template-columns: 1fr;
}

.inline-action {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 2rpx;
  box-sizing: border-box;
  min-height: 62rpx;
  border: 1rpx solid rgba(184, 149, 90, 0.18);
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.66);
  padding: 11rpx 54rpx 11rpx 16rpx;
}

.inline-action.more {
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.78) 0%, rgba(250, 246, 238, 0.7) 100%);
}

.inline-action-main {
  color: var(--tree-text-primary, #1e293b);
  font-size: 24rpx;
  font-weight: 800;
  line-height: 1.25;
}

.inline-action-sub {
  color: var(--tree-text-secondary, #64748b);
  font-size: 19rpx;
  line-height: 1.25;
}

.inline-action-arrow {
  position: absolute;
  right: 18rpx;
  top: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34rpx;
  height: 34rpx;
  border-radius: 999rpx;
  background: rgba(184, 149, 90, 0.11);
  color: var(--tree-gold-text, #8a6d2f);
  font-size: 26rpx;
  font-weight: 800;
  transform: translateY(-50%);
}

.member-actions {
  margin: 12rpx 0 14rpx;
  padding: 20rpx 22rpx 22rpx;
  border: 1rpx solid rgba(47, 107, 87, 0.16);
  border-radius: 24rpx;
  background:
    linear-gradient(135deg, rgba(243, 250, 247, 0.98) 0%, rgba(255, 255, 255, 0.96) 100%);
  box-shadow: 0 12rpx 30rpx rgba(24, 54, 83, 0.06);
}

.member-actions-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12rpx;
  margin-bottom: 14rpx;
}

.member-actions-title {
  color: var(--tree-text-primary, #1e293b);
  font-size: 24rpx;
  font-weight: 700;
  line-height: 1.4;
}

.member-actions-target {
  color: var(--tree-text-secondary, #64748b);
  font-size: 22rpx;
  line-height: 1.4;
  text-align: right;
}

.member-actions :deep(.mini-notice) {
  margin-bottom: 14rpx;
}

.node-action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12rpx;
  margin-bottom: 16rpx;
}

.invite-panel {
  margin-bottom: 20rpx;
}

.invite-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.invite-title,
.invite-desc {
  display: block;
}

.invite-title {
  color: var(--tree-text);
  font-size: 28rpx;
  font-weight: 700;
}

.invite-desc {
  margin-top: 8rpx;
}

.invite-close {
  padding: 0 8rpx;
  color: var(--tree-text-secondary);
  font-size: 40rpx;
  line-height: 1;
}

.wechat-share-button {
  margin: 16rpx 0 12rpx;
  border: 0;
  border-radius: 16rpx;
  background: var(--tree-green, #2f6b57);
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
}

.wechat-share-button::after {
  border: 0;
}
</style>
