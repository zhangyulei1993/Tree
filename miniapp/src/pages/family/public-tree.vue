<template>
  <view class="tree-page">
    <MiniBackHome />
    <MiniCard v-if="loading">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载公开家谱..." />
    </MiniCard>

    <MiniCard v-else-if="errorMessage">
      <MiniSectionHeader title="公开家谱" />
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadTree">重新加载</MiniButton>
    </MiniCard>

    <template v-else-if="tree">
      <view class="tree-tool-banner genealogy-banner">
        <view class="banner-copy">
          <text class="tree-tool-banner-title">公开家谱</text>
          <text class="tree-tool-banner-desc">{{ treeViewLabel(tree.treeMode) }}</text>
          <view class="tree-archive-ribbon">
            <text class="tree-archive-chip">成员 {{ tree.nodes.length }}</text>
            <text class="tree-archive-chip green">关系 {{ edgeCount }}</text>
          </view>
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

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">家谱结构</text>
          <text class="tree-space-subtitle">公开可见的家谱树结构</text>
        </view>
        <view class="tree-space-body section-pad">
          <TreeViewModeSwitch v-model="viewMode" />

          <template v-if="viewMode === 'structure'">
            <MiniEmptyState
              v-if="tree.nodes.length === 0"
              symbol="谱"
              title="暂无公开成员"
              description="该家庭暂未公开家谱成员。"
            />
            <FamilyTreeStructureView v-else :tree="tree" />
          </template>

          <template v-else>
            <MiniEmptyState
              v-if="edgeCount === 0"
              symbol="亲"
              title="暂无公开关系"
              description="该家庭暂未公开家谱关系。"
            />
            <RelationSentenceList v-else :items="relationSentences" />
          </template>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import TreeViewModeSwitch from '@/components/family/TreeViewModeSwitch.vue'
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
.genealogy-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.banner-copy {
  flex: 1;
  min-width: 0;
}

.section-pad {
  padding: 8rpx 20rpx 16rpx;
}
</style>
