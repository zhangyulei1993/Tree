<template>
  <PageShell>
    <main class="container detail-wrap">
      <div v-if="loading" class="state-card">正在加载内容...</div>
      <div v-else-if="error" class="state-card error">
        <strong>内容加载失败</strong>
        <p>{{ error }}</p>
        <RouterLink class="button secondary" to="/content">返回阅读</RouterLink>
      </div>
      <article v-else-if="article" class="article-detail">
        <RouterLink class="back-link" to="/content">返回阅读</RouterLink>
        <header>
          <span>{{ article.categoryName }}</span>
          <h1>{{ article.title }}</h1>
          <p>{{ article.summary }}</p>
          <div class="meta">
            <span>{{ article.authorName || 'Tree 编辑部' }}</span>
            <span v-if="article.source">{{ article.source }}</span>
            <span v-if="article.publishedAt">{{ formatDate(article.publishedAt) }}</span>
          </div>
        </header>
        <section class="body">
          <p v-for="(paragraph, index) in paragraphs" :key="index">{{ paragraph }}</p>
        </section>
      </article>
    </main>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { getContentArticle } from '@/api/content'
import PageShell from '@/components/PageShell.vue'
import type { ContentArticleDetail } from '@/types/api'

const route = useRoute()
const article = ref<ContentArticleDetail | null>(null)
const loading = ref(false)
const error = ref('')

const paragraphs = computed(() => article.value?.body.split('\n').map((item) => item.trim()).filter(Boolean) || [])

onMounted(loadArticle)
watch(() => route.params.articleId, loadArticle)

async function loadArticle() {
  const id = String(route.params.articleId || '').trim()
  if (!id) return
  loading.value = true
  error.value = ''
  article.value = null
  try {
    article.value = await getContentArticle(id)
  } catch (err) {
    error.value = apiErrorMessage(err, '内容不存在或暂未发布')
  } finally {
    loading.value = false
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.detail-wrap {
  padding: 70px 0;
}

.article-detail {
  max-width: 860px;
  margin: 0 auto;
}

.back-link {
  color: var(--color-heritage-green);
  font-weight: 800;
  text-decoration: none;
}

header {
  margin-top: 18px;
  border: 1px solid rgba(31, 58, 95, 0.08);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.9);
  padding: 44px;
  box-shadow: var(--shadow-light);
}

header span {
  color: var(--color-heritage-green);
  font-weight: 800;
}

h1 {
  margin: 14px 0 14px;
  color: var(--color-primary);
  font-size: clamp(32px, 4vw, 54px);
  line-height: 1.12;
}

header p {
  color: var(--color-text-secondary);
  font-size: 18px;
  line-height: 1.75;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 18px;
}

.meta span {
  border-radius: 999px;
  background: var(--color-bg-soft);
  color: var(--color-text-muted);
  padding: 6px 12px;
  font-size: 13px;
}

.body {
  margin-top: 22px;
  border-radius: 26px;
  background: rgba(255, 255, 255, 0.84);
  padding: 34px 42px;
  color: var(--color-text-primary);
  font-size: 18px;
  line-height: 1.9;
  box-shadow: var(--shadow-light);
}

.body p {
  margin: 0 0 20px;
}

.state-card {
  max-width: 760px;
  margin: 0 auto;
  border: 1px solid var(--color-border);
  border-radius: 24px;
  background: #fff;
  padding: 32px;
  color: var(--color-text-secondary);
}

.state-card.error {
  color: var(--color-danger);
}

@media (max-width: 720px) {
  .detail-wrap {
    padding: 36px 0;
  }

  header,
  .body {
    padding: 24px;
  }
}
</style>
