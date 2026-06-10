<template>
  <view class="page">
    <view class="card">
      <text class="title">绑定手机号</text>
      <template v-if="isRealApiMode">
        <text class="muted">当前暂不支持微信手机号绑定，请使用手机号登录或注册。</text>
        <button class="button" @click="go('/pages/auth/phone-login')">使用手机号登录</button>
        <button class="button secondary" @click="go('/pages/auth/register-phone')">手机号注册</button>
      </template>
      <template v-else>
        <text class="muted">当前为本地体验模式，可先体验手机号绑定流程。</text>
        <input class="input" placeholder="138****0000" />
        <input class="input" placeholder="验证码" />
        <button class="button" @click="bind">完成绑定</button>
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
