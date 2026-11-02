<template>
  <view class="archive-page private-tree-page">
    <MiniBackHome :return-url="familyReturnUrl" :return-label="familyReturnLabel" />

    <view v-if="errorMessage && !tree" class="tree-error archive-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadTree">重新加载</MiniButton>
    </view>

    <view v-else-if="loading && !tree" class="tree-loading archive-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在组装家庭树..." />
    </view>

    <template v-else-if="tree">
      <view class="tree-head">
        <view>
        <text class="archive-kicker">Family Tree</text>
        <text class="archive-title">家庭树</text>
        <text class="archive-subtitle">
            {{ family?.familyName || '家庭关系' }} · 家庭树 · {{ family?.memberCount ?? tree.nodes.length }} / {{ family?.memberLimit || '—' }} 位成员
        </text>
        </view>
        <view class="tree-head-actions">
          <view class="tree-export-action" :class="{ loading: treeExporting }" @tap="exportFamilyTree">
            <text>{{ treeExporting ? '生成中' : '导出家庭树' }}</text>
          </view>
          <view v-if="family" class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
        </view>
      </view>

      <MiniNotice v-if="guideMode" tone="security" class="tree-guide-note" title="从这里认识家庭树">
        先看“当前视角”是谁；点击任意人物可查看称谓、切换视角或继续处理该节点。资料或关系待补充会显示红点。
      </MiniNotice>

      <view v-if="tree.nodes.length > 0" class="viewer-panel archive-form-panel">
        <view class="viewer-copy">
          <text class="viewer-label">当前视角</text>
          <text class="viewer-name">{{ currentViewerName }}</text>
          <text class="viewer-hint">深蓝为直系关系，点击人物卡可切换视角。</text>
        </view>
        <view class="viewer-actions">
          <view v-if="canResetViewerToSelf" class="viewer-action" @tap="resetViewerToSelf">回到我的视角</view>
          <view v-if="viewerMemberId !== null" class="viewer-action" @tap="showAllTree">查看全树</view>
        </view>
      </view>

      <view class="tree-affairs archive-form-panel" @tap="openTreeAffairs">
        <view class="tree-affairs-copy">
          <text class="tree-affairs-title">家庭事务</text>
          <text class="tree-affairs-desc">{{ treeAffairsDescription }}</text>
        </view>
        <view class="tree-affairs-meta">
          <text v-if="treeIssueCount" class="tree-affairs-count">{{ treeIssueCount }}</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>

      <view v-if="Object.keys(healthHints).length" class="tree-health-note">
        <text class="tree-health-dot">!</text>
        <text>红点表示资料或关系待补充，可点击人物卡处理或查看家庭健康。</text>
      </view>

      <view class="tree-status-legend" aria-label="家庭树状态说明">
        <view class="tree-status-item">
          <text class="legend-sample health">!</text>
          <text>资料或关系待补充</text>
        </view>
        <view class="tree-status-item">
          <text class="legend-sample deceased">名</text>
          <text>已离世亲人</text>
        </view>
        <view class="tree-status-item">
          <text class="legend-sample binding">已绑</text>
          <text>已绑定账号</text>
        </view>
        <view class="tree-status-item">
          <text class="legend-sample pending">待绑</text>
          <text>尚未绑定账号</text>
        </view>
      </view>

      <view class="tree-mode-line ui-simplified-hidden">
        <view
          class="tree-mode-item"
          :class="{ active: viewMode === 'structure' }"
          @click="viewMode = 'structure'"
        >
          <text>树结构</text>
        </view>
        <view
          class="tree-mode-item"
          :class="{ active: viewMode === 'detail' }"
          @click="viewMode = 'detail'"
        >
          <text>关系明细</text>
        </view>
      </view>

      <view class="tree-scroll-note ui-simplified-hidden">
        <text>树状家谱可左右滑动查看；配偶卡带有婚嫁标记，不作为关系扩展入口。</text>
      </view>

      <template v-if="viewMode === 'structure'">
        <view v-if="tree.nodes.length === 0" class="empty-tree-guide archive-form-panel">
          <MiniEmptyState
            symbol="树"
            title="暂无成员"
            description="请先添加家庭成员，再查看树状家谱。"
          />
        </view>
        <FamilyTreeStructureView
          v-else
          class="archive-tree-structure"
          :tree="displayTree || tree"
          :viewer-member-id="viewerMemberId"
          show-binding
          :interactive="true"
          selectable-all
          :selectable-deceased="grayAccessAvailable"
          :health-hints="healthHints"
          portrait-theme="regional"
          :portrait-region="family?.regionText || family?.nativePlace"
          :portrait-region-code="family?.regionCode"
          @select="openNodeActions"
        />
      </template>

      <template v-else>
        <MiniEmptyState
          v-if="edgeCount === 0"
          symbol="亲"
          title="暂无关系"
          description="暂无父母子女或配偶关系。"
        />
        <RelationSentenceList v-else class="relation-list" :items="relationSentences" />
      </template>

      <view v-if="selectedNode" class="node-action-panel archive-form-panel">
        <view class="node-action-head">
          <view class="node-action-copy">
            <text class="node-action-label">已选成员</text>
            <text class="node-action-name">{{ selectedNode.displayName || '未命名成员' }}</text>
          </view>
          <view class="node-action-head-actions">
            <view v-if="nodeActionMenu !== 'MAIN'" class="node-action-back" @tap="openNodeActionMenu('MAIN')">返回</view>
            <view class="node-action-close" @tap="closeNodeActions">收起</view>
          </view>
        </view>
        <text class="node-action-hint">{{ nodeActionMenuTitle }}</text>
        <view v-if="nodeActionMenu === 'MAIN'" class="node-action-grid">
          <view
            v-for="action in selectedNodeActions"
            :key="action.label"
            class="node-action-item"
            :class="{ danger: action.danger, primary: action.primary }"
            @tap="runNodeAction(action)"
          >
            {{ action.label }}
          </view>
        </view>
        <view v-else-if="nodeActionMenu === 'RELATION'" class="node-action-grid">
          <view
            v-for="relation in selectedNodeRelationAdditions"
            :key="relation.addType"
            class="node-action-item primary"
            @tap="openSelectedNodeRelation(relation.addType)"
          >
            {{ relation.label }}
          </view>
        </view>
        <view v-else class="node-action-grid">
          <view
            v-for="action in selectedNodeMoreActions"
            :key="action.label"
            class="node-action-item"
            :class="{ danger: action.danger }"
            @tap="runNodeAction(action)"
          >
            {{ action.label }}
          </view>
        </view>
      </view>

      <view class="tree-legend ui-simplified-hidden">
        <view class="legend-item">
          <text class="legend-name male">名</text>
          <text>男 · 墨蓝姓名</text>
        </view>
        <view class="legend-item">
          <text class="legend-name female">名</text>
          <text>女 · 朱砂姓名</text>
        </view>
        <view class="legend-item">
          <text class="legend-name unknown">名</text>
          <text>未知 · 灰墨姓名</text>
        </view>
        <view class="legend-item">
          <text class="legend-tag lineage" />
          <text>家庭成员</text>
        </view>
        <view class="legend-item">
          <text class="legend-tag current" />
          <text>当前中心</text>
        </view>
        <view class="legend-item">
          <text class="legend-tag spouse" />
          <text>配偶</text>
        </view>
      </view>

    </template>

    <view v-if="memorialMember" class="memorial-mask" @click="closeMemorial">
      <view class="memorial-sheet" @click.stop>
        <view class="memorial-sheet-head">
          <view>
            <text class="memorial-kicker">IN MEMORY</text>
            <text class="memorial-name">{{ memorialMember.name || memorialNode?.displayName || '亲人' }}</text>
          </view>
          <view class="memorial-seal">念</view>
        </view>

        <view class="memorial-time">
          <text class="memorial-time-label">离世时间</text>
          <text class="memorial-time-value">{{ formatDeathTime(memorialMember) }}</text>
        </view>

        <view class="memorial-content">
          <text class="memorial-content-label">亲人简介</text>
          <text class="memorial-content-value">
            {{ memorialDescriptionText }}
          </text>
        </view>

        <view class="memorial-foot">
          <text class="memorial-footnote">{{ memorialFootnote }}</text>
          <view class="memorial-actions">
            <view class="memorial-action secondary" @click="closeMemorial">关闭</view>
            <view class="memorial-action secondary" @click="switchMemorialViewer">以此亲人为视角</view>
            <view class="memorial-action primary" @click="viewMemorialMember">查看资料</view>
          </view>
          <view v-if="canDeleteMemorialMember" class="memorial-action danger memorial-delete-action" @click="deleteMemorialMember">删除节点</view>
          <text v-else-if="memorialDeleteRestriction" class="memorial-delete-restriction">{{ memorialDeleteRestriction }}</text>
        </view>
      </view>
    </view>

    <MiniPhoneUpgradeDialog v-model:visible="phoneUpgradeVisible" @bound="continueAfterPhoneUpgrade" />
    <canvas
      v-if="treePosterCanvasHeight"
      canvas-id="privateTreePosterCanvas"
      id="privateTreePosterCanvas"
      class="tree-export-canvas"
      :style="{ width: `${treePosterCanvasWidth}px`, height: `${treePosterCanvasHeight}px` }"
      :width="treePosterCanvasWidth"
      :height="treePosterCanvasHeight"
    />
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, getCurrentInstance, nextTick, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { fetchCapabilities } from '@/api/capabilities'
import { getFamilyDetail, leaveFamily } from '@/api/families'
import { listFamilyInvitations } from '@/api/invitations'
import { deleteFamilyMember, getFamilyMember, listFamilyMembers, updateFamilyMember } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniPhoneUpgradeDialog from '@/components/base/MiniPhoneUpgradeDialog.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import { resolveCurrentMemberId } from '@/features/family-tree/resolveCurrentMember'
import { buildRelationSentences, countVisibleEdges } from '@/features/family-tree/relationSentences'
import { inspectRelationshipHealth } from '@/features/family-tree/relationshipHealth'
import { buildFamilyTreePosterPage, drawExportTreePoster, resolveExportTreePortraitPaths } from '@/features/family-tree/exportTreePoster'
import type { FamilyTreeViewMode } from '@/features/family-tree/types'
import { isLineageMember, isSpouseMember, isExternalMember } from '@/features/family-tree/graph'
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyMember, FamilyTreeResult, Invitation, RelationshipAddType, TreeNode, UserCapabilities } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const returnTo = ref('')
const guideMode = ref(false)
const familyReturnUrl = computed(() => returnTo.value === 'health'
  ? `/pages/family/health?familyId=${encodeURIComponent(familyId.value)}`
  : `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const familyReturnLabel = computed(() => returnTo.value === 'health' ? '返回家庭健康' : '返回家庭详情')
const family = ref<FamilyDetail | null>(null)
const tree = ref<FamilyTreeResult | null>(null)
const invitations = ref<Invitation[]>([])
const loading = ref(false)
const errorMessage = ref('')
const viewMode = ref<FamilyTreeViewMode>('structure')
const viewerMemberId = ref<number | null>(null)
const requestedViewerMemberId = ref<number | null>(null)
const requestedRelationAddType = ref<RelationshipAddType | null>(null)
const requestedNodeAction = ref<'INVITE' | null>(null)
const currentUserMemberId = ref<number | null>(null)
const memorialMember = ref<FamilyMember | null>(null)
const memorialNode = ref<TreeNode | null>(null)
const grayAccessAvailable = ref(false)
const phoneUpgradeVisible = ref(false)
const treeExporting = ref(false)
const treePosterPaths = ref<string[]>([])
const treePosterCanvasWidth = ref(750)
const treePosterCanvasHeight = ref(0)
const instance = getCurrentInstance()
const leaving = ref(false)
const pendingAddNode = ref<TreeNode | null>(null)
const pendingAddType = ref<RelationshipAddType | null>(null)
type NodeAction = {
  label: string
  primary?: boolean
  danger?: boolean
  keepPanelOpen?: boolean
  run: () => void
}
type NodeActionMenu = 'MAIN' | 'RELATION' | 'MORE'
const selectedNode = ref<TreeNode | null>(null)
const nodeActionMenu = ref<NodeActionMenu>('MAIN')
const edgeCount = computed(() => (tree.value ? countVisibleEdges(tree.value) : 0))
const relationSentences = computed(() => (tree.value ? buildRelationSentences(tree.value) : []))
const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const canLeaveCurrentFamily = computed(() =>
  family.value?.role !== 'FOUNDER' && family.value?.status === 'NORMAL' && currentUserMemberId.value != null
)
const viewerNodes = computed(() => tree.value?.nodes || [])
const pendingInviteMemberIds = computed(() => {
  const now = Date.now()
  return new Set(
    invitations.value
      .filter((item) => item.status === 'PENDING' && Date.parse(item.expiredAt) > now)
      .map((item) => Number(item.targetMemberId))
      .filter((memberId) => Number.isFinite(memberId))
  )
})
const healthHints = computed<Record<number, string[]>>(() => {
  const hints: Record<number, string[]> = {}
  const relationshipHints = inspectRelationshipHealth(tree.value).hintsByMemberId
  const itemsByMemberID = new Map((tree.value?.tree || []).map((item) => [item.memberId, item]))
  for (const node of tree.value?.nodes || []) {
    const nodeHints: string[] = []
    if (node.gender === 'UNKNOWN') nodeHints.push('待补性别')
    if (!node.birthDate) nodeHints.push('待补出生年份')
    const item = itemsByMemberID.get(node.memberId)
    if (item && item.parentIds.length === 0 && item.childrenIds.length === 0 && item.spouseIds.length === 0 && (tree.value?.nodes.length || 0) > 1) nodeHints.push('待补关系')
    nodeHints.push(...(relationshipHints[node.memberId] || []))
    if (nodeHints.length) hints[node.memberId] = nodeHints
  }
  return hints
})
const treeIssueCount = computed(() => Object.keys(healthHints.value).length)
const treeAffairsDescription = computed(() => {
  if (canManageFamily.value) {
    return treeIssueCount.value
      ? `${treeIssueCount.value} 个成员资料或关系待补充，也可处理邀请与管理申请`
      : '处理邀请、审核与家庭设置'
  }
  return treeIssueCount.value
    ? `${treeIssueCount.value} 个成员资料或关系待补充`
    : '查看家庭健康并生成家谱长卷'
})
const displayTree = computed<FamilyTreeResult | null>(() => {
  if (!tree.value || pendingInviteMemberIds.value.size === 0) return tree.value
  return {
    ...tree.value,
    nodes: tree.value.nodes.map((node) =>
      pendingInviteMemberIds.value.has(node.memberId) && node.userBindingState === 'UNBOUND'
        ? { ...node, userBindingState: 'INVITING' }
        : node
    )
  }
})
const currentViewerNode = computed(() =>
  viewerNodes.value.find((node) => node.memberId === viewerMemberId.value) || null
)
const currentViewerName = computed(() => currentViewerNode.value?.displayName || '全树视角')
const canResetViewerToSelf = computed(() => {
  const selfViewerId = normalizeViewerMemberId(currentUserMemberId.value || tree.value?.nodes[0]?.memberId || null)
  return viewerMemberId.value !== selfViewerId
})

function deleteRestriction(node: TreeNode | null | undefined): string {
  if (!node || !canManageFamily.value) return ''
  if (node.memberId === family.value?.currentFounderMemberId) return '当前创建人需先完成转让，不能删除。'
  const hasChildren = (tree.value?.edges || []).some((edge) =>
    edge.relationshipType === 'PARENT_CHILD' && edge.fromMemberId === node.memberId
  )
  return hasChildren ? '该成员已有子女，不能删除。' : ''
}

function canDeleteTreeNode(node: TreeNode) {
  return canManageFamily.value && !deleteRestriction(node)
}

const memorialDeleteRestriction = computed(() => deleteRestriction(memorialNode.value))
const canDeleteMemorialMember = computed(() => Boolean(memorialNode.value && canDeleteTreeNode(memorialNode.value)))

function canSwitchViewer(node: TreeNode) {
	return Boolean(tree.value?.nodes.some((item) => item.memberId === node.memberId))
}

function normalizeViewerMemberId(memberId: number | null | undefined): number | null {
  if (memberId == null || !tree.value) return memberId ?? null
  const node = tree.value.nodes.find((item) => item.memberId === memberId)
  if (!isSpouseMember(node)) return memberId

  for (const edge of tree.value.edges) {
    if (edge.relationshipType !== 'SPOUSE') continue
    const spouseId =
      edge.fromMemberId === memberId
        ? edge.toMemberId
        : edge.toMemberId === memberId
          ? edge.fromMemberId
          : null
    if (spouseId == null) continue
    const spouseNode = tree.value.nodes.find((item) => item.memberId === spouseId)
    if (isLineageMember(spouseNode)) return spouseId
  }

  return memberId
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  tree.value = null
  requestedViewerMemberId.value = null
  requestedRelationAddType.value = null
  requestedNodeAction.value = null
  invitations.value = []
  family.value = null
  viewerMemberId.value = null
  currentUserMemberId.value = null
  memorialMember.value = null
  memorialNode.value = null
  selectedNode.value = null
  grayAccessAvailable.value = false
}

async function loadTree() {
  session.restoreSession()
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = !tree.value
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    const familyResult = await getFamilyDetail(familyId.value)
    const canManage = familyResult.role === 'FOUNDER' || familyResult.role === 'FAMILY_ADMIN'
    const [treeResult, invitationResult] = await Promise.all([
      getPrivateTree(familyId.value),
      canManage ? listFamilyInvitations(familyId.value) : Promise.resolve([])
    ])
    family.value = familyResult
    tree.value = treeResult
    invitations.value = invitationResult
    grayAccessAvailable.value = Boolean(treeResult.grayAccessEnabled)
    let resolvedViewerId = currentUserMemberId.value
    if (session.user?.id && currentUserMemberId.value == null) {
      try {
        const memberList = await listFamilyMembers(familyId.value)
        resolvedViewerId = resolveCurrentMemberId(session.user.id, memberList)
        currentUserMemberId.value = resolvedViewerId
      } catch {
        // Tree nodes lack userId; members list is the supported way to locate "me".
      }
    }
    const requestedViewer = requestedViewerMemberId.value
    if (requestedViewer && treeResult.nodes.some((node) => node.memberId === requestedViewer)) {
      viewerMemberId.value = normalizeViewerMemberId(requestedViewer)
      requestedViewerMemberId.value = null
      const addType = requestedRelationAddType.value
      requestedRelationAddType.value = null
      const nodeAction = requestedNodeAction.value
      requestedNodeAction.value = null
      if (addType && canManageFamily.value) {
        const target = treeResult.nodes.find((node) => node.memberId === requestedViewer)
        if (target) setTimeout(() => { void openAddRelation(target, addType) }, 0)
      }
      if (nodeAction === 'INVITE' && canManageFamily.value) {
        const target = treeResult.nodes.find((node) => node.memberId === requestedViewer)
        if (target) setTimeout(() => openInviteMember(target), 0)
      }
      return
    }
    const hasCurrentViewer = treeResult.nodes.some((node) => node.memberId === viewerMemberId.value)
    if (!hasCurrentViewer) {
      viewerMemberId.value = normalizeViewerMemberId(resolvedViewerId || treeResult.nodes[0]?.memberId || null)
    } else {
      viewerMemberId.value = normalizeViewerMemberId(viewerMemberId.value)
    }
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭树加载失败。')
  } finally {
    loading.value = false
  }
}

function treeCanvasToTempFilePath(width: number, height: number): Promise<string> {
  return new Promise((resolve, reject) => {
    uni.canvasToTempFilePath({
      canvasId: 'privateTreePosterCanvas', x: 0, y: 0, width, height,
      destWidth: width, destHeight: height, fileType: 'png',
      success: (result) => resolve(result.tempFilePath), fail: reject
    }, instance?.proxy)
  })
}

async function exportFamilyTree() {
  if (!tree.value || !family.value || treeExporting.value) return
  treeExporting.value = true
  try {
    // Tree export intentionally keeps the complete current family graph. Archive
    // export remains the separate, scope-controlled long-scroll workflow.
    const page = buildFamilyTreePosterPage(
      tree.value,
      `${family.value.familyName}家庭树`,
      `${family.value.memberCount ?? tree.value.nodes.length} 位成员 · 完整家庭树`
    )
    treePosterCanvasWidth.value = page.width
    treePosterCanvasHeight.value = page.height
    await nextTick()
    const context = uni.createCanvasContext('privateTreePosterCanvas', instance?.proxy)
    const portraitPaths = await resolveExportTreePortraitPaths(page.avatarUrlByMemberId)
    drawExportTreePoster(context, page, portraitPaths)
    await new Promise<void>((resolve) => context.draw(false, resolve))
    treePosterPaths.value = [await treeCanvasToTempFilePath(page.width, page.height)]
    uni.showActionSheet({
      itemList: ['保存到相册', '预览或分享图片'],
      success: (result) => {
        if (result.tapIndex === 0) void saveFamilyTreeImage()
        else shareFamilyTreeImage()
      }
    })
  } catch (error) {
    uni.showModal({ title: '生成失败', content: apiErrorMessage(error, '家庭树图片生成失败，请稍后重试。'), showCancel: false })
  } finally {
    treeExporting.value = false
  }
}

async function saveFamilyTreeImage() {
  const filePath = treePosterPaths.value[0]
  if (!filePath) return
  try {
    await new Promise<void>((resolve, reject) => uni.saveImageToPhotosAlbum({ filePath, success: () => resolve(), fail: reject }))
    uni.showToast({ title: '已保存到相册', icon: 'success' })
  } catch {
    uni.showModal({ title: '无法保存图片', content: '请允许小程序保存到相册后重试。', showCancel: false })
  }
}

function shareFamilyTreeImage() {
  const filePath = treePosterPaths.value[0]
  if (!filePath) return
  const wxApi = (globalThis as typeof globalThis & { wx?: { showShareImageMenu?: (options: { path: string }) => void } }).wx
  if (wxApi?.showShareImageMenu) {
    wxApi.showShareImageMenu({ path: filePath })
    return
  }
  uni.previewImage({ urls: treePosterPaths.value, current: filePath })
}

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function resetViewerToSelf() {
  viewerMemberId.value = normalizeViewerMemberId(currentUserMemberId.value || tree.value?.nodes[0]?.memberId || null)
  uni.showToast({ title: '已回到我的视角', icon: 'success' })
}

function showAllTree() {
  viewerMemberId.value = null
  uni.showToast({ title: '已显示全树', icon: 'success' })
}

function switchViewerToNode(node: TreeNode) {
	if (!canSwitchViewer(node)) return
	viewerMemberId.value = normalizeViewerMemberId(node.memberId)
  viewMode.value = 'structure'
  uni.showToast({ title: '已切换视角', icon: 'success' })
}

function openNodeActions(node: TreeNode) {
  if (node.isLiving === false && grayAccessAvailable.value) {
    selectedNode.value = null
    showMemorial(node)
    return
  }
  nodeActionMenu.value = 'MAIN'
  selectedNode.value = node
}

const selectedNodeRelationAdditions = computed(() => {
  const node = selectedNode.value
  if (!node || !canManageFamily.value || isSpouseMember(node) || isExternalMember(node)) return []
  return availableRelationAdditions(node)
})

const selectedNodeActions = computed<NodeAction[]>(() => {
  const node = selectedNode.value
  if (!node) return []
  const actions: NodeAction[] = []
  const relationAdditions = selectedNodeRelationAdditions.value
  actions.push({ label: '查看资料', primary: true, run: () => openMemberView(node) })
  if (healthHints.value[node.memberId]?.length && (canManageFamily.value || node.memberId === currentUserMemberId.value)) {
    actions.push({ label: '补充资料', primary: true, run: () => openMemberEdit(node) })
  }
  if (canManageFamily.value) {
    if (node.userBindingState === 'UNBOUND') {
      actions.push({ label: '邀请本人绑定', primary: true, run: () => openInviteMember(node) })
    }
    if (node.userBindingState === 'INVITING') {
      actions.push({ label: '继续邀请', primary: true, run: () => openInviteMember(node) })
    }
    if (relationAdditions.length > 0) {
      actions.push({ label: '添加关系', primary: true, keepPanelOpen: true, run: () => openNodeActionMenu('RELATION') })
    }
  }
  if (selectedNodeMoreActions.value.length > 0) {
    actions.push({ label: '更多操作', keepPanelOpen: true, run: () => openNodeActionMenu('MORE') })
  }
  return actions
})

const selectedNodeMoreActions = computed<NodeAction[]>(() => {
  const node = selectedNode.value
  if (!node) return []
  const actions: NodeAction[] = []
  if (canManageFamily.value) {
    if (selectedNodeRelationAdditions.value.length > 0) {
      actions.push({ label: '调整已有关系', run: () => openAdjustRelation(node) })
    }
    if (node.isLiving === true) {
      actions.push({ label: '标记为已故', run: () => updateLivingState(node, false) })
    } else if (node.isLiving === false) {
      actions.push({ label: '标记为健在', run: () => updateLivingState(node, true) })
    } else {
      actions.push({ label: '标记健在状态', run: () => openLivingStateActions(node) })
    }
    if (canDeleteTreeNode(node)) actions.push({ label: '删除节点', danger: true, run: () => confirmDeleteMember(node) })
  }
  if (canSwitchViewer(node)) actions.push({ label: '切换到此视角', run: () => switchViewerToNode(node) })
  if (canLeaveCurrentFamily.value && node.memberId === currentUserMemberId.value) {
    actions.push({ label: '退出当前家庭', danger: true, run: () => confirmLeave() })
  }
  return actions
})

const nodeActionMenuTitle = computed(() => {
  if (nodeActionMenu.value === 'RELATION') return '选择要添加的关系'
  if (nodeActionMenu.value === 'MORE') return '更多成员操作'
  return '请选择下一步操作'
})

function openNodeActionMenu(menu: NodeActionMenu) {
  nodeActionMenu.value = menu
}

function closeNodeActions() {
  selectedNode.value = null
  nodeActionMenu.value = 'MAIN'
}

function runNodeAction(action: NodeAction) {
  if (!action.keepPanelOpen) closeNodeActions()
  action.run()
}

function openSelectedNodeRelation(addType: RelationshipAddType) {
  const node = selectedNode.value
  if (!node) return
  closeNodeActions()
  void openAddRelation(node, addType)
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
    await leaveFamily(familyId.value, '用户从家庭树退出')
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

function formatDeathTime(member: { deathDate?: string | null; deathYear?: number | null }) {
  if (member.deathDate) return member.deathDate.slice(0, 10)
  if (member.deathYear) return `${member.deathYear} 年`
  return '未记录'
}

async function showMemorial(node: TreeNode) {
  try {
    const member = await getFamilyMember(familyId.value, node.memberId)
    memorialNode.value = node
    memorialMember.value = member
  } catch (error) {
    uni.showModal({
      title: '读取失败',
      content: apiErrorMessage(error, '亲人简介读取失败。'),
      showCancel: false
    })
  }
}

const memorialDescription = computed(() => memorialMember.value?.description?.trim() || '')
const memorialDescriptionText = computed(() => {
  if (memorialDescription.value) return memorialDescription.value
  if (memorialMember.value?.memorialVisible === false) return '这位亲友的简介暂未向家庭成员展示。'
  return '暂未记录亲人简介。'
})
const memorialFootnote = computed(() =>
  memorialMember.value?.memorialVisible === false && !canManageFamily.value
    ? '创建人或管理员可选择是否展示这段记述。'
    : '愿这段记述，留存于家庭的记忆中。'
)

function closeMemorial() {
  memorialMember.value = null
  memorialNode.value = null
}

function viewMemorialMember() {
  const node = memorialNode.value
  closeMemorial()
  if (node) openMemberView(node)
}

function switchMemorialViewer() {
  const node = memorialNode.value
  closeMemorial()
  if (node) switchViewerToNode(node)
}

function deleteMemorialMember() {
  const node = memorialNode.value
  closeMemorial()
  if (node) confirmDeleteMember(node)
}

function availableRelationAdditions(node: TreeNode) {
  const parentIds = new Set(
    (tree.value?.edges || [])
      .filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.toMemberId === node.memberId)
      .map((edge) => edge.fromMemberId)
  )
  const parents = (tree.value?.nodes || []).filter((item) => parentIds.has(item.memberId))
  const hasFather = parents.some((parent) => parent.gender === 'MALE')
  const hasMother = parents.some((parent) => parent.gender === 'FEMALE')
  const canExpandByGender = node.gender === 'MALE' || node.gender === 'FEMALE'
  const additions: Array<{ label: string; addType: RelationshipAddType }> = []
  if (!hasFather) additions.push({ label: '添加父亲', addType: 'ADD_FATHER' })
  if (!hasMother) additions.push({ label: '添加母亲', addType: 'ADD_MOTHER' })
  if (canExpandByGender) {
    additions.push({ label: '添加子女', addType: 'ADD_CHILD' })
    additions.push({ label: '添加配偶', addType: 'ADD_SPOUSE' })
  }
  if (parents.length > 0) additions.push({ label: '添加兄弟姐妹', addType: 'ADD_SIBLING' })
  return additions
}

function openLivingStateActions(node: TreeNode) {
  const items =
    node.isLiving === true
      ? ['标记为已故']
      : node.isLiving === false
        ? ['标记为健在']
        : ['标记为健在', '标记为已故']
  uni.showActionSheet({
    itemList: items,
    success: (result) => {
      const action = items[result.tapIndex]
      updateLivingState(node, action === '标记为健在')
    }
  })
}

function openMemberView(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/family/member-view?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(String(node.memberId))}&returnTo=tree`
  })
}

