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

      <view class="home-spine archive-book-spine" aria-hidden="true">
        <view class="spine-title">
          <text>家</text>
          <text>谱</text>
          <text>册</text>
        </view>
        <view class="spine-stitches">
          <view v-for="item in 5" :key="item" class="spine-stitch" />
        </view>
      </view>

    </view>

    <view v-if="featuredShowcase" class="home-showcase-panel" @click="openHistoricalShowcase(featuredShowcase.id)">
      <view class="home-showcase-head">
        <view>
          <text class="archive-kicker">Imperial Lineage</text>
          <text class="home-showcase-title">帝王世家</text>
        </view>
        <text class="home-tools-more" @click.stop="openHistoricalShowcase()">查看全部 ›</text>
      </view>
      <view class="home-showcase-card">
        <image v-if="featuredShowcase.coverImageUrl" class="home-showcase-cover" :src="featuredShowcase.coverImageUrl" mode="aspectFill" />
        <view class="home-showcase-seal">谱</view>
        <view class="home-showcase-copy">
          <text class="home-showcase-name">{{ featuredShowcase.displayName }}</text>
          <text class="home-showcase-meta">{{ featuredShowcase.regionText || '历史示例' }} · {{ featuredShowcase.memberCount }} 位人物</text>
          <text class="home-showcase-desc">{{ featuredShowcase.summary }}</text>
        </view>
        <text class="archive-arrow">›</text>
      </view>
      <text class="home-showcase-note">历史示例 · 仅供浏览</text>
    </view>

    <view class="home-tools-panel">
      <view class="home-tools-head">
        <view>
          <text class="home-tools-title">常用工具</text>
        </view>
        <text class="home-tools-more" @click="go('/pages/tools/index')">查看全部 ›</text>
      </view>
      <view class="home-tools-grid">
        <view
          v-for="tool in visibleToolItems"
          :key="tool.key"
          class="home-tool-card"
          :class="{ 'is-tool-highlighted': isToolHighlighted(tool.key), 'is-tool-disabled': !isToolEnabled(tool.key) }"
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

    <view class="home-journey" @click="openHomePrimaryEntry">
      <view class="home-journey-head">
        <view>
          <text class="home-journey-title">从一页家谱开始</text>
          <text class="home-journey-desc">建立关系，慢慢补全属于家人的记忆</text>
        </view>
        <text class="archive-arrow">›</text>
      </view>
      <view class="home-journey-steps">
        <view
          v-for="(item, index) in journeySteps"
          :key="item.key"
          class="home-journey-step"
          @click.stop="openJourneyStep(item.key)"
        >
          <text class="home-journey-index">0{{ index + 1 }}</text>
          <text>{{ item.label }}</text>
        </view>
      </view>
    </view>

    <view class="home-page-note" aria-hidden="true">
      <view class="home-page-note-line" />
      <text>每一次补充，都是留给家人的一页。</text>
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

import { listMyFamilies } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import { listPublishedShowcaseFamilies } from '@/api/showcaseFamilies'
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
import { buildHomeSharePayload, loadShareConfigs } from '@/features/share/wechatShare'
import { useSessionStore } from '@/stores/session'
import type {
  FamilyMember,
  FamilySummary,
  ShowcaseFamilyItem,
} from '@/types/api'

interface ArchiveStat {
  label: string
  value: string
}

const session = useSessionStore()
const myFamilies = ref<FamilySummary[]>([])
const archiveStats = ref<ArchiveStat[]>([
  { label: '家庭记录', value: '私密' },
  { label: '常用工具', value: '—' },
  { label: '资料访问', value: '授权' }
])
const hasOwnFamilies = ref(false)
const featuredShowcase = ref<ShowcaseFamilyItem | null>(null)
const visibleToolItems = computed(() => getVisibleToolDefinitions())

