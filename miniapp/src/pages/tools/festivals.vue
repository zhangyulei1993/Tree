<template>
  <view class="archive-page festivals-page">
    <MiniBackHome />

    <view class="festivals-head archive-page-head">
      <view>
        <text class="archive-kicker">Chinese Calendar</text>
        <text class="archive-title">传统节日</text>
        <text class="archive-subtitle">2026年传统节日与法定假期安排</text>
      </view>
      <view class="archive-seal">节</view>
    </view>

    <view v-if="selectedFestival" id="festival-detail" class="festival-detail-panel archive-panel">
      <view class="festival-detail-head">
        <view>
          <text class="festival-modal-kicker">FESTIVAL NOTE</text>
          <text class="festival-modal-title">{{ selectedFestival.name }}</text>
        </view>
        <view class="festival-detail-close" @click="closeFestival">收起</view>
      </view>
      <view class="festival-modal-date">
        <text>{{ formatDate(selectedFestival.date) }}</text>
        <text>{{ selectedFestival.lunarDate }}</text>
      </view>
      <text class="festival-modal-description">{{ selectedFestival.description }}</text>
      <view class="festival-modal-detail">
        <text class="festival-detail-label">节日简介</text>
        <text class="festival-detail-text">{{ selectedFestival.introduction }}</text>
        <text class="festival-detail-label">常见习俗</text>
        <text class="festival-detail-text">{{ selectedFestival.customs }}</text>
      </view>
      <text v-if="selectedFestival.holidayRange" class="festival-modal-line">{{ selectedFestival.holidayRange }}</text>
      <text v-if="selectedFestival.workdays" class="festival-modal-line">{{ selectedFestival.workdays }}</text>

      <view class="festival-detail-actions">
        <MiniButton size="sm" @click="showCopyComposer = !showCopyComposer">
          {{ showCopyComposer ? '收起文案' : '生成节日文案' }}
        </MiniButton>
        <view class="festival-detail-link" @click="closeFestival">返回节日列表</view>
      </view>

      <view v-if="showCopyComposer" class="festival-copy-composer">
        <text class="festival-copy-title">生成节日文案</text>
        <text class="festival-copy-hint">文案在当前设备生成，不会上传家庭信息。</text>

        <text class="festival-copy-label">文案类型</text>
        <view class="festival-copy-options">
          <text
            v-for="option in festivalCopySceneOptions"
            :key="option.value"
            class="festival-copy-option"
            :class="{ active: selectedScene === option.value }"
            @click="selectedScene = option.value"
          >
            {{ option.label }}
          </text>
        </view>

        <view class="festival-copy-preview">
          <text>{{ festivalCopyText }}</text>
        </view>
        <view class="festival-copy-actions">
          <MiniButton variant="secondary" size="sm" @click="nextFestivalCopy">换一条</MiniButton>
          <MiniButton size="sm" @click="copyFestivalCopy">复制文案</MiniButton>
        </view>
      </view>
    </view>

    <view
      v-if="featuredFestival"
      class="next-festival archive-panel"
      :class="{ 'is-today': todayFestival }"
      @click="openFestival(featuredFestival)"
    >
      <view class="next-heading">
        <text class="next-label">{{ todayFestival ? '今天的节日' : '最近的节日' }}</text>
        <text class="next-countdown">{{ featuredCountdown }}</text>
      </view>
      <view class="next-main">
        <view class="next-date-block">
          <text class="next-month">{{ monthOf(featuredFestival.date) }}月</text>
          <text class="next-day">{{ dayOf(featuredFestival.date) }}</text>
        </view>
        <view class="next-copy">
          <text class="next-name">{{ featuredFestival.name }}</text>
          <text class="next-date">{{ formatDate(featuredFestival.date) }} · {{ featuredFestival.lunarDate }}</text>
          <text class="next-desc">{{ featuredFestival.description }}</text>
        </view>
      </view>
      <view class="next-footer">
        <text>查看节日介绍，生成分享文案</text>
        <text>→</text>
      </view>
    </view>

    <view v-if="upcomingFestivals.length > 0" class="recent-section">
      <view class="festival-section-head recent-section-head">
        <view>
          <text class="archive-kicker">UP NEXT</text>
          <text class="section-title">接下来</text>
        </view>
        <text class="section-count">近 {{ upcomingFestivals.length }} 项</text>
      </view>
      <view class="recent-grid">
        <view
          v-for="item in upcomingFestivals"
          :key="`upcoming-${item.name}`"
          class="recent-card"
          :class="{ 'is-today': item.date === todayKey }"
          @click="openFestival(item)"
        >
          <text class="recent-month">{{ monthOf(item.date) }}月</text>
          <text class="recent-day">{{ dayOf(item.date) }}</text>
          <text class="recent-name">{{ item.name }}</text>
          <text class="recent-state">{{ item.date === todayKey ? '今天' : daysUntil(item.date) + '天后' }}</text>
        </view>
      </view>
    </view>

    <MiniNotice tone="info" class="source-notice">
      法定假期按国务院办公厅公布的 2026 年安排整理，调休日期请以临近节日的官方通知为准。
    </MiniNotice>

    <view class="festival-section-head">
      <view>
        <text class="archive-kicker">2026</text>
        <text class="section-title">节日时间线</text>
      </view>
      <text class="section-count">{{ festivals.length }} 个</text>
    </view>

    <view class="festival-list">
      <view
        v-for="item in festivals"
        :key="item.name"
        class="festival-card"
        :class="festivalCardClass(item)"
        @click="openFestival(item)"
      >
        <view class="festival-date-block">
          <text class="festival-month">{{ monthOf(item.date) }}月</text>
          <text class="festival-day">{{ dayOf(item.date) }}</text>
        </view>
        <view class="festival-copy">
          <view class="festival-title-row">
            <text class="festival-name">{{ item.name }}</text>
            <text v-if="item.date === todayKey" class="festival-status">今天</text>
            <text v-else-if="isNextFestival(item)" class="festival-status upcoming">即将到来</text>
            <text v-else-if="isPastFestival(item)" class="festival-status past">已过</text>
            <text class="festival-tag" :class="{ statutory: item.category === '法定节假日' }">
              {{ item.category === '法定节假日' ? '法定' : '传统' }}
            </text>
          </view>
          <text class="festival-lunar">{{ item.lunarDate }}</text>
          <text class="festival-description">{{ item.description }}</text>
          <text v-if="item.holidayRange" class="festival-holiday">{{ item.holidayRange }}</text>
          <text v-if="item.workdays" class="festival-workday">{{ item.workdays }}</text>
        </view>
        <text class="festival-arrow">›</text>
      </view>
    </view>

    <text class="festival-footnote">传统节日日期按 2026 年农历与节气换算展示。</text>

  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, nextTick, ref } from 'vue'

