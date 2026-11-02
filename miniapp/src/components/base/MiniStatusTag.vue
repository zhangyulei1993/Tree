<template>
  <text class="mini-status-tag" :class="tone">{{ displayLabel }}</text>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import { statusTagTone } from './formatStatus'

const props = defineProps<{
  status?: string
  label?: string
  tone?: 'active' | 'pending' | 'danger' | 'muted'
}>()

const displayLabel = computed(() => props.label || '未知')
const tone = computed(() => props.tone || statusTagTone(props.status))
</script>

<style scoped>
.mini-status-tag {
  display: inline-flex;
  align-items: center;
  border-radius: 999rpx;
  padding: 8rpx 16rpx;
  font-size: 22rpx;
  font-weight: 600;
}
.mini-status-tag.active {
  background: var(--tree-green-light, #e8f3ee);
  color: var(--tree-green, #2f6b57);
}
.mini-status-tag.pending {
  background: var(--tree-gold-light, #f6eedc);
  color: var(--tree-gold-text, #8a6d2f);
}
.mini-status-tag.danger {
  background: #fde8e6;
  color: #c0392b;
}
.mini-status-tag.muted {
  background: #f3f4f6;
  color: #6b7280;
}
</style>
