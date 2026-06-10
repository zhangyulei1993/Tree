<template>
  <view class="page">
    <view class="card">
      <text class="title">私有家庭树</text>
      <text v-if="tree" class="tag">{{ tree.treeMode }}</text>
      <text v-if="tree" class="muted">graphVersion：{{ tree.graphVersion }}</text>
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
      <button v-if="errorMessage" class="button secondary" @click="loadTree">重新加载</button>
    </view>
    <view v-if="loading" class="card state-card">
      <text class="muted">正在组装家庭树...</text>
    </view>
    <template v-else-if="tree">
      <view class="card">
        <text class="section-title">成员节点（{{ tree.nodes.length }}）</text>
        <text v-if="tree.nodes.length === 0" class="muted">暂无节点。</text>
        <view v-for="node in tree.nodes" :key="node.memberId" class="data-row">
          <view>
            <text>{{ node.displayName }}</text>
            <text class="muted">ID {{ node.memberId }} · {{ node.gender }}</text>
          </view>
          <text class="tag">{{ node.isLiving === false ? '已故' : '健在/未知' }}</text>
        </view>
      </view>
      <view class="card">
        <text class="section-title">关系边（{{ visibleEdges.length }}）</text>
        <text v-if="visibleEdges.length === 0" class="muted">暂无父子或配偶关系。</text>
        <view v-for="edge in visibleEdges" :key="edge.relationshipId" class="data-row">
          <view>
            <text>{{ edge.fromMemberId }} → {{ edge.toMemberId }}</text>
            <text class="muted">关系 ID {{ edge.relationshipId }}</text>
          </view>
          <text class="tag">{{ edge.relationshipType }}</text>
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

async function loadTree() {
  if (!familyId.value) {
    errorMessage.value = '缺少 familyId。'
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
</style>
