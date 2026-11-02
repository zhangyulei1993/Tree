<template>
  <PageShell>
    <section class="content-hero">
      <div class="container hero-inner">
        <span class="eyebrow">Tree Reading</span>
        <h1>阅读家族故事、姓氏典故与使用教程</h1>
        <p>这里汇总平台教程、家族文化内容和精选文章。先了解方法，再开始整理家族记忆。</p>
      </div>
    </section>

    <main class="container content-main">
      <aside class="filters">
        <button :class="{ active: !categoryKey }" @click="selectCategory('')">全部内容</button>
        <button
          v-for="category in categories"
          :key="category.key"
          :class="{ active: categoryKey === category.key }"
          @click="selectCategory(category.key)"
        >
          <strong>{{ category.name }}</strong>
          <small>{{ category.description }}</small>
        </button>
      </aside>

      <section class="article-area">
        <div class="search-row">
          <input v-model="keyword" placeholder="搜索标题、摘要或正文" @keydown.enter="reload" />
          <button class="button" :disabled="loading" @click="reload">搜索</button>
        </div>

        <div v-if="error" class="state-card error">{{ error }}</div>
        <div v-else-if="loading" class="state-card">正在加载内容...</div>
        <div v-else-if="articles.length === 0" class="state-card">暂无内容，换个分类或关键词试试。</div>

        <div v-else class="article-list">
          <RouterLink v-for="article in articles" :key="article.id" class="article-card" :to="`/content/${article.slug || article.id}`">
            <div class="article-mark">{{ shortCategory(article.categoryName) }}</div>
            <div>
              <div class="article-meta">
                <span>{{ article.categoryName }}</span>
                <span v-if="article.isFeatured">精选</span>
                <span>{{ readMinutes(article) }} 分钟阅读</span>
              </div>
              <h2>{{ article.title }}</h2>
              <p>{{ article.summary || '暂无摘要' }}</p>
            </div>
          </RouterLink>
        </div>
      </section>
    </main>
  </PageShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listContentArticles, listContentCategories } from '@/api/content'
import PageShell from '@/components/PageShell.vue'
import type { ContentArticleSummary, ContentCategory } from '@/types/api'

const categories = ref<ContentCategory[]>([])
const articles = ref<ContentArticleSummary[]>([])
const categoryKey = ref('')
const keyword = ref('')
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  await Promise.all([loadCategories(), loadArticles()])
})

async function loadCategories() {
  try {
    categories.value = await listContentCategories()
  } catch (err) {
    error.value = apiErrorMessage(err, '内容分类加载失败')
  }
}

async function loadArticles() {
  loading.value = true
  error.value = ''
  try {
    const result = await listContentArticles({
      categoryKey: categoryKey.value || undefined,
      keyword: keyword.value || undefined,
      page: 1,
      pageSize: 50
    })
    articles.value = result.items
  } catch (err) {
    error.value = apiErrorMessage(err, '内容列表加载失败')
  } finally {
    loading.value = false
  }
}

function selectCategory(key: string) {
  categoryKey.value = key
  void loadArticles()
}

function reload() {
  void loadArticles()
}

function shortCategory(name: string) {
  return name.slice(0, 1)
}

function readMinutes(article: ContentArticleSummary) {
  const chars = `${article.title}${article.summary || ''}`.length
  return Math.max(1, Math.round(chars / 120))
}
</script>

<style scoped>
.content-hero {
  padding: 72px 0 30px;
}

.hero-inner {
  border: 1px solid rgba(31, 58, 95, 0.08);
  border-radius: 30px;
  background:
    radial-gradient(circle at 88% 20%, rgba(47, 107, 87, 0.14), transparent 260px),
    linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(244, 250, 247, 0.94));
  padding: 52px;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

h1 {
  max-width: 820px;
  margin: 16px 0;
  color: var(--color-primary);
  font-size: clamp(34px, 4vw, 56px);
  line-height: 1.08;
}

.hero-inner p {
  max-width: 720px;
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 18px;
  line-height: 1.8;
}

.content-main {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 24px;
  padding: 22px 0 64px;
}

.filters {
  position: sticky;
  top: 88px;
  display: grid;
  gap: 10px;
  align-self: start;
}

.filters button,
.article-card,
.state-card {
  border: 1px solid var(--color-border);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.86);
  box-shadow: var(--shadow-light);
}

.filters button {
  display: grid;
  gap: 4px;
  padding: 16px;
  color: var(--color-text-secondary);
  text-align: left;
}

.filters button.active {
  border-color: rgba(47, 107, 87, 0.26);
  background: var(--color-heritage-green-light);
  color: var(--color-primary);
}

.filters small {
  line-height: 1.5;
}

.search-row {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.search-row input {
  flex: 1;
  min-width: 0;
}

.article-list {
  display: grid;
  gap: 14px;
}

.article-card {
  display: grid;
  grid-template-columns: 66px minmax(0, 1fr);
  gap: 18px;
  padding: 22px;
  color: inherit;
  text-decoration: none;
}

.article-card:hover {
  transform: translateY(-2px);
}

.article-mark {
  display: grid;
  width: 56px;
  height: 56px;
  place-items: center;
  border-radius: 18px;
  background: var(--color-warm-gold-light);
  color: var(--color-warm-gold-text);
  font-size: 24px;
  font-weight: 900;
}

.article-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  color: var(--color-text-muted);
  font-size: 13px;
}

.article-card h2 {
  margin: 8px 0;
  color: var(--color-primary);
  font-size: 22px;
}

.article-card p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.75;
}

.state-card {
  padding: 28px;
  color: var(--color-text-secondary);
}

.state-card.error {
  color: var(--color-danger);
}

@media (max-width: 760px) {
  .hero-inner {
    padding: 30px 22px;
  }

  .content-main {
    grid-template-columns: 1fr;
  }

  .filters {
    position: static;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .article-card {
    grid-template-columns: 1fr;
  }
}
</style>
