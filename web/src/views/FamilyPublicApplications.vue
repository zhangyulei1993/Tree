<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">公开展示</span>
          <h1>家庭公开申请</h1>
        </div>
        <RouterLink class="button secondary" :to="`/families/${familyId}`">返回家庭详情</RouterLink>
      </div>

      <section class="card form-card">
        <h2>提交申请</h2>
        <p class="muted">审核通过后，公开家庭主页和公开家庭树才可匿名访问。</p>
        <form @submit.prevent="submitApplication">
          <label>
            <span>申请理由</span>
            <textarea
              v-model.trim="applicationReason"
              class="field textarea"
              maxlength="500"
              placeholder="说明希望公开展示该家庭的原因（可选）"
            />
          </label>
          <button class="button" :disabled="submitting">
            {{ submitting ? '提交中...' : '提交公开申请' }}
          </button>
        </form>
        <p v-if="submitError" class="feedback error" role="alert">{{ submitError }}</p>
        <p v-if="submitSuccess" class="feedback success" role="status">{{ submitSuccess }}</p>
      </section>

      <section class="records">
        <div class="section-heading">
          <h2>申请记录</h2>
          <button class="button secondary" :disabled="loading" @click="loadApplications">刷新</button>
        </div>
        <div v-if="loading" class="card state-panel">正在加载申请记录...</div>
        <div v-else-if="error" class="card state-panel error" role="alert">
          <strong>申请记录加载失败</strong>
          <span>{{ error }}</span>
          <button class="button secondary" @click="loadApplications">重新加载</button>
        </div>
        <div v-else-if="applications.length === 0" class="card state-panel">
          暂无公开展示申请。
        </div>
        <div v-else class="record-list">
          <article v-for="application in applications" :key="application.applicationId" class="card record">
            <div>
              <div class="record-title">
                <strong>申请 #{{ application.applicationId }}</strong>
                <span class="status" :data-status="application.status">{{ application.status }}</span>
              </div>
              <p>{{ application.reason || '未填写申请理由' }}</p>
              <p class="muted">提交时间：{{ formatDate(application.createdAt) }}</p>
              <p v-if="application.reviewComment" class="muted">
                审核意见：{{ application.reviewComment }}
              </p>
              <p v-if="application.cancelReason" class="muted">
                取消原因：{{ application.cancelReason }}
              </p>
            </div>
            <button
              v-if="application.status === 'PENDING'"
              class="button danger"
              :disabled="cancellingId === application.applicationId"
              @click="cancelApplication(application.applicationId)"
            >
              {{ cancellingId === application.applicationId ? '取消中...' : '取消申请' }}
            </button>
          </article>
        </div>
      </section>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import {
  cancelPublicApplication,
  listPublicApplications,
  submitPublicApplication
} from '@/api/publicApplications'
import PageShell from '@/components/PageShell.vue'
import type { PublicApplication } from '@/types/api'

const route = useRoute()
const familyId = computed(() => String(route.params.familyId))
const applications = ref<PublicApplication[]>([])
const applicationReason = ref('')
const loading = ref(true)
const submitting = ref(false)
const cancellingId = ref<number | null>(null)
const error = ref('')
const submitError = ref('')
const submitSuccess = ref('')

function formatDate(value: string) {
  return new Date(value).toLocaleString('zh-CN')
}

async function loadApplications() {
  loading.value = true
  error.value = ''
  try {
    const result = await listPublicApplications(familyId.value, { page: 1, pageSize: 50 })
    applications.value = result.items
  } catch (requestError) {
    applications.value = []
    error.value = apiErrorMessage(requestError, '无法加载公开申请记录。')
  } finally {
    loading.value = false
  }
}

async function submitApplication() {
  submitting.value = true
  submitError.value = ''
  submitSuccess.value = ''
  try {
    await submitPublicApplication(familyId.value, {
      applicationReason: applicationReason.value || undefined
    })
    applicationReason.value = ''
    submitSuccess.value = '公开申请已提交，等待后台审核。'
    await loadApplications()
  } catch (requestError) {
    submitError.value = apiErrorMessage(requestError, '公开申请提交失败。')
  } finally {
    submitting.value = false
  }
}

async function cancelApplication(applicationId: number) {
  if (!window.confirm('确定取消这条待审核公开申请吗？')) return
  cancellingId.value = applicationId
  error.value = ''
  try {
    await cancelPublicApplication(familyId.value, applicationId)
    await loadApplications()
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '取消公开申请失败。')
  } finally {
    cancellingId.value = null
  }
}

onMounted(loadApplications)
</script>

<style scoped>
.page-section {
  display: grid;
  gap: 20px;
  padding: 32px 0;
}

.page-heading,
.section-heading,
.record,
.record-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

h1,
h2,
p {
  margin-top: 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.form-card {
  padding: 22px;
}

form,
label,
.records,
.record-list {
  display: grid;
  gap: 14px;
}

label span {
  display: block;
  margin-bottom: 8px;
  font-weight: 700;
}

form .button {
  justify-self: start;
}

.record {
  align-items: flex-start;
  padding: 18px;
}

.record p {
  margin: 10px 0 0;
}

.status {
  border-radius: 6px;
  background: #f4f1e8;
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 700;
}

.status[data-status='PENDING'] {
  color: var(--color-warning);
}

.status[data-status='APPROVED'] {
  color: var(--color-success);
}

.button.danger {
  background: var(--color-danger);
}

.state-panel {
  display: grid;
  min-height: 150px;
  place-items: center;
  gap: 12px;
  padding: 24px;
  text-align: center;
}

.state-panel.error,
.feedback.error {
  color: var(--color-danger);
}

.feedback.success {
  color: var(--color-success);
}

@media (max-width: 680px) {
  .page-heading,
  .section-heading,
  .record {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
