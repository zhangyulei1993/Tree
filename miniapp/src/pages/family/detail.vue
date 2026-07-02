<template>
  <view class="tree-page detail-page">
    <MiniBackHome />
    <MiniFamilyPageSkeleton v-if="loading && !family" variant="detail" />

    <MiniCard v-else-if="errorMessage && !family">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </MiniCard>

    <template v-else-if="family">
      <MiniCard variant="hero" class="dossier-hero tree-pedigree-watermark">
        <view class="tree-dossier-head">
          <view class="tree-dossier-seal">{{ family.familySurname.slice(0, 1) }}</view>
          <view class="tree-dossier-copy">
            <text class="dossier-eyebrow">家庭空间</text>
            <text class="tree-page-title">{{ family.familyName }}</text>
            <view class="tag-row">
              <MiniStatusTag :label="familyStatusText(family.status)" :tone="statusTagTone(family.status)" />
              <MiniStatusTag
                :label="publicStatusText(family.publicDisplayStatus)"
                :tone="statusTagTone(family.publicDisplayStatus)"
              />
            </view>
          </view>
        </view>
        <text class="tree-muted dossier-desc">{{ family.description || '暂未填写家庭简介。' }}</text>
        <view class="tree-archive-ribbon">
          <text v-if="family.familySurname" class="tree-archive-chip gold">{{ family.familySurname }}氏</text>
          <text v-if="family.regionText || family.nativePlace" class="tree-archive-chip">
            {{ family.regionText || family.nativePlace }}
          </text>
        </view>
      </MiniCard>

      <MiniCard variant="soft" class="identity-card">
        <view class="identity-row">
          <text class="identity-label">我的角色</text>
          <text class="identity-value">{{ roleText(family.role) }}</text>
        </view>
        <view class="identity-row">
          <text class="identity-label">家庭状态</text>
          <text class="identity-value">{{ familyStatusText(family.status) }}</text>
        </view>
      </MiniCard>

      <MiniCard class="directory-card directory-card--list">
        <MiniDirectoryTile
          title="家庭档案"
          desc="姓氏、地区、简介与公开状态"
          icon-class="tree-symbol-doc"
          @click="openProfile"
        />
        <MiniDirectoryTile
          title="成员名册"
          desc="查看成员基本信息与绑定状态"
          icon-class="tree-symbol-users"
          @click="openMembers"
        />
        <MiniDirectoryTile
          title="家谱"
          desc="查看父母子女与配偶关系"
          icon-class="tree-symbol-folder"
          @click="openTree"
        />
        <MiniDirectoryTile
          v-if="canManageFamily"
          title="家庭管理"
          desc="加入申请、邀请与公开展示权限"
          icon-class="tree-symbol-tool"
          @click="openManageCenter"
        />
      </MiniCard>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniDirectoryTile from '@/components/base/MiniDirectoryTile.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniStatusTag from '@/components/base/MiniStatusTag.vue'
import { statusTagTone } from '@/components/base/formatStatus'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const navigating = ref(false)

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

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
    family.value = await getFamilyDetail(familyId.value)
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

function openMembers() {
  navigateOnce(`/pages/family/members?familyId=${encodeURIComponent(familyId.value)}`)
}

function openTree() {
  navigateOnce(`/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`)
}

function openManageCenter() {
  navigateOnce(`/pages/family/manage-center?familyId=${encodeURIComponent(familyId.value)}`)
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

function familyStatusText(status: string) {
  switch (status) {
    case 'NORMAL':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'DISSOLVED':
      return '已解散'
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
  min-height: auto;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 20%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.dossier-hero {
  margin-bottom: 18rpx;
  padding-top: 34rpx;
  padding-bottom: 34rpx;
}

.dossier-hero :deep(.tree-page-title),
.dossier-hero .tree-page-title {
  color: #fff;
}

.dossier-eyebrow {
  display: block;
  margin-bottom: 8rpx;
  color: rgba(248, 231, 194, 0.92);
  font-size: 22rpx;
  font-weight: 600;
  letter-spacing: 2rpx;
}

.dossier-desc {
  display: block;
  margin-top: 16rpx;
  color: rgba(255, 255, 255, 0.76);
  line-height: 1.7;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 12rpx;
}

.identity-card {
  margin-bottom: 16rpx;
}

.identity-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 12rpx 0;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.12);
}

.identity-row:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.identity-label {
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}

.identity-value {
  color: var(--tree-text-primary);
  font-size: 26rpx;
  font-weight: 700;
}

.directory-card {
  margin-bottom: 12rpx;
}

.directory-card--list {
  padding-top: 8rpx;
  padding-bottom: 8rpx;
}
</style>
