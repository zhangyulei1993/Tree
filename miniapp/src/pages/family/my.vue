<template>
  <view class="archive-page my-family-page">
    <view v-if="isResolvingMyFamilyMode" class="my-state archive-form-panel my-resolving-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">正在翻阅家庭册页</text>
        <text class="archive-section-subtitle">请稍候，正在确认你的家庭记录。</text>
      </view>
      <view class="my-loading-lines">
        <view class="my-loading-line long" />
        <view class="my-loading-line" />
        <view class="my-loading-line short" />
      </view>
    </view>

    <template v-else-if="hasMyFamilies">
      <view class="my-head archive-page-head">
        <view>
          <text class="archive-kicker">My Archive</text>
          <text class="archive-title">我的家庭</text>
          <text class="archive-subtitle">你创建或加入的家庭</text>
        </view>
        <view class="archive-seal">家</view>
      </view>

      <view class="my-toolbar">
        <text class="archive-chip">共 {{ families.length }} 个家庭</text>
        <text class="archive-thin-button" @click="openCreateFamily">创建家庭</text>
      </view>

      <view v-if="errorMessage" class="my-state archive-form-panel">
        <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
        <MiniButton variant="secondary" @click="loadMyFamilies">重新加载</MiniButton>
      </view>

      <view v-else-if="loadingMyFamilies" class="my-state archive-form-panel">
        <MiniEmptyState symbol="…" title="正在加载" description="正在加载家庭列表..." />
      </view>

      <view v-else class="my-family-list archive-list">
        <view
          v-for="family in families"
          :key="family.id"
          class="archive-row"
          @click="openFamily(family.id)"
        >
          <text class="archive-surname-stamp">{{ family.familySurname.slice(0, 1) }}</text>
          <view class="archive-row-main">
            <text class="archive-row-title">{{ family.familyName }}</text>
            <text class="archive-row-desc">{{ familyRegionLabel(family) }}</text>
          </view>
          <text class="archive-row-meta">{{ roleText(family.role) }}</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </template>

    <template v-else>
      <view class="my-head archive-page-head">
        <view>
          <text class="archive-kicker">Public Archive</text>
          <text class="archive-title">展示家庭</text>
          <text class="archive-subtitle">浏览经平台审核的公开家庭主页</text>
        </view>
        <view class="archive-seal">展</view>
      </view>

      <view class="showcase-notice archive-panel">
        <text class="showcase-notice-label">公开册页</text>
        <text class="showcase-notice-text">加入家庭请通过成员邀请，暂不提供公开搜索或申请加入。</text>
      </view>

      <view class="create-family-entry archive-panel" @click="openCreateFamily">
        <view class="create-family-copy">
          <text class="create-family-title">新建家谱册</text>
          <text class="create-family-desc">{{ createFamilyHint }}</text>
        </view>
        <text class="archive-thin-button">{{ createFamilyActionText }}</text>
      </view>

      <view v-if="session.isLoggedIn" class="showcase-invite-link" @click="openInvitations">
        <text class="showcase-invite-label">查看收到的家庭邀请</text>
        <text class="archive-arrow">›</text>
      </view>

      <view v-if="showcaseError" class="my-state archive-form-panel">
        <MiniNotice tone="warm" title="加载失败">{{ showcaseError }}</MiniNotice>
        <MiniButton variant="secondary" @click="reloadShowcase">重新加载</MiniButton>
      </view>

      <view v-else-if="showcaseLoading && showcaseItems.length === 0" class="my-state archive-form-panel">
        <MiniEmptyState symbol="展" title="正在加载" description="正在读取展示家庭列表..." />
      </view>

      <template v-else>
        <view v-if="showcaseItems.length > 0" class="showcase-list archive-list">
          <text class="showcase-count">共 {{ showcaseTotal }} 个展示家庭</text>
          <view
            v-for="family in showcaseItems"
            :key="family.id"
            class="archive-row"
            @click="openShowcaseFamily(family.id)"
          >
            <text class="archive-surname-stamp">{{ family.familySurname.slice(0, 1) }}</text>
            <view class="archive-row-main">
              <text class="archive-row-title">{{ family.familyName }}</text>
              <text class="archive-row-desc">{{ showcaseDesc(family) }}</text>
            </view>
            <text class="archive-row-meta">公开</text>
            <text class="archive-arrow">›</text>
          </view>
        </view>

        <view v-else class="my-state archive-form-panel">
          <MiniEmptyState
            symbol="展"
            title="暂无展示家庭"
            description="当前没有可浏览的展示家庭。你可以通过家人分享的链接访问家庭主页。"
          />
        </view>

        <view v-if="showcaseHasMore" class="showcase-load-more">
          <text
            class="archive-thin-button"
            :class="{ disabled: showcaseLoadingMore }"
            @click="loadMoreShowcase"
          >
            {{ showcaseLoadingMore ? '加载中…' : '加载更多' }}
          </text>
        </view>
      </template>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { apiErrorMessage } from '@/api/client'
import { listMyFamilies, listPublicFamilyShowcase } from '@/api/families'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilySummary, PublicFamilyShowcaseItem } from '@/types/api'

const session = useSessionStore()
const families = ref<FamilySummary[]>([])
const loadingMyFamilies = ref(false)
const errorMessage = ref('')
const myFamiliesLoaded = ref(false)

