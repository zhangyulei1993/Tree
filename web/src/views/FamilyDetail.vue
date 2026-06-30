<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">私有家庭</span>
          <h1>{{ family?.familyName || '家庭详情' }}</h1>
        </div>
        <RouterLink class="button secondary" to="/me/families">返回我的家庭</RouterLink>
      </div>

      <section v-if="loading" class="card state-panel">正在加载家庭详情...</section>
      <section v-else-if="error" class="card state-panel error" role="alert">
        <strong>{{ forbidden ? '无权查看该家庭' : '家庭详情加载失败' }}</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="loadFamily">重新加载</button>
      </section>
      <template v-else-if="family">
        <section class="card detail-card">
          <div class="summary">
            <div><span>姓氏</span><strong>{{ family.familySurname }}</strong></div>
            <div><span>我的身份</span><strong>{{ roleText(family.role) }}</strong></div>
            <div><span>家庭状态</span><strong>{{ statusText(family.status) }}</strong></div>
            <div><span>公开状态</span><strong>{{ publicStatusText(family.publicDisplayStatus) }}</strong></div>
            <div><span>家谱版本</span><strong>第 {{ family.graphVersion }} 版</strong></div>
            <div><span>创建者记录</span><strong>{{ family.currentFounderMemberId ? '已确认' : '未确认' }}</strong></div>
          </div>
          <p v-if="family.description" class="description">{{ family.description }}</p>
          <p class="muted">
            {{ family.nativePlace || '未填写籍贯' }}
            <template v-if="family.regionText"> · {{ family.regionText }}</template>
          </p>
          <div
            v-if="canManage && family.status === 'NORMAL' && family.publicDisplayStatus === 'APPROVED'"
            class="public-status-actions"
          >
            <RouterLink class="button secondary" :to="`/families/${family.id}/public`">
              查看公开主页
            </RouterLink>
            <RouterLink class="button secondary" :to="`/families/${family.id}/public-applications`">
              公开记录
            </RouterLink>
            <button class="button danger" :disabled="closingPublic" @click="confirmClosePublic">
              {{ closingPublic ? '关闭中...' : '关闭公开展示' }}
            </button>
          </div>
          <p v-if="operationError" class="operation-error" role="alert">{{ operationError }}</p>
        </section>

        <div class="entry-grid">
          <RouterLink class="card entry" :to="`/families/${family.id}/members`">
            <strong>家庭成员</strong>
            <span>{{ canManage ? '查看并维护成员和新亲属关系' : '查看家庭成员' }}</span>
          </RouterLink>
          <RouterLink class="card entry" :to="`/families/${family.id}/tree`">
            <strong>私有家庭树</strong>
            <span>查看成员列表、亲属关系和家谱版本</span>
          </RouterLink>
          <RouterLink class="card entry" :to="`/families/${family.id}/manage`">
            <strong>{{ canManage ? '家庭管理' : '成员身份' }}</strong>
            <span>{{ canManage ? '处理申请、邀请、角色与高风险操作' : '查看身份或退出该家庭' }}</span>
          </RouterLink>
          <RouterLink
            v-if="canManage && family.status === 'NORMAL'"
            class="card entry"
            :to="`/families/${family.id}/public-applications`"
          >
            <strong>{{ publicEntryTitle }}</strong>
            <span>{{ publicEntryDescription }}</span>
          </RouterLink>
          <div v-else-if="canManage" class="card entry entry-disabled">
            <strong>公开展示暂不可调整</strong>
            <span>家庭当前为{{ statusText(family.status) }}状态</span>
          </div>
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
import { getFamilyDetail } from '@/api/families'
import { closePublicFamily } from '@/api/publicApplications'
import PageShell from '@/components/PageShell.vue'
import type { ApiResponse, FamilyDetail } from '@/types/api'

