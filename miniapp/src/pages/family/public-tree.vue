<template>
  <view class="archive-page public-tree-page">
    <MiniBackHome />

    <view v-if="loading && !tree" class="tree-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载公开家谱..." />
    </view>

    <view v-else-if="errorMessage && !tree" class="tree-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" class="retry-button" @click="loadTree">重新加载</MiniButton>
    </view>

    <template v-else-if="tree">
      <view class="tree-head archive-page-head">
        <view>
          <text class="archive-kicker">Public Genealogy</text>
          <text class="archive-title">公开家谱</text>
          <text class="archive-subtitle">
            {{ treeViewLabel(tree.treeMode) }} · {{ tree.nodes.length }} 位成员 · {{ edgeCount }} 条关系
          </text>
        </view>
        <view class="archive-seal">{{ surnameSeal }}</view>
      </view>

      <view class="public-notice archive-panel">
        <text class="public-notice-label">公开册页</text>
        <text class="public-notice-text">
          以下家谱经平台审核后对外展示，姓名等信息已按公开规则脱敏，仅供浏览，不提供申请加入。
        </text>
      </view>

      <view class="public-stats">
        <text class="archive-chip">成员 {{ tree.nodes.length }}</text>
        <text class="archive-chip">关系 {{ edgeCount }}</text>
        <text class="archive-chip">脱敏展示</text>
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
        <text>纸签家谱可左右滑动查看；配偶为浅纸签，不作为关系扩展入口。</text>
      </view>

      <template v-if="viewMode === 'structure'">
        <MiniEmptyState
          v-if="tree.nodes.length === 0"
          symbol="谱"
          title="暂无公开成员"
          description="该家庭暂未公开家谱成员。"
        />
        <FamilyTreeStructureView v-else class="archive-tree-structure" :tree="tree" />
      </template>

      <template v-else>
        <MiniEmptyState
          v-if="edgeCount === 0"
          symbol="亲"
          title="暂无公开关系"
          description="该家庭暂未公开家谱关系。"
        />
        <RelationSentenceList v-else class="public-relation-list" :items="relationSentences" />
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
          <text class="legend-tag spouse" />
          <text>配偶</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, onMounted, onUnmounted, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPublicTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import {
  buildRelationSentences,
  countVisibleEdges,
  treeViewLabel
} from '@/features/family-tree/relationSentences'
import type { FamilyTreeViewMode } from '@/features/family-tree/types'
import type { FamilyTreeResult } from '@/types/api'

const familyId = ref('')
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const viewMode = ref<FamilyTreeViewMode>('structure')
let requestVersion = 0

const edgeCount = computed(() => (tree.value ? countVisibleEdges(tree.value) : 0))
const relationSentences = computed(() => (tree.value ? buildRelationSentences(tree.value) : []))
const surnameSeal = computed(() => {
  const surname = tree.value?.nodes.find((node) => node.surname)?.surname
  return surname?.slice(0, 1) || '谱'
})

function resetPageState() {
  tree.value = null
  errorMessage.value = ''
  loading.value = false
}

function familyIdFromCurrentHash() {
  if (typeof window === 'undefined') return ''
  const [, query = ''] = window.location.hash.split('?')
  return new URLSearchParams(query).get('familyId') || ''
}

function isCurrentTreeRoute() {
  if (typeof window === 'undefined') return false
  return window.location.hash.split('?')[0] === '#/pages/family/public-tree'
}

function reloadForFamilyId(nextFamilyId: string) {
  requestVersion += 1
  familyId.value = nextFamilyId.trim()
  resetPageState()

  if (!familyId.value) {
    errorMessage.value = '公开家谱暂不可访问。'
    return
  }

  loadTree()
}

function handleHashChange() {
  if (!isCurrentTreeRoute()) return
  const nextFamilyId = familyIdFromCurrentHash()
  if (nextFamilyId !== familyId.value) {
    reloadForFamilyId(nextFamilyId)
  }
}

async function loadTree() {
  if (!familyId.value) {
    tree.value = null
    errorMessage.value = '公开家谱暂不可访问。'
    return
  }
  const version = requestVersion
  const currentFamilyId = familyId.value
  tree.value = null
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await getPublicTree(currentFamilyId)
    if (version === requestVersion && currentFamilyId === familyId.value) {
      tree.value = result
    }
  } catch (error) {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      tree.value = null
      errorMessage.value = apiErrorMessage(error, '公开家谱暂不可访问。')
    }
  } finally {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      loading.value = false
    }
  }
}

onLoad((options) => {
  reloadForFamilyId(String(options?.familyId || ''))
})

onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener('hashchange', handleHashChange)
  }
})

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('hashchange', handleHashChange)
  }
})
</script>

<style scoped>
.public-tree-page {
  padding-top: 28rpx;
}

.public-tree-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.tree-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.retry-button {
  margin-top: 16rpx;
}

.tree-head {
  margin-bottom: 18rpx;
}

.public-notice {
  margin-bottom: 18rpx;
  padding: 22rpx 0;
}

.public-notice-label {
  display: block;
  margin-bottom: 8rpx;
  color: var(--archive-cinnabar);
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 4rpx;
}

.public-notice-text {
  display: block;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.65;
}

.public-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 22rpx;
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

.public-relation-list {
  display: block;
  border-top: 1rpx solid var(--archive-line);
}

.public-tree-page :deep(.public-relation-list .relation-item) {
  border-top-color: var(--archive-line);
}

.public-tree-page :deep(.public-relation-list .relation-dot) {
  background: var(--archive-cinnabar);
}

.public-tree-page :deep(.public-relation-list .sentence) {
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 27rpx;
  line-height: 1.68;
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
</style>
