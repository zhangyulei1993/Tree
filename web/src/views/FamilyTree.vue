<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">私有家庭树</span>
          <h1>私有家庭树</h1>
        </div>
        <div class="heading-actions">
          <RouterLink class="button secondary" :to="`/families/${familyId}`">家庭详情</RouterLink>
          <RouterLink class="button secondary" :to="`/families/${familyId}/members`">家庭成员</RouterLink>
        </div>
      </div>

      <section v-if="loading" class="card state-panel">正在加载家庭树...</section>
      <section v-else-if="error" class="card state-panel error" role="alert">
        <strong>{{ forbidden ? '无权查看家庭树' : '家庭树加载失败' }}</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="loadTree">重新加载</button>
      </section>
      <template v-else-if="tree">
        <section class="card metrics">
          <div><span>展示方式</span><strong>{{ treeModeText(tree.treeMode) }}</strong></div>
          <div><span>家谱版本</span><strong>第 {{ tree.graphVersion }} 版</strong></div>
          <div><span>成员</span><strong>{{ tree.nodes.length }}</strong></div>
          <div><span>关系</span><strong>{{ visibleEdges.length }}</strong></div>
        </section>

        <div class="content-grid">
          <section>
            <h2>成员</h2>
            <div v-if="tree.nodes.length === 0" class="card state-panel">暂无成员</div>
            <div v-else class="list">
              <article v-for="node in tree.nodes" :key="node.memberId" class="card list-item">
                <div class="item-title">
                  <strong>{{ node.displayName }}</strong>
                  <span>{{ genderLabel(node.gender) }}</span>
                </div>
                <p class="muted">
                  {{ lifeText(node) }} · {{ bindingText(node.userBindingState) }} · {{ memberTypeText(node.memberType) }}
                </p>
              </article>
            </div>
          </section>

          <section>
            <h2>亲属关系</h2>
            <p v-if="hiddenSiblingCount" class="warning">
              已隐藏 {{ hiddenSiblingCount }} 条不应直接展示的兄弟姐妹关系。
            </p>
            <div v-if="visibleEdges.length === 0" class="card state-panel">暂无关系</div>
            <div v-else class="list">
              <article v-for="edge in visibleEdges" :key="edge.relationshipId" class="card list-item">
                <div class="item-title">
                  <strong>{{ relationTypeText(edge.relationshipType) }}</strong>
                </div>
                <p class="muted">
                  {{ memberName(edge.fromMemberId) }} → {{ memberName(edge.toMemberId) }}
                  <template v-if="edge.parentLinkType"> · {{ parentLinkText(edge.parentLinkType) }}</template>
                </p>
              </article>
            </div>
          </section>
        </div>
      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { getPrivateTree } from '@/api/tree'
import PageShell from '@/components/PageShell.vue'
import type { ApiResponse, FamilyTreeResult, TreeNode } from '@/types/api'

const route = useRoute()
const familyId = computed(() => String(route.params.familyId))
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(true)
const error = ref('')
const forbidden = ref(false)
const visibleEdges = computed(() => tree.value?.edges.filter(
  (edge) => String(edge.relationshipType) !== 'SIBLING'
) || [])
const hiddenSiblingCount = computed(() => (tree.value?.edges.length || 0) - visibleEdges.value.length)

function errorText(requestError: unknown) {
  if (axios.isAxiosError<ApiResponse<unknown>>(requestError)) {
    forbidden.value = requestError.response?.status === 403
    if (requestError.response?.data?.code) {
      return `${requestError.response.data.code}：${requestError.response.data.message}`
    }
  }
  return apiErrorMessage(requestError, '无法加载私有家庭树。')
}

function genderLabel(gender: string) {
  return gender === 'MALE' ? '男' : gender === 'FEMALE' ? '女' : '未知'
}

function lifeText(node: TreeNode) {
  if (node.isLiving === true) return '健在'
  if (node.isLiving === false) return '已故'
  return '生卒状态未填写'
}

function memberName(memberId: number) {
  const node = tree.value?.nodes.find((item) => item.memberId === memberId)
  return node ? node.displayName : '未知成员'
}

function treeModeText(mode: string) {
  return mode === 'LIST_TREE' ? '列表家谱' : '家谱'
}

function bindingText(state?: string) {
  if (state === 'BOUND') return '已绑定账号'
  if (state === 'UNBOUND') return '未绑定账号'
  if (state === 'NOT_REQUIRED') return '无需绑定'
  return '绑定状态未填写'
}

function memberTypeText(type?: string) {
  if (type === 'REAL') return '真实成员'
  if (type === 'VIRTUAL') return '记录节点'
  return '成员节点'
}

function relationTypeText(type?: string) {
  if (type === 'PARENT_CHILD') return '父母子女'
  if (type === 'SPOUSE') return '配偶'
  return '亲属关系'
}

function parentLinkText(type?: string) {
  if (type === 'PRIMARY') return '主要关系'
  if (type === 'STEP') return '继亲'
  if (type === 'ADOPTIVE') return '收养'
  if (type === 'SUCCESSION') return '承继'
  if (type === 'NOTE_ONLY') return '备注关系'
  if (type === 'OTHER') return '其他关系'
  return '关系属性未填写'
}

async function loadTree() {
  loading.value = true
  error.value = ''
  forbidden.value = false
  try {
    tree.value = await getPrivateTree(familyId.value)
  } catch (requestError) {
    tree.value = null
    error.value = errorText(requestError)
  } finally {
    loading.value = false
  }
}

onMounted(loadTree)
</script>

<style scoped>
.page-section {
  padding: 44px 0;
}

.page-heading,
.heading-actions,
.item-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.page-heading {
  align-items: flex-end;
  margin-bottom: 18px;
}

h1 {
  margin: 8px 0 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.state-panel {
  display: grid;
  min-height: 160px;
  place-items: center;
  gap: 12px;
  padding: 24px;
  text-align: center;
}

.state-panel.error {
  color: var(--color-danger);
}

.metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  padding: 22px;
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.07), transparent 150px),
    rgba(255, 255, 255, 0.94);
}

.metrics div {
  display: grid;
  gap: 6px;
}

.metrics span {
  color: var(--color-text-secondary);
  font-size: 12px;
}

.content-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
  margin-top: 22px;
}

.list {
  display: grid;
  gap: 10px;
}

.list-item {
  padding: 18px;
}

.list-item p {
  margin-bottom: 0;
}

.item-title span {
  font-size: 13px;
}

.warning {
  color: var(--color-danger);
  font-weight: 600;
}

@media (max-width: 760px) {
  .page-heading,
  .heading-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .metrics,
  .content-grid {
    grid-template-columns: 1fr;
  }
}
</style>
