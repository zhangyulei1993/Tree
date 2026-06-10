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
      <text class="muted">当前账号还没有加入家庭。</text>
    </view>
    <view
      v-for="family in families"
      :key="family.id"
      class="card family-card"
      @click="openFamily(family.id)"
    >
      <view class="section-row">
        <text class="section-title">{{ family.familyName }}</text>
        <text class="tag">{{ family.role }}</text>
      </view>
      <text class="muted">姓氏：{{ family.familySurname }}</text>
      <text v-if="family.regionText || family.nativePlace" class="muted">
        地区：{{ family.regionText || family.nativePlace }}
      </text>
      <text class="muted">家庭状态：{{ family.status }}</text>
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

.family-card {
  cursor: pointer;
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
