<template>
  <view class="archive-page notifications-page">
    <MiniBackHome />
    <view class="archive-page-head">
      <text class="archive-kicker">Updates</text>
      <text class="archive-title">消息中心</text>
      <text class="archive-subtitle">家庭、邀请、支付与审核更新</text>
    </view>
    <view class="notification-filter">
      <view class="filter-tab" :class="{ active: filter === 'ALL' }" @tap="filter = 'ALL'">全部</view>
      <view class="filter-tab" :class="{ active: filter === 'UNREAD' }" @tap="filter = 'UNREAD'">未读<text v-if="unreadCount" class="filter-count">{{ unreadCount }}</text></view>
      <view v-if="unreadCount" class="mark-all" @tap="markAllRead">全部已读</view>
    </view>
    <view v-if="loading && !items.length" class="archive-form-panel"><MiniEmptyState symbol="…" title="正在加载" description="正在同步消息..." /></view>
    <view v-else-if="errorMessage" class="notification-error archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" size="sm" @click="load(true)">重新加载</MiniButton>
    </view>
    <template v-else-if="visibleItems.length">
      <view class="notification-list">
        <view v-for="item in visibleItems" :key="item.id" class="notification-card" :class="{ unread: !item.read }" @tap="open(item)">
          <view class="notification-head"><text class="notification-category" :class="`category-${item.category.toLowerCase()}`">{{ categoryLabel(item.category) }}</text><text v-if="!item.read" class="notification-dot" /><text class="notification-time">{{ formatTime(item.createdAt) }}</text></view>
          <text class="notification-title">{{ item.title }}</text><text class="notification-content">{{ item.content }}</text>
        </view>
      </view>
      <view v-if="hasMore" class="load-more" :class="{ disabled: loadingMore }" @tap="loadMore">{{ loadingMore ? '正在加载...' : '加载更多消息' }}</view>
      <text v-else class="list-finished">已经到底了</text>
    </template>
    <view v-else class="archive-form-panel"><MiniEmptyState symbol="信" title="暂无消息" description="家庭事务中的待处理邀请和申请会在对应页面显示。" /></view>
  </view>
</template>

<script setup lang="ts">
import { onReachBottom, onShow } from '@dcloudio/uni-app'
import { computed, ref, watch } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { listNotifications, markAllNotificationsRead, markNotificationRead, type NotificationItem } from '@/api/notifications'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'

const pageSize = 20
const items = ref<NotificationItem[]>([])
const page = ref(1)
const total = ref(0)
const unreadCount = ref(0)
const loading = ref(false)
const loadingMore = ref(false)
const errorMessage = ref('')
const filter = ref<'ALL' | 'UNREAD'>('ALL')
const visibleItems = computed(() => filter.value === 'UNREAD' ? items.value.filter((item) => !item.read) : items.value)
const hasMore = computed(() => items.value.length < total.value)
const tabBarRoutes = new Set(['/pages/home/index', '/pages/family/my', '/pages/content/index', '/pages/me/index'])

