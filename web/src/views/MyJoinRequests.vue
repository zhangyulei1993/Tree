<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">加入流程</span>
          <h1>我的加入申请</h1>
        </div>
        <button class="button secondary" :disabled="loading" @click="loadRequests">刷新</button>
      </div>

      <section v-if="loading" class="card state-panel">正在加载加入申请...</section>
      <section v-else-if="error" class="card state-panel error" role="alert">
        <strong>加入申请加载失败</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="loadRequests">重新加载</button>
      </section>
      <section v-else-if="requests.length === 0" class="card state-panel">暂无加入申请。</section>
      <div v-else class="list">
        <article v-for="item in requests" :key="item.requestId" class="card item">
          <div class="item-main">
            <div class="title">
              <strong>{{ item.familyName || `家庭 ${item.familyId}` }}</strong>
              <span class="status">{{ item.requestStatus }}</span>
            </div>
            <p v-if="item.applicantRealName">申请人：{{ item.applicantRealName }}</p>
            <p>{{ item.applicantMessage || '未填写申请理由' }}</p>
            <p v-if="item.handleComment">审核意见：{{ item.handleComment }}</p>
            <p class="muted">提交时间：{{ formatDate(item.createdAt) }}</p>
          </div>
          <button
            v-if="item.requestStatus === 'PENDING'"
            class="button danger"
            :disabled="cancellingId === item.requestId"
            @click="cancel(item)"
          >
            {{ cancellingId === item.requestId ? '取消中...' : '取消申请' }}
          </button>
        </article>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { cancelJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import PageShell from '@/components/PageShell.vue'
import type { JoinRequest } from '@/types/api'

const requests = ref<JoinRequest[]>([])
const loading = ref(true)
const error = ref('')
const cancellingId = ref<number | string | null>(null)

function formatDate(value: string) {
  return new Date(value).toLocaleString('zh-CN')
}

async function loadRequests() {
  loading.value = true
  error.value = ''
  try {
    requests.value = await listMyJoinRequests()
  } catch (requestError) {
    requests.value = []
    error.value = apiErrorMessage(requestError, '无法加载加入申请。')
  } finally {
    loading.value = false
  }
}

async function cancel(item: JoinRequest) {
  if (!window.confirm(`确定取消对 ${item.familyName || `家庭 ${item.familyId}`} 的加入申请吗？`)) return
  cancellingId.value = item.requestId
  error.value = ''
  try {
    await cancelJoinRequest(item.familyId, item.requestId)
    await loadRequests()
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '取消加入申请失败。')
  } finally {
    cancellingId.value = null
  }
}

onMounted(loadRequests)
</script>

<style scoped>
.page-section {
  padding: 32px 0;
}

.page-heading,
.title,
.item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.page-heading {
  margin-bottom: 18px;
}

h1,
p {
  margin-top: 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.list {
  display: grid;
  gap: 14px;
}

.item {
  align-items: flex-start;
  padding: 18px;
}

.item-main {
  flex: 1;
}

.status {
  border-radius: 6px;
  background: #f4f1e8;
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 700;
}

.button.danger {
  background: var(--color-danger);
}

.state-panel {
  display: grid;
  min-height: 180px;
  place-items: center;
  gap: 12px;
  padding: 24px;
  text-align: center;
}

.error {
  color: var(--color-danger);
}

@media (max-width: 680px) {
  .page-heading,
  .item {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
