<template>
  <view class="page">
    <view class="card">
      <text class="title">微信快捷登录</text>
      <text class="muted">当前为 mock 登录，不调用真实微信接口，不配置 AppSecret。</text>
      <button class="button" @click="login">模拟微信登录</button>
      <button class="button secondary" @click="go('/pages/auth/bind-phone')">去绑定手机号</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()

function login() {
  session.mockWechatLogin()
  uni.showToast({ title: 'mock 登录成功', icon: 'none' })
  setTimeout(() => {
    uni.switchTab({ url: '/pages/me/index' })
  }, 500)
}

function go(url: string) {
  uni.navigateTo({ url })
}
</script>
