<template>
  <view class="page">
    <view class="card">
      <text class="title">个人中心</text>
      <template v-if="session.isLoggedIn && session.user">
        <text class="muted">昵称：{{ session.user.nickname || '未设置' }}</text>
        <text class="muted">手机号：{{ maskedPhone }}</text>
        <text class="muted">账号状态：{{ accountStatusText }}</text>
        <text class="muted">手机号验证：{{ session.user.phoneVerified ? '已验证' : '未验证' }}</text>
        <button class="button" @click="go('/pages/family/my')">我的家庭</button>
        <button class="button secondary" @click="go('/pages/invite/my')">我的邀请</button>
        <button class="button secondary" @click="go('/pages/join/my')">我的加入申请</button>
      </template>
      <template v-else>
        <text class="muted">当前未登录。请使用手机号登录或注册账号。</text>
        <button class="button" @click="go('/pages/auth/phone-login')">手机号登录</button>
        <button class="button secondary" @click="go('/pages/auth/register-phone')">注册账号</button>
      </template>
    </view>
    <view class="card">
      <text class="section-title">协议与隐私</text>
      <view class="link-row" @click="go('/pages/legal/user-agreement')">
        <text>用户协议</text><text class="muted">查看</text>
      </view>
      <view class="link-row" @click="go('/pages/legal/privacy-policy')">
        <text>隐私政策</text><text class="muted">查看</text>
      </view>
      <button
        v-if="session.isLoggedIn"
        class="button secondary"
        :disabled="loggingOut"
        :loading="loggingOut"
        @click="logout"
      >
        退出登录
      </button>
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const loggingOut = ref(false)
const errorMessage = ref('')
const maskedPhone = computed(() => {
  const phone = session.user?.phone || ''
  if (!phone) return '未绑定'
  if (phone.includes('*')) return phone
  return phone.length === 11 ? `${phone.slice(0, 3)}****${phone.slice(-4)}` : phone
})
const accountStatusText = computed(() => {
  switch (session.user?.status) {
    case 'ACTIVE':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'CANCELLED':
      return '已注销'
    case 'PENDING_PHONE_BIND':
      return '待绑定手机号'
    default:
      return session.user?.status || '未知'
  }
})

function go(url: string) {
  uni.navigateTo({ url })
}

async function logout() {
  loggingOut.value = true
  errorMessage.value = ''
  try {
    await session.logout()
    uni.showToast({ title: '已退出', icon: 'none' })
    setTimeout(() => uni.reLaunch({ url: '/pages/auth/phone-login' }), 200)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '退出请求失败，本地登录状态已清理。')
  } finally {
    loggingOut.value = false
  }
}

onShow(() => session.restoreSession())
</script>

<style scoped>
.link-row {
  display: flex;
  justify-content: space-between;
  padding: 22rpx 0;
  border-top: 1rpx solid #e5e0d6;
}

.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}
</style>