const shouldPromptCreateFamily = computed(() => !hasOwnFamilies.value && session.isLoggedIn)
const directoryTagText = computed(() =>
  hasOwnFamilies.value ? ['家', '树'] : ['建', '树']
)
const journeySteps = computed(() => [
  { key: 'START', label: hasOwnFamilies.value ? '查看家树' : '建立家谱' },
  { key: 'MEMBERS', label: '补充成员' },
  { key: 'INVITE', label: '邀请家人' }
])
const primaryEntryTitle = computed(() =>
  hasOwnFamilies.value
    ? myFamilies.value.length === 1
      ? '打开家庭树'
      : '进入我的家庭'
    : shouldPromptCreateFamily.value
      ? '创建我的家庭'
      : '登录后建立家谱'
)
const primaryEntryDesc = computed(() =>
  hasOwnFamilies.value
    ? myFamilies.value.length === 1
      ? '查看家庭成员与关系图'
      : '选择要查看的家庭树'
    : shouldPromptCreateFamily.value
      ? '填写姓氏，建立自己的家庭树'
      : '登录并完善资料后，开始记录家庭关系'
)

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
    { label: '家庭记录', value: '私密' },
    { label: '常用工具', value: String(visibleToolItems.value.length) },
    { label: '资料访问', value: '授权' }
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

function go(url: string) {
  uni.navigateTo({ url })
}

function openHistoricalShowcase(showcaseFamilyId?: number) {
  const query = showcaseFamilyId ? `?familyId=${encodeURIComponent(String(showcaseFamilyId))}` : ''
  go(`/pages/family/historical-showcase${query}`)
}

async function loadFeaturedShowcase() {
  try {
    const items = await listPublishedShowcaseFamilies()
    featuredShowcase.value = items.find((item) => item.rotationPinned) || items[0] || null
  } catch {
    featuredShowcase.value = null
  }
}

function toolTitle(key: string) {
  return getToolDisplayName(key)
}

function toolDescription(key: string) {
  if (!isToolEnabled(key)) return '暂未开放'
  if (!getToolDefinition(key)?.path) return '功能暂未接入'
  return getToolDescription(key)
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
  if (!hasOwnFamilies.value) {
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

function openJourneyStep(step: 'START' | 'MEMBERS' | 'INVITE') {
  if (step === 'START') {
    openHomePrimaryEntry()
    return
  }

  if (!hasOwnFamilies.value || myFamilies.value.length !== 1) {
    goMyFamilyTab()
    return
  }

  const familyId = encodeURIComponent(String(myFamilies.value[0].id))
  uni.navigateTo({
    url: `/pages/family/members?familyId=${familyId}`
  })
}

async function refreshHome() {
  await Promise.all([loadArchiveStats(), loadFeaturedShowcase()])
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
  padding-bottom: calc(58rpx + env(safe-area-inset-bottom));
}

.home-book {
  position: relative;
  display: flex;
  min-height: 570rpx;
  margin: 4rpx 0 0;
}

.home-book-main {
  position: relative;
  flex: 1;
  min-width: 0;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background: rgba(255, 252, 244, 0.2);
  padding: 36rpx 112rpx 28rpx 4rpx;
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
  flex-direction: row;
  gap: 12rpx;
  width: auto;
  margin-top: 30rpx;
}

.home-stat {
  flex: 1;
  border-left: 3rpx solid var(--archive-cinnabar);
  padding-left: 14rpx;
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
  font-size: 22rpx;
  line-height: 1.2;
}

.bamboo-ink {
  right: 94rpx;
  top: 96rpx;
  width: 202rpx;
  height: 332rpx;
  opacity: 0.44;
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
  right: 92rpx;
  bottom: 24rpx;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  border: 1rpx solid rgba(35, 73, 98, 0.22);
  min-height: 106rpx;
  border-radius: 10rpx;
  background: rgba(255, 252, 244, 0.92);
  padding: 22rpx 20rpx;
  box-shadow: 0 8rpx 18rpx rgba(77, 59, 35, 0.05);
}

.home-tools-panel {
  margin: 22rpx 30rpx 30rpx 0;
  border: 1rpx solid var(--archive-line);
  border-radius: 8rpx;
  background: rgba(255, 252, 244, 0.34);
  padding: 22rpx 18rpx 18rpx;
}

.home-showcase-panel {
  margin: 0 30rpx 22rpx 0;
  border: 1rpx solid rgba(200, 164, 93, 0.52);
  border-radius: 8rpx;
  background: rgba(255, 251, 239, 0.72);
  padding: 20rpx 18rpx 16rpx;
}

.home-showcase-head { display: flex; align-items: baseline; justify-content: space-between; gap: 16rpx; }
.home-showcase-title { display: block; margin-top: 4rpx; color: var(--archive-ink); font-size: 27rpx; font-weight: 750; }
.home-showcase-card { position: relative; display: flex; align-items: center; gap: 14rpx; min-height: 126rpx; margin-top: 16rpx; overflow: hidden; border: 1rpx solid var(--archive-line); border-radius: 6rpx; background: rgba(255, 252, 244, 0.92); padding: 16rpx; }
.home-showcase-cover { width: 112rpx; height: 92rpx; flex: 0 0 auto; border-radius: 4rpx; }
.home-showcase-seal { display: grid; width: 56rpx; height: 56rpx; flex: 0 0 auto; place-items: center; border: 2rpx solid var(--archive-cinnabar); color: var(--archive-cinnabar); font-family: 'Songti SC', 'STSong', serif; font-size: 28rpx; transform: rotate(-7deg); }
.home-showcase-copy { display: grid; min-width: 0; flex: 1; gap: 4rpx; }
.home-showcase-name, .home-showcase-meta, .home-showcase-desc { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.home-showcase-name { color: var(--archive-ink); font-size: 26rpx; font-weight: 700; }
.home-showcase-meta { color: var(--archive-blue); font-size: 20rpx; }
.home-showcase-desc { color: var(--archive-ink-soft); font-size: 21rpx; }
.home-showcase-note { display: block; margin-top: 12rpx; color: var(--archive-ink-soft); font-size: 20rpx; }

.home-page-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  margin: 0 30rpx 0 0;
  color: rgba(101, 112, 128, 0.78);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 20rpx;
  letter-spacing: 1rpx;
  line-height: 1.4;
}

.home-journey {
  margin: 0 30rpx 22rpx 0;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background: rgba(255, 252, 244, 0.3);
  padding: 20rpx 18rpx 18rpx;
}

.home-journey-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
}

.home-journey-title,
.home-journey-desc,
.home-journey-index,
.home-journey-step > text:last-child {
  display: block;
}

.home-journey-title {
  color: var(--archive-ink);
  font-size: 26rpx;
  font-weight: 700;
}

.home-journey-desc {
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.4;
}

.home-journey-steps {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10rpx;
  margin-top: 16rpx;
}

.home-journey-step {
  position: relative;
  min-width: 0;
  border-left: 2rpx solid rgba(168, 59, 45, 0.52);
  padding-left: 10rpx;
}

.home-journey-step::after {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 0;
  height: 0;
  border-top: 4rpx solid transparent;
  border-bottom: 4rpx solid transparent;
  border-left: 5rpx solid rgba(101, 112, 128, 0.58);
  content: '';
}

.home-journey-index {
  color: var(--archive-cinnabar);
  font-size: 17rpx;
  font-weight: 700;
  letter-spacing: 1rpx;
}

.home-journey-step > text:last-child {
  margin-top: 3rpx;
  color: var(--archive-ink);
  font-size: 22rpx;
  line-height: 1.3;
  white-space: nowrap;
}

.home-page-note-line {
  width: 46rpx;
  height: 1rpx;
  background: var(--archive-line-strong);
}

.home-tools-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16rpx;
}

