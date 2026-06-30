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
      <view class="card-tail">
        <text class="card-arrow">›</text>
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
  position: relative;
  margin-bottom: 18rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: 32rpx;
  background:
    radial-gradient(circle at 100% 0%, rgba(216, 175, 104, 0.14), transparent 180rpx),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(250, 246, 238, 0.94) 50%, rgba(244, 250, 247, 0.94) 100%);
  padding: 26rpx 24rpx;
  overflow: hidden;
  box-shadow: 0 18rpx 44rpx rgba(24, 54, 83, 0.08);
}

.family-mini-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 28rpx;
  bottom: 28rpx;
  width: 7rpx;
  border-radius: 0 999rpx 999rpx 0;
  background: linear-gradient(180deg, var(--tree-primary) 0%, var(--tree-green) 100%);
}

.family-mini-card:active {
  opacity: 0.92;
  transform: scale(0.995);
}

.card-layout {
  display: flex;
  align-items: center;
  gap: 18rpx;
}

.family-avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 82rpx;
  height: 82rpx;
  border-radius: 26rpx;
  background: linear-gradient(145deg, #17304c 0%, #2f6b57 100%);
  box-shadow:
    inset 0 1rpx 0 rgba(255, 255, 255, 0.22),
    0 14rpx 30rpx rgba(24, 54, 83, 0.14);
}

.avatar-char {
  color: #fff;
  font-size: 34rpx;
  font-weight: 800;
}

.card-body {
  flex: 1;
  min-width: 0;
}

.card-tail {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  align-self: stretch;
}

.card-arrow {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40rpx;
  height: 40rpx;
  margin-top: 4rpx;
  border-radius: 999rpx;
  background: rgba(24, 54, 83, 0.08);
  color: var(--tree-primary);
  font-size: 28rpx;
  font-weight: 700;
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
  font-size: 32rpx;
  font-weight: 800;
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
