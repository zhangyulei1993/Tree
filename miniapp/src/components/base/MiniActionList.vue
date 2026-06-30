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
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 2rpx;
}
.action-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  border: 1rpx solid rgba(226, 232, 240, 0.86);
  border-radius: 22rpx;
  background:
    linear-gradient(135deg, rgba(248, 250, 252, 0.98) 0%, rgba(255, 255, 255, 0.92) 100%);
  padding: 22rpx 18rpx 22rpx 22rpx;
  box-shadow: 0 6rpx 18rpx rgba(24, 54, 83, 0.035);
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.action-item:first-child {
  border-top: 1rpx solid rgba(226, 232, 240, 0.86);
  padding-top: 22rpx;
}
.action-item:active {
  opacity: 0.9;
  transform: translateY(2rpx);
}
.action-main {
  flex: 1;
  min-width: 0;
}
.action-title {
  display: block;
  color: var(--tree-text-primary, #1e293b);
  font-size: 28rpx;
  font-weight: 700;
  line-height: 1.42;
}
.action-title.danger {
  color: #b4533a;
  font-weight: 500;
}
.action-desc {
  display: block;
  margin-top: 6rpx;
  color: var(--tree-text-secondary, #64748b);
  font-size: 21rpx;
  line-height: 1.55;
}
.action-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44rpx;
  height: 44rpx;
  border-radius: 999rpx;
  background: rgba(24, 54, 83, 0.06);
  color: var(--tree-primary);
  font-size: 30rpx;
  font-weight: 700;
}
.action-item.danger {
  border-color: rgba(180, 83, 58, 0.18);
  background: rgba(255, 247, 243, 0.92);
}
.action-item.danger .action-arrow {
  background: rgba(180, 83, 58, 0.08);
  color: #b4533a;
}
</style>
