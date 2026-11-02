<template>
  <view class="archive-page manage-center-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />

    <MiniFamilyPageSkeleton v-if="loading && !family" variant="manage" />

    <view v-else-if="errorMessage && !family" class="manage-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </view>

    <template v-else-if="family">
      <view v-if="errorMessage" class="manage-refresh-error">
        <MiniNotice tone="warm" title="刷新未完成">{{ errorMessage }}</MiniNotice>
        <MiniButton size="sm" variant="secondary" @click="loadFamily">重新加载</MiniButton>
      </view>
      <view class="manage-head archive-page-head">
        <view>
          <text class="archive-kicker">Family Affairs</text>
          <text class="archive-title">家庭管理</text>
          <text class="archive-subtitle">
            {{ family.familyName }} · 待办汇总与管理设置
          </text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="manage-context">
        <text class="context-back" @click="openFamilyDetail">{{ returnLabel }}</text>
      </view>

      <view v-if="!canManageFamily" class="manage-state archive-form-panel">
        <MiniNotice tone="warm" title="无管理权限">只有家庭创建者或家庭管理员可以进入家庭管理。</MiniNotice>
        <MiniButton variant="secondary" @click="openFamilyDetail">返回家庭详情</MiniButton>
      </view>

      <template v-else>
        <view v-if="pendingTasks.length" class="manage-tasks archive-form-panel">
          <view class="archive-section-head">
            <text class="archive-section-title">待处理事项</text>
            <text class="archive-section-subtitle">有 <text class="manage-pending-count">{{ pendingTasks.length }}</text> 项需要处理</text>
          </view>
          <view class="archive-list">
            <view v-for="task in pendingTasks" :key="task.key" class="archive-row" @tap="openPendingTask(task)">
              <view class="archive-row-main">
                <text class="archive-row-title">{{ task.title }}</text>
                <text class="archive-row-desc">{{ task.description }}</text>
              </view>
              <text class="task-status" :class="`task-status--${task.tone}`">{{ task.statusLabel }}</text>
              <text class="archive-arrow">›</text>
            </view>
          </view>
        </view>
        <view v-else class="manage-clear-state">
          <view class="manage-clear-mark"><text>已</text></view>
          <view class="manage-clear-copy">
            <text class="manage-clear-title">事项已处理</text>
            <text class="manage-clear-desc">当前没有需要处理的家庭资料、订单或管理申请。</text>
          </view>
        </view>

        <view class="manage-directory">
          <view class="manage-directory-head">
            <text class="manage-directory-title">管理设置</text>
            <text class="manage-directory-desc">需要时可查看记录或调整家庭设置</text>
          </view>
          <view class="archive-list">
            <view class="archive-row" @click="openJoinRequests">
              <view class="archive-row-main">
                <text class="archive-row-title">历史加入记录</text>
                <text class="archive-row-desc">仅查看和处理存量申请</text>
              </view>
              <text class="archive-row-meta">申请</text>
              <text class="archive-arrow">›</text>
            </view>
            <view class="archive-row" @click="openSettings">
              <view class="archive-row-main">
                <text class="archive-row-title">家庭设置</text>
                <text class="archive-row-desc">展示、成员权限提醒、创建者转让、解散与操作记录</text>
              </view>
              <text class="archive-row-meta">设置</text>
              <text class="archive-arrow">›</text>
            </view>
          </view>
        </view>
      </template>

    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getCurrentDissolution } from '@/api/dissolutions'
import { getFamilyDetail } from '@/api/families'
import { getCurrentFounderTransfer } from '@/api/founderTransfers'
import { listFamilyInvitations } from '@/api/invitations'
import { listFamilyJoinRequests } from '@/api/joinRequests'
import { listCapacityOrders, type CapacityOrder } from '@/api/capacity'
import { listFamilyMembers } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { summarizeFamilyHealth } from '@/features/family-tree/familyHealthSummary'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, Invitation, JoinRequest } from '@/types/api'

