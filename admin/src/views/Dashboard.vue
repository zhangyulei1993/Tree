<template>
  <div class="card">
    <h2 class="page-title">首页看板</h2>

    <div class="stats">
      <div class="stat">
        <span>轮播图</span>
        <strong>{{ data.banner_count }}</strong>
      </div>

      <div class="stat">
        <span>家族故事</span>
        <strong>{{ data.article_count }}</strong>
      </div>

      <div class="stat">
        <span>留言</span>
        <strong>{{ data.contact_count }}</strong>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { api } from '../api'

const data = reactive({
  banner_count: 0,
  article_count: 0,
  contact_count: 0
})

onMounted(async () => {
  const res = await api.dashboard()
  Object.assign(data, res)
})
</script>

<style scoped>
.stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18px;
}

.stat {
  border-radius: 18px;
  padding: 26px;
  background: linear-gradient(135deg, #b91c1c, #7f1d1d);
  color: #fff7ed;
}

.stat span {
  display: block;
  color: #fde7c7;
}

.stat strong {
  display: block;
  margin-top: 18px;
  font-size: 42px;
}

@media (max-width: 800px) {
  .stats {
    grid-template-columns: 1fr;
  }
}
</style>