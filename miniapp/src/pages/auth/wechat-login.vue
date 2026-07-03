<template>
  <view class="tree-page auth-page">
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">微信登录</text>
      <text class="tree-auth-brand-desc">使用微信授权，快速进入家脉亲缘</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <AuthLegalConsent v-model="legalAccepted" />
      <template v-if="isRealApiMode && isMpWeixin">
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
        <view class="btn-stack">
          <MiniButton :disabled="submitting || !legalAccepted" :loading="submitting" @click="loginWithWechat">
            微信登录
          </MiniButton>
        </view>
      </template>
      <template v-else-if="isRealApiMode">
        <text class="tree-muted">{{ h5LoginHint }}</text>
      </template>
      <template v-else>
        <text class="tree-muted">本地体验模式下，可先体验登录流程。</text>
        <view class="btn-stack">
          <MiniButton @click="mockLogin(false)">微信快捷登录（待完善资料）</MiniButton>
          <MiniButton variant="secondary" @click="mockLogin(true)">微信快捷登录（已完善资料）</MiniButton>
        </view>
      </template>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { apiErrorMessage, isRealApiMode } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import AuthLegalConsent from '@/components/legal/AuthLegalConsent.vue'
import { ensurePrivacyConsentForLogin } from '@/features/legal/privacyConsent'
import { isMpWeixinPlatform, resolveWechatLoginUnsupportedMessage } from '@/features/session/wechatLogin'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
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
</script>

<style scoped>
.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 22rpx;
}

.auth-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.16), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.10), transparent 300rpx);
}

.tree-auth-card {
  border-color: rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.10), transparent 180rpx),
    rgba(255, 255, 255, 0.94);
  box-shadow: 0 18rpx 44rpx rgba(24, 54, 83, 0.08);
}
</style>
