<template>
  <view class="tree-page auth-page tree-page-lineage">
    <MiniBackHome />
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-eyebrow">Tree</text>
      <text class="tree-auth-brand-title">注册账号</text>
      <text class="tree-auth-brand-desc">填写手机号与验证码，完成注册后即可登录</text>
    </view>

    <MiniCard variant="info" class="tree-auth-card paper-surface">
      <text class="tree-field-label">手机号</text>
      <input
        v-model.trim="phone"
        class="input"
        type="number"
        maxlength="11"
        placeholder="请输入手机号"
      />
      <text class="tree-field-label">验证码</text>
      <view class="code-row">
        <input
          v-model.trim="code"
          class="input code-input"
          type="number"
          maxlength="8"
          placeholder="验证码"
        />
        <MiniButton
          size="sm"
          :block="false"
          class="code-button"
          :disabled="sending || cooldown > 0"
          :loading="sending"
          @click="send"
        >
          {{ cooldown > 0 ? `${cooldown}s` : '发送验证码' }}
        </MiniButton>
      </view>
      <text class="tree-field-label">昵称（可选）</text>
      <input v-model.trim="nickname" class="input" maxlength="30" placeholder="昵称（可选）" />
      <text class="tree-field-label">密码</text>
      <input
        v-model="password"
        class="input"
        password
        maxlength="64"
        placeholder="设置密码（至少 6 位）"
      />
      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <text v-if="sendResult" class="tree-field-success">{{ sendResult }}</text>
      <view class="btn-stack">
        <MiniButton :disabled="submitting" :loading="submitting" @click="submit">
          注册并登录
        </MiniButton>
        <MiniButton variant="secondary" :disabled="submitting" @click="goLogin">
          已有账号，去登录
        </MiniButton>
      </view>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onUnmounted, ref } from 'vue'

import { sendCode } from '@/api/auth'
import { apiErrorMessage } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
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
    sendResult.value = '验证码已发送，请查看短信。'
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
  flex-shrink: 0;
  width: 220rpx;
  margin-top: 16rpx;
}

.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-top: 24rpx;
}
</style>
