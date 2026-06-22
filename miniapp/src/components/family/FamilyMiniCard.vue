<template>
  <view class="family-mini-card" @click="$emit('click')">
    <view class="card-layout">
      <view class="family-avatar">
        <text class="avatar-char">{{ sealText }}</text>
      </view>
      <view class="card-body">
        <view class="card-head">
          <text class="name">{{ name }}</text>
          <MiniStatusTag v-if="statusLabel" :label="statusLabel" :tone="statusTone" />
        </view>
        <view v-if="surname || region" class="meta-row">
          <text v-if="surname" class="meta-chip">{{ surname }}氏</text>
          <text v-if="region" class="meta-chip muted">{{ region }}</text>
        </view>
        <text v-if="roleLabel" class="role">{{ roleLabel }}</text>
        <text v-if="desc" class="desc">{{ desc }}</text>
        <slot />
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import MiniStatusTag from '../base/MiniStatusTag.vue'

const props = defineProps<{
  name: string
  surname?: string
  region?: string
  roleLabel?: string
  desc?: string
  statusLabel?: string
  statusTone?: 'active' | 'pending' | 'danger' | 'muted'
}>()

defineEmits<{ click: [] }>()

const sealText = computed(() => (props.surname?.trim() || props.name.trim()).slice(0, 1) || '家')
</script>

<style scoped>
.family-mini-card {
  margin-bottom: 16rpx;
  border: 1rpx solid var(--tree-border-subtle, #eef1f4);
  border-radius: var(--tree-radius-lg, 24rpx);
  background: var(--tree-surface, #fff);
  padding: 22rpx 24rpx;
  box-shadow: var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.04));
}

.family-mini-card:active {
  background: var(--tree-surface-muted, #f8fafc);
}

.card-layout {
  display: flex;
  gap: 18rpx;
}

.family-avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 72rpx;
  height: 72rpx;
  border-radius: 20rpx;
  background: linear-gradient(145deg, var(--tree-accent-blue-light, #eef4fb) 0%, var(--tree-green-light, #e8f3ee) 100%);
}

.avatar-char {
  color: var(--tree-primary, #1f3a5f);
  font-size: 32rpx;
  font-weight: 700;
}

.card-body {
  flex: 1;
  min-width: 0;
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12rpx;
}

.name {
  flex: 1;
  color: var(--tree-text-primary, #1e293b);
  font-size: 30rpx;
  font-weight: 600;
  line-height: 1.35;
}

.meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 10rpx;
}

.meta-chip {
  display: inline-flex;
  border-radius: 8rpx;
  background: var(--tree-green-light, #e8f3ee);
  color: var(--tree-green, #2f6b57);
  padding: 4rpx 12rpx;
  font-size: 22rpx;
  font-weight: 500;
}

.meta-chip.muted {
  background: var(--tree-surface-muted, #f1f5f9);
  color: var(--tree-text-secondary, #64748b);
}

.role {
  display: block;
  margin-top: 10rpx;
  color: var(--tree-text-secondary, #64748b);
  font-size: 24rpx;
}

.desc {
  display: block;
  margin-top: 8rpx;
  color: var(--tree-text-weak, #94a3b8);
  font-size: 24rpx;
  line-height: 1.6;
}
</style>
