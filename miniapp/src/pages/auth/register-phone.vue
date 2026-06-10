<template>
  <view class="page">
    <view class="card">
      <text class="title">手机号注册</text>
      <text class="muted">验证码由后端发送。页面不会预置或保存验证码。</text>
      <input
        v-model.trim="phone"
        class="input"
        type="number"
        maxlength="11"
        placeholder="请输入手机号"
      />
      <view class="code-row">
        <input
          v-model.trim="code"
          class="input code-input"
          type="number"
          maxlength="8"
          placeholder="验证码"
        />
        <button
          class="code-button"
          :disabled="sending || cooldown > 0"
          :loading="sending"
          @click="send"
        >
          {{ cooldown > 0 ? `${cooldown}s` : '发送验证码' }}
        </button>
      </view>
      <input v-model.trim="nickname" class="input" maxlength="30" placeholder="昵称（可选）" />
      <input
        v-model="password"
        class="input"
        password
        maxlength="64"
        placeholder="设置密码（至少 6 位）"
      />
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
      <text v-if="sendResult" class="success">{{ sendResult }}</text>
      <button class="button" :disabled="submitting" :loading="submitting" @click="submit">
        注册并登录
      </button>
      <button class="button secondary" :disabled="submitting" @click="goLogin">已有账号，去登录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onUnmounted, ref } from 'vue'

import { sendCode } from '@/api/auth'
import { apiErrorMessage } from '@/api/client'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const phone = ref('')
const code = ref('')
const nickname = ref('')
const password = ref('')
const sending = ref(false)
const submitting = ref(false)
const cooldown = ref(0)
const errorMessage = ref('')
const sendResult = ref('')
let timer: ReturnType<typeof setInterval> | undefined

async function send() {
  errorMessage.value = ''
  sendResult.value = ''
  if (!/^1\d{10}$/.test(phone.value)) {
    errorMessage.value = '请输入正确的 11 位手机号。'
    return
  }

  sending.value = true
  try {
    const result = await sendCode(phone.value, 'REGISTER')
    cooldown.value = result.cooldownSeconds
    sendResult.value = '验证码已发送，请查看开发环境响应或短信通道。'
    timer = setInterval(() => {
      cooldown.value -= 1
      if (cooldown.value <= 0 && timer) {
        clearInterval(timer)
        timer = undefined
      }
    }, 1000)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '验证码发送失败。')
  } finally {
    sending.value = false
  }
}

async function submit() {
  errorMessage.value = ''
  if (!/^1\d{10}$/.test(phone.value)) {
    errorMessage.value = '请输入正确的 11 位手机号。'
    return
  }
  if (!code.value) {
    errorMessage.value = '请输入验证码。'
    return
  }
  if (password.value.length < 6) {
    errorMessage.value = '密码至少需要 6 位。'
    return
  }

  submitting.value = true
  try {
    await session.register({
      phone: phone.value,
      code: code.value,
      password: password.value,
      nickname: nickname.value || undefined
    })
    uni.showToast({ title: '注册成功', icon: 'success' })
    setTimeout(() => session.finishLogin(), 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '注册失败。')
  } finally {
    submitting.value = false
  }
}

function goLogin() {
  uni.navigateBack({
    fail: () => uni.reLaunch({ url: '/pages/auth/phone-login' })
  })
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.code-row {
  display: flex;
  align-items: stretch;
  gap: 16rpx;
}

.code-input {
  flex: 1;
}

.code-button {
  width: 220rpx;
  margin-top: 16rpx;
  border-radius: 16rpx;
  background: #2f6b57;
  color: #fff;
  font-size: 24rpx;
}

.error,
.success {
  display: block;
  margin-top: 16rpx;
  font-size: 24rpx;
}

.error {
  color: #c0392b;
}

.success {
  color: #2f6b57;
}

.button[disabled],
.code-button[disabled] {
  opacity: 0.55;
}
</style>
