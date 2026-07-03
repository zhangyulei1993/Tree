<template>
  <view class="tree-page private-tree-page">
    <MiniBackHome />
    <MiniCard v-if="errorMessage && !tree">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadTree">重新加载</MiniButton>
    </MiniCard>

    <MiniCard v-else-if="loading && !tree">
      <MiniEmptyState symbol="…" title="正在加载" description="正在组装家庭树..." />
    </MiniCard>

    <template v-else-if="tree">
      <FamilyContextHeader
        v-if="family"
        :family-name="family.familyName"
        section="家谱"
        :subtitle="canManageFamily ? '查看家谱结构，维护亲属关系' : '查看家谱结构与亲属关系描述'"
        :role-label="family.role === 'FOUNDER' ? '创建者' : family.role === 'FAMILY_ADMIN' ? '管理员' : '成员'"
        @back="openFamilyOverview"
      />
      <view class="tree-tool-banner tree-banner">
        <view class="banner-copy">
          <text class="tree-tool-banner-title">家谱概览</text>
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

      <view v-if="canManageFamily" class="genealogy-tools">
        <text class="genealogy-tools-label">家谱管理</text>
        <view class="genealogy-tools-links">
          <button class="genealogy-tool-link" @click="openManageSection('create')">添加亲属</button>
          <button class="genealogy-tool-link" @click="openManageSection('locate')">定位成员</button>
          <button class="genealogy-tool-link" @click="openManageSection('relations')">关系维护</button>
        </view>
      </view>

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">家谱结构</text>
          <text class="tree-space-subtitle">夫妻单元向下展开，可滑动查看</text>
        </view>
        <view class="tree-space-body section-pad">
          <TreeViewModeSwitch v-model="viewMode" />
          <MiniNotice v-if="canManageFamily && viewMode === 'structure'" tone="security" title="可直接管理家谱节点">
            带“管理”标记的族内成员节点可点击，添加父母、子女、配偶和兄弟姐妹；配偶节点仅可编辑资料。
          </MiniNotice>

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
            <RelationSentenceList v-else :items="relationSentences" />
          </template>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyContextHeader from '@/components/family/FamilyContextHeader.vue'
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
import { isLineageMember, isSpouseMember, isExternalMember } from '@/features/family-tree/graph'
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
.tree-banner {
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

.tree-banner .tree-tool-banner-title {
  color: #fff;
}

.tree-banner .tree-tool-banner-desc {
  color: rgba(255, 255, 255, 0.72);
}

.section-pad {
  padding: 12rpx 20rpx 20rpx;
}

.private-tree-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.genealogy-tools {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12rpx 16rpx;
  margin-bottom: 20rpx;
  border: 1rpx solid rgba(47, 107, 87, 0.12);
  border-radius: 20rpx;
  background: rgba(255, 255, 255, 0.72);
  padding: 16rpx 20rpx;
}

.genealogy-tools-label {
  color: var(--tree-text-secondary, #64748b);
  font-size: 22rpx;
  font-weight: 600;
}

.genealogy-tools-links {
  display: grid;
  flex: 1;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8rpx;
}

.genealogy-tool-link {
  min-height: 64rpx;
  margin: 0;
  border: 1rpx solid rgba(47, 107, 87, 0.12);
  border-radius: 14rpx;
  background: rgba(240, 248, 244, 0.9);
  color: var(--tree-green, #2f6b57);
  padding: 8rpx 10rpx;
  font-size: 24rpx;
  font-weight: 700;
  line-height: 1.35;
}

.genealogy-tool-link::after {
  border: 0;
}

.genealogy-tool-link:active {
  opacity: 0.72;
}

</style>