function openMemberEdit(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/family/member-edit?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(String(node.memberId))}&returnTo=tree`
  })
}

function openTreeAffairs() {
  const page = canManageFamily.value ? 'manage-center' : 'health'
  uni.navigateTo({
    url: `/pages/family/${page}?familyId=${encodeURIComponent(familyId.value)}&returnTo=tree`
  })
}

function openInviteMember(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}&mode=node&memberId=${encodeURIComponent(String(node.memberId))}&memberName=${encodeURIComponent(node.displayName || '')}&returnTo=tree`
  })
}

async function openAddRelation(node: TreeNode, addType: RelationshipAddType) {
  if (!await ensureMemberCapacity(node, addType)) return
  uni.navigateTo({
    url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}&section=node&baseMemberId=${encodeURIComponent(String(node.memberId))}&addType=${encodeURIComponent(addType)}&returnTo=tree`
  })
}

async function ensureMemberCapacity(node: TreeNode, addType: RelationshipAddType) {
  let capabilities: UserCapabilities
  try {
    capabilities = await fetchCapabilities()
  } catch {
    uni.showModal({
      title: '容量信息暂不可用',
      content: '暂时无法确认当前家庭容量，请稍后重试。',
      showCancel: false
    })
    return false
  }
  const usage = capabilities.usage.membersPerOwnedFamily[String(familyId.value)]
  const limit = capabilities.usage.membersLimitsByFamily?.[String(familyId.value)] ?? capabilities.limits.maxMembersPerOwnedFamily
  if (typeof usage !== 'number' || usage < limit) return true
  if (!session.user?.phoneVerified) {
    pendingAddNode.value = node
    pendingAddType.value = addType
    phoneUpgradeVisible.value = true
    return false
  }
  openCapacityPurchase(usage, limit)
  return false
}

function openCapacityPurchase(usage: number, limit: number) {
  uni.showActionSheet({ itemList: [`购买扩容（当前 ${usage}/${limit}）`, '邀请家人解锁奖励名额'], success: (result) => {
    const page = result.tapIndex === 0 ? 'capacity' : 'invitation-reward'
    uni.navigateTo({ url: `/pages/family/${page}?familyId=${encodeURIComponent(familyId.value)}&returnTo=tree` })
  } })
}

function continueAfterPhoneUpgrade() {
  const node = pendingAddNode.value
  const addType = pendingAddType.value
  pendingAddNode.value = null
  pendingAddType.value = null
  if (node && addType) void openAddRelation(node, addType)
}

function openAdjustRelation(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}&section=relations&baseMemberId=${encodeURIComponent(String(node.memberId))}&returnTo=tree`
  })
}

