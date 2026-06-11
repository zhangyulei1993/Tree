<template>
  <view class="page">
    <view class="card">
      <text class="title">私有家庭树</text>
      <text v-if="tree" class="tag">{{ treeModeText(tree.treeMode) }}</text>
      <text v-if="tree" class="muted">家谱版本：{{ tree.graphVersion }}</text>
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
      <button v-if="errorMessage" class="button secondary" @click="loadTree">重新加载</button>
    </view>
    <view v-if="loading" class="card state-card">
      <text class="muted">正在组装家庭树...</text>
    </view>
    <template v-else-if="tree">
      <view class="card">
        <text class="section-title">成员（{{ tree.nodes.length }}）</text>
        <text v-if="tree.nodes.length === 0" class="muted">暂无成员。请先在家庭成员中添加成员后再查看家谱。</text>
        <view v-for="node in tree.nodes" :key="node.memberId" class="data-row">
          <view>
            <text>{{ node.displayName }}</text>
            <text class="muted">{{ genderText(node.gender) }} · {{ livingText(node.isLiving) }}</text>
          </view>
          <text class="tag">{{ bindingStateText(node.userBindingState) }}</text>
        </view>
      </view>
      <view class="card">
        <text class="section-title">关系（{{ visibleEdges.length }}）</text>
        <text v-if="visibleEdges.length === 0" class="muted">暂无父母子女或配偶关系。当前仅展示已维护的家庭关系。</text>
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
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPrivateTree } from '@/api/tree'
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

function treeModeText(mode: string) {
  switch (mode) {
    case 'LIST_TREE':
      return '列表树'
    case 'GRAPH_TREE':
      return '图谱树'
    default:
      return '家谱'
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
.state-card {
  text-align: center;
}

.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
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
</style>
