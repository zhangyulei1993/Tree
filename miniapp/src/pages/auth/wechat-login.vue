<template>
  <view class="tree-page auth-page">
    <MiniBackHome />
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">微信登录</text>
      <text class="tree-auth-brand-desc">使用微信授权，快速进入家脉亲缘</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <template v-if="isRealApiMode">
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
        <view class="btn-stack">
          <MiniButton :disabled="submitting" :loading="submitting" @click="loginWithWechat">
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
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const submitting = ref(false)
const errorMessage = ref('')

async function loginWithWechat() {
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
</style>
