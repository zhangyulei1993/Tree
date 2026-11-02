<template>
  <view class="archive-page capacity-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />
    <view class="archive-page-head">
      <view>
        <text class="archive-kicker">Family Capacity</text>
        <text class="archive-title">成员名额与订单</text>
        <text class="archive-subtitle">购买扩容、继续支付或管理退款。</text>
      </view>
    </view>
    <view v-if="loading" class="capacity-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在同步可购买档位..." />
    </view>
    <view v-else-if="loadError" class="capacity-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" size="sm" @click="load">重新加载</MiniButton>
    </view>
    <template v-else>
      <MiniNotice tone="info">购买的是成员名额，不会自动创建成员。</MiniNotice>
      <view v-if="reward?.enabled && reward.active" class="reward-entry" @tap="openInvitationReward"><view><text class="reward-entry-title">邀请家人解锁奖励名额</text><text class="reward-entry-desc">{{ reward.validInvites }} 位有效邀请，已奖励 +{{ reward.rewardedMembers }} 位</text></view><text class="capacity-order-arrow">›</text></view>
      <view v-if="products.length" class="capacity-picker">
      <view class="capacity-picker-head"><text class="picker-label">选择扩容档位</text><text class="picker-count">{{ productIndex + 1 }} / {{ products.length }}</text></view>
      <swiper class="capacity-swiper" :current="productIndex" :duration="220" :indicator-dots="true" indicator-color="rgba(38, 58, 82, .18)" indicator-active-color="#b64032" @change="onProductChange">
        <swiper-item v-for="product in products" :key="product.productKey">
          <view class="capacity-item"><view><text class="capacity-name">{{ product.displayName }}</text><text class="capacity-meta">增加 {{ product.addedMembers }} 位可添加成员</text><text class="capacity-meta">支付 {{ (product.amountFen / 100).toFixed(2) }} 元 · {{ product.refundWindowMinutes }} 分钟内可申请退款</text></view></view>
        </swiper-item>
      </swiper>
      <view class="capacity-purchase"><text class="capacity-selected">当前选择：{{ selectedProduct?.displayName || '请选择档位' }}</text><view class="capacity-buy" :class="{ disabled: buying }" @tap="buySelected"><text>{{ buying ? '处理中' : '立即购买' }}</text><text class="capacity-arrow">›</text></view></view>
      </view>
      <MiniEmptyState v-else symbol="扩" title="暂无可购买档位" description="当前没有可用的家庭扩容商品，请稍后再试。" />
      <view class="capacity-order-entry" @tap="openOrders"><view><text class="capacity-order-title">查看订单与退款</text><text class="capacity-order-desc">继续支付、申请退款并查看历史订单</text></view><text class="capacity-order-arrow">›</text></view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { CapacityPaymentPendingError, listCapacityProducts, payCapacityOrder, type CapacityProduct } from '@/api/capacity'
import { capacityPaymentFeedback } from '@/features/capacity/paymentOutcome'
import { getInvitationReward, type InvitationRewardSummary } from '@/api/invitationReward'
import { requestSubscribeMessage } from '@/features/notifications/subscribe'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'

const familyId = ref('')
const returnTo = ref('')
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : '返回家庭详情')
const products = ref<CapacityProduct[]>([])
const productIndex = ref(0)
const buying = ref(false)
const loading = ref(false)
const loadError = ref('')
const reward = ref<InvitationRewardSummary>()
const navigating = ref(false)
const selectedProduct = computed(() => products.value[productIndex.value])
let skipInitialShow = true

