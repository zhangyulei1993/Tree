<template>
  <view class="tree-page content-page">
    <view class="reading-hero">
      <view class="hero-text">
        <text class="hero-kicker">Tree Reading</text>
        <text class="hero-title">阅读</text>
        <text class="hero-desc">故事、典故、教程与宗亲文章，帮助你更好地整理家庭记忆。</text>
      </view>
      <view class="hero-book" aria-hidden="true">
        <view class="book-page book-page-left" />
        <view class="book-page book-page-right" />
      </view>
    </view>

    <scroll-view class="category-scroll" scroll-x>
      <view class="category-row">
        <view
          v-for="cat in categoryFilters"
          :key="cat.key"
          class="category-chip"
          :class="{ active: activeCategory === cat.key }"
          @click="activeCategory = cat.key"
        >
          <text>{{ cat.title }}</text>
        </view>
      </view>
    </scroll-view>

    <view
      v-if="spotlightArticle"
      class="spotlight-card"
      @click="goArticle(spotlightArticle)"
    >
      <view class="spotlight-art">
        <view class="art-line art-line-a" />
        <view class="art-line art-line-b" />
        <view class="art-dot art-dot-a" />
        <view class="art-dot art-dot-b" />
        <view class="art-dot art-dot-c" />
      </view>
      <view class="spotlight-copy">
        <view class="spotlight-meta">
          <view class="spotlight-badges">
            <text class="spotlight-badge">精选推荐</text>
            <text v-if="spotlightArticle.contentType === 'WECHAT_OFFICIAL'" class="spotlight-badge spotlight-badge-wechat">
              公众号
            </text>
          </view>
          <text class="spotlight-read">{{ readMinutes(spotlightArticle) }} 分钟</text>
        </view>
        <text class="spotlight-category">{{ spotlightArticle.categoryName }}</text>
        <text class="spotlight-title">{{ spotlightArticle.title }}</text>
        <text class="spotlight-summary">{{ spotlightArticle.summary }}</text>
      </view>
    </view>

    <view class="article-section">
      <view class="section-head">
        <text class="section-title">{{ listTitle }}</text>
        <text class="section-count">{{ articles.length }} 篇</text>
      </view>

      <view v-if="loading" class="content-state">
        <MiniEmptyState title="正在加载内容" description="正在读取阅读列表，请稍候。" />
      </view>

      <view v-else-if="error" class="content-state">
        <MiniNotice tone="warm" title="阅读列表加载失败">{{ error }}</MiniNotice>
        <MiniButton variant="secondary" class="retry-button" @click="loadArticles">重新加载</MiniButton>
      </view>

      <view v-else-if="articles.length === 0" class="content-state">
        <MiniEmptyState title="暂无内容" description="当前分类暂时没有已发布文章。" />
      </view>

      <template v-else>
        <view v-if="displayArticles.length === 0" class="content-note">
          <text>当前分类暂无更多文章，可先阅读上方精选内容。</text>
        </view>
        <view
          v-for="article in displayArticles"
          :key="article.id"
          class="article-card"
          @click="goArticle(article)"
        >
          <view class="article-leading" :class="`article-leading-${article.categoryKey}`">
            <text>{{ categoryShort(article.categoryKey) }}</text>
          </view>
          <view class="article-copy">
            <view class="article-meta">
              <view class="article-tags">
                <text class="article-tag">{{ article.categoryName }}</text>
                <text v-if="article.contentType === 'WECHAT_OFFICIAL'" class="article-tag article-tag-wechat">公众号</text>
              </view>
              <text class="article-read">{{ readMinutes(article) }} 分钟</text>
            </view>
            <text class="article-title">{{ article.title }}</text>
            <text class="article-summary">{{ article.summary }}</text>
          </view>
        </view>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listContentArticles, listContentCategories } from '@/api/content'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { openWechatOfficialArticle } from '@/features/content/wechatOfficialArticle'
import type { ContentArticleSummary, ContentCategory } from '@/types/api'

const activeCategory = ref('all')
const categories = ref<ContentCategory[]>([])
const articles = ref<ContentArticleSummary[]>([])
const spotlightArticle = ref<ContentArticleSummary | null>(null)
const loading = ref(false)
const error = ref('')

const categoryFilters = computed(() => [
  { key: 'all', title: '全部' },
  ...categories.value.map((item) => ({ key: item.key, title: item.name }))
])

const displayArticles = computed(() => articles.value.filter((item) => item.id !== spotlightArticle.value?.id))

const listTitle = computed(() => {
  if (activeCategory.value === 'all') return '精选内容'
  return categories.value.find((item) => item.key === activeCategory.value)?.name || '内容列表'
})

onMounted(async () => {
  await Promise.all([loadCategories(), loadArticles()])
})

watch(activeCategory, () => {
  void loadArticles()
})

async function loadCategories() {
  try {
    categories.value = await listContentCategories()
  } catch (err) {
    error.value = apiErrorMessage(err)
  }
}

