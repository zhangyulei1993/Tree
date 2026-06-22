<template>
  <view class="tree-page">
    <MiniCard v-if="loading">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载家庭详情..." />
    </MiniCard>

    <MiniCard v-else-if="errorMessage">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </MiniCard>

    <template v-else-if="family">
      <MiniCard variant="hero" class="dossier-hero tree-pedigree-watermark">
        <view class="tree-dossier-head">
          <view class="tree-dossier-seal">{{ family.familySurname.slice(0, 1) }}</view>
          <view class="tree-dossier-copy">
            <text class="dossier-eyebrow">家族档案</text>
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

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">档案信息</text>
          <text class="tree-space-subtitle">家庭基本资料与公开展示状态</text>
        </view>
        <view class="tree-space-body section-pad">
          <view class="tree-info-row">
            <text class="tree-info-label">姓氏</text>
            <text class="tree-info-value">{{ family.familySurname }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">籍贯</text>
            <text class="tree-info-value">{{ family.nativePlace || '未设置' }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">地区</text>
            <text class="tree-info-value">{{ family.regionText || '未设置' }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">公开状态</text>
            <text class="tree-info-value">{{ publicStatusText(family.publicDisplayStatus) }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">允许被搜索</text>
            <text class="tree-info-value">{{ family.searchable ? '是' : '否' }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">公开联系方式</text>
            <text class="tree-info-value">{{ publicContactText }}</text>
          </view>
        </view>
      </view>

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">我的身份</text>
          <text class="tree-space-subtitle">你在该家庭中的角色</text>
        </view>
        <view class="tree-space-body section-pad">
          <view class="tree-info-row">
            <text class="tree-info-label">家庭角色</text>
            <text class="tree-info-value">{{ roleText(family.role) }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">家庭状态</text>
            <text class="tree-info-value">{{ familyStatusText(family.status) }}</text>
          </view>
        </view>
      </view>

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">家族操作</text>
          <text class="tree-space-subtitle">成员管理与家谱查看</text>
        </view>
        <view class="tree-space-body">
          <view class="tree-action-tile" @click="openMembers">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-users" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">成员列表</text>
              <text class="tree-action-tile-desc">查看与管理家庭成员</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
          <view class="tree-action-tile" @click="openTree">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-folder" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">私有家谱</text>
              <text class="tree-action-tile-desc">查看成员与亲属关系</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
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
const publicContactText = computed(() => {
  if (!family.value?.publicContactVisible) return '未公开'
  const contacts = [
    family.value.publicContactName,
    family.value.publicContactPhone,
    family.value.publicContactWechat,
    family.value.publicContactNote
  ].filter(Boolean)
  return contacts.length > 0 ? contacts.join(' / ') : '未设置'
})

async function loadFamily() {
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  errorMessage.value = ''
  try {
    family.value = await getFamilyDetail(familyId.value)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭详情加载失败。')
  } finally {
    loading.value = false
  }
}

function openMembers() {
  uni.navigateTo({
    url: `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function openTree() {
  uni.navigateTo({
    url: `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function roleText(role: string) {
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    case 'ROOT_ADMIN':
      return '平台管理员'
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
  loadFamily()
})
</script>

<style scoped>
.dossier-hero {
  margin-bottom: 20rpx;
}

.dossier-eyebrow {
  display: block;
  margin-bottom: 8rpx;
  color: var(--tree-gold-text);
  font-size: 22rpx;
  font-weight: 600;
  letter-spacing: 2rpx;
}

.dossier-desc {
  display: block;
  margin-top: 16rpx;
  line-height: 1.7;
}

.section-pad {
  padding: 8rpx 24rpx 16rpx;
}

.entry-panel {
  margin-top: 4rpx;
}

.feature-arrow {
  color: var(--tree-text-weak);
  font-size: 28rpx;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 12rpx;
}
</style>
