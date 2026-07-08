<template>
  <view
    class="tree-member-chip"
    :class="[
      livingClass,
      nodeKindClass,
      {
        self: isSelf,
        compact,
        interactive
      }
    ]"
    @tap="handleSelect"
  >
    <view v-if="showBinding && bindingText && !isDeceased" class="binding-strip" :class="bindingClass" />
    <view class="chip-head">
      <view class="name-wrap">
        <text v-if="shouldEmphasizeSurname" class="name-emphasis">
          <text class="surname-seal">{{ surnamePart.slice(0, 1) }}</text>
          <text class="given-name" :class="genderNameClass">{{ surnamePart.slice(1) }}{{ givenPart }}</text>
        </text>
        <text v-else class="plain-name" :class="genderNameClass">{{ displayName }}</text>
        <text v-if="isSelf" class="self-tag">我</text>
      </view>
    </view>
    <text v-if="kinshipTitle && !isSelf" class="kinship-title">{{ kinshipTitle }}</text>
    <view v-if="isSelf || (showBinding && bindingText && !isDeceased)" class="chip-meta">
      <view v-if="isSelf" class="gender-dot" :class="genderDotClass" />
      <view v-if="showBinding && bindingText && !isDeceased" class="binding-dot" :class="bindingClass" />
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import type { TreeNode } from '@/types/api'
import { formatTreeBindingState } from '@/utils/memberFormat'

const props = defineProps<{
  node: TreeNode
  isSelf?: boolean
  showBinding?: boolean
  kinshipTitle?: string
  compact?: boolean
  familySurname?: string
  interactive?: boolean
}>()

const emit = defineEmits<{
  select: [node: TreeNode]
}>()

