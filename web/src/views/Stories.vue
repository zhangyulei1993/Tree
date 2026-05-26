<template>
  <section class="section">
    <div class="container">
      <p class="eyebrow">&#x5BB6;&#x65CF;&#x8BB0;&#x5FC6;</p>
      <h1 class="section-title">&#x5BB6;&#x65CF;&#x6545;&#x4E8B;</h1>
      <p class="section-subtitle">
        &#x9605;&#x8BFB;&#x5BB6;&#x65CF;&#x4E2D;&#x7684;&#x4EBA;&#x3001;&#x4E8B;&#x3001;&#x8BB0;&#x5FC6;&#x548C;&#x4F20;&#x627F;&#xFF0C;&#x8BA9;&#x4E00;&#x5BB6;&#x4EBA;&#x7684;&#x6765;&#x5904;&#x66F4;&#x52A0;&#x6E05;&#x6670;&#x3002;
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

      <div v-else class="empty">&#x6682;&#x65E0;&#x5BB6;&#x65CF;&#x6545;&#x4E8B;</div>

      <div class="pager" v-if="total > pageSize">
        <button class="btn secondary" :disabled="page <= 1" @click="prev">&#x4E0A;&#x4E00;&#x9875;</button>
        <span>&#x7B2C; {{ page }} &#x9875;&#xFF0C;&#x5171; {{ total }} &#x6761;</span>
        <button class="btn secondary" :disabled="page * pageSize >= total" @click="next">&#x4E0B;&#x4E00;&#x9875;</button>
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