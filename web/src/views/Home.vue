<template>
  <div>
    <section class="hero" :style="heroStyle">
      <div class="hero-mask">
        <div class="container hero-content">
          <p class="eyebrow">&#x5BB6;&#x5EAD;&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x7ED3;&#x6784;</p>
          <h1>{{ site.site_name || defaultTitle }}</h1>
          <p>{{ site.description || defaultDescription }}</p>

          <div class="hero-actions">
            <RouterLink to="/family-tree" class="btn">&#x67E5;&#x770B;&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x6811;</RouterLink>
            <RouterLink to="/stories" class="btn secondary">&#x9605;&#x8BFB;&#x5BB6;&#x65CF;&#x6545;&#x4E8B;</RouterLink>
          </div>
        </div>
      </div>
    </section>

    <section class="section">
      <div class="container">
        <h2 class="section-title">&#x6E29;&#x6696;&#x7684;&#x5BB6;&#x65CF;&#x8BB0;&#x5FC6;</h2>
        <p class="section-subtitle">&#x8BB0;&#x5F55;&#x4E00;&#x5BB6;&#x4EBA;&#x7684;&#x5173;&#x7CFB;&#x3001;&#x6545;&#x4E8B;&#x548C;&#x4F20;&#x627F;&#xFF0C;&#x8BA9;&#x4EB2;&#x60C5;&#x6709;&#x8FF9;&#x53EF;&#x5FAA;&#x3002;</p>

        <div v-if="banners.length" class="grid-3">
          <a v-for="item in banners" :key="item.id" :href="item.link_url || '#'" class="banner-card card">
            <img :src="item.image_url" :alt="item.title" />
            <div>
              <h3>{{ item.title }}</h3>
            </div>
          </a>
        </div>

        <div v-else class="empty">&#x6682;&#x65E0;&#x8F6E;&#x64AD;&#x56FE;</div>
      </div>
    </section>

    <section class="section light">
      <div class="container">
        <h2 class="section-title">&#x5BB6;&#x65CF;&#x6545;&#x4E8B;</h2>
        <p class="section-subtitle">&#x4ECE;&#x957F;&#x8F88;&#x8BB0;&#x5FC6;&#x5230;&#x665A;&#x8F88;&#x6210;&#x957F;&#xFF0C;&#x628A;&#x4E00;&#x5BB6;&#x4EBA;&#x7684;&#x6545;&#x4E8B;&#x597D;&#x597D;&#x7559;&#x4E0B;&#x6765;&#x3002;</p>

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
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type Article, type Banner, type SiteConfig } from '../api'

const defaultTitle = '\u5BB6\u8109\u4EB2\u7F18'
const defaultDescription = '\u4E13\u6CE8\u4E8E\u5BB6\u5EAD\u4EB2\u5C5E\u5173\u7CFB\u7ED3\u6784\u68B3\u7406\u3001\u5BB6\u65CF\u6545\u4E8B\u8BB0\u5F55\u4E0E\u4EB2\u60C5\u4F20\u627F\u3002'

const site = reactive<SiteConfig>({
  site_name: '',
  logo: '',
  phone: '',
  email: '',
  address: '',
  description: ''
})

const banners = ref<Banner[]>([])
const articles = ref<Article[]>([])

const heroStyle = computed(() => {
  const image = banners.value[0]?.image_url
  return image ? { backgroundImage: 'url(' + image + ')' } : {}
})

onMounted(async () => {
  try {
    const siteRes = await api.siteConfig()
    Object.assign(site, siteRes)

    banners.value = await api.bannerList()

    const articleRes = await api.articleList({ page: 1, page_size: 6 })
    articles.value = articleRes.list
  } catch {
    // ignore
  }
})
</script>

<style scoped>
.hero {
  min-height: 620px;
  background: linear-gradient(135deg, #3b1f13, #9f1d1d);
  background-size: cover;
  background-position: center;
}

.hero-mask {
  min-height: 620px;
  background: linear-gradient(90deg, rgba(59,31,19,0.88), rgba(59,31,19,0.35));
  display: flex;
  align-items: center;
}

.hero-content {
  color: #fff7ed;
}

.eyebrow {
  color: #fde68a;
  font-weight: 900;
}

.hero h1 {
  margin: 16px 0;
  font-size: clamp(42px, 7vw, 76px);
  line-height: 1.05;
  max-width: 760px;
}

.hero p {
  max-width: 680px;
  line-height: 1.9;
  font-size: 18px;
}

.hero-actions {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  margin-top: 32px;
}

.light {
  background: #fff4e5;
}

.banner-card img,
.article-card img {
  width: 100%;
  height: 220px;
  object-fit: cover;
}

.banner-card div,
.article-card div {
  padding: 20px;
}

.article-card p {
  color: #7c4a2a;
  line-height: 1.7;
}
</style>