import { ensureToolEnabled } from '@/api/toolConfigs'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import {
  dateKey,
  getFestivalItems,
  type FestivalItem
} from '@/features/festivals/festivalData'
import {
  festivalCopySceneOptions,
  generateFestivalCopy,
  type FestivalCopyScene
} from '@/features/festivals/festivalCopy'

const currentYear = 2026
const festivals = getFestivalItems(currentYear)
const todayKey = dateKey(new Date())
const todayFestival = festivals.find((item) => item.date === todayKey) || null
const nextFestival = computed(() => festivals.find((item) => item.date > todayKey) || null)
const featuredFestival = computed(() => todayFestival || nextFestival.value)
const upcomingFestivals = computed(() => festivals.filter((item) => item.date >= todayKey).slice(0, 3))
const featuredCountdown = computed(() => {
  if (!featuredFestival.value) return ''
  const days = daysUntil(featuredFestival.value.date)
  return days === 0 ? '今天' : `${days} 天后`
})
const selectedFestival = ref<FestivalItem | null>(null)
const showCopyComposer = ref(false)
const selectedScene = ref<FestivalCopyScene>('personal')
const copyVariantIndex = ref(0)
const festivalCopyText = computed(() =>
  selectedFestival.value
    ? generateFestivalCopy(selectedFestival.value, selectedScene.value, copyVariantIndex.value)
    : ''
)

onLoad(() => {
  void ensureToolEnabled('TRADITIONAL_FESTIVALS')
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

function isNextFestival(item: FestivalItem) {
  return !todayFestival && nextFestival.value?.date === item.date
}

function isPastFestival(item: FestivalItem) {
  return item.date < todayKey
}

function festivalCardClass(item: FestivalItem) {
  return {
    'is-today': item.date === todayKey,
    'is-upcoming': isNextFestival(item)
  }
}

function openFestival(item: FestivalItem) {
  selectedFestival.value = item
  showCopyComposer.value = false
  selectedScene.value = 'personal'
  copyVariantIndex.value = 0
  nextTick(() => {
    uni.pageScrollTo({ selector: '#festival-detail', duration: 180 })
  })
}

function nextFestivalCopy() {
  let nextIndex = copyVariantIndex.value
  while (nextIndex === copyVariantIndex.value) {
    nextIndex = Math.floor(Math.random() * 10)
  }
  copyVariantIndex.value = nextIndex
}

function closeFestival() {
  selectedFestival.value = null
  showCopyComposer.value = false
}

function copyFestivalCopy() {
  uni.setClipboardData({
    data: festivalCopyText.value,
    success: () => {
      uni.showToast({ title: '文案已复制', icon: 'none' })
    }
  })
}
</script>

<style scoped>
.festivals-page {
  padding-top: 28rpx;
}

.source-notice {
  margin-top: 24rpx;
}

.next-festival {
  margin-top: 18rpx;
  border-top: 4rpx solid var(--archive-cinnabar);
  padding: 24rpx 24rpx 18rpx;
  background: rgba(255, 252, 244, 0.78);
}

.next-festival.is-today {
  border-color: var(--archive-cinnabar);
  background: rgba(168, 59, 45, 0.07);
}

.next-heading,
.next-label,
.next-countdown,
.next-name,
.next-date,
.next-desc,
.next-footer {
  display: block;
}

.next-heading,
.next-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.next-label {
  color: var(--archive-cinnabar);
  font-size: 23rpx;
  font-weight: 750;
}

.next-countdown {
  color: var(--archive-blue);
  font-size: 21rpx;
  font-weight: 700;
}

.next-main {
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin-top: 20rpx;
}

.next-date-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 112rpx;
  height: 112rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-cinnabar);
  background: rgba(168, 59, 45, 0.055);
}

