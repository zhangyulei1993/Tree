<template>
  <view class="export-lineage-tree">
    <view v-for="group in lineage.generations" :key="group.generation" class="generation-section">
      <view class="generation-head">
        <text>第 {{ group.generation }} 代</text>
        <view class="generation-rule" />
      </view>
      <view class="generation-members">
        <view v-for="item in group.members" :key="item.id" class="lineage-member">
          <text class="member-name">{{ display(item).name }}</text>
          <text v-if="display(item).yearText" class="member-year">{{ display(item).yearText }}</text>
          <text v-for="label in display(item).parentLabels" :key="label" class="parent-note">{{ label }}</text>
          <text v-for="label in display(item).spouseLabels" :key="label" class="spouse-note">（{{ label }}）</text>
        </view>
      </view>
    </view>
    <view v-if="lineage.unlocatedMembers.length" class="unlocated-note">
      <text>待补关系：</text>
      <text>{{ unlocatedText }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import { buildExportLineageTree, type ExportLineageMode } from '@/features/family-tree/exportLineageTree'
import { getExportLineageDisplay, type ExportLineageScope } from '@/features/family-tree/exportLineagePresentation'
import type { FamilyTreeResult } from '@/types/api'

const props = withDefaults(defineProps<{ tree: FamilyTreeResult; scope?: ExportLineageScope; memberMode?: ExportLineageMode }>(), { scope: 'INTERNAL', memberMode: 'SURNAME' })
const lineage = computed(() => buildExportLineageTree(props.tree, props.memberMode))
const unlocatedText = computed(() => lineage.value.unlocatedMembers.map((member) => props.scope === 'PUBLIC' ? `${member.displayName.slice(0, 1)}*` : member.displayName).join('、'))
function display(item: typeof lineage.value.generations[number]['members'][number]) { return getExportLineageDisplay(item, props.scope) }
</script>

<style scoped>
.export-lineage-tree{margin-top:26rpx;border-top:1rpx solid #c9a56b;border-bottom:1rpx solid #c9a56b}.generation-section{padding:25rpx 0 4rpx}.generation-head{display:flex;align-items:center;gap:14rpx;color:#9e4c3b;font-size:24rpx}.generation-rule{height:1rpx;flex:1;background:#c9a56b}.generation-members{display:grid;margin-top:16rpx;grid-template-columns:repeat(2,minmax(0,1fr));gap:14rpx}.lineage-member{box-sizing:border-box;display:flex;min-height:132rpx;flex-direction:column;justify-content:center;padding:17rpx 16rpx;border:1rpx solid rgba(184,138,82,.58);background:rgba(255,255,255,.32)}.member-name{color:#193856;font-size:30rpx;font-weight:700;line-height:1.25}.member-year{margin-top:8rpx;color:#718096;font-size:21rpx;line-height:1.2}.parent-note{margin-top:6rpx;color:#62768d;font-size:20rpx;line-height:1.2}.spouse-note{margin-top:6rpx;color:#855f4e;font-size:20rpx;line-height:1.2}.unlocated-note{padding:16rpx 0;color:#9e4c3b;font-size:22rpx;line-height:1.5}
</style>
