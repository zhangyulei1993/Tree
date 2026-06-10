<template>
  <view class="page">
    <view class="card">
      <text class="title">绑定手机号</text>
      <template v-if="isRealApiMode">
        <text class="muted">M18-1A 暂不提供真实微信手机号绑定，请使用手机号账号登录或注册。</text>
        <button class="button" @click="go('/pages/auth/phone-login')">手机号登录</button>
        <button class="button secondary" @click="go('/pages/auth/register-phone')">手机号注册</button>
      </template>
      <template v-else>
        <text class="muted">手机号和验证码均为 UI 示例，不发送真实验证码。</text>
        <input class="input" placeholder="138****0000" />
        <input class="input" placeholder="验证码" />
        <button class="button" @click="bind">完成 mock 绑定</button>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { isRealApiMode } from '@/api/client'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()

function bind() {
  if (isRealApiMode) return
  session.mockBindPhone()
  uni.showToast({ title: '已绑定', icon: 'success' })
  setTimeout(() => {
    uni.switchTab({ url: '/pages/me/index' })
  }, 500)
}

function go(url: string) {
  uni.navigateTo({ url })
}
</script>
