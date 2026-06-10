<template>
  <view class="page">
    <view class="card">
      <text class="title">手机号登录</text>
      <text class="muted">登录 Tree，查看你的家庭、成员与私有家庭树。</text>
      <input
        v-model.trim="phone"
        class="input"
        type="number"
        maxlength="11"
        placeholder="请输入手机号"
      />
      <input
        v-model="password"
        class="input"
        password
        maxlength="64"
        placeholder="请输入密码"
      />
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
      <button class="button" :disabled="submitting" :loading="submitting" @click="submit">
        登录
      </button>
      <button class="button secondary" :disabled="submitting" @click="goRegister">注册新账号</button>
    </view>
    <view class="card">
      <text class="section-title">微信登录</text>
      <text class="muted">真实微信登录尚未启用，请先使用手机号登录。</text>
      <button class="button secondary" @click="goWechat">查看微信登录说明</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const phone = ref('')
const password = ref('')
const submitting = ref(false)
const errorMessage = ref('')

async function submit() {
  errorMessage.value = ''
  if (!/^1\d{10}$/.test(phone.value)) {
    errorMessage.value = '请输入正确的 11 位手机号。'
    return
  }
  if (!password.value) {
    errorMessage.value = '请输入密码。'
    return
  }

  submitting.value = true
  try {
    await session.login({ phone: phone.value, password: password.value })
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => session.finishLogin(), 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '登录失败，请检查手机号和密码。')
  } finally {
    submitting.value = false
  }
}

function goRegister() {
  uni.navigateTo({ url: '/pages/auth/register-phone' })
}

function goWechat() {
  uni.navigateTo({ url: '/pages/auth/wechat-login' })
}
</script>

<style scoped>
.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.button[disabled] {
  opacity: 0.55;
}
</style>
