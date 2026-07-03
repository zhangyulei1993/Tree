<template>
  <view class="tree-page">
    <MiniBackHome />
    <MiniCard v-if="loading">
      <view class="state-block">
        <text class="tree-muted">正在加载家庭信息...</text>
      </view>
    </MiniCard>

    <MiniCard v-else-if="familyError">
      <MiniEmptyState
        symbol="!"
        title="家庭信息加载失败"
        :description="familyError"
        action-text="重新加载"
        @action="loadFamily"
      />
    </MiniCard>

    <template v-else-if="family">
      <MiniNotice tone="security" title="隐私说明">
        你的申请信息仅家庭管理员可见，用于核实身份。我们不会向无关人员公开你的联系方式。
      </MiniNotice>

      <FamilyMiniCard
        :name="family.familyName"
        :surname="family.familySurname"
        :region="family.regionText || '未设置'"
        :desc="family.description || '该家庭暂未填写公开简介。'"
      />

      <MiniCard v-if="!session.isLoggedIn">
        <MiniNotice tone="warm" title="需要登录">
          请先登录，再提交加入申请。
        </MiniNotice>
        <MiniButton @click="requireLogin">微信登录</MiniButton>
      </MiniCard>

      <MiniCard v-else-if="!session.isProfileComplete">
        <MiniNotice tone="warm" title="需要完善资料">
          请先设置昵称，再提交加入申请。
        </MiniNotice>
        <MiniButton @click="requireProfileComplete">去完善资料</MiniButton>
      </MiniCard>

      <MiniCard v-else-if="alreadyMember" variant="soft">
        <MiniEmptyState
          symbol="✓"
          title="你已经是该家庭成员"
          description="无需重复提交申请，可以直接进入“我的家庭”查看。"
          action-text="查看我的家庭"
          @action="go('/pages/family/my')"
        />
      </MiniCard>

      <MiniCard v-else-if="submittedRequest?.requestStatus === 'PENDING'" variant="soft">
        <MiniEmptyState
          symbol="…"
          title="加入申请审核中"
          description="家庭管理员尚未处理，请勿重复提交。"
          action-text="查看或取消申请"
          @action="go('/pages/me/family-affairs?tab=requests')"
        />
      </MiniCard>

      <MiniCard v-else-if="!submittedRequest">
        <text class="form-title">填写申请信息</text>
        <text class="tree-weak form-hint">请如实填写，方便管理员快速审核。</text>
        <text class="tree-field-label">真实姓名（可选）</text>
        <input
          v-model.trim="applicantRealName"
          class="tree-input"
          maxlength="100"
          placeholder="你的真实姓名"
        />
        <text class="tree-field-label">申请人性别</text>
        <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onSelectGender">
          <view class="picker-field">{{ genderLabels[genderIndex] }}</view>
        </picker>
        <text class="tree-field-label">申请说明</text>
        <textarea
          v-model.trim="applicantMessage"
          class="tree-textarea"
          maxlength="500"
          placeholder="请简单说明你与该家庭或成员的关系，方便管理员审核"
        />
        <MiniButton :disabled="submitting" :loading="submitting" @click="submitApplication">
          提交加入申请
        </MiniButton>
      </MiniCard>

      <MiniCard v-else variant="soft">
        <MiniEmptyState
          symbol="✓"
          title="申请已提交"
          description="请等待家庭管理员审核。你可以在「我的加入申请」中查看审核进度。"
          action-text="查看我的加入申请"
          @action="go('/pages/me/family-affairs?tab=requests')"
        />
      </MiniCard>

      <text v-if="submitError" class="tree-field-error">{{ submitError }}</text>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage, pendingRouteKey } from '@/api/client'
import { getPublicFamilyDetail, listMyFamilies } from '@/api/families'
import { createJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyMiniCard from '@/components/family/FamilyMiniCard.vue'
import { useSessionStore } from '@/stores/session'
import type { Gender, JoinRequest, PublicFamily } from '@/types/api'

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
  submitting.value = true
  submitError.value = ''
  try {
    submittedRequest.value = await createJoinRequest(familyId.value, {
      applicantRealName: applicantRealName.value || undefined,
      applicantGender: applicantGender.value,
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
.state-block {
  padding: 32rpx 0;
  text-align: center;
}

.form-title {
  display: block;
  margin-bottom: 8rpx;
  color: #1f2937;
  font-size: 28rpx;
  font-weight: 700;
}

.form-hint {
  display: block;
  margin-bottom: 8rpx;
}

.picker-field {
  box-sizing: border-box;
  width: 100%;
  min-height: 84rpx;
  margin-bottom: 18rpx;
  padding: 22rpx 24rpx;
  border: 1rpx solid var(--tree-border-warm, #ebe4d6);
  border-radius: 16rpx;
  background: #fff;
  color: var(--tree-text, #1f2937);
  font-size: 26rpx;
}
</style>
