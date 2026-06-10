<template>
  <view class="page">
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载家庭信息...</text>
    </view>
    <view v-else-if="familyError" class="card state-card">
      <text class="title">加入家庭申请</text>
      <text class="error">{{ familyError }}</text>
      <button class="button secondary" @click="loadFamily">重新加载</button>
    </view>
    <view v-else-if="family" class="card">
      <text class="title">加入家庭申请</text>
      <text class="muted">家庭：{{ family.familyName }}</text>
      <text class="muted">姓氏：{{ family.familySurname }}</text>
      <text class="muted">地区：{{ family.regionText || '未设置' }}</text>
      <text class="muted">{{ family.description || '该家庭暂未填写公开简介。' }}</text>

      <view v-if="!session.isLoggedIn">
        <text class="notice">请先使用手机号登录或注册，再提交加入申请。</text>
        <button class="button" @click="requireLogin">手机号登录</button>
      </view>
      <view v-else-if="!session.isPhoneBound">
        <text class="notice">当前账号尚未完成手机号验证，请使用手机号账号登录或注册。</text>
        <button class="button" @click="requireLogin">手机号登录</button>
      </view>
      <view v-else-if="!submittedRequest">
        <input
          v-model.trim="applicantRealName"
          class="input"
          maxlength="100"
          placeholder="申请人真实姓名（可选）"
        />
        <textarea
          v-model.trim="applicantMessage"
          class="textarea"
          maxlength="500"
          placeholder="申请理由（可选，建议说明与该家庭的关系）"
        />
        <button class="button" :disabled="submitting" :loading="submitting" @click="submitApplication">
          提交加入申请
        </button>
      </view>
      <view v-else class="result-card">
        <text class="success">申请已提交。</text>
        <text class="muted">申请 ID：{{ submittedRequest.requestId }}</text>
        <text class="muted">当前状态：{{ submittedRequest.requestStatus }}</text>
        <button class="button secondary" @click="go('/pages/join/my')">查看我的加入申请</button>
      </view>
      <text v-if="submitError" class="error">{{ submitError }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPublicFamilyDetail } from '@/api/families'
import { createJoinRequest } from '@/api/joinRequests'
import { useSessionStore } from '@/stores/session'
import type { JoinRequest, PublicFamily } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<PublicFamily | null>(null)
const applicantRealName = ref('')
const applicantMessage = ref('')
const submittedRequest = ref<JoinRequest | null>(null)
const loading = ref(false)
const submitting = ref(false)
const familyError = ref('')
const submitError = ref('')

function go(url: string) {
  uni.navigateTo({ url })
}

function currentRoute() {
  return `/pages/join/apply?familyId=${encodeURIComponent(familyId.value)}`
}

function requireLogin() {
  session.requireLogin(currentRoute())
}

async function loadFamily() {
  if (!familyId.value) {
    familyError.value = '缺少 familyId。'
    return
  }
  loading.value = true
  familyError.value = ''
  submitError.value = ''
  try {
    family.value = await getPublicFamilyDetail(familyId.value)
  } catch (error) {
    family.value = null
    familyError.value = apiErrorMessage(error, '家庭公开信息加载失败。')
  } finally {
    loading.value = false
  }
}

async function submitApplication() {
  if (!session.requireLogin(currentRoute())) return
  submitting.value = true
  submitError.value = ''
  try {
    submittedRequest.value = await createJoinRequest(familyId.value, {
      applicantRealName: applicantRealName.value || undefined,
      applicantMessage: applicantMessage.value || undefined
    })
  } catch (error) {
    submitError.value = apiErrorMessage(error, '加入申请提交失败。')
  } finally {
    submitting.value = false
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  session.restoreSession()
  loadFamily()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.notice,
.error,
.success {
  display: block;
  margin-top: 18rpx;
  font-size: 24rpx;
}

.notice {
  color: #6b7280;
}

.error {
  color: #c0392b;
}

.success {
  color: #2f6b57;
  font-weight: 600;
}

.result-card {
  margin-top: 20rpx;
}

.button[disabled] {
  opacity: 0.55;
}
</style>
