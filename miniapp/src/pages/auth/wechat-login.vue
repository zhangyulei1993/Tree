<template>
  <view class="archive-page auth-page">
    <view class="auth-head archive-page-head">
      <view>
        <text class="archive-kicker">WeChat Entry</text>
        <text class="archive-title">微信登录</text>
        <text class="archive-subtitle">使用微信账号进入 Tree</text>
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

      <view :class="['consent-wrap', { 'is-required': consentRequired }]">
        <AuthLegalConsent v-model="legalAccepted" @update:model-value="onLegalAcceptedChanged" />
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
        <text class="login-note">登录后可直接进入，头像和昵称都可以稍后补充。</text>
        <text class="login-note">首次使用也请先微信登录；进入添加、邀请等核心操作时再按需绑定手机号。</text>
        <view class="assist-login" @click="goAssistanceLogin"><text>亲友给了你协助码？</text><text>协助登录 ›</text></view>
      </template>
      <template v-else-if="isRealApiMode">
        <text class="platform-hint">{{ h5LoginHint }}</text>
      </template>
      <template v-else-if="showLocalMockActions">
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
      <template v-else>
        <text class="platform-hint">当前小程序包不是 real 模式构建，请使用 staging 或正式配置重新构建后再登录。</text>
      </template>
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
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { isMpWeixinPlatform, resolveWechatLoginUnsupportedMessage } from '@/features/session/wechatLogin'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const productName = LEGAL_PRODUCT_NAME
const submitting = ref(false)
const errorMessage = ref('')
const legalAccepted = ref(false)
const consentRequired = ref(false)
const isMpWeixin = isMpWeixinPlatform()
const h5LoginHint = computed(() => resolveWechatLoginUnsupportedMessage('h5') || '')
const showLocalMockActions = computed(() => !isRealApiMode && !isMpWeixin)

async function loginWithWechat() {
  if (!ensureLegalAccepted()) return
  errorMessage.value = ''
  submitting.value = true
  try {
    await ensurePrivacyConsentForLogin()
    await session.loginWithWechat()
    markNewUserGuideMilestone('LOGIN_DONE')
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => session.routeAfterAuth(), 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '微信登录失败，请稍后重试。')
  } finally {
    submitting.value = false
  }
}

function goAssistanceLogin() { uni.navigateTo({ url: '/pages/auth/assistance-login' }) }

function ensureLegalAccepted() {
  if (legalAccepted.value) {
    consentRequired.value = false
    return true
  }
  consentRequired.value = true
  errorMessage.value = '请先勾选《用户协议》和《隐私政策》，再继续登录。'
  uni.showToast({ title: '请先勾选用户协议和隐私政策', icon: 'none' })
  return false
}

function onLegalAcceptedChanged(value: boolean) {
  if (value) {
    consentRequired.value = false
    errorMessage.value = ''
  }
}

function mockLogin(withProfile: boolean) {
  session.mockWechatLogin(withProfile)
  markNewUserGuideMilestone('LOGIN_DONE')
  uni.showToast({ title: '登录成功', icon: 'none' })
  setTimeout(() => session.routeAfterAuth(), 300)
}

</script>

<style scoped>
.auth-page {
  padding-top: 48rpx;
}

.phone-login-action {
  width: 100%;
  min-height: 88rpx;
  margin-top: 18rpx;
  border: 1rpx solid var(--archive-blue);
  border-radius: 10rpx;
  background: transparent;
  color: var(--archive-blue);
  font-size: 28rpx;
  font-weight: 700;
}

.phone-login-action::after { border: 0; }
.phone-login-action[disabled] { opacity: .55; }
.phone-login-note {
  display: block;
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.phone-login-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  min-height: 68rpx;
  margin-top: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
}
.assist-login { display:flex; align-items:center; justify-content:space-between; margin-top:26rpx; padding-top:22rpx; border-top:1rpx solid var(--archive-line); color:var(--archive-ink-soft); font-size:24rpx; }
.assist-login text:last-child { color:var(--archive-cinnabar); font-weight:650; }

.phone-login-trigger:active,
.phone-login-collapse:active { opacity: .68; }
.phone-login-trigger-arrow { color: var(--archive-cinnabar); font-size: 32rpx; line-height: 1; }
.phone-login-panel {
  margin-top: 18rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 18rpx;
}
.phone-login-title { display: block; color: var(--archive-blue); font-size: 26rpx; font-weight: 700; }
.phone-login-collapse {
  display: block;
  margin-top: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  text-align: center;
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

.consent-wrap.is-required :deep(.auth-legal-consent) {
  border-color: var(--archive-cinnabar);
  box-shadow: 0 0 0 3rpx rgba(168, 59, 45, 0.12);
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

.login-note {
  display: block;
  margin-top: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.5;
  text-align: center;
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

</style>