async function loadArticles() {
  loading.value = true
  error.value = ''
  try {
    const result = await listContentArticles({
      categoryKey: activeCategory.value === 'all' ? undefined : activeCategory.value,
      page: 1,
      pageSize: 50
    })
    articles.value = result.items
    spotlightArticle.value = result.items.find((item) => item.isFeatured) || result.items[0] || null
  } catch (err) {
    error.value = apiErrorMessage(err)
    articles.value = []
    spotlightArticle.value = null
  } finally {
    loading.value = false
  }
}

function categoryShort(key: string) {
  const map: Record<string, string> = {
    tutorial: '教',
    story: '故',
    surname: '姓',
    article: '文'
  }
  return map[key] || '文'
}

function readMinutes(article: ContentArticleSummary) {
  const chars = `${article.title}${article.summary || ''}`.replace(/\s/g, '').length
  return Math.max(1, Math.round(chars / 400))
}

function articleUrl(id: number | string) {
  const article = articles.value.find((item) => item.id === id)
  return `/pages/content/detail?id=${encodeURIComponent(article?.slug || String(id))}`
}

async function goArticle(article: ContentArticleSummary) {
  if (article.contentType === 'WECHAT_OFFICIAL') {
    try {
      await openWechatOfficialArticle(article.externalUrl || '')
    } catch (err) {
      uni.showToast({ title: err instanceof Error ? err.message : '公众号文章暂时无法打开', icon: 'none' })
    }
    return
  }
  uni.navigateTo({ url: articleUrl(article.id) })
}
</script>

<style scoped>
.content-page {
  background:
    radial-gradient(circle at 92% 5%, rgba(47, 107, 87, 0.08), transparent 240rpx),
    radial-gradient(circle at 8% 30%, rgba(31, 58, 95, 0.06), transparent 260rpx);
}

.reading-hero {
  position: relative;
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin-bottom: 22rpx;
  border: 1rpx solid rgba(148, 163, 184, 0.14);
  border-radius: 32rpx;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.96) 0%, rgba(243, 250, 247, 0.96) 56%, rgba(238, 246, 252, 0.94) 100%);
  padding: 32rpx 28rpx;
  overflow: hidden;
  box-shadow: 0 16rpx 44rpx rgba(31, 58, 95, 0.07);
}

.hero-text {
  position: relative;
  z-index: 1;
  flex: 1;
  min-width: 0;
}

.hero-kicker {
  display: block;
  color: var(--tree-green);
  font-size: 20rpx;
  font-weight: 800;
  letter-spacing: 3rpx;
}

.hero-title {
  display: block;
  margin-top: 10rpx;
  color: var(--tree-text-primary);
  font-size: 46rpx;
  font-weight: 800;
  line-height: 1.2;
}

.hero-desc {
  display: block;
  margin-top: 12rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.58;
}

.hero-book {
  position: relative;
  width: 132rpx;
  height: 130rpx;
  flex-shrink: 0;
}

.book-page {
  position: absolute;
  top: 22rpx;
  width: 54rpx;
  height: 82rpx;
  border: 2rpx solid rgba(31, 58, 95, 0.18);
  background: rgba(255, 255, 255, 0.76);
  box-shadow: 0 8rpx 22rpx rgba(31, 58, 95, 0.06);
}

.book-page-left {
  left: 16rpx;
  border-radius: 16rpx 8rpx 8rpx 16rpx;
  transform: rotate(-8deg);
}

.book-page-right {
  right: 16rpx;
  border-radius: 8rpx 16rpx 16rpx 8rpx;
  transform: rotate(8deg);
}

.category-scroll {
  width: 100%;
  margin-bottom: 20rpx;
  white-space: nowrap;
}

.category-row {
  display: inline-flex;
  gap: 12rpx;
  padding: 2rpx 2rpx 4rpx;
}

.category-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 126rpx;
  border: 1rpx solid rgba(148, 163, 184, 0.16);
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.82);
  padding: 14rpx 22rpx;
  box-shadow: 0 6rpx 18rpx rgba(31, 58, 95, 0.035);
}

.category-chip.active {
  border-color: rgba(47, 107, 87, 0.18);
  background: linear-gradient(135deg, rgba(31, 58, 95, 0.95) 0%, rgba(47, 107, 87, 0.92) 100%);
}

.category-chip text {
  color: var(--tree-text-secondary);
  font-size: 23rpx;
  font-weight: 700;
  white-space: nowrap;
}

.category-chip.active text {
  color: #fff;
}

.spotlight-card {
  position: relative;
  margin-bottom: 22rpx;
  border-radius: 32rpx;
  background:
    linear-gradient(142deg, rgba(31, 58, 95, 0.96) 0%, rgba(47, 107, 87, 0.92) 100%);
  padding: 28rpx;
  overflow: hidden;
  box-shadow: 0 18rpx 46rpx rgba(31, 58, 95, 0.12);
}

.spotlight-card:active,
.article-card:active {
  transform: scale(0.992);
  opacity: 0.96;
}

