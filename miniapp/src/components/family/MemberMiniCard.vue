<template>
  <view class="member-mini-card">
    <view class="card-layout">
      <view class="avatar">{{ avatarLetter }}</view>
      <view class="card-body">
        <view class="card-head">
          <text class="name">{{ name }}</text>
          <view v-if="metaItems.length" class="meta-chips">
            <text
              v-for="item in metaItems"
              :key="item.key"
              class="meta-chip"
              :class="item.tone"
            >
              {{ item.text }}
            </text>
          </view>
          <view v-if="roleLabel || statusLabel" class="head-tags">
            <MiniStatusTag v-if="roleLabel" :label="roleLabel" tone="muted" />
            <MiniStatusTag v-if="statusLabel" :label="statusLabel" :tone="statusTone" />
          </view>
        </view>
        <view v-if="remark" class="remark">
          <text>{{ remark }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import MiniStatusTag from '../base/MiniStatusTag.vue'

type MetaItem = {
  key: string
  text: string
  tone?: 'active' | 'pending' | 'muted'
}

const props = defineProps<{
  name: string
  ageText?: string
  genderText?: string
  livingText?: string
  bindingText?: string
  showBinding?: boolean
  roleLabel?: string
  statusLabel?: string
  statusTone?: 'active' | 'pending' | 'danger' | 'muted'
  remark?: string
}>()

const avatarLetter = computed(() => props.name.trim().slice(0, 1) || '·')

const metaItems = computed<MetaItem[]>(() => {
  const items: MetaItem[] = []
  if (props.ageText) items.push({ key: 'age', text: props.ageText })
  if (props.genderText) items.push({ key: 'gender', text: props.genderText })
  if (props.livingText) {
    items.push({
      key: 'living',
      text: props.livingText,
      tone: props.livingText === '健在' ? 'active' : props.livingText === '已故' ? 'muted' : undefined
    })
  }
  if (props.showBinding && props.bindingText) {
    items.push({
      key: 'binding',
      text: props.bindingText,
      tone:
        props.bindingText === '已绑定'
          ? 'active'
          : props.bindingText === '未绑定' || props.bindingText === '待绑定'
            ? 'pending'
            : undefined
    })
  }
  return items
})
</script>

<style scoped>
.member-mini-card {
  margin-bottom: 10rpx;
  border: 1rpx solid var(--tree-border-subtle, #eef1f4);
  border-radius: var(--tree-radius-md, 18rpx);
  background: var(--tree-surface, #fff);
  padding: 16rpx 18rpx;
  box-shadow: var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.04));
}
.card-layout {
  display: flex;
  align-items: center;
  gap: 14rpx;
}
.avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 52rpx;
  height: 52rpx;
  border-radius: 50%;
  background: var(--tree-green-light);
  color: var(--tree-green);
  font-size: 24rpx;
  font-weight: 600;
}
.card-body {
  flex: 1;
  min-width: 0;
}
.card-head {
  display: flex;
  align-items: center;
  gap: 8rpx;
  min-width: 0;
}
.name {
  flex-shrink: 0;
  color: var(--tree-text);
  font-size: 28rpx;
  font-weight: 600;
  line-height: 1.35;
}
.meta-chips {
  display: flex;
  flex: 1;
  flex-wrap: wrap;
  align-items: center;
  gap: 6rpx;
  min-width: 0;
}
.meta-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999rpx;
  background: #f3f4f6;
  color: var(--tree-text-secondary);
  font-size: 20rpx;
  line-height: 1.2;
  padding: 4rpx 10rpx;
}
.meta-chip.active {
  background: var(--tree-green-light, #e8f3ee);
  color: var(--tree-green, #2f6b57);
}
.meta-chip.pending {
  background: var(--tree-gold-light, #f6eedc);
  color: var(--tree-gold-text, #8a6d2f);
}
.meta-chip.muted {
  background: #f3f4f6;
  color: var(--tree-text-weak, #9ca3af);
}
.head-tags {
  display: flex;
  flex-shrink: 0;
  gap: 6rpx;
}
.remark {
  margin-top: 6rpx;
  color: var(--tree-text-weak);
  font-size: 22rpx;
  line-height: 1.4;
}
</style>
