<template>
  <view class="archive-page about-page">
    <MiniBackHome />

    <view class="about-head archive-page-head">
      <view>
        <text class="archive-kicker">About Tree</text>
        <text class="archive-title">关于 Tree</text>
        <text class="archive-subtitle">协议、隐私与版本信息</text>
      </view>
      <view class="archive-seal">谱</view>
    </view>

    <view class="about-folio">
      <view class="about-mark">
        <text class="about-brand">Tree</text>
        <text class="about-product">{{ productName }}</text>
      </view>
      <view class="about-folio-copy">
        <text class="about-tagline">家庭族谱记录与协作平台</text>
        <text class="about-version">v{{ buildInfo.version }} · {{ buildInfo.environment }}</text>
        <text class="about-intro">
          帮助家庭记录成员关系、维护族谱档案，并在授权后对外展示经审核的家族信息。
        </text>
      </view>
    </view>

    <view class="about-section archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">功能说明</text>
        <text class="archive-section-subtitle">Tree 小程序主要能力</text>
      </view>
      <view class="about-list archive-list">
        <view v-for="item in featureItems" :key="item.title" class="archive-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">{{ item.title }}</text>
            <text class="archive-row-desc">{{ item.desc }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="about-section archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">合规说明</text>
        <text class="archive-section-subtitle">隐私保护与公开展示原则</text>
      </view>
      <view class="about-list archive-list">
        <view v-for="item in complianceItems" :key="item.title" class="archive-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">{{ item.title }}</text>
            <text class="archive-row-desc">{{ item.desc }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="about-section archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">联系方式</text>
        <text class="archive-section-subtitle">产品咨询与隐私相关联络</text>
      </view>
      <view class="about-list archive-list">
        <view v-for="item in contactItems" :key="item.title" class="archive-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">{{ item.title }}</text>
            <text class="archive-row-desc">{{ item.desc }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="about-section archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">法律文件</text>
        <text class="archive-section-subtitle">查看完整协议与隐私政策</text>
      </view>
      <view class="about-list archive-list">
        <view
          v-for="item in legalItems"
          :key="item.key"
          class="archive-row"
          @click="onLegalSelect(item.key)"
        >
          <view class="archive-row-main">
            <text class="archive-row-title">{{ item.title }}</text>
            <text class="archive-row-desc">{{ item.desc }}</text>
          </view>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </view>

    <view class="about-version archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">版本信息</text>
        <text class="archive-section-subtitle">当前小程序运行环境</text>
      </view>
      <view class="version-grid">
        <view class="tree-info-row">
          <text class="tree-info-label">小程序版本</text>
          <text class="tree-info-value">{{ buildInfo.version }}</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">当前环境</text>
          <text class="tree-info-value">{{ buildInfo.environment }}</text>
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
        <view class="tree-info-row">
          <text class="tree-info-label">提交版本</text>
          <text class="tree-info-value mono">{{ buildInfo.commit }}{{ buildInfo.dirty ? ' dirty' : '' }}</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">接口地址</text>
          <text class="tree-info-value mono">{{ buildInfo.apiBaseUrl || '未配置' }}</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">构建时间</text>
          <text class="tree-info-value mono">{{ buildInfo.buildTime || '—' }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onHide, onUnload } from '@dcloudio/uni-app'
import { nextTick, ref } from 'vue'

import { buildInfo } from '@/buildInfo'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import { LEGAL_OPERATOR, LEGAL_PRODUCT_NAME } from '@/features/legal/legalMeta'

const TECH_PANEL_MS = 200

const productName = LEGAL_PRODUCT_NAME

const showTechInfo = ref(false)
const techPanelMounted = ref(false)
const techPanelExpanded = ref(false)
let techPanelTimer: ReturnType<typeof setTimeout> | null = null
let transitionVersion = 0

const featureItems = [
  {
    title: '家庭族谱',
    desc: '创建家庭、录入成员与关系，维护私密家谱结构'
  },
  {
    title: '邀请与加入',
    desc: '通过邀请链接或申请流程加入家庭，协作维护族谱'
  },
  {
    title: '公开展示',
    desc: '经审核后可对外展示脱敏后的家族简介与公开家谱'
  }
]

const complianceItems = [
  {
    title: '协议与隐私',
    desc: '用户协议与隐私政策遵循现行法律文本，详见下方法律文件入口'
  },
  {
    title: '信息处理',
    desc: '个人信息处理范围、保存期限与权利行使方式见隐私政策'
  },
  {
    title: '公开脱敏',
    desc: '公开展示的信息经审核与脱敏处理，不默认公开在世成员详细资料'
  }
]

const contactItems = [
  {
    title: '隐私联系邮箱',
    desc: LEGAL_OPERATOR.privacyEmail
  },
  {
    title: '联系电话',
    desc: LEGAL_OPERATOR.phone
  },
  {
    title: '运营主体',
    desc: LEGAL_OPERATOR.name
  }
]

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
  padding-top: 28rpx;
}

.about-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.about-folio {
  display: flex;
  align-items: stretch;
  gap: 24rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.42), transparent 38%),
    rgba(255, 252, 245, 0.52);
  padding: 24rpx 0;
}

.about-mark {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 128rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-line-strong);
  background:
    radial-gradient(circle at 35% 28%, rgba(255, 255, 255, 0.9), transparent 34rpx),
    #eadfc8;
  padding: 18rpx 8rpx;
}

.about-brand {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 34rpx;
  font-weight: 800;
  line-height: 1.2;
}

.about-product {
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 18rpx;
  line-height: 1.35;
  text-align: center;
}

.about-folio-copy {
  flex: 1;
  min-width: 0;
}

.about-tagline {
  display: block;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 34rpx;
  font-weight: 800;
  line-height: 1.35;
}

.about-version {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.4;
}

.about-intro {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.55;
}

.about-section,
.about-version {
  margin-bottom: 24rpx;
}

.about-list {
  margin-top: 8rpx;
}

.version-grid {
  display: grid;
  gap: 0;
}

.tech-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 72rpx;
  margin-top: 12rpx;
  padding: 12rpx 0 4rpx;
  border-top: 1rpx solid var(--archive-line);
}

.tech-toggle-label {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
}

.tech-toggle-arrow {
  color: var(--archive-blue);
  font-size: 24rpx;
  font-weight: 600;
}

.tech-panel {
  margin-top: 8rpx;
  opacity: 0;
  transform: translateY(8rpx);
  transition: opacity 200ms ease-out, transform 200ms ease-out;
}

.tech-panel--visible {
  opacity: 1;
  transform: translateY(0);
}

.tree-info-value.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 22rpx;
  overflow-wrap: anywhere;
}
</style>
