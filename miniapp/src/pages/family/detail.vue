<template>
  <view class="archive-page detail-page">
    <MiniBackHome />
    <MiniFamilyPageSkeleton v-if="loading && !family" variant="detail" />

    <view v-else-if="errorMessage && !family" class="detail-error archive-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </view>

    <template v-else-if="family">
      <view class="detail-head">
        <view>
          <text class="archive-kicker">Family Archive</text>
          <text class="archive-title">家庭详情</text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="family-cover">
        <view class="family-cover-copy">
          <text class="cover-name">{{ family.familyName }}</text>
          <text class="cover-desc">{{ family.description || '暂未填写家庭简介。' }}</text>
        </view>
        <view class="cover-book archive-book-spine">
          <text>家</text>
          <text>庭</text>
          <text>树</text>
        </view>
      </view>

      <view class="dossier-grid">
        <view v-for="item in dossierItems" :key="item.label" class="dossier-cell">
          <text class="dossier-label">{{ item.label }}</text>
          <text class="dossier-value">{{ item.value }}</text>
        </view>
      </view>

      <view class="status-line">
        <text class="archive-chip">{{ familyStatusText(family.status) }}</text>
        <text class="archive-chip">{{ publicStatusText(family.publicDisplayStatus) }}</text>
      </view>

      <view class="detail-directory archive-list">
        <view class="archive-row" @click="openProfile">
          <view class="archive-row-main">
            <text class="archive-row-title">家族简介 / 家庭档案</text>
            <text class="archive-row-desc">姓氏、地区、简介与公开状态</text>
          </view>
          <text class="archive-row-meta">档案</text>
          <text class="archive-arrow">›</text>
        </view>
        <navigator class="archive-row" :url="familyTreeUrl" hover-class="none">
          <view class="archive-row-main">
            <text class="archive-row-title">家庭树</text>
            <text class="archive-row-desc">在家庭树与成员表之间切换查看</text>
          </view>
          <text class="archive-row-meta">家庭树 / 成员表</text>
          <text class="archive-arrow">›</text>
        </navigator>
        <view v-if="canManageFamily" class="archive-row" @click="openManageCenter">
          <view class="archive-row-main">
            <text class="archive-row-title">家庭管理</text>
            <text class="archive-row-desc">加入申请、邀请与公开展示权限</text>
          </view>
          <text class="archive-row-meta">管理</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyMember } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const loading = ref(false)
const errorMessage = ref('')
const navigating = ref(false)

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const familyTreeUrl = computed(() =>
  `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
)
const lastMemberUpdatedAtText = computed(() => formatLatestMemberUpdate(members.value))

const dossierItems = computed(() => {
  const currentFamily = family.value
  if (!currentFamily) return []
  return [
    { label: '姓氏', value: `${currentFamily.familySurname || '未录'}氏` },
    { label: '成员数', value: `${members.value.length} 位` },
    { label: '地区', value: currentFamily.regionText || currentFamily.nativePlace || '未填写' },
    { label: '最近更新', value: lastMemberUpdatedAtText.value }
  ]
})

async function loadFamily() {
  session.restoreSession()
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = !family.value
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    const [familyDetail, familyMembers] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyMembers(familyId.value)
    ])
    family.value = familyDetail
    members.value = familyMembers
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭详情加载失败。')
  } finally {
    loading.value = false
  }
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  family.value = null
  members.value = []
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

function openProfile() {
  navigateOnce(`/pages/family/profile?familyId=${encodeURIComponent(familyId.value)}`)
}

function openManageCenter() {
  navigateOnce(`/pages/family/manage-center?familyId=${encodeURIComponent(familyId.value)}`)
}

function familyStatusText(status: string) {
  switch (status) {
    case 'NORMAL':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'DISSOLVED':
      return '已解散'
    case 'DISSOLUTION_COOLDOWN':
      return '恢复冷静期'
    case 'DISSOLUTION_PENDING':
      return '解散待审核'
    case 'PENDING':
      return '待审核'
    case 'APPROVED':
      return '已通过'
    case 'REJECTED':
      return '已拒绝'
    case 'TAKEN_DOWN':
      return '已下架'
    case 'DELETED':
      return '已删除'
    default:
      return '未知状态'
  }
}

function formatLatestMemberUpdate(memberList: FamilyMember[]) {
  const latestTime = memberList
    .map((member) => Date.parse(member.updatedAt || member.createdAt || ''))
    .filter((time) => Number.isFinite(time))
    .sort((a, b) => b - a)[0]
  if (!latestTime) return '暂无记录'
  return new Date(latestTime).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  })
}

function publicStatusText(status: string) {
  switch (status) {
    case 'APPROVED':
      return '已公开'
    case 'PENDING':
      return '待审核'
    case 'REJECTED':
      return '已拒绝'
    case 'PRIVATE':
      return '未公开'
    case 'TAKEN_DOWN':
      return '已下架'
    default:
      return '未知状态'
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
})
onShow(loadFamily)
onUnload(resetPageData)
</script>

<style scoped>
.detail-page {
  padding-top: 28rpx;
}

.detail-page :deep(.mini-back-home) {
  margin-bottom: 24rpx;
}

.detail-error {
  padding: 28rpx 0;
}

.detail-error :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24rpx;
  margin-bottom: 28rpx;
}

.family-cover {
  display: flex;
  min-height: 210rpx;
  margin-bottom: 30rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.family-cover-copy {
  flex: 1;
  min-width: 0;
  padding: 28rpx 28rpx 28rpx 0;
}

.cover-name,
.cover-desc {
  display: block;
}

.cover-name {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 42rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  line-height: 1.28;
}

.cover-desc {
  margin-top: 14rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.7;
}

.cover-book {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  width: 118rpx;
  flex-shrink: 0;
}

.cover-book text {
  color: rgba(255, 255, 255, 0.94);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 800;
}

.dossier-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border-top: 1rpx solid var(--archive-line);
  border-left: 1rpx solid var(--archive-line);
  margin-bottom: 20rpx;
}

.dossier-cell {
  min-height: 112rpx;
  border-right: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 20rpx;
  box-sizing: border-box;
}

.dossier-label,
.dossier-value {
  display: block;
}

.dossier-label {
  color: var(--archive-cinnabar);
  font-size: 21rpx;
  line-height: 1.2;
}

.dossier-value {
  margin-top: 12rpx;
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 700;
  line-height: 1.35;
}

.status-line {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 28rpx;
}

.detail-directory {
  margin-top: 8rpx;
}
</style>