const route = useRoute()
const family = ref<FamilyDetail | null>(null)
const loading = ref(true)
const error = ref('')
const forbidden = ref(false)
const closingPublic = ref(false)
const operationError = ref('')
const familyId = computed(() => String(route.params.familyId))
const canManage = computed(() => family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN')
const publicEntryTitle = computed(() => {
  if (family.value?.publicDisplayStatus === 'APPROVED') return '公开展示管理'
  if (family.value?.publicDisplayStatus === 'PENDING') return '公开申请审核中'
  return '申请公开展示'
})
const publicEntryDescription = computed(() => {
  if (family.value?.publicDisplayStatus === 'APPROVED') return '查看公开主页，或关闭公开展示'
  if (family.value?.publicDisplayStatus === 'PENDING') return '查看审核进度，或取消待审核申请'
  if (family.value?.publicDisplayStatus === 'REJECTED') return '查看审核意见并重新提交申请'
  if (family.value?.publicDisplayStatus === 'TAKEN_DOWN') return '当前已关闭，可重新申请公开'
  return '提交公开申请并查看审核进度'
})

function roleText(role?: string) {
  if (role === 'FOUNDER') return '家庭创建者'
  if (role === 'FAMILY_ADMIN') return '家庭管理员'
  if (role === 'MEMBER') return '家庭成员'
  return '未知'
}

function statusText(status?: string) {
  if (status === 'NORMAL') return '正常'
  if (status === 'DISSOLUTION_PENDING') return '解散待审核'
  if (status === 'DISSOLVED') return '已解散'
  if (status === 'DISABLED') return '已停用'
  return '未知'
}

function publicStatusText(status?: string) {
  if (status === 'APPROVED') return '已公开'
  if (status === 'PENDING') return '审核中'
  if (status === 'REJECTED') return '未通过'
  if (status === 'TAKEN_DOWN') return '已关闭'
  if (status === 'PRIVATE') return '未公开'
  return '未知'
}

async function confirmClosePublic() {
  if (!family.value || closingPublic.value) return
  if (!window.confirm('关闭后公开主页和公开家谱将立即不可访问，重新公开需要再次提交审核。确定继续吗？')) return
  closingPublic.value = true
  operationError.value = ''
  try {
    await closePublicFamily(familyId.value, {})
    await loadFamily()
  } catch (requestError) {
    operationError.value = apiErrorMessage(requestError, '关闭公开展示失败。')
  } finally {
    closingPublic.value = false
  }
}

function errorText(requestError: unknown) {
  if (axios.isAxiosError<ApiResponse<unknown>>(requestError)) {
    forbidden.value = requestError.response?.status === 403
    if (requestError.response?.data?.code) {
      return `${requestError.response.data.code}：${requestError.response.data.message}`
    }
  }
  return apiErrorMessage(requestError, '无法加载家庭详情。')
}

async function loadFamily() {
  loading.value = true
  error.value = ''
  operationError.value = ''
  forbidden.value = false
  try {
    family.value = await getFamilyDetail(familyId.value)
  } catch (requestError) {
    family.value = null
    error.value = errorText(requestError)
  } finally {
    loading.value = false
  }
}

onMounted(loadFamily)
</script>

<style scoped>
.page-section {
  padding: 44px 0;
}

.page-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
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
  min-height: 190px;
  place-items: center;
  gap: 12px;
  padding: 28px;
  text-align: center;
}

.state-panel.error {
  color: var(--color-danger);
}

.detail-card {
  position: relative;
  overflow: hidden;
  padding: 28px;
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.08), transparent 170px),
    rgba(255, 255, 255, 0.94);
}

.detail-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 4px;
  background: linear-gradient(90deg, var(--color-primary), var(--color-heritage-green), var(--color-warm-gold));
}

.summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.summary div {
  display: grid;
  gap: 7px;
  border-left: 3px solid var(--color-warm-gold);
  padding-left: 12px;
}

.summary span,
.entry span {
  color: var(--color-text-secondary);
  font-size: 13px;
}

.description {
  margin: 24px 0 8px;
  line-height: 1.7;
}

.public-status-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 20px;
}

.operation-error {
  margin: 12px 0 0;
  color: var(--color-danger);
}

.entry-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-top: 18px;
}

.entry {
  display: grid;
  gap: 8px;
  padding: 24px;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.entry-disabled {
  cursor: not-allowed;
  opacity: 0.66;
}

.entry:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-soft);
  transform: translateY(-2px);
}

@media (max-width: 760px) {
  .page-heading {
    align-items: stretch;
    flex-direction: column;
  }

  .summary,
  .entry-grid {
    grid-template-columns: 1fr;
  }
}
</style>
