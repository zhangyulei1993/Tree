<template>
  <view class="page">
    <view class="card">
      <text class="title">注销账号</text>
      <text class="warning">注销后当前账号将不可继续登录，家庭树成员节点不会因此被删除。</text>
      <view class="check-row" @click="confirmed = !confirmed">
        <checkbox :checked="confirmed" color="#1f3a5f" />
        <text>我已了解注销影响，并确认继续</text>
      </view>
      <textarea v-model.trim="reason" class="textarea" maxlength="200" placeholder="注销原因，可选" />
      <button class="button secondary" @click="back">返回</button>
      <button class="button danger" :disabled="!confirmed || cancelled" @click="cancelAccount">
        {{ cancelled ? '已完成 mock 注销' : '确认 mock 注销' }}
      </button>
      <text v-if="cancelled" class="result">账号状态已模拟更新为 CANCELLED。</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const reason = ref('')
const confirmed = ref(false)
const cancelled = ref(false)

function back() {
  uni.navigateBack()
}

function cancelAccount() {
  uni.showModal({
    title: '确认注销账号',
    content: '这是高风险操作。当前仅执行 mock 状态变更，不请求真实后端。',
    confirmText: '确认注销',
    confirmColor: '#c0392b',
    success: ({ confirm }) => {
      if (!confirm) return
      cancelled.value = true
      session.logout()
      uni.showToast({ title: 'mock 注销完成', icon: 'none' })
    }
  })
}
</script>

<style scoped>
.warning {
  display: block;
  border-left: 6rpx solid #c0392b;
  background: #fff5f4;
  color: #7f1d1d;
  padding: 18rpx;
  font-size: 26rpx;
  line-height: 1.7;
}

.check-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 22rpx;
  font-size: 26rpx;
}

.button.danger {
  background: #c0392b;
}

.button:disabled {
  opacity: 0.45;
}

.result {
  display: block;
  margin-top: 18rpx;
  color: #2f6b57;
  text-align: center;
}
</style>
