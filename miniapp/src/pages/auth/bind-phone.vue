<template>
  <view class="page">
    <view class="card">
      <text class="title">绑定手机号</text>
      <text class="muted">手机号和验证码均为 UI 示例，不发送真实验证码。</text>
      <input class="input" placeholder="138****0000" />
      <input class="input" placeholder="验证码" />
      <button class="button" @click="bind">完成 mock 绑定</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()

function bind() {
  session.mockBindPhone()
  uni.showToast({ title: '已绑定', icon: 'success' })
  setTimeout(() => {
    uni.switchTab({ url: '/pages/me/index' })
  }, 500)
}
</script>
