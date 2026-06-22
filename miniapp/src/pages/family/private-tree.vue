<template>
  <view class="tree-page private-tree-page">
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
            {{ treeViewLabel(tree.treeMode) }} · {{ tree.nodes.length }} 位成员 · {{ visibleEdges.length }} 条关系
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
          <text class="tree-space-title">家族成员</text>
          <text class="tree-space-subtitle">成员及其账号绑定状态</text>
        </view>
        <view class="tree-space-body section-pad">
          <MiniEmptyState
            v-if="tree.nodes.length === 0"
            symbol="谱"
            title="暂无成员"
            description="请先在家庭成员中添加成员后再查看家谱。"
          />
          <MemberMiniCard
            v-for="node in tree.nodes"
            :key="node.memberId"
            :name="node.displayName"
            :gender-label="genderText(node.gender)"
            :life-info="livingText(node.isLiving)"
            :bind-label="bindingStateText(node.userBindingState)"
          />
        </view>
      </view>

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">亲属关系</text>
          <text class="tree-space-subtitle">父母子女与配偶关系</text>
        </view>
        <view class="tree-space-body section-pad">
          <MiniEmptyState
            v-if="visibleEdges.length === 0"
            symbol="亲"
            title="暂无关系"
            description="暂无父母子女或配偶关系。当前仅展示已维护的家庭关系。"
          />
          <RelationSentenceList v-else :items="relationSentences" />
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPrivateTree } from '@/api/tree'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MemberMiniCard from '@/components/family/MemberMiniCard.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyTreeResult } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const visibleEdges = computed(() =>
  (tree.value?.edges || []).filter((edge) => String(edge.relationshipType) !== 'SIBLING')
)
const nodeNameMap = computed(() => new Map((tree.value?.nodes || []).map((node) => [node.memberId, node.displayName])))
const nodeGenderMap = computed(() => new Map((tree.value?.nodes || []).map((node) => [node.memberId, node.gender])))
const relationSentences = computed(() =>
  visibleEdges.value.map((edge) => buildRelationSentence(edge))
)

async function loadTree() {
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  errorMessage.value = ''
  try {
    tree.value = await getPrivateTree(familyId.value)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭树加载失败。')
  } finally {
    loading.value = false
  }
}

function nodeName(memberId: number) {
  return nodeNameMap.value.get(memberId) || '未知成员'
}

function treeViewLabel(mode: string) {
  switch (mode) {
    case 'LIST_TREE':
      return '家谱列表'
    case 'GRAPH_TREE':
      return '家谱图谱'
    default:
      return '私有家谱'
  }
}

function buildRelationSentence(edge: FamilyTreeResult['edges'][number]) {
  const fromName = nodeName(edge.fromMemberId)
  const toName = nodeName(edge.toMemberId)
  if (edge.relationshipType === 'SPOUSE') {
    let sentence = `${fromName} 与 ${toName} 是配偶`
    if (edge.relationNote) sentence += `，${edge.relationNote}`
    return sentence
  }
  const fromGender = nodeGenderMap.value.get(edge.fromMemberId)
  let parentRole = '父母'
  if (fromGender === 'MALE') parentRole = '父亲'
  else if (fromGender === 'FEMALE') parentRole = '母亲'
  let sentence = `${fromName} 是 ${toName} 的${parentRole}`
  const linkLabel = parentLinkText(edge.parentLinkType)
  if (linkLabel && linkLabel !== '父母子女') {
    sentence += `（${linkLabel}）`
  }
  if (edge.relationNote) sentence += `，${edge.relationNote}`
  return sentence
}

function parentLinkText(type?: string | null) {
  switch (type) {
    case 'PRIMARY':
      return '亲生'
    case 'STEP':
      return '继亲'
    case 'ADOPTIVE':
      return '收养'
    case 'SUCCESSION':
      return '过继'
    case 'NOTE_ONLY':
      return '备注'
    case 'OTHER':
      return '其他'
    default:
      return '父母子女'
  }
}

function genderText(gender: string) {
  switch (gender) {
    case 'MALE':
      return '男'
    case 'FEMALE':
      return '女'
    default:
      return '未知性别'
  }
}

function livingText(isLiving?: boolean | null) {
  if (isLiving === false) return '已故'
  if (isLiving === true) return '健在'
  return '生卒未知'
}

function bindingStateText(state: string) {
  switch (state) {
    case 'BOUND':
      return '已绑定'
    case 'UNBOUND':
      return '未绑定'
    case 'INVITING':
      return '邀请中'
    case 'NOT_REQUIRED':
      return '无需绑定'
    default:
      return '未绑定'
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
