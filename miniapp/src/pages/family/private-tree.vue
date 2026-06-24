<template>
  <view class="tree-page private-tree-page">
    <MiniBackHome />
    <MiniCard v-if="errorMessage">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadTree">重新加载</MiniButton>
    </MiniCard>

    <MiniCard v-if="loading">
      <MiniEmptyState symbol="…" title="正在加载" description="正在组装家庭树..." />
    </MiniCard>

    <template v-else-if="tree">
      <view class="tree-tool-banner tree-banner">
        <view class="banner-copy">
          <text class="tree-tool-banner-title">私有家谱</text>
          <text class="tree-tool-banner-desc">
            {{ treeViewLabel(tree.treeMode) }} · {{ tree.nodes.length }} 位成员 · {{ edgeCount }} 条关系
          </text>
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
          <text class="tree-space-subtitle">夫妻单元向下展开，可滑动查看</text>
        </view>
        <view class="tree-space-body section-pad">
          <TreeViewModeSwitch v-model="viewMode" />

          <template v-if="viewMode === 'structure'">
            <MiniEmptyState
              v-if="tree.nodes.length === 0"
              symbol="谱"
              title="暂无成员"
              description="请先在家庭成员中添加成员后再查看家谱。"
            />
            <FamilyTreeStructureView
              v-else
              :tree="tree"
              :viewer-member-id="viewerMemberId"
              show-binding
            />
          </template>

          <template v-else>
            <MiniEmptyState
              v-if="edgeCount === 0"
              symbol="亲"
              title="暂无关系"
              description="暂无父母子女或配偶关系。"
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
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listFamilyMembers } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import TreeViewModeSwitch from '@/components/family/TreeViewModeSwitch.vue'
import { resolveCurrentMemberId } from '@/features/family-tree/resolveCurrentMember'
import {
  buildRelationSentences,
  countVisibleEdges,
  treeViewLabel
} from '@/features/family-tree/relationSentences'
import type { FamilyTreeViewMode } from '@/features/family-tree/types'
import { useSessionStore } from '@/stores/session'
import type { FamilyTreeResult } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const viewMode = ref<FamilyTreeViewMode>('structure')
const viewerMemberId = ref<number | null>(null)

const edgeCount = computed(() => (tree.value ? countVisibleEdges(tree.value) : 0))
const relationSentences = computed(() => (tree.value ? buildRelationSentences(tree.value) : []))

async function loadTree() {
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  errorMessage.value = ''
  viewerMemberId.value = null
  try {
    tree.value = await getPrivateTree(familyId.value)
    if (session.user?.id) {
      try {
        const members = await listFamilyMembers(familyId.value)
        viewerMemberId.value = resolveCurrentMemberId(session.user.id, members)
      } catch {
        // Tree nodes lack userId; members list is the supported way to locate "me".
      }
    }
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭树加载失败。')
  } finally {
    loading.value = false
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  loadTree()
})
</script>

<style scoped>
.tree-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.banner-copy {
  flex: 1;
  min-width: 0;
}

.section-pad {
  padding: 8rpx 24rpx 16rpx;
}
</style>
