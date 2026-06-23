<template>
  <view class="tree-page family-space-page">
    <MiniBackHome />
    <view class="tree-space tree-pedigree-watermark">
      <view class="tree-space-head">
        <text class="tree-space-title">我的家庭</text>
        <text class="tree-space-subtitle">你创建或加入的家族空间</text>
        <view v-if="!loading && !errorMessage && families.length > 0" class="tree-archive-ribbon">
          <text class="tree-archive-chip">共 {{ families.length }} 个家庭</text>
        </view>
      </view>
      <view class="tree-space-body archive-body">
        <MiniCard v-if="errorMessage" flat class="state-card">
          <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
          <MiniButton variant="secondary" @click="loadFamilies">重新加载</MiniButton>
        </MiniCard>

        <MiniCard v-else-if="loading" flat class="state-card">
          <MiniEmptyState title="正在加载" description="正在加载家庭列表..." />
        </MiniCard>

        <MiniCard v-else-if="families.length === 0" flat class="state-card">
          <MiniEmptyState
            title="暂无家庭"
            description="当前账号还没有创建或加入家庭。你可以查看公开家庭，或通过家人发送的邀请加入家庭。"
            action-text="查看公开家庭"
            @action="openSearch"
          />
          <MiniButton variant="secondary" @click="openInvitations">查看我的邀请</MiniButton>
        </MiniCard>

        <view v-else class="family-stack">
          <FamilyMiniCard
            v-for="family in families"
            :key="family.id"
            :name="family.familyName"
            :surname="family.familySurname"
            :region="family.regionText || family.nativePlace || undefined"
            :role-label="roleText(family.role)"
            :status-label="familyStatusText(family.status)"
            :status-tone="familyStatusTone(family.status)"
            @click="openFamily(family.id)"
          />
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listMyFamilies } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { statusTagTone } from '@/components/base/formatStatus'
import FamilyMiniCard from '@/components/family/FamilyMiniCard.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilySummary } from '@/types/api'

const session = useSessionStore()
const families = ref<FamilySummary[]>([])
const loading = ref(false)
const errorMessage = ref('')

async function loadFamilies() {
  if (!session.requirePhoneBound('/pages/family/my')) return
  loading.value = true
  errorMessage.value = ''
  try {
    families.value = await listMyFamilies()
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭列表加载失败。')
  } finally {
    loading.value = false
  }
}

function openFamily(familyId: number | string) {
  uni.navigateTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(String(familyId))}` })
}

function openSearch() {
  uni.switchTab({ url: '/pages/family/search' })
}

function openInvitations() {
  uni.navigateTo({ url: '/pages/invite/my' })
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

function familyStatusTone(status: string) {
  return statusTagTone(status)
}

onShow(loadFamilies)
</script>

<style scoped>
.archive-body {
  padding: 16rpx 20rpx 20rpx;
}

.state-card {
  margin-bottom: 0;
}

.family-stack {
  padding: 0;
}
</style>
