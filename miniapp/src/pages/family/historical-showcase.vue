<template>
  <view class="archive-page historical-showcase-page">
    <view class="archive-page-head">
      <view>
        <text class="archive-kicker">Imperial Lineage</text>
        <text class="archive-title">帝王世家</text>
        <text class="archive-subtitle">历史示例，按后台发布内容展示，仅供浏览</text>
      </view>
      <view class="archive-seal">谱</view>
    </view>

    <view v-if="loading && items.length === 0" class="archive-form-panel page-state"><MiniEmptyState symbol="谱" title="正在加载" description="正在读取展示家谱，请稍候。" /></view>
    <view v-else-if="errorMessage && items.length === 0" class="archive-form-panel page-state"><MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice><MiniButton variant="secondary" @click="load">重新加载</MiniButton></view>
    <view v-else-if="items.length === 0" class="archive-form-panel page-state">
      <MiniEmptyState symbol="谱" title="示例整理中" description="当前暂无可浏览的历史示例，不影响继续创建自己的家谱。" />
      <view v-if="guideMode" class="guide-next-panel empty-guide-next" @click="openGuideNext">
        <view>
          <text class="guide-next-kicker">NEXT STEP</text>
          <text class="guide-next-title">{{ guideNextTitle }}</text>
          <text class="guide-next-desc">{{ guideNextDesc }}</text>
        </view>
        <MiniButton size="sm" class="guide-next-action" @click="openGuideNext">{{ guideNextAction }}</MiniButton>
      </view>
    </view>

    <template v-else>
      <MiniNotice v-if="switchNotice" tone="warm" title="已切换展示">
        {{ switchNotice }}
      </MiniNotice>
      <view v-if="!guideMode" class="historical-list archive-panel">
        <text class="historical-list-label">已发布世系</text>
        <view v-for="item in items" :key="item.id" class="historical-row" :class="{ active: item.id === selectedId }" @click="selectFamily(item.id)">
          <view><text class="historical-name">{{ item.displayName }}</text><text class="historical-meta">{{ item.regionText || '历史示例' }} · {{ item.memberCount }} 位人物</text></view>
          <text class="archive-arrow">›</text>
        </view>
      </view>

      <view v-if="selected" class="historical-intro archive-panel"><text class="historical-label">历史示例</text><text class="historical-summary">{{ selected.summary }}</text><text class="historical-count">{{ selected.memberCount }} 位人物 · {{ selected.relationshipCount }} 条关系</text></view>
      <view v-if="guideMode && tree" class="guide-next-panel" @click="openGuideNext">
        <view>
          <text class="guide-next-kicker">NEXT STEP</text>
          <text class="guide-next-title">{{ guideNextTitle }}</text>
          <text class="guide-next-desc">{{ guideNextDesc }}</text>
        </view>
        <MiniButton size="sm" class="guide-next-action" @click="openGuideNext">{{ guideNextAction }}</MiniButton>
      </view>
      <view v-if="treeLoading" class="archive-form-panel page-state"><MiniEmptyState symbol="…" title="正在展开家谱" description="正在整理人物关系。" /></view>
      <view v-else-if="tree" class="historical-tree"><view class="tree-note"><text>可左右、上下滑动查看完整世系；人物仅展示姓名与称号/封号。</text></view><FamilyTreeStructureView :tree="tree" portrait-theme="imperial" /></view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { listMyFamilies } from '@/api/families'
import { getPublishedShowcaseFamilyTree, listPublishedShowcaseFamilies } from '@/api/showcaseFamilies'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { useSessionStore } from '@/stores/session'
import type { FamilySummary, FamilyTreeResult, ShowcaseFamilyItem } from '@/types/api'

const session = useSessionStore()
const items = ref<ShowcaseFamilyItem[]>([])
const myFamilies = ref<FamilySummary[]>([])
const selectedId = ref<number | null>(null)
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const treeLoading = ref(false)
const errorMessage = ref('')
const switchNotice = ref('')
const guideMode = ref(false)
const selected = computed(() => items.value.find((item) => item.id === selectedId.value) || null)
const hasOwnFamilies = computed(() => myFamilies.value.length > 0)
const guideNextTitle = computed(() => {
  if (hasOwnFamilies.value) return '继续完善家谱'
  if (!session.isLoggedIn) return '示例看完了，登录后继续建谱'
  if (!session.isProfileComplete) return '继续完善资料'
  return '创建自己的第一本家谱'
})
const guideNextDesc = computed(() => {
  if (hasOwnFamilies.value) return '先补充成员和关系，再邀请家人一起核对。'
  if (!session.isLoggedIn) return '登录完成后会回到当前示例页，再继续下一步。'
  if (!session.isProfileComplete) return '补充昵称后，再创建家庭树。'
  return '从姓氏和家庭名称开始，搭起自己的家族关系树。'
})
const guideNextAction = computed(() => {
  if (hasOwnFamilies.value) return '补充成员'
  if (!session.isLoggedIn) return '登录继续'
  if (!session.isProfileComplete) return '完善资料'
  return '创建家谱'
})

async function refreshMyFamilies() {
  session.restoreSession()
  if (!session.isLoggedIn || !session.isProfileComplete) {
    myFamilies.value = []
    return
  }
  try {
    myFamilies.value = await listMyFamilies()
  } catch {
    myFamilies.value = []
  }
}

