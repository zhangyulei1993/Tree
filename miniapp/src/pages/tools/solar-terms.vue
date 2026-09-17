<template>
  <view class="archive-page solar-terms-page">
    <MiniBackHome />

    <view class="solar-terms-head archive-page-head">
      <view>
        <text class="archive-kicker">SEASONAL MARKERS</text>
        <text class="archive-title">农历节气</text>
        <text class="archive-subtitle">顺着节气看一年四时变化</text>
      </view>
      <view class="archive-seal">时</view>
    </view>

    <view v-if="nextSolarTerm" class="solar-term-feature archive-panel" :class="{ 'is-today': todaySolarTerm }">
      <view class="solar-term-feature-head">
        <view>
          <text class="archive-kicker">NEXT SEASON</text>
          <text class="section-title">最近节气</text>
        </view>
        <text class="solar-term-countdown">{{ solarTermCountdown }}</text>
      </view>
      <view class="solar-term-feature-main">
        <view class="solar-term-date-block">
          <text>{{ monthOf(nextSolarTerm.date) }}月</text>
          <text>{{ dayOf(nextSolarTerm.date) }}</text>
        </view>
        <view class="solar-term-copy">
          <text class="solar-term-name">{{ nextSolarTerm.name }}</text>
          <text class="solar-term-date">{{ formatDate(nextSolarTerm.date) }}</text>
          <text class="solar-term-description">{{ nextSolarTerm.description }}</text>
        </view>
      </view>
    </view>

    <MiniNotice tone="info" class="solar-term-notice">
      节气日期按 2026 年公历换算展示，具体交节时刻会因年份和时区略有差异。
    </MiniNotice>

    <view class="solar-term-section">
      <view class="solar-term-section-head">
        <view>
          <text class="archive-kicker">2026</text>
          <text class="section-title">二十四节气</text>
        </view>
        <text class="section-count">{{ solarTerms.length }} 个</text>
      </view>
      <view class="solar-term-grid">
        <view
          v-for="item in solarTerms"
          :key="item.name"
          class="solar-term-card"
          :class="solarTermCardClass(item)"
        >
          <view class="solar-term-card-date">
            <text>{{ monthOf(item.date) }}月</text>
            <text>{{ dayOf(item.date) }}</text>
          </view>
          <view class="solar-term-card-copy">
            <view class="solar-term-card-title-row">
              <text class="solar-term-card-name">{{ item.name }}</text>
              <text v-if="item.date === todayKey" class="solar-term-status">今天</text>
              <text v-else-if="isNextSolarTerm(item)" class="solar-term-status upcoming">即将到来</text>
            </view>
            <text class="solar-term-card-description">{{ item.description }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed } from 'vue'

import { ensureToolEnabled } from '@/api/toolConfigs'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { dateKey, getSolarTermItems, type SolarTermItem } from '@/features/festivals/festivalData'

const currentYear = 2026
const solarTerms = getSolarTermItems(currentYear)
const todayKey = dateKey(new Date())
const todaySolarTerm = solarTerms.find((item) => item.date === todayKey) || null
const nextSolarTerm = computed(() => solarTerms.find((item) => item.date >= todayKey) || null)
const solarTermCountdown = computed(() => {
  if (!nextSolarTerm.value) return ''
  const days = daysUntil(nextSolarTerm.value.date)
  return days === 0 ? '今天' : `${days} 天后`
})

onLoad(() => {
  void ensureToolEnabled('SOLAR_TERMS')
})

function monthOf(date: string) {
  return Number(date.slice(5, 7))
}

function dayOf(date: string) {
  return Number(date.slice(8, 10))
}

function formatDate(date: string) {
  return `${monthOf(date)}月${dayOf(date)}日`
}

function daysUntil(date: string) {
  const [year, month, day] = date.split('-').map(Number)
  const target = new Date(year, month - 1, day).getTime()
  const today = new Date()
  const todayStart = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime()
  return Math.max(0, Math.round((target - todayStart) / 86400000))
}

