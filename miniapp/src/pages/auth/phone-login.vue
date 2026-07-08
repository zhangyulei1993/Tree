<template>
  <view class="archive-page auth-page">
    <view class="auth-head archive-page-head">
      <view>
        <text class="archive-kicker">Backup Login</text>
        <text class="archive-title">手机号备用登录</text>
        <text class="archive-subtitle">使用已绑定的手机号与密码登录，作为微信登录的备用方式</text>
      </view>
      <view class="archive-seal">备</view>
    </view>

    <view class="auth-folio">
      <view class="auth-mark">
        <text class="auth-brand">Tree</text>
        <text class="auth-product">{{ productName }}</text>
      </view>
      <view class="auth-folio-copy">
        <text class="auth-tagline">备用入册</text>
        <text class="auth-intro">仅在无法使用微信时使用。需先在资料页设置手机号与登录密码。</text>
      </view>
    </view>

    <view class="archive-form-panel auth-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">备用登录</text>
        <text class="archive-section-subtitle">输入已绑定的手机号与登录密码</text>
      </view>

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

      <view class="consent-wrap">
        <AuthLegalConsent v-model="legalAccepted" />
      </view>

      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>

      <MiniButton
        class="login-action"
        :disabled="submitting"
        :loading="submitting"
        @click="submitLogin"
      >
        登录
      </MiniButton>
    </view>

    <view class="auth-foot">
      <text class="auth-foot-hint">推荐使用微信快捷登录。</text>
      <text class="auth-switch-link" @click="goWechatLogin">返回微信登录</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import AuthLegalConsent from '@/components/legal/AuthLegalConsent.vue'
import { ensurePrivacyConsentForLogin } from '@/features/legal/privacyConsent'
import { LEGAL_PRODUCT_NAME } from '@/features/legal/legalMeta'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const productName = LEGAL_PRODUCT_NAME
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
.auth-page {
  padding-top: 48rpx;
}

.auth-folio {
  display: flex;
  align-items: stretch;
  gap: 24rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.42), transparent 38%),
    rgba(255, 252, 245, 0.52);
  padding: 24rpx 0;
}

.auth-mark {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 128rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-line-strong);
  background:
    radial-gradient(circle at 35% 28%, rgba(255, 255, 255, 0.9), transparent 34rpx),
    #eadfc8;
  padding: 18rpx 8rpx;
}

.auth-brand {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 34rpx;
  font-weight: 800;
  line-height: 1.2;
}

.auth-product {
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 18rpx;
  line-height: 1.35;
  text-align: center;
}

.auth-folio-copy {
  flex: 1;
  min-width: 0;
}

.auth-tagline {
  display: block;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 34rpx;
  font-weight: 800;
  line-height: 1.35;
}

.auth-intro {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.55;
}

.auth-panel {
  margin-bottom: 20rpx;
}

.consent-wrap {
  margin-top: 22rpx;
}

.consent-wrap :deep(.auth-legal-consent) {
  border-top: 1rpx solid rgba(168, 59, 45, 0.24);
  border-bottom: 1rpx solid rgba(168, 59, 45, 0.16);
  border-radius: 0;
  background: rgba(168, 59, 45, 0.05);
  padding: 18rpx 0;
}

.consent-wrap :deep(.consent-text) {
  color: var(--archive-ink);
  font-size: 25rpx;
  line-height: 1.65;
}

.consent-wrap :deep(.consent-link) {
  color: var(--archive-cinnabar);
  font-weight: 650;
}

.login-action {
  margin-top: 22rpx;
  align-self: stretch;
}

.auth-foot {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  align-items: center;
  padding: 8rpx 0 24rpx;
}

.auth-foot-hint {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.5;
}

.auth-switch-link {
  display: inline-flex;
  min-height: 56rpx;
  align-items: center;
  color: var(--archive-blue);
  font-size: 24rpx;
  font-weight: 650;
  line-height: 1.4;
  text-decoration: underline;
  text-underline-offset: 4rpx;
}
</style>
