<template>
  <view v-if="visible" class="mini-back-home">
    <button class="back-button" @click="goHome">
      <text class="back-icon">‹</text>
      <text>{{ label }}</text>
    </button>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

const props = defineProps<{
  returnUrl?: string
  returnLabel?: string
}>()

const visible = ref(false)
const label = computed(() => props.returnLabel || (getCurrentPages().length > 1 ? '返回上一页' : '返回首页'))

onMounted(() => {
  visible.value = true
})

function goHome() {
  if (props.returnUrl) {
    const pages = getCurrentPages()
    const targetPath = props.returnUrl.split('?')[0]
    const targetPageIndex = pages
      .slice(0, -1)
      .map((page) => `/${page.route}`)
      .lastIndexOf(targetPath)
    if (targetPageIndex >= 0) {
      uni.navigateBack({ delta: pages.length - 1 - targetPageIndex })
      return
    }
    if (props.returnUrl.startsWith('/pages/home/') || props.returnUrl.startsWith('/pages/family/my')) {
      uni.switchTab({ url: props.returnUrl })
    } else {
      uni.redirectTo({ url: props.returnUrl })
    }
    return
  }
  const pages = getCurrentPages()
  if (pages.length > 1) {
    uni.navigateBack()
    return
  }
  uni.switchTab({ url: '/pages/home/index' })
}
</script>

<style scoped>
.mini-back-home {
  margin-bottom: 18rpx;
}

.back-button {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  width: auto;
  min-height: 80rpx;
  margin: 0;
  border: 1rpx solid rgba(31, 58, 95, 0.1);
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.86);
  color: var(--tree-text-secondary, #6b7280);
  padding: 8rpx 18rpx 8rpx 14rpx;
  font-size: 26rpx;
  font-weight: 600;
  line-height: 1.4;
  box-shadow: 0 8rpx 24rpx rgba(31, 58, 95, 0.06);
  transition: transform 180ms ease-out, background-color 180ms ease-out;
}

.back-button:active {
  background-color: rgba(24, 54, 83, 0.05);
  transform: translateY(1rpx) scale(0.99);
}

.back-button:active .back-icon {
  transform: translateX(-4rpx);
}

.back-button::after {
  border: none;
}

.back-icon {
  color: var(--tree-text-primary, #1f3a5f);
  font-size: 34rpx;
  line-height: 1;
  transition: transform 180ms ease-out;
}
</style>