.next-month,
.next-day {
  display: block;
}

.next-month {
  color: var(--archive-cinnabar);
  font-size: 20rpx;
}

.next-day {
  margin-top: 2rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 48rpx;
  font-weight: 800;
  line-height: 1;
}

.next-copy {
  min-width: 0;
  flex: 1;
}

.next-name {
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 38rpx;
  font-weight: 800;
}

.next-date {
  color: var(--archive-blue);
  font-size: 25rpx;
  font-weight: 700;
}

.next-desc {
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.5;
}

.next-footer {
  margin-top: 20rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 15rpx;
  color: var(--archive-blue);
  font-size: 20rpx;
}

.festival-section-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-top: 32rpx;
}

.recent-section {
  margin-top: 30rpx;
}

.recent-section-head {
  margin-top: 0;
}

.recent-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12rpx;
  margin-top: 14rpx;
}

.recent-card {
  min-height: 156rpx;
  border: 1rpx solid var(--archive-line-strong);
  background: rgba(255, 252, 244, 0.66);
  padding: 14rpx 12rpx 12rpx;
  box-sizing: border-box;
}

.recent-card.is-today {
  border-color: rgba(168, 59, 45, 0.52);
  background: rgba(168, 59, 45, 0.07);
}

.recent-month,
.recent-day,
.recent-name,
.recent-state {
  display: block;
}

.recent-month {
  color: var(--archive-cinnabar);
  font-size: 18rpx;
}

.recent-day {
  margin-top: 2rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 35rpx;
  font-weight: 800;
  line-height: 1;
}

.recent-name {
  margin-top: 12rpx;
  color: var(--archive-ink);
  font-size: 23rpx;
  font-weight: 750;
}

.recent-state {
  margin-top: 5rpx;
  color: var(--archive-blue);
  font-size: 18rpx;
}

.section-title {
  display: block;
  margin-top: 5rpx;
  color: var(--archive-ink);
  font-size: 29rpx;
  font-weight: 750;
}

.section-count {
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.festival-list {
  position: relative;
  margin-top: 14rpx;
  padding-left: 28rpx;
}

.festival-list::before {
  position: absolute;
  top: 14rpx;
  bottom: 14rpx;
  left: 6rpx;
  width: 1rpx;
  background: var(--archive-line-strong);
  content: '';
}

.festival-card {
  position: relative;
  display: flex;
  gap: 18rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.festival-card::before {
  position: absolute;
  top: 48rpx;
  left: -28rpx;
  width: 13rpx;
  height: 13rpx;
  border: 2rpx solid var(--archive-line-strong);
  border-radius: 50%;
  background: var(--archive-paper-light);
  content: '';
}

.festival-date-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 76rpx;
  height: 76rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-line-strong);
  background: rgba(255, 252, 244, 0.72);
}

.festival-month,
.festival-day {
  display: block;
}

.festival-month {
  color: var(--archive-cinnabar);
  font-size: 18rpx;
}

.festival-day {
  margin-top: 2rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 800;
}

.festival-copy {
  min-width: 0;
  flex: 1;
}

.festival-title-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
}

.festival-status {
  border: 1rpx solid rgba(168, 59, 45, 0.38);
  color: var(--archive-cinnabar);
  padding: 2rpx 8rpx;
  font-size: 16rpx;
}

.festival-status.upcoming {
  border-color: rgba(35, 73, 98, 0.32);
  color: var(--archive-blue);
}

.festival-status.past {
  border-color: var(--archive-line-strong);
  color: var(--archive-ink-soft);
}

.festival-name {
  color: var(--archive-ink);
  font-size: 27rpx;
  font-weight: 750;
}

.festival-tag {
  border: 1rpx solid var(--archive-line-strong);
  color: var(--archive-ink-soft);
  padding: 2rpx 8rpx;
  font-size: 16rpx;
}