async function load() {
  if (loading.value) return
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    const [productResult, rewardResult] = await Promise.allSettled([
      listCapacityProducts(),
      getInvitationReward(familyId.value)
    ])
    if (productResult.status === 'rejected') throw productResult.reason
    products.value = productResult.value
    reward.value = rewardResult.status === 'fulfilled' ? rewardResult.value : undefined
    if (productIndex.value >= products.value.length) productIndex.value = 0
  } catch (error) {
    loadError.value = apiErrorMessage(error, '商品加载失败。')
  } finally { loading.value = false }
}
function onProductChange(event: { detail: { current: number } }) { productIndex.value = Number(event.detail.current) }
async function buySelected() {
  const product = selectedProduct.value
  if (!product || buying.value) return
  buying.value = true
  try {
    await requestSubscribeMessage('payment')
    const order = await payCapacityOrder(familyId.value, product.productKey)
    uni.showToast({ title: capacityPaymentFeedback(order.status), icon: 'none' })
    if (order.status === 'PAID') {
      if (returnTo.value === 'add') {
        setTimeout(() => uni.navigateBack(), 450)
      } else {
        setTimeout(() => uni.navigateTo({
          url: `/pages/family/capacity-orders?familyId=${encodeURIComponent(familyId.value)}&orderNo=${encodeURIComponent(order.orderNo)}${returnTo.value === 'tree' ? '&returnTo=tree' : ''}`
        }), 450)
      }
    }
  } catch (error) {
    if (error instanceof CapacityPaymentPendingError) {
      uni.showToast({ title: error.message, icon: 'none' })
      setTimeout(() => uni.navigateTo({
        url: `/pages/family/capacity-orders?familyId=${encodeURIComponent(familyId.value)}&orderNo=${encodeURIComponent(error.order.orderNo)}${returnTo.value === 'tree' ? '&returnTo=tree' : ''}`
      }), 350)
      return
    }
    uni.showToast({ title: apiErrorMessage(error, '支付未完成'), icon: 'none' })
  } finally { buying.value = false }
}
function navigateOnce(url: string) {
  if (navigating.value) return
  navigating.value = true
  uni.navigateTo({ url, complete: () => { navigating.value = false } })
}
function openOrders() { navigateOnce(`/pages/family/capacity-orders?familyId=${encodeURIComponent(familyId.value)}${returnTo.value === 'tree' ? '&returnTo=tree' : ''}`) }
function openInvitationReward() { navigateOnce(`/pages/family/invitation-reward?familyId=${encodeURIComponent(familyId.value)}${returnTo.value === 'tree' ? '&returnTo=tree' : ''}`) }
onLoad((options) => { familyId.value = String(options?.familyId || ''); returnTo.value = String(options?.returnTo || ''); void load() })
onShow(() => { if (skipInitialShow) { skipInitialShow = false; return }; if (familyId.value) void load() })
</script>

<style scoped>
.capacity-page { padding-top: 28rpx; }.capacity-state { margin-top: 24rpx; }.capacity-state :deep(.mini-notice) { margin-bottom: 18rpx; }.reward-entry { display:flex;align-items:center;justify-content:space-between;margin-top:24rpx;padding:24rpx 22rpx;background:rgba(181,122,41,.10);border:1px solid var(--archive-line);border-radius:14rpx; }.reward-entry-title,.reward-entry-desc{display:block}.reward-entry-title{color:var(--archive-ink);font-size:28rpx;font-weight:700}.reward-entry-desc{margin-top:8rpx;color:var(--archive-ink-soft);font-size:23rpx}.capacity-picker { margin-top: 24rpx; padding: 22rpx 20rpx 18rpx; background: var(--archive-paper-deep); border: 1px solid var(--archive-line); border-radius: 14rpx; }.capacity-picker-head,.capacity-purchase { display:flex; align-items:center; justify-content:space-between; gap:16rpx; }.picker-label,.capacity-selected { color:var(--archive-ink); font-size:25rpx; font-weight:700; }.picker-count { color:var(--archive-ink-soft); font-size:22rpx; }.capacity-swiper { height:190rpx; margin:8rpx -4rpx 2rpx; }.capacity-item { display:flex; align-items:center; min-height:142rpx; padding:24rpx; margin:0 4rpx; background:var(--archive-paper); border:1px solid var(--archive-line); border-radius:14rpx; }.capacity-item:active { background:var(--archive-paper-deep); }.capacity-name,.capacity-meta { display:block; }.capacity-name { color:var(--archive-ink); font-size:31rpx; font-weight:700; }.capacity-meta { color:var(--archive-ink-soft); font-size:23rpx; margin-top:8rpx; }.capacity-purchase { padding-top:10rpx; }.capacity-selected { max-width:58%; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }.capacity-buy { display:flex; align-items:center; gap:8rpx; padding:12rpx 18rpx; color:var(--archive-paper); background:var(--archive-cinnabar); border-radius:10rpx; font-size:25rpx; }.capacity-buy.disabled { opacity:.55; }.capacity-arrow { color:var(--archive-paper); font-size:34rpx; }.capacity-order-entry { display:flex; align-items:center; justify-content:space-between; margin-top:24rpx; padding:24rpx 22rpx; background:var(--archive-paper-deep); border:1px solid var(--archive-line); border-radius:14rpx; }.capacity-order-entry:active { background:var(--archive-paper); }.capacity-order-title,.capacity-order-desc { display:block; }.capacity-order-title { color:var(--archive-ink); font-size:28rpx; font-weight:700; }.capacity-order-desc { margin-top:8rpx; color:var(--archive-ink-soft); font-size:23rpx; }.capacity-order-arrow { color:var(--archive-cinnabar); font-size:44rpx; }
</style>
