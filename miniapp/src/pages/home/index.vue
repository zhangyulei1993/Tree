<template>
  <view class="archive-page home-page" :class="{ 'home-page--splashing': launchSplashVisible }">
    <view class="home-page-content">
    <view class="home-book">
      <view class="home-book-main">
        <view class="home-title-row">
          <view class="home-title-copy">
            <text class="archive-kicker">Tree</text>
            <text class="archive-title">家族关系成树</text>
            <text class="archive-subtitle">从成员、亲属关系到家族脉络，慢慢留下一本可延续的家谱。</text>
          </view>
          <view class="home-title-seal archive-seal">谱</view>
        </view>

        <view class="home-stats-caption">
          <text>{{ platformStatsCaption }}</text>
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

    <swiper
      class="home-discovery-swiper"
      circular
      :autoplay="true"
      :interval="5200"
      :duration="360"
      :indicator-dots="true"
      indicator-color="rgba(31, 58, 95, .18)"
      indicator-active-color="#a83b2d"
    >
      <swiper-item>
        <view class="home-showcase-panel home-quality-panel" @click="go('/pages/family/showcase')">
      <view class="home-showcase-head">
        <view>
            <text class="archive-kicker">优质家谱</text>
          <text class="home-showcase-title">优质家庭</text>
        </view>
        <text class="home-tools-more">去学习 ›</text>
      </view>
      <view class="home-showcase-card">
        <view class="home-showcase-seal home-quality-seal">优</view>
        <view class="home-showcase-copy">
          <text class="home-showcase-name">优质家庭案例</text>
          <text class="home-showcase-meta">平台审核 · 学习参考</text>
          <text class="home-showcase-desc">浏览真实家庭的记录方式与树形结构</text>
        </view>
        <text class="archive-arrow">›</text>
      </view>
      <text class="home-showcase-note">不开放陌生人申请加入</text>
        </view>
      </swiper-item>

      <swiper-item>
    <view class="home-showcase-panel home-imperial-panel" @click="featuredShowcase ? openHistoricalShowcase(featuredShowcase.id) : openHistoricalShowcase()">
      <view class="home-showcase-head">
        <view>
          <text class="archive-kicker">历史世系</text>
          <text class="home-showcase-title">帝王世家</text>
        </view>
        <text class="home-tools-more" @click.stop="openHistoricalShowcase()">查看全部 ›</text>
      </view>
      <view class="home-showcase-card">
        <image v-if="featuredShowcase?.coverImageUrl" class="home-showcase-cover" :src="featuredShowcase.coverImageUrl" mode="aspectFill" />
        <view v-else class="home-showcase-seal">谱</view>
        <view class="home-showcase-copy">
          <text class="home-showcase-name">{{ featuredShowcase?.displayName || '帝王世家整理中' }}</text>
          <text class="home-showcase-meta">{{ featuredShowcase ? `${featuredShowcase.regionText || '历史示例'} · ${featuredShowcase.memberCount} 位人物` : '历史示例 · 仅供浏览' }}</text>
          <text class="home-showcase-desc">{{ featuredShowcase?.summary || (showcaseLoadFailed ? '暂时无法读取展示内容，请稍后再试。' : '后台发布后，这里会展示置顶世系。') }}</text>
        </view>
        <text class="archive-arrow">›</text>
      </view>
      <text class="home-showcase-note">历史示例 · 仅供浏览</text>
    </view>
      </swiper-item>
    </swiper>

    <view v-if="shouldShowJourney" class="home-journey" @click="openJourneyPrimary">
      <view class="home-journey-head">
        <view>
          <text class="home-journey-kicker">{{ journeyKicker }}</text>
          <text class="home-journey-title">{{ journeyTitle }}</text>
          <text class="home-journey-desc">{{ journeyDesc }}</text>
        </view>
        <text class="archive-arrow">›</text>
      </view>
      <view class="home-journey-steps">
        <view
          v-for="(item, index) in journeySteps"
          :key="item.key"
          class="home-journey-step"
          hover-class="home-journey-step--pressed"
          hover-stay-time="80"
          :class="{ 'is-current': nextJourneyStepKey === item.key, 'is-pressed': pressedJourneyStepKey === item.key }"
          @touchstart="pressJourneyStep(item.key)"
          @touchend="clearPressedJourneyStep"
          @touchcancel="clearPressedJourneyStep"
          @mousedown="pressJourneyStep(item.key)"
          @mouseup="clearPressedJourneyStep"
          @mouseleave="clearPressedJourneyStep"
          @click.stop="openJourneyStep(item.key)"
        >
          <text class="home-journey-index">{{ journeyStepDone(item.key) ? '✓' : `0${index + 1}` }}</text>
          <view class="home-journey-step-copy">
            <text class="home-journey-label">{{ item.label }}</text>
            <text class="home-journey-step-desc">{{ item.description }}</text>
          </view>
          <text v-if="journeyStepDone(item.key)" class="home-journey-done">已了解</text>
          <text v-else-if="nextJourneyStepKey === item.key" class="home-journey-current">建议先做</text>
        </view>
      </view>
      <view class="home-journey-footer">
        <text class="home-journey-progress">{{ guideProgressText }}</text>
        <view class="home-journey-continue" hover-class="home-journey-reset--pressed" hover-stay-time="80" @click.stop="continueNewUserGuide">
          <text>继续引导</text>
        </view>
        <view class="home-journey-reset" hover-class="home-journey-reset--pressed" hover-stay-time="80" @click.stop="dismissNewUserGuideCard">
          <text>稍后再说</text>
        </view>
        <view
          class="home-journey-reset"
          hover-class="home-journey-reset--pressed"
          hover-stay-time="80"
          @click.stop="restartNewUserGuide"
        >
          <text>重新开始</text>
        </view>
      </view>
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

    <view class="home-page-note" aria-hidden="true">
      <view class="home-page-note-line" />
      <text>每一次补充，都是留给家人的一页。</text>
    </view>

      <!-- #ifdef MP-WEIXIN -->
      <view class="home-share ui-simplified-hidden">
        <view class="home-share-copy">
          <text class="home-share-label">把家庭记录分享给家人</text>
          <text class="home-share-desc">邀请家人一起补充家庭树</text>
        </view>
        <button class="wechat-share-button" open-type="share">分享</button>
      </view>
      <!-- #endif -->
    </view>

    <MiniLaunchSplash :visible="launchSplashVisible" :config="launchSplashConfig" @close="closeLaunchSplash" @visual-ready="startLaunchSplashTimer" />
  </view>
