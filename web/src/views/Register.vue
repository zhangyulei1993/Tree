<template>
  <PageShell>
    <section class="container auth-page">
      <form class="card auth-card" @submit.prevent="submit">
        <div>
          <span class="eyebrow">{{ session.mode === 'real' ? '真实 API' : 'Mock 模式' }}</span>
          <h1>注册账号</h1>
          <p class="muted">验证码不会写入页面文件或浏览器存储。</p>
        </div>
        <label>
          <span>手机号</span>
          <input v-model.trim="phone" class="field" inputmode="numeric" autocomplete="username" placeholder="请输入手机号" />
        </label>
        <div class="code-row">
          <label>
            <span>验证码</span>
            <input v-model.trim="code" class="field" inputmode="numeric" autocomplete="one-time-code" placeholder="请输入验证码" />
          </label>
          <button class="button secondary code-button" type="button" :disabled="sending || cooldown > 0" @click="requestCode">
            {{ cooldown > 0 ? `${cooldown}s` : sending ? '发送中...' : '发送验证码' }}
          </button>
        </div>
        <label>
          <span>密码</span>
          <input v-model="password" class="field" type="password" autocomplete="new-password" placeholder="至少 6 位" />
        </label>
        <label>
          <span>昵称（可选）</span>
          <input v-model.trim="nickname" class="field" maxlength="100" placeholder="请输入昵称" />
        </label>
        <p v-if="notice" class="feedback success" role="status">{{ notice }}</p>
        <p v-if="error" class="feedback error" role="alert">{{ error }}</p>
        <button class="button" :disabled="submitting">
          {{ submitting ? '注册中...' : '完成注册' }}
        </button>
        <RouterLink to="/login">已有账号？返回登录</RouterLink>
      </form>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'

import { sendCode } from '@/api/auth'
import { apiErrorMessage } from '@/api/client'
import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'

const router = useRouter()
const session = useSessionStore()
const phone = ref('')
const code = ref('')
const password = ref('')
const nickname = ref('')
const sending = ref(false)
const submitting = ref(false)
const cooldown = ref(0)
const notice = ref('')
const error = ref('')
let cooldownTimer: number | undefined

function validPhone() {
  return /^1[3-9]\d{9}$/.test(phone.value)
}

async function requestCode() {
  error.value = ''
  notice.value = ''
  if (!validPhone()) {
    error.value = '请输入正确的手机号。'
    return
  }

  sending.value = true
  try {
    const result = await sendCode({ phone: phone.value, scene: 'REGISTER', clientType: 'H5_WEB' })
    notice.value = '验证码已发送，请从本地测试渠道获取。'
    cooldown.value = result.cooldownSeconds
    cooldownTimer = window.setInterval(() => {
      cooldown.value -= 1
      if (cooldown.value <= 0 && cooldownTimer) {
        window.clearInterval(cooldownTimer)
        cooldownTimer = undefined
      }
    }, 1000)
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '验证码发送失败。')
  } finally {
    sending.value = false
  }
}

async function submit() {
  error.value = ''
  notice.value = ''
  if (!validPhone()) {
    error.value = '请输入正确的手机号。'
    return
  }
  if (!code.value) {
    error.value = '请输入验证码。'
    return
  }
  if (password.value.length < 6) {
    error.value = '密码至少需要 6 位。'
    return
  }

  submitting.value = true
  try {
    await session.register({
      phone: phone.value,
      code: code.value,
      password: password.value,
      nickname: nickname.value || undefined,
      clientType: 'H5_WEB'
    })
    await router.push('/me')
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '注册失败。')
  } finally {
    submitting.value = false
  }
}

onBeforeUnmount(() => {
  if (cooldownTimer) window.clearInterval(cooldownTimer)
})
</script>

<style scoped>
.auth-page {
  display: grid;
  min-height: 620px;
  place-items: center;
  padding: 24px 0;
}

.auth-card {
  display: grid;
  width: min(460px, 100%);
  gap: 16px;
  padding: 24px;
}

.auth-card h1 {
  margin: 8px 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

label {
  display: grid;
  gap: 7px;
  font-size: 14px;
  font-weight: 600;
}

.code-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 132px;
  gap: 10px;
  align-items: end;
}

.code-button {
  min-height: 42px;
}

.feedback {
  margin: 0;
  font-weight: 600;
}

.error {
  color: var(--color-danger);
}

.success {
  color: var(--color-success);
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 520px) {
  .code-row {
    grid-template-columns: 1fr;
  }
}
</style>