function handleSelect() {
  if (props.interactive) emit('select', props.node)
}

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
const genderNameClass = computed(() => {
  if (props.node.gender === 'MALE') return 'name-male'
  if (props.node.gender === 'FEMALE') return 'name-female'
  return 'name-unknown'
})
const genderDotClass = computed(() => {
  if (props.node.gender === 'MALE') return 'gender-male'
  if (props.node.gender === 'FEMALE') return 'gender-female'
  return 'gender-unknown'
})
const bindingText = computed(() =>
  props.showBinding ? formatTreeBindingState(props.node.userBindingState) : ''
)
const isDeceased = computed(() => props.node.isLiving === false)
const livingClass = computed(() => {
  if (props.node.isLiving === true) return 'living-active'
  if (isDeceased.value) return 'living-deceased'
  return 'living-unknown'
})
const nodeKindClass = computed(() => {
  if (props.node.memberType === 'SPOUSE') return 'spouse-tag'
  if (props.node.memberType === 'EXTERNAL_MEMBER') return 'external-tag'
  if (props.node.memberType === 'LINEAGE_MEMBER') return 'lineage-tag'
  return 'lineage-tag'
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
  border: 1rpx solid var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  border-radius: 4rpx;
  background-color: #fff7e8;
  background-image: linear-gradient(rgba(92, 74, 48, 0.028) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
  padding: 10rpx 9rpx;
  box-shadow: none;
}
.tree-member-chip.self {
  border-color: var(--archive-blue, #163353);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.08), transparent),
    var(--archive-blue, #163353);
  box-shadow: none;
}
.tree-member-chip.interactive {
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  box-shadow: inset 0 0 0 2rpx rgba(255, 255, 255, 0.34);
}
.tree-member-chip.interactive::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 34rpx;
  height: 34rpx;
  border-radius: 0 0 0 16rpx;
  background: var(--archive-cinnabar, #a83b2d);
}
.tree-member-chip.interactive::after {
  content: '管';
  position: absolute;
  top: 2rpx;
  right: 6rpx;
  color: #fff;
  font-size: 16rpx;
  font-weight: 700;
  line-height: 1;
}
.tree-member-chip.compact {
  width: 112rpx;
  height: 112rpx;
  min-width: 112rpx;
  max-width: 112rpx;
  padding: 6rpx 7rpx;
  border-radius: 10rpx;
}
.tree-member-chip.compact .plain-name,
.tree-member-chip.compact .name-emphasis,
.tree-member-chip.compact .given-name {
  font-size: 19rpx;
  line-height: 1.25;
}
.tree-member-chip.compact .gender-dot {
  width: 8rpx;
  height: 8rpx;
}
.tree-member-chip.compact .chip-meta {
  margin-top: 2rpx;
}
.tree-member-chip.compact .meta {
  font-size: 15rpx;
  line-height: 1.35;
}
.tree-member-chip.living-active {
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
}
.tree-member-chip.living-deceased {
  border-color: rgba(101, 112, 128, 0.32);
  background-color: #eee7d8;
  background-image: linear-gradient(rgba(92, 74, 48, 0.026) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
}
.tree-member-chip.living-unknown {
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
}
.tree-member-chip.self.living-active {
  border-color: var(--archive-blue, #163353);
}
.tree-member-chip.self.living-deceased {
  border-color: var(--archive-blue, #163353);
}
.tree-member-chip.self.living-unknown {
  border-color: var(--archive-blue, #163353);
}
.tree-member-chip.lineage-tag {
  border-style: solid;
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  background-color: #fff7e8;
  background-image: linear-gradient(rgba(92, 74, 48, 0.028) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
}
.tree-member-chip.spouse-tag {
  border-style: dashed;
  background-color: #fbf0dc;
  background-image: linear-gradient(rgba(92, 74, 48, 0.022) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
  opacity: 0.92;
}
.tree-member-chip.external-tag {
  background: #f2eadc;
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
.name-emphasis,
.plain-name {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
  max-width: 100%;
  min-width: 0;
}
.plain-name {
  display: block;
  overflow: hidden;
  font-size: 26rpx;
  font-weight: 500;
  line-height: 1.35;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.plain-name.name-male,
.given-name.name-male {
  color: var(--archive-blue, #163353);
}
.plain-name.name-female,
.given-name.name-female {
  color: var(--archive-cinnabar, #a83b2d);
}
.plain-name.name-unknown,
.given-name.name-unknown {
  color: var(--archive-ink-soft, #657080);
}
.living-deceased .plain-name,
.living-deceased .given-name {
  opacity: 0.62;
}
.surname-seal {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 28rpx;
  height: 28rpx;
  border: 1rpx solid var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  background: rgba(255, 248, 234, 0.88);
  color: var(--archive-ink, #243244);
  font-size: 18rpx;
  font-weight: 800;
  line-height: 1;
}
.given-name {
  overflow: hidden;
  min-width: 0;
  font-size: 26rpx;
  font-weight: 500;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tree-member-chip.compact .surname-seal {
  width: 22rpx;
  height: 22rpx;
  font-size: 15rpx;
}
.tree-member-chip.compact .given-name {
  font-size: 19rpx;
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
  background: var(--archive-cinnabar, #a83b2d);
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
  color: var(--archive-blue, #163353);
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
.gender-dot {
  display: inline-flex;
  flex-shrink: 0;
  width: 10rpx;
  height: 10rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: 50%;
}
.gender-dot.gender-male {
  background: var(--archive-blue, #163353);
}
.gender-dot.gender-female {
  background: var(--archive-cinnabar, #a83b2d);
}
.gender-dot.gender-unknown {
  background: #a8b0bb;
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
  color: var(--archive-ink-soft, #657080);
  font-size: 20rpx;
  line-height: 1.4;
}
.tree-member-chip.self .plain-name,
.tree-member-chip.self .given-name,
.tree-member-chip.self .surname-seal,
.tree-member-chip.self .kinship-title {
  color: rgba(255, 255, 255, 0.92);
}
.tree-member-chip.self .surname-seal {
  border-color: rgba(255, 255, 255, 0.42);
  background: rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.95);
}
.tree-member-chip.self .gender-dot {
  border-color: rgba(255, 255, 255, 0.55);
}
.tree-member-chip.self .gender-dot.gender-unknown {
  background: rgba(255, 255, 255, 0.72);
}
</style>
