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

    <view v-if="featuredRead" class="home-reading archive-panel">
      <view class="home-section-head">
        <text class="archive-kicker">Reading</text>
        <text class="home-section-title">阅读精选</text>
      </view>
      <view class="home-note" @click="openReadingArticle(featuredRead)">
        <text class="home-note-label">{{ categoryTitle(featuredRead.categoryKey) }}</text>
        <text class="home-note-title">{{ featuredRead.title }}</text>
      </view>
    </view>

    <view v-if="!hasOwnFamilies && showcaseFamilies.length > 0" class="home-showcase archive-list">
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

    <view class="home-privacy">
      <text>家庭资料仅在授权范围内可见，公开展示需经过审核。</text>
    </view>

    <!-- #ifdef MP-WEIXIN -->
    <view class="home-share">
      <text class="home-share-label">分享小程序</text>
      <button class="wechat-share-button" open-type="share">分享给微信好友</button>
    </view>
    <!-- #endif -->
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShareAppMessage, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { listContentArticles, listContentCategories } from '@/api/content'
import { getFamilyDetail, listMyFamilies, listPublicFamilyShowcase } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import { promptPrivacyConsentIfNeeded } from '@/features/legal/privacyConsent'
import { openWechatOfficialArticle } from '@/features/content/wechatOfficialArticle'
import { buildHomeSharePayload } from '@/features/share/wechatShare'
import { useSessionStore } from '@/stores/session'
import type { ContentArticleSummary, ContentCategory, PublicFamilyShowcaseItem } from '@/types/api'

interface ArchiveStat {
  label: string
  value: string
}

const session = useSessionStore()
const categories = ref<ContentCategory[]>([])
const articles = ref<ContentArticleSummary[]>([])
const showcaseFamilies = ref<PublicFamilyShowcaseItem[]>([])
const showcaseTotal = ref(0)
const archiveStats = ref<ArchiveStat[]>([
  { label: '展示家庭', value: '—' },
  { label: '传世故事', value: '—' },
  { label: '阅读篇目', value: '—' }
])
const hasOwnFamilies = ref(false)

const featuredRead = computed(() => articles.value.find((item) => item.isFeatured) || articles.value[0] || null)
const shouldPromptCreateFamily = computed(() => !hasOwnFamilies.value && session.isLoggedIn)
const primaryEntryTitle = computed(() =>
  hasOwnFamilies.value ? '翻开我的家谱' : shouldPromptCreateFamily.value ? '创建我的家庭' : '浏览展示家庭'
)
const primaryEntryDesc = computed(() =>
  hasOwnFamilies.value
    ? '查看成员、关系和家谱册页'
    : shouldPromptCreateFamily.value
      ? '填写姓氏，建立自己的家谱册页'
      : '无需登录，先看公开展示册页'
)
const directoryTagText = computed(() =>
  hasOwnFamilies.value ? ['家', '谱'] : shouldPromptCreateFamily.value ? ['建', '谱'] : ['展', '示']
)

function categoryTitle(key: string) {
  return categories.value.find((item) => item.key === key)?.name || '内容'
}

function articleUrl(id: number | string) {
  return `/pages/content/detail?id=${encodeURIComponent(id)}`
}

function familyRoleText(role?: string) {
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    default:
      return '成员'
  }
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
      const myFamilies = await listMyFamilies()
      if (myFamilies.length > 0) {
        hasOwnFamilies.value = true
        const primary = myFamilies[0]
        const [members, detail] = await Promise.all([
          listFamilyMembers(primary.id),
          getFamilyDetail(primary.id)
        ])
        archiveStats.value = [
          { label: '成员人数', value: String(members.length) },
          { label: '家庭角色', value: familyRoleText(detail.role) },
          { label: '谱系版本', value: `v${detail.graphVersion}` }
        ]
        return
      }
    } catch {
      // Fall through to browse stats.
    }
  }
  hasOwnFamilies.value = false
  archiveStats.value = [
    { label: '展示家庭', value: showcaseTotal.value > 0 ? String(showcaseTotal.value) : '—' },
    { label: '传世故事', value: articles.value.length > 0 ? String(articles.value.length) : '—' },
    { label: '阅读篇目', value: categories.value.length > 0 ? String(categories.value.length) : '—' }
  ]
}

function openShowcaseFamily(familyId: number | string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(String(familyId))}`
  })
}

function go(url: string) {
  uni.navigateTo({ url })
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
  goMyFamilyTab()
}

async function refreshHome() {
  await loadReading()
  await loadShowcasePreview()
  await loadArchiveStats()
}

onLoad(refreshHome)

onShareAppMessage(() => buildHomeSharePayload())

onShow(() => {
  promptPrivacyConsentIfNeeded()
  refreshHome()
})
</script>

<style scoped>
.home-page {
  padding-right: 0;
}

.home-book {
  position: relative;
  display: flex;
  min-height: 720rpx;
  margin: 4rpx 0 34rpx;
}

.home-book-main {
  position: relative;
  flex: 1;
  min-width: 0;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 48rpx 148rpx 36rpx 4rpx;
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
  margin-top: 14rpx;
  font-size: 56rpx;
}

.home-title-seal {
  flex-shrink: 0;
  margin-top: 8rpx;
}

.home-stats {
  display: flex;
  flex-direction: column;
  gap: 22rpx;
  width: 200rpx;
  margin-top: 64rpx;
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
  bottom: 32rpx;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  border-top: 1rpx solid var(--archive-line);
  padding: 24rpx 8rpx 0 0;
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
  min-height: 720rpx;
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
  gap: 18rpx;
  margin: 24rpx 30rpx 0 0;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 18rpx;
}

.home-share-label {
  flex-shrink: 0;
  color: var(--archive-ink);
  font-size: 25rpx;
  font-weight: 650;
  white-space: nowrap;
}

.wechat-share-button {
  flex: 1;
  min-width: 0;
  margin: 0;
  border: 1rpx solid var(--archive-cinnabar);
  border-radius: 0;
  background: transparent;
  color: var(--archive-cinnabar);
  font-size: 23rpx;
  line-height: 2.1;
}

.wechat-share-button::after {
  border: 0;
}
</style>