.festival-tag.statutory {
  border-color: rgba(168, 59, 45, 0.35);
  color: var(--archive-cinnabar);
}

.festival-lunar,
.festival-description,
.festival-holiday,
.festival-workday {
  display: block;
  margin-top: 5rpx;
  font-size: 20rpx;
  line-height: 1.4;
}

.festival-lunar {
  color: var(--archive-blue);
}

.festival-description {
  color: var(--archive-ink-soft);
}

.festival-holiday,
.festival-workday {
  color: var(--archive-cinnabar);
}

.festival-workday {
  color: var(--archive-ink-soft);
}

.festival-card.is-today {
  border-bottom-color: rgba(168, 59, 45, 0.45);
  background: rgba(168, 59, 45, 0.045);
}

.festival-card.is-today::before {
  border-color: var(--archive-cinnabar);
  background: var(--archive-cinnabar);
}

.festival-card.is-today .festival-name {
  color: var(--archive-cinnabar);
}

.festival-card.is-upcoming {
  border-bottom-color: rgba(35, 73, 98, 0.35);
  background: rgba(35, 73, 98, 0.035);
}

.festival-card.is-upcoming .festival-name {
  color: var(--archive-blue);
}

.festival-arrow {
  align-self: center;
  color: var(--archive-cinnabar);
  font-size: 32rpx;
  line-height: 1;
}

.festival-footnote {
  display: block;
  margin-top: 18rpx;
  border-top: 1rpx solid var(--archive-line);
  padding: 18rpx 0 30rpx;
  color: var(--archive-ink-soft);
  font-size: 19rpx;
  line-height: 1.5;
}

.festival-detail-panel {
  margin-top: 18rpx;
  border-top: 4rpx solid var(--archive-cinnabar);
  background: rgba(255, 252, 244, 0.82);
  padding: 24rpx;
}

.festival-detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding-bottom: 18rpx;
}

.festival-detail-close {
  color: var(--archive-blue);
  font-size: 21rpx;
  font-weight: 700;
}

.festival-detail-actions {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-top: 22rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 18rpx;
}

.festival-detail-link {
  margin-left: auto;
  color: var(--archive-blue);
  font-size: 21rpx;
}

.festival-modal-kicker,
.festival-modal-title,
.festival-modal-description,
.festival-modal-line {
  display: block;
}

.festival-modal-kicker {
  color: var(--archive-cinnabar);
  font-size: 18rpx;
  letter-spacing: 3rpx;
}

.festival-modal-title {
  margin-top: 8rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 38rpx;
  font-weight: 800;
}

.festival-modal-date {
  display: flex;
  gap: 16rpx;
  margin-top: 22rpx;
  color: var(--archive-blue);
  font-size: 23rpx;
}

.festival-modal-description {
  margin-top: 20rpx;
  color: var(--archive-ink);
  font-size: 26rpx;
  line-height: 1.65;
}

.festival-modal-detail {
  margin-top: 18rpx;
  border-top: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 16rpx 0;
}

.festival-detail-label,
.festival-detail-text {
  display: block;
}

.festival-detail-label {
  color: var(--archive-blue);
  font-size: 20rpx;
  font-weight: 700;
}

.festival-detail-label:not(:first-child) {
  margin-top: 14rpx;
}

.festival-detail-text {
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.festival-modal-line {
  margin-top: 10rpx;
  color: var(--archive-cinnabar);
  font-size: 21rpx;
  line-height: 1.45;
}

.festival-copy-composer {
  margin-top: 24rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 20rpx;
}

.festival-copy-title,
.festival-copy-hint,
.festival-copy-label {
  display: block;
}

.festival-copy-title {
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 750;
}

.festival-copy-hint {
  margin-top: 6rpx;
  color: var(--archive-ink-soft);
  font-size: 19rpx;
}

.festival-copy-label {
  margin-top: 18rpx;
  color: var(--archive-blue);
  font-size: 21rpx;
  font-weight: 700;
}

.festival-copy-options {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 8rpx;
}

.festival-copy-option {
  border: 1rpx solid var(--archive-line-strong);
  padding: 7rpx 16rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.festival-copy-option.active {
  border-color: var(--archive-blue);
  background: rgba(35, 73, 98, 0.08);
  color: var(--archive-blue);
  font-weight: 700;
}

.festival-copy-preview {
  margin: 20rpx 0 16rpx;
  border: 1rpx solid var(--archive-line-strong);
  background: rgba(255, 252, 244, 0.7);
  padding: 18rpx;
  color: var(--archive-ink);
  font-size: 24rpx;
  line-height: 1.65;
}

.festival-copy-actions {
  display: flex;
  align-items: center;
  gap: 10rpx;
}

</style>
