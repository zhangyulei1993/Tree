<template>
  <view class="archive-page members-page">
    <MiniBackHome />

    <view class="members-head">
      <view>
        <text class="archive-kicker">Family Roster</text>
        <text class="archive-title">成员名册</text>
        <text class="archive-subtitle">
          {{ family?.familyName || '查看成员档案、账号绑定和邀请状态' }}
        </text>
      </view>
      <view v-if="family" class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
    </view>

    <view class="members-tools">
      <view class="members-search">
        <text class="search-mark">⌕</text>
        <input
          v-model.trim="searchKeyword"
          class="members-search-input"
          confirm-type="search"
          placeholder="搜索成员姓名"
          placeholder-class="members-search-placeholder"
        />
      </view>
      <view class="members-filter archive-paper-tag">
        <text>筛选</text>
      </view>
    </view>

    <view v-if="family" class="members-context">
      <text class="archive-chip">{{ roleText(family.role) || '成员' }}</text>
      <text class="archive-chip">共 {{ filteredMembers.length }} 位</text>
    </view>

    <view v-if="errorMessage && members.length === 0" class="members-error archive-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadMembers">重新加载</MiniButton>
    </view>

    <template v-else>
      <view v-if="loading && members.length === 0" class="members-empty archive-panel">
        <MiniEmptyState symbol="…" title="正在加载" description="正在加载成员..." />
      </view>

      <view v-else-if="filteredMembers.length === 0" class="members-empty archive-panel">
        <MiniEmptyState
          symbol="员"
          title="暂无成员"
          description="当前条件下没有可展示的成员。"
        />
      </view>

      <view v-else class="members-roster">
        <view v-for="group in memberGroups" :key="group.key" class="member-generation">
          <view class="generation-divider">
            <text>{{ group.title }}</text>
            <view class="generation-line" />
          </view>

          <view
            v-for="member in group.items"
            :key="member.memberId"
            class="member-entry"
            :class="{ selected: isActionMember(member), actionable: hasNodeAction(member) }"
            @tap="handleMemberRowTap(member)"
          >
            <view class="member-row">
              <text class="archive-surname-stamp">{{ surnameLetter(member) }}</text>
              <view class="member-main">
                <view class="member-title-line">
                  <text class="member-name">{{ member.name }}</text>
                  <text v-if="member.status !== 'ACTIVE'" class="member-state">{{ memberStatusText(member.status) }}</text>
                </view>
                <view class="member-meta-line">
                  <text>{{ memberYearText(member) }}</text>
                  <text>{{ memberGenerationLabel(member) }}</text>
                  <text>{{ formatMemberGender(member.gender) }}</text>
                </view>
              </view>
              <view class="member-side">
                <text class="member-role">{{ roleText(member.boundFamilyRole) || formatMemberBindingNeed(member) }}</text>
                <text v-if="memberInvitation(member)" class="member-invite">{{ invitationSummary(memberInvitation(member)!) }}</text>
              </view>
              <text class="member-more" @tap.stop="handleMemberMoreTap(member)">
                {{ isActionMember(member) ? '收' : '⋯' }}
              </text>
            </view>

            <view v-if="isActionMember(member) && hasNodeAction(member)" class="member-actions">
              <view class="member-actions-head">
                <text class="member-actions-title">{{ inlineActionMain(member) }}</text>
                <text class="member-actions-target">{{ inlineActionSub(member) }}</text>
              </view>
              <MiniNotice v-if="actionError" tone="warm" title="操作失败">
                {{ actionError }}
              </MiniNotice>
              <view v-if="canManageFamily" class="node-action-grid">
                <MiniButton variant="secondary" size="sm" @click="goEditMember(member)">
                  编辑资料
                </MiniButton>
                <MiniButton
                  v-if="canDeleteForMember(member)"
                  variant="secondary"
                  size="sm"
                  :loading="deletingMemberId === String(member.memberId)"
                  :disabled="deletingMemberId === String(member.memberId)"
                  @click="confirmDeleteMember(member)"
                >
                  删除节点
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
                邀请本人绑定此节点
              </MiniButton>
              <MiniButton
                v-else-if="memberInvitation(member)"
                variant="secondary"
                size="sm"
                @click="openSentInvitations"
              >
                查看邀请记录
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

            <view
              v-if="isInvitePanelMember(member)"
              :id="`invite-panel-${member.memberId}`"
              class="invite-panel archive-form-panel"
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
            </view>
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
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type {
  CreatedInvitation,
  FamilyDetail,
  FamilyMember,
  Invitation
} from '@/types/api'
import { optionalText, validateTextFields } from '@/utils/inputValidation'
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
  formatMemberBindingNeed,
  formatMemberGender
} from '@/utils/memberFormat'

