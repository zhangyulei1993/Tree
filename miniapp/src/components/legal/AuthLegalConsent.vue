<template>
  <view class="auth-legal-consent">
    <view class="consent-row" @click="toggle">
      <checkbox :checked="modelValue" color="#2F6B57" @click.stop="toggle" />
      <text class="consent-text">
        我已阅读并同意
        <text class="consent-link" @click.stop="openUserAgreement">《用户协议》</text>
        和
        <text class="consent-link" @click.stop="openPrivacyPolicy">《隐私政策》</text>
      </text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { openPrivacyPolicy, openUserAgreement } from '@/features/legal/privacyConsent'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<style scoped>
.auth-legal-consent {
  border-radius: 18rpx;
  background: rgba(47, 107, 87, 0.045);
  padding: 18rpx 16rpx;
}

.consent-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
  min-height: 52rpx;
}

.consent-row:active {
  opacity: 0.78;
}

checkbox {
  flex-shrink: 0;
  transform: scale(0.86);
}

.consent-text {
  flex: 1;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.65;
}

.consent-link {
  color: var(--tree-green);
  font-weight: 600;
}
</style>