const showcaseItems = ref<PublicFamilyShowcaseItem[]>([])
const showcaseTotal = ref(0)
const showcasePage = ref(1)
const showcaseLoading = ref(false)
const showcaseLoadingMore = ref(false)
const showcaseError = ref('')
const showcasePageSize = 20

const hasMyFamilies = computed(() => myFamiliesLoaded.value && families.value.length > 0)
const isResolvingMyFamilyMode = computed(() => loadingMyFamilies.value && !myFamiliesLoaded.value)
const showcaseHasMore = computed(() => showcaseItems.value.length < showcaseTotal.value)
const createFamilyActionText = computed(() => (session.isLoggedIn ? '创建家庭' : '登录后创建'))
const createFamilyHint = computed(() =>
  session.isLoggedIn
    ? session.isProfileComplete
      ? '准备自己整理家谱？填写姓氏与地区，开启新的家谱册。'
      : '准备自己整理家谱？完善资料后即可创建家庭。'
    : '准备自己整理家谱？登录并完善资料后可创建家庭。'
)

async function loadMyFamilies() {
  session.restoreSession()
  if (!session.isLoggedIn || !session.isProfileComplete) {
    myFamiliesLoaded.value = true
    families.value = []
    return
  }
  loadingMyFamilies.value = true
  errorMessage.value = ''
  try {
    families.value = await listMyFamilies()
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭列表加载失败。')
    families.value = []
  } finally {
    loadingMyFamilies.value = false
    myFamiliesLoaded.value = true
  }
}

async function fetchShowcasePage(nextPage: number, append: boolean) {
  if (append) {
    showcaseLoadingMore.value = true
  } else {
    showcaseLoading.value = true
    showcaseError.value = ''
  }
  try {
    const result = await listPublicFamilyShowcase({ page: nextPage, pageSize: showcasePageSize })
    showcaseTotal.value = result.total
    showcasePage.value = result.page
    showcaseItems.value = append ? [...showcaseItems.value, ...result.items] : result.items
  } catch (error) {
    if (!append) {
      showcaseItems.value = []
      showcaseTotal.value = 0
    }
    showcaseError.value = apiErrorMessage(error, '展示家庭列表加载失败。')
  } finally {
    showcaseLoading.value = false
    showcaseLoadingMore.value = false
  }
}

function reloadShowcase() {
  showcasePage.value = 1
  fetchShowcasePage(1, false)
}

function loadMoreShowcase() {
  if (!showcaseHasMore.value || showcaseLoadingMore.value) return
  fetchShowcasePage(showcasePage.value + 1, true)
}

async function refreshPage() {
  await loadMyFamilies()
  if (!hasMyFamilies.value) {
    await fetchShowcasePage(1, false)
  }
}

function openFamily(familyId: number | string) {
  uni.navigateTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(String(familyId))}` })
}

function openShowcaseFamily(familyId: number | string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(String(familyId))}`
  })
}

function openInvitations() {
  if (!session.requireLogin('/pages/me/family-affairs?tab=invitations')) return
  uni.navigateTo({ url: '/pages/me/family-affairs?tab=invitations' })
}

function openCreateFamily() {
  if (!session.requireProfileComplete('/pages/family/create')) return
  uni.navigateTo({ url: '/pages/family/create' })
}

function familyRegionLabel(family: FamilySummary) {
  return family.regionText || family.nativePlace || '暂未填写地区'
}

function showcaseDesc(family: PublicFamilyShowcaseItem) {
  return family.description || family.regionText || family.nativePlace || '已公开展示'
}

function roleText(role: string) {
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    default:
      return role ? '未知角色' : '成员'
  }
}

onShow(refreshPage)
</script>

<style scoped>
.my-family-page {
  padding-top: 28rpx;
}

.my-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 22rpx;
}

.my-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.showcase-notice {
  margin-bottom: 18rpx;
  padding: 22rpx 0;
}

.showcase-notice-label {
  display: block;
  margin-bottom: 8rpx;
  color: var(--archive-cinnabar);
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 4rpx;
}

.showcase-notice-text {
  display: block;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.65;
}

.create-family-entry {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  margin-bottom: 18rpx;
  padding: 22rpx 0;
}

.create-family-copy {
  flex: 1;
  min-width: 0;
}

.create-family-title,
.create-family-desc {
  display: block;
}

.create-family-title {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 30rpx;
  font-weight: 700;
  letter-spacing: 1rpx;
}

.create-family-desc {
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.62;
}

.showcase-invite-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22rpx;
  border-top: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 16rpx 0;
}

.showcase-invite-label {
  color: var(--archive-blue);
  font-size: 24rpx;
  font-weight: 650;
}

.showcase-count {
  display: block;
  margin-bottom: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

.showcase-load-more {
  display: flex;
  justify-content: center;
  margin-top: 24rpx;
}

.showcase-load-more .disabled {
  opacity: 0.55;
}

.my-resolving-panel {
  margin-top: 28rpx;
}

.my-loading-lines {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-top: 26rpx;
}

.my-loading-line {
  width: 72%;
  height: 1rpx;
  background: linear-gradient(90deg, var(--archive-line-strong), rgba(92, 74, 48, 0.04));
}

.my-loading-line.long {
  width: 88%;
}

.my-loading-line.short {
  width: 46%;
}
</style>
