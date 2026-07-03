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
  margin-top: 20rpx;
}

.consent-row {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
}

.consent-text {
  flex: 1;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.7;
}

.consent-link {
  color: var(--tree-green);
}
</style>
