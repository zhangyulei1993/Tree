<template>
  <view class="archive-page tools-page">
    <MiniBackHome />

    <view class="tools-head archive-page-head">
      <view>
        <text class="archive-kicker">Tree Tools</text>
        <text class="archive-title">常用工具</text>
        <text class="archive-subtitle">家庭记录中的实用小工具</text>
      </view>
      <view class="archive-seal">用</view>
    </view>

    <view class="tool-grid">
      <view
        v-for="tool in visibleToolItems"
        :key="tool.key"
        class="tool-card"
        :class="[toolCardClass(tool.key), { 'is-tool-highlighted': isToolHighlighted(tool.key), 'is-tool-disabled': !isToolEnabled(tool.key) }]"
        @click="openCommonTool(tool.key)"
      >
        <view v-if="tool.key === 'KINSHIP_QUERY'" class="tool-icon relationship-tool-icon" aria-hidden="true">
          <view class="tool-icon-line line-one" />
          <view class="tool-icon-line line-two" />
          <view class="tool-icon-node node-top" />
          <view class="tool-icon-node node-left" />
          <view class="tool-icon-node node-right" />
        </view>
        <view v-else class="tool-icon">{{ tool.icon }}</view>
        <text class="tool-title">{{ toolTitle(tool.key) }}</text>
        <text class="tool-desc">{{ toolDescription(tool.key) }}</text>
        <text class="tool-arrow">›</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed } from 'vue'

import {
  getVisibleToolDefinitions,
  getToolDefinition,
  getToolDescription,
  getToolDisplayName,
  isToolEnabled,
  isToolHighlighted,
  loadToolConfigs
} from '@/api/toolConfigs'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import { dateKey, festivalDateLabel, getFestivalItems } from '@/features/festivals/festivalData'

const festivalItems = getFestivalItems(2026)
const todayKey = dateKey(new Date())
const todayFestival = festivalItems.find((item) => item.date === todayKey)
const nextFestival = festivalItems.find((item) => item.date > todayKey)
const festivalToolClass = computed(() => ({
  'is-festival-today': Boolean(todayFestival),
  'is-festival-upcoming': !todayFestival && Boolean(nextFestival)
}))
const festivalToolDescription = computed(() =>
  todayFestival
    ? `今天是${todayFestival.name} · ${festivalDateLabel(todayFestival.date)}`
    : nextFestival
      ? `${nextFestival.name} · ${festivalDateLabel(nextFestival.date)}即将到来`
      : '查看节日与法定假期'
)
const visibleToolItems = computed(() => getVisibleToolDefinitions())

function go(url: string) {
  uni.navigateTo({ url })
}

function toolTitle(key: string) {
  return getToolDisplayName(key)
}

function toolDescription(key: string) {
  if (!isToolEnabled(key)) return '暂未开放'
  if (key === 'TRADITIONAL_FESTIVALS' && (todayFestival || nextFestival)) return festivalToolDescription.value
  if (!getToolDefinition(key)?.path) return '功能暂未接入'
  return getToolDescription(key)
}

function toolCardClass(key: string) {
  return key === 'TRADITIONAL_FESTIVALS' ? festivalToolClass.value : {}
}

function openCommonTool(key: string) {
  if (!isToolEnabled(key)) {
    uni.showToast({ title: '该工具暂未开放', icon: 'none' })
    return
  }
  const definition = getToolDefinition(key)
  if (!definition?.path) {
    uni.showToast({ title: '该工具暂未接入', icon: 'none' })
    return
  }
  go(definition.path)
}

onLoad(() => {
  void loadToolConfigs()
})

onShow(() => {
  void loadToolConfigs({ force: true })
})
</script>

<style scoped>
.tools-page {
  padding-top: 28rpx;
}

.tool-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18rpx;
  margin-top: 28rpx;
}

.tool-card {
  position: relative;
  min-height: 214rpx;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 8rpx;
  background: rgba(255, 252, 244, 0.56);
  padding: 22rpx 20rpx 20rpx;
  box-sizing: border-box;
}

.tool-card.is-tool-highlighted {
  border-color: rgba(168, 59, 45, 0.72);
  box-shadow: 0 4rpx 0 rgba(168, 59, 45, 0.16);
}

.tool-card.is-tool-disabled {
  opacity: 0.62;
}

.tool-card.is-festival-today {
  border-color: rgba(168, 59, 45, 0.55);
  background: rgba(168, 59, 45, 0.07);
}

.tool-card.is-festival-today .tool-title,
.tool-card.is-festival-today .tool-desc {
  color: var(--archive-cinnabar);
}

.tool-card.is-festival-upcoming {
  border-color: rgba(35, 73, 98, 0.42);
  background: rgba(35, 73, 98, 0.055);
}

.tool-card.is-festival-upcoming .tool-title {
  color: var(--archive-blue);
}

.tool-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 58rpx;
  height: 58rpx;
  border: 1rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  line-height: 1;
}

.relationship-tool-icon {
  position: relative;
  border-color: var(--archive-line-strong);
  border-radius: 50%;
  background: rgba(255, 252, 244, 0.72);
}

.tool-icon-line {
  position: absolute;
  top: 28rpx;
  left: 28rpx;
  width: 25rpx;
  height: 2rpx;
  background: var(--archive-cinnabar);
  transform-origin: left center;
}

.relationship-tool-icon .line-one {
  transform: rotate(148deg);
}

.relationship-tool-icon .line-two {
  transform: rotate(32deg);
}

.tool-icon-node {
  position: absolute;
  width: 16rpx;
  height: 16rpx;
  border: 2rpx solid var(--archive-blue);
  border-radius: 50%;
  background: var(--archive-paper-light);
}

.relationship-tool-icon .node-top {
  top: 7rpx;
  left: 21rpx;
}

.relationship-tool-icon .node-left {
  bottom: 8rpx;
  left: 7rpx;
}

.relationship-tool-icon .node-right {
  right: 7rpx;
  bottom: 8rpx;
}

.tool-title,
.tool-desc {
  display: block;
}

.tool-title {
  margin-top: 20rpx;
  color: var(--archive-ink);
  font-size: 27rpx;
  font-weight: 700;
  line-height: 1.3;
}

.tool-desc {
  margin-top: 7rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.4;
}

.tool-arrow {
  position: absolute;
  right: 18rpx;
  bottom: 17rpx;
  color: var(--archive-cinnabar);
  font-size: 32rpx;
  line-height: 1;
}
</style>
