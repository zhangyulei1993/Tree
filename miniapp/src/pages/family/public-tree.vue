<template>
  <view class="page">
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载公开家谱...</text>
    </view>
    <view v-else-if="errorMessage" class="card state-card">
      <text class="title">公开家谱</text>
      <text class="error">{{ errorMessage }}</text>
      <button class="button secondary" @click="loadTree">重新加载</button>
    </view>
    <template v-else-if="tree">
      <view class="card">
        <text class="title">公开家谱</text>
        <text class="tag">{{ treeModeText(tree.treeMode) }}</text>
        <text class="muted version-text">家谱版本：{{ tree.graphVersion }}</text>
      </view>

      <view class="card">
        <text class="section-title">成员（{{ tree.nodes.length }}）</text>
        <text v-if="tree.nodes.length === 0" class="muted">该家庭暂未公开家谱关系。</text>
        <view v-for="node in tree.nodes" :key="node.memberId" class="data-row">
          <view>
            <text>{{ node.displayName }}</text>
            <text class="muted">{{ genderText(node.gender) }} · {{ livingText(node.isLiving) }}</text>
          </view>
        </view>
      </view>

      <view class="card">
        <text class="section-title">关系（{{ visibleEdges.length }}）</text>
        <text v-if="visibleEdges.length === 0" class="muted">该家庭暂未公开家谱关系。</text>
        <view v-for="edge in visibleEdges" :key="edge.relationshipId" class="data-row">
          <view>
            <text>{{ nodeName(edge.fromMemberId) }} → {{ nodeName(edge.toMemberId) }}</text>
            <text class="muted">{{ relationTypeText(edge.relationshipType) }}</text>
            <text v-if="edge.relationNote" class="subtle">{{ edge.relationNote }}</text>
          </view>
          <text class="tag">{{ relationTagText(edge) }}</text>
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

function nodeName(memberId: number) {
  return nodeNameMap.value.get(memberId) || '未知成员'
}

function treeModeText(mode: string) {
  switch (mode) {
    case 'LIST_TREE':
      return '列表家谱'
    case 'GRAPH_TREE':
      return '图谱家谱'
    default:
      return '公开家谱'
  }
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
.state-card {
  text-align: center;
}

.version-text {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #9ca3af;
}

.data-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 18rpx 0;
  border-top: 1rpx solid #e5e0d6;
}

.data-row view,
.data-row text {
  display: block;
}

.subtle {
  margin-top: 6rpx;
  color: #9ca3af;
  font-size: 22rpx;
}

.error {
  display: block;
  margin-top: 18rpx;
  color: #c0392b;
  font-size: 24rpx;
}
</style>
