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
            <text class="tree-archive-chip green">关系 {{ visibleEdges.length }}</text>
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
          <text class="tree-space-title">成员档案</text>
          <text class="tree-space-subtitle">该家庭公开可见的成员信息</text>
        </view>
        <view class="tree-space-body section-pad">
        <MiniEmptyState
          v-if="tree.nodes.length === 0"
          title="暂无公开成员"
          description="该家庭暂未公开家谱成员。"
        />
        <MemberMiniCard
          v-for="node in tree.nodes"
          :key="node.memberId"
          :name="node.displayName"
          :gender-label="genderText(node.gender)"
          :life-info="livingText(node.isLiving)"
        />
        </view>
      </view>

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">亲属关系</text>
          <text class="tree-space-subtitle">以自然语言描述的公开亲属关系</text>
        </view>
        <view class="tree-space-body section-pad">
        <MiniEmptyState
          v-if="visibleEdges.length === 0"
          title="暂无公开关系"
          description="该家庭暂未公开家谱关系。"
        />
        <RelationSentenceList v-else :items="relationSentences" />
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
import MemberMiniCard from '@/components/family/MemberMiniCard.vue'
import RelationSentenceList from '@/components/family/RelationSentenceList.vue'
import type { FamilyTreeResult } from '@/types/api'

const familyId = ref('')
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const errorMessage = ref('')
let requestVersion = 0
const visibleEdges = computed(() =>
  tree.value?.edges.filter((edge) => String(edge.relationshipType) !== 'SIBLING') || []
)
const nodeNameMap = computed(() => new Map((tree.value?.nodes || []).map((node) => [node.memberId, node.displayName])))
const nodeGenderMap = computed(() => new Map((tree.value?.nodes || []).map((node) => [node.memberId, node.gender])))
const relationSentences = computed(() =>
  visibleEdges.value.map((edge) => buildRelationSentence(edge))
)

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
      return '公开家谱'
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

function relationTypeText(type: string) {
  switch (type) {
    case 'PARENT_CHILD':
      return '父母子女'
    case 'SPOUSE':
      return '配偶'
    case 'SIBLING':
      return '兄弟姐妹'
    default:
      return '家庭关系'
  }
}

function relationTagText(edge: FamilyTreeResult['edges'][number]) {
  if (edge.relationshipType === 'PARENT_CHILD') {
    return parentLinkText(edge.parentLinkType)
  }
  return relationTypeText(edge.relationshipType)
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
