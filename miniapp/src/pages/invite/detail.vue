<template>
  <view class="page">
    <view class="card">
      <text class="title">邀请确认</text>
      <text class="muted">家庭：{{ invite.familyName }}</text>
      <text class="muted">成员：{{ invite.memberName }}</text>
      <text class="muted">邀请人：{{ invite.inviterName }}</text>
      <text class="muted">有效期至：{{ invite.expiresAt }}</text>
      <view v-if="session.state === 'guest'">
        <button class="button" @click="go('/pages/auth/wechat-login')">先微信登录</button>
      </view>
      <view v-else-if="!session.isPhoneBound">
        <button class="button" @click="go('/pages/auth/bind-phone')">先绑定手机号</button>
      </view>
      <view v-else>
        <button class="button" @click="result = '已接受邀请'">接受绑定</button>
        <button class="button secondary" @click="result = '已拒绝邀请'">拒绝</button>
      </view>
      <text v-if="result" class="muted">{{ result }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { invite } from '@/mock/data'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const result = ref('')

function go(url: string) {
  uni.navigateTo({ url })
}
</script>
