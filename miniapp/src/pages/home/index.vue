<template>
  <view class="tree-page home-page">
    <view class="hero-card">
      <view class="hero-copy">
        <text class="hero-brand">Tree</text>
        <view class="hero-title">
          <text>把家人关系</text>
          <text>整理清楚</text>
        </view>
        <text class="hero-subtitle">成员、关系、故事与邀请，放在一个清晰的家庭空间。</text>
        <view class="hero-badges">
          <text class="hero-badge">私密协作</text>
          <text class="hero-badge">审核公开</text>
        </view>
      </view>
      <view class="hero-visual" aria-hidden="true">
        <view class="visual-ring visual-ring-a" />
        <view class="visual-ring visual-ring-b" />
        <view class="visual-line visual-line-a" />
        <view class="visual-line visual-line-b" />
        <view class="visual-line visual-line-c" />
        <view class="visual-node visual-node-core" />
        <view class="visual-node visual-node-top" />
        <view class="visual-node visual-node-left" />
        <view class="visual-node visual-node-right" />
        <view class="visual-node visual-node-bottom" />
      </view>
    </view>

    <view class="primary-grid">
      <navigator
        class="primary-card primary-card-main"
        hover-class="navigator-hover"
        url="/pages/family/my"
      >
        <view class="primary-card-glow" />
        <view class="primary-card-content">
          <view class="primary-icon primary-icon-home" />
          <text class="primary-title">我的家庭</text>
          <text class="primary-desc">查看成员、关系和私有家谱</text>
        </view>
        <text class="primary-arrow">›</text>
      </navigator>
      <navigator
        class="primary-card primary-card-light"
        hover-class="navigator-hover"
        open-type="switchTab"
        url="/pages/family/search"
      >
        <view class="primary-card-content">
          <view class="primary-icon primary-icon-search" />
          <text class="primary-title">寻找家族</text>
          <text class="primary-desc">浏览公开主页</text>
        </view>
        <text class="primary-arrow">›</text>
      </navigator>
    </view>

    <view class="dashboard-card">
      <view class="section-row">
        <view>
          <text class="section-kicker">今日事项</text>
          <text class="section-title">处理邀请与申请</text>
        </view>
        <text class="section-note">保持家庭资料同步</text>
      </view>
      <view class="task-grid">
        <navigator class="task-card" hover-class="navigator-hover" url="/pages/invite/my">
          <view class="task-mark task-mark-blue" />
          <text class="task-title">我的邀请</text>
          <text class="task-desc">查看收到的家庭邀请</text>
        </navigator>
        <navigator class="task-card" hover-class="navigator-hover" url="/pages/join/my">
          <view class="task-mark task-mark-green" />
          <text class="task-title">加入申请</text>
          <text class="task-desc">查看申请处理进度</text>
        </navigator>
      </view>
    </view>

    <navigator class="kinship-card" hover-class="navigator-hover" url="/pages/tools/kinship">
      <view class="kinship-orb">
        <view class="kinship-dot kinship-dot-a" />
        <view class="kinship-dot kinship-dot-b" />
        <view class="kinship-dot kinship-dot-c" />
      </view>
      <view class="kinship-copy">
        <text class="kinship-title">亲属称谓工具</text>
        <text class="kinship-desc">按关系路径推测常见称呼</text>
      </view>
      <text class="kinship-arrow">›</text>
    </navigator>

    <view class="read-card">
      <navigator
        class="section-row read-head"
        hover-class="navigator-hover-soft"
        open-type="switchTab"
        url="/pages/content/index"
      >
        <view>
          <text class="section-kicker">阅读精选</text>
          <text class="section-title">故事、典故与使用指南</text>
        </view>
        <text class="read-more">进入阅读</text>
      </navigator>

      <navigator
        v-if="featuredRead"
        class="featured-read"
        hover-class="navigator-hover"
        :url="articleUrl(featuredRead.id)"
      >
        <text class="featured-label">{{ categoryTitle(featuredRead.category) }}</text>
        <text class="featured-title">{{ featuredRead.title }}</text>
        <text class="featured-summary">{{ featuredRead.summary }}</text>
      </navigator>

      <view class="read-mini-grid">
        <navigator
          v-for="item in secondaryReadHighlights"
          :key="item.category.key"
          class="read-mini"
          :class="`read-mini-${item.category.key}`"
          hover-class="navigator-hover"
          :open-type="item.article ? 'navigate' : 'switchTab'"
          :url="item.article ? articleUrl(item.article.id) : '/pages/content/index'"
        >
          <text class="read-mini-label">{{ item.category.title }}</text>
          <text class="read-mini-title">{{ item.article?.title || item.category.desc }}</text>
        </navigator>
      </view>
    </view>

    <navigator
      class="public-card"
      hover-class="navigator-hover"
      open-type="switchTab"
      url="/pages/family/search"
    >
      <view class="public-copy">
        <text class="section-kicker">公开家族</text>
        <text class="public-title">看看别人如何展示家族主页</text>
        <text class="public-desc">浏览已审核公开的家族简介、公开树与留言。</text>
      </view>
      <view class="public-avatar">
        <text>张</text>
      </view>
    </navigator>

    <view class="home-privacy">
      <text>家庭资料仅在授权范围内可见，公开展示需经过审核。</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { getCategoryMeta, getHomeReadHighlights, type ContentCategory } from '@/mock/content'

