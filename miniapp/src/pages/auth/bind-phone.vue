<template>
  <view class="tree-page auth-page">
    <MiniBackHome />
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">{{ pageTitle }}</text>
      <text class="tree-auth-brand-desc">{{ pageDesc }}</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <template v-if="isRealApiMode">
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
        <text v-if="showTestCodeHint" class="tree-field-hint">测试阶段验证码为 123456。</text>
        <text class="tree-field-label">登录密码</text>
        <input
          v-model="password"
          class="input"
          password
          maxlength="64"
          placeholder="设置密码（至少 6 位）"
        />
        <text class="tree-field-label">确认密码</text>
        <input
          v-model="confirmPassword"
          class="input"
          password
          maxlength="64"
          placeholder="再次输入密码"
        />
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
        <text v-if="sendResult" class="tree-field-success">{{ sendResult }}</text>
        <MiniButton class="btn-top" :disabled="submitting" :loading="submitting" @click="submit">
          {{ submitText }}
        </MiniButton>
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
import { computed, onUnmounted, ref } from 'vue'

import { onShow } from '@dcloudio/uni-app'

import { sendCode } from '@/api/auth'
import { apiErrorMessage, isRealApiMode } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const phone = ref('')
const code = ref('')
const password = ref('')
const confirmPassword = ref('')
const sending = ref(false)
const submitting = ref(false)
const cooldown = ref(0)
const errorMessage = ref('')
const sendResult = ref('')
const showTestCodeHint = import.meta.env.MODE !== 'production'
let timer: ReturnType<typeof setInterval> | undefined

const alreadyBound = computed(() => session.user?.phoneVerified === true)
const pageTitle = computed(() => (alreadyBound.value ? '更改手机号' : '绑定手机号'))
const pageDesc = computed(() =>
  alreadyBound.value
    ? '更换绑定手机号，并同步确认登录密码'
    : '绑定手机号并设置登录密码，可使用完整家庭功能与电脑网页登录'
)
const submitText = computed(() => (alreadyBound.value ? '完成更改' : '完成绑定'))

async function send() {
  errorMessage.value = ''
  sendResult.value = ''
  if (!/^1\d{10}$/.test(phone.value)) {
    errorMessage.value = '请输入正确的 11 位手机号。'
    return
  }

  sending.value = true
  try {
    const result = await sendCode(phone.value, 'BIND_PHONE')
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
    errorMessage.value = '密码至少 6 位。'
    return
  }
  if (password.value !== confirmPassword.value) {
    errorMessage.value = '两次输入的密码不一致。'
    return
  }

  submitting.value = true
  try {
    await session.bindPhone({
      phone: phone.value,
      code: code.value,
      password: password.value
    })
    uni.showToast({ title: '绑定成功', icon: 'success' })
    setTimeout(() => session.finishBind(), 300)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '绑定失败，请检查信息后重试。')
  } finally {
    submitting.value = false
  }
}

function bind() {
  session.mockBindPhone()
  uni.showToast({ title: '已绑定', icon: 'success' })
  setTimeout(() => session.finishBind(), 500)
}

onShow(() => {
  session.restoreSession()
  if (isRealApiMode && !session.isLoggedIn) {
    uni.reLaunch({ url: '/pages/auth/wechat-login' })
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.code-row {
  display: flex;
  gap: 16rpx;
  align-items: center;
}

.code-input {
  flex: 1;
}

.code-button {
  flex-shrink: 0;
}

.btn-top {
  margin-top: 22rpx;
}
</style>
