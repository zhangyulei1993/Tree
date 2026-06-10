<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">账号消息</span>
          <h1>我的邀请</h1>
        </div>
        <button class="button secondary" :disabled="loading" @click="loadInvitations">刷新</button>
      </div>

      <section v-if="loading" class="card state-panel">正在加载邀请...</section>
      <section v-else-if="loadError" class="card state-panel error" role="alert">
        <strong>邀请列表加载失败</strong>
        <span>{{ loadError }}</span>
        <button class="button secondary" @click="loadInvitations">重新加载</button>
      </section>
      <template v-else>
        <p v-if="actionError" class="action-error" role="alert">
          操作失败：{{ actionError }}
        </p>
        <section v-if="invitations.length === 0" class="card state-panel">暂无收到的站内邀请。</section>
        <div v-else class="list">
          <article v-for="item in invitations" :key="item.invitationId" class="card item">
            <div class="item-main">
              <div class="title">
                <strong>{{ item.familyName }}</strong>
                <span class="status">{{ item.status }}</span>
              </div>
              <p>目标成员：{{ item.targetMemberName }}</p>
              <p>邀请渠道：{{ item.inviteChannel }} · 接受后角色：{{ item.familyRoleAfterAccept }}</p>
              <p v-if="item.inviteMessage">{{ item.inviteMessage }}</p>
              <p class="muted">有效期至：{{ formatDate(item.expiredAt) }}</p>
            </div>
            <div v-if="item.status === 'PENDING'" class="actions">
              <button class="button" :disabled="actingId === item.invitationId" @click="accept(item)">
                接受
              </button>
              <button class="button secondary" :disabled="actingId === item.invitationId" @click="reject(item)">
                拒绝
              </button>
            </div>
          </article>
        </div>
      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { acceptInvitation, listMyInvitations, rejectInvitation } from '@/api/invitations'
import PageShell from '@/components/PageShell.vue'
import type { Invitation } from '@/types/api'

const invitations = ref<Invitation[]>([])
const loading = ref(true)
const loadError = ref('')
const actionError = ref('')
const actingId = ref<number | string | null>(null)

function formatDate(value: string) {
  return new Date(value).toLocaleString('zh-CN')
}

async function loadInvitations() {
  loading.value = true
  loadError.value = ''
  try {
    invitations.value = await listMyInvitations()
  } catch (requestError) {
    invitations.value = []
    loadError.value = apiErrorMessage(requestError, '无法加载邀请列表。')
  } finally {
    loading.value = false
  }
}

async function accept(item: Invitation) {
  actingId.value = item.invitationId
  actionError.value = ''
  try {
    await acceptInvitation(item.invitationId)
    await loadInvitations()
  } catch (requestError) {
    actionError.value = apiErrorMessage(requestError, '接受邀请失败。')
  } finally {
    actingId.value = null
  }
}

async function reject(item: Invitation) {
  if (!window.confirm(`确定拒绝 ${item.familyName} 的邀请吗？`)) return
  actingId.value = item.invitationId
  actionError.value = ''
  try {
    await rejectInvitation(item.invitationId)
    await loadInvitations()
  } catch (requestError) {
    actionError.value = apiErrorMessage(requestError, '拒绝邀请失败。')
  } finally {
    actingId.value = null
  }
}

onMounted(loadInvitations)
</script>

<style scoped>
.page-section {
  padding: 32px 0;
}

.page-heading,
.title,
.item,
.actions {
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

.action-error {
  border-left: 3px solid var(--color-danger);
  background: #fff5f3;
  padding: 12px;
  color: var(--color-danger);
  font-weight: 700;
}

@media (max-width: 680px) {
  .page-heading,
  .item {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