const readHighlights = getHomeReadHighlights()
const featuredRead = readHighlights.find((item) => item.category.key === 'tutorial')?.article || null
const secondaryReadHighlights = readHighlights.filter((item) => item.article?.id !== featuredRead?.id)

function categoryTitle(key: ContentCategory) {
  return getCategoryMeta(key)?.title || '内容'
}

function articleUrl(id: string) {
  return `/pages/content/detail?id=${encodeURIComponent(id)}`
}
</script>

<style scoped>
.home-page {
  background:
    radial-gradient(circle at 90% 4%, rgba(47, 107, 87, 0.08), transparent 240rpx),
    radial-gradient(circle at 10% 18%, rgba(39, 76, 119, 0.07), transparent 260rpx);
}

.hero-card {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 230rpx;
  margin-bottom: 24rpx;
  border: 1rpx solid rgba(148, 163, 184, 0.16);
  border-radius: 32rpx;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.96) 0%, rgba(242, 250, 247, 0.94) 52%, rgba(238, 246, 252, 0.94) 100%);
  padding: 34rpx 30rpx;
  overflow: hidden;
  box-shadow: 0 18rpx 48rpx rgba(31, 58, 95, 0.08);
}

.hero-card::after {
  content: '';
  position: absolute;
  right: -40rpx;
  bottom: -80rpx;
  width: 260rpx;
  height: 260rpx;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(47, 107, 87, 0.12), transparent 68%);
}

.hero-copy {
  position: relative;
  z-index: 1;
  flex: 1;
  min-width: 0;
}

.hero-brand {
  display: block;
  color: var(--tree-green);
  font-size: 22rpx;
  font-weight: 800;
  letter-spacing: 5rpx;
}

.hero-title {
  display: flex;
  flex-direction: column;
  margin-top: 14rpx;
  max-width: 410rpx;
  color: var(--tree-text-primary);
  font-size: 45rpx;
  font-weight: 800;
  line-height: 1.22;
}

.hero-title text {
  display: block;
}

.hero-subtitle {
  display: block;
  margin-top: 14rpx;
  max-width: 400rpx;
  color: var(--tree-text-secondary);
  font-size: 25rpx;
  line-height: 1.55;
}

.hero-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 22rpx;
}

.hero-badge {
  border: 1rpx solid rgba(47, 107, 87, 0.14);
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.72);
  color: var(--tree-green);
  padding: 7rpx 15rpx;
  font-size: 20rpx;
  font-weight: 600;
}

