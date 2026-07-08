<template>
  <view class="archive-page cancel-page">
    <MiniBackHome />

    <view class="cancel-head archive-page-head">
      <view>
        <text class="archive-kicker">Account Closure</text>
        <text class="archive-title">注销账号</text>
        <text class="archive-subtitle">此操作不可撤销，注销前请仔细确认影响范围</text>
      </view>
      <view class="archive-seal">销</view>
    </view>

    <view class="archive-danger-panel">
      <text class="archive-danger-title">不可逆风险</text>
      <text class="archive-danger-desc">
        注销后当前账号将不可继续登录，家庭树成员节点不会因此被删除。注销前需使用当前微信重新认证。
      </text>
    </view>

    <view class="cancel-impact archive-list">
      <view class="archive-row static">
        <view class="archive-row-main">
          <text class="archive-row-title">账号登录</text>
          <text class="archive-row-desc">注销后无法继续使用此账号登录 Tree</text>
        </view>
      </view>
      <view class="archive-row static">
        <view class="archive-row-main">
          <text class="archive-row-title">家庭树节点</text>
          <text class="archive-row-desc">已录入的家庭成员节点不会被删除，仍保留在对应家庭中</text>
        </view>
      </view>
      <view class="archive-row static">
        <view class="archive-row-main">
          <text class="archive-row-title">微信重新认证</text>
          <text class="archive-row-desc">提交注销前需使用当前微信完成一次身份确认</text>
        </view>
      </view>
      <view class="archive-row static">
        <view class="archive-row-main">
          <text class="archive-row-title">隐私与安全</text>
          <text class="archive-row-desc">注销过程不会记录或展示完整微信凭证，仅用于身份核验</text>
        </view>
      </view>
    </view>

    <view class="archive-form-panel cancel-form">
      <view class="archive-section-head">
        <text class="archive-section-title">确认注销</text>
        <text class="archive-section-subtitle">勾选确认并可选填写注销原因</text>
      </view>

      <view class="check-row" @click="confirmed = !confirmed">
        <checkbox :checked="confirmed" color="#a83b2d" />
        <text>我已了解注销影响，并确认继续</text>
      </view>

      <text class="tree-field-label">注销原因（可选）</text>
      <textarea
        v-model.trim="reason"
        class="tree-textarea"
        maxlength="200"
        placeholder="注销原因，可选"
      />

      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>

      <view class="btn-stack">
        <MiniButton variant="secondary" @click="back">返回</MiniButton>
        <MiniButton
          class="cancel-submit"
          variant="danger"
          size="sm"
          :disabled="!confirmed || cancelled || submitting"
          :loading="submitting"
          @click="confirmCancelAccount"
        >
          {{ cancelled ? '已完成注销' : '微信认证并注销' }}
        </MiniButton>
      </view>

      <MiniNotice v-if="cancelled" tone="info" class="result-notice">
        账号已注销，当前登录状态已清除。
      </MiniNotice>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { cancelAccountByWechat } from '@/api/auth'
import { apiErrorMessage } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import { optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const reason = ref('')
const confirmed = ref(false)
const cancelled = ref(false)
const submitting = ref(false)
const errorMessage = ref('')

function back() {
  uni.navigateBack()
}

function confirmCancelAccount() {
  const validationMessage = validateTextFields([
    { value: reason.value, label: '注销原因', kind: 'multiLine', maxLength: 200 }
  ])
  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }
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
        const loginResult = await new Promise<UniApp.LoginRes>((resolve, reject) => {
          uni.login({ provider: 'weixin', success: resolve, fail: reject })
        })
        if (!loginResult.code) {
          throw new Error('微信认证失败，请稍后重试。')
        }
        await cancelAccountByWechat({ code: loginResult.code, cancelReason: optionalText(reason.value) })
        cancelled.value = true
        session.clearSession()
        uni.showToast({ title: '注销完成', icon: 'none' })
        setTimeout(() => uni.reLaunch({ url: '/pages/home/index' }), 500)
      } catch (error) {
        errorMessage.value = apiErrorMessage(error, '账号注销失败。请确认已退出所有家庭。')
      } finally {
        submitting.value = false
      }
    }
  })
}

onShow(() => {
  session.restoreSession()
  if (!session.isLoggedIn) {
    uni.reLaunch({ url: '/pages/auth/wechat-login' })
  }
})
</script>

<style scoped>
.cancel-page {
  padding-top: 28rpx;
}

.cancel-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.cancel-impact {
  margin-bottom: 24rpx;
}

.cancel-form {
  margin-bottom: 24rpx;
}

.check-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 8rpx;
  border-top: 1rpx solid rgba(168, 59, 45, 0.22);
  border-bottom: 1rpx solid rgba(168, 59, 45, 0.14);
  background: rgba(168, 59, 45, 0.03);
  padding: 18rpx 0;
  color: var(--archive-ink);
  font-size: 26rpx;
  line-height: 1.5;
}

.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 22rpx;
}

.cancel-submit {
  align-self: flex-start;
}

.result-notice {
  margin-top: 18rpx;
}
</style>
