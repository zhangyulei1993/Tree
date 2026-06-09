<template>
  <view class="page">
    <view class="card">
      <text class="title">加入家庭申请</text>
      <text class="muted">申请加入张氏家族，mock 流程不提交真实后端。</text>
      <view v-if="session.state === 'guest'">
        <button class="button" @click="go('/pages/auth/wechat-login')">先登录</button>
      </view>
      <view v-else-if="!session.isPhoneBound">
        <button class="button" @click="go('/pages/auth/bind-phone')">先绑定手机号</button>
      </view>
      <view v-else>
        <textarea
          v-model="reason"
          class="textarea"
          maxlength="300"
          placeholder="填写申请理由"
          @input="submitted = false"
        />
        <text v-if="showReasonError" class="form-error">请填写申请理由</text>
        <button class="button" :disabled="!canSubmit" @click="submitApplication">
          提交申请
        </button>
      </view>
      <text v-if="submitted" class="muted">申请已提交，状态：PENDING。</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const reason = ref('')
const submitted = ref(false)
const showReasonError = computed(() => !submitted.value && reason.value.length > 0 && !reason.value.trim())
const canSubmit = computed(() => reason.value.trim().length > 0 && !submitted.value)

function go(url: string) {
  uni.navigateTo({ url })
}

function submitApplication() {
  if (!reason.value.trim()) {
    uni.showToast({ title: '请填写申请理由', icon: 'none' })
    return
  }

  submitted.value = true
}
</script>

<style scoped>
.form-error {
  display: block;
  margin: -12rpx 0 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.button[disabled] {
  opacity: 0.45;
}
</style>
