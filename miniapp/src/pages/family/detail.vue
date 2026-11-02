<template>
  <view class="archive-page detail-page">
    <MiniBackHome />
    <MiniFamilyPageSkeleton v-if="loading && !family" variant="detail" />

    <view v-else-if="errorMessage && !family" class="detail-error archive-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </view>

    <template v-else-if="family">
      <view v-if="errorMessage" class="detail-refresh-error">
        <MiniNotice tone="warm" title="刷新未完成">{{ errorMessage }}</MiniNotice>
        <MiniButton size="sm" variant="secondary" @click="loadFamily">重新加载</MiniButton>
      </view>
      <view class="detail-head">
        <view>
          <text class="archive-kicker">Family Archive</text>
          <text class="archive-title">家庭详情</text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="family-cover">
        <view class="family-cover-copy">
          <text class="cover-name">{{ family.familyName }}</text>
          <text class="cover-desc">{{ family.description || '暂未填写家庭简介。' }}</text>
        </view>
        <view class="cover-book archive-book-spine">
          <text>家</text>
          <text>庭</text>
          <text>树</text>
        </view>
      </view>

      <view class="dossier-grid">
        <view v-for="item in dossierItems" :key="item.label" class="dossier-cell">
          <text class="dossier-label">{{ item.label }}</text>
          <text class="dossier-value">{{ item.value }}</text>
        </view>
      </view>

      <view v-if="canManageFamily" class="detail-profile-action">
        <text>家庭名称、籍贯或简介需要调整？</text>
        <text class="archive-thin-button" @tap="openProfile">编辑家庭资料</text>
      </view>

      <view class="status-line">
        <text class="archive-chip">{{ familyStatusText(family.status) }}</text>
        <text class="archive-chip">{{ publicStatusText(family.publicDisplayStatus) }}</text>
      </view>

      <view v-if="exitedMember" class="member-exit-notice" @tap="openExitedMemberInTree">
        <view>
          <text class="member-exit-title">{{ exitedMember.name || '该成员' }}已退出家庭</text>
          <text class="member-exit-desc">账号绑定已解除，家庭树资料仍保留。查看该节点 ›</text>
        </view>
      </view>

      <view v-if="invitationReward?.enabled" class="invitation-progress" :class="{ urgent: invitationReward.active }" @click="openInvitationReward">
        <view class="invitation-progress-head"><view><text class="invitation-progress-title">家庭邀请进度</text><text class="invitation-progress-desc">{{ invitationReward.validInvites }} 位有效邀请 · 已解锁 +{{ invitationReward.rewardedMembers }} 位名额</text></view><text class="archive-arrow">›</text></view>
        <view v-if="invitationReward.tiers.length" class="invitation-progress-track"><view class="invitation-progress-fill" :style="{ width: `${Math.min(100, invitationReward.validInvites / invitationReward.tiers[invitationReward.tiers.length - 1].requiredInvites * 100)}%` }" /><view v-for="tier in invitationReward.tiers" :key="tier.requiredInvites" class="invitation-progress-mark" :class="{ reached: invitationReward.validInvites >= tier.requiredInvites }" :style="{ left: `${tier.requiredInvites / invitationReward.tiers[invitationReward.tiers.length - 1].requiredInvites * 100}%` }" /></view>
        <view v-if="invitationReward.tiers.length" class="invitation-progress-labels"><text v-for="tier in invitationReward.tiers" :key="tier.requiredInvites">{{ tier.requiredInvites }} 人 · +{{ tier.cumulativeRewardMembers }}</text></view>
        <text v-if="invitationReward.nextRequiredInvites" class="invitation-progress-tip">距下一档还差 {{ invitationReward.nextRequiredInvites - invitationReward.validInvites }} 人{{ invitationReward.active ? '，现在可发起助力邀请' : '' }}</text>
      </view>

      <view v-if="qualityEligibility?.eligible && canManageFamily && family.publicDisplayStatus !== 'APPROVED'" class="quality-family-invite" @click="openQualityFamilySettings">
        <view class="quality-family-invite-main">
          <text class="quality-family-invite-title">已获得优质家庭申请资格</text>
          <text class="quality-family-invite-desc">平台邀请你提交申请，是否申请由家庭创建者决定</text>
        </view>
        <text class="archive-arrow">›</text>
      </view>

      <view v-if="family.role === 'FOUNDER' && recoverableMemberCount" class="recovery-reminder archive-row" @click="openRecovery">
        <view class="archive-row-main">
          <text class="archive-row-title">最近删除可恢复</text>
          <text class="archive-row-desc">成员删除后 7 天内可恢复，账号绑定不会恢复</text>
        </view>
        <text class="affair-badge">{{ recoverableMemberCount }}</text>
        <text class="archive-arrow">›</text>
      </view>

      <view v-if="showGuideNudge" class="detail-guide-nudge archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">下一步怎么做</text>
          <text class="archive-section-subtitle">{{ guideNudgeDescription }}</text>
        </view>
        <MiniButton size="sm" @click="openGuideNextAction">{{ guideNudgeAction }}</MiniButton>
        <text class="detail-guide-more" @click="openGuideHome">查看完整新手引导 ›</text>
      </view>

      <view v-if="showTreeStarter" class="tree-starter archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">开始搭建家庭树</text>
          <text class="archive-section-subtitle">先添加父母、配偶、子女或兄弟姐妹。</text>
        </view>
        <view class="tree-starter-actions">
          <MiniButton size="sm" @click="openAddFirstParent">添加父母</MiniButton>
          <MiniButton size="sm" variant="secondary" @click="openAddFirstRelative('ADD_SPOUSE')">添加配偶</MiniButton>
          <MiniButton size="sm" variant="secondary" @click="openAddFirstRelative('ADD_CHILD')">添加子女</MiniButton>
        </view>
      </view>

      <view v-if="family.status === 'DISSOLUTION_COOLDOWN'" class="cooldown-panel archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">恢复冷静期</text>
          <text class="archive-section-subtitle">
            {{ cooldownUntilText }}
          </text>
        </view>
        <MiniNotice tone="warm" title="冷静期家庭仍保留">
          冷静期内家庭可查看但暂不可编辑。创建者可在此直接恢复家庭，或完成解散。
        </MiniNotice>
        <view v-if="family.role === 'FOUNDER'" class="cooldown-actions">
          <MiniButton :loading="cooldownSubmitting" :disabled="cooldownSubmitting" @click="confirmRestoreFromCooldown">
            恢复家庭
          </MiniButton>
          <MiniButton
            variant="secondary"
            :loading="cooldownSubmitting"
            :disabled="cooldownSubmitting"
            @click="confirmFinalizeDissolution"
          >
            跳过冷静期并完成解散
          </MiniButton>
        </view>
        <MiniNotice v-else tone="security" title="只读状态">
          只有家庭创建者可以恢复家庭或完成解散。
        </MiniNotice>
      </view>

      <view v-else-if="family.status === 'DISSOLUTION_PENDING'" class="cooldown-panel archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">解散申请待审核</text>
          <text class="archive-section-subtitle">后台审核前可以撤销申请</text>
        </view>
        <MiniNotice tone="warm" title="申请处理中">
          {{ currentDissolution?.requestReason || '家庭解散申请正在等待后台审核。' }}
        </MiniNotice>
        <view v-if="family.role === 'FOUNDER'" class="cooldown-actions">
          <MiniButton
            variant="secondary"
            :loading="dissolutionSubmitting"
            :disabled="dissolutionSubmitting || !currentDissolution"
            @click="confirmCancelDissolution"
          >
            撤销解散申请
          </MiniButton>
        </view>
      </view>

      <view class="detail-directory archive-list">
        <view class="archive-row" @click="openFamilyTree">
          <view class="archive-row-main">
            <text class="archive-row-title">家庭树</text>
            <text class="archive-row-desc">{{ canManageFamily ? '查看成员、编辑关系和邀请本人绑定' : '查看成员与亲属关系' }}</text>
          </view>
          <text class="archive-row-meta">树 / 邀请</text>
          <text class="archive-arrow">›</text>
        </view>
        <view v-if="canManageFamily" class="archive-row" @click="openManageCenter">
          <view class="archive-row-main">
            <text class="archive-row-title">家庭管理</text>
            <text class="archive-row-desc">集中处理待办、审核与权限设置</text>
          </view>
          <text v-if="pendingAffairCount" class="affair-badge" @tap.stop="openNextPendingAffair">{{ pendingAffairLabel }}</text>
          <text v-else class="archive-row-meta">事务</text>
          <text class="archive-arrow">›</text>
        </view>
        <view v-if="canManageFamily" class="archive-row" @click="openFamilyHealth">
          <view class="archive-row-main">
            <text class="archive-row-title">家庭健康与档案导出</text>
            <text class="archive-row-desc">先补齐资料和关系，再生成家谱长卷</text>
          </view>
          <text v-if="pendingHealthCount" class="affair-badge health-badge" @tap.stop="openFamilyHealth">{{ pendingHealthLabel }}</text>
          <text v-else class="archive-row-meta">检查</text>
          <text class="archive-arrow">›</text>
        </view>
        <view v-if="family.role === 'FOUNDER'" class="archive-row" @click="openCapacity">
          <view class="archive-row-main">
            <text class="archive-row-title">成员名额与订单</text>
            <text class="archive-row-desc">购买扩容、继续支付或申请退款</text>
          </view>
          <text class="archive-row-meta">名额</text>
          <text class="archive-arrow">›</text>
        </view>
        <view v-if="canLeaveCurrentFamily" class="archive-row detail-leave-row" @click="confirmLeave">
          <view class="archive-row-main">
            <text class="archive-row-title">退出当前家庭</text>
            <text class="archive-row-desc">仅解除本人绑定</text>
          </view>
          <text class="archive-row-meta">退出</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { formatFamilyRegionText } from '@/features/native-place/geoHierarchy'
