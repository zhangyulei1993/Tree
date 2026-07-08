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
          <text class="archive-kicker">Genealogy</text>
          <text class="archive-title">家谱</text>
          <text class="archive-subtitle">
            {{ family?.familyName || '家谱结构' }} · {{ treeViewLabel(tree.treeMode) }} · {{ tree.nodes.length }} 位成员
          </text>
        </view>
        <view v-if="family" class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="generation-tabs">
        <view
          v-for="item in generationTabs"
          :key="item"
          class="generation-tab"
          :class="{ active: activeGeneration === item }"
          @click="activeGeneration = item"
        >
          <text>{{ item }}</text>
        </view>
      </view>

      <view v-if="canManageFamily" class="genealogy-tools">
        <text class="genealogy-tools-label">家谱管理</text>
        <view class="genealogy-tools-links">
          <button class="genealogy-tool-link" @click="openManageSection('create')">添加亲属</button>
          <button class="genealogy-tool-link" @click="openManageSection('locate')">定位成员</button>
          <button class="genealogy-tool-link" @click="openManageSection('relations')">关系维护</button>
        </view>
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
        <text>纸签家谱可左右滑动查看；配偶为浅纸签，不作为主干展开入口。</text>
      </view>

      <template v-if="viewMode === 'structure'">
        <MiniEmptyState
          v-if="tree.nodes.length === 0"
          symbol="谱"
          title="暂无成员"
          description="请先在家庭成员中添加成员后再查看家谱。"
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
          <text>族内成员</text>
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
import { listFamilyMembers } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import { resolveCurrentMemberId } from '@/features/family-tree/resolveCurrentMember'
import {
  buildRelationSentences,
  countVisibleEdges,
  treeViewLabel
} from '@/features/family-tree/relationSentences'
import type { FamilyTreeViewMode } from '@/features/family-tree/types'
import { isSpouseMember, isExternalMember } from '@/features/family-tree/graph'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyTreeResult, RelationshipAddType, TreeNode } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const viewMode = ref<FamilyTreeViewMode>('structure')
const viewerMemberId = ref<number | null>(null)
const generationTabs = ['一世', '二世', '三世', '四世', '五世']
const activeGeneration = ref('一世')

const edgeCount = computed(() => (tree.value ? countVisibleEdges(tree.value) : 0))
const relationSentences = computed(() => (tree.value ? buildRelationSentences(tree.value) : []))
const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

const nodeActionTypes: RelationshipAddType[] = [
  'ADD_FATHER',
  'ADD_MOTHER',
  'ADD_CHILD',
  'ADD_SPOUSE',
  'ADD_SIBLING'
]

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  tree.value = null
  family.value = null
  viewerMemberId.value = null
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
    if (session.user?.id) {
      try {
        const memberList = await listFamilyMembers(familyId.value)
        viewerMemberId.value = resolveCurrentMemberId(session.user.id, memberList)
      } catch {
        // Tree nodes lack userId; members list is the supported way to locate "me".
      }
    }
    if (!viewerMemberId.value) {
      viewerMemberId.value = treeResult.nodes[0]?.memberId ?? null
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

function openManageSection(section: 'create' | 'locate' | 'relations') {
  uni.navigateTo({
    url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}&section=${section}`
  })
}

function openNodeActions(node: TreeNode) {
  if (!canManageFamily.value) return
  if (isSpouseMember(node) || isExternalMember(node)) {
    uni.navigateTo({
      url: `/pages/family/member-edit?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(String(node.memberId))}`
    })
    return
  }
  uni.showActionSheet({
    itemList: ['编辑成员资料', '添加父亲', '添加母亲', '添加子女', '添加配偶', '添加兄弟姐妹'],
    success: (result) => {
      if (result.tapIndex === 0) {
        uni.navigateTo({
          url: `/pages/family/member-edit?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(String(node.memberId))}`
        })
        return
      }
      const addType = nodeActionTypes[result.tapIndex - 1]
      if (!addType) return
      uni.navigateTo({
        url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}&section=create&baseMemberId=${encodeURIComponent(String(node.memberId))}&addType=${encodeURIComponent(addType)}`
      })
    }
  })
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

.generation-tabs {
  display: flex;
  align-items: center;
  gap: 26rpx;
  margin-bottom: 24rpx;
  border-bottom: 1rpx solid var(--archive-line);
}

.generation-tab {
  position: relative;
  padding: 0 0 16rpx;
  color: var(--archive-ink-soft);
  font-size: 25rpx;
  line-height: 1.4;
}

.generation-tab.active {
  color: var(--archive-cinnabar);
  font-weight: 750;
}

.generation-tab.active::after {
  content: '';
  position: absolute;
  right: 0;
  bottom: -1rpx;
  left: 0;
  height: 3rpx;
  background: var(--archive-cinnabar);
}

.genealogy-tools {
  margin-bottom: 22rpx;
  border-top: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.genealogy-tools-label {
  display: block;
  margin-bottom: 12rpx;
  color: var(--archive-cinnabar);
  font-size: 21rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.genealogy-tools-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10rpx;
}

.genealogy-tool-link {
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
