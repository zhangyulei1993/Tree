<template>
  <view class="archive-page home-page">
    <view class="home-book">
      <view class="home-book-main">
        <view class="home-title-row">
          <view class="home-title-copy">
            <text class="archive-kicker">Tree</text>
            <text class="archive-title">我的家庭</text>
            <text class="archive-subtitle">家之有谱，世代相传</text>
          </view>
          <view class="home-title-seal archive-seal">谱</view>
        </view>

        <view class="home-stats">
          <view v-for="item in archiveStats" :key="item.label" class="home-stat">
            <text class="home-stat-value">{{ item.value }}</text>
            <text class="home-stat-label">{{ item.label }}</text>
          </view>
        </view>

        <view class="bamboo-ink" aria-hidden="true">
          <view class="bamboo-stem bamboo-stem-a" />
          <view class="bamboo-stem bamboo-stem-b" />
          <view class="bamboo-stem bamboo-stem-c" />
          <view class="bamboo-leaf bamboo-leaf-a" />
          <view class="bamboo-leaf bamboo-leaf-b" />
          <view class="bamboo-leaf bamboo-leaf-c" />
          <view class="bamboo-leaf bamboo-leaf-d" />
          <view class="bamboo-leaf bamboo-leaf-e" />
          <view class="bamboo-leaf bamboo-leaf-f" />
          <view class="bamboo-mist bamboo-mist-a" />
          <view class="bamboo-mist bamboo-mist-b" />
        </view>

      <view class="home-primary-entry" @click="openHomePrimaryEntry">
          <view>
            <text class="entry-title">{{ primaryEntryTitle }}</text>
            <text class="entry-desc">{{ primaryEntryDesc }}</text>
          </view>
          <text class="archive-arrow">›</text>
        </view>
      </view>

      <view class="home-directory-tag archive-paper-tag" @click="goMyFamilyTab">
        <text>{{ directoryTagText[0] }}</text>
        <text>{{ directoryTagText[1] }}</text>
      </view>

      <view class="home-spine archive-book-spine">
        <view class="spine-title">
          <text>家</text>
          <text>谱</text>
          <text>册</text>
        </view>
        <view class="spine-stitches">
          <view v-for="item in 6" :key="item" class="spine-stitch" />
        </view>
      </view>

    </view>

    <view class="home-tools-panel">
      <view class="home-tools-head">
        <view>
          <text class="home-tools-kicker">TOOLS</text>
          <text class="home-tools-title">常用工具</text>
        </view>
        <text class="home-tools-more" @click="go('/pages/tools/index')">查看全部 ›</text>
      </view>
      <view class="home-tools-grid">
        <view
          v-for="tool in visibleToolItems"
          :key="tool.key"
          class="home-tool-card"
          :class="[toolCardClass(tool.key), { 'is-tool-highlighted': isToolHighlighted(tool.key), 'is-tool-disabled': !isToolEnabled(tool.key) }]"
          @click="openCommonTool(tool.key)"
        >
          <view v-if="tool.key === 'KINSHIP_QUERY'" class="home-tool-icon relation-tool-icon" aria-hidden="true">
            <view class="home-tool-dot dot-top" />
            <view class="home-tool-dot dot-left" />
            <view class="home-tool-dot dot-right" />
            <view class="home-tool-branch branch-left" />
            <view class="home-tool-branch branch-right" />
          </view>
          <view v-else class="home-tool-icon">{{ tool.icon }}</view>
          <view class="home-tool-copy">
            <text class="home-tool-title">{{ toolTitle(tool.key) }}</text>
            <text class="home-tool-desc">{{ toolDescription(tool.key) }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="!hasOwnFamilies && showcaseFamilies.length > 0" class="home-showcase archive-list ui-simplified-hidden">
      <view
        v-for="family in showcaseFamilies"
        :key="family.id"
        class="archive-row"
        @click="openShowcaseFamily(family.id)"
      >
        <text class="archive-surname-stamp">{{ family.familySurname.slice(0, 1) }}</text>
        <view class="archive-row-main">
          <text class="archive-row-title">{{ family.familyName }}</text>
          <text class="archive-row-desc">{{ family.regionText || family.nativePlace || '已公开展示' }}</text>
        </view>
        <text class="archive-arrow">›</text>
      </view>
      <view class="home-showcase-more" @click="goMyFamilyTab">
        <text>在「我的家庭」查看更多展示家庭</text>
        <text class="archive-arrow">›</text>
      </view>
    </view>

    <view class="home-privacy ui-simplified-hidden">
      <text>家庭资料仅在授权范围内可见，公开展示需经过审核。</text>
    </view>

    <!-- #ifdef MP-WEIXIN -->
    <view class="home-share ui-simplified-hidden">
      <view class="home-share-copy">
        <text class="home-share-label">把家庭记录分享给家人</text>
        <text class="home-share-desc">邀请家人一起补充和维护家庭关系</text>
      </view>
      <button class="wechat-share-button" open-type="share">分享</button>
    </view>
    <!-- #endif -->
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShareAppMessage, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { listContentArticles, listContentCategories } from '@/api/content'
import { listMyFamilies, listPublicFamilyShowcase } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import {
  getVisibleToolDefinitions,
  getToolDefinition,
  getToolDescription,
  getToolDisplayName,
  isToolEnabled,
  isToolHighlighted,
  loadToolConfigs
} from '@/api/toolConfigs'
import { promptPrivacyConsentIfNeeded } from '@/features/legal/privacyConsent'
import { openWechatOfficialArticle } from '@/features/content/wechatOfficialArticle'
import { buildHomeSharePayload, loadShareConfigs } from '@/features/share/wechatShare'
import { dateKey, festivalDateLabel, getFestivalItems } from '@/features/festivals/festivalData'
import { useSessionStore } from '@/stores/session'
import type {
  ContentArticleSummary,
  ContentCategory,
  FamilyMember,
  FamilySummary,
  PublicFamilyShowcaseItem
} from '@/types/api'

interface ArchiveStat {
  label: string
  value: string
}

const session = useSessionStore()
const categories = ref<ContentCategory[]>([])
const articles = ref<ContentArticleSummary[]>([])
const showcaseFamilies = ref<PublicFamilyShowcaseItem[]>([])
const showcaseTotal = ref(0)
const myFamilies = ref<FamilySummary[]>([])
const archiveStats = ref<ArchiveStat[]>([
  { label: '展示家庭', value: '—' },
  { label: '家庭故事', value: '—' },
  { label: '阅读篇目', value: '—' }
])
const hasOwnFamilies = ref(false)
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

const featuredRead = computed(() => articles.value.find((item) => item.isFeatured) || articles.value[0] || null)
const shouldPromptCreateFamily = computed(() => !hasOwnFamilies.value && session.isLoggedIn)
const primaryEntryTitle = computed(() =>
  hasOwnFamilies.value
    ? myFamilies.value.length === 1
      ? '打开家庭树'
      : '进入我的家庭'
    : shouldPromptCreateFamily.value
      ? '创建我的家庭'
      : '浏览展示家庭'
)
const primaryEntryDesc = computed(() =>
  hasOwnFamilies.value
    ? myFamilies.value.length === 1
      ? '查看家庭成员与关系图'
      : '选择要查看的家庭树'
    : shouldPromptCreateFamily.value
      ? '填写姓氏，建立自己的家庭树'
      : '无需登录，先看公开展示册页'
)
const directoryTagText = computed(() =>
  hasOwnFamilies.value ? ['家', '树'] : shouldPromptCreateFamily.value ? ['建', '树'] : ['展', '示']
)

function categoryTitle(key: string) {
  return categories.value.find((item) => item.key === key)?.name || '内容'
}

function articleUrl(id: number | string) {
  return `/pages/content/detail?id=${encodeURIComponent(id)}`
}

async function openReadingArticle(article: ContentArticleSummary) {
  if (article.contentType === 'WECHAT_OFFICIAL') {
    try {
      await openWechatOfficialArticle(article.externalUrl || '')
    } catch (err) {
      uni.showToast({ title: err instanceof Error ? err.message : '公众号文章暂时无法打开', icon: 'none' })
    }
    return
  }
  go(articleUrl(article.id))
}

async function loadReading() {
  try {
    const [categoryResult, articleResult] = await Promise.all([
      listContentCategories(),
      listContentArticles({ page: 1, pageSize: 12 })
    ])
    categories.value = categoryResult
    articles.value = articleResult.items
  } catch {
    categories.value = []
    articles.value = []
  }
}

async function loadShowcasePreview() {
  try {
    const result = await listPublicFamilyShowcase({ page: 1, pageSize: 2 })
    showcaseFamilies.value = result.items
    showcaseTotal.value = result.total
  } catch {
    showcaseFamilies.value = []
    showcaseTotal.value = 0
  }
}

async function loadArchiveStats() {
  session.restoreSession()
  if (session.isLoggedIn && session.isProfileComplete) {
    try {
      myFamilies.value = await listMyFamilies()
      if (myFamilies.value.length > 0) {
        hasOwnFamilies.value = true
        const primary = myFamilies.value[0]
        const members = await listFamilyMembers(primary.id)
        archiveStats.value = [
          { label: '成员人数', value: String(members.length) },
          { label: '最近更新', value: formatLatestMemberUpdate(members) },
          { label: '我的家庭', value: String(myFamilies.value.length) }
        ]
        return
      }
    } catch {
      // Fall through to browse stats.
    }
  }
  hasOwnFamilies.value = false
  myFamilies.value = []
  archiveStats.value = [
    { label: '展示家庭', value: showcaseTotal.value > 0 ? String(showcaseTotal.value) : '—' },
    { label: '家庭故事', value: articles.value.length > 0 ? String(articles.value.length) : '—' },
    { label: '阅读篇目', value: categories.value.length > 0 ? String(categories.value.length) : '—' }
  ]
}

function formatLatestMemberUpdate(members: FamilyMember[]) {
  const latestTime = members
    .map((member) => Date.parse(member.updatedAt || member.createdAt || ''))
    .filter((time) => Number.isFinite(time))
    .sort((a, b) => b - a)[0]
  if (!latestTime) return '暂无记录'
  return new Date(latestTime).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

function openShowcaseFamily(familyId: number | string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(String(familyId))}`
  })
}

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
  if (key === 'FAMILY_STORIES') {
    openFamilyStories()
    return
  }
  const definition = getToolDefinition(key)
  if (!definition?.path) {
    uni.showToast({ title: '该工具暂未接入', icon: 'none' })
    return
  }
  go(definition.path)
}

function openFamilyStories() {
  if (!session.isLoggedIn) {
    session.requireLogin('/pages/family/stories')
    return
  }
  if (myFamilies.value.length === 0) {
    uni.showToast({ title: '加入或创建家庭后查看', icon: 'none' })
    return
  }
  const familyId = myFamilies.value.length === 1
    ? `?familyId=${encodeURIComponent(String(myFamilies.value[0].id))}`
    : ''
  go(`/pages/family/stories${familyId}`)
}

function goMyFamilyTab() {
  uni.switchTab({ url: '/pages/family/my' })
}

function openCreateFamily() {
  if (!session.requireProfileComplete('/pages/family/create')) return
  uni.navigateTo({ url: '/pages/family/create' })
}

function openHomePrimaryEntry() {
  if (!hasOwnFamilies.value && session.isLoggedIn) {
    openCreateFamily()
    return
  }
  if (hasOwnFamilies.value && myFamilies.value.length === 1) {
    uni.navigateTo({
      url: `/pages/family/private-tree?familyId=${encodeURIComponent(String(myFamilies.value[0].id))}`
    })
    return
  }
  goMyFamilyTab()
}

async function refreshHome() {
  await loadReading()
  await loadShowcasePreview()
  await loadArchiveStats()
}

onLoad(() => {
  void loadToolConfigs()
  void refreshHome()
})
loadShareConfigs()

onShareAppMessage(() => buildHomeSharePayload())

onShow(() => {
  promptPrivacyConsentIfNeeded()
  void loadToolConfigs({ force: true })
  void loadShareConfigs({ force: true })
  refreshHome()
})
</script>

<style scoped>
.ui-simplified-hidden {
  display: none !important;
}

.home-page {
  padding-right: 0;
}

.home-book {
  position: relative;
  display: flex;
  min-height: 638rpx;
  margin: 4rpx 0 0;
}

.home-book-main {
  position: relative;
  flex: 1;
  min-width: 0;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 42rpx 148rpx 30rpx 4rpx;
  overflow: hidden;
}

.home-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.home-title-copy {
  flex: 1;
  min-width: 0;
}

.home-title-copy .archive-title {
  margin-top: 10rpx;
  font-size: 52rpx;
}

.home-title-seal {
  flex-shrink: 0;
  margin-top: 8rpx;
}

.home-stats {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
  width: 200rpx;
  margin-top: 48rpx;
}

.home-stat {
  border-left: 3rpx solid var(--archive-cinnabar);
  padding-left: 18rpx;
}

.home-stat-value,
.home-stat-label {
  display: block;
}

.home-stat-value {
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 36rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  line-height: 1.15;
}

.home-stat-label {
  margin-top: 7rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.2;
}

.bamboo-ink {
  position: absolute;
  right: 138rpx;
  top: 130rpx;
  width: 220rpx;
  height: 380rpx;
  opacity: 0.82;
}

.bamboo-stem {
  position: absolute;
  bottom: 0;
  width: 4rpx;
  border-radius: 999rpx;
  background: var(--archive-bamboo);
  transform-origin: bottom;
}

.bamboo-stem-a {
  left: 58rpx;
  height: 340rpx;
  transform: rotate(8deg);
}

.bamboo-stem-b {
  left: 98rpx;
  height: 300rpx;
  transform: rotate(-5deg);
}

.bamboo-stem-c {
  left: 138rpx;
  height: 260rpx;
  transform: rotate(4deg);
}

.bamboo-leaf {
  position: absolute;
  width: 88rpx;
  height: 22rpx;
  border-radius: 100% 0 100% 0;
  background: var(--archive-bamboo);
  transform-origin: left center;
}

.bamboo-leaf-a {
  top: 36rpx;
  left: 68rpx;
  transform: rotate(-32deg);
}

.bamboo-leaf-b {
  top: 88rpx;
  left: 22rpx;
  transform: rotate(208deg);
}

.bamboo-leaf-c {
  top: 128rpx;
  left: 88rpx;
  transform: rotate(-16deg);
}

.bamboo-leaf-d {
  top: 196rpx;
  left: 28rpx;
  transform: rotate(192deg);
}

.bamboo-leaf-e {
  top: 248rpx;
  left: 112rpx;
  transform: rotate(-24deg);
}

.bamboo-leaf-f {
  top: 302rpx;
  left: 52rpx;
  transform: rotate(186deg);
}

.bamboo-mist {
  position: absolute;
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: rgba(37, 76, 59, 0.06);
}

.bamboo-mist-a {
  top: 60rpx;
  left: 10rpx;
}

.bamboo-mist-b {
  top: 210rpx;
  left: 70rpx;
}

.home-primary-entry {
  position: absolute;
  right: 128rpx;
  bottom: 24rpx;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  border: 1rpx solid rgba(35, 73, 98, 0.22);
  border-radius: 8rpx;
  background: rgba(255, 252, 244, 0.64);
  padding: 20rpx 18rpx;
  box-shadow: 0 8rpx 18rpx rgba(77, 59, 35, 0.05);
}

.home-tools-panel {
  margin: 22rpx 30rpx 30rpx 0;
  border: 1rpx solid var(--archive-line);
  border-radius: 8rpx;
  background: rgba(255, 252, 244, 0.34);
  padding: 22rpx 18rpx 18rpx;
}

.home-tools-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16rpx;
}

.home-tools-kicker,
.home-tools-title,
.home-tools-more {
  display: block;
}

.home-tools-kicker {
  color: var(--archive-cinnabar);
  font-size: 18rpx;
  letter-spacing: 3rpx;
}

.home-tools-title {
  margin-top: 5rpx;
  color: var(--archive-ink);
  font-size: 27rpx;
  font-weight: 750;
}

.home-tools-more {
  color: var(--archive-blue);
  font-size: 21rpx;
}

.home-tools-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14rpx;
  margin-top: 16rpx;
}

.home-tool-card {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-width: 0;
  min-height: 112rpx;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 6rpx;
  background: rgba(255, 252, 244, 0.76);
  padding: 16rpx 14rpx;
  box-sizing: border-box;
}

.home-tool-card.is-tool-highlighted {
  border-color: rgba(168, 59, 45, 0.72);
  box-shadow: 0 4rpx 0 rgba(168, 59, 45, 0.16);
}

.home-tool-card.is-tool-disabled {
  opacity: 0.62;
}

.home-tool-card.is-festival-today {
  border-color: rgba(168, 59, 45, 0.55);
  background: rgba(168, 59, 45, 0.08);
}

.home-tool-card.is-festival-today .home-tool-title,
.home-tool-card.is-festival-today .home-tool-desc {
  color: var(--archive-cinnabar);
}

.home-tool-card.is-festival-upcoming {
  border-color: rgba(35, 73, 98, 0.42);
  background: rgba(35, 73, 98, 0.055);
}

.home-tool-card.is-festival-upcoming .home-tool-title {
  color: var(--archive-blue);
}

.home-tool-copy {
  min-width: 0;
  flex: 1;
}

.home-tool-icon {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42rpx;
  height: 42rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-cinnabar);
  border-radius: 50%;
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 22rpx;
  font-weight: 700;
}

.relation-tool-icon {
  border-color: var(--archive-line-strong);
}

.home-tool-dot {
  position: absolute;
  width: 7rpx;
  height: 7rpx;
  border: 1rpx solid var(--archive-blue);
  border-radius: 50%;
  background: var(--archive-paper-light);
}

.dot-top {
  top: 6rpx;
  left: 16rpx;
}

.dot-left {
  bottom: 7rpx;
  left: 5rpx;
}

.dot-right {
  right: 5rpx;
  bottom: 7rpx;
}

.home-tool-branch {
  position: absolute;
  top: 19rpx;
  left: 19rpx;
  width: 12rpx;
  height: 1rpx;
  background: var(--archive-cinnabar);
  transform-origin: left center;
}

.branch-left {
  transform: rotate(145deg);
}

.branch-right {
  transform: rotate(35deg);
}

.home-tool-title,
.home-tool-desc {
  display: block;
  overflow: hidden;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-tool-title {
  color: var(--archive-ink);
  font-size: 24rpx;
  font-weight: 700;
}

.home-tool-desc {
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 19rpx;
}

.entry-title,
.entry-desc {
  display: block;
}

.entry-title {
  color: var(--archive-blue);
  font-size: 30rpx;
  font-weight: 750;
  line-height: 1.3;
}

.entry-desc {
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.4;
}

.home-directory-tag {
  position: absolute;
  right: 96rpx;
  top: 248rpx;
  z-index: 4;
  width: 58rpx;
  min-height: 178rpx;
  flex-direction: column;
  gap: 14rpx;
  box-shadow: 6rpx 10rpx 18rpx rgba(77, 59, 35, 0.12);
  transform: rotate(2deg);
}

.home-spine {
  width: 96rpx;
  min-height: 680rpx;
  flex-shrink: 0;
}

.spine-title {
  position: absolute;
  top: 88rpx;
  left: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  color: rgba(255, 255, 255, 0.92);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 700;
  line-height: 1.2;
  transform: translateX(-50%);
}

.spine-stitches {
  position: absolute;
  right: 36rpx;
  bottom: 88rpx;
  display: flex;
  flex-direction: column;
  gap: 32rpx;
}

.spine-stitch {
  width: 18rpx;
  height: 18rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.78);
  border-radius: 50%;
}

.home-reading {
  margin: 30rpx 30rpx 0 0;
  padding: 26rpx 0 24rpx;
}

.home-section-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 16rpx;
}

.home-section-title {
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 750;
}

.home-note {
  border-top: 1rpx solid var(--archive-line);
  padding-top: 18rpx;
}

.home-note-label,
.home-note-title {
  display: block;
}

.home-note-label {
  color: var(--archive-cinnabar);
  font-size: 21rpx;
  line-height: 1.4;
}

.home-note-title {
  margin-top: 6rpx;
  color: var(--archive-ink);
  font-size: 27rpx;
  font-weight: 650;
  line-height: 1.45;
}

.home-showcase {
  margin: 24rpx 30rpx 0 0;
}

.home-showcase-more {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1rpx solid var(--archive-line);
  color: var(--archive-blue);
  padding: 18rpx 0;
  font-size: 24rpx;
  font-weight: 650;
}

.home-privacy {
  margin: 26rpx 30rpx 0 0;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.7;
}

.home-share {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 22rpx;
  margin: 24rpx 30rpx 0 0;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 18rpx;
}

.home-share-copy {
  flex: 1;
  min-width: 0;
}

.home-share-label {
  display: block;
  color: var(--archive-ink);
  font-size: 25rpx;
  font-weight: 650;
  line-height: 1.4;
}

.home-share-desc {
  display: block;
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.45;
}

.wechat-share-button {
  flex-shrink: 0;
  width: 148rpx;
  margin: 0;
  border: 0;
  border-radius: 999rpx;
  background: var(--archive-blue);
  color: var(--archive-paper-light);
  font-size: 23rpx;
  font-weight: 650;
  line-height: 2.4;
}

.wechat-share-button::after {
  border: 0;
}
</style>
