<template>
  <view class="page">
    <view class="card">
      <text class="title">Tree 家脉亲缘</text>
      <text class="muted">保存家族记忆，查看公开家族，处理邀请与加入申请。</text>
      <button class="button" @click="go('/pages/family/search')">搜索家庭</button>
      <button class="button secondary" @click="go('/pages/auth/wechat-login')">微信快捷登录</button>
    </view>
    <view class="card">
      <text class="section-title">公开家庭推荐</text>
      <view v-for="family in families" :key="family.id" class="family-row" @click="openFamily(family.id)">
        <text>{{ family.name }}</text>
        <text class="muted">{{ family.regionText }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { families } from '@/mock/data'

function go(url: string) {
  if (url === '/pages/family/search' || url === '/pages/me/index') {
    uni.switchTab({ url })
    return
  }
  uni.navigateTo({ url })
}

function openFamily(familyId: string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(familyId)}`
  })
}
</script>

<style scoped>
.family-row {
  padding: 18rpx 0;
  border-top: 1rpx solid #e5e0d6;
}
</style>
