<template>
  <view class="archive-page content-detail-page">
    <MiniBackHome />

    <view v-if="loading" class="detail-state archive-form-panel">
      <MiniEmptyState title="正在加载" description="正在读取内容，请稍候。" />
    </view>

    <view v-else-if="!article" class="detail-state archive-form-panel">
      <MiniEmptyState title="内容不存在" :description="error || '请返回阅读页重新选择。'" />
      <MiniButton variant="secondary" class="retry-button" @click="goBack">返回阅读</MiniButton>
    </view>

    <template v-else>
      <view class="scroll-head archive-page-head">
        <view class="scroll-head-copy">
          <text class="archive-kicker">Tree Reading</text>
          <text class="scroll-title">{{ article.title }}</text>
          <text v-if="article.summary" class="scroll-lead">{{ article.summary }}</text>
        </view>
        <view class="scroll-seal archive-seal">{{ categoryShort(article.categoryKey) }}</view>
      </view>

      <view class="scroll-meta archive-panel">
        <view class="scroll-meta-row">
          <view class="scroll-badges">
            <text class="scroll-badge">{{ categoryTitle }}</text>
            <text v-if="article.contentType === 'WECHAT_OFFICIAL'" class="scroll-badge scroll-badge-wechat">
              公众号
            </text>
          </view>
          <text class="scroll-read">{{ readMinutes(article) }} 分钟</text>
        </view>
      </view>

      <view v-if="article.contentType === 'WECHAT_OFFICIAL'" class="scroll-external archive-form-panel">
        <text class="external-note">本文发布于微信公众号，将通过微信官方页面打开。</text>
        <MiniButton @click="openOfficialArticle">阅读公众号文章</MiniButton>
      </view>

      <view v-else class="scroll-body">
        <view class="scroll-body-frame">
          <view class="scroll-spine">
            <text>卷</text>
            <text>文</text>
          </view>
          <view class="scroll-content">
            <image
              v-if="article.coverUrl"
              class="scroll-cover"
              :src="article.coverUrl"
              mode="widthFix"
            />
            <template v-for="(block, index) in bodyBlocks" :key="`${block.type}-${index}`">
              <view v-if="block.type === 'image'" class="scroll-image-block">
                <image
                  class="scroll-body-image"
                  :src="block.url"
                  mode="widthFix"
                />
                <text v-if="block.caption" class="scroll-image-caption">{{ block.caption }}</text>
              </view>
              <text v-else class="scroll-paragraph">{{ block.text }}</text>
            </template>
          </view>
        </view>
      </view>

      <!-- #ifdef MP-WEIXIN -->
      <view class="scroll-share">
        <text class="scroll-share-label">觉得有帮助？</text>
        <button class="wechat-share-button" open-type="share">分享给微信好友</button>
      </view>
      <!-- #endif -->
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getContentArticle } from '@/api/content'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { openWechatOfficialArticle } from '@/features/content/wechatOfficialArticle'
import { buildArticleSharePayload } from '@/features/share/wechatShare'
import type { ContentArticleDetail } from '@/types/api'

const articleId = ref('')
const article = ref<ContentArticleDetail | null>(null)
const loading = ref(false)
const error = ref('')

type ArticleBodyBlock =
  | { type: 'text'; text: string }
  | { type: 'image'; url: string; caption: string }

const categoryTitle = computed(() => {
  if (!article.value) return ''
  return article.value.categoryName || '内容'
})

const bodyBlocks = computed<ArticleBodyBlock[]>(() => {
  if (!article.value) return []
  return article.value.body
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map(parseArticleBodyLine)
})

function parseArticleBodyLine(line: string): ArticleBodyBlock {
  const markdownImage = line.match(/^!\[([^\]]*)\]\(([^)]+)\)$/)
  if (markdownImage) {
    const caption = markdownImage[1].trim()
    const url = markdownImage[2].trim()
    if (isSupportedImageSrc(url)) {
      return { type: 'image', url, caption }
    }
  }

  if (isSupportedImageSrc(line)) {
    return { type: 'image', url: line, caption: '' }
  }

  return { type: 'text', text: line }
}

