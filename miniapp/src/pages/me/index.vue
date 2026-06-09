<template>
  <view class="page">
    <view class="card">
      <text class="title">个人中心</text>
      <text class="muted">昵称：{{ session.user.nickname }}</text>
      <text class="muted">手机号：{{ session.user.maskedPhone }}</text>
      <text class="muted">状态：{{ session.state }}</text>
      <button class="button" @click="go('/pages/family/my')">我的家庭</button>
      <button v-if="!session.isLoggedIn" class="button" @click="go('/pages/auth/wechat-login')">去登录</button>
      <button v-else-if="!session.isPhoneBound" class="button" @click="go('/pages/auth/bind-phone')">绑定手机号</button>
      <button class="button secondary" @click="go('/pages/account/cancel')">注销账号</button>
    </view>
    <view class="card">
      <text class="section-title">协议与隐私</text>
      <view class="link-row" @click="go('/pages/legal/user-agreement')">
        <text>用户协议</text><text class="muted">查看</text>
      </view>
      <view class="link-row" @click="go('/pages/legal/privacy-policy')">
        <text>隐私政策</text><text class="muted">查看</text>
      </view>
      <button v-if="session.isLoggedIn" class="button secondary" @click="logout">退出 mock 登录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()

function go(url: string) {
  uni.navigateTo({ url })
}

function logout() {
  session.logout()
  uni.showToast({ title: '已退出', icon: 'none' })
}
</script>

<style scoped>
.link-row {
  display: flex;
  justify-content: space-between;
  padding: 22rpx 0;
  border-top: 1rpx solid #e5e0d6;
}
</style>
