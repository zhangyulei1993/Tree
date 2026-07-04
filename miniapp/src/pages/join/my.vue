<template>
  <view class="tree-page join-page">
    <MiniBackHome />
    <template v-if="authChecked">
      <MiniCard variant="hero" class="join-hero">
        <MiniNotice tone="security">
          申请信息仅家庭管理员可见，用于核实身份。审核结果会在此页面更新。
        </MiniNotice>
        <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadRequests">
          刷新列表
        </MiniButton>
        <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
      </MiniCard>

      <MiniCard v-if="loading && requests.length === 0">
        <view class="state-block">
          <text class="tree-muted">正在加载加入申请...</text>
        </view>
      </MiniCard>

      <MiniCard v-else-if="loadError && requests.length === 0">
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
          description="你提交的家庭加入申请会显示在这里。请家人发送家庭邀请。"
        />
      </MiniCard>

      <template v-else>
        <view v-for="group in requestGroups" :key="group.key" class="request-group">
          <MiniSectionHeader :title="group.title" :subtitle="group.subtitle" />
          <MiniCard v-for="item in group.items" :key="item.requestId" variant="soft" class="request-card">
            <view class="item-head">
              <view>
                <text class="item-kicker">加入申请</text>
                <text class="item-name">{{ item.familyName || '未知家庭' }}</text>
              </view>
              <MiniStatusTag :status="item.requestStatus" :label="joinRequestStatusText(item.requestStatus)" />
            </view>
            <view class="request-detail-grid">
              <text v-if="item.applicantRealName" class="item-meta">申请人：{{ item.applicantRealName }}</text>
              <text class="item-meta full">申请理由：{{ item.applicantMessage || '未填写' }}</text>
              <text v-if="item.handleComment" class="item-meta full">审核意见：{{ item.handleComment }}</text>
              <text class="item-meta weak">提交时间：{{ formatDate(item.createdAt) }}</text>
              <text v-if="item.cancelledAt" class="item-meta weak">取消时间：{{ formatDate(item.cancelledAt) }}</text>
              <text v-else-if="item.requestStatus !== 'PENDING'" class="item-meta weak">
                更新时间：{{ formatDate(item.updatedAt) }}
              </text>
            </view>
            <view class="item-actions">
              <MiniButton
                v-if="item.requestStatus === 'PENDING'"
                variant="danger"
                :disabled="cancellingId === item.requestId"
                :loading="cancellingId === item.requestId"
                @click="confirmCancel(item)"
              >
                取消申请
              </MiniButton>
              <MiniButton
                v-else-if="item.requestStatus === 'APPROVED'"
                variant="secondary"
                @click="openMyFamilies"
              >
                进入我的家庭
              </MiniButton>
            </view>
          </MiniCard>
        </view>
      </template>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onHide, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { cancelJoinRequest, listMyJoinRequests } from '@/api/joinRequests'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
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
const authChecked = ref(false)
const requestGroups = computed(() => {
  const pending = requests.value.filter((item) => item.requestStatus === 'PENDING')
  const history = requests.value.filter((item) => item.requestStatus !== 'PENDING')
  return [
    { key: 'pending', title: '待家庭审核', subtitle: '可在审核前取消申请', items: pending },
    { key: 'history', title: '历史申请', subtitle: '已通过、驳回或取消的记录', items: history }
  ].filter((group) => group.items.length > 0)
})

function resetTransientUI() {
  actionError.value = ''
  cancellingId.value = null
}

function resetPageData() {
  authChecked.value = false
  loading.value = false
  loadError.value = ''
  requests.value = []
  resetTransientUI()
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

function openMyFamilies() {
  uni.navigateTo({ url: '/pages/family/my' })
}

async function loadRequests() {
  session.restoreSession()
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin('/pages/join/my')
    return
  }
  authChecked.value = true
  const isInitialLoad = requests.value.length === 0
  loading.value = isInitialLoad
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

onShow(loadRequests)
onHide(resetTransientUI)
onUnload(resetPageData)
</script>

<style scoped>
.join-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.join-hero {
  margin-bottom: 26rpx;
}

.join-hero :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.state-block {
  padding: 32rpx 0;
  text-align: center;
}

.request-group {
  margin-bottom: 26rpx;
}

.item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.item-kicker {
  display: block;
  margin-bottom: 6rpx;
  color: var(--tree-green);
  font-size: 21rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}

.item-name {
  flex: 1;
  color: var(--tree-text-primary);
  font-size: 34rpx;
  font-weight: 800;
}

.item-meta {
  display: block;
  color: var(--tree-text-secondary);
  font-size: 23rpx;
  line-height: 1.55;
}

.item-meta.weak {
  color: var(--tree-text-weak);
}

.item-meta.full {
  grid-column: 1 / -1;
}

.request-detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10rpx 16rpx;
  border-radius: 22rpx;
  background: rgba(248, 250, 252, 0.82);
  padding: 18rpx;
}

.item-actions {
  margin-top: 20rpx;
}

.request-card {
  position: relative;
  border-color: rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.12), transparent 160rpx),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(244, 250, 247, 0.94) 100%);
  box-shadow: 0 16rpx 40rpx rgba(24, 54, 83, 0.07);
}

.request-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 26rpx;
  bottom: 26rpx;
  width: 7rpx;
  border-radius: 0 999rpx 999rpx 0;
  background: linear-gradient(180deg, var(--tree-green) 0%, var(--tree-primary) 100%);
}
</style>
