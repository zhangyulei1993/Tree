<template>
  <view class="tree-page auth-page">
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">手机号登录</text>
      <text class="tree-auth-brand-desc">使用已绑定的手机号与密码登录，作为微信登录的备用方式</text>
    </view>

    <view class="tree-auth-card">
      <text class="tree-field-label">手机号</text>
      <input
        v-model.trim="phone"
        class="tree-input"
        type="number"
        maxlength="11"
        placeholder="请输入手机号"
      />

      <text class="tree-field-label">登录密码</text>
      <input
        v-model="password"
        class="tree-input"
        password
        maxlength="64"
        placeholder="请输入登录密码"
      />

      <AuthLegalConsent v-model="legalAccepted" class="phone-legal-consent" />

      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>

      <view class="btn-stack">
        <MiniButton :disabled="submitting" :loading="submitting" @click="submitLogin">
          登录
        </MiniButton>
      </view>

      <view class="tree-auth-foot auth-switch">
        <text class="tree-muted">推荐使用微信快捷登录。</text>
        <text class="auth-switch-link" @click="goWechatLogin">返回微信登录</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import AuthLegalConsent from '@/components/legal/AuthLegalConsent.vue'
import { ensurePrivacyConsentForLogin } from '@/features/legal/privacyConsent'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const phone = ref('')
const password = ref('')
const submitting = ref(false)
const errorMessage = ref('')
const legalAccepted = ref(false)

function validatePhone(value: string) {
  return /^1[3-9]\d{9}$/.test(value)
}

async function submitLogin() {
  if (!legalAccepted.value) {
    errorMessage.value = '请先阅读并同意用户协议与隐私政策。'
    return
  }
  const phoneValue = phone.value.trim()
  if (!validatePhone(phoneValue)) {
    errorMessage.value = '请输入有效的手机号。'
    return
  }
  if (!password.value) {
    errorMessage.value = '请输入登录密码。'
    return
  }

  errorMessage.value = ''
  submitting.value = true
  try {
    await ensurePrivacyConsentForLogin()
    await session.loginWithPhone({ phone: phoneValue, password: password.value })
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => session.routeAfterAuth(), 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '登录失败，请检查手机号与密码。')
  } finally {
    submitting.value = false
  }
}

function goWechatLogin() {
  uni.redirectTo({ url: '/pages/auth/wechat-login' })
}
</script>

<style scoped>
.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 22rpx;
}

.phone-legal-consent {
  margin-top: 22rpx;
}

.auth-page {
  background: #f3f6f3;
}

.tree-auth-card {
  margin-bottom: 24rpx;
  border: 1rpx solid rgba(216, 229, 220, 0.86);
  border-radius: var(--tree-radius-lg);
  background: rgba(250, 252, 251, 0.96);
  overflow: hidden;
  box-shadow: none;
}

.auth-switch {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  align-items: center;
}

.auth-switch-link {
  display: inline-flex;
  min-height: 60rpx;
  align-items: center;
  color: var(--tree-green);
  font-size: 26rpx;
  font-weight: 600;
}
</style>
