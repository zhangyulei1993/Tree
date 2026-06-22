<template>
  <view class="tree-page auth-page">
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">微信登录</text>
      <text class="tree-auth-brand-desc">使用微信授权，快速进入家脉亲缘</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <template v-if="isRealApiMode">
        <MiniNotice tone="info">微信一键登录即将开放，当前可使用手机号登录或注册。</MiniNotice>
        <view class="btn-stack">
          <MiniButton @click="go('/pages/auth/phone-login')">使用手机号登录</MiniButton>
          <MiniButton variant="secondary" @click="go('/pages/auth/register-phone')">
            手机号注册
          </MiniButton>
        </view>
      </template>
      <template v-else>
        <text class="tree-muted">本地体验模式下，可先体验登录流程。</text>
        <view class="btn-stack">
          <MiniButton @click="login">微信快捷登录</MiniButton>
          <MiniButton variant="secondary" @click="go('/pages/auth/bind-phone')">
            去绑定手机号
          </MiniButton>
        </view>
      </template>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { isRealApiMode } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
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

<style scoped>
.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 22rpx;
}
</style>
