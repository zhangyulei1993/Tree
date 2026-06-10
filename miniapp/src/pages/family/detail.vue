<template>
  <view class="page">
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载家庭详情...</text>
    </view>
    <view v-else-if="errorMessage" class="card state-card">
      <text class="error">{{ errorMessage }}</text>
      <button class="button secondary" @click="loadFamily">重新加载</button>
    </view>
    <template v-else-if="family">
      <view class="card">
        <text class="title">{{ family.familyName }}</text>
        <text class="tag">{{ family.role }}</text>
        <text class="tag">{{ family.status }}</text>
        <text class="tag">{{ family.publicDisplayStatus }}</text>
        <view class="info-grid">
          <view><text class="label">姓氏</text><text>{{ family.familySurname }}</text></view>
          <view><text class="label">籍贯</text><text>{{ family.nativePlace || '未设置' }}</text></view>
          <view><text class="label">地区</text><text>{{ family.regionText || '未设置' }}</text></view>
          <view><text class="label">图版本</text><text>{{ family.graphVersion }}</text></view>
          <view>
            <text class="label">创始成员 ID</text>
            <text>{{ family.currentFounderMemberId ?? '未设置' }}</text>
          </view>
          <view><text class="label">允许搜索</text><text>{{ family.searchable ? '是' : '否' }}</text></view>
        </view>
        <text class="muted">{{ family.description || '暂未填写家庭简介。' }}</text>
      </view>
      <view class="card">
        <text class="section-title">家庭数据</text>
        <button class="button" @click="openMembers">查看成员列表</button>
        <button class="button secondary" @click="openTree">查看私有家庭树</button>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')

async function loadFamily() {
  if (!familyId.value) {
    errorMessage.value = '缺少 familyId。'
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

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  loadFamily()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.error {
  display: block;
  color: #c0392b;
  font-size: 24rpx;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
  margin: 20rpx 0;
}

.info-grid view {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.label {
  color: #6b7280;
  font-size: 22rpx;
}
</style>
