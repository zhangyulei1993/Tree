<template>
  <view class="tree-page">
    <MiniSectionHeader title="我的加入申请" subtitle="查看你提交的家庭加入申请及审核进度。" />

    <MiniCard>
      <MiniNotice tone="security">
        申请信息仅家庭管理员可见，用于核实身份。审核结果会在此页面更新。
      </MiniNotice>
      <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadRequests">
        刷新列表
      </MiniButton>
      <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
    </MiniCard>

    <MiniCard v-if="loading">
      <view class="state-block">
        <text class="tree-muted">正在加载加入申请...</text>
      </view>
    </MiniCard>

    <MiniCard v-else-if="loadError">
      <MiniEmptyState
        symbol="!"
        title="加载失败"
        :description="loadError"
        action-text="重新加载"
        @action="loadRequests"
      />
    </MiniCard>

    <MiniCard v-else-if="requests.length === 0">
      <MiniEmptyState
        symbol="申"
        title="暂无加入申请"
        description="你提交的家庭加入申请会显示在这里。可在公开家庭主页提交加入申请，或请家人发送邀请。"
        action-text="寻找家族"
        @action="openSearch"
      />
    </MiniCard>

    <template v-else>
      <MiniCard v-for="item in requests" :key="item.requestId" variant="soft">
      <view class="item-head">
        <text class="item-name">{{ item.familyName || '未知家庭' }}</text>
        <MiniStatusTag :status="item.requestStatus" :label="joinRequestStatusText(item.requestStatus)" />
      </view>
      <text v-if="item.applicantRealName" class="tree-muted item-meta">申请人：{{ item.applicantRealName }}</text>
      <text class="tree-muted item-meta">申请理由：{{ item.applicantMessage || '未填写' }}</text>
      <text v-if="item.handleComment" class="tree-muted item-meta">审核意见：{{ item.handleComment }}</text>
      <text class="tree-weak item-meta">提交时间：{{ formatDate(item.createdAt) }}</text>
      <text v-if="item.cancelledAt" class="tree-weak item-meta">取消时间：{{ formatDate(item.cancelledAt) }}</text>
      <text v-else-if="item.requestStatus !== 'PENDING'" class="tree-weak item-meta">
        更新时间：{{ formatDate(item.updatedAt) }}
      </text>
      <MiniButton
        v-if="item.requestStatus === 'PENDING'"
        variant="danger"
        :disabled="cancellingId === item.requestId"
        :loading="cancellingId === item.requestId"
        @click="confirmCancel(item)"
      >
        取消申请
      </MiniButton>
      </MiniCard>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { cancelJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { joinRequestStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import MiniStatusTag from '@/components/base/MiniStatusTag.vue'
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

function openSearch() {
  uni.switchTab({ url: '/pages/family/search' })
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
    content: `确定取消对「${item.familyName || '该家庭'}」的加入申请吗？取消后需要重新提交申请。`,
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
.state-block {
  padding: 32rpx 0;
  text-align: center;
}

.item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.item-name {
  flex: 1;
  color: #1f2937;
  font-size: 30rpx;
  font-weight: 700;
}

.item-meta {
  display: block;
  margin-top: 6rpx;
}
</style>
