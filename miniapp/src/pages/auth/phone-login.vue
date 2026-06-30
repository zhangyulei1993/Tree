<template>
  <view class="tree-page auth-page">
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-eyebrow">Tree</text>
      <text class="tree-auth-brand-title">家脉亲缘</text>
      <text class="tree-auth-brand-desc">使用手机号登录，进入家庭空间</text>
    </view>

    <view class="auth-card auth-card-main">
      <text class="tree-field-label">手机号</text>
      <input
        v-model.trim="phone"
        class="input"
        type="number"
        maxlength="11"
        placeholder="请输入手机号"
      />
      <text class="tree-field-label">密码</text>
      <input
        v-model="password"
        class="input"
        password
        maxlength="64"
        placeholder="请输入密码"
      />
      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <view class="btn-stack">
        <MiniButton :disabled="submitting" :loading="submitting" @click="submit">
          登录
        </MiniButton>
        <MiniButton variant="secondary" :disabled="submitting" @click="goRegister">
          注册新账号
        </MiniButton>
      </view>
    </view>

    <view class="auth-card auth-card-foot">
      <MiniNotice tone="info">
        当前小程序暂不支持手机号一键登录。你可以使用微信登录后绑定手机号，或使用手机号密码登录。
      </MiniNotice>
      <MiniButton variant="secondary" class="btn-top" @click="goWechat">
        微信登录
      </MiniButton>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const phone = ref('')
const password = ref('')
const submitting = ref(false)
const errorMessage = ref('')

async function submit() {
  errorMessage.value = ''
  if (!/^1\d{10}$/.test(phone.value)) {
    errorMessage.value = '请输入正确的 11 位手机号。'
    return
  }
  if (!password.value) {
    errorMessage.value = '请输入密码。'
    return
  }

  submitting.value = true
  try {
    await session.login({ phone: phone.value, password: password.value })
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => session.finishLogin(), 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '登录失败，请检查手机号和密码。')
  } finally {
    submitting.value = false
  }
}

function goRegister() {
  uni.navigateTo({ url: '/pages/auth/register-phone' })
}

function goWechat() {
  uni.navigateTo({ url: '/pages/auth/wechat-login' })
}
</script>

<style scoped>
.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 58rpx;
}

.btn-top {
  margin-top: 16rpx;
}

.auth-page {
  overflow-x: hidden;
  background: linear-gradient(180deg, #f6efe5 0%, #f1f6f3 48%, #edf4f4 100%);
}

.auth-page::before,
.auth-page::after {
  display: none;
  content: none;
}

.auth-card {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  margin-bottom: 24rpx;
  border: 1rpx solid rgba(216, 229, 220, 0.95);
  border-radius: 32rpx;
  background: #f3faf7;
  padding: 30rpx 28rpx;
  overflow: hidden;
  box-shadow: 0 18rpx 44rpx rgba(24, 54, 83, 0.08);
}

.auth-card-foot {
  background: #f6faf8;
}
</style>