</template>

<script setup lang="ts">
import { onLoad, onReady, onShareAppMessage, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { getPublicPlatformStats, listMyFamilies } from '@/api/families'
import { defaultLaunchSplashConfig, loadLaunchSplashConfig, type LaunchSplashConfig } from '@/api/launchSplash'
import { listPublishedShowcaseFamilies } from '@/api/showcaseFamilies'
import MiniLaunchSplash from '@/components/base/MiniLaunchSplash.vue'
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
import { markLaunchSplashShown, shouldShowLaunchSplash } from '@/features/home/launchSplash'
import { buildHomeSharePayload, loadShareConfigs } from '@/features/share/wechatShare'
import {
  dismissNewUserGuide,
  hasNewUserGuideMilestone,
  countNewUserGuideCompleted,
  areNewUserGuideMilestonesComplete,
  isNewUserGuideDismissed,
  isNewUserGuideReplay,
  readNewUserGuideProgress,
  resetNewUserGuideProgress,
  type NewUserGuideMilestone,
  type NewUserGuideProgress
} from '@/features/onboarding/newUserGuide'
import { useSessionStore } from '@/stores/session'
import type {
  FamilySummary,
  PublicPlatformStats,
  ShowcaseFamilyItem,
} from '@/types/api'

interface ArchiveStat {
  label: string
  value: string
}

type JourneyStepKey = 'SHOWCASE' | 'LOGIN' | 'PROFILE' | 'CREATE' | 'MEMBERS' | 'RELATIONS' | 'INVITE' | 'TREE'
type JourneyStep = { key: JourneyStepKey; label: string; description: string }

const session = useSessionStore()
const launchSplashVisible = ref(shouldShowLaunchSplash())
const launchSplashConfig = ref<LaunchSplashConfig>(defaultLaunchSplashConfig)
let launchSplashTimer: ReturnType<typeof setTimeout> | null = null
let launchSplashConfigReady = false
let launchSplashVisualReady = false
let launchSplashFallbackTimer: ReturnType<typeof setTimeout> | null = null
const myFamilies = ref<FamilySummary[]>([])
const archiveStats = ref<ArchiveStat[]>([
  { label: '家族', value: '—' },
  { label: '成员', value: '—' },
  { label: '关系', value: '—' }
])
const platformStatsLoaded = ref(false)
const hasOwnFamilies = ref(false)
const featuredShowcase = ref<ShowcaseFamilyItem | null>(null)
const showcaseLoaded = ref(false)
const showcaseLoadFailed = ref(false)
const pressedJourneyStepKey = ref<JourneyStepKey | null>(null)
const newUserGuideProgress = ref<NewUserGuideProgress>(readNewUserGuideProgress())
let pressedJourneyClearTimer: ReturnType<typeof setTimeout> | null = null
let primaryEntryOpening = false
const navigating = ref(false)
const visibleToolItems = computed(() => getVisibleToolDefinitions())
const platformStatsCaption = computed(() => platformStatsLoaded.value ? '平台累计' : '累计数据整理中')
const hasViewedShowcase = computed(() =>
  hasNewUserGuideMilestone(newUserGuideProgress.value, 'SHOWCASE_VIEWED')
)
const guideFamily = computed(() => myFamilies.value.find((family) => family.status === 'NORMAL') || myFamilies.value[0])
const guideFamilyCanManage = computed(() =>
  guideFamily.value?.role === 'FOUNDER' || guideFamily.value?.role === 'FAMILY_ADMIN'
)
const guideProgressText = computed(() =>
  guideFamily.value && !guideFamilyCanManage.value
    ? '按自己的节奏熟悉家庭树'
    : `已了解 ${guideCompletedCount.value}/${guideTotalCount} 个关键步骤`
)

const shouldPromptCreateFamily = computed(() => !hasOwnFamilies.value && session.isLoggedIn)
const shouldPromptProfile = computed(() => session.isLoggedIn && !session.isProfileComplete)
const directoryTagText = computed(() =>
  hasOwnFamilies.value ? ['家', '树'] : ['建', '树']
)
const journeyKicker = computed(() => {
  if (hasOwnFamilies.value) return '下一步'
  if (shouldPromptProfile.value) return '完善资料'
  if (!session.isLoggedIn && hasViewedShowcase.value) return '继续建谱'
  return session.isLoggedIn ? '开始建谱' : '新手引导'
})
const journeyTitle = computed(() => {
  if (hasOwnFamilies.value) return '继续完善家谱'
  if (shouldPromptProfile.value) return '先完善个人资料'
  if (!session.isLoggedIn && hasViewedShowcase.value) return '登录后继续建谱'
  return session.isLoggedIn ? '建立第一本家谱' : '先看示例，再建家谱'
})
const journeyDesc = computed(() => {
  if (hasOwnFamilies.value) return '先补齐成员和关系，再邀请家人'
  if (shouldPromptProfile.value) return '先补充昵称，再创建和邀请'
  if (!session.isLoggedIn && hasViewedShowcase.value) return '已看过示例，登录后会从创建和完善资料继续'
  return session.isLoggedIn ? '先填写姓氏，生成本人节点' : '先浏览示例，再创建家庭树'
})
const journeySteps = computed<JourneyStep[]>(() => {
  const inviteLabel = session.user?.phoneVerified ? '邀请家人' : '绑定并邀请'
  if (guideFamily.value && !guideFamilyCanManage.value) {
    return [
      { key: 'TREE', label: '看看家庭树', description: '从自己的节点认识家人关系' },
      { key: 'PROFILE', label: '补充个人资料', description: '方便家人识别你，可稍后完成' },
      { key: 'INVITE', label: '帮家庭助力', description: '分享助力邀请，为家庭解锁名额' }
    ]
  }
  if (guideFamily.value) {
    return [
      { key: 'MEMBERS', label: '补充成员', description: '从一位最熟悉的亲属开始' },
      { key: 'RELATIONS', label: '连好关系', description: '用父母、配偶、子女建立脉络' },
      { key: 'INVITE', label: inviteLabel, description: '邀请对应节点的家人来确认身份' }
    ]
  }
  if (shouldPromptProfile.value) {
    return [
      { key: 'PROFILE', label: '补充昵称', description: '头像可跳过，昵称用于家人识别' },
      { key: 'CREATE', label: '创建家谱', description: '填写本人、姓氏和籍贯即可开始' },
      { key: 'SHOWCASE', label: '先看示例', description: '参考真实家庭如何整理家谱' }
    ]
  }
  if (session.isLoggedIn) {
    return [
      { key: 'CREATE', label: '创建家谱', description: '填写本人、姓氏和籍贯即可开始' },
      { key: 'SHOWCASE', label: '先看示例', description: '先了解一棵家庭树如何组织' },
      { key: 'PROFILE', label: '补充昵称', description: '需要创建或邀请时再补齐也可以' }
    ]
  }
  if (hasViewedShowcase.value) {
    return [
      { key: 'LOGIN', label: '微信登录', description: '登录后保存你的家庭记录' },
      { key: 'CREATE', label: '创建家谱', description: '登录后填写本人和家庭信息' },
      { key: 'SHOWCASE', label: '继续看示例', description: '随时回来参考整理方式' }
    ]
  }
  return [
    { key: 'SHOWCASE', label: '看示例', description: '先用一棵示例树熟悉家谱' },
    { key: 'LOGIN', label: '微信登录', description: '登录后可创建、加入和维护家庭' },
    { key: 'CREATE', label: '创建家谱', description: '从自己的信息开始建立第一棵树' }
  ]
})
const journeyMilestoneByStep: Partial<Record<JourneyStepKey, NewUserGuideMilestone>> = {
  SHOWCASE: 'SHOWCASE_VIEWED',
  LOGIN: 'LOGIN_DONE',
  PROFILE: 'PROFILE_DONE',
  CREATE: 'FAMILY_CREATED',
  MEMBERS: 'MEMBER_GUIDE_OPENED',
  RELATIONS: 'RELATION_GUIDE_OPENED',
  INVITE: 'INVITE_GUIDE_OPENED',
  TREE: 'TREE_GUIDE_OPENED'
}
const journeyRequiredMilestones = computed(() =>
  journeySteps.value
    // A joined member can fill in a nickname later. Before a first family is
    // created, profile completion is the required gate for the next action.
    .filter((item) => item.key !== 'PROFILE' || !hasOwnFamilies.value)
    .map((item) => journeyMilestoneByStep[item.key])
    .filter((milestone): milestone is NewUserGuideMilestone => Boolean(milestone))
)
const nextJourneyStepKey = computed(() =>
  journeySteps.value.find((item) => !journeyStepDone(item.key))?.key || null
)
const guideCompletedCount = computed(() =>
  countNewUserGuideCompleted(newUserGuideProgress.value, journeyRequiredMilestones.value)
)
const guideTotalCount = computed(() => journeyRequiredMilestones.value.length)
const shouldShowJourney = computed(() =>
  !areNewUserGuideMilestonesComplete(newUserGuideProgress.value, journeyRequiredMilestones.value) &&
  !isNewUserGuideDismissed(newUserGuideProgress.value)
)
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
      ? '查看家庭成员与家谱结构'
      : '选择要查看的家庭树'
    : shouldPromptCreateFamily.value
      ? '填写姓氏，建立自己的家庭树'
      : '登录并完善资料后，开始记录家庭关系'
)

