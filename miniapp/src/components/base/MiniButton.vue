<template>
  <button
    class="mini-button"
    :class="[variant, size, { block }]"
    :disabled="disabled"
    :loading="loading"
    @click="$emit('click')"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
    size?: 'md' | 'sm'
    block?: boolean
    disabled?: boolean
    loading?: boolean
  }>(),
  { variant: 'primary', size: 'md', block: true, disabled: false, loading: false }
)
defineEmits<{ click: [] }>()
</script>

<style scoped>
.mini-button {
  margin: 0;
  border: none;
  border-radius: var(--tree-radius-md, 16rpx);
  font-weight: 600;
  line-height: 1.4;
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
}
.mini-button.secondary {
  background: #fff;
  color: var(--tree-text-primary, #1f3a5f);
  border: 1rpx solid var(--tree-border-warm, #ebe4d6);
}
.mini-button.danger {
  background: #b4533a;
  color: #fff;
}
.mini-button.ghost {
  background: transparent;
  color: var(--tree-text-secondary, #6b7280);
  border: 1rpx solid var(--tree-border-warm, #ebe4d6);
}
.mini-button[disabled] {
  opacity: 0.55;
}
</style>
