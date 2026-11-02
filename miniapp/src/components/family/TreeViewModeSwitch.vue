<template>
  <view class="tree-mode-switch">
    <view
      class="switch-item"
      :class="{ active: modelValue === 'structure' }"
      @click="emit('update:modelValue', 'structure')"
    >
      <text>树状结构</text>
    </view>
    <view
      class="switch-item"
      :class="{ active: modelValue === 'detail' }"
      @click="emit('update:modelValue', 'detail')"
    >
      <text>关系明细</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import type { FamilyTreeViewMode } from '@/features/family-tree/types'

defineProps<{
  modelValue: FamilyTreeViewMode
}>()

const emit = defineEmits<{
  'update:modelValue': [value: FamilyTreeViewMode]
}>()
</script>

<style scoped>
.tree-mode-switch {
  display: flex;
  gap: 8rpx;
  margin-bottom: 16rpx;
  border-radius: 999rpx;
  background: #eef1f4;
  padding: 6rpx;
}
.switch-item {
  flex: 1;
  border-radius: 999rpx;
  padding: 14rpx 0;
  text-align: center;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  font-weight: 500;
}
.switch-item.active {
  background: #fff;
  box-shadow: 0 2rpx 8rpx rgba(15, 23, 42, 0.06);
  color: var(--tree-green);
  font-weight: 600;
}
</style>
