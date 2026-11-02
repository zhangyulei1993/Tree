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
        <text class="archive-chip ui-simplified-hidden">创建 {{ ownedFamilies.length }} · 加入 {{ joinedFamilies.length }}</text>
        <text class="archive-thin-button" @click="openCreateFamily">创建家庭</text>
      </view>

      <view v-if="errorMessage" class="my-state archive-form-panel">
        <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
        <MiniButton variant="secondary" @click="loadMyFamilies">重新加载</MiniButton>
      </view>

      <view v-else-if="loadingMyFamilies" class="my-state archive-form-panel">
        <MiniEmptyState symbol="…" title="正在加载" description="正在加载家庭列表..." />
      </view>

      <view v-else class="my-family-groups">
        <view v-if="ownedFamilies.length > 0" class="archive-form-panel">
          <view class="archive-section-head">
            <text class="archive-section-title">我创建的家庭</text>
            <text class="archive-section-subtitle">共 {{ ownedFamilies.length }} 个</text>
          </view>
          <view class="my-family-list archive-list">
            <view
              v-for="family in ownedFamilies"
              :key="family.id"
              class="archive-row"
              @click="openFamily(family.id)"
            >
              <text class="archive-surname-stamp">{{ family.familySurname.slice(0, 1) }}</text>
              <view class="archive-row-main">
                <text class="archive-row-title">{{ family.familyName }}</text>
                <text class="archive-row-desc">{{ familyDesc(family) }}</text>
              </view>
              <text class="archive-row-meta" :class="{ 'is-cooldown': family.status === 'DISSOLUTION_COOLDOWN' }">
                {{ familyRowStatus(family) }}
              </text>
              <text class="archive-arrow">›</text>
            </view>
          </view>
        </view>

        <view v-if="joinedFamilies.length > 0" class="archive-form-panel">
          <view class="archive-section-head">
            <text class="archive-section-title">我加入的家庭</text>
            <text class="archive-section-subtitle">共 {{ joinedFamilies.length }} 个</text>
          </view>
          <view class="my-family-list archive-list">
            <view
              v-for="family in joinedFamilies"
              :key="family.id"
              class="archive-row"
              @click="openFamily(family.id)"
            >
              <text class="archive-surname-stamp">{{ family.familySurname.slice(0, 1) }}</text>
              <view class="archive-row-main">
                <text class="archive-row-title">{{ family.familyName }}</text>
                <text class="archive-row-desc">{{ familyDesc(family) }}</text>
              </view>
              <text class="archive-row-meta" :class="{ 'is-cooldown': family.status === 'DISSOLUTION_COOLDOWN' }">
                {{ familyRowStatus(family) }}
              </text>
              <text class="archive-arrow">›</text>
            </view>
          </view>
        </view>
      </view>
    </template>

    <template v-else>
      <view class="my-head archive-page-head">
        <view>
          <text class="archive-kicker">My Archive</text>
          <text class="archive-title">我的家庭</text>
          <text class="archive-subtitle">还没有家庭，可创建或等待邀请</text>
        </view>
        <view class="archive-seal">家</view>
      </view>

      <view class="my-empty-actions">
        <view class="my-empty-action my-empty-action--quality" @click="openShowcase">
          <text class="my-empty-action-mark">优</text>
          <text class="my-empty-action-title">优质家庭</text>
          <text class="my-empty-action-desc">学习真实家庭树案例</text>
          <text class="my-empty-action-link">去看看 ›</text>
        </view>
        <view class="my-empty-action my-empty-action--create" @click="openCreateFamily">
          <text class="my-empty-action-mark">建</text>
          <text class="my-empty-action-title">创建家庭</text>
          <text class="my-empty-action-desc">建立自己的家庭树</text>
          <text class="my-empty-action-link">{{ createFamilyActionText }} ›</text>
        </view>
      </view>

      <view v-if="session.isLoggedIn" class="showcase-invite-link" @click="openInvitations">
        <text class="showcase-invite-label">查看收到的家庭邀请</text>
        <text class="archive-arrow">›</text>
      </view>

      <template v-if="customerPublicShowcaseAvailable">
      <view v-if="showcaseError" class="my-state archive-form-panel">
        <MiniNotice tone="warm" title="加载失败">{{ showcaseError }}</MiniNotice>
        <MiniButton variant="secondary" @click="reloadShowcase">重新加载</MiniButton>
      </view>

      <view v-else-if="showcaseLoading && showcaseItems.length === 0" class="my-state archive-form-panel">
        <MiniEmptyState symbol="展" title="正在加载" description="正在读取优质家庭列表..." />
      </view>

      <template v-else>
        <view v-if="showcaseItems.length > 0" class="showcase-list archive-list">
          <text class="showcase-count">共 {{ showcaseTotal }} 个优质家庭</text>
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
            title="暂无优质家庭"
            description="暂无可供学习的优质家庭案例。"
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
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { apiErrorMessage } from '@/api/client'
import { formatFamilyRegionText } from '@/features/native-place/geoHierarchy'
import { fetchCapabilities } from '@/api/capabilities'
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
const navigating = ref(false)

const showcaseItems = ref<PublicFamilyShowcaseItem[]>([])
const showcaseTotal = ref(0)
const showcasePage = ref(1)
const showcaseLoading = ref(false)
const showcaseLoadingMore = ref(false)
const showcaseError = ref('')
const showcasePageSize = 20
const customerPublicShowcaseAvailable = false

