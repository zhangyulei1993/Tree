<template>
  <view class="mini-card" :class="[variant, { 'no-margin': noMargin, flat }]">
    <slot />
  </view>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'default' | 'soft' | 'hero' | 'elevated' | 'archive' | 'info'
    noMargin?: boolean
    flat?: boolean
  }>(),
  { variant: 'default', noMargin: false, flat: false }
)
</script>

<style scoped>
.mini-card {
  margin-bottom: 20rpx;
  border: 1rpx solid var(--tree-border-subtle, #eef1f4);
  border-radius: var(--tree-radius-lg, 24rpx);
  background: var(--tree-surface, #fff);
  padding: var(--tree-space-card, 28rpx);
  box-shadow: var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.04));
}
.mini-card.soft {
  background: var(--tree-surface-muted, #f8fafc);
  border-color: var(--tree-border-subtle, #eef1f4);
  box-shadow: none;
}
.mini-card.hero {
  border-color: rgba(47, 107, 87, 0.12);
  background: linear-gradient(145deg, #fff 0%, var(--tree-green-soft, #f3faf7) 100%);
  box-shadow: var(--tree-shadow-md, 0 6rpx 24rpx rgba(15, 23, 42, 0.06));
}
.mini-card.elevated {
  border-color: var(--tree-border-subtle, #eef1f4);
  background: var(--tree-surface, #fff);
  box-shadow: var(--tree-shadow-md, 0 6rpx 24rpx rgba(15, 23, 42, 0.06));
}
.mini-card.archive {
  position: relative;
  padding-left: 28rpx;
  border-color: var(--tree-border-subtle, #eef1f4);
  background: var(--tree-surface, #fff);
  overflow: hidden;
}
.mini-card.archive::before {
  content: '';
  position: absolute;
  left: 0;
  top: 18rpx;
  bottom: 18rpx;
  width: 4rpx;
  border-radius: 0 4rpx 4rpx 0;
  background: linear-gradient(180deg, var(--tree-primary, #1f3a5f) 0%, var(--tree-green, #2f6b57) 100%);
  opacity: 0.55;
}
.mini-card.info {
  border-color: var(--tree-border-green, #d8e5dc);
  background: var(--tree-green-soft, #f3faf7);
  box-shadow: none;
}
.mini-card.flat {
  border-color: transparent;
  background: #fff;
}
.mini-card.no-margin {
  margin-bottom: 0;
}
</style>