async function loadArchiveStats() {
  session.restoreSession()
  await loadPlatformStats()
  await refreshMyFamilies()
  syncNewUserGuideProgress()
}

async function refreshMyFamilies(): Promise<FamilySummary[] | null> {
  if (!session.isLoggedIn || !session.isProfileComplete) {
    myFamilies.value = []
    hasOwnFamilies.value = false
    return []
  }
  try {
    const families = await listMyFamilies()
    myFamilies.value = families
    hasOwnFamilies.value = families.length > 0
    return families
  } catch {
    return null
  }
}

async function loadPlatformStats() {
  try {
    const stats = await getPublicPlatformStats()
    archiveStats.value = platformStatsToArchiveStats(stats)
    platformStatsLoaded.value = true
  } catch {
    archiveStats.value = [
      { label: '家族', value: '—' },
      { label: '成员', value: '—' },
      { label: '关系', value: '—' }
    ]
    platformStatsLoaded.value = false
  }
}

function platformStatsToArchiveStats(stats: PublicPlatformStats): ArchiveStat[] {
  return [
    { label: '家族', value: formatCount(stats.familyCount) },
    { label: '成员', value: formatCount(stats.memberCount) },
    { label: '关系', value: formatCount(stats.relationshipCount) }
  ]
}

function formatCount(value: number) {
  const count = Number.isFinite(value) && value > 0 ? Math.floor(value) : 0
  if (count >= 10000) {
    const wan = count / 10000
    return `${wan >= 100 ? Math.round(wan) : wan.toFixed(1).replace(/\.0$/, '')}万`
  }
  return String(count)
}

