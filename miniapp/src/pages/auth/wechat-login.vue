<template>
  <view class="archive-page auth-page">
    <view class="auth-head archive-page-head">
      <view>
        <text class="archive-kicker">WeChat Entry</text>
        <text class="archive-title">微信快捷登录</text>
        <text class="archive-subtitle">使用微信授权，快速进入家脉亲缘</text>
      </view>
      <view class="archive-seal">登</view>
    </view>

    <view class="auth-folio">
      <view class="auth-mark">
        <text class="auth-brand">Tree</text>
        <text class="auth-product">{{ productName }}</text>
      </view>
      <view class="auth-folio-copy">
        <text class="auth-tagline">入册登录</text>
        <text class="auth-intro">登录后可创建家庭、维护族谱、处理邀请与加入申请。</text>
      </view>
    </view>

    <view class="archive-form-panel auth-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">登录确认</text>
        <text class="archive-section-subtitle">请先阅读并同意用户协议与隐私政策</text>
      </view>

      <view class="consent-wrap">
        <AuthLegalConsent v-model="legalAccepted" />
      </view>

      <template v-if="isRealApiMode && isMpWeixin">
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
        <MiniButton
          class="login-action"
          :disabled="submitting"
          :loading="submitting"
          @click="loginWithWechat"
        >
          微信登录
        </MiniButton>
      </template>
      <template v-else-if="isRealApiMode">
        <text class="platform-hint">{{ h5LoginHint }}</text>
      </template>
      <template v-else>
        <text class="platform-hint">本地体验模式下，可先体验登录流程。</text>
        <view class="mock-actions">
          <MiniButton class="login-action" @click="mockLogin(false)">
            微信快捷登录（待完善资料）
          </MiniButton>
          <MiniButton variant="secondary" size="sm" @click="mockLogin(true)">
            微信快捷登录（已完善资料）
          </MiniButton>
        </view>
      </template>
    </view>

    <view v-if="isRealApiMode" class="auth-foot">
      <text class="auth-switch-link" @click="goPhoneLogin">使用手机号密码登录</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { apiErrorMessage, isRealApiMode } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import AuthLegalConsent from '@/components/legal/AuthLegalConsent.vue'
import { ensurePrivacyConsentForLogin } from '@/features/legal/privacyConsent'
import { LEGAL_PRODUCT_NAME } from '@/features/legal/legalMeta'
import { isMpWeixinPlatform, resolveWechatLoginUnsupportedMessage } from '@/features/session/wechatLogin'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const productName = LEGAL_PRODUCT_NAME
const submitting = ref(false)
const errorMessage = ref('')
const legalAccepted = ref(false)
const isMpWeixin = isMpWeixinPlatform()
const h5LoginHint = computed(() => resolveWechatLoginUnsupportedMessage('h5') || '')

async function loginWithWechat() {
  if (!legalAccepted.value) {
    errorMessage.value = '请先阅读并同意用户协议与隐私政策。'
    return
  }
  errorMessage.value = ''
  submitting.value = true
  try {
    await ensurePrivacyConsentForLogin()
    await session.loginWithWechat()
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => session.routeAfterAuth(), 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '微信登录失败，请稍后重试。')
  } finally {
    submitting.value = false
  }
}

function mockLogin(withProfile: boolean) {
  session.mockWechatLogin(withProfile)
  uni.showToast({ title: '登录成功', icon: 'none' })
  setTimeout(() => session.routeAfterAuth(), 300)
}

function goPhoneLogin() {
  uni.navigateTo({ url: '/pages/auth/phone-login' })
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
  margin-top: 8rpx;
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

.mock-actions {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 18rpx;
}

.platform-hint {
  display: block;
  margin-top: 18rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.6;
}

.auth-foot {
  display: flex;
  justify-content: center;
  padding: 8rpx 0 24rpx;
}

.auth-switch-link {
  display: inline-flex;
  min-height: 56rpx;
  align-items: center;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.4;
  text-decoration: underline;
  text-underline-offset: 4rpx;
}
</style>
