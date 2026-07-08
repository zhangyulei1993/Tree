<template>
  <view class="archive-page content-page">
    <view class="reading-head archive-page-head">
      <view>
        <text class="archive-kicker">Tree Reading</text>
        <text class="archive-title">阅读</text>
        <text class="archive-subtitle">故事、典故、教程与宗亲文章，帮助你整理家庭记忆。</text>
      </view>
      <view class="archive-seal">读</view>
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
      class="spotlight-card archive-panel"
      @click="goArticle(spotlightArticle)"
    >
      <view class="spotlight-spine">
        <text>卷</text>
        <text>首</text>
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
      <text class="archive-arrow spotlight-arrow">›</text>
    </view>

    <view class="article-section">
      <view class="section-head">
        <view>
          <text class="archive-kicker">Catalog</text>
          <text class="section-title">{{ listTitle }}</text>
        </view>
        <text class="section-count">{{ articles.length }} 篇</text>
      </view>

      <view v-if="loading" class="content-state archive-form-panel">
        <MiniEmptyState title="正在加载内容" description="正在读取阅读列表，请稍候。" />
      </view>

      <view v-else-if="error" class="content-state archive-form-panel">
        <MiniNotice tone="warm" title="阅读列表加载失败">{{ error }}</MiniNotice>
        <MiniButton variant="secondary" class="retry-button" @click="loadArticles">重新加载</MiniButton>
      </view>

      <view v-else-if="articles.length === 0" class="content-state archive-form-panel">
        <MiniEmptyState title="暂无内容" description="当前分类暂时没有已发布文章。" />
      </view>

      <template v-else>
        <view v-if="displayArticles.length === 0" class="content-note">
          <text>当前分类暂无更多文章，可先阅读上方精选内容。</text>
        </view>
        <view class="article-list archive-list">
          <view
            v-for="article in displayArticles"
            :key="article.id"
            class="article-row archive-row"
            @click="goArticle(article)"
          >
            <view class="article-mark" :class="`article-mark-${article.categoryKey}`">
              <text>{{ categoryShort(article.categoryKey) }}</text>
            </view>
            <view class="archive-row-main article-copy">
              <view class="article-meta">
                <view class="article-tags">
                  <text class="article-tag">{{ article.categoryName }}</text>
                  <text v-if="article.contentType === 'WECHAT_OFFICIAL'" class="article-tag article-tag-wechat">
                    公众号
                  </text>
                </view>
                <text class="article-read">{{ readMinutes(article) }} 分钟</text>
              </view>
              <text class="archive-row-title article-title">{{ article.title }}</text>
              <text class="archive-row-desc article-summary">{{ article.summary }}</text>
            </view>
            <text class="archive-arrow">›</text>
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
  padding-top: 28rpx;
}

.reading-head {
  margin-bottom: 18rpx;
}

.category-scroll {
  width: 100%;
  margin-bottom: 22rpx;
  white-space: nowrap;
}

.category-row {
  display: inline-flex;
  gap: 22rpx;
  padding: 0 2rpx 8rpx;
}

.category-chip {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 104rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding: 12rpx 4rpx 16rpx;
}

.category-chip.active {
  border-bottom-color: var(--archive-cinnabar);
}

.category-chip text {
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  font-weight: 650;
  white-space: nowrap;
}

.category-chip.active text {
  color: var(--archive-cinnabar);
}

.spotlight-card {
  position: relative;
  display: flex;
  gap: 24rpx;
  margin-bottom: 22rpx;
  padding: 24rpx 34rpx 24rpx 0;
  overflow: hidden;
}

.spotlight-card:active,
.article-row:active {
  transform: scale(0.992);
  opacity: 0.96;
}

.navigator-hover {
  opacity: 0.92;
}

.spotlight-spine {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6rpx;
  align-self: stretch;
  width: 58rpx;
  background: var(--archive-blue);
  color: rgba(255, 255, 255, 0.88);
  padding-top: 24rpx;
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 24rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.spotlight-copy {
  flex: 1;
  min-width: 0;
}

.spotlight-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.spotlight-badge {
  display: inline-flex;
  border: 1rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  padding: 5rpx 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 1rpx;
}

.spotlight-badges,
.article-tags {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.spotlight-badge-wechat {
  border-color: var(--archive-blue);
  color: var(--archive-blue);
}

.spotlight-read {
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.spotlight-category {
  display: block;
  margin-top: 18rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  font-weight: 650;
}

.spotlight-title {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 34rpx;
  font-weight: 700;
  line-height: 1.34;
}

.spotlight-summary {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.62;
}

.spotlight-arrow {
  position: absolute;
  right: 12rpx;
  top: 50%;
  margin-top: -18rpx;
}

.article-section {
  margin-bottom: 16rpx;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding: 0 0 16rpx;
}

.section-title {
  display: block;
  margin-top: 4rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 32rpx;
  font-weight: 700;
}

.section-count {
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.content-state {
  margin-top: 14rpx;
}

.retry-button {
  margin-top: 16rpx;
}

.content-note {
  border-top: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 22rpx;
}

.content-note text {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.7;
}

.article-row {
  align-items: flex-start;
  padding-top: 22rpx;
  padding-bottom: 22rpx;
}

.article-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 58rpx;
  height: 58rpx;
  border: 2rpx solid var(--archive-cinnabar);
  background: rgba(251, 246, 234, 0.62);
  color: var(--archive-cinnabar);
  flex-shrink: 0;
  transform: rotate(-2deg);
}

.article-mark text {
  color: inherit;
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 24rpx;
  font-weight: 700;
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
  margin-bottom: 8rpx;
}

.article-tag {
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  font-weight: 650;
}

.article-tag-wechat {
  color: var(--archive-blue);
}

.article-read {
  color: var(--archive-ink-soft);
  font-size: 20rpx;
  white-space: nowrap;
}

.article-title {
  display: block;
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 700;
  line-height: 1.4;
}

.article-summary {
  display: block;
  margin-top: 7rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.55;
}
</style>