function go(url: string) {
  if (navigating.value) return
  navigating.value = true
  uni.navigateTo({
    url,
    complete: () => {
      navigating.value = false
    }
  })
}

function openHistoricalShowcase(showcaseFamilyId?: number, options: { guide?: boolean } = {}) {
  const params: string[] = []
  if (showcaseFamilyId) params.push(`familyId=${encodeURIComponent(String(showcaseFamilyId))}`)
  if (options.guide) params.push('guide=1')
  const query = params.length > 0 ? `?${params.join('&')}` : ''
  go(`/pages/family/historical-showcase${query}`)
}

async function loadFeaturedShowcase() {
  showcaseLoaded.value = false
  showcaseLoadFailed.value = false
  try {
    const items = await listPublishedShowcaseFamilies()
    featuredShowcase.value = items.find((item) => item.rotationPinned) || items[0] || null
  } catch {
    featuredShowcase.value = null
    showcaseLoadFailed.value = true
  } finally {
    showcaseLoaded.value = true
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
  if (navigating.value) return
  navigating.value = true
  uni.switchTab({
    url: '/pages/family/my',
    complete: () => {
      navigating.value = false
    }
  })
}

async function openCreateFamily() {
  if (!session.isLoggedIn) {
    session.requireLogin('/pages/home/index')
    return
  }
  if (!session.isProfileComplete) {
    openProfileOnboarding()
    return
  }
  const families = await refreshMyFamilies()
  if (families === null) {
    uni.showToast({ title: '家庭信息加载失败，请稍后重试', icon: 'none' })
    return
  }
  if (openExistingFamily(families)) return
  go('/pages/family/create')
}

function openLoginContinuation() {
  session.requireLogin('/pages/home/index')
}

function openProfileOnboarding() {
  if (!session.isLoggedIn) {
    session.requireLogin('/pages/me/profile?onboarding=1')
    return
  }
  go('/pages/me/profile?onboarding=1')
}

function openExistingFamily(families: FamilySummary[]) {
  const normalFamilies = families.filter((family) => family.status === 'NORMAL')
  if (normalFamilies.length === 1) {
    go(`/pages/family/detail?familyId=${encodeURIComponent(String(normalFamilies[0].id))}`)
    return true
  }
  if (families.length > 0) {
    goMyFamilyTab()
    return true
  }
  return false
}

async function openHomePrimaryEntry() {
  if (primaryEntryOpening) return
  primaryEntryOpening = true
  try {
    await openCreateFamily()
  } finally {
    primaryEntryOpening = false
  }
}

function openJourneyPrimary() {
  continueNewUserGuide()
}

function continueNewUserGuide() {
  const nextStep = nextJourneyStepKey.value
  if (nextStep) openJourneyStep(nextStep)
}

function pressJourneyStep(step: JourneyStepKey) {
  if (pressedJourneyClearTimer) clearTimeout(pressedJourneyClearTimer)
  pressedJourneyStepKey.value = step
}

function clearPressedJourneyStep() {
  if (pressedJourneyClearTimer) clearTimeout(pressedJourneyClearTimer)
  pressedJourneyClearTimer = setTimeout(() => {
    pressedJourneyStepKey.value = null
    pressedJourneyClearTimer = null
  }, 120)
}

function openJourneyStep(step: JourneyStepKey) {
  clearPressedJourneyStep()
  if (step === 'SHOWCASE') {
    openHistoricalShowcase(featuredShowcase.value?.id, { guide: true })
    return
  }
  if (step === 'LOGIN') {
    openLoginContinuation()
    return
  }
  if (step === 'PROFILE') {
    openProfileOnboarding()
    return
  }
  if (step === 'CREATE') {
    openCreateFamily()
    return
  }

  if (!guideFamily.value) {
    openCreateFamily()
    return
  }

  const familyId = encodeURIComponent(String(guideFamily.value.id))
  if (step === 'TREE') {
    go(`/pages/family/private-tree?familyId=${familyId}&guide=1`)
    return
  }
  if (step === 'INVITE') {
    if (!guideFamilyCanManage.value) {
      go(`/pages/family/invitation-reward?familyId=${familyId}&guide=1`)
      return
    }
    go(`/pages/invite/sent?familyId=${familyId}&guide=1`)
    return
  }
  if (!guideFamilyCanManage.value) {
    go(`/pages/family/private-tree?familyId=${familyId}&guide=1`)
    return
  }
  if (step === 'RELATIONS') {
    go(`/pages/family/manage?familyId=${familyId}&section=relations&guide=1`)
    return
  }
  go(`/pages/family/manage?familyId=${familyId}&section=node&guide=1`)
}

function journeyStepDone(step: JourneyStepKey) {
  const milestone = journeyMilestoneByStep[step]
  return milestone ? hasNewUserGuideMilestone(newUserGuideProgress.value, milestone) : false
}

function dismissNewUserGuideCard() {
  newUserGuideProgress.value = dismissNewUserGuide(3)
  uni.showToast({ title: '3 天内不再提示，可在“我的”重新打开', icon: 'none' })
}

function restartNewUserGuide() {
  uni.showModal({
    title: '重新开始新手引导',
    content: '会重置本机引导进度，不会影响已创建的家庭和成员资料。',
    confirmText: '重新开始',
    cancelText: '取消',
    success(result) {
      if (!result.confirm) return
      newUserGuideProgress.value = resetNewUserGuideProgress()
      uni.showToast({ title: '已重置，请从第一步开始', icon: 'none' })
    }
  })
}

function syncNewUserGuideProgress() {
  if (isNewUserGuideReplay(newUserGuideProgress.value)) return
  // Do not infer completion from existing data. A guide step only completes after
  // the user explicitly follows it, otherwise returning users lose their route.
  newUserGuideProgress.value = readNewUserGuideProgress()
}

async function refreshHome() {
  if (homeRefreshedAt > 0 && Date.now() - homeRefreshedAt < homeCacheTtlMs) return
  if (homeRefreshPromise) return homeRefreshPromise
  homeRefreshPromise = Promise.all([loadArchiveStats(), loadFeaturedShowcase()]).then(() => {
    homeRefreshedAt = Date.now()
  }).then(() => undefined).finally(() => {
    homeRefreshPromise = null
  })
  return homeRefreshPromise
}

let homeRefreshPromise: Promise<void> | null = null
let homeRefreshedAt = 0
const homeCacheTtlMs = 15_000

function closeLaunchSplash() {
  launchSplashVisible.value = false
  markLaunchSplashShown()
  if (launchSplashTimer) {
    clearTimeout(launchSplashTimer)
    launchSplashTimer = null
  }
  if (launchSplashFallbackTimer) { clearTimeout(launchSplashFallbackTimer); launchSplashFallbackTimer = null }
}

function startLaunchSplashTimer() {
  launchSplashVisualReady = true
  if (!launchSplashVisible.value || !launchSplashConfigReady || launchSplashTimer) return
  launchSplashTimer = setTimeout(closeLaunchSplash, launchSplashConfig.value.durationMs)
}

onLoad(() => {
  void loadToolConfigs()
  void refreshHome()
  void loadLaunchSplashConfig().then((config) => {
    launchSplashConfig.value = config
    launchSplashConfigReady = true
    if (!config.enabled) {
      closeLaunchSplash()
      return
    }
    if (!config.imageUrl && !config.dayImageUrl && !config.nightImageUrl) startLaunchSplashTimer()
    if (launchSplashVisualReady) startLaunchSplashTimer()
  })
})

onReady(() => { if (launchSplashVisible.value) launchSplashFallbackTimer = setTimeout(startLaunchSplashTimer, 2800) })
loadShareConfigs()

onShareAppMessage(() => buildHomeSharePayload())

onShow(() => {
  promptPrivacyConsentIfNeeded()
  void loadToolConfigs()
  void loadShareConfigs()
  void refreshHome()
})
</script>

<style scoped>
.ui-simplified-hidden {
  display: none !important;
}

.home-page {
  padding-right: 0;
  padding-bottom: calc(132rpx + env(safe-area-inset-bottom));
}

.home-page--splashing .home-page-content {
  visibility: hidden;
}

.home-book {
  position: relative;
  display: flex;
  min-height: 464rpx;
  margin: 0;
}

.home-book-main {
  position: relative;
  flex: 1;
  min-width: 0;
  border-top: 1rpx solid rgba(92, 74, 48, 0.17);
  border-bottom: 1rpx solid rgba(92, 74, 48, 0.11);
  background: linear-gradient(180deg, rgba(255, 252, 244, 0.48), rgba(248, 239, 220, 0.2));
  padding: 38rpx 88rpx 30rpx 6rpx;
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
  max-width: 560rpx;
  font-size: 60rpx;
  letter-spacing: 1rpx;
  line-height: 1.06;
}

.home-title-copy .archive-subtitle {
  margin-top: 14rpx;
  max-width: 486rpx;
  font-size: 30rpx;
  line-height: 1.42;
}

.home-title-seal {
  flex-shrink: 0;
  margin-top: 2rpx;
  transform: scale(0.86) rotate(-7deg);
  transform-origin: top right;
}

.home-stats-caption {
  display: inline-flex;
  align-items: center;
  margin-top: 24rpx;
  border-left: 4rpx solid var(--archive-cinnabar);
  padding-left: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 25rpx;
  line-height: 1.2;
}

.home-stats {
  display: flex;
  flex-direction: row;
  gap: 0;
  width: auto;
  margin-top: 16rpx;
  border-top: 1rpx solid rgba(92, 74, 48, 0.16);
  border-bottom: 1rpx solid rgba(92, 74, 48, 0.1);
  background: rgba(255, 252, 244, 0.42);
}

.home-stat {
  flex: 1;
  min-width: 0;
  border-right: 1rpx solid rgba(92, 74, 48, 0.12);
  padding: 18rpx 14rpx 17rpx;
}

.home-stat:last-child {
  border-right: 0;
}

.home-stat-value,
.home-stat-label {
  display: block;
}

.home-stat-value {
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 46rpx;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.15;
  font-variant-numeric: tabular-nums;
}

.home-stat-label {
  margin-top: 6rpx;
  color: var(--archive-ink-soft);
  font-size: 25rpx;
  line-height: 1.2;
}

.bamboo-ink {
  position: absolute;
  right: 62rpx;
  top: 90rpx;
  width: 174rpx;
  height: 250rpx;
  opacity: 0.18;
  pointer-events: none;
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
  left: 48rpx;
  height: 260rpx;
  transform: rotate(8deg);
}

.bamboo-stem-b {
  left: 80rpx;
  height: 230rpx;
  transform: rotate(-5deg);
}

.bamboo-stem-c {
  left: 112rpx;
  height: 202rpx;
  transform: rotate(4deg);
}

.bamboo-leaf {
  position: absolute;
  width: 70rpx;
  height: 18rpx;
  border-radius: 100% 0 100% 0;
  background: var(--archive-bamboo);
  transform-origin: left center;
}

.bamboo-leaf-a {
  top: 30rpx;
  left: 56rpx;
  transform: rotate(-32deg);
}

.bamboo-leaf-b {
  top: 70rpx;
  left: 18rpx;
  transform: rotate(208deg);
}

.bamboo-leaf-c {
  top: 100rpx;
  left: 72rpx;
  transform: rotate(-16deg);
}

.bamboo-leaf-d {
  top: 150rpx;
  left: 24rpx;
  transform: rotate(192deg);
}

.bamboo-leaf-e {
  top: 192rpx;
  left: 92rpx;
  transform: rotate(-24deg);
}

.bamboo-leaf-f {
  top: 232rpx;
  left: 42rpx;
  transform: rotate(186deg);
}

.bamboo-mist {
  position: absolute;
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  background: rgba(37, 76, 59, 0.06);
}

.bamboo-mist-a {
  top: 48rpx;
  left: 8rpx;
}

.bamboo-mist-b {
  top: 160rpx;
  left: 58rpx;
}

.home-primary-entry {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 22rpx;
  min-height: 136rpx;
  margin-top: 26rpx;
  border: 1rpx solid rgba(255, 252, 244, 0.18);
  border-radius: 10rpx;
  background: linear-gradient(135deg, var(--archive-blue) 0%, #1f466c 100%);
  padding: 25rpx 28rpx;
  box-shadow: 0 14rpx 28rpx rgba(22, 51, 83, 0.18), inset 0 1rpx 0 rgba(255, 255, 255, 0.18);
}

.home-primary-entry .archive-arrow {
  color: rgba(255, 252, 244, 0.94);
  font-size: 46rpx;
}

.home-tools-panel {
  margin: 22rpx 24rpx 24rpx 0;
  border-top: 1rpx solid rgba(92, 74, 48, 0.18);
  border-bottom: 1rpx solid rgba(92, 74, 48, 0.1);
  background: rgba(255, 252, 244, 0.22);
  padding: 24rpx 18rpx 22rpx;
}

.home-showcase-panel {
  min-width: 0;
  border: 1rpx solid rgba(180, 144, 71, 0.26);
  border-radius: 8rpx;
  background: linear-gradient(180deg, rgba(255, 251, 239, 0.76), rgba(255, 252, 244, 0.42));
  box-shadow: 0 6rpx 18rpx rgba(91, 70, 40, 0.045), inset 0 1rpx 0 rgba(255, 255, 255, 0.72);
  padding: 22rpx 20rpx 18rpx;
}

.home-discovery-swiper {
  height: 356rpx;
  margin: 20rpx 24rpx 18rpx 0;
}

.home-discovery-swiper swiper-item {
  height: 100%;
}

.home-discovery-swiper .home-showcase-panel {
  height: 100%;
  box-sizing: border-box;
}

.home-discovery-swiper .home-showcase-head {
  align-items: flex-start;
  min-height: 74rpx;
}

.home-discovery-swiper .home-showcase-title {
  font-size: 30rpx;
}

.home-discovery-swiper .home-tools-more {
  flex: 0 0 auto;
  font-size: 22rpx;
}

.home-discovery-swiper .home-showcase-card {
  min-height: 164rpx;
  margin-top: 12rpx;
  padding: 16rpx;
}

.home-discovery-swiper .home-showcase-seal {
  width: 52rpx;
  height: 52rpx;
  font-size: 26rpx;
}

.home-discovery-swiper .home-showcase-name {
  font-size: 27rpx;
  white-space: normal;
}

.home-discovery-swiper .home-showcase-meta,
.home-discovery-swiper .home-showcase-desc {
  font-size: 21rpx;
  white-space: normal;
  line-height: 1.45;
}

.home-discovery-swiper .home-showcase-note {
  min-height: 34rpx;
  font-size: 20rpx;
}

.home-quality-panel {
  border-color: rgba(47, 107, 87, 0.24);
  background: linear-gradient(180deg, rgba(241, 249, 243, 0.88), rgba(255, 252, 244, 0.5));
}

.home-quality-seal {
  border-color: var(--archive-green, #2f6b57) !important;
  color: var(--archive-green, #2f6b57) !important;
}

.home-showcase-head { display: flex; align-items: baseline; justify-content: space-between; gap: 16rpx; }
.home-showcase-title { display: block; margin-top: 3rpx; color: var(--archive-ink); font-size: 34rpx; font-weight: 780; }
.home-showcase-card { position: relative; display: flex; align-items: center; gap: 16rpx; min-height: 138rpx; margin-top: 16rpx; overflow: hidden; border: 1rpx solid rgba(92, 74, 48, 0.14); border-radius: 10rpx; background: rgba(255, 252, 244, 0.8); padding: 20rpx; box-shadow: inset 0 1rpx 0 rgba(255, 255, 255, 0.58); }
.home-showcase-cover { width: 116rpx; height: 92rpx; flex: 0 0 auto; border-radius: 8rpx; }
.home-showcase-seal { display: grid; width: 62rpx; height: 62rpx; flex: 0 0 auto; place-items: center; border: 2rpx solid var(--archive-cinnabar); color: var(--archive-cinnabar); font-family: 'Songti SC', 'STSong', serif; font-size: 30rpx; transform: rotate(-7deg); }
.home-showcase-copy { display: grid; min-width: 0; flex: 1; gap: 5rpx; }
.home-showcase-name, .home-showcase-meta, .home-showcase-desc { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.home-showcase-name { color: var(--archive-ink); font-size: 31rpx; font-weight: 760; }
.home-showcase-meta { color: var(--archive-blue); font-size: 25rpx; }
.home-showcase-desc { color: var(--archive-ink-soft); font-size: 24rpx; }
.home-showcase-note { display: block; margin-top: 12rpx; color: var(--archive-ink-soft); font-size: 23rpx; }

.home-page-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  margin: 0 24rpx 0 0;
  color: rgba(101, 112, 128, 0.78);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 21rpx;
  letter-spacing: 1rpx;
  line-height: 1.4;
}

.home-journey {
  position: relative;
  margin: 0 24rpx 20rpx 0;
  border: 1rpx solid rgba(168, 59, 45, 0.28);
  border-left: 6rpx solid var(--archive-cinnabar);
  border-radius: 10rpx;
  background:
    linear-gradient(135deg, rgba(168, 59, 45, 0.1), transparent 44%),
    linear-gradient(180deg, rgba(255, 252, 244, 0.88), rgba(255, 248, 235, 0.54));
  box-shadow: 0 10rpx 24rpx rgba(91, 70, 40, 0.06), inset 0 1rpx 0 rgba(255, 255, 255, 0.72);
  padding: 22rpx 20rpx 18rpx;
}

.home-journey-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
}

.home-journey-title,
.home-journey-desc,
.home-journey-kicker,
.home-journey-index,
.home-journey-label,
.home-journey-step-desc,
.home-journey-current,
.home-journey-done,
.home-journey-step > text:last-child {
  display: block;
}

.home-journey-kicker {
  color: var(--archive-cinnabar);
  font-size: 19rpx;
  font-weight: 800;
  letter-spacing: 4rpx;
  line-height: 1.2;
}

.home-journey-title {
  margin-top: 6rpx;
  color: var(--archive-ink);
  font-size: 36rpx;
  font-weight: 820;
  line-height: 1.22;
}

.home-journey-desc {
  margin-top: 6rpx;
  color: var(--archive-ink-soft);
  font-size: 26rpx;
  line-height: 1.35;
}

.home-journey-steps {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0;
  margin-top: 18rpx;
  overflow: hidden;
  border: 1rpx solid rgba(92, 74, 48, 0.12);
  border-radius: 8rpx;
  background: rgba(255, 252, 244, 0.58);
}

.home-journey-step {
  display: flex;
  flex-direction: column;
  position: relative;
  min-width: 0;
  border-left: 1rpx solid rgba(168, 59, 45, 0.16);
  padding: 15rpx 10rpx 12rpx;
  transition: background-color 120ms ease, transform 120ms ease, opacity 120ms ease;
}

.home-journey-step:first-child {
  border-left: 0;
}

.home-journey-step--pressed,
.home-journey-step.is-pressed {
  background: rgba(168, 59, 45, 0.13);
  opacity: 0.9;
  transform: translateY(1rpx);
}

.home-journey-step.is-current {
  background:
    linear-gradient(180deg, rgba(168, 59, 45, 0.12), rgba(168, 59, 45, 0.045));
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
  font-size: 21rpx;
  font-weight: 820;
  letter-spacing: 1rpx;
}

.home-journey-step-copy {
  min-width: 0;
}

.home-journey-label {
  margin-top: 2rpx;
  color: var(--archive-ink);
  font-size: 26rpx;
  font-weight: 660;
  line-height: 1.3;
  white-space: nowrap;
}

.home-journey-step-desc {
  min-height: 54rpx;
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
  line-height: 1.35;
}

.home-journey-done {
  margin-top: 8rpx;
  color: var(--archive-green);
  font-size: 20rpx;
  font-weight: 700;
}

.home-journey-current {
  margin-top: 5rpx;
  color: rgba(168, 59, 45, 0.78);
  font-size: 19rpx;
  line-height: 1.2;
  white-space: nowrap;
}

.home-journey-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  margin-top: 16rpx;
  border-top: 1rpx solid rgba(92, 74, 48, 0.1);
  padding-top: 16rpx;
}

.home-journey-progress {
  display: block;
  min-width: 0;
  flex: 1;
  color: rgba(101, 112, 128, 0.86);
  font-size: 23rpx;
  line-height: 1.3;
}

.home-journey-continue,
.home-journey-reset {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 52rpx;
  flex-shrink: 0;
  border: 1rpx solid rgba(31, 58, 95, 0.22);
  border-radius: 8rpx;
  background: rgba(255, 252, 244, 0.64);
  padding: 0 18rpx;
  color: var(--archive-blue);
  font-size: 24rpx;
  font-weight: 700;
  line-height: 1;
  transition: transform 120ms ease, background-color 120ms ease, opacity 120ms ease;
}

.home-journey-continue {
  border-color: rgba(168, 59, 45, 0.36);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
}

.home-journey-reset--pressed {
  background: rgba(31, 58, 95, 0.1);
  opacity: 0.88;
  transform: translateY(1rpx);
}

.home-page-note-line {
  width: 46rpx;
  height: 1rpx;
  background: rgba(92, 74, 48, 0.2);
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
  font-size: 34rpx;
  font-weight: 780;
}

.home-tools-more {
  display: inline-flex;
  align-items: center;
  min-height: 50rpx;
  color: var(--archive-blue);
  font-size: 27rpx;
}

.home-tools-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14rpx;
  margin-top: 18rpx;
}

.home-tool-card {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-width: 0;
  min-height: 146rpx;
  border: 1rpx solid rgba(92, 74, 48, 0.16);
  border-radius: 10rpx;
  background: rgba(255, 252, 244, 0.7);
  padding: 19rpx 17rpx;
  box-sizing: border-box;
  box-shadow: inset 0 1rpx 0 rgba(255, 255, 255, 0.52);
}

.home-tool-card.is-tool-highlighted {
  border-color: rgba(168, 59, 45, 0.44);
  background: rgba(255, 250, 241, 0.76);
  box-shadow: 0 4rpx 12rpx rgba(138, 74, 52, 0.06), inset 0 1rpx 0 rgba(255, 255, 255, 0.6);
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
  width: 54rpx;
  height: 54rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-cinnabar);
  border-radius: 50%;
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 26rpx;
  font-weight: 700;
}

