<template>
  <view class="page">
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载公开家庭树...</text>
    </view>
    <view v-else-if="errorMessage" class="card state-card">
      <text class="title">公开家庭树</text>
      <text class="error">公开家庭树不可访问：{{ errorMessage }}</text>
      <button class="button secondary" @click="loadTree">重新加载</button>
    </view>
    <template v-else-if="tree">
      <view class="card">
        <text class="title">公开家庭树</text>
        <text class="muted">列表树展示，不包含账号、手机号、openid、unionid 或 token。</text>
        <text class="tag">模式：{{ tree.treeMode }}</text>
        <text class="tag">Graph Version：{{ tree.graphVersion }}</text>
        <text class="tag">节点：{{ tree.nodes.length }}</text>
        <text class="tag">关系：{{ visibleEdges.length }}</text>
      </view>

      <view v-for="item in tree.tree" :key="item.memberId" class="card">
        <text class="section-title">{{ nodeName(item.memberId) }}</text>
        <text class="tag">{{ nodeGender(item.memberId) }}</text>
        <text class="muted">父母：{{ names(item.parentIds) || '未展示' }}</text>
        <text class="muted">配偶：{{ names(item.spouseIds) || '无' }}</text>
        <text class="muted">子女：{{ names(item.childrenIds) || '无' }}</text>
      </view>

      <view class="card">
        <text class="section-title">公开关系</text>
        <text v-if="visibleEdges.length === 0" class="muted">暂无公开关系。</text>
        <view v-for="edge in visibleEdges" v-else :key="edge.relationshipId" class="edge-row">
          <text class="tag">{{ edge.relationshipType }}</text>
          <text class="muted">{{ nodeName(edge.fromMemberId) }} → {{ nodeName(edge.toMemberId) }}</text>
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

function node(memberId: number) {
  return tree.value?.nodes.find((item) => item.memberId === memberId)
}

function nodeName(memberId: number) {
  return node(memberId)?.displayName || `成员 ${memberId}`
}

function nodeGender(memberId: number) {
  const gender = node(memberId)?.gender
  return gender === 'MALE' ? '男' : gender === 'FEMALE' ? '女' : '未知'
}

function names(ids: number[]) {
  return ids.map(nodeName).join('、')
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
    errorMessage.value = '缺少 familyId。'
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
    errorMessage.value = '缺少 familyId。'
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
      errorMessage.value = apiErrorMessage(error, '无法加载公开家庭树。')
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

.edge-row {
  padding: 14rpx 0;
  border-top: 1rpx solid #e5e0d6;
}

.error {
  display: block;
  margin-top: 18rpx;
  color: #c0392b;
  font-size: 24rpx;
}
</style>
