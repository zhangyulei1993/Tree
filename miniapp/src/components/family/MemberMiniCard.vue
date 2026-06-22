<template>
  <view class="member-mini-card">
    <view class="card-layout">
      <view class="avatar">{{ avatarLetter }}</view>
      <view class="card-body">
        <view class="card-head">
          <text class="name">{{ name }}</text>
          <MiniStatusTag v-if="roleLabel" :label="roleLabel" tone="muted" />
        </view>
        <view class="meta-row">
          <text v-if="genderLabel" class="meta">{{ genderLabel }}</text>
          <MiniStatusTag v-if="statusLabel" :label="statusLabel" :tone="statusTone" />
        </view>
        <text v-if="lifeInfo" class="meta">{{ lifeInfo }}</text>
        <text v-if="bindLabel" class="meta">{{ bindLabel }}</text>
        <text v-if="remark" class="remark">{{ remark }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import MiniStatusTag from '../base/MiniStatusTag.vue'

const props = defineProps<{
  name: string
  genderLabel?: string
  lifeInfo?: string
  bindLabel?: string
  roleLabel?: string
  statusLabel?: string
  statusTone?: 'active' | 'pending' | 'danger' | 'muted'
  remark?: string
}>()

const avatarLetter = computed(() => props.name.trim().slice(0, 1) || '·')
</script>

<style scoped>
.member-mini-card {
  margin-bottom: 14rpx;
  border: 1rpx solid var(--tree-border-subtle, #eef1f4);
  border-radius: var(--tree-radius-md, 18rpx);
  background: var(--tree-surface, #fff);
  padding: 22rpx;
  box-shadow: var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.04));
}
.card-layout {
  display: flex;
  gap: 18rpx;
}
.avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: var(--tree-green-light);
  color: var(--tree-green);
  font-size: 28rpx;
  font-weight: 600;
}
.card-body {
  flex: 1;
  min-width: 0;
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
}
.name {
  color: var(--tree-text);
  font-size: 28rpx;
  font-weight: 600;
}
.meta-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8rpx;
  margin-top: 8rpx;
}
.meta {
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.5;
}
.remark {
  display: block;
  margin-top: 8rpx;
  color: var(--tree-text-weak);
  font-size: 22rpx;
  line-height: 1.5;
}
</style>
