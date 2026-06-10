<template>
  <view class="page">
    <view class="card">
      <text class="title">我的加入申请</text>
      <button class="button secondary" :disabled="loading" :loading="loading" @click="loadRequests">
        刷新
      </button>
      <text v-if="actionError" class="error">{{ actionError }}</text>
    </view>

    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载加入申请...</text>
    </view>
    <view v-else-if="loadError" class="card state-card">
      <text class="error">{{ loadError }}</text>
      <button class="button secondary" @click="loadRequests">重新加载</button>
    </view>
    <view v-else-if="requests.length === 0" class="card state-card">
      <text class="muted">暂无加入申请。</text>
    </view>
    <view v-else>
      <view v-for="item in requests" :key="item.requestId" class="card item-card">
        <view class="item-header">
          <text class="section-title">{{ item.familyName || `家庭 ${item.familyId}` }}</text>
          <text class="tag">{{ item.requestStatus }}</text>
        </view>
        <text v-if="item.applicantRealName" class="muted">申请人：{{ item.applicantRealName }}</text>
        <text class="muted">申请理由：{{ item.applicantMessage || '未填写' }}</text>
        <text v-if="item.handleComment" class="muted">审核意见：{{ item.handleComment }}</text>
        <text class="muted">创建时间：{{ formatDate(item.createdAt) }}</text>
        <button
          v-if="item.requestStatus === 'PENDING'"
          class="button danger"
          :disabled="cancellingId === item.requestId"
          :loading="cancellingId === item.requestId"
          @click="confirmCancel(item)"
        >
          取消申请
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { cancelJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import { useSessionStore } from '@/stores/session'
import type { JoinRequest } from '@/types/api'

const session = useSessionStore()
const requests = ref<JoinRequest[]>([])
const loading = ref(false)
const loadError = ref('')
const actionError = ref('')
const cancellingId = ref<number | string | null>(null)

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

async function loadRequests() {
  if (!session.requireLogin('/pages/join/my')) return
  loading.value = true
  loadError.value = ''
  actionError.value = ''
  try {
    requests.value = await listMyJoinRequests()
  } catch (error) {
    loadError.value = apiErrorMessage(error, '加入申请列表加载失败。')
  } finally {
    loading.value = false
  }
}

function confirmCancel(item: JoinRequest) {
  uni.showModal({
    title: '取消加入申请',
    content: `确定取消对 ${item.familyName || `家庭 ${item.familyId}`} 的加入申请吗？`,
    success: (result) => {
      if (result.confirm) cancel(item)
    }
  })
}

async function cancel(item: JoinRequest) {
  cancellingId.value = item.requestId
  actionError.value = ''
  try {
    await cancelJoinRequest(item.familyId, item.requestId, { cancelReason: '用户主动取消' })
    await loadRequests()
  } catch (error) {
    actionError.value = apiErrorMessage(error, '取消加入申请失败。')
  } finally {
    cancellingId.value = null
  }
}

onLoad(() => {
  session.restoreSession()
  loadRequests()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.item-card {
  margin-top: 20rpx;
}

.item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.button.danger {
  background: #c0392b;
}

.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.button[disabled] {
  opacity: 0.55;
}
</style>
