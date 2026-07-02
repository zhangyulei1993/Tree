<template>
  <view class="tree-page about-page">
    <MiniBackHome />

    <MiniCard>
      <MiniSectionHeader title="关于 Tree" subtitle="协议、隐私与版本信息" accent />
      <MiniActionList :items="legalItems" @select="onLegalSelect" />
    </MiniCard>

    <MiniCard variant="soft">
      <MiniSectionHeader title="版本信息" subtitle="当前小程序运行环境" accent />
      <view class="version-grid">
        <view class="version-row">
          <text class="version-label">小程序版本</text>
          <text class="version-value">{{ buildInfo.version }}</text>
        </view>
        <view class="version-row">
          <text class="version-label">当前环境</text>
          <text class="version-value">{{ buildInfo.environment }}</text>
        </view>
      </view>
      <view class="tech-toggle" @click="toggleTechInfo">
        <text class="tech-toggle-label">技术信息</text>
        <text class="tech-toggle-arrow">{{ showTechInfo ? '收起' : '展开' }}</text>
      </view>
      <view
        v-if="techPanelMounted"
        class="tech-panel"
        :class="{ 'tech-panel--visible': techPanelExpanded }"
      >
        <view class="version-row">
          <text class="version-label">提交版本</text>
          <text class="version-value mono">{{ buildInfo.commit }}{{ buildInfo.dirty ? ' dirty' : '' }}</text>
        </view>
        <view class="version-row">
          <text class="version-label">接口地址</text>
          <text class="version-value mono">{{ buildInfo.apiBaseUrl || '未配置' }}</text>
        </view>
        <view class="version-row">
          <text class="version-label">构建时间</text>
          <text class="version-value mono">{{ buildInfo.buildTime || '—' }}</text>
        </view>
      </view>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onHide, onUnload } from '@dcloudio/uni-app'
import { nextTick, ref } from 'vue'

import { buildInfo } from '@/buildInfo'
import MiniActionList from '@/components/base/MiniActionList.vue'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'

const TECH_PANEL_MS = 200

const showTechInfo = ref(false)
const techPanelMounted = ref(false)
const techPanelExpanded = ref(false)
let techPanelTimer: ReturnType<typeof setTimeout> | null = null
let transitionVersion = 0

const legalItems = [
  { key: 'user-agreement', title: '用户协议', desc: '查看平台服务条款' },
  { key: 'privacy-policy', title: '隐私政策', desc: '查看个人信息处理规则' }
]

function clearTechPanelTimer() {
  if (techPanelTimer) {
    clearTimeout(techPanelTimer)
    techPanelTimer = null
  }
}

async function toggleTechInfo() {
  const token = transitionVersion

  if (showTechInfo.value) {
    showTechInfo.value = false
    techPanelExpanded.value = false
    clearTechPanelTimer()
    techPanelTimer = setTimeout(() => {
      if (token !== transitionVersion) return
      techPanelMounted.value = false
      techPanelTimer = null
    }, TECH_PANEL_MS)
    return
  }
  showTechInfo.value = true
  techPanelMounted.value = true
  await nextTick()
  if (token !== transitionVersion || !showTechInfo.value) return
  clearTechPanelTimer()
  techPanelTimer = setTimeout(() => {
    if (token !== transitionVersion || !showTechInfo.value) return
    techPanelExpanded.value = true
    techPanelTimer = null
  }, 16)
}

function go(url: string) {
  uni.navigateTo({ url })
}

function onLegalSelect(key: string) {
  switch (key) {
    case 'user-agreement':
      go('/pages/legal/user-agreement')
      break
    case 'privacy-policy':
      go('/pages/legal/privacy-policy')
      break
  }
}

function resetTransientUI() {
  transitionVersion += 1
  clearTechPanelTimer()
  showTechInfo.value = false
  techPanelMounted.value = false
  techPanelExpanded.value = false
}

onHide(resetTransientUI)
onUnload(resetTransientUI)
</script>

<style scoped>
.about-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.version-grid,
.tech-panel {
  display: grid;
  gap: 14rpx;
}

.version-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
  padding-bottom: 14rpx;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.14);
}

.version-row:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.version-label {
  flex: 0 0 140rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}

.version-value {
  min-width: 0;
  color: var(--tree-text-primary);
  font-size: 24rpx;
  line-height: 1.45;
  text-align: right;
  overflow-wrap: anywhere;
}

.version-value.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.tech-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 88rpx;
  margin-top: 12rpx;
  padding: 16rpx 8rpx 4rpx;
  border-top: 1rpx solid rgba(148, 163, 184, 0.14);
  border-radius: 12rpx;
  transition: transform 180ms ease-out, background-color 180ms ease-out;
}

.tech-toggle:active {
  background-color: rgba(24, 54, 83, 0.05);
  transform: translateY(1rpx) scale(0.99);
}

.tech-toggle-label {
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}

.tech-toggle-arrow {
  color: var(--tree-green);
  font-size: 24rpx;
  font-weight: 600;
  transition: transform 180ms ease-out;
}

.tech-toggle:active .tech-toggle-arrow {
  transform: translateX(4rpx);
}

.tech-panel {
  margin-top: 16rpx;
  padding-top: 8rpx;
  opacity: 0;
  transform: translateY(8rpx);
  transition: opacity 200ms ease-out, transform 200ms ease-out;
}

.tech-panel--visible {
  opacity: 1;
  transform: translateY(0);
}
</style>
