<template>
  <view class="family-context">
    <view class="context-copy">
      <text class="context-kicker">当前家庭</text>
      <text class="context-family">{{ familyName || '家庭空间' }}</text>
      <text class="context-section">{{ section }}</text>
      <text v-if="subtitle" class="context-subtitle">{{ subtitle }}</text>
    </view>
    <view class="context-side">
      <text v-if="roleLabel" class="context-role">{{ roleLabel }}</text>
      <button class="context-back" @click="$emit('back')">{{ backLabel }}</button>
    </view>
  </view>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    familyName?: string
    section: string
    subtitle?: string
    roleLabel?: string
    backLabel?: string
  }>(),
  {
    backLabel: '家庭概览'
  }
)

defineEmits<{ back: [] }>()
</script>

<style scoped>
.family-context {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
  margin-bottom: 20rpx;
  border: 1rpx solid rgba(47, 107, 87, 0.16);
  border-radius: 24rpx;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.98), rgba(240, 248, 244, 0.94));
  padding: 24rpx;
}

.context-copy {
  flex: 1;
  min-width: 0;
}

.context-kicker,
.context-family,
.context-section,
.context-subtitle {
  display: block;
}

.context-kicker {
  color: var(--tree-green, #2f6b57);
  font-size: 20rpx;
  font-weight: 600;
}

.context-family {
  margin-top: 4rpx;
  overflow: hidden;
  color: var(--tree-text-primary, #1f3a5f);
  font-size: 26rpx;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.context-section {
  margin-top: 16rpx;
  color: var(--tree-text-primary, #1f2937);
  font-size: 34rpx;
  font-weight: 700;
  line-height: 1.3;
}

.context-subtitle {
  margin-top: 6rpx;
  color: var(--tree-text-secondary, #6b7280);
  font-size: 22rpx;
  line-height: 1.5;
}

.context-side {
  display: flex;
  flex: 0 0 auto;
  flex-direction: column;
  align-items: flex-end;
  gap: 12rpx;
}

.context-role {
  border-radius: 999rpx;
  background: rgba(47, 107, 87, 0.1);
  color: var(--tree-green, #2f6b57);
  padding: 6rpx 12rpx;
  font-size: 20rpx;
}

.context-back {
  width: auto;
  min-height: 52rpx;
  margin: 0;
  border: 1rpx solid var(--tree-border-warm, #ebe4d6);
  border-radius: 999rpx;
  background: #fff;
  color: var(--tree-text-secondary, #6b7280);
  padding: 6rpx 16rpx;
  font-size: 22rpx;
  line-height: 1.4;
  transition: transform 180ms ease-out, background-color 180ms ease-out;
}

.context-back:active {
  background-color: rgba(24, 54, 83, 0.05);
  transform: translateY(1rpx) scale(0.99);
}

.context-back::after {
  border: 0;
}
</style>