type PendingTask = {
  key: string
  title: string
  description: string
  statusLabel: string
  tone: 'attention' | 'review' | 'risk'
  route: string
}

const session = useSessionStore()
const familyId = ref('')
const returnTo = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const pendingTasks = ref<PendingTask[]>([])

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : '返回家庭详情')

async function loadFamily() {
  session.restoreSession()
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/manage-center?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = !family.value
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    const familyDetail = await getFamilyDetail(familyId.value)
    family.value = familyDetail
    if (familyDetail.role === 'FOUNDER' || familyDetail.role === 'FAMILY_ADMIN') {
      const [invitations, joinRequests, transfer, dissolution, memberRows, tree, capacityOrders] = await Promise.all([
        listFamilyInvitations(familyId.value),
        listFamilyJoinRequests(familyId.value),
        getCurrentFounderTransfer(familyId.value).catch(() => null),
        getCurrentDissolution(familyId.value).catch(() => null),
        listFamilyMembers(familyId.value),
        getPrivateTree(familyId.value).catch(() => null),
        familyDetail.role === 'FOUNDER' ? listCapacityOrders(familyId.value).catch(() => []) : Promise.resolve([])
      ])
      pendingTasks.value = buildPendingTasks({
        invitations,
        joinRequests,
        hasTransfer: transfer?.requestStatus === 'PENDING',
        hasDissolution: dissolution?.requestStatus === 'PENDING',
        healthIssueCount: summarizeFamilyHealth(familyDetail, memberRows, tree).total,
        capacityOrders
      })
    } else {
      pendingTasks.value = []
    }
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭管理加载失败。')
  } finally {
    loading.value = false
  }
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  family.value = null
  pendingTasks.value = []
}

function openFamilyDetail() {
  uni.redirectTo({ url: returnUrl.value })
}

function openJoinRequests() {
  uni.navigateTo({ url: `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}${returnTo.value === 'tree' ? '&returnTo=tree' : ''}` })
}

function openSettings() {
  uni.navigateTo({ url: `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}${returnTo.value === 'tree' ? '&returnTo=tree' : ''}` })
}

