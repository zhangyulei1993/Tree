<template>
  <view class="mini-notice" :class="tone">
    <text v-if="title" class="notice-title">{{ title }}</text>
    <text class="notice-text"><slot /></text>
  </view>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{ title?: string; tone?: 'info' | 'warm' | 'security' }>(),
  { title: '', tone: 'info' }
)
</script>

<style scoped>
.mini-notice {
  border-radius: var(--tree-radius-md, 16rpx);
  padding: 18rpx 22rpx;
}
.mini-notice.info {
  background: #f8f6f1;
  color: var(--tree-text-secondary, #6b7280);
  border: 1rpx solid var(--tree-border-warm, #ebe4d6);
}
.mini-notice.warm {
  background: #fdf8ee;
  color: var(--tree-gold-text, #8a6d2f);
  border: 1rpx solid #f0e2c4;
}
.mini-notice.security {
  background: #f4f8f5;
  color: var(--tree-text-secondary, #6b7280);
  border: 1rpx solid var(--tree-border-green, #d8e5dc);
}
.notice-title {
  display: block;
  margin-bottom: 6rpx;
  color: var(--tree-text-primary, #1f3a5f);
  font-size: 24rpx;
  font-weight: 600;
}
.notice-text {
  display: block;
  font-size: 24rpx;
  line-height: 1.7;
}
</style>
