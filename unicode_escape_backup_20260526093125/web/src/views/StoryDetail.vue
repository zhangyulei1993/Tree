<template>
  <section class="section">
    <div class="container detail">
      <RouterLink to="/stories" class="back">返回家族故事</RouterLink>

      <article v-if="article" class="card">
        <img v-if="article.cover" :src="article.cover" :alt="article.title" class="cover-img" />

        <div class="content">
          <p class="eyebrow">家族故事</p>
          <h1>{{ article.title }}</h1>
          <p v-if="article.summary" class="summary">{{ article.summary }}</p>
          <div class="body">{{ article.content || emptyText }}</div>
        </div>
      </article>

      <div v-else class="empty">加载中...</div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Article } from '../api'

const route = useRoute()
const article = ref<Article | null>(null)
const emptyText = '\u6682\u65E0\u5185\u5BB9'

onMounted(async () => {
  article.value = await api.articleDetail(String(route.params.id))
})
</script>

<style scoped>
.detail {
  max-width: 920px;
}

.back {
  display: inline-block;
  margin-bottom: 20px;
  color: #9f1d1d;
  font-weight: 700;
}

.cover-img {
  width: 100%;
  max-height: 430px;
  object-fit: cover;
}

.content {
  padding: 34px;
}

.eyebrow {
  margin: 0 0 10px;
  color: #b91c1c;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.content h1 {
  margin: 0 0 18px;
  font-size: 34px;
}

.summary {
  padding: 18px;
  background: #fff4e5;
  color: #7c4a2a;
  border-radius: 14px;
  line-height: 1.8;
}

.body {
  margin-top: 28px;
  line-height: 2;
  white-space: pre-wrap;
}
</style>