function isSupportedImageSrc(value: string) {
  const src = value.trim()
  const isImage = /\.(png|jpe?g|webp|gif)(\?.*)?$/i.test(src)
  return isImage && (src.startsWith('https://') || src.startsWith('/static/'))
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

function readMinutes(item: Pick<ContentArticleDetail, 'title' | 'summary' | 'body'>) {
  const chars = `${item.title}${item.summary || ''}${item.body || ''}`.replace(/\s/g, '').length
  return Math.max(1, Math.round(chars / 400))
}

function goBack() {
  uni.switchTab({ url: '/pages/content/index' })
}

onLoad((options) => {
  articleId.value = String(options?.id || '').trim()
  void loadArticle()
})

onShareAppMessage(() => {
  if (!article.value) {
    return buildArticleSharePayload({ id: articleId.value || '0', title: '阅读精选' })
  }
  return buildArticleSharePayload({
    id: article.value.id,
    title: article.value.title,
    coverUrl: article.value.coverUrl
  })
})

async function loadArticle() {
  if (!articleId.value) {
    article.value = null
    error.value = '内容信息缺失'
    return
  }
  loading.value = true
  error.value = ''
  article.value = null
  try {
    article.value = await getContentArticle(articleId.value)
  } catch (err) {
    error.value = apiErrorMessage(err, '内容不存在或暂未发布')
  } finally {
    loading.value = false
  }
}

async function openOfficialArticle() {
  if (!article.value?.externalUrl) {
    uni.showToast({ title: '公众号文章链接缺失', icon: 'none' })
    return
  }
  try {
    await openWechatOfficialArticle(article.value.externalUrl)
  } catch (err) {
    uni.showToast({ title: err instanceof Error ? err.message : '公众号文章暂时无法打开', icon: 'none' })
  }
}
</script>

<style scoped>
.content-detail-page {
  padding-top: 28rpx;
}

.content-detail-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.detail-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.retry-button {
  margin-top: 16rpx;
}

.scroll-head {
  margin-bottom: 18rpx;
}

.scroll-head-copy {
  flex: 1;
  min-width: 0;
}

.scroll-title {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 44rpx;
  font-weight: 700;
  letter-spacing: 1rpx;
  line-height: 1.32;
}

.scroll-lead {
  display: block;
  margin-top: 14rpx;
  color: var(--archive-ink-soft);
  font-size: 25rpx;
  line-height: 1.68;
}

.scroll-seal {
  flex-shrink: 0;
  margin-top: 8rpx;
}

.scroll-meta {
  margin-bottom: 24rpx;
  padding: 20rpx 0;
}

.scroll-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.scroll-badges {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.scroll-badge {
  display: inline-flex;
  border: 1rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  padding: 5rpx 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  letter-spacing: 1rpx;
}

.scroll-badge-wechat {
  border-color: var(--archive-blue);
  color: var(--archive-blue);
}

.scroll-read {
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.scroll-external {
  margin-bottom: 24rpx;
}

.external-note {
  display: block;
  margin-bottom: 18rpx;
  color: var(--archive-ink-soft);
  font-size: 25rpx;
  line-height: 1.68;
}

.scroll-body {
  margin-bottom: 24rpx;
}

.scroll-body-frame {
  display: flex;
  gap: 0;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.scroll-spine {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  width: 58rpx;
  flex-shrink: 0;
  background: var(--archive-blue);
  color: rgba(255, 255, 255, 0.88);
  padding-top: 28rpx;
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 24rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.scroll-content {
  flex: 1;
  min-width: 0;
  padding: 28rpx 0 32rpx 24rpx;
}

.scroll-cover {
  display: block;
  width: 100%;
  margin-bottom: 28rpx;
  border: 1rpx solid var(--archive-line);
}

.scroll-image-block {
  margin-bottom: 30rpx;
  border-top: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0 16rpx;
}

.scroll-body-image {
  display: block;
  width: 100%;
  border: 1rpx solid var(--archive-line);
  background: rgba(255, 255, 255, 0.32);
}

.scroll-image-caption {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.6;
  text-align: center;
}

.scroll-paragraph {
  display: block;
  margin-bottom: 28rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', 'PingFang SC', serif;
  font-size: 29rpx;
  font-weight: 450;
  line-height: 1.88;
  letter-spacing: 0.5rpx;
  text-align: justify;
}

.scroll-paragraph:last-child {
  margin-bottom: 0;
}

.scroll-share {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  margin-top: 8rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 20rpx;
}

.scroll-share-label {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
}

.wechat-share-button {
  margin: 0;
  border: 1rpx solid var(--archive-cinnabar);
  border-radius: 0;
  background: transparent;
  color: var(--archive-cinnabar);
  font-size: 23rpx;
  line-height: 2.1;
}

.wechat-share-button::after {
  border: 0;
}

.content-detail-page :deep(.mini-button) {
  border-radius: 0;
  box-shadow: none;
}
</style>
