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
  position: relative;
  margin: 0 0 22rpx;
  padding-bottom: 18rpx;
}
.mini-section-header::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: 0;
  width: 96rpx;
  height: 5rpx;
  border-radius: 999rpx;
  background: linear-gradient(90deg, rgba(31, 58, 95, 0.22) 0%, rgba(47, 107, 87, 0.08) 100%);
}
.mini-section-header.accent .title {
  position: relative;
  display: inline-flex;
  align-items: center;
  border-radius: 999rpx;
  background: rgba(24, 54, 83, 0.055);
  padding: 8rpx 18rpx 8rpx 28rpx;
}
.mini-section-header.accent .title::before {
  content: '';
  position: absolute;
  left: 12rpx;
  top: 50%;
  width: 8rpx;
  height: 28rpx;
  margin-top: -14rpx;
  border-radius: 999rpx;
  background: linear-gradient(180deg, var(--tree-primary, #1f3a5f) 0%, var(--tree-green, #2f6b57) 100%);
  box-shadow: 0 4rpx 10rpx rgba(31, 58, 95, 0.12);
}
.title {
  display: block;
  color: var(--tree-text-primary, #1f3a5f);
  font-size: 35rpx;
  font-weight: 800;
  line-height: 1.35;
}
.subtitle {
  display: block;
  width: 100%;
  margin-top: 8rpx;
  color: var(--tree-text-secondary, #64748b);
  font-size: 22rpx;
  line-height: 1.55;
}
</style>
