<template>
  <view class="tree-page cancel-page">
    <MiniBackHome />
    <MiniCard variant="soft" class="cancel-card">
      <text class="tree-page-title">注销账号</text>
      <MiniNotice tone="warm" title="重要提示">
        注销后当前账号将不可继续登录，家庭树成员节点不会因此被删除。
      </MiniNotice>
      <view class="check-row" @click="confirmed = !confirmed">
        <checkbox :checked="confirmed" color="#1f3a5f" />
        <text>我已了解注销影响，并确认继续</text>
      </view>
      <textarea
        v-model.trim="reason"
        class="textarea"
        maxlength="200"
        placeholder="注销原因，可选"
      />
      <input v-model.trim="phoneCode" class="input" maxlength="6" type="number" placeholder="请输入当前手机号验证码" />
      <MiniButton variant="secondary" :disabled="sendingCode" @click="sendCancelCode">
        {{ sendingCode ? '发送中...' : '获取注销验证码' }}
      </MiniButton>
      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <view class="btn-stack">
        <MiniButton variant="secondary" @click="back">返回</MiniButton>
        <MiniButton variant="danger" :disabled="!confirmed || !phoneCode || cancelled || submitting" @click="confirmCancelAccount">
          {{ cancelled ? '已完成注销' : '确认注销' }}
        </MiniButton>
      </view>
      <MiniNotice v-if="cancelled" tone="info" class="result-notice">
        账号已注销，当前登录状态已清除。
      </MiniNotice>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { cancelAccount, sendCode } from '@/api/auth'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const reason = ref('')
const confirmed = ref(false)
const cancelled = ref(false)
const phoneCode = ref('')
const sendingCode = ref(false)
const submitting = ref(false)
const errorMessage = ref('')

function back() {
  uni.navigateBack()
}

async function sendCancelCode() {
  const phone = session.user?.phone
  if (!phone) { errorMessage.value = '当前账号未绑定手机号。'; return }
  sendingCode.value = true
  errorMessage.value = ''
  try {
    await sendCode(phone, 'CANCEL_ACCOUNT')
    uni.showToast({ title: '验证码已发送', icon: 'none' })
  } catch (error) { errorMessage.value = apiErrorMessage(error, '验证码发送失败。') }
  finally { sendingCode.value = false }
}

function confirmCancelAccount() {
  uni.showModal({
    title: '确认注销账号',
    content: '注销后账号将不可继续登录。确定要继续吗？',
    confirmText: '确认注销',
    confirmColor: '#b4533a',
    success: async ({ confirm }) => {
      if (!confirm) return
      submitting.value = true
      errorMessage.value = ''
      try {
        await cancelAccount({ phoneCode: phoneCode.value, cancelReason: reason.value || undefined })
        cancelled.value = true
        session.clearSession()
        uni.showToast({ title: '注销完成', icon: 'none' })
        setTimeout(() => uni.reLaunch({ url: '/pages/home/index' }), 500)
      } catch (error) { errorMessage.value = apiErrorMessage(error, '账号注销失败。请确认已退出所有家庭。') }
      finally { submitting.value = false }
    }
  })
}

onShow(() => {
  session.restoreSession()
  if (!session.isLoggedIn) {
    uni.reLaunch({ url: '/pages/auth/wechat-login' })
    return
  }
  if (!session.isPhoneBound) {
    uni.redirectTo({ url: '/pages/auth/bind-phone' })
  }
})
</script>

<style scoped>
.check-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 22rpx;
  border: 1rpx solid rgba(181, 71, 60, 0.18);
  border-radius: 22rpx;
  background: rgba(255, 248, 246, 0.92);
  padding: 18rpx;
  color: var(--tree-text);
  font-size: 26rpx;
  line-height: 1.5;
}

.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 22rpx;
}

.result-notice {
  margin-top: 22rpx;
}

.cancel-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(181, 71, 60, 0.10), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.cancel-card {
  border-color: rgba(181, 71, 60, 0.18);
  background:
    radial-gradient(circle at 100% 0%, rgba(181, 71, 60, 0.08), transparent 180rpx),
    rgba(255, 255, 255, 0.94);
  box-shadow: 0 18rpx 44rpx rgba(24, 54, 83, 0.08);
}
</style>
