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
  position: relative;
  margin-bottom: 24rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: var(--tree-radius-lg, 24rpx);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96) 0%, rgba(255, 252, 247, 0.92) 100%);
  padding: 30rpx;
  overflow: hidden;
  box-shadow: var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.04));
}
.mini-card.soft {
  border-color: rgba(216, 229, 220, 0.82);
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.08), transparent 180rpx),
    rgba(247, 250, 248, 0.94);
  box-shadow: 0 8rpx 26rpx rgba(24, 54, 83, 0.04);
}
.mini-card.hero {
  border-color: rgba(255, 255, 255, 0.18);
  background:
    radial-gradient(circle at 88% 0%, rgba(216, 175, 104, 0.22), transparent 220rpx),
    linear-gradient(135deg, #17304c 0%, #235a54 100%);
  color: #fff;
  box-shadow: var(--tree-shadow-lg, 0 18rpx 44rpx rgba(24, 54, 83, 0.14));
}
.mini-card.elevated {
  border-color: rgba(255, 255, 255, 0.72);
  background: rgba(255, 255, 255, 0.96);
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
  box-shadow: none;
}
.mini-card.no-margin {
  margin-bottom: 0;
}
</style>