.hero-visual {
  position: relative;
  z-index: 1;
  width: 160rpx;
  height: 170rpx;
  flex-shrink: 0;
}

.visual-ring {
  position: absolute;
  border-radius: 50%;
}

.visual-ring-a {
  inset: 8rpx 0 2rpx 0;
  border: 2rpx dashed rgba(31, 58, 95, 0.14);
}

.visual-ring-b {
  top: 50rpx;
  left: 46rpx;
  width: 66rpx;
  height: 66rpx;
  border: 1rpx solid rgba(47, 107, 87, 0.12);
  background: rgba(255, 255, 255, 0.56);
}

.visual-line {
  position: absolute;
  height: 3rpx;
  border-radius: 999rpx;
  background: rgba(100, 116, 139, 0.22);
  transform-origin: center;
}

.visual-line-a {
  top: 48rpx;
  left: 78rpx;
  width: 42rpx;
  transform: rotate(88deg);
}

.visual-line-b {
  top: 92rpx;
  left: 38rpx;
  width: 56rpx;
  transform: rotate(-28deg);
}

.visual-line-c {
  top: 96rpx;
  right: 26rpx;
  width: 54rpx;
  transform: rotate(28deg);
}

.visual-node {
  position: absolute;
  border-radius: 50%;
  background: #fff;
  box-sizing: border-box;
}

.visual-node-core {
  top: 66rpx;
  left: 64rpx;
  width: 42rpx;
  height: 42rpx;
  border: 4rpx solid var(--tree-primary);
  box-shadow: 0 0 0 10rpx rgba(31, 58, 95, 0.08);
}

.visual-node-top {
  top: 22rpx;
  left: 76rpx;
  width: 18rpx;
  height: 18rpx;
  border: 3rpx solid var(--tree-green);
}

.visual-node-left,
.visual-node-right,
.visual-node-bottom {
  width: 20rpx;
  height: 20rpx;
}

.visual-node-left {
  left: 24rpx;
  top: 116rpx;
  border: 3rpx solid var(--tree-accent-blue);
}

.visual-node-right {
  right: 16rpx;
  top: 106rpx;
  border: 3rpx solid #7fb7a4;
}

.visual-node-bottom {
  left: 76rpx;
  bottom: 8rpx;
  border: 3rpx solid #d6af68;
}

