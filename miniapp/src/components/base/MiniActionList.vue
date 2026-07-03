<template>
  <view class="mini-action-list">
    <view
      v-for="item in items"
      :key="item.key"
      class="action-item"
      :class="{ danger: item.danger }"
      @click="handleSelect(item.key)"
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
import { ref } from 'vue'

export interface ActionListItem {
  key: string
  title: string
  desc?: string
  danger?: boolean
}

defineProps<{ items: ActionListItem[] }>()
const emit = defineEmits<{ select: [key: string] }>()

const selecting = ref(false)

function handleSelect(key: string) {
  if (selecting.value) return
  selecting.value = true
  emit('select', key)
  setTimeout(() => {
    selecting.value = false
  }, 500)
}
</script>

<style scoped>
.mini-action-list {
  display: flex;
  flex-direction: column;
  margin-top: 2rpx;
}

.action-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  min-height: 88rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.16);
  background: transparent;
  transition: transform 180ms ease-out, background-color 180ms ease-out;
}

.action-item:last-child {
  border-bottom: 0;
}

.action-item:active {
  background-color: rgba(24, 54, 83, 0.05);
  transform: translateY(1rpx) scale(0.99);
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
  flex-shrink: 0;
  color: var(--tree-text-secondary, #64748b);
  font-size: 32rpx;
  font-weight: 400;
  line-height: 1;
  transition: transform 180ms ease-out;
}

.action-item:active .action-arrow {
  transform: translateX(4rpx);
}

.action-item.danger:active {
  background-color: rgba(180, 83, 58, 0.06);
}
</style>
