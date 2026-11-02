<template>
  <PageShell>
    <section class="container auth-page">
      <form class="card auth-card" @submit.prevent="submit">
        <div class="auth-intro">
          <span class="eyebrow">创建账号</span>
          <h1>先建立可信身份，再进入家庭协作。</h1>
          <p>手机号用于确认身份，后续可接收家庭邀请和加入申请处理结果。</p>
          <div class="intro-points">
            <span>验证码验证手机号</span>
            <span>不公开展示联系方式</span>
            <span>登录后管理家庭资料</span>
          </div>
        </div>
        <div class="auth-form">
          <h2>手机号注册</h2>
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
        </div>
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
  min-height: 660px;
  place-items: center;
  padding: 48px 0;
}

.auth-card {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(380px, 0.78fr);
  width: min(960px, 100%);
  gap: 0;
  padding: 0;
  overflow: hidden;
}

.auth-intro {
  display: grid;
  align-content: center;
  padding: 42px;
  background:
    radial-gradient(circle at 88% 18%, rgba(47, 107, 87, 0.14), transparent 160px),
    var(--gradient-hero);
}

.auth-intro h1 {
  margin: 16px 0 14px;
  color: var(--color-primary);
  font-size: clamp(30px, 4vw, 46px);
  letter-spacing: -0.04em;
  line-height: 1.1;
}

.auth-intro p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.8;
}

.intro-points {
  display: grid;
  gap: 10px;
  margin-top: 28px;
}

.intro-points span {
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  padding: 12px 14px;
  color: var(--color-primary);
  font-weight: 800;
}

.auth-form {
  display: grid;
  align-content: center;
  gap: 16px;
  padding: 42px;
}

.auth-form h2 {
  margin: 0;
  color: var(--color-primary);
  font-size: 28px;
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
  .auth-card {
    grid-template-columns: 1fr;
  }

  .auth-intro,
  .auth-form {
    padding: 28px;
  }

  .code-row {
    grid-template-columns: 1fr;
  }
}
</style>