type MemberGroup = {
  key: string
  title: string
  items: FamilyMember[]
}

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
const searchKeyword = ref('')

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

const filteredMembers = computed(() => {
  const keyword = searchKeyword.value.trim()
  if (!keyword) return members.value
  return members.value.filter((member) => member.name.includes(keyword))
})

const memberGroups = computed<MemberGroup[]>(() => {
  const groups = new Map<string, MemberGroup>()
  for (const member of filteredMembers.value) {
    const title = memberGroupTitle(member)
    const key = title
    if (!groups.has(key)) {
      groups.set(key, { key, title, items: [] })
    }
    groups.get(key)!.items.push(member)
  }
  return Array.from(groups.values())
})

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
  return hasMoreAction(member) ? `${member.name} 可继续维护资料与账号绑定` : `${member.name} 暂无可用管理操作`
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

function handleMemberRowTap(member: FamilyMember) {
  if (hasNodeAction(member)) {
    toggleActionMember(member)
    return
  }
  goEditMember(member)
}

function handleMemberMoreTap(member: FamilyMember) {
  if (hasNodeAction(member)) {
    toggleActionMember(member)
    return
  }
  goEditMember(member)
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
    content: `确定删除“${member.name}”吗？父母或配偶关系会一并解除；已有子女、创建者或管理员身份时系统会拒绝。`,
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
    uni.showModal({
      title: '删除失败',
      content: actionError.value,
      showCancel: false
    })
  } finally {
    deletingMemberId.value = ''
  }
}

