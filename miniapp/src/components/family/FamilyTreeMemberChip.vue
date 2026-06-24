<template>
  <view
    class="tree-member-chip"
    :class="[
      livingClass,
      {
        self: isSelf,
        compact
      }
    ]"
  >
    <view v-if="showBinding && bindingText && !isDeceased" class="binding-strip" :class="bindingClass" />
    <view class="chip-head">
      <view class="name-wrap">
        <text class="name">
          <template v-if="shouldEmphasizeSurname">
            <text class="surname">{{ surnamePart }}</text>{{ givenPart }}
          </template>
          <template v-else>{{ displayName }}</template>
        </text>
        <text v-if="isSelf" class="self-tag">我</text>
      </view>
    </view>
    <text v-if="kinshipTitle && !isSelf" class="kinship-title">{{ kinshipTitle }}</text>
    <view class="chip-meta">
      <text class="meta">{{ genderText }}</text>
      <view v-if="showBinding && bindingText && !isDeceased" class="binding-dot" :class="bindingClass" />
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import type { TreeNode } from '@/types/api'
import { formatMemberGender, formatTreeBindingState } from '@/utils/memberFormat'

const props = defineProps<{
  node: TreeNode
  isSelf?: boolean
  showBinding?: boolean
  kinshipTitle?: string
  compact?: boolean
  familySurname?: string
}>()

const displayName = computed(() => props.node.displayName.trim() || '未命名')
const surnamePart = computed(() => {
  const surname = (props.node.surname || displayName.value.slice(0, 1)).trim()
  if (!surname || !displayName.value.startsWith(surname)) return ''
  return surname
})
const givenPart = computed(() => displayName.value.slice(surnamePart.value.length))
const shouldEmphasizeSurname = computed(() => {
  const familySurname = props.familySurname?.trim()
  return Boolean(familySurname && surnamePart.value === familySurname)
})
const genderText = computed(() => formatMemberGender(props.node.gender))
const bindingText = computed(() =>
  props.showBinding ? formatTreeBindingState(props.node.userBindingState) : ''
)
const isDeceased = computed(() => props.node.isLiving === false)
const livingClass = computed(() => {
  if (props.node.isLiving === true) return 'living-active'
  if (isDeceased.value) return 'living-deceased'
  return 'living-unknown'
})

const bindingClass = computed(() => {
  if (bindingText.value === '已绑定') return 'tone-active'
  if (bindingText.value === '未绑定' || bindingText.value === '待绑定') return 'tone-pending'
  if (bindingText.value === '无需绑定') return 'tone-muted'
  return 'tone-unknown'
})
</script>

<style scoped>
.tree-member-chip {
  position: relative;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 144rpx;
  height: 144rpx;
  min-width: 144rpx;
  max-width: 144rpx;
  overflow: hidden;
  border: 1rpx solid var(--tree-border-subtle, #eef1f4);
  border-radius: 12rpx;
  background: #fff;
  padding: 9rpx 10rpx;
  box-shadow: var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.04));
}
.tree-member-chip.self {
  border-width: 2rpx;
  border-color: var(--tree-green, #2f6b57);
  box-shadow:
    0 0 0 2rpx rgba(47, 107, 87, 0.14),
    var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.06));
}
.tree-member-chip.compact {
  width: 112rpx;
  height: 112rpx;
  min-width: 112rpx;
  max-width: 112rpx;
  padding: 6rpx 7rpx;
  border-radius: 10rpx;
}
.tree-member-chip.compact .name {
  font-size: 19rpx;
  line-height: 1.25;
}
.tree-member-chip.compact .chip-meta {
  margin-top: 2rpx;
}
.tree-member-chip.compact .meta {
  font-size: 15rpx;
  line-height: 1.35;
}
.tree-member-chip.living-active {
  border-color: #8fc8a5;
  background: linear-gradient(180deg, #d8f1e2 0%, #f7fffa 100%);
}
.tree-member-chip.living-deceased {
  border-color: #a9b2bf;
  background: linear-gradient(180deg, #cfd5de 0%, #eef1f5 100%);
}
.tree-member-chip.living-unknown {
  border-color: #e1c589;
  background: linear-gradient(180deg, #ffedbf 0%, #fffaf0 100%);
}
.tree-member-chip.self.living-active {
  border-color: var(--tree-green, #2f6b57);
}
.tree-member-chip.self.living-deceased {
  border-color: #5f7a6d;
}
.tree-member-chip.self.living-unknown {
  border-color: var(--tree-green, #2f6b57);
}
.chip-head {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
}
.name-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5rpx;
  width: 100%;
  min-width: 0;
}
.name {
  display: block;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  color: var(--tree-text);
  font-size: 26rpx;
  font-weight: 500;
  line-height: 1.35;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.surname {
  font-weight: 800;
}
.binding-dot {
  display: inline-flex;
  flex-shrink: 0;
  width: 10rpx;
  height: 10rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  box-shadow: 0 2rpx 6rpx rgba(15, 23, 42, 0.12);
}
.binding-strip {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  height: 7rpx;
  border-radius: 10rpx 10rpx 0 0;
}
.binding-strip.tone-active,
.binding-dot.tone-active {
  background: #18c76f;
}
.binding-strip.tone-pending,
.binding-dot.tone-pending {
  background: var(--tree-gold, #d6ad60);
}
.binding-strip.tone-muted,
.binding-dot.tone-muted {
  background: #a8b0bb;
}
.binding-strip.tone-unknown,
.binding-dot.tone-unknown {
  background: #cbd5e1;
}
.self-tag {
  flex-shrink: 0;
  border-radius: 999rpx;
  background: var(--tree-green);
  color: #fff;
  font-size: 16rpx;
  line-height: 1;
  padding: 3rpx 8rpx;
}
.kinship-title {
  display: block;
  overflow: hidden;
  width: 100%;
  margin-top: 3rpx;
  color: var(--tree-green, #2f6b57);
  font-size: 18rpx;
  font-weight: 500;
  line-height: 1.25;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.living-deceased .kinship-title {
  color: #6b7280;
}
.tree-member-chip.compact .kinship-title {
  font-size: 15rpx;
  margin-top: 1rpx;
}
.chip-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5rpx;
  width: 100%;
  margin-top: 5rpx;
}
.meta {
  color: var(--tree-text-secondary);
  font-size: 20rpx;
  line-height: 1.4;
}
</style>
