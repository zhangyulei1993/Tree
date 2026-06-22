<template>
  <view class="relation-list">
    <view v-for="(item, index) in items" :key="index" class="relation-item">
      <view class="relation-dot" />
      <text class="sentence">{{ item }}</text>
    </view>
    <view v-if="!items.length" class="empty-hint">
      <text class="tree-muted">暂无关系记录</text>
    </view>
  </view>
</template>

<script setup lang="ts">
defineProps<{ items: string[] }>()
</script>

<style scoped>
.relation-item {
  display: flex;
  align-items: flex-start;
  gap: 14rpx;
  padding: 14rpx 0;
  border-top: 1rpx solid var(--tree-border-subtle, #f0ebe3);
}
.relation-item:first-child {
  border-top: none;
  padding-top: 0;
}
.relation-dot {
  flex-shrink: 0;
  width: 10rpx;
  height: 10rpx;
  margin-top: 12rpx;
  border-radius: 50%;
  background: var(--tree-green);
}
.sentence {
  flex: 1;
  color: var(--tree-text);
  font-size: 26rpx;
  line-height: 1.65;
}
.empty-hint {
  padding: 20rpx 0;
  text-align: center;
}
</style>
