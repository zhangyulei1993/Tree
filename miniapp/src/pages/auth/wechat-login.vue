<template>
  <view class="tree-page auth-page">
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">微信登录</text>
      <text class="tree-auth-brand-desc">使用微信授权，快速进入家脉亲缘</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <AuthLegalConsent v-model="legalAccepted" />
      <template v-if="isRealApiMode">
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
        <view class="btn-stack">
          <MiniButton :disabled="submitting || !legalAccepted" :loading="submitting" @click="loginWithWechat">
            微信登录
          </MiniButton>
          <MiniButton variant="secondary" :disabled="submitting" @click="go('/pages/auth/phone-login')">
            使用手机号密码登录
          </MiniButton>
        </view>
      </template>
      <template v-else>
        <text class="tree-muted">本地体验模式下，可先体验登录流程。</text>
        <view class="btn-stack">
          <MiniButton @click="mockLogin">微信快捷登录</MiniButton>
          <MiniButton variant="secondary" @click="go('/pages/auth/bind-phone')">
            去绑定手机号
          </MiniButton>
        </view>
      </template>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { apiErrorMessage, isRealApiMode } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import AuthLegalConsent from '@/components/legal/AuthLegalConsent.vue'
import { hasPrivacyConsent } from '@/features/legal/privacyConsent'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const submitting = ref(false)
const errorMessage = ref('')
const legalAccepted = ref(false)

async function loginWithWechat() {
  if (!legalAccepted.value) {
    errorMessage.value = '请先阅读并同意用户协议与隐私政策。'
    return
  }
  if (!hasPrivacyConsent()) {
    errorMessage.value = '请先在首页隐私提示中同意个人信息处理规则。'
    return
  }
  errorMessage.value = ''
  submitting.value = true
  try {
    const result = await session.loginWithWechat()
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => {
      if (result.user.phoneVerified) {
        session.finishLogin()
      } else {
        uni.navigateTo({ url: '/pages/auth/bind-phone' })
      }
    }, 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '微信登录失败，请稍后重试。')
  } finally {
    submitting.value = false
  }
}

function mockLogin() {
  session.mockWechatLogin()
  uni.showToast({ title: '登录成功', icon: 'none' })
  setTimeout(() => {
    uni.navigateTo({ url: '/pages/auth/bind-phone' })
  }, 500)
}

function go(url: string) {
  uni.navigateTo({ url })
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
