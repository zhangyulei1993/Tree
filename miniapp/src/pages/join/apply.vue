<template>
  <view class="archive-page apply-page">
    <MiniBackHome />

    <view v-if="loading" class="apply-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载家庭信息..." />
    </view>

    <view v-else-if="familyError" class="apply-state archive-form-panel">
      <MiniEmptyState
        symbol="!"
        title="家庭信息加载失败"
        :description="familyError"
        action-text="重新加载"
        @action="loadFamily"
      />
    </view>

    <template v-else-if="family">
      <view class="apply-head archive-page-head">
        <view>
          <text class="archive-kicker">Join Application</text>
          <text class="archive-title">提交加入申请</text>
          <text class="archive-subtitle">向家庭管理员提交加入申请，审核通过后可进入家庭</text>
        </view>
        <view class="archive-seal">申</view>
      </view>

      <view class="apply-guide archive-list">
        <view class="archive-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">隐私说明</text>
            <text class="archive-row-desc">
              你的申请信息仅家庭管理员可见，用于核实身份。我们不会向无关人员公开你的联系方式。
            </text>
          </view>
        </view>
        <view class="archive-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">审核流程</text>
            <text class="archive-row-desc">
              提交后可在「我提交的加入申请」查看进度；家庭管理员在「收到的加入申请」中审核。
            </text>
          </view>
        </view>
      </view>

      <view class="apply-family-folio">
        <view class="apply-family-copy">
          <text class="apply-surname">{{ family.familySurname }}氏</text>
          <text class="apply-name">{{ family.familyName }}</text>
          <text class="apply-desc">{{ family.description || '该家庭暂未填写公开简介。' }}</text>
          <view class="apply-meta">
            <text class="archive-chip">{{ family.regionText || '未设置地区' }}</text>
            <text class="archive-chip">公开家庭</text>
          </view>
        </view>
        <view class="apply-spine archive-book-spine">
          <text>加</text>
          <text>入</text>
          <text>册</text>
        </view>
      </view>

      <view v-if="!session.isLoggedIn" class="apply-state archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">需要登录</text>
          <text class="archive-section-subtitle">请先登录，再提交加入申请</text>
        </view>
        <MiniButton class="apply-action" @click="requireLogin">去登录</MiniButton>
      </view>

      <view v-else-if="!session.isProfileComplete" class="apply-state archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">需要完善资料</text>
          <text class="archive-section-subtitle">请先设置昵称，再提交加入申请</text>
        </view>
        <MiniButton class="apply-action" @click="requireProfileComplete">去完善资料</MiniButton>
      </view>

      <view v-else-if="alreadyMember" class="apply-state archive-form-panel">
        <MiniEmptyState
          symbol="✓"
          title="你已经是该家庭成员"
          description="无需重复提交申请，可以直接进入「我的家庭」查看。"
          action-text="查看我的家庭"
          @action="go('/pages/family/my')"
        />
      </view>

      <view v-else-if="submittedRequest?.requestStatus === 'PENDING'" class="apply-state archive-form-panel">
        <MiniEmptyState
          symbol="…"
          title="加入申请审核中"
          description="家庭管理员尚未处理，请勿重复提交。"
          action-text="查看或取消申请"
          @action="go('/pages/me/family-affairs?tab=requests')"
        />
      </view>

      <view v-else-if="!submittedRequest" class="apply-form archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">填写申请信息</text>
          <text class="archive-section-subtitle">请如实填写，方便管理员快速审核</text>
        </view>

        <text class="tree-field-label">真实姓名（可选）</text>
        <input
          v-model.trim="applicantRealName"
          class="tree-input"
          maxlength="100"
          placeholder="你的真实姓名"
        />

        <text class="tree-field-label">申请人性别</text>
        <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onSelectGender">
          <view class="field-picker">{{ genderLabels[genderIndex] }}</view>
        </picker>

        <text class="tree-field-label">申请说明</text>
        <textarea
          v-model.trim="applicantMessage"
          class="tree-textarea"
          maxlength="500"
          placeholder="请简单说明你与该家庭或成员的关系，方便管理员审核"
        />

        <text v-if="submitError" class="tree-field-error">{{ submitError }}</text>

        <MiniButton
          class="apply-action"
          size="sm"
          :disabled="submitting"
          :loading="submitting"
          @click="submitApplication"
        >
          提交加入申请
        </MiniButton>
      </view>

      <view v-else class="apply-state archive-form-panel">
        <MiniEmptyState
          symbol="✓"
          title="申请已提交"
          description="请等待家庭管理员审核。你可以在「我的加入申请」中查看审核进度。"
          action-text="查看我的加入申请"
          @action="go('/pages/me/family-affairs?tab=requests')"
        />
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPublicFamilyDetail, listMyFamilies } from '@/api/families'
import { createJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { useSessionStore } from '@/stores/session'
import type { Gender, JoinRequest, PublicFamily } from '@/types/api'
import { optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const family = ref<PublicFamily | null>(null)
const applicantRealName = ref('')
const applicantGender = ref<Gender>('MALE')
const genderIndex = ref(0)
const applicantMessage = ref('')
const submittedRequest = ref<JoinRequest | null>(null)
const alreadyMember = ref(false)
const loading = ref(false)
const submitting = ref(false)
const familyError = ref('')
const submitError = ref('')

function go(url: string) {
  if (url === '/pages/family/my') {
    uni.switchTab({ url })
    return
  }
  uni.navigateTo({ url })
}

function currentRoute() {
  return `/pages/join/apply?familyId=${encodeURIComponent(familyId.value)}`
}

function requireLogin() {
  session.requireLogin(currentRoute())
}

function requireProfileComplete() {
  session.requireProfileComplete(currentRoute())
}

const genders: Gender[] = ['MALE', 'FEMALE']
const genderLabels = ['男', '女']

function onSelectGender(event: { detail: { value: number | string } }) {
  genderIndex.value = Number(event.detail.value) || 0
  applicantGender.value = genders[genderIndex.value] || 'MALE'
}

async function loadFamily() {
  if (!familyId.value) {
    familyError.value = '家庭信息缺失。'
    return
  }
  loading.value = true
  familyError.value = ''
  submitError.value = ''
  submittedRequest.value = null
  alreadyMember.value = false
  try {
    family.value = await getPublicFamilyDetail(familyId.value)
    if (session.isLoggedIn && session.isProfileComplete) {
      const [requests, myFamilies] = await Promise.all([
        listMyJoinRequests(),
        listMyFamilies()
      ])
      alreadyMember.value = myFamilies.some((item) => String(item.id) === familyId.value)
      submittedRequest.value = requests.find(
        (item) => String(item.familyId) === familyId.value && item.requestStatus === 'PENDING'
      ) || null
    }
  } catch (error) {
    family.value = null
    familyError.value = apiErrorMessage(error, '家庭公开信息加载失败。')
  } finally {
    loading.value = false
  }
}

async function submitApplication() {
  if (!session.requireProfileComplete(currentRoute())) return
  const validationMessage = validateTextFields([
    { value: applicantRealName.value, label: '真实姓名', kind: 'name', maxLength: 80 },
    { value: applicantMessage.value, label: '申请说明', kind: 'multiLine', maxLength: 300 }
  ])
  if (validationMessage) {
    submitError.value = validationMessage
    return
  }
  submitting.value = true
  submitError.value = ''
  try {
    submittedRequest.value = await createJoinRequest(familyId.value, {
      applicantRealName: optionalText(applicantRealName.value),
      applicantGender: applicantGender.value,
      applicantMessage: optionalText(applicantMessage.value)
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
.apply-page {
  padding-top: 28rpx;
}

.apply-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.apply-head {
  margin-bottom: 24rpx;
}

.apply-guide {
  margin-bottom: 24rpx;
}

.apply-family-folio {
  display: flex;
  min-height: 220rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.42), transparent 38%),
    rgba(255, 252, 245, 0.52);
}

.apply-family-copy {
  flex: 1;
  min-width: 0;
  padding: 28rpx 24rpx 28rpx 0;
}

.apply-surname {
  display: block;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  font-weight: 650;
  letter-spacing: 3rpx;
}

.apply-name {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 40rpx;
  font-weight: 800;
  line-height: 1.28;
}

.apply-desc {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.55;
}

.apply-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 14rpx;
}

.apply-spine {
  flex-shrink: 0;
}

.apply-state,
.apply-form {
  margin-bottom: 24rpx;
}

.apply-action {
  margin-top: 18rpx;
  align-self: flex-start;
}
</style>
