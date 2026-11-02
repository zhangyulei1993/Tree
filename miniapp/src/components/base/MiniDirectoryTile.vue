<template>
  <view
    class="directory-tile"
    :class="{ grouped }"
    @click="handleClick"
  >
    <view class="directory-tile-icon">
      <view class="tree-symbol" :class="iconClass" />
    </view>
    <view class="directory-tile-copy">
      <text class="directory-tile-title">{{ title }}</text>
      <text class="directory-tile-desc">{{ desc }}</text>
    </view>
    <text class="directory-arrow">›</text>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

withDefaults(
  defineProps<{
    title: string
    desc: string
    iconClass: string
    grouped?: boolean
  }>(),
  { grouped: true }
)

const emit = defineEmits<{ click: [] }>()

const navigating = ref(false)

function handleClick() {
  if (navigating.value) return
  navigating.value = true
  emit('click')
  setTimeout(() => {
    navigating.value = false
  }, 500)
}
</script>

<style scoped>
.directory-tile {
  display: flex;
  align-items: center;
  gap: 18rpx;
  min-height: 88rpx;
  padding: 20rpx 0;
  transition: transform 180ms ease-out, background-color 180ms ease-out;
}

.directory-tile.grouped {
  margin-bottom: 0;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.16);
  border-radius: 0;
  background: transparent;
}

.directory-tile.grouped:last-child {
  border-bottom: 0;
}

.directory-tile:active {
  background-color: rgba(24, 54, 83, 0.05);
  transform: translateY(1rpx) scale(0.99);
}

.directory-tile-icon {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 72rpx;
  height: 72rpx;
  border-radius: 20rpx;
  background: rgba(24, 54, 83, 0.06);
}

.directory-tile-copy {
  flex: 1;
  min-width: 0;
}

.directory-tile-title {
  display: block;
  color: var(--tree-text-primary, #1e293b);
  font-size: 28rpx;
  font-weight: 700;
  line-height: 1.4;
}

.directory-tile-desc {
  display: block;
  margin-top: 6rpx;
  color: var(--tree-text-secondary, #64748b);
  font-size: 22rpx;
  line-height: 1.5;
}

.directory-arrow {
  flex-shrink: 0;
  color: var(--tree-text-secondary, #64748b);
  font-size: 32rpx;
  font-weight: 400;
  line-height: 1;
  transition: transform 180ms ease-out;
}

.directory-tile:active .directory-arrow {
  transform: translateX(4rpx);
}
</style>
