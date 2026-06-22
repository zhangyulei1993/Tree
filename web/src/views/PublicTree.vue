<template>
  <PageShell>
    <section class="container tree-page">
      <div class="page-heading">
        <div>
          <span class="eyebrow">公开家庭树</span>
          <h1>家庭树</h1>
        </div>
        <RouterLink class="button secondary" :to="`/families/${familyId}/public`">返回公开主页</RouterLink>
      </div>

      <section v-if="loading" class="card state-panel">正在加载公开家庭树...</section>
      <section v-else-if="error" class="card state-panel error" role="alert">
        <strong>公开家庭树不可访问</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="loadTree">重新加载</button>
      </section>
      <template v-else-if="tree">
        <div class="meta">
          <span>展示方式：{{ treeModeText(tree.treeMode) }}</span>
          <span>家谱版本：第 {{ tree.graphVersion }} 版</span>
          <span>成员：{{ tree.nodes.length }}</span>
          <span>关系：{{ visibleEdges.length }}</span>
        </div>

        <section class="tree-list">
          <article v-for="item in tree.tree" :key="item.memberId" class="card tree-node">
            <div class="node-heading">
              <div>
                <h2>{{ nodeName(item.memberId) }}</h2>
                <span>{{ nodeGender(item.memberId) }}</span>
              </div>
            </div>
            <div class="relations">
              <span>父母：{{ names(item.parentIds) || '未展示' }}</span>
              <span>配偶：{{ names(item.spouseIds) || '无' }}</span>
              <span>子女：{{ names(item.childrenIds) || '无' }}</span>
            </div>
          </article>
        </section>

        <section class="card edge-panel">
          <h2>公开关系</h2>
          <p v-if="visibleEdges.length === 0" class="muted">暂无公开关系。</p>
          <div v-else class="edge-list">
            <div v-for="edge in visibleEdges" :key="edge.relationshipId" class="edge-row">
              <strong>{{ relationTypeText(edge.relationshipType) }}</strong>
              <span>{{ nodeName(edge.fromMemberId) }} → {{ nodeName(edge.toMemberId) }}</span>
            </div>
          </div>
        </section>
      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { getPublicTree } from '@/api/tree'
import PageShell from '@/components/PageShell.vue'
import type { FamilyTreeResult } from '@/types/api'

const route = useRoute()
const familyId = computed(() => String(route.params.familyId))
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(true)
const error = ref('')
const visibleEdges = computed(() =>
  tree.value?.edges.filter((edge) => String(edge.relationshipType) !== 'SIBLING') || []
)

function node(memberId: number) {
  return tree.value?.nodes.find((item) => item.memberId === memberId)
}

function nodeName(memberId: number) {
  return node(memberId)?.displayName || '未知成员'
}

function nodeGender(memberId: number) {
  const gender = node(memberId)?.gender
  return gender === 'MALE' ? '男' : gender === 'FEMALE' ? '女' : '未知'
}

function treeModeText(mode: string) {
  return mode === 'LIST_TREE' ? '列表家谱' : '家谱'
}

function relationTypeText(type?: string) {
  if (type === 'PARENT_CHILD') return '父母子女'
  if (type === 'SPOUSE') return '配偶'
  return '亲属关系'
}

function names(ids: number[]) {
  return ids.map(nodeName).join('、')
}

async function loadTree() {
  loading.value = true
  error.value = ''
  try {
    tree.value = await getPublicTree(familyId.value)
  } catch (requestError) {
    tree.value = null
    error.value = apiErrorMessage(requestError, '无法加载公开家庭树。')
  } finally {
    loading.value = false
  }
}

onMounted(loadTree)
</script>

<style scoped>
.tree-page {
  display: grid;
  gap: 18px;
  padding: 48px 0 32px;
}

.page-heading,
.node-heading,
.meta,
.edge-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

h1,
h2 {
  margin: 0;
}

.eyebrow {
  display: block;
  margin-bottom: 8px;
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.meta {
  justify-content: flex-start;
  flex-wrap: wrap;
  color: var(--color-text-secondary);
  font-size: 13px;
}

.meta span,
.relations span {
  border-radius: 999px;
  background: var(--color-bg-soft);
  padding: 6px 10px;
}

.tree-list,
.edge-list {
  display: grid;
  gap: 12px;
}

.tree-node,
.edge-panel {
  padding: 22px;
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.06), transparent 140px),
    rgba(255, 255, 255, 0.94);
}

.node-heading span,
.edge-row span {
  color: var(--color-text-secondary);
}

.relations {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
  color: var(--color-text-secondary);
  font-size: 13px;
}

.edge-panel {
  display: grid;
  gap: 14px;
}

.edge-row {
  justify-content: flex-start;
  border-top: 1px solid var(--color-border);
  padding-top: 12px;
}

.state-panel {
  display: grid;
  min-height: 220px;
  place-items: center;
  gap: 12px;
  padding: 28px;
  text-align: center;
}

.state-panel.error {
  color: var(--color-danger);
}

@media (max-width: 680px) {
  .page-heading,
  .node-heading {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
