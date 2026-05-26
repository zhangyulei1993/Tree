<template>
  <div>
    <header class="site-header">
      <div class="container nav">
        <RouterLink to="/" class="brand">
          <img v-if="site.logo" :src="site.logo" alt="logo" />
          <span>{{ site.site_name || defaultTitle }}</span>
        </RouterLink>

        <button class="menu-btn" @click="menuOpen = !menuOpen">&#x83DC;&#x5355;</button>

        <nav :class="{ open: menuOpen }">
          <RouterLink to="/" @click="closeMenu">&#x9996;&#x9875;</RouterLink>
          <RouterLink to="/about" @click="closeMenu">&#x5173;&#x4E8E;&#x6211;&#x4EEC;</RouterLink>
          <RouterLink to="/stories" @click="closeMenu">&#x5BB6;&#x65CF;&#x6545;&#x4E8B;</RouterLink>
          <RouterLink to="/family-tree" @click="closeMenu">&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x6811;</RouterLink>
          <RouterLink to="/contact" class="nav-cta" @click="closeMenu">&#x8054;&#x7CFB;&#x6211;&#x4EEC;</RouterLink>
        </nav>
      </div>
    </header>

    <main>
      <RouterView />
    </main>

    <footer class="site-footer">
      <div class="container footer-grid">
        <div>
          <h3>{{ site.site_name || defaultTitle }}</h3>
          <p>{{ site.description || defaultDescription }}</p>
        </div>

        <div>
          <h4>&#x8054;&#x7CFB;&#x65B9;&#x5F0F;</h4>
          <p>&#x7535;&#x8BDD;&#xFF1A;{{ site.phone || '-' }}</p>
          <p>&#x90AE;&#x7BB1;&#xFF1A;{{ site.email || '-' }}</p>
          <p>&#x5730;&#x5740;&#xFF1A;{{ site.address || '-' }}</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type SiteConfig } from '../api'

const menuOpen = ref(false)
const defaultTitle = '\u5BB6\u8109\u4EB2\u7F18'
const defaultDescription = '\u8BB0\u5F55\u5BB6\u5EAD\u4EB2\u5C5E\u5173\u7CFB\u3001\u5BB6\u65CF\u6545\u4E8B\u4E0E\u6E29\u6696\u8BB0\u5FC6\u3002'

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

function closeMenu() {
  menuOpen.value = false
}
</script>

<style scoped>
.site-header {
  position: sticky;
  top: 0;
  z-index: 20;
  background: rgba(255, 250, 243, 0.9);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid #ead8c0;
}

.nav {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 900;
}

.brand img {
  width: 44px;
  height: 44px;
  object-fit: contain;
  border-radius: 12px;
}

nav {
  display: flex;
  align-items: center;
  gap: 26px;
  color: #5b3924;
  font-weight: 700;
}

.nav-cta {
  background: #9f1d1d;
  color: #fff;
  padding: 10px 18px;
  border-radius: 999px;
}

.menu-btn {
  display: none;
  border: none;
  border-radius: 10px;
  padding: 10px 12px;
  background: #f3dfc7;
  color: #6b2f16;
}

.site-footer {
  background: #3b1f13;
  color: #f8ead8;
  padding: 54px 0;
}

.footer-grid {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 36px;
}

.site-footer p {
  line-height: 1.8;
  color: #e7c9a8;
}

@media (max-width: 768px) {
  .menu-btn {
    display: block;
  }

  nav {
    display: none;
    position: absolute;
    top: 72px;
    left: 16px;
    right: 16px;
    background: #fffaf3;
    padding: 18px;
    border-radius: 18px;
    box-shadow: 0 18px 50px rgba(80, 45, 20, 0.14);
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  nav.open {
    display: flex;
  }

  .footer-grid {
    grid-template-columns: 1fr;
  }
}
</style>