.home-tools-title,
.home-tools-more {
  display: block;
}

.home-tools-title {
  color: var(--archive-ink);
  font-size: 27rpx;
  font-weight: 750;
}

.home-tools-more {
  display: inline-flex;
  align-items: center;
  min-height: 64rpx;
  color: var(--archive-blue);
  font-size: 24rpx;
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
  min-height: 128rpx;
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
}

.home-tool-title {
  color: var(--archive-ink);
  font-size: 24rpx;
  font-weight: 700;
}

.home-tool-desc {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.35;
  white-space: normal;
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
  right: 70rpx;
  top: 218rpx;
  z-index: 4;
  width: 52rpx;
  min-height: 158rpx;
  flex-direction: column;
  gap: 12rpx;
  box-shadow: 5rpx 8rpx 16rpx rgba(77, 59, 35, 0.1);
  transform: rotate(2deg);
}

.home-spine {
  width: 72rpx;
  min-height: 594rpx;
  flex-shrink: 0;
}

.spine-title {
  position: absolute;
  top: 76rpx;
  left: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10rpx;
  color: rgba(255, 255, 255, 0.92);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 700;
  line-height: 1.2;
  transform: translateX(-50%);
}

.spine-stitches {
  position: absolute;
  right: 26rpx;
  bottom: 60rpx;
  display: flex;
  flex-direction: column;
  gap: 28rpx;
}

.spine-stitch {
  width: 15rpx;
  height: 15rpx;
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
