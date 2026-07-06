<template>
  <view class="tree-page detail-page">
    <MiniBackHome />
    <MiniCard v-if="loading">
      <MiniEmptyState title="正在加载" description="正在读取内容，请稍候。" />
    </MiniCard>

    <MiniCard v-else-if="!article">
      <MiniEmptyState title="内容不存在" :description="error || '请返回阅读页重新选择。'" />
      <MiniButton variant="secondary" @click="goBack">返回阅读</MiniButton>
    </MiniCard>

    <template v-else>
      <view class="article-hero">
        <text class="article-category">{{ categoryTitle }}</text>
        <text class="article-title">{{ article.title }}</text>
        <text class="article-summary">{{ article.summary }}</text>
      </view>

      <view v-if="article.contentType === 'WECHAT_OFFICIAL'" class="article-body external-article-card">
        <text class="external-article-note">本文发布于微信公众号，将通过微信官方页面打开。</text>
        <MiniButton @click="openOfficialArticle">阅读公众号文章</MiniButton>
      </view>

      <view v-else class="article-body">
        <text
          v-for="(paragraph, index) in bodyParagraphs"
          :key="index"
          class="article-paragraph"
        >
          {{ paragraph }}
        </text>
      </view>

      <!-- #ifdef MP-WEIXIN -->
      <view v-if="article" class="article-share-card">
        <text class="article-share-label">觉得有帮助？</text>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { openWechatOfficialArticle } from '@/features/content/wechatOfficialArticle'
import { buildArticleSharePayload } from '@/features/share/wechatShare'
import type { ContentArticleDetail } from '@/types/api'

const articleId = ref('')
const article = ref<ContentArticleDetail | null>(null)
const loading = ref(false)
const error = ref('')

const categoryTitle = computed(() => {
  if (!article.value) return ''
  return article.value.categoryName || '内容'
})

const bodyParagraphs = computed(() => {
  if (!article.value) return []
  return article.value.body
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
})

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
.detail-page {
  background: linear-gradient(180deg, #f7faf9 0%, #f5f8fc 100%);
}

.article-hero {
  margin-bottom: 20rpx;
  border-radius: var(--tree-radius-lg);
  background: #fff;
  padding: 28rpx 24rpx;
  box-shadow: 0 2rpx 14rpx rgba(15, 23, 42, 0.04);
}

.article-category {
  display: inline-flex;
  border-radius: 8rpx;
  background: #f1f5f9;
  color: var(--tree-text-secondary);
  padding: 6rpx 12rpx;
  font-size: 22rpx;
  font-weight: 500;
}

.article-title {
  display: block;
  margin-top: 16rpx;
  color: var(--tree-text);
  font-size: 36rpx;
  font-weight: 600;
  line-height: 1.4;
}

.article-summary {
  display: block;
  margin-top: 12rpx;
  color: var(--tree-text-secondary);
  font-size: 26rpx;
  line-height: 1.6;
}

.article-body {
  border-radius: var(--tree-radius-lg);
  background: #fff;
  padding: 28rpx 24rpx;
  box-shadow: 0 2rpx 14rpx rgba(15, 23, 42, 0.04);
}

.external-article-card {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.external-article-note {
  color: var(--tree-text-secondary);
  font-size: 26rpx;
  line-height: 1.65;
}

.article-paragraph {
  display: block;
  margin-bottom: 20rpx;
  color: var(--tree-text);
  font-size: 28rpx;
  line-height: 1.75;
}

.article-paragraph:last-child {
  margin-bottom: 0;
}

.article-share-card {
  margin-top: 20rpx;
  border-radius: var(--tree-radius-lg);
  background: #fff;
  padding: 24rpx;
  box-shadow: 0 2rpx 14rpx rgba(15, 23, 42, 0.04);
}

.article-share-label {
  display: block;
  margin-bottom: 16rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}
</style>
