<template>
  <view class="mini-action-list">
    <view
      v-for="item in items"
      :key="item.key"
      class="action-item"
      :class="{ danger: item.danger }"
      @click="$emit('select', item.key)"
    >
      <view class="action-main">
        <text class="action-title" :class="{ danger: item.danger }">{{ item.title }}</text>
        <text v-if="item.desc" class="action-desc">{{ item.desc }}</text>
      </view>
      <text class="action-arrow">›</text>
    </view>
  </view>
</template>

<script setup lang="ts">
export interface ActionListItem {
  key: string
  title: string
  desc?: string
  danger?: boolean
}

defineProps<{ items: ActionListItem[] }>()
defineEmits<{ select: [key: string] }>()
</script>

<style scoped>
.mini-action-list {
  margin-top: -4rpx;
}
.action-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 22rpx 0;
  border-top: 1rpx solid var(--tree-border-subtle, #f0ebe3);
}
.action-item:first-child {
  border-top: none;
  padding-top: 4rpx;
}
.action-item:active {
  opacity: 0.72;
}
.action-title {
  display: block;
  color: var(--tree-text);
  font-size: 28rpx;
  font-weight: 600;
}
.action-title.danger {
  color: #b4533a;
  font-weight: 500;
}
.action-desc {
  display: block;
  margin-top: 4rpx;
  color: var(--tree-text-weak);
  font-size: 22rpx;
  line-height: 1.5;
}
.action-arrow {
  color: var(--tree-text-weak);
  font-size: 28rpx;
}
</style>