.relation-tool-icon {
  border-color: var(--archive-line-strong);
}

.home-tool-dot {
  position: absolute;
  width: 8rpx;
  height: 8rpx;
  border: 1rpx solid var(--archive-blue);
  border-radius: 50%;
  background: var(--archive-paper-light);
}

.dot-top {
  top: 8rpx;
  left: 20rpx;
}

.dot-left {
  bottom: 9rpx;
  left: 8rpx;
}

.dot-right {
  right: 8rpx;
  bottom: 9rpx;
}

.home-tool-branch {
  position: absolute;
  top: 24rpx;
  left: 24rpx;
  width: 16rpx;
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
  font-size: 30rpx;
  font-weight: 760;
}

.home-tool-desc {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  margin-top: 4rpx;
  color: var(--archive-ink-soft);
  font-size: 25rpx;
  line-height: 1.32;
  white-space: normal;
}

.entry-title,
.entry-desc {
  display: block;
}

.entry-title {
  color: rgba(255, 252, 244, 0.96);
  font-size: 38rpx;
  font-weight: 780;
  line-height: 1.22;
}

.entry-desc {
  margin-top: 5rpx;
  color: rgba(255, 252, 244, 0.76);
  font-size: 27rpx;
  line-height: 1.32;
}

.home-directory-tag {
  position: absolute;
  right: 52rpx;
  top: 188rpx;
  z-index: 4;
  width: 44rpx;
  min-height: 116rpx;
  flex-direction: column;
  gap: 6rpx;
  box-shadow: 4rpx 6rpx 14rpx rgba(77, 59, 35, 0.1);
  transform: rotate(2deg);
}

.home-spine {
  width: 58rpx;
  min-height: 464rpx;
  flex-shrink: 0;
}

.spine-title {
  position: absolute;
  top: 68rpx;
  left: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7rpx;
  color: rgba(255, 255, 255, 0.92);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 25rpx;
  font-weight: 700;
  line-height: 1.2;
  transform: translateX(-50%);
}

.spine-stitches {
  position: absolute;
  right: 21rpx;
  bottom: 48rpx;
  display: flex;
  flex-direction: column;
  gap: 19rpx;
}

.spine-stitch {
  width: 12rpx;
  height: 12rpx;
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