import { summarizeFamilyHealth } from '@/features/family-tree/familyHealthSummary'
import { cancelDissolution, finalizeDissolution, getCurrentDissolution, restoreDissolution } from '@/api/dissolutions'
import { getFamilyDetail, leaveFamily } from '@/api/families'
import { getCurrentFounderTransfer } from '@/api/founderTransfers'
import { listCapacityOrders } from '@/api/capacity'
import { getInvitationReward, type InvitationRewardSummary } from '@/api/invitationReward'
import { listFamilyInvitations } from '@/api/invitations'
import { listFamilyJoinRequests } from '@/api/joinRequests'
import { listFamilyMembers, listRecoverableFamilyMembers } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import { getQualityEligibility, type QualityEligibility } from '@/api/publicApplications'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { areNewUserGuideMilestonesComplete, readNewUserGuideProgress, resumeNewUserGuide, type NewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { useSessionStore } from '@/stores/session'
import type { DissolutionRequest, FamilyDetail, FamilyMember, FamilyTreeResult, FounderTransferRequest, Invitation, JoinRequest, RelationshipAddType } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const exitedMemberId = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const loading = ref(false)
const errorMessage = ref('')
const navigating = ref(false)
const cooldownSubmitting = ref(false)
const dissolutionSubmitting = ref(false)
const leaving = ref(false)
const currentDissolution = ref<DissolutionRequest | null>(null)
const invitationReward = ref<InvitationRewardSummary>()
const qualityEligibility = ref<QualityEligibility | null>(null)
const pendingInvitations = ref<Invitation[]>([])
const pendingJoinRequests = ref<JoinRequest[]>([])
const currentTransfer = ref<FounderTransferRequest | null>(null)
const healthTree = ref<FamilyTreeResult | null>(null)
const recoverableMemberCount = ref(0)
const capacityPendingTaskCount = ref(0)
const exitedMember = computed(() => members.value.find((member) => String(member.memberId) === exitedMemberId.value) || null)
const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const canLeaveCurrentFamily = computed(() =>
  family.value?.role !== 'FOUNDER' && family.value?.status === 'NORMAL'
)
const guideNudgeMilestones = computed<NewUserGuideMilestone[]>(() =>
  canManageFamily.value
    ? ['MEMBER_GUIDE_OPENED', 'RELATION_GUIDE_OPENED', 'INVITE_GUIDE_OPENED']
    : ['TREE_GUIDE_OPENED', 'INVITE_GUIDE_OPENED']
)
const showGuideNudge = computed(() =>
  family.value?.status === 'NORMAL' &&
  !areNewUserGuideMilestonesComplete(readNewUserGuideProgress(), guideNudgeMilestones.value)
)
const guideNudgeAction = computed(() => {
  if (!canManageFamily.value) return '查看家庭树'
  return members.value.length <= 1 ? '添加第一位亲属' : '继续整理关系'
})
const guideNudgeDescription = computed(() => {
  if (!canManageFamily.value) return '先从家庭树认识自己的位置和亲属关系。'
  return members.value.length <= 1
    ? '先补充一位最熟悉的亲属，家庭树就有了第一条关系。'
    : '成员已经录入，接下来把父母、配偶或子女关系连起来。'
})
const pendingAffairCount = computed(() => {
  const invitationCount = pendingInvitations.value.filter((item) => item.status === 'PENDING' && new Date(item.expiredAt).getTime() > Date.now()).length
  const joinRequestCount = pendingJoinRequests.value.filter((item) => item.requestStatus === 'PENDING').length
  const transferCount = currentTransfer.value?.requestStatus === 'PENDING' ? 1 : 0
  const dissolutionCount = currentDissolution.value?.requestStatus === 'PENDING' ? 1 : 0
  const healthTaskCount = pendingHealthCount.value > 0 ? 1 : 0
  return invitationCount + joinRequestCount + transferCount + dissolutionCount + healthTaskCount + capacityPendingTaskCount.value
})
const pendingAffairLabel = computed(() => pendingAffairCount.value > 99 ? '99+' : String(pendingAffairCount.value))
const pendingHealthCount = computed(() => summarizeFamilyHealth(family.value, members.value, healthTree.value).total)
const pendingHealthLabel = computed(() => pendingHealthCount.value > 99 ? '99+' : String(pendingHealthCount.value))
const nextPendingAffairRoute = computed(() => {
  if (currentDissolution.value?.requestStatus === 'PENDING') return `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}&section=dissolution`
  if (currentTransfer.value?.requestStatus === 'PENDING') return `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}&section=transfer`
  if (pendingHealthCount.value > 0) return `/pages/family/health?familyId=${encodeURIComponent(familyId.value)}`
  if (capacityPendingTaskCount.value > 0) return `/pages/family/capacity-orders?familyId=${encodeURIComponent(familyId.value)}`
  const joinRequest = pendingJoinRequests.value.find((item) => item.requestStatus === 'PENDING')
  if (joinRequest) return `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}&requestId=${encodeURIComponent(String(joinRequest.requestId))}`
  const invitation = pendingInvitations.value.find((item) =>
    item.status === 'PENDING' && new Date(item.expiredAt).getTime() > Date.now()
  )
  if (invitation) {
    return `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}&mode=node&memberId=${encodeURIComponent(String(invitation.targetMemberId))}&memberName=${encodeURIComponent(invitation.targetMemberName || '')}`
  }
  return ''
})
const showTreeStarter = computed(() =>
  canManageFamily.value && family.value?.status === 'NORMAL' && members.value.length <= 1
)
const familyTreeUrl = computed(() =>
  `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
)
const lastMemberUpdatedAtText = computed(() => formatLatestMemberUpdate(members.value))
const cooldownUntilText = computed(() => {
  const value = family.value?.dissolutionCooldownUntil
  return value ? `冷静期截止 ${formatDateTime(value)}` : '冷静期中'
})

const dossierItems = computed(() => {
  const currentFamily = family.value
  if (!currentFamily) return []
  return [
    { label: '姓氏', value: `${currentFamily.familySurname || '未录'}氏` },
    { label: '成员数', value: `${currentFamily.memberCount ?? members.value.length} / ${currentFamily.memberLimit || '—'} 位` },
    { label: '地区', value: formatFamilyRegionText(currentFamily) },
    { label: '最近更新', value: lastMemberUpdatedAtText.value }
  ]
})

async function loadFamily() {
  session.restoreSession()
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = !family.value
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    const [familyDetail, familyMembers, reward] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyMembers(familyId.value),
      getInvitationReward(familyId.value).catch(() => undefined)
    ])
    family.value = familyDetail
    members.value = familyMembers
    invitationReward.value = reward
    qualityEligibility.value = (familyDetail.role === 'FOUNDER' || familyDetail.role === 'FAMILY_ADMIN')
      ? await getQualityEligibility(familyId.value).catch(() => null)
      : null
    currentDissolution.value = familyDetail.status === 'DISSOLUTION_PENDING'
      ? await getCurrentDissolution(familyId.value)
      : null
    if (familyDetail.role === 'FOUNDER' || familyDetail.role === 'FAMILY_ADMIN') {
      const [invitations, joinRequests, transfer, treeResult, recoverableMembers, capacityOrders] = await Promise.all([
        listFamilyInvitations(familyId.value),
        listFamilyJoinRequests(familyId.value),
        getCurrentFounderTransfer(familyId.value).catch(() => null),
        getPrivateTree(familyId.value).catch(() => null),
        familyDetail.role === 'FOUNDER' ? listRecoverableFamilyMembers(familyId.value).catch(() => []) : Promise.resolve([]),
        familyDetail.role === 'FOUNDER' ? listCapacityOrders(familyId.value).catch(() => []) : Promise.resolve([])
      ])
      pendingInvitations.value = invitations
      pendingJoinRequests.value = joinRequests
      currentTransfer.value = transfer
      healthTree.value = treeResult
      recoverableMemberCount.value = recoverableMembers.length
      capacityPendingTaskCount.value = capacityOrders.filter((item) =>
        item.status === 'REFUNDING' || (item.status === 'PENDING' && (!item.expiresAt || Date.parse(item.expiresAt) > Date.now()))
      ).length
    } else {
      pendingInvitations.value = []
      pendingJoinRequests.value = []
      currentTransfer.value = null
      healthTree.value = null
      recoverableMemberCount.value = 0
      capacityPendingTaskCount.value = 0
    }
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭详情加载失败。')
  } finally {
    loading.value = false
  }
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  family.value = null
  members.value = []
  currentDissolution.value = null
  invitationReward.value = undefined
  qualityEligibility.value = null
  pendingInvitations.value = []
  pendingJoinRequests.value = []
  currentTransfer.value = null
  healthTree.value = null
  recoverableMemberCount.value = 0
  capacityPendingTaskCount.value = 0
}

function openQualityFamilySettings() {
  navigateOnce(`/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}&section=public`)
}

function navigateOnce(url: string) {
  if (navigating.value) return
  navigating.value = true
  uni.navigateTo({
    url,
    complete: () => {
      navigating.value = false
    }
  })
}

function openProfile() {
  navigateOnce(`/pages/family/profile?familyId=${encodeURIComponent(familyId.value)}`)
}

function openFamilyTree() {
  navigateOnce(familyTreeUrl.value)
}

function openGuideNextAction() {
  if (!canManageFamily.value) {
    navigateOnce(`${familyTreeUrl.value}&guide=1`)
    return
  }
  const section = members.value.length <= 1 ? 'node' : 'relations'
  navigateOnce(`/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}&section=${section}&guide=1`)
}

function openGuideHome() {
  resumeNewUserGuide()
  uni.switchTab({ url: '/pages/home/index' })
}

function openExitedMemberInTree() {
  if (!exitedMember.value) return
  navigateOnce(`${familyTreeUrl.value}&memberId=${encodeURIComponent(String(exitedMember.value.memberId))}`)
}

function openManageCenter() {
  navigateOnce(`/pages/family/manage-center?familyId=${encodeURIComponent(familyId.value)}`)
}

function openFamilyHealth() {
  navigateOnce(`/pages/family/health?familyId=${encodeURIComponent(familyId.value)}`)
}

function openNextPendingAffair() {
  if (!nextPendingAffairRoute.value) {
    openManageCenter()
    return
  }
  navigateOnce(nextPendingAffairRoute.value)
}

function openCapacity() {
  navigateOnce(`/pages/family/capacity?familyId=${encodeURIComponent(familyId.value)}`)
}

function openRecovery() {
  navigateOnce(`/pages/family/recovery?familyId=${encodeURIComponent(familyId.value)}`)
}

function confirmLeave() {
  if (!family.value || leaving.value) return
  uni.showModal({
    title: '确认退出家庭',
    content: '退出后将解除你与当前成员节点的绑定，家庭树资料不会删除。再次加入需要重新申请或接受邀请。',
    confirmColor: '#B5473C',
    success: (result) => {
      if (result.confirm) void submitLeave()
    }
  })
}

async function submitLeave() {
  leaving.value = true
  try {
    await leaveFamily(familyId.value, '用户从家庭详情退出')
    uni.showToast({ title: '已退出家庭', icon: 'success' })
    setTimeout(() => uni.switchTab({ url: '/pages/family/my' }), 500)
  } catch (error) {
    uni.showModal({
      title: '退出失败',
      content: apiErrorMessage(error, '暂时无法退出该家庭。'),
      showCancel: false
    })
  } finally {
    leaving.value = false
  }
}
function openInvitationReward() { navigateOnce(`/pages/family/invitation-reward?familyId=${encodeURIComponent(familyId.value)}`) }

function openAddFirstParent() {
  uni.showActionSheet({
    itemList: ['添加父亲', '添加母亲'],
    success: (result) => {
      openAddFirstRelative(result.tapIndex === 0 ? 'ADD_FATHER' : 'ADD_MOTHER')
    }
  })
}

function openAddFirstRelative(addType: RelationshipAddType) {
  const baseMemberId = members.value[0]?.memberId
  const query = [
    `familyId=${encodeURIComponent(familyId.value)}`,
    'section=node',
    `addType=${encodeURIComponent(addType)}`,
    baseMemberId ? `baseMemberId=${encodeURIComponent(String(baseMemberId))}` : ''
  ].filter(Boolean).join('&')
  navigateOnce(`/pages/family/manage?${query}`)
}

function confirmRestoreFromCooldown() {
  if (!family.value || family.value.role !== 'FOUNDER' || cooldownSubmitting.value) return
  uni.showModal({
    title: '恢复家庭',
    content: '恢复后家庭将重新回到正常状态，并恢复到家庭列表。确认恢复吗？',
    confirmColor: '#163353',
    success: async (result) => {
      if (!result.confirm) return
      cooldownSubmitting.value = true
      try {
        family.value = await restoreDissolution(familyId.value)
        uni.showToast({ title: '家庭已恢复', icon: 'success' })
        await loadFamily()
      } catch (error) {
        errorMessage.value = apiErrorMessage(error, '恢复家庭失败，请稍后重试。')
      } finally {
        cooldownSubmitting.value = false
      }
    }
  })
}

function confirmFinalizeDissolution() {
  if (!family.value || family.value.role !== 'FOUNDER' || cooldownSubmitting.value) return
  uni.showModal({
    title: '跳过冷静期',
    content: '跳过后家庭将完成解散，并从用户端家庭列表隐藏。此后如需恢复，请联系后台处理。确认继续吗？',
    confirmColor: '#A83B2D',
    success: async (result) => {
      if (!result.confirm) return
      cooldownSubmitting.value = true
      try {
        await finalizeDissolution(familyId.value)
        uni.showToast({ title: '已完成解散', icon: 'success' })
        uni.switchTab({ url: '/pages/family/my' })
      } catch (error) {
        errorMessage.value = apiErrorMessage(error, '完成解散失败，请稍后重试。')
      } finally {
        cooldownSubmitting.value = false
      }
    }
  })
}

function confirmCancelDissolution() {
  if (!family.value || family.value.role !== 'FOUNDER' || !currentDissolution.value || dissolutionSubmitting.value) return
  uni.showModal({
    title: '撤销解散申请',
    content: '确定撤销当前待审核的家庭解散申请吗？撤销后家庭会恢复正常状态。',
    success: async (result) => {
      if (!result.confirm || !currentDissolution.value) return
      dissolutionSubmitting.value = true
      try {
        await cancelDissolution(familyId.value, currentDissolution.value.requestId, '创建者主动撤销')
        uni.showToast({ title: '已撤销申请', icon: 'success' })
        await loadFamily()
      } catch (error) {
        errorMessage.value = apiErrorMessage(error, '撤销解散申请失败，请稍后重试。')
      } finally {
        dissolutionSubmitting.value = false
      }
    }
  })
}

function familyStatusText(status: string) {
  switch (status) {
    case 'NORMAL':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'DISSOLVED':
      return '已解散'
    case 'DISSOLUTION_COOLDOWN':
      return '恢复冷静期'
    case 'DISSOLUTION_PENDING':
      return '解散待审核'
    case 'PENDING':
      return '待审核'
    case 'APPROVED':
      return '已通过'
    case 'REJECTED':
      return '已拒绝'
    case 'TAKEN_DOWN':
      return '已下架'
    case 'DELETED':
      return '已删除'
    default:
      return '未知状态'
  }
}

function formatLatestMemberUpdate(memberList: FamilyMember[]) {
  const latestTime = memberList
    .map((member) => Date.parse(member.updatedAt || member.createdAt || ''))
    .filter((time) => Number.isFinite(time))
    .sort((a, b) => b - a)[0]
  if (!latestTime) return '暂无记录'
  return new Date(latestTime).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  })
}

function formatDateTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  })
}

function publicStatusText(status: string) {
  switch (status) {
    case 'APPROVED':
      return '已公开'
    case 'PENDING':
      return '待审核'
    case 'REJECTED':
      return '已拒绝'
    case 'PRIVATE':
      return '未公开'
    case 'TAKEN_DOWN':
      return '已下架'
    default:
      return '未知状态'
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  exitedMemberId.value = String(options?.memberId || '')
})
onShow(loadFamily)
onUnload(resetPageData)
</script>

<style scoped>
.detail-page {
  padding-top: 28rpx;
}

.detail-page :deep(.mini-back-home) {
  margin-bottom: 24rpx;
}

.member-exit-notice {
  display: flex;
  align-items: center;
  margin-top: 16rpx;
  padding: 18rpx 20rpx;
  border: 1rpx solid rgba(168, 59, 45, 0.22);
  border-radius: var(--tree-radius-sm, 12rpx);
  background: rgba(168, 59, 45, 0.06);
}

.quality-family-invite {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  margin-top: 16rpx;
  padding: 22rpx 24rpx;
  border: 1rpx solid rgba(168, 59, 45, 0.3);
  border-radius: var(--tree-radius-sm, 12rpx);
  background: rgba(168, 59, 45, 0.07);
}

.quality-family-invite-main { min-width: 0; }
.quality-family-invite-title,
.quality-family-invite-desc { display: block; }
.quality-family-invite-title { color: var(--archive-cinnabar); font-size: 27rpx; font-weight: 700; }
.quality-family-invite-desc { margin-top: 7rpx; color: var(--archive-ink-soft); font-size: 23rpx; line-height: 1.45; }

.member-exit-title,
.member-exit-desc {
  display: block;
}

.member-exit-title {
  color: var(--archive-ink);
  font-size: 25rpx;
  font-weight: 700;
}

.member-exit-desc {
  margin-top: 6rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.5;
}

.detail-error {
  padding: 28rpx 0;
}

.detail-error :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24rpx;
  margin-bottom: 28rpx;
}

.family-cover {
  display: flex;
  min-height: 210rpx;
  margin-bottom: 30rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.family-cover-copy {
  flex: 1;
  min-width: 0;
  padding: 28rpx 28rpx 28rpx 0;
}

.cover-name,
.cover-desc {
  display: block;
}

.cover-name {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 42rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  line-height: 1.28;
}

.cover-desc {
  margin-top: 14rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.7;
}

.cover-book {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  width: 118rpx;
  flex-shrink: 0;
}

.cover-book text {
  color: rgba(255, 255, 255, 0.94);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 800;
}

.dossier-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border-top: 1rpx solid var(--archive-line);
  border-left: 1rpx solid var(--archive-line);
  margin-bottom: 20rpx;
}

.dossier-cell {
  min-height: 112rpx;
  border-right: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 20rpx;
  box-sizing: border-box;
}

.dossier-label,
.dossier-value {
  display: block;
}

.dossier-label {
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.2;
}

.dossier-value {
  margin-top: 12rpx;
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 700;
  line-height: 1.35;
}

.status-line {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 28rpx;
}

.detail-guide-nudge {
  margin: -4rpx 0 24rpx;
  border-color: rgba(47, 107, 87, 0.25);
  background: linear-gradient(135deg, rgba(47, 107, 87, 0.08), rgba(255, 252, 244, 0.78));
}

.detail-guide-nudge :deep(.mini-button) {
  width: auto;
  min-width: 238rpx;
}

.detail-guide-more {
  display: inline-block;
  margin-top: 18rpx;
  color: var(--archive-blue);
  font-size: 23rpx;
  line-height: 1.4;
}

.recovery-reminder {
  margin: 0 0 18rpx;
}

.detail-profile-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  margin: 18rpx 0 24rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.4;
}

.recovery-reminder .archive-row-title,
.recovery-reminder .archive-arrow {
  color: var(--archive-cinnabar);
}

.invitation-progress {
  margin: -4rpx 0 26rpx;
  padding: 22rpx;
  border: 1rpx solid var(--archive-line);
  border-radius: 12rpx;
  background: var(--archive-paper-deep);
}

.invitation-progress.urgent {
  border-color: rgba(168, 59, 45, 0.42);
  background: rgba(168, 59, 45, 0.055);
}

.invitation-progress-head { display:flex; align-items:center; justify-content:space-between; gap:18rpx; }
.invitation-progress-title,.invitation-progress-desc,.invitation-progress-tip { display:block; }
.invitation-progress-title { color:var(--archive-ink); font-size:28rpx; font-weight:700; }
.invitation-progress-desc,.invitation-progress-tip { color:var(--archive-ink-soft); font-size:24rpx; line-height:1.5; }
.invitation-progress-desc { margin-top:6rpx; }
.invitation-progress-track { position:relative; height:10rpx; margin-top:18rpx; overflow:visible; background:rgba(31,58,95,.1); border-radius:99rpx; }
.invitation-progress-fill { height:100%; background:var(--archive-cinnabar); border-radius:inherit; transition:width .2s ease; }
.invitation-progress-mark { position:absolute; top:50%; width:16rpx; height:16rpx; box-sizing:border-box; border:3rpx solid var(--archive-paper-deep); border-radius:50%; background:var(--archive-ink-soft); transform:translate(-50%,-50%); }.invitation-progress-mark.reached { background:var(--archive-cinnabar); }
.invitation-progress-labels { display:flex; justify-content:space-between; gap:8rpx; margin-top:13rpx; color:var(--archive-ink-soft); font-size:21rpx; }
.invitation-progress-tip { margin-top:10rpx; }

.detail-directory {
  margin-top: 8rpx;
}

.detail-refresh-error {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.detail-refresh-error :deep(.mini-notice) {
  flex: 1;
}

.detail-leave-row .archive-row-title,
.detail-leave-row .archive-row-meta,
.detail-leave-row .archive-arrow {
  color: var(--archive-cinnabar);
}

.affair-badge {
  display: inline-flex;
  min-width: 36rpx;
  height: 36rpx;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border-radius: 18rpx;
  background: var(--archive-cinnabar);
  color: #fff;
  font-size: 20rpx;
  font-weight: 700;
  line-height: 1;
  padding: 0 10rpx;
}

.health-badge {
  background: #9a5d14;
}

.tree-starter {
  margin-bottom: 26rpx;
}

.tree-starter-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.cooldown-panel {
  margin-bottom: 26rpx;
}

.cooldown-panel :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.cooldown-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 16rpx;
}
</style>
