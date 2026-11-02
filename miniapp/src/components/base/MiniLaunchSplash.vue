<template>
  <view v-if="visible" class="launch-splash" :class="{ 'launch-splash--night': activeImageIsNight }">
    <view v-if="config.skipEnabled" class="launch-splash-skip" @tap.stop="emit('close')">跳过</view>

    <image
      v-if="activeImageUrl"
      class="launch-splash-image"
      :class="{ 'launch-splash-image--ready': imageReady }"
      :src="activeImageUrl"
      mode="aspectFill"
      :lazy-load="false"
      @load="handleImageLoad"
      @error="handleImageError"
    />
    <view class="launch-splash-copy">
      <view v-if="!activeImageUrl" class="launch-splash-seal">谱</view>
      <text class="launch-splash-name">{{ config.title }}</text>
      <text class="launch-splash-tagline">{{ config.tagline }}</text>
    </view>

    <view class="launch-splash-footer">
      <view class="launch-splash-line" />
      <text>家谱记录与家人协作</text>
      <view class="launch-splash-line" />
    </view>
  </view>
</template>

<script setup lang="ts">
import type { LaunchSplashConfig } from '@/api/launchSplash'
import { computed, ref, watch } from 'vue'

const props = defineProps<{ visible: boolean; config: LaunchSplashConfig }>()
const imageReady = ref(false)

const activeImageUrl = computed(() => {
  const hour = new Date().getHours()
  const isDaytime = hour >= 6 && hour < 18
  const preferred = isDaytime ? props.config.dayImageUrl : props.config.nightImageUrl
  const fallback = isDaytime ? props.config.nightImageUrl : props.config.dayImageUrl
  return preferred || fallback || props.config.imageUrl
})

const activeImageIsNight = computed(() => Boolean(props.config.nightImageUrl && activeImageUrl.value === props.config.nightImageUrl))

function handleImageLoad() {
  imageReady.value = true
  emit('visualReady')
}

function handleImageError() {
  imageReady.value = false
  emit('visualReady')
}

watch(activeImageUrl, () => {
  imageReady.value = false
})

const emit = defineEmits<{ close: []; visualReady: [] }>()
</script>

<style scoped>
.launch-splash {
  position: fixed;
  inset: 0;
  z-index: 9000;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  padding: calc(76rpx + env(safe-area-inset-top)) 48rpx calc(62rpx + env(safe-area-inset-bottom));
  overflow: hidden;
  background-color: var(--archive-paper);
  background-image:
    linear-gradient(rgba(92, 74, 48, 0.03) 1rpx, transparent 1rpx),
    linear-gradient(90deg, rgba(92, 74, 48, 0.022) 1rpx, transparent 1rpx),
    linear-gradient(180deg, var(--archive-paper-light) 0%, var(--archive-paper) 58%, var(--archive-paper-deep) 100%);
  background-size: 100% 34rpx, 34rpx 100%, 100% 100%;
  animation: launch-splash-in 260ms ease-out both;
}

.launch-splash-skip {
  position: absolute;
  top: calc(36rpx + env(safe-area-inset-top));
  right: 34rpx;
  z-index: 1;
  padding: 14rpx 20rpx;
  border: 1rpx solid rgba(22, 51, 83, 0.16);
  border-radius: 999rpx;
  color: var(--archive-ink-soft);
  background: rgba(255, 255, 255, 0.42);
  font-size: 24rpx;
  line-height: 1;
}

.launch-splash-copy {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: launch-splash-copy-in 560ms 120ms cubic-bezier(.22, .8, .24, 1) both;
}

.launch-splash-image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  background: var(--archive-paper);
  opacity: 0;
  transition: opacity 220ms ease-out;
}

.launch-splash-image--ready {
  opacity: 1;
}

.launch-splash-image + .launch-splash-copy {
  padding: 24rpx 34rpx;
  border: 1rpx solid rgba(255, 252, 244, 0.42);
  border-radius: 16rpx;
  background: rgba(255, 252, 244, 0.22);
  box-shadow: 0 12rpx 32rpx rgba(15, 32, 51, 0.12);
  backdrop-filter: blur(12rpx);
}

.launch-splash--night .launch-splash-image + .launch-splash-copy {
  border-color: rgba(236, 242, 248, 0.3);
  background: rgba(12, 29, 48, 0.38);
  box-shadow: 0 12rpx 32rpx rgba(3, 12, 23, 0.2);
}

.launch-splash--night .launch-splash-name {
  color: rgba(248, 245, 234, 0.96);
}

.launch-splash--night .launch-splash-tagline,
.launch-splash--night .launch-splash-footer {
  color: rgba(239, 243, 246, 0.82);
}

.launch-splash--night .launch-splash-line {
  background: rgba(239, 243, 246, 0.42);
}

.launch-splash-seal {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 106rpx;
  height: 106rpx;
  border: 3rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 52rpx;
  font-weight: 700;
  line-height: 1;
  transform: rotate(-7deg);
  animation: launch-splash-seal-in 520ms 80ms cubic-bezier(.22, .8, .24, 1) both;
}

.launch-splash-name {
  display: block;
  margin-top: 30rpx;
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 58rpx;
  font-weight: 700;
  line-height: 1.15;
  letter-spacing: 2rpx;
}

.launch-splash-tagline {
  display: block;
  margin-top: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 25rpx;
  letter-spacing: 3rpx;
}

.launch-splash-footer {
  position: absolute;
  bottom: calc(58rpx + env(safe-area-inset-bottom));
  left: 0;
  right: 0;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  color: rgba(101, 112, 128, 0.78);
  font-size: 20rpx;
  letter-spacing: 2rpx;
}

.launch-splash-line {
  width: 48rpx;
  height: 1rpx;
  background: rgba(92, 74, 48, 0.28);
}

@keyframes launch-splash-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes launch-splash-copy-in {
  from { opacity: 0; transform: translateY(24rpx); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes launch-splash-seal-in {
  from { opacity: 0; transform: rotate(-7deg) scale(.88); }
  to { opacity: 1; transform: rotate(-7deg) scale(1); }
}
</style>
