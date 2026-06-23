<template>
  <view class="profile-header">
    <image v-if="avatarUrl" class="avatar avatar-image" :src="avatarUrl" mode="aspectFill" />
    <view v-else class="avatar">{{ avatarText }}</view>
    <view class="info">
      <text class="name">{{ name }}</text>
      <text v-if="subtitle" class="subtitle">{{ subtitle }}</text>
      <view v-if="tags?.length" class="tags">
        <MiniStatusTag v-for="(tag, i) in tags" :key="i" :label="tag.label" :tone="tag.tone" />
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import MiniStatusTag from '../base/MiniStatusTag.vue'

defineProps<{
  name: string
  subtitle?: string
  avatarText?: string
  avatarUrl?: string
  tags?: Array<{ label: string; tone?: 'active' | 'pending' | 'danger' | 'muted' }>
}>()
</script>

<style scoped>
.profile-header {
  display: flex;
  align-items: center;
  gap: 22rpx;
}
.avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: linear-gradient(145deg, var(--tree-primary, #1f3a5f) 0%, var(--tree-green, #2f6b57) 100%);
  color: #fff;
  font-size: 38rpx;
  font-weight: 600;
}
.avatar-image {
  background: #eef4f1;
}
.name {
  display: block;
  color: var(--tree-text);
  font-size: 34rpx;
  font-weight: 600;
  line-height: 1.35;
}
.subtitle {
  display: block;
  margin-top: 6rpx;
  color: var(--tree-text-secondary);
  font-size: 26rpx;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 12rpx;
}
</style>
