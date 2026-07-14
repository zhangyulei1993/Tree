<template>
  <view class="archive-page private-tree-page">
    <MiniBackHome />

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
            {{ family?.familyName || '家庭关系' }} · 家庭树 · {{ tree.nodes.length }} 位成员
        </text>
        </view>
        <view v-if="family" class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="archive-segment-tabs family-tree-tabs">
        <view class="archive-segment-tab active">
          <text>家庭树</text>
        </view>
        <view class="archive-segment-tab" @click="openMembers">
          <text>成员表</text>
        </view>
      </view>

      <view v-if="tree.nodes.length > 0" class="viewer-panel archive-form-panel">
        <view class="viewer-copy">
          <text class="viewer-label">当前视角</text>
          <text class="viewer-name">{{ currentViewerName }}</text>
          <text class="viewer-hint">仅用于查看称谓，不影响家庭树关系。</text>
        </view>
        <picker :range="viewerPickerLabels" :value="viewerPickerIndex" @change="changeViewer">
          <view class="viewer-action">更换</view>
        </picker>
      </view>

      <view v-if="canManageFamily" class="genealogy-tools">
        <button v-if="canManageFamily" class="genealogy-tool-link" @click="openManage">关系管理</button>
      </view>

      <view class="tree-mode-line">
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

      <view class="tree-scroll-note">
        <text>关系图可左右滑动查看；配偶为浅纸签，不作为关系扩展入口。</text>
      </view>

      <template v-if="viewMode === 'structure'">
        <MiniEmptyState
          v-if="tree.nodes.length === 0"
          symbol="树"
          title="暂无成员"
          description="请先在家庭树成员中添加成员后再查看关系图。"
        />
        <FamilyTreeStructureView
          v-else
          class="archive-tree-structure"
          :tree="tree"
          :viewer-member-id="viewerMemberId"
          show-binding
          :interactive="canManageFamily"
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

      <view class="tree-legend">
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
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { deleteFamilyMember, listFamilyMembers, updateFamilyMember } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import { resolveCurrentMemberId } from '@/features/family-tree/resolveCurrentMember'
import { buildRelationSentences, countVisibleEdges } from '@/features/family-tree/relationSentences'
import type { FamilyTreeViewMode } from '@/features/family-tree/types'
import { isLineageMember, isSpouseMember, isExternalMember } from '@/features/family-tree/graph'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyTreeResult, TreeNode } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const viewMode = ref<FamilyTreeViewMode>('structure')
const viewerMemberId = ref<number | null>(null)
const currentUserMemberId = ref<number | null>(null)
const edgeCount = computed(() => (tree.value ? countVisibleEdges(tree.value) : 0))
const relationSentences = computed(() => (tree.value ? buildRelationSentences(tree.value) : []))
const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const viewerNodes = computed(() => tree.value?.nodes || [])
const currentViewerNode = computed(() =>
  viewerNodes.value.find((node) => node.memberId === viewerMemberId.value) || null
)
const currentViewerName = computed(() => currentViewerNode.value?.displayName || '未选择')
const viewerPickerLabels = computed(() =>
  viewerNodes.value.map((node) => {
    const suffix = node.memberId === currentUserMemberId.value ? '（我）' : ''
    return `${node.displayName || '未命名成员'}${suffix}`
  })
)
const viewerPickerIndex = computed(() => {
  const index = viewerNodes.value.findIndex((node) => node.memberId === viewerMemberId.value)
  return index >= 0 ? index : 0
})

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
  family.value = null
  viewerMemberId.value = null
  currentUserMemberId.value = null
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
    const [familyResult, treeResult] = await Promise.all([
      getFamilyDetail(familyId.value),
      getPrivateTree(familyId.value)
    ])
    family.value = familyResult
    tree.value = treeResult
    let resolvedViewerId = currentUserMemberId.value
    if (session.user?.id) {
      try {
        const memberList = await listFamilyMembers(familyId.value)
        resolvedViewerId = resolveCurrentMemberId(session.user.id, memberList)
        currentUserMemberId.value = resolvedViewerId
      } catch {
        // Tree nodes lack userId; members list is the supported way to locate "me".
      }
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

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function openMembers() {
  uni.redirectTo({ url: `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}` })
}

function openManage() {
  uni.navigateTo({
    url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function changeViewer(event: { detail: { value: number | string } }) {
  const index = Number(event.detail.value)
  const node = viewerNodes.value[index]
  if (!node) return
  viewerMemberId.value = normalizeViewerMemberId(node.memberId)
  viewMode.value = 'structure'
}

function openNodeActions(node: TreeNode) {
  if (!canManageFamily.value) return
  const items = ['编辑成员资料']
  if (node.userBindingState === 'UNBOUND') items.push('邀请本人绑定')
  if (!isSpouseMember(node) && !isExternalMember(node)) items.push('关系操作')
  items.push('健在状态', '删除成员（谨慎）')
  uni.showActionSheet({
    itemList: items,
    success: (result) => {
      const action = items[result.tapIndex]
      if (action === '编辑成员资料') {
        openMemberEdit(node)
        return
      }
      if (action === '邀请本人绑定') {
        openInviteMember(node)
        return
      }
      if (action === '关系操作') {
        openRelationActions(node)
        return
      }
      if (action === '健在状态') {
        openLivingStateActions(node)
        return
      }
      if (action === '删除成员（谨慎）') {
        confirmDeleteMember(node)
      }
    }
  })
}

function openRelationActions(node: TreeNode) {
  const items = ['添加关系', '调整关系']
  uni.showActionSheet({
    itemList: items,
    success: (result) => {
      const action = items[result.tapIndex]
      if (action === '添加关系') openAddRelation(node)
      if (action === '调整关系') openAdjustRelation(node)
    }
  })
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

function openMemberEdit(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/family/member-edit?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(String(node.memberId))}`
  })
}

function openInviteMember(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}&mode=node&memberId=${encodeURIComponent(String(node.memberId))}&memberName=${encodeURIComponent(node.displayName || '')}`
  })
}

function openAddRelation(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}&section=node&baseMemberId=${encodeURIComponent(String(node.memberId))}`
  })
}

function openAdjustRelation(node: TreeNode) {
  uni.navigateTo({
    url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}&section=relations&baseMemberId=${encodeURIComponent(String(node.memberId))}`
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
  uni.showModal({
    title: '删除成员',
    content: `确定删除“${node.displayName || '该成员'}”吗？没有子女时会一并解除父母/配偶关系；已有子女时系统会拒绝。`,
    success: (result) => {
      if (result.confirm) deleteMember(node)
    }
  })
}

async function deleteMember(node: TreeNode) {
  try {
    await deleteFamilyMember(familyId.value, node.memberId, '小程序家庭树节点删除')
    uni.showToast({ title: '成员已删除', icon: 'success' })
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
})
onShow(loadTree)
onUnload(resetPageData)
</script>

<style scoped>
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

.family-tree-tabs {
  margin-bottom: 20rpx;
}

.viewer-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  margin-bottom: 20rpx;
  padding: 18rpx 0;
  border-left: 0;
  border-right: 0;
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
}

.viewer-hint {
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.45;
}

.viewer-action {
  min-width: 96rpx;
  border: 1rpx solid rgba(168, 59, 45, 0.62);
  color: var(--archive-cinnabar);
  padding: 12rpx 18rpx;
  text-align: center;
  font-size: 23rpx;
  font-weight: 700;
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
</style>