async function updateLivingState(node: TreeNode, isAlive: boolean) {
  try {
    await updateFamilyMember(familyId.value, node.memberId, { isAlive })
    uni.showToast({ title: isAlive ? '已标记健在' : '已标记已故', icon: 'success' })
    await loadTree()
  } catch (error) {
    uni.showModal({
      title: '更新失败',
      content: apiErrorMessage(error, '更新健在状态失败。'),
      showCancel: false
    })
  }
}

function confirmDeleteMember(node: TreeNode) {
  const restriction = deleteRestriction(node)
  if (restriction) {
    uni.showModal({ title: '暂不能删除', content: restriction, showCancel: false })
    return
  }
  uni.showModal({
    title: '删除节点',
    content: `确定删除“${node.displayName || '该成员'}”吗？相关父母或配偶关系会一并解除。`,
    success: (result) => {
      if (result.confirm) deleteMember(node)
    }
  })
}

async function deleteMember(node: TreeNode) {
  try {
    await deleteFamilyMember(familyId.value, node.memberId, '小程序家庭树节点删除')
    uni.showToast({ title: '节点已删除', icon: 'success' })
    await loadTree()
  } catch (error) {
    uni.showModal({
      title: '删除失败',
      content: apiErrorMessage(error, '删除成员失败。'),
      showCancel: false
    })
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  returnTo.value = String(options?.returnTo || '')
  guideMode.value = String(options?.guide || '') === '1'
  if (guideMode.value) markNewUserGuideMilestone('TREE_GUIDE_OPENED')
  const memberId = Number(options?.memberId)
  requestedViewerMemberId.value = Number.isFinite(memberId) && memberId > 0 ? memberId : null
  requestedRelationAddType.value = String(options?.task || '') === 'ADD_SPOUSE' ? 'ADD_SPOUSE' : null
  requestedNodeAction.value = String(options?.task || '') === 'INVITE' ? 'INVITE' : null
})
onShow(loadTree)
onUnload(resetPageData)
</script>