.navigator-hover {
  opacity: 0.92;
}

.spotlight-art {
  position: absolute;
  right: -8rpx;
  top: 14rpx;
  width: 180rpx;
  height: 160rpx;
  opacity: 0.48;
}

.art-line {
  position: absolute;
  height: 3rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.34);
}

.art-line-a {
  top: 64rpx;
  left: 42rpx;
  width: 84rpx;
  transform: rotate(24deg);
}

.art-line-b {
  top: 92rpx;
  left: 42rpx;
  width: 84rpx;
  transform: rotate(-24deg);
}

.art-dot {
  position: absolute;
  border-radius: 50%;
  border: 3rpx solid rgba(255, 255, 255, 0.78);
  background: rgba(255, 255, 255, 0.12);
}

.art-dot-a {
  left: 24rpx;
  top: 64rpx;
  width: 22rpx;
  height: 22rpx;
}

.art-dot-b {
  right: 40rpx;
  top: 34rpx;
  width: 28rpx;
  height: 28rpx;
}

.art-dot-c {
  right: 34rpx;
  bottom: 28rpx;
  width: 20rpx;
  height: 20rpx;
}

.spotlight-copy {
  position: relative;
  z-index: 1;
  max-width: 500rpx;
}

.spotlight-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.spotlight-badge {
  display: inline-flex;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.13);
  color: rgba(255, 255, 255, 0.82);
  padding: 6rpx 14rpx;
  font-size: 20rpx;
  font-weight: 800;
}

.spotlight-badges,
.article-tags {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.spotlight-badge-wechat {
  background: rgba(82, 196, 26, 0.18);
  color: rgba(236, 255, 229, 0.92);
}

.spotlight-read {
  color: rgba(255, 255, 255, 0.62);
  font-size: 21rpx;
}

.spotlight-category {
  display: block;
  margin-top: 18rpx;
  color: rgba(255, 255, 255, 0.7);
  font-size: 22rpx;
  font-weight: 700;
}

.spotlight-title {
  display: block;
  margin-top: 10rpx;
  color: #fff;
  font-size: 35rpx;
  font-weight: 800;
  line-height: 1.34;
}

.spotlight-summary {
  display: block;
  margin-top: 12rpx;
  color: rgba(255, 255, 255, 0.76);
  font-size: 24rpx;
  line-height: 1.62;
}

.article-section {
  margin-bottom: 16rpx;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14rpx;
  padding: 0 4rpx;
}

.section-title {
  color: var(--tree-text-primary);
  font-size: 31rpx;
  font-weight: 800;
}

.section-count {
  color: var(--tree-text-weak);
  font-size: 21rpx;
}

.content-state {
  margin-top: 14rpx;
  border: 1rpx solid rgba(148, 163, 184, 0.14);
  border-radius: 28rpx;
  background: rgba(255, 255, 255, 0.9);
  padding: 16rpx;
  box-shadow: 0 10rpx 28rpx rgba(31, 58, 95, 0.04);
}

.retry-button {
  margin-top: 16rpx;
}

.content-note {
  border: 1rpx solid rgba(216, 229, 220, 0.9);
  border-radius: 24rpx;
  background: rgba(244, 248, 245, 0.9);
  padding: 22rpx;
}

.content-note text {
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.7;
}

.article-card {
  display: flex;
  gap: 18rpx;
  margin-bottom: 14rpx;
  border: 1rpx solid rgba(148, 163, 184, 0.14);
  border-radius: 28rpx;
  background: rgba(255, 255, 255, 0.94);
  padding: 22rpx;
  box-shadow: 0 10rpx 28rpx rgba(31, 58, 95, 0.045);
}

.article-leading {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 58rpx;
  height: 58rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
}

.article-leading text {
  font-size: 24rpx;
  font-weight: 800;
}

.article-leading-tutorial {
  background: #eef4fb;
  color: var(--tree-primary);
}

.article-leading-story {
  background: #fff2f7;
  color: #9f5872;
}

.article-leading-surname {
  background: #fff8e8;
  color: #936f2f;
}

.article-leading-article {
  background: #edf8f2;
  color: var(--tree-green);
}

.article-copy {
  flex: 1;
  min-width: 0;
}

.article-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  margin-bottom: 7rpx;
}

.article-tag {
  color: var(--tree-text-secondary);
  font-size: 21rpx;
  font-weight: 700;
}

.article-tag-wechat {
  border-radius: 999rpx;
  background: rgba(47, 107, 87, 0.1);
  color: var(--tree-green);
  padding: 3rpx 10rpx;
}

.article-read {
  color: var(--tree-text-weak);
  font-size: 20rpx;
}

.article-title {
  display: block;
  color: var(--tree-text-primary);
  font-size: 28rpx;
  font-weight: 800;
  line-height: 1.4;
}

.article-summary {
  display: block;
  margin-top: 7rpx;
  color: var(--tree-text-secondary);
  font-size: 23rpx;
  line-height: 1.55;
}
</style>
