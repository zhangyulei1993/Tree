<template>
  <view class="tree-page">
    <MiniBackHome />
    <MiniCard variant="soft">
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
      <view class="btn-stack">
        <MiniButton variant="secondary" @click="back">返回</MiniButton>
        <MiniButton variant="danger" :disabled="!confirmed || cancelled" @click="cancelAccount">
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
import { ref } from 'vue'

import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
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
    content: '注销后账号将不可继续登录。确定要继续吗？',
    confirmText: '确认注销',
    confirmColor: '#b4533a',
    success: ({ confirm }) => {
      if (!confirm) return
      cancelled.value = true
      session.logout()
      uni.showToast({ title: '注销完成', icon: 'none' })
    }
  })
}
</script>

<style scoped>
.check-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 22rpx;
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
</style>
