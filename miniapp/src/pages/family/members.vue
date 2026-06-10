<template>
  <view class="page">
    <view class="card">
      <text class="title">家庭成员</text>
      <text class="muted">家庭 ID：{{ familyId || '-' }}</text>
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
      <button v-if="errorMessage" class="button secondary" @click="loadMembers">重新加载</button>
    </view>
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载成员...</text>
    </view>
    <view v-else-if="!errorMessage && members.length === 0" class="card state-card">
      <text class="section-title">暂无成员</text>
    </view>
    <view v-for="member in members" :key="member.memberId" class="card">
      <view class="section-row">
        <text class="section-title">{{ member.name }}</text>
        <text class="tag">{{ member.boundFamilyRole || '未绑定角色' }}</text>
      </view>
      <view class="info-row"><text class="label">成员 ID</text><text>{{ member.memberId }}</text></view>
      <view class="info-row"><text class="label">性别</text><text>{{ member.gender }}</text></view>
      <view class="info-row"><text class="label">状态</text><text>{{ member.status }}</text></view>
      <view class="info-row">
        <text class="label">绑定用户</text>
        <text>{{ member.boundUserId ? '已绑定' : '未绑定' }}</text>
      </view>
      <view class="info-row">
        <text class="label">绑定策略</text>
        <text>{{ member.userBindingPolicy }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listFamilyMembers } from '@/api/members'
import { useSessionStore } from '@/stores/session'
import type { FamilyMember } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const members = ref<FamilyMember[]>([])
const loading = ref(false)
const errorMessage = ref('')

async function loadMembers() {
  if (!familyId.value) {
    errorMessage.value = '缺少 familyId。'
    return
  }
  const route = `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  errorMessage.value = ''
  try {
    members.value = await listFamilyMembers(familyId.value)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '成员列表加载失败。')
  } finally {
    loading.value = false
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  loadMembers()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.section-row,
.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.section-row .section-title {
  margin-bottom: 8rpx;
}

.info-row {
  padding: 12rpx 0;
  border-top: 1rpx solid #e5e0d6;
  font-size: 26rpx;
}

.label {
  color: #6b7280;
}
</style>
