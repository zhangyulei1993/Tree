<template>
  <view class="mini-empty-state">
    <view class="symbol-ring" />
    <text class="title">{{ title }}</text>
    <text v-if="description" class="description">{{ description }}</text>
    <MiniButton
      v-if="actionText"
      variant="primary"
      size="sm"
      :block="false"
      class="action-btn"
      @click="$emit('action')"
    >
      {{ actionText }}
    </MiniButton>
  </view>
</template>

<script setup lang="ts">
import MiniButton from './MiniButton.vue'

withDefaults(
  defineProps<{
    symbol?: string
    title: string
    description?: string
    actionText?: string
  }>(),
  { symbol: '', description: '', actionText: '' }
)
defineEmits<{ action: [] }>()
</script>

<style scoped>
.mini-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40rpx 20rpx;
  text-align: center;
}
.symbol-ring {
  width: 72rpx;
  height: 72rpx;
  margin-bottom: 20rpx;
  border: 3rpx solid var(--tree-border-warm, #ebe4d6);
  border-radius: 50%;
  background: radial-gradient(circle at 50% 40%, #fff 0%, var(--tree-green-light, #e8f3ee) 100%);
}
.title {
  color: var(--tree-text);
  font-size: 30rpx;
  font-weight: 600;
}
.description {
  margin-top: 10rpx;
  color: var(--tree-text-secondary);
  font-size: 26rpx;
  line-height: 1.7;
}
.action-btn {
  margin-top: 22rpx;
}
</style>