async function selectFamily(id: number) {
  selectedId.value = id
  tree.value = null
  treeLoading.value = true
  try {
    tree.value = await getPublishedShowcaseFamilyTree(id)
    if (guideMode.value) {
      markNewUserGuideMilestone('SHOWCASE_VIEWED')
    }
  } catch (error) {
    const fallback = items.value.find((item) => item.id !== id)
    if (fallback) {
      switchNotice.value = '所选世系正在修订，已自动切换到其他已发布世系。'
      items.value = items.value.filter((item) => item.id !== id)
      await selectFamily(fallback.id)
    } else {
      errorMessage.value = apiErrorMessage(error, '展示家谱暂不可访问。')
    }
  } finally {
    treeLoading.value = false
  }
}

async function load(preferredId?: number) {
  loading.value = true
  errorMessage.value = ''
  switchNotice.value = ''
  try {
    items.value = await listPublishedShowcaseFamilies()
    const preferred = items.value.find((item) => item.id === preferredId)?.id
    const next = preferred || items.value.find((item) => item.rotationPinned)?.id || items.value[0]?.id
    if (preferredId && !preferred && next) {
      switchNotice.value = '所选世系正在修订，已自动切换到其他已发布世系。'
    }
    if (next) await selectFamily(next)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '展示家谱列表加载失败。')
  } finally {
    loading.value = false
  }
}

function currentGuideRoute() {
  const params = ['guide=1']
  if (selectedId.value) params.unshift(`familyId=${encodeURIComponent(String(selectedId.value))}`)
  return `/pages/family/historical-showcase?${params.join('&')}`
}

function openGuideNext() {
  markNewUserGuideMilestone('SHOWCASE_VIEWED')
  const route = currentGuideRoute()
  if (!session.isLoggedIn) {
    session.requireLogin(route)
    return
  }
  if (!session.isProfileComplete) {
    session.requireProfileComplete(route)
    return
  }
  if (hasOwnFamilies.value) {
    if (myFamilies.value.length === 1) {
      const family = myFamilies.value[0]
      const canManage = family.role === 'FOUNDER' || family.role === 'FAMILY_ADMIN'
      uni.navigateTo({
        url: canManage
          ? `/pages/family/manage?familyId=${encodeURIComponent(String(family.id))}&section=node&guide=1`
          : `/pages/family/private-tree?familyId=${encodeURIComponent(String(family.id))}&guide=1`
      })
      return
    }
    uni.switchTab({ url: '/pages/family/my' })
    return
  }
  uni.navigateTo({ url: '/pages/family/create' })
}

let skipInitialShow = true

onLoad((options) => {
  const id = Number(options?.familyId)
  guideMode.value = String(options?.guide || '') === '1'
  void refreshMyFamilies()
  void load(Number.isFinite(id) && id > 0 ? id : undefined)
})

onShow(() => {
  if (skipInitialShow) { skipInitialShow = false; return }
  void refreshMyFamilies()
})
</script>

<style scoped>
.historical-showcase-page { padding-top: 28rpx; padding-bottom: calc(80rpx + env(safe-area-inset-bottom)); }
.page-state { margin-top: 24rpx; }
.historical-list, .historical-intro, .historical-tree { margin-top: 20rpx; }
.historical-list { padding: 18rpx 0; }
.historical-list-label, .historical-label { display: block; color: var(--archive-cinnabar); font-size: 20rpx; font-weight: 700; letter-spacing: 3rpx; }
.historical-row { display: flex; align-items: center; justify-content: space-between; gap: 14rpx; border-top: 1rpx solid var(--archive-line); padding: 18rpx 0; }
.historical-row:first-of-type { margin-top: 12rpx; }
.historical-row.active .historical-name { color: var(--archive-cinnabar); }
.historical-name, .historical-meta, .historical-summary, .historical-count { display: block; }
.historical-name { color: var(--archive-ink); font-size: 27rpx; font-weight: 700; }
.historical-meta, .historical-count { margin-top: 5rpx; color: var(--archive-ink-soft); font-size: 21rpx; }
.historical-intro { padding: 20rpx 0; }
.historical-summary { margin-top: 12rpx; color: var(--archive-ink); font-size: 25rpx; line-height: 1.7; }
.tree-note { margin-bottom: 12rpx; color: var(--archive-ink-soft); font-size: 21rpx; line-height: 1.5; }
.guide-next-panel { display:flex; align-items:center; justify-content:space-between; gap:18rpx; margin-top:22rpx; border:1rpx solid rgba(168,59,45,.28); border-left:6rpx solid var(--archive-cinnabar); border-radius:10rpx; background:linear-gradient(135deg, rgba(168,59,45,.1), transparent 46%), rgba(255,252,244,.78); padding:22rpx 18rpx; box-shadow:0 10rpx 24rpx rgba(91,70,40,.06); }
.guide-next-kicker, .guide-next-title, .guide-next-desc { display:block; }
.guide-next-kicker { color:var(--archive-cinnabar); font-size:18rpx; font-weight:800; letter-spacing:4rpx; line-height:1.2; }
.guide-next-title { margin-top:6rpx; color:var(--archive-ink); font-size:32rpx; font-weight:800; line-height:1.25; }
.guide-next-desc { margin-top:5rpx; color:var(--archive-ink-soft); font-size:24rpx; line-height:1.35; }
.guide-next-action { flex-shrink:0; }
</style>