const hasMyFamilies = computed(() => myFamiliesLoaded.value && families.value.length > 0)
const isResolvingMyFamilyMode = computed(() => loadingMyFamilies.value && !myFamiliesLoaded.value)
const showcaseHasMore = computed(() => showcaseItems.value.length < showcaseTotal.value)
const ownedFamilies = computed(() => families.value.filter((family) => family.role === 'FOUNDER'))
const joinedFamilies = computed(() => families.value.filter((family) => family.role !== 'FOUNDER'))
const createFamilyActionText = computed(() => (session.isLoggedIn ? '创建家庭' : '登录后创建'))
const createFamilyHint = computed(() =>
  session.isLoggedIn
    ? session.isProfileComplete
      ? '准备记录家庭关系？填写姓氏与地区，建立新的家庭树。'
      : '准备记录家庭关系？完善资料后即可创建家庭。'
    : '准备记录家庭关系？登录并完善资料后可创建家庭。'
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
    showcaseError.value = apiErrorMessage(error, '优质家庭列表加载失败。')
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
  if (customerPublicShowcaseAvailable && !hasMyFamilies.value) {
    await fetchShowcasePage(1, false)
  }
}

function openFamily(familyId: number | string) {
  navigateOnce(`/pages/family/detail?familyId=${encodeURIComponent(String(familyId))}`)
}

function openShowcaseFamily(familyId: number | string) {
  navigateOnce(`/pages/family/public-profile?familyId=${encodeURIComponent(String(familyId))}`)
}

function openShowcase() {
  navigateOnce('/pages/family/showcase')
}

function openInvitations() {
  if (!session.requireLogin('/pages/invite/my')) return
  navigateOnce('/pages/invite/my')
}

function navigateOnce(url: string) {
  if (navigating.value) return
  navigating.value = true
  uni.navigateTo({
    url,
    complete: () => {
      navigating.value = false
    }
  })
}

async function openCreateFamily() {
  if (!session.requireProfileComplete('/pages/family/create')) return
  try {
    const capabilities = await fetchCapabilities()
    if (capabilities.usage.ownedFamilies >= capabilities.limits.maxOwnedFamilies) {
      uni.showModal({
        title: '暂不能创建家庭',
        content: `当前账号可创建家庭 ${capabilities.usage.ownedFamilies}/${capabilities.limits.maxOwnedFamilies}。冷静期家庭仍会占用创建名额，完成解散后才会释放。`,
        confirmText: '知道了',
        showCancel: false
      })
      return
    }
  } catch (error) {
    uni.showToast({ title: apiErrorMessage(error, '无法确认创建资格'), icon: 'none' })
    return
  }
  navigateOnce('/pages/family/create')
}

function familyRegionLabel(family: FamilySummary) {
  return formatFamilyRegionText(family)
}

function familyDesc(family: FamilySummary) {
  const region = familyRegionLabel(family)
  if (family.status !== 'DISSOLUTION_COOLDOWN') return region
  if (!family.dissolutionCooldownUntil) return `${region} · 冷静期中`
  return `${region} · 截止 ${formatCooldown(family.dissolutionCooldownUntil)}`
}

function familyRowStatus(family: FamilySummary) {
  if (family.status === 'DISSOLUTION_COOLDOWN') return '冷静期'
  if (family.status === 'DISSOLUTION_PENDING') return '待审核'
  return '进入'
}

function formatCooldown(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '冷静期中'
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  const hour = `${date.getHours()}`.padStart(2, '0')
  const minute = `${date.getMinutes()}`.padStart(2, '0')
  return `${month}-${day} ${hour}:${minute}`
}

function showcaseDesc(family: PublicFamilyShowcaseItem) {
  return family.description || family.regionText || family.nativePlace || '已入选优质家庭'
}

onShow(refreshPage)
</script>

<style scoped>
.ui-simplified-hidden {
  display: none !important;
}

.my-empty-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14rpx;
  margin-bottom: 18rpx;
}

.my-empty-action {
  display: flex;
  min-height: 190rpx;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-start;
  border: 1rpx solid var(--archive-line);
  border-radius: 10rpx;
  background: rgba(255, 252, 244, 0.64);
  padding: 18rpx;
  box-shadow: 0 6rpx 16rpx rgba(91, 70, 40, 0.04);
}

.my-empty-action--quality {
  border-color: rgba(47, 107, 87, 0.24);
  background: linear-gradient(180deg, rgba(241, 249, 243, 0.88), rgba(255, 252, 244, 0.54));
}

.my-empty-action-mark {
  display: grid;
  width: 48rpx;
  height: 48rpx;
  place-items: center;
  border: 2rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 25rpx;
  transform: rotate(-6deg);
}

.my-empty-action--quality .my-empty-action-mark {
  border-color: var(--archive-green, #2f6b57);
  color: var(--archive-green, #2f6b57);
}

.my-empty-action-title {
  margin-top: 14rpx;
  color: var(--archive-ink);
  font-size: 29rpx;
  font-weight: 760;
}

.my-empty-action-desc {
  margin-top: 7rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.45;
}

.my-empty-action-link {
  margin-top: auto;
  color: var(--archive-blue);
  font-size: 22rpx;
  font-weight: 650;
}

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

.my-family-groups {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.archive-row-meta.is-cooldown {
  color: var(--archive-cinnabar);
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