function isNextSolarTerm(item: SolarTermItem) {
  return !todaySolarTerm && nextSolarTerm.value?.date === item.date
}

function solarTermCardClass(item: SolarTermItem) {
  return {
    'is-today': item.date === todayKey,
    'is-upcoming': isNextSolarTerm(item)
  }
}
</script>

<style scoped>
.solar-terms-page {
  padding-top: 28rpx;
}

.solar-term-notice {
  margin-top: 24rpx;
}

.solar-term-feature {
  margin-top: 18rpx;
  border-top: 4rpx solid var(--archive-blue);
  padding: 22rpx 24rpx 20rpx;
  background: rgba(35, 73, 98, 0.055);
}

.solar-term-feature.is-today {
  border-color: var(--archive-cinnabar);
  background: rgba(168, 59, 45, 0.07);
}

.solar-term-feature-head,
.solar-term-feature-main,
.solar-term-card-title-row {
  display: flex;
  align-items: center;
}

.solar-term-feature-head,
.solar-term-section-head {
  justify-content: space-between;
  gap: 16rpx;
}

.solar-term-feature-head {
  display: flex;
}

.solar-term-countdown {
  color: var(--archive-blue);
  font-size: 21rpx;
  font-weight: 700;
}

.solar-term-feature-main {
  gap: 18rpx;
  margin-top: 18rpx;
}

.solar-term-date-block,
.solar-term-card-date {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-blue);
  color: var(--archive-blue);
  background: rgba(255, 252, 244, 0.64);
}

.solar-term-date-block {
  width: 100rpx;
  height: 100rpx;
}

.solar-term-date-block text:first-child,
.solar-term-card-date text:first-child {
  font-size: 19rpx;
}

.solar-term-date-block text:last-child,
.solar-term-card-date text:last-child {
  margin-top: 2rpx;
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 38rpx;
  font-weight: 800;
  line-height: 1;
}

.solar-term-copy,
.solar-term-name,
.solar-term-date,
.solar-term-description,
.solar-term-card-copy,
.solar-term-card-name,
.solar-term-card-description {
  display: block;
}

.solar-term-copy,
.solar-term-card-copy {
  min-width: 0;
}

.solar-term-name {
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 800;
}

.solar-term-date {
  margin-top: 4rpx;
  color: var(--archive-blue);
  font-size: 20rpx;
}

.solar-term-description {
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.5;
}

.solar-term-section {
  margin-top: 34rpx;
}

.solar-term-section-head {
  display: flex;
  align-items: center;
}

.solar-term-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12rpx;
  margin-top: 14rpx;
}

.solar-term-card {
  display: flex;
  align-items: center;
  gap: 12rpx;
  min-height: 116rpx;
  border: 1rpx solid var(--archive-line);
  background: rgba(255, 252, 244, 0.54);
  padding: 12rpx;
  box-sizing: border-box;
}

.solar-term-card.is-today {
  border-color: rgba(168, 59, 45, 0.55);
  background: rgba(168, 59, 45, 0.07);
}

.solar-term-card.is-upcoming {
  border-color: rgba(35, 73, 98, 0.45);
  background: rgba(35, 73, 98, 0.055);
}

.solar-term-card-date {
  width: 62rpx;
  height: 68rpx;
}

.solar-term-card-date text:last-child {
  font-size: 27rpx;
}

.solar-term-card-title-row {
  gap: 6rpx;
  min-width: 0;
}

.solar-term-card-name {
  color: var(--archive-ink);
  font-size: 24rpx;
  font-weight: 750;
}

.solar-term-status {
  flex-shrink: 0;
  color: var(--archive-cinnabar);
  font-size: 17rpx;
}

.solar-term-status.upcoming {
  color: var(--archive-blue);
}

.solar-term-card-description {
  display: -webkit-box;
  overflow: hidden;
  margin-top: 4rpx;
  color: var(--archive-ink-soft);
  font-size: 18rpx;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
</style>