function buildPendingTasks(input: {
  invitations: Invitation[]
  joinRequests: JoinRequest[]
  hasTransfer: boolean
  hasDissolution: boolean
  healthIssueCount: number
  capacityOrders: CapacityOrder[]
}): PendingTask[] {
  const tasks: PendingTask[] = []
  const treeReturn = returnTo.value === 'tree' ? '&returnTo=tree' : ''
  if (input.hasDissolution) tasks.push({ key: 'dissolution', title: '家庭解散 · 审核中', description: '该操作影响全体成员，可查看进度或撤销申请', statusLabel: '需关注', tone: 'risk', route: `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}&section=dissolution${treeReturn}` })
  if (input.hasTransfer) tasks.push({ key: 'transfer', title: '创建者转让 · 审核中', description: '可查看审核进度或撤销申请', statusLabel: '审核中', tone: 'review', route: `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}&section=transfer${treeReturn}` })
  if (input.healthIssueCount > 0) tasks.push({ key: 'health', title: '完善家庭资料与关系', description: `发现 ${input.healthIssueCount} 项待补充内容，完成后可生成家庭档案长卷`, statusLabel: '待完善', tone: 'attention', route: `/pages/family/health?familyId=${encodeURIComponent(familyId.value)}${treeReturn}` })
  for (const item of input.capacityOrders) {
    const paymentPending = item.status === 'PENDING' && (!item.expiresAt || Date.parse(item.expiresAt) > Date.now())
    if (paymentPending) tasks.push({ key: `payment-${item.orderNo}`, title: `继续支付 ${item.productName}`, description: '订单尚未完成支付，名额暂未生效', statusLabel: '待支付', tone: 'attention', route: `/pages/family/capacity-orders?familyId=${encodeURIComponent(familyId.value)}&orderNo=${encodeURIComponent(item.orderNo)}${treeReturn}` })
    if (item.status === 'REFUNDING') tasks.push({ key: `refund-${item.orderNo}`, title: `${item.productName}退款审核中`, description: '扩容名额已暂时冻结，审核结果会通过消息通知', statusLabel: '审核中', tone: 'review', route: `/pages/family/capacity-orders?familyId=${encodeURIComponent(familyId.value)}&orderNo=${encodeURIComponent(item.orderNo)}${treeReturn}` })
  }
  for (const item of input.joinRequests) {
    if (item.requestStatus !== 'PENDING') continue
    tasks.push({ key: `join-${item.requestId}`, title: '处理历史加入申请', description: item.applicantRealName || '存量申请等待处理', statusLabel: '待处理', tone: 'attention', route: `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}&requestId=${encodeURIComponent(String(item.requestId))}${treeReturn}` })
  }
  for (const item of input.invitations) {
    if (item.status !== 'PENDING' || Date.parse(item.expiredAt) <= Date.now()) continue
    tasks.push({ key: `invite-${item.invitationId}`, title: `继续邀请 ${item.targetMemberName}`, description: '直接进入该成员的邀请处理页', statusLabel: '待接受', tone: 'attention', route: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}&mode=node&memberId=${encodeURIComponent(String(item.targetMemberId))}&memberName=${encodeURIComponent(item.targetMemberName)}${treeReturn}` })
  }
  return tasks
}

function openPendingTask(task: PendingTask) {
  uni.navigateTo({ url: task.route })
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  returnTo.value = String(options?.returnTo || '')
})
onShow(loadFamily)
onUnload(resetPageData)
</script>

<style scoped>
.manage-center-page {
  padding-top: 28rpx;
}

.manage-center-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.manage-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.manage-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 24rpx;
}

.context-back {
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.4;
}

.context-back:active {
  color: var(--archive-cinnabar);
}

.manage-directory {
  margin-top: 30rpx;
}

.manage-directory-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 18rpx;
  padding: 0 4rpx 14rpx;
}

.manage-directory-title {
  flex-shrink: 0;
  color: var(--archive-ink);
  font-size: 29rpx;
  font-weight: 800;
}

.manage-directory-desc {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.5;
  text-align: right;
}

.manage-clear-state {
  display: flex;
  align-items: center;
  gap: 18rpx;
  padding: 24rpx 4rpx 6rpx;
}

.manage-clear-mark {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 58rpx;
  height: 58rpx;
  border: 1rpx solid rgba(42, 111, 91, 0.28);
  border-radius: 50%;
  background: rgba(42, 111, 91, 0.08);
  color: #2a6f5b;
  font-size: 24rpx;
  font-weight: 800;
}

.manage-clear-copy {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 5rpx;
}

.manage-clear-title {
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 800;
}

.manage-clear-desc {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.55;
}

.manage-refresh-error {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.manage-refresh-error :deep(.mini-notice) {
  flex: 1;
}

.manage-tasks {
  margin-top: 24rpx;
}

.manage-pending-count {
  color: var(--archive-cinnabar);
  font-size: 27rpx;
  font-weight: 800;
}

.task-status {
  flex-shrink: 0;
  min-width: 66rpx;
  box-sizing: border-box;
  padding: 7rpx 10rpx;
  border: 1rpx solid transparent;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 700;
  line-height: 1.25;
  text-align: center;
}

.task-status--attention {
  border-color: rgba(168, 59, 45, 0.32);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
}

.task-status--review {
  border-color: rgba(31, 58, 95, 0.24);
  background: rgba(31, 58, 95, 0.08);
  color: var(--archive-blue);
}

.task-status--risk {
  border-color: rgba(181, 122, 41, 0.34);
  background: rgba(181, 122, 41, 0.1);
  color: #9a5d14;
}
</style>