function categoryLabel(category: string) { return ({ PAYMENT: '支付', REFUND: '退款', AUDIT: '审核', MEMBER: '成员', MEMBER_CHANGE: '成员变更', INVITATION: '邀请' } as Record<string, string>)[category] || '家庭动态' }
function formatTime(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN', { hour12: false }) }
async function load(reset = false): Promise<boolean> {
  if (loading.value || loadingMore.value) return false
  if (reset) { page.value = 1; items.value = [] }
  const isFirstPage = page.value === 1
  if (isFirstPage) loading.value = true; else loadingMore.value = true
  errorMessage.value = ''
  try {
    const result = await listNotifications(page.value, pageSize)
    const existing = new Set(items.value.map((item) => String(item.id)))
    items.value.push(...result.items.filter((item) => !existing.has(String(item.id))))
    total.value = result.total
    unreadCount.value = result.unreadCount
    return true
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '消息加载失败')
    return false
  } finally { loading.value = false; loadingMore.value = false }
}
async function loadMore() {
  if (!hasMore.value || loading.value || loadingMore.value) return
  const previousPage = page.value
  page.value += 1
  if (!await load()) page.value = previousPage
}
function openRoute(route: string, onSuccess: () => void) {
  const normalized = route.startsWith('/') ? route : `/${route}`
  const basePath = normalized.split(/[?#]/, 1)[0]
  if (tabBarRoutes.has(basePath)) {
    uni.switchTab({ url: basePath, success: onSuccess, fail: () => uni.showToast({ title: '打开消息页面失败', icon: 'none' }) })
    return
  }
  uni.navigateTo({ url: normalized, success: onSuccess, fail: () => uni.showToast({ title: '打开消息页面失败', icon: 'none' }) })
}
async function markRead(item: NotificationItem) {
  if (!item.read) {
    try {
      await markNotificationRead(item.id)
      item.read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    } catch {
      uni.showToast({ title: '消息状态更新失败', icon: 'none' })
    }
  }
}
function open(item: NotificationItem) {
  if (item.route) {
    openRoute(item.route, () => { void markRead(item) })
    return
  }
  void markRead(item)
  uni.showToast({ title: '该消息暂无可打开的关联页面', icon: 'none' })
}
async function markAllRead() {
  if (!unreadCount.value) return
  try {
    await markAllNotificationsRead()
    items.value.forEach((item) => { item.read = true })
    unreadCount.value = 0
    uni.showToast({ title: '已全部标记为已读', icon: 'success' })
  } catch {
    uni.showToast({ title: '操作失败，请稍后重试', icon: 'none' })
  }
}
watch(filter, async () => {
  while (filter.value === 'UNREAD' && !visibleItems.value.length && hasMore.value && !errorMessage.value) {
    await loadMore()
  }
})
onShow(() => { void load(true) })
onReachBottom(loadMore)
</script>

<style scoped>
.notifications-page { padding-top: 28rpx; }.notification-filter { display:flex; gap:14rpx; align-items:center; margin-top:24rpx; }.filter-tab { display:inline-flex; align-items:center; justify-content:center; min-height:60rpx; padding:0 24rpx; border:1rpx solid var(--archive-line); border-radius:999rpx; color:var(--archive-ink-soft); font-size:25rpx; }.filter-tab.active { border-color:var(--archive-cinnabar); color:var(--archive-cinnabar); background:var(--archive-paper); }.filter-count { min-width:30rpx; height:30rpx; margin-left:8rpx; border-radius:50%; background:var(--archive-cinnabar); color:#fff; font-size:19rpx; line-height:30rpx; text-align:center; }.mark-all { margin-left:auto; color:var(--archive-cinnabar); font-size:24rpx; }.notification-error{margin-top:22rpx}.notification-error :deep(.mini-notice){margin-bottom:18rpx}.notification-list { display:grid; gap:16rpx; margin-top:22rpx; }.notification-card { padding:24rpx; background:var(--archive-paper-deep); border:1rpx solid var(--archive-line); border-radius:14rpx; }.notification-card.unread { border-color:var(--archive-line-strong); box-shadow:inset 5rpx 0 0 rgba(168,59,45,.72); }.notification-head { display:flex; align-items:center; gap:10rpx; }.notification-category { padding:4rpx 10rpx; border-radius:6rpx; color:var(--archive-blue); background:rgba(22,51,83,.08); font-size:20rpx; }.category-payment { color:var(--archive-green,#2f6b57); background:rgba(47,107,87,.09); }.category-refund,.category-audit { color:var(--archive-cinnabar); background:rgba(168,59,45,.08); }.notification-dot { width:12rpx; height:12rpx; border-radius:50%; background:var(--archive-cinnabar); }.notification-time { margin-left:auto; color:var(--archive-ink-soft); font-size:21rpx; }.notification-title { display:block; margin-top:14rpx; color:var(--archive-ink); font-size:27rpx; font-weight:700; }.notification-content { display:block; margin-top:9rpx; color:var(--archive-ink-soft); font-size:24rpx; line-height:1.58; }.load-more,.list-finished { display:block; margin:22rpx auto 8rpx; color:var(--archive-blue); font-size:24rpx; text-align:center; }.load-more.disabled,.list-finished { color:var(--archive-ink-soft); }
</style>
