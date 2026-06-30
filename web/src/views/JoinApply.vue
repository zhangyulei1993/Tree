<template>
  <PageShell>
    <section class="container join-page">
      <section v-if="loading" class="card state-card">正在加载家庭信息...</section>
      <section v-else-if="familyError" class="card state-card error" role="alert">
        <strong>家庭信息不可访问</strong>
        <span>{{ familyError }}</span>
        <button class="button secondary" @click="loadFamily">重新加载</button>
      </section>
      <form v-else-if="family" class="card join-card" @submit.prevent="submit">
        <div>
          <span class="eyebrow">加入家庭</span>
          <h1>申请加入 {{ family.familyName }}</h1>
          <p class="muted">{{ family.familySurname }}氏 · {{ family.regionText || '未填写地区' }}</p>
        </div>

        <div v-if="!session.isLoggedIn" class="notice">
          请先登录后申请加入。
          <button class="button" type="button" @click="goToLogin">去登录</button>
        </div>
        <div v-else-if="!session.isPhoneBound" class="notice">
          当前账号尚未完成手机号验证，请先完成真实手机号绑定后再申请。
        </div>
        <section v-else-if="alreadyMember" class="result-card">
          <strong>你已经是该家庭成员</strong>
          <span>无需重复提交加入申请，可以直接进入“我的家庭”查看。</span>
          <RouterLink class="button secondary" to="/me/families">查看我的家庭</RouterLink>
        </section>
        <section v-else-if="submittedRequest?.requestStatus === 'PENDING'" class="result-card">
          <strong>加入申请审核中</strong>
          <span>家庭管理员尚未处理，请勿重复提交。</span>
          <RouterLink class="button secondary" to="/me/join-requests">查看或取消申请</RouterLink>
        </section>
        <template v-else-if="!submittedRequest">
          <label>
            <span>申请人真实姓名</span>
            <input
              v-model.trim="applicantRealName"
              class="field"
              maxlength="100"
              placeholder="可选"
            />
          </label>
          <label>
            <span>申请人性别</span>
            <select v-model="applicantGender" class="field">
              <option value="MALE">男</option>
              <option value="FEMALE">女</option>
            </select>
          </label>
          <label>
            <span>申请理由</span>
            <textarea
              v-model.trim="applicantMessage"
              class="field textarea"
              maxlength="500"
              placeholder="请说明与该家庭的关系或加入原因"
            />
          </label>
          <button class="button" :disabled="submitting">
            {{ submitting ? '提交中...' : '提交加入申请' }}
          </button>
        </template>

        <section v-else class="result-card">
          <strong>申请已提交</strong>
          <span>当前状态：{{ requestStatusText(submittedRequest.requestStatus) }}</span>
          <RouterLink class="button secondary" to="/me/join-requests">查看我的加入申请</RouterLink>
        </section>
        <p v-if="error" class="feedback error" role="alert">{{ error }}</p>
      </form>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { getPublicFamilyDetail, listMyFamilies } from '@/api/families'
import { createJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'
import type { Gender, JoinRequest, PublicFamily } from '@/types/api'

const route = useRoute()
const router = useRouter()
const session = useSessionStore()
const familyId = computed(() => String(route.params.familyId))
const family = ref<PublicFamily | null>(null)
const applicantRealName = ref('')
const applicantGender = ref<Gender>('MALE')
const applicantMessage = ref('')
const submittedRequest = ref<JoinRequest | null>(null)
const alreadyMember = ref(false)
const loading = ref(true)
const submitting = ref(false)
const familyError = ref('')
const error = ref('')

function goToLogin() {
  router.push({ path: '/login', query: { redirect: route.fullPath } })
}

async function loadFamily() {
  loading.value = true
  familyError.value = ''
  submittedRequest.value = null
  alreadyMember.value = false
  try {
    family.value = await getPublicFamilyDetail(familyId.value)
    if (session.isLoggedIn && session.isPhoneBound) {
      const [requests, myFamilies] = await Promise.all([
        listMyJoinRequests(),
        listMyFamilies()
      ])
      alreadyMember.value = myFamilies.some((item) => String(item.id) === familyId.value)
      submittedRequest.value = requests.find(
        (item) => String(item.familyId) === familyId.value && item.requestStatus === 'PENDING'
      ) || null
    }
  } catch (requestError) {
    family.value = null
    familyError.value = apiErrorMessage(requestError, '无法加载公开家庭信息。')
  } finally {
    loading.value = false
  }
}

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    submittedRequest.value = await createJoinRequest(familyId.value, {
      applicantRealName: applicantRealName.value || undefined,
      applicantGender: applicantGender.value,
      applicantMessage: applicantMessage.value || undefined
    })
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '加入申请提交失败。')
  } finally {
    submitting.value = false
  }
}

onMounted(loadFamily)

function requestStatusText(status?: string) {
  if (status === 'PENDING') return '待审核'
  if (status === 'APPROVED') return '已通过'
  if (status === 'REJECTED') return '已驳回'
  if (status === 'CANCELLED') return '已取消'
  return '未知状态'
}
</script>

<style scoped>
.join-page {
  display: grid;
  min-height: 560px;
  place-items: center;
  padding: 32px 0;
}

.join-card,
.state-card {
  display: grid;
  width: min(600px, 100%);
  gap: 14px;
  padding: 24px;
}

.state-card {
  min-height: 220px;
  place-items: center;
  text-align: center;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

label {
  display: grid;
  gap: 7px;
  font-weight: 600;
}

.notice,
.result-card {
  display: grid;
  justify-items: start;
  gap: 10px;
  border-left: 3px solid var(--color-warm-gold);
  background: #fbf7ed;
  padding: 14px;
}

.feedback {
  font-weight: 700;
}

.error {
  color: var(--color-danger);
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>
