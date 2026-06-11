<template>
  <view class="page">
    <view class="card">
      <text class="title">微信快捷登录</text>
      <template v-if="isRealApiMode">
        <text class="muted">当前暂不支持微信一键登录，请使用手机号登录。</text>
        <button class="button" @click="go('/pages/auth/phone-login')">使用手机号登录</button>
        <button class="button secondary" @click="go('/pages/auth/register-phone')">手机号注册</button>
      </template>
      <template v-else>
        <text class="muted">当前为本地体验模式，可先体验微信登录流程。</text>
        <button class="button" @click="login">体验微信登录</button>
        <button class="button secondary" @click="go('/pages/auth/bind-phone')">去绑定手机号</button>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { isRealApiMode } from '@/api/client'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()

function login() {
  if (isRealApiMode) return
  session.mockWechatLogin()
  uni.showToast({ title: '登录成功', icon: 'none' })
  setTimeout(() => {
    uni.switchTab({ url: '/pages/me/index' })
  }, 500)
}

function go(url: string) {
  uni.navigateTo({ url })
}
</script>
