<template>
  <view class="tree-page auth-page">
    <MiniBackHome />
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">绑定手机号</text>
      <text class="tree-auth-brand-desc">完成验证后，可使用完整家庭功能</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <template v-if="isRealApiMode">
        <MiniNotice tone="info">微信手机号绑定即将开放，请使用手机号登录或注册。</MiniNotice>
        <view class="btn-stack">
          <MiniButton @click="go('/pages/auth/phone-login')">使用手机号登录</MiniButton>
          <MiniButton variant="secondary" @click="go('/pages/auth/register-phone')">
            手机号注册
          </MiniButton>
        </view>
      </template>
      <template v-else>
        <text class="tree-muted">请填写手机号并完成验证码校验。</text>
        <text class="tree-field-label">手机号</text>
        <input class="input" placeholder="请输入手机号" />
        <text class="tree-field-label">验证码</text>
        <input class="input" placeholder="请输入验证码" />
        <MiniButton class="btn-top" @click="bind">完成绑定</MiniButton>
      </template>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { isRealApiMode } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
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

<style scoped>
.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 22rpx;
}

.btn-top {
  margin-top: 22rpx;
}
</style>
