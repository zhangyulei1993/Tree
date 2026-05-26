<template>
  <section class="section">
    <div class="container about">
      <div>
        <p class="eyebrow">&#x5173;&#x4E8E;&#x6211;&#x4EEC;</p>
        <h1 class="section-title">&#x8BA9;&#x5BB6;&#x5EAD;&#x5173;&#x7CFB;&#x66F4;&#x6E05;&#x6670;&#xFF0C;&#x8BA9;&#x4EB2;&#x60C5;&#x8BB0;&#x5FC6;&#x88AB;&#x597D;&#x597D;&#x4FDD;&#x7559;</h1>
        <p class="section-subtitle">
          {{ site.description || defaultDescription }}
        </p>

        <div class="features">
          <div class="card feature">
            <h3>&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x68B3;&#x7406;</h3>
            <p>&#x4EE5;&#x7ED3;&#x6784;&#x5316;&#x65B9;&#x5F0F;&#x6574;&#x7406;&#x7236;&#x6BCD;&#x3001;&#x5B50;&#x5973;&#x3001;&#x914D;&#x5076;&#x3001;&#x957F;&#x8F88;&#x3001;&#x665A;&#x8F88;&#x7B49;&#x5173;&#x7CFB;&#x3002;</p>
          </div>

          <div class="card feature">
            <h3>&#x5BB6;&#x65CF;&#x6545;&#x4E8B;&#x8BB0;&#x5F55;</h3>
            <p>&#x5C06;&#x957F;&#x8F88;&#x53E3;&#x8FF0;&#x3001;&#x8001;&#x7167;&#x7247;&#x3001;&#x5BB6;&#x5EAD;&#x4E8B;&#x4EF6;&#x548C;&#x6E29;&#x6696;&#x8BB0;&#x5FC6;&#x6C47;&#x805A;&#x6210;&#x5BB6;&#x65CF;&#x6545;&#x4E8B;&#x3002;</p>
          </div>

          <div class="card feature">
            <h3>&#x4F20;&#x7EDF;&#x4E0E;&#x6E29;&#x60C5;</h3>
            <p>&#x5C0A;&#x91CD;&#x4F20;&#x7EDF;&#x5BB6;&#x65CF;&#x6587;&#x5316;&#xFF0C;&#x7528;&#x73B0;&#x4EE3;&#x5316;&#x7F51;&#x7AD9;&#x65B9;&#x5F0F;&#x627F;&#x8F7D;&#x4EB2;&#x60C5;&#x4E0E;&#x8840;&#x8109;&#x3002;</p>
          </div>
        </div>
      </div>

      <div class="card contact-card">
        <img v-if="site.logo" :src="site.logo" alt="logo" />
        <h3>{{ site.site_name || defaultTitle }}</h3>
        <p><strong>&#x7535;&#x8BDD;&#xFF1A;</strong>{{ site.phone || '-' }}</p>
        <p><strong>&#x90AE;&#x7BB1;&#xFF1A;</strong>{{ site.email || '-' }}</p>
        <p><strong>&#x5730;&#x5740;&#xFF1A;</strong>{{ site.address || '-' }}</p>
        <RouterLink to="/contact" class="btn">&#x8054;&#x7CFB;&#x6211;&#x4EEC;</RouterLink>
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