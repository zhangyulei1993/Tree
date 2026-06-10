<template>
  <PageShell>
    <section class="container my-page">
      <div class="page-heading">
        <div>
          <span class="eyebrow">{{ apiMode === 'real' ? '真实家庭数据' : 'Mock 家庭数据' }}</span>
          <h1>我的家庭</h1>
        </div>
        <div class="heading-actions">
          <RouterLink class="button" to="/me/families/new">创建家庭</RouterLink>
          <RouterLink class="button secondary" to="/me/invitations">我的邀请</RouterLink>
          <RouterLink class="button secondary" to="/me/join-requests">加入申请</RouterLink>
          <button class="button secondary" :disabled="loading" @click="loadFamilies">刷新</button>
        </div>
      </div>

      <section v-if="loading" class="card state-panel" aria-live="polite">
        正在加载家庭列表...
      </section>
      <section v-else-if="error" class="card state-panel error" role="alert">
        <strong>家庭列表加载失败</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="loadFamilies">重新加载</button>
      </section>
      <section v-else-if="families.length === 0" class="card state-panel">
        <strong>还没有加入家庭</strong>
        <span class="muted">后续可以创建家庭，或通过邀请和加入申请进入家庭。</span>
      </section>
      <div v-else class="grid">
        <article v-for="family in families" :key="family.id" class="card item">
          <div>
            <div class="item-title">
              <h3>{{ family.familyName }}</h3>
              <span class="role">{{ family.role }}</span>
            </div>
            <p class="muted">
              姓氏：{{ family.familySurname }}
              <template v-if="family.regionText"> · 地区：{{ family.regionText }}</template>
            </p>
            <p class="meta">
              家庭状态：{{ family.status }} · 公开状态：{{ family.publicDisplayStatus }}
            </p>
          </div>
          <RouterLink class="button secondary" :to="`/families/${family.id}`">进入家庭</RouterLink>
        </article>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { apiErrorMessage, apiMode } from '@/api/client'
import { listMyFamilies } from '@/api/families'
import PageShell from '@/components/PageShell.vue'
import type { FamilySummary } from '@/types/api'

const families = ref<FamilySummary[]>([])
const loading = ref(true)
const error = ref('')

async function loadFamilies() {
  loading.value = true
  error.value = ''
  try {
    families.value = await listMyFamilies()
  } catch (requestError) {
    families.value = []
    error.value = apiErrorMessage(requestError, '无法加载家庭列表。')
  } finally {
    loading.value = false
  }
}

onMounted(loadFamilies)
</script>

<style scoped>
.my-page {
  padding: 32px 0;
}

.page-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.heading-actions {
  display: flex;
  gap: 10px;
}

.page-heading h1 {
  margin: 8px 0 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
}

.item-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.item-title h3 {
  margin: 0;
}

.role {
  border-radius: 6px;
  background: #edf4ef;
  padding: 4px 8px;
  color: var(--color-success);
  font-size: 12px;
  font-weight: 700;
}

.meta {
  margin-bottom: 0;
  font-size: 13px;
}

.state-panel {
  display: grid;
  min-height: 180px;
  place-items: center;
  gap: 12px;
  padding: 28px;
  text-align: center;
}

.state-panel.error {
  color: var(--color-danger);
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 640px) {
  .item,
  .page-heading {
    align-items: stretch;
    flex-direction: column;
  }

  .heading-actions {
    flex-direction: column;
  }
}
</style>
