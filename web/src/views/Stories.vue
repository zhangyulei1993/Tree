<template>
  <section class="section">
    <div class="container">
      <p class="eyebrow">家族记忆</p>
      <h1 class="section-title">家族故事</h1>
      <p class="section-subtitle">
        阅读家族中的人、事、记忆和传承，让一家人的来处更加清晰。
      </p>

      <div v-if="articles.length" class="grid-3">
        <RouterLink v-for="item in articles" :key="item.id" :to="'/story/' + item.id" class="article-card card">
          <img v-if="item.cover" :src="item.cover" :alt="item.title" />
          <div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.summary }}</p>
          </div>
        </RouterLink>
      </div>

      <div v-else class="empty">暂无家族故事</div>

      <div class="pager" v-if="total > pageSize">
        <button class="btn secondary" :disabled="page <= 1" @click="prev">上一页</button>
        <span>第 {{ page }} 页，共 {{ total }} 条</span>
        <button class="btn secondary" :disabled="page * pageSize >= total" @click="next">下一页</button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, type Article } from '../api'

const articles = ref<Article[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(9)

onMounted(load)

async function load() {
  const res = await api.articleList({
    page: page.value,
    page_size: pageSize.value
  })

  articles.value = res.list
  total.value = res.total
}

function prev() {
  if (page.value <= 1) return
  page.value--
  load()
}

function next() {
  if (page.value * pageSize.value >= total.value) return
  page.value++
  load()
}
</script>

<style scoped>
.eyebrow {
  margin: 0 0 10px;
  color: #b91c1c;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.article-card img {
  width: 100%;
  height: 220px;
  object-fit: cover;
}

.article-card div {
  padding: 20px;
}

.article-card p {
  color: #7c4a2a;
  line-height: 1.7;
}

.pager {
  margin-top: 32px;
  display: flex;
  justify-content: center;
  gap: 14px;
  align-items: center;
  flex-wrap: wrap;
}
</style>