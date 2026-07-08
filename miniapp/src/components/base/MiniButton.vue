<template>
  <view
    class="mini-button"
    :class="[variant, size, { block, disabled }]"
    @tap.stop="handleTap"
  >
    <text v-if="loading">处理中...</text>
    <slot v-else />
  </view>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
    size?: 'md' | 'sm'
    block?: boolean
    disabled?: boolean
    loading?: boolean
  }>(),
  { variant: 'primary', size: 'md', block: true, disabled: false, loading: false }
)
const emit = defineEmits<{ click: [] }>()

function handleTap() {
  if (props.disabled || props.loading) return
  emit('click')
}
</script>

<style scoped>
.mini-button {
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  margin: 0;
  border: 1rpx solid transparent;
  border-radius: var(--tree-radius-md, 16rpx);
  font-weight: 600;
  line-height: 1.4;
  text-align: center;
  transition: transform 0.18s ease, opacity 0.18s ease, box-shadow 0.18s ease;
}
.mini-button.block {
  width: 100%;
}
.mini-button.md {
  min-height: 88rpx;
  padding: 22rpx 28rpx;
  font-size: 28rpx;
}
.mini-button.sm {
  min-height: 68rpx;
  padding: 14rpx 22rpx;
  font-size: 24rpx;
}
.mini-button.primary {
  background: linear-gradient(135deg, var(--tree-primary, #1f3a5f) 0%, var(--tree-green, #2f6b57) 100%);
  color: #fff;
  box-shadow: 0 12rpx 28rpx rgba(31, 58, 95, 0.16);
}
.mini-button.secondary {
  background: #fff;
  color: var(--tree-text-primary, #1f3a5f);
  border-color: var(--tree-border-warm, #ebe4d6);
  box-shadow: 0 6rpx 16rpx rgba(15, 23, 42, 0.05);
}
.mini-button.danger {
  background: #b4533a;
  color: #fff;
  box-shadow: 0 10rpx 22rpx rgba(180, 83, 58, 0.16);
}
.mini-button.ghost {
  background: rgba(255, 255, 255, 0.56);
  color: var(--tree-text-primary, #1f3a5f);
  border-color: var(--tree-border-subtle, #eef1f4);
}
.mini-button:active {
  transform: translateY(1rpx) scale(0.992);
  opacity: 0.88;
}
.mini-button.disabled {
  opacity: 0.55;
}
</style>