async function createShareInvitation() {
  const member = selectedInviteMember.value
  if (!member || inviteSubmitting.value) return
  const validationMessage = validateTextFields([
    { value: inviteMessage.value, label: '邀请说明', kind: 'multiLine', maxLength: 300 }
  ])
  if (validationMessage) {
    inviteError.value = validationMessage
    return
  }
  inviteSubmitting.value = true
  inviteError.value = ''
  try {
    const result = await createInvitation(familyId.value, member.memberId, {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: optionalText(inviteMessage.value),
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

function surnameLetter(member: FamilyMember) {
  return member.name.trim().slice(0, 1) || '氏'
}

function memberYearText(member: FamilyMember) {
  if (member.birthYear && member.deathYear) return `${member.birthYear}-${member.deathYear}`
  if (member.birthYear) return `${member.birthYear} 年生`
  if (member.deathYear) return `${member.deathYear} 年殁`
  return '年份未录'
}

function memberGroupTitle(member: FamilyMember) {
  if (!member.birthYear) return '世系待补'
  const decade = Math.floor(member.birthYear / 10) * 10
  return `${decade} 年代`
}

function memberGenerationLabel(member: FamilyMember) {
  if (member.birthYear) {
    const decade = Math.floor(member.birthYear / 10) * 10
    return `${decade}代`
  }
  return '世代待补'
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
.members-page {
  padding-top: 28rpx;
}

.members-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.members-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24rpx;
  margin-bottom: 26rpx;
}

.members-tools {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin-bottom: 18rpx;
}

.members-search {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  height: 72rpx;
  border: 1rpx solid var(--archive-line-strong);
  background: rgba(255, 249, 236, 0.42);
  padding: 0 20rpx;
  box-sizing: border-box;
}

.search-mark {
  flex-shrink: 0;
  color: var(--archive-cinnabar);
  font-size: 26rpx;
  margin-right: 12rpx;
}

.members-search-input {
  flex: 1;
  min-width: 0;
  color: var(--archive-ink);
  font-size: 25rpx;
}

.members-search-placeholder {
  color: rgba(101, 112, 128, 0.58);
}

.members-filter {
  width: 76rpx;
  height: 72rpx;
  flex-shrink: 0;
}

.members-context {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 20rpx;
}

.members-error,
.members-empty {
  padding: 28rpx 0;
}

.members-error :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.member-generation {
  margin-bottom: 24rpx;
}

.generation-divider {
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin-bottom: 10rpx;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.generation-line {
  flex: 1;
  height: 1rpx;
  background: var(--archive-line-strong);
}

.member-entry {
  border-bottom: 1rpx solid var(--archive-line);
}

.member-entry.selected {
  background: rgba(255, 248, 234, 0.48);
}

.member-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-height: 112rpx;
  padding: 18rpx 0;
}

.member-main {
  flex: 1;
  min-width: 0;
}

.member-title-line {
  display: flex;
  align-items: center;
  gap: 10rpx;
  min-width: 0;
}

.member-name {
  overflow: hidden;
  color: var(--archive-ink);
  font-size: 30rpx;
  font-weight: 750;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-state {
  flex-shrink: 0;
  color: var(--archive-cinnabar);
  font-size: 20rpx;
}

.member-meta-line {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx 16rpx;
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.35;
}

.member-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  max-width: 150rpx;
  flex-shrink: 0;
}

.member-role {
  color: var(--archive-blue);
  font-size: 22rpx;
  font-weight: 700;
  line-height: 1.35;
  text-align: right;
}

.member-invite {
  margin-top: 5rpx;
  color: var(--archive-cinnabar);
  font-size: 19rpx;
  line-height: 1.3;
  text-align: right;
}

.member-more {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48rpx;
  height: 48rpx;
  border: 1rpx solid var(--archive-line);
  color: var(--archive-cinnabar);
  font-size: 25rpx;
  line-height: 1;
}

.member-actions {
  margin: 0 0 18rpx 72rpx;
  border-left: 3rpx solid var(--archive-cinnabar);
  background: rgba(255, 249, 236, 0.44);
  padding: 18rpx 0 18rpx 20rpx;
}

.member-actions-head {
  margin-bottom: 16rpx;
}

.member-actions-title,
.member-actions-target,
.invite-title,
.invite-desc {
  display: block;
}

.member-actions-title {
  color: var(--archive-ink);
  font-size: 25rpx;
  font-weight: 750;
}

.member-actions-target {
  margin-top: 6rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.45;
}

.member-actions :deep(.mini-notice) {
  margin-bottom: 14rpx;
}

.node-action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12rpx;
  margin-bottom: 14rpx;
}

.member-actions :deep(.mini-button),
.invite-panel :deep(.mini-button) {
  border-radius: 0;
  box-shadow: none;
}

.invite-panel {
  margin: 0 0 20rpx 72rpx;
  border-radius: var(--archive-radius);
  box-shadow: none;
}

.invite-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.invite-title {
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 700;
}

.invite-desc {
  margin-top: 8rpx;
}

.invite-close {
  padding: 0 8rpx;
  color: var(--archive-ink-soft);
  font-size: 40rpx;
  line-height: 1;
}

.wechat-share-button {
  margin: 16rpx 0 12rpx;
  border: 1rpx solid var(--archive-cinnabar);
  border-radius: 0;
  background: transparent;
  color: var(--archive-cinnabar);
  font-size: 26rpx;
  font-weight: 600;
}

.wechat-share-button::after {
  border: 0;
}
</style>