.primary-grid {
  display: grid;
  grid-template-columns: 1.08fr 0.92fr;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.primary-card {
  position: relative;
  min-height: 158rpx;
  border-radius: 30rpx;
  padding: 24rpx;
  overflow: hidden;
  box-sizing: border-box;
  box-shadow: 0 14rpx 34rpx rgba(31, 58, 95, 0.08);
}

.primary-card:active,
.task-card:active,
.kinship-card:active,
.featured-read:active,
.read-mini:active,
.public-card:active {
  transform: scale(0.992);
  opacity: 0.96;
}

.navigator-hover {
  opacity: 0.92;
}

.navigator-hover-soft {
  opacity: 0.86;
}

.primary-card-main {
  background: linear-gradient(135deg, #1f3a5f 0%, #28605b 100%);
  color: #fff;
}

.primary-card-light {
  border: 1rpx solid rgba(47, 107, 87, 0.15);
  background: linear-gradient(145deg, #fff 0%, #f2faf7 100%);
}

.primary-card-glow {
  position: absolute;
  right: -46rpx;
  top: -40rpx;
  width: 170rpx;
  height: 170rpx;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(214, 175, 104, 0.32), transparent 66%);
}

.primary-card-content {
  position: relative;
  z-index: 1;
}

.primary-icon {
  position: relative;
  width: 52rpx;
  height: 52rpx;
  margin-bottom: 20rpx;
  border-radius: 16rpx;
}

.primary-icon-home {
  background: rgba(255, 255, 255, 0.16);
}

.primary-icon-home::before,
.primary-icon-search::before,
.primary-icon-search::after {
  content: '';
  position: absolute;
  box-sizing: border-box;
}

.primary-icon-home::before {
  left: 14rpx;
  top: 15rpx;
  width: 24rpx;
  height: 18rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.74);
  border-top-left-radius: 4rpx;
  border-top-right-radius: 4rpx;
}

.primary-icon-search {
  background: rgba(47, 107, 87, 0.1);
}

.primary-icon-search::before {
  top: 14rpx;
  left: 14rpx;
  width: 22rpx;
  height: 22rpx;
  border: 3rpx solid var(--tree-green);
  border-radius: 50%;
}

.primary-icon-search::after {
  top: 33rpx;
  left: 33rpx;
  width: 12rpx;
  height: 3rpx;
  border-radius: 999rpx;
  background: var(--tree-green);
  transform: rotate(45deg);
}

.primary-title {
  display: block;
  color: inherit;
  font-size: 31rpx;
  font-weight: 800;
  line-height: 1.32;
}

.primary-card-light .primary-title {
  color: var(--tree-text-primary);
}

.primary-desc {
  display: block;
  margin-top: 7rpx;
  color: rgba(255, 255, 255, 0.76);
  font-size: 22rpx;
  line-height: 1.45;
}

.primary-card-light .primary-desc {
  color: var(--tree-text-secondary);
}

.primary-arrow {
  position: absolute;
  right: 24rpx;
  top: 26rpx;
  color: rgba(255, 255, 255, 0.72);
  font-size: 34rpx;
  font-weight: 300;
}

.primary-card-light .primary-arrow {
  color: var(--tree-text-weak);
}

.dashboard-card,
.read-card,
.public-card,
.kinship-card {
  margin-bottom: 20rpx;
  border: 1rpx solid rgba(148, 163, 184, 0.14);
  border-radius: 30rpx;
  background: rgba(255, 255, 255, 0.92);
  padding: 26rpx;
  box-shadow: 0 10rpx 34rpx rgba(31, 58, 95, 0.055);
}

.section-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18rpx;
}

.section-kicker {
  display: block;
  color: var(--tree-green);
  font-size: 21rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}

.section-title {
  display: block;
  margin-top: 7rpx;
  color: var(--tree-text-primary);
  font-size: 31rpx;
  font-weight: 800;
  line-height: 1.35;
}

.section-note,
.read-more {
  flex-shrink: 0;
  margin-top: 8rpx;
  color: var(--tree-text-secondary);
  font-size: 22rpx;
}

.read-more {
  color: var(--tree-green);
  font-weight: 700;
}

.task-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14rpx;
  margin-top: 22rpx;
}

.task-card {
  border-radius: 22rpx;
  background: var(--tree-surface-muted);
  padding: 20rpx;
  box-sizing: border-box;
}

.task-mark {
  width: 34rpx;
  height: 34rpx;
  margin-bottom: 18rpx;
  border-radius: 12rpx;
}

.task-mark-blue {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
}

.task-mark-green {
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
}

.task-title {
  display: block;
  color: var(--tree-text-primary);
  font-size: 26rpx;
  font-weight: 800;
}

.task-desc {
  display: block;
  margin-top: 7rpx;
  color: var(--tree-text-secondary);
  font-size: 21rpx;
  line-height: 1.45;
}

.kinship-card {
  display: flex;
  align-items: center;
  gap: 18rpx;
  background:
    linear-gradient(135deg, rgba(245, 247, 255, 0.98) 0%, rgba(255, 255, 255, 0.94) 55%, rgba(243, 250, 247, 0.96) 100%);
}

.kinship-orb {
  position: relative;
  width: 70rpx;
  height: 70rpx;
  border-radius: 22rpx;
  background: rgba(31, 58, 95, 0.08);
  flex-shrink: 0;
}

.kinship-dot {
  position: absolute;
  border-radius: 50%;
  background: #fff;
  box-sizing: border-box;
}

.kinship-dot-a {
  top: 16rpx;
  left: 26rpx;
  width: 18rpx;
  height: 18rpx;
  border: 3rpx solid var(--tree-primary);
}

