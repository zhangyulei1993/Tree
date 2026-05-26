<template>
  <section class="section">
    <div class="container about">
      <div>
        <p class="eyebrow">关于我们</p>
        <h1 class="section-title">让家庭关系更清晰，让亲情记忆被好好保留</h1>
        <p class="section-subtitle">
          {{ site.description || defaultDescription }}
        </p>

        <div class="features">
          <div class="card feature">
            <h3>亲属关系梳理</h3>
            <p>以结构化方式整理父母、子女、配偶、长辈、晚辈等关系。</p>
          </div>

          <div class="card feature">
            <h3>家族故事记录</h3>
            <p>将长辈口述、老照片、家庭事件和温暖记忆汇聚成家族故事。</p>
          </div>

          <div class="card feature">
            <h3>传统与温情</h3>
            <p>尊重传统家族文化，用现代化网站方式承载亲情与血脉。</p>
          </div>
        </div>
      </div>

      <div class="card contact-card">
        <img v-if="site.logo" :src="site.logo" alt="logo" />
        <h3>{{ site.site_name || defaultTitle }}</h3>
        <p><strong>电话：</strong>{{ site.phone || '-' }}</p>
        <p><strong>邮箱：</strong>{{ site.email || '-' }}</p>
        <p><strong>地址：</strong>{{ site.address || '-' }}</p>
        <RouterLink to="/contact" class="btn">联系我们</RouterLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { api, type SiteConfig } from '../api'

const defaultTitle = '\u5BB6\u8109\u4EB2\u7F18'
const defaultDescription = '\u6211\u4EEC\u4E13\u6CE8\u4E8E\u5BB6\u5EAD\u4EB2\u5C5E\u5173\u7CFB\u7ED3\u6784\u68B3\u7406\u3001\u5BB6\u65CF\u6545\u4E8B\u8BB0\u5F55\u548C\u4EB2\u60C5\u8BB0\u5FC6\u4F20\u627F\u3002'

const site = reactive<SiteConfig>({
  site_name: '',
  logo: '',
  phone: '',
  email: '',
  address: '',
  description: ''
})

onMounted(async () => {
  try {
    const res = await api.siteConfig()
    Object.assign(site, res)
  } catch {
    // ignore
  }
})
</script>

<style scoped>
.about {
  display: grid;
  grid-template-columns: 1.25fr 0.75fr;
  gap: 34px;
  align-items: start;
}

.eyebrow {
  margin: 0 0 10px;
  color: #b91c1c;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.features {
  display: grid;
  gap: 16px;
}

.feature {
  padding: 24px;
}

.feature h3 {
  margin: 0 0 10px;
}

.feature p,
.contact-card p {
  color: #7c4a2a;
  line-height: 1.8;
}

.contact-card {
  padding: 26px;
}

.contact-card img {
  width: 150px;
  max-height: 80px;
  object-fit: contain;
  margin-bottom: 18px;
}

.contact-card .btn {
  margin-top: 12px;
}

@media (max-width: 768px) {
  .about {
    grid-template-columns: 1fr;
  }
}
</style>