<template>
  <view class="page">
    <view class="card">
      <text class="title">我的家庭</text>
      <text class="muted">查看当前账号已创建或加入的家庭。</text>
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
      <button v-if="errorMessage" class="button secondary" @click="loadFamilies">重新加载</button>
    </view>
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载家庭列表...</text>
    </view>
    <view v-else-if="!errorMessage && families.length === 0" class="card state-card">
      <text class="section-title">暂无家庭</text>
      <text class="muted">当前账号还没有创建或加入家庭。你可以查看公开家庭，或通过家人发送的邀请加入家庭。</text>
      <button class="button" @click="openSearch">查看公开家庭</button>
      <button class="button secondary" @click="openInvitations">查看我的邀请</button>
    </view>
    <view
      v-for="family in families"
      :key="family.id"
      class="card family-card"
      @click="openFamily(family.id)"
    >
      <view class="section-row">
        <text class="section-title">{{ family.familyName }}</text>
        <text class="tag">{{ roleText(family.role) }}</text>
      </view>
      <text class="muted">姓氏：{{ family.familySurname }}</text>
      <text v-if="family.regionText || family.nativePlace" class="muted">
        地区：{{ family.regionText || family.nativePlace }}
      </text>
      <text class="muted">家庭状态：{{ familyStatusText(family.status) }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listMyFamilies } from '@/api/families'
import { useSessionStore } from '@/stores/session'
import type { FamilySummary } from '@/types/api'

const session = useSessionStore()
const families = ref<FamilySummary[]>([])
const loading = ref(false)
const errorMessage = ref('')

async function loadFamilies() {
  if (!session.requireLogin('/pages/family/my')) return
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

onShow(loadFamilies)
</script>

<style scoped>
.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.state-card {
  text-align: center;
}

.section-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.section-row .section-title {
  margin-bottom: 8rpx;
}
</style>
