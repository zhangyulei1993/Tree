<template>
  <view class="archive-page join-my-page">
    <MiniBackHome />

    <view class="join-head archive-page-head">
      <view>
        <text class="archive-kicker">Join Requests</text>
        <text class="archive-title">我提交的加入申请</text>
        <text class="archive-subtitle">查看申请进度与审核结果，待审核时可主动取消</text>
      </view>
      <view class="archive-seal">申</view>
    </view>

    <template v-if="authChecked">
      <view class="join-summary">
        <text class="archive-chip">审核中 {{ pendingCount }}</text>
        <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadRequests">
          刷新列表
        </MiniButton>
      </view>

      <MiniNotice tone="security" title="申请说明">
        申请信息仅家庭管理员可见，用于核实身份。审核结果会在此页面更新。
      </MiniNotice>

      <view v-if="loading && requests.length === 0" class="join-state archive-form-panel">
        <MiniEmptyState symbol="…" title="正在加载" description="正在同步加入申请..." />
      </view>

      <view v-else-if="loadError && requests.length === 0" class="join-state archive-form-panel">
        <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
        <MiniButton variant="secondary" size="sm" class="join-action" @click="loadRequests">重新加载</MiniButton>
      </view>

      <view v-else-if="requests.length === 0" class="join-state archive-form-panel">
        <MiniEmptyState
          symbol="申"
          title="暂无加入申请"
          description="你提交的家庭加入申请会显示在这里。请家人发送家庭邀请。"
        />
      </view>

      <template v-else>
        <view v-for="group in requestGroups" :key="group.key" class="join-group archive-form-panel">
          <view class="archive-section-head">
            <text class="archive-section-title">{{ group.title }}</text>
            <text class="archive-section-subtitle">{{ group.subtitle }}</text>
          </view>
          <view class="join-list archive-list">
            <view v-for="item in group.items" :key="item.requestId" class="join-item">
              <view class="join-item-head">
                <view class="archive-row-main">
                  <text class="archive-row-title">{{ item.familyName || '未知家庭' }}</text>
                  <text class="archive-row-desc">
                    申请人：{{ item.applicantRealName || '未填写姓名' }}
                  </text>
                </view>
                <text class="archive-status-tag" :class="joinStatusClass(item.requestStatus)">
                  {{ joinRequestStatusText(item.requestStatus) }}
                </text>
              </view>

              <view class="join-meta">
                <text>提交时间：{{ formatDate(item.createdAt) }}</text>
                <text>处理结果：{{ handleResultText(item) }}</text>
              </view>

              <view v-if="expandedRequestId === item.requestId" class="join-detail">
                <text>申请理由：{{ item.applicantMessage || '未填写' }}</text>
                <text v-if="item.handleComment">审核意见：{{ item.handleComment }}</text>
                <text v-if="item.cancelledAt" class="join-meta-weak">取消时间：{{ formatDate(item.cancelledAt) }}</text>
                <text v-else-if="item.requestStatus !== 'PENDING'" class="join-meta-weak">
                  更新时间：{{ formatDate(item.updatedAt) }}
                </text>
              </view>

              <view class="join-actions">
                <MiniButton
                  v-if="item.requestStatus === 'PENDING'"
                  size="sm"
                  variant="secondary"
                  :disabled="cancellingId === item.requestId"
                  :loading="cancellingId === item.requestId"
                  @click="confirmCancel(item)"
                >
                  取消申请
                </MiniButton>
                <MiniButton
                  v-else-if="item.requestStatus === 'APPROVED'"
                  size="sm"
                  variant="secondary"
                  @click="openMyFamilies"
                >
                  进入我的家庭
                </MiniButton>
                <MiniButton size="sm" variant="secondary" @click="toggleRequestDetail(item)">
                  {{ expandedRequestId === item.requestId ? '收起详情' : '查看详情' }}
                </MiniButton>
              </view>
            </view>
          </view>
        </view>
      </template>

      <text v-if="actionError" class="tree-field-error join-error">{{ actionError }}</text>
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
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { joinRequestStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { JoinRequest } from '@/types/api'

const session = useSessionStore()
const requests = ref<JoinRequest[]>([])
const loading = ref(false)
const loadError = ref('')
const actionError = ref('')
const cancellingId = ref<number | string | null>(null)
const authChecked = ref(false)
const expandedRequestId = ref<number | string | null>(null)

const pendingCount = computed(() =>
  requests.value.filter((item) => item.requestStatus === 'PENDING').length
)

const requestGroups = computed(() => {
  const pending = requests.value.filter((item) => item.requestStatus === 'PENDING')
  const history = requests.value.filter((item) => item.requestStatus !== 'PENDING')
  return [
    { key: 'pending', title: '待家庭审核', subtitle: '可在审核前取消申请', items: pending },
    { key: 'history', title: '历史申请', subtitle: '已通过、驳回或取消的记录', items: history }
  ].filter((group) => group.items.length > 0)
})

function joinStatusClass(status: string) {
  if (status === 'PENDING') return 'is-cinnabar'
  if (status === 'APPROVED') return 'is-ink'
  return 'is-muted'
}

function handleResultText(item: JoinRequest) {
  if (item.requestStatus === 'PENDING') return '等待家庭管理员审核'
  if (item.requestStatus === 'APPROVED') return item.handleComment || '申请已通过'
  if (item.requestStatus === 'REJECTED') return item.handleComment || '申请已被驳回'
  if (item.requestStatus === 'CANCELLED') return '已主动取消'
  return '—'
}

function resetTransientUI() {
  actionError.value = ''
  cancellingId.value = null
  expandedRequestId.value = null
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

function toggleRequestDetail(item: JoinRequest) {
  expandedRequestId.value =
    expandedRequestId.value === item.requestId ? null : item.requestId
}

function openMyFamilies() {
  uni.switchTab({ url: '/pages/family/my' })
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
.join-my-page {
  padding-top: 28rpx;
}

.join-my-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.join-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 22rpx;
}

.join-my-page :deep(.mini-notice) {
  margin-bottom: 20rpx;
}

.join-group {
  margin-bottom: 24rpx;
}

.join-list {
  margin-top: 8rpx;
}

.join-item {
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.join-item:last-child {
  border-bottom: 0;
}

.join-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.archive-status-tag {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx solid var(--archive-line);
  padding: 0 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  line-height: 1.2;
}

.archive-status-tag.is-cinnabar {
  border-color: rgba(168, 59, 45, 0.28);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
}

.archive-status-tag.is-ink {
  border-color: rgba(22, 51, 83, 0.22);
  background: rgba(22, 51, 83, 0.08);
  color: var(--archive-blue);
}

.archive-status-tag.is-muted {
  background: rgba(255, 248, 234, 0.58);
  color: var(--archive-ink-soft);
}

.join-meta,
.join-detail {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.join-detail {
  margin-top: 10rpx;
  padding-top: 10rpx;
  border-top: 1rpx dashed rgba(44, 36, 22, 0.12);
}

.join-meta-weak {
  opacity: 0.88;
}

.join-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.join-action {
  margin-top: 16rpx;
  align-self: flex-start;
}

.join-error {
  display: block;
  margin-top: 16rpx;
}
</style>