.kinship-dot-b,
.kinship-dot-c {
  bottom: 16rpx;
  width: 14rpx;
  height: 14rpx;
  border: 3rpx solid var(--tree-green);
}

.kinship-dot-b {
  left: 15rpx;
}

.kinship-dot-c {
  right: 15rpx;
}

.kinship-copy {
  flex: 1;
  min-width: 0;
}

.kinship-title {
  display: block;
  color: var(--tree-text-primary);
  font-size: 28rpx;
  font-weight: 800;
}

.kinship-desc {
  display: block;
  margin-top: 6rpx;
  color: var(--tree-text-secondary);
  font-size: 22rpx;
}

.kinship-arrow {
  color: var(--tree-text-weak);
  font-size: 34rpx;
}

.read-head {
  margin-bottom: 20rpx;
}

.featured-read {
  position: relative;
  margin-bottom: 14rpx;
  border-radius: 26rpx;
  background:
    linear-gradient(140deg, rgba(31, 58, 95, 0.94) 0%, rgba(47, 107, 87, 0.92) 100%);
  padding: 24rpx;
  overflow: hidden;
}

.featured-read::after {
  content: '';
  position: absolute;
  right: -32rpx;
  bottom: -48rpx;
  width: 170rpx;
  height: 170rpx;
  border-radius: 50%;
  border: 1rpx solid rgba(255, 255, 255, 0.16);
}

.featured-label {
  position: relative;
  z-index: 1;
  display: inline-flex;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.13);
  color: rgba(255, 255, 255, 0.8);
  padding: 5rpx 13rpx;
  font-size: 20rpx;
  font-weight: 700;
}

.featured-title {
  position: relative;
  z-index: 1;
  display: block;
  margin-top: 16rpx;
  color: #fff;
  font-size: 30rpx;
  font-weight: 800;
  line-height: 1.38;
}

.featured-summary {
  position: relative;
  z-index: 1;
  display: block;
  margin-top: 8rpx;
  color: rgba(255, 255, 255, 0.76);
  font-size: 22rpx;
  line-height: 1.55;
}

.read-mini-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14rpx;
}

.read-mini {
  min-height: 128rpx;
  border-radius: 24rpx;
  padding: 18rpx;
  box-sizing: border-box;
}

.read-mini-story {
  background: #fff2f7;
}

.read-mini-surname {
  background: #fff8e8;
}

.read-mini-article {
  background: #edf8f2;
}

.read-mini-tutorial {
  background: #eef4fb;
}

.read-mini-label {
  display: block;
  color: var(--tree-text-secondary);
  font-size: 20rpx;
  font-weight: 800;
}

.read-mini-title {
  display: block;
  margin-top: 9rpx;
  color: var(--tree-text-primary);
  font-size: 24rpx;
  font-weight: 700;
  line-height: 1.45;
}

.public-card {
  display: flex;
  align-items: center;
  gap: 18rpx;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.96) 0%, rgba(248, 250, 252, 0.92) 100%);
}

.public-copy {
  flex: 1;
  min-width: 0;
}

.public-title {
  display: block;
  margin-top: 7rpx;
  color: var(--tree-text-primary);
  font-size: 28rpx;
  font-weight: 800;
  line-height: 1.4;
}

.public-desc {
  display: block;
  margin-top: 6rpx;
  color: var(--tree-text-secondary);
  font-size: 22rpx;
  line-height: 1.5;
}

.public-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80rpx;
  height: 80rpx;
  border-radius: 26rpx;
  background: linear-gradient(145deg, #eff6ff 0%, #e6f6ee 100%);
  color: var(--tree-primary);
  font-size: 32rpx;
  font-weight: 800;
  flex-shrink: 0;
}

.home-privacy {
  padding: 4rpx 18rpx 18rpx;
  text-align: center;
}

.home-privacy text {
  color: var(--tree-text-weak);
  font-size: 20rpx;
  line-height: 1.7;
}
</style>
