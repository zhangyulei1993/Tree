<template>
  <view class="mini-section-header" :class="{ accent }">
    <text class="title">{{ title }}</text>
    <text v-if="subtitle" class="subtitle">{{ subtitle }}</text>
  </view>
</template>

<script setup lang="ts">
defineProps<{ title: string; subtitle?: string; accent?: boolean }>()
</script>

<style scoped>
.mini-section-header {
  margin: 4rpx 0 18rpx;
}
.mini-section-header.accent .title {
  position: relative;
  padding-left: 18rpx;
}
.mini-section-header.accent .title::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  width: 6rpx;
  height: 24rpx;
  margin-top: -12rpx;
  border-radius: 3rpx;
  background: linear-gradient(180deg, var(--tree-primary, #1f3a5f) 0%, var(--tree-green, #2f6b57) 100%);
  opacity: 0.75;
}
.title {
  display: block;
  color: var(--tree-text-primary, #1f3a5f);
  font-size: 30rpx;
  font-weight: 600;
  line-height: 1.4;
}
.subtitle {
  display: block;
  width: 100%;
  margin-top: 6rpx;
  color: var(--tree-text-secondary, #6b7280);
  font-size: 24rpx;
  line-height: 1.6;
}
</style>