<style scoped>
.ui-simplified-hidden {
  display: none !important;
}

.private-tree-page {
  padding-top: 28rpx;
}

.private-tree-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.tree-error,
.tree-loading {
  padding: 28rpx 0;
}

.tree-error :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.tree-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24rpx;
  margin-bottom: 24rpx;
}

.tree-head-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 12rpx;
}

.tree-export-action {
  padding: 10rpx 14rpx;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 5rpx;
  background: rgba(255, 250, 239, 0.72);
  color: var(--archive-blue);
  font-size: 21rpx;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

.tree-export-action:active,
.tree-export-action.loading {
  background: var(--archive-paper-deep);
  opacity: 0.72;
}

.tree-export-canvas {
  position: fixed;
  top: -10000px;
  left: -10000px;
  pointer-events: none;
  opacity: 0;
}

.family-tree-tabs {
  margin-bottom: 20rpx;
}

.viewer-panel {
  display: flex;
  align-items: center;
  margin-bottom: 20rpx;
  padding: 18rpx;
  border: 1rpx solid rgba(22, 51, 83, 0.16);
  border-left: 5rpx solid var(--archive-blue);
  border-radius: 8rpx;
  background: rgba(255, 251, 242, 0.54);
}

.viewer-copy {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: 5rpx;
}

.viewer-label {
  color: var(--archive-cinnabar);
  font-size: 20rpx;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.viewer-name {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 31rpx;
  font-weight: 750;
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.viewer-hint {
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.45;
}
.viewer-actions {
  display: flex;
  flex-shrink: 0;
  flex-direction: column;
  gap: 10rpx;
  margin-left: 18rpx;
}
.viewer-action {
  padding: 10rpx 14rpx;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 6rpx;
  color: var(--archive-blue);
  background: rgba(255, 250, 239, 0.72);
  font-size: 21rpx;
  line-height: 1.2;
  white-space: nowrap;
}
.viewer-action:active {
  background: var(--archive-paper-deep);
  opacity: 0.76;
}

.tree-health-note {
  display: flex;
  align-items: flex-start;
  gap: 10rpx;
  margin: 0 0 20rpx;
  padding: 12rpx 14rpx;
  border-left: 3rpx solid var(--archive-cinnabar);
  background: rgba(168, 59, 45, 0.045);
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.55;
}

.tree-affairs {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  margin: -2rpx 0 20rpx;
}

.tree-affairs:active {
  background: rgba(31, 58, 95, 0.045);
}

.tree-affairs-copy {
  min-width: 0;
}

.tree-affairs-title,
.tree-affairs-desc {
  display: block;
}

.tree-affairs-title {
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 800;
}

.tree-affairs-desc {
  margin-top: 6rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.5;
}

.tree-affairs-meta {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 12rpx;
}

.tree-affairs-count {
  min-width: 38rpx;
  box-sizing: border-box;
  padding: 4rpx 10rpx;
  border-radius: 999rpx;
  background: var(--archive-cinnabar);
  color: #fff;
  font-size: 21rpx;
  font-weight: 800;
  line-height: 1.3;
  text-align: center;
}

.tree-health-dot {
  display: grid;
  flex: 0 0 auto;
  width: 22rpx;
  height: 22rpx;
  margin-top: 4rpx;
  place-items: center;
  border-radius: 50%;
  background: var(--archive-cinnabar);
  color: #fff;
  font-size: 16rpx;
  font-weight: 800;
  line-height: 1;
}

.tree-status-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx 20rpx;
  margin: -6rpx 0 20rpx;
  padding: 0 2rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.35;
}

.tree-guide-note {
  margin: 16rpx 0 20rpx;
}
.tree-status-item {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  white-space: nowrap;
}
.legend-sample {
  display: inline-grid;
  width: 28rpx;
  height: 28rpx;
  place-items: center;
  box-sizing: border-box;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 4rpx;
  background: var(--archive-paper, #fff7e8);
  color: var(--archive-blue);
  font-size: 15rpx;
  font-weight: 700;
  line-height: 1;
}
.legend-sample.health {
  width: 22rpx;
  height: 22rpx;
  border: 0;
  border-radius: 50%;
  background: var(--archive-cinnabar);
  color: #fff;
  font-size: 14rpx;
}
.legend-sample.deceased {
  border-color: rgba(101, 112, 128, 0.32);
  background: #eee7d8;
  color: rgba(36, 50, 68, 0.58);
}
.legend-sample.binding {
  border-color: rgba(42, 112, 87, 0.28);
  background: rgba(42, 112, 87, 0.09);
  color: var(--archive-green, #2a7057);
  font-size: 17rpx;
}
.legend-sample.pending {
  border-color: rgba(168, 59, 45, 0.28);
  background: rgba(168, 59, 45, 0.07);
  color: var(--archive-cinnabar);
  font-size: 17rpx;
}

.genealogy-tools {
  display: flex;
  gap: 14rpx;
  margin-bottom: 22rpx;
  border-top: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.genealogy-tool-link {
  display: block;
  min-width: 180rpx;
  min-height: 62rpx;
  margin: 0;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 0;
  background: rgba(255, 248, 234, 0.5);
  color: var(--archive-blue);
  padding: 8rpx 10rpx;
  font-size: 23rpx;
  font-weight: 700;
  line-height: 1.35;
}

.genealogy-tool-link::after {
  border: 0;
}

.genealogy-tool-link[disabled] {
  opacity: 0.58;
}


.tree-mode-line {
  display: flex;
  gap: 22rpx;
  margin-bottom: 18rpx;
}

.tree-mode-item {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  padding-bottom: 8rpx;
}

.tree-mode-item.active {
  color: var(--archive-blue);
  font-weight: 750;
  border-bottom: 2rpx solid var(--archive-blue);
}

.tree-scroll-note {
  margin-bottom: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.55;
}

.archive-tree-structure {
  margin-left: -8rpx;
  margin-right: -8rpx;
}

.node-action-panel {
  position: fixed;
  right: 20rpx;
  bottom: calc(20rpx + env(safe-area-inset-bottom));
  left: 20rpx;
  z-index: 98;
  max-height: 720rpx;
  margin: 0;
  overflow-y: auto;
  padding: 20rpx;
  border: 1rpx solid rgba(22, 51, 83, 0.22);
  border-left: 5rpx solid var(--archive-blue);
  border-radius: 8rpx;
  background: rgba(255, 251, 242, 0.84);
  box-shadow: 0 10rpx 34rpx rgba(27, 42, 57, 0.24);
}

.node-action-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.node-action-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 4rpx;
}

.node-action-label {
  color: var(--archive-cinnabar);
  font-size: 20rpx;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.node-action-name {
  overflow: hidden;
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 31rpx;
  font-weight: 750;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-action-close {
  flex: 0 0 auto;
  padding: 8rpx 4rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.2;
}

.node-action-head-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 18rpx;
}

.node-action-back {
  padding: 8rpx 4rpx;
  color: var(--archive-blue);
  font-size: 22rpx;
  font-weight: 700;
  line-height: 1.2;
}

.node-action-hint {
  display: block;
  margin: 12rpx 0 14rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.4;
}

.node-action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12rpx;
}

.node-action-item {
  display: flex;
  min-height: 66rpx;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  padding: 10rpx 12rpx;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 5rpx;
  color: var(--archive-blue);
  background: rgba(255, 249, 238, 0.72);
  font-size: 22rpx;
  font-weight: 700;
  line-height: 1.25;
  text-align: center;
}

.node-action-item.primary {
  border-color: rgba(31, 79, 121, 0.48);
  background: rgba(31, 79, 121, 0.08);
}

.node-action-item.danger {
  border-color: rgba(168, 59, 45, 0.42);
  color: var(--archive-cinnabar);
}

.node-action-item:active,
.node-action-close:active,
.node-action-back:active {
  opacity: 0.64;
}

.empty-tree-guide {
  padding: 28rpx 0;
}

.empty-tree-action {
  margin-top: 18rpx;
}

.relation-list {
  display: block;
  border-top: 1rpx solid var(--archive-line);
}

.tree-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx 22rpx;
  margin-top: 22rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 16rpx;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.legend-name {
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 24rpx;
  font-weight: 600;
  line-height: 1;
}

.legend-name.male {
  color: var(--archive-blue);
}

.legend-name.female {
  color: var(--archive-cinnabar);
}

.legend-name.unknown {
  color: var(--archive-ink-soft);
}

.legend-tag {
  width: 22rpx;
  height: 34rpx;
  border: 1rpx solid var(--archive-line-strong);
  background: #fff7e8;
}

.legend-tag.lineage {
  border-style: solid;
  background: #fff7e8;
}

.legend-tag.spouse {
  border-style: dashed;
  background: #fbf0dc;
}

.legend-tag.current {
  border-color: var(--archive-blue);
  background: var(--archive-blue);
}

.memorial-mask {
  position: fixed;
  inset: 0;
  z-index: 99;
  display: flex;
  align-items: flex-end;
  background: rgba(27, 42, 57, 0.42);
}

.memorial-sheet {
  box-sizing: border-box;
  width: 100%;
  min-height: 560rpx;
  border-top: 1rpx solid rgba(31, 55, 79, 0.32);
  background:
    repeating-linear-gradient(
      0deg,
      rgba(153, 122, 73, 0.045) 0,
      rgba(153, 122, 73, 0.045) 1rpx,
      transparent 1rpx,
      transparent 42rpx
    ),
    var(--archive-paper);
  padding: 46rpx 42rpx calc(42rpx + env(safe-area-inset-bottom));
  box-shadow: 0 -16rpx 38rpx rgba(27, 42, 57, 0.16);
}

.memorial-sheet-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding-bottom: 26rpx;
}

.memorial-kicker,
.memorial-name,
.memorial-time-label,
.memorial-time-value,
.memorial-content-label,
.memorial-content-value,
.memorial-footnote {
  display: block;
}

.memorial-kicker {
  color: var(--archive-cinnabar);
  font-size: 19rpx;
  font-weight: 700;
  letter-spacing: 0.13em;
}

.memorial-name {
  margin-top: 8rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 43rpx;
  font-weight: 700;
  line-height: 1.3;
}

.memorial-seal {
  display: flex;
  width: 54rpx;
  height: 54rpx;
  align-items: center;
  justify-content: center;
  border: 2rpx solid rgba(168, 59, 45, 0.68);
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  transform: rotate(-5deg);
}

.memorial-time {
  display: flex;
  align-items: baseline;
  gap: 20rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding: 22rpx 0;
}

.memorial-time-label,
.memorial-content-label {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  letter-spacing: 0.08em;
}

.memorial-time-value {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 700;
}

.memorial-content {
  padding: 28rpx 0 34rpx;
}

.memorial-content-value {
  min-height: 132rpx;
  margin-top: 14rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 27rpx;
  line-height: 1.9;
  white-space: pre-wrap;
}

.memorial-foot {
  border-top: 1rpx solid var(--archive-line);
  padding-top: 22rpx;
}

.memorial-footnote {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.55;
}

.memorial-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 16rpx;
  margin-top: 28rpx;
}

.memorial-action {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 126rpx;
  min-height: 80rpx;
  border: 1rpx solid var(--archive-line-strong);
  box-sizing: border-box;
  padding: 12rpx 18rpx;
  text-align: center;
  font-size: 24rpx;
  font-weight: 700;
}

.memorial-action.secondary {
  color: var(--archive-ink-soft);
}

.memorial-action.primary {
  border-color: var(--archive-cinnabar);
  color: var(--archive-cinnabar);
}

.memorial-delete-restriction {
  display: block;
  margin-top: 20rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.55;
  text-align: right;
}

.memorial-delete-action {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 76rpx;
  margin-top: 16rpx;
  border-top: 1rpx solid rgba(168, 59, 45, 0.24);
  color: var(--archive-cinnabar);
  font-size: 24rpx;
  font-weight: 700;
  text-align: center;
}
</style>
