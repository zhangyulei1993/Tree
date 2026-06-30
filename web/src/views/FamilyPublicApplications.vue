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

      <section v-if="loading" class="card state-panel">正在确认家庭公开状态...</section>

      <section v-else-if="error" class="card state-panel error" role="alert">
        <strong>公开状态加载失败</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="loadApplications">重新加载</button>
      </section>

      <section v-else-if="family?.status !== 'NORMAL'" class="card form-card">
        <h2>公开展示暂不可调整</h2>
        <p class="muted">家庭当前为{{ familyStatusText(family?.status) }}状态，请先完成当前家庭状态处理。</p>
      </section>

      <section v-else-if="family?.publicDisplayStatus === 'APPROVED'" class="card form-card public-active">
        <h2>家庭当前已公开</h2>
        <p class="muted">公开家庭主页和公开家谱可被匿名访问。</p>
        <div class="public-actions">
          <RouterLink class="button secondary" :to="`/families/${familyId}/public`">查看公开主页</RouterLink>
          <RouterLink class="button secondary" :to="`/families/${familyId}/tree/public`">查看公开家谱</RouterLink>
        </div>
        <label>
          <span>关闭说明</span>
          <textarea
            v-model.trim="takeDownReason"
            class="field textarea"
            maxlength="500"
            placeholder="可选，仅用于操作记录"
          />
        </label>
        <p class="privacy-note">家庭主动关闭公开展示会立即生效，无需再次等待后台审核；重新公开时需要重新提交申请。</p>
        <button class="button danger" :disabled="takingDown" @click="closePublicDisplay">
          {{ takingDown ? '关闭中...' : '关闭公开展示' }}
        </button>
      </section>

      <section v-else-if="family?.publicDisplayStatus === 'PENDING'" class="card form-card">
        <h2>公开申请审核中</h2>
        <p class="muted">后台审核完成前，公开主页和公开家谱不会对外开放。你可以在下方申请记录中取消待审核申请。</p>
      </section>

      <section v-else class="card form-card">
        <h2>提交申请</h2>
        <p v-if="family?.publicDisplayStatus === 'REJECTED'" class="feedback error">
          上次公开申请未通过，请查看审核意见后重新提交。
        </p>
        <p v-else-if="family?.publicDisplayStatus === 'TAKEN_DOWN'" class="feedback">
          该家庭当前已关闭公开展示，如需重新公开请再次提交审核。
        </p>
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
      </section>
      <p v-if="submitError" class="feedback error" role="alert">{{ submitError }}</p>
      <p v-if="submitSuccess" class="feedback success" role="status">{{ submitSuccess }}</p>

      <section class="records">
        <div class="section-heading">
          <h2>申请记录</h2>
          <button class="button secondary" :disabled="loading" @click="loadApplications">刷新</button>
        </div>
        <div v-if="loading" class="card state-panel">正在加载申请记录...</div>
        <div v-else-if="applications.length === 0" class="card state-panel">
          暂无公开展示申请。
        </div>
        <div v-else class="record-list">
          <article v-for="application in applications" :key="application.applicationId" class="card record">
            <div>
              <div class="record-title">
                <strong>公开展示申请</strong>
                <span class="status" :data-status="application.status">{{ applicationStatusText(application.status) }}</span>
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
import { getFamilyDetail } from '@/api/families'
import {
  cancelPublicApplication,
  closePublicFamily,
  listPublicApplications,
  submitPublicApplication
} from '@/api/publicApplications'
import PageShell from '@/components/PageShell.vue'
import type { FamilyDetail, PublicApplication } from '@/types/api'

const route = useRoute()
const familyId = computed(() => String(route.params.familyId))
const family = ref<FamilyDetail | null>(null)
const applications = ref<PublicApplication[]>([])
const applicationReason = ref('')
const takeDownReason = ref('')
const loading = ref(true)
const submitting = ref(false)
const takingDown = ref(false)
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
    const [familyResult, result] = await Promise.all([
      getFamilyDetail(familyId.value),
      listPublicApplications(familyId.value, { page: 1, pageSize: 50 })
    ])
    family.value = familyResult
    applications.value = result.items
  } catch (requestError) {
    family.value = null
    applications.value = []
    error.value = apiErrorMessage(requestError, '无法加载公开申请记录。')
  } finally {
    loading.value = false
  }
}

async function closePublicDisplay() {
  if (!window.confirm('关闭后公开主页和公开家谱将立即不可访问。确定关闭公开展示吗？')) return
  takingDown.value = true
  submitError.value = ''
  submitSuccess.value = ''
  try {
    await closePublicFamily(familyId.value, {
      reason: takeDownReason.value || undefined
    })
    takeDownReason.value = ''
    submitSuccess.value = '公开展示已关闭。'
    await loadApplications()
  } catch (requestError) {
    submitError.value = apiErrorMessage(requestError, '关闭公开展示失败。')
  } finally {
    takingDown.value = false
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

function applicationStatusText(status?: string) {
  if (status === 'PENDING') return '待审核'
  if (status === 'APPROVED') return '已通过'
  if (status === 'REJECTED') return '已驳回'
  if (status === 'CANCELLED') return '已取消'
  return '未知'
}

function familyStatusText(status?: string) {
  if (status === 'DISSOLUTION_PENDING') return '解散待审核'
  if (status === 'DISSOLVED') return '已解散'
  if (status === 'DISABLED') return '已停用'
  if (status === 'NORMAL') return '正常'
  return '不可用'
}
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

.public-active {
  border-color: rgba(47, 107, 87, 0.35);
}

.public-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 16px;
}

.privacy-note {
  border-left: 3px solid var(--color-warning);
  background: #fff8e9;
  padding: 10px 12px;
  color: var(--color-text-secondary);
  line-height: 1.6;
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
