<template>
  <div class="layout">
    <aside class="aside">
      <div class="brand">
        <img class="brand-logo" src="/brand/favicon.svg" alt="logo" />
        <div>
          <strong>&#x5BB6;&#x8109;&#x4EB2;&#x7F18;</strong>
          <span>&#x540E;&#x53F0;&#x7BA1;&#x7406;</span>
        </div>
      </div>

      <nav>
        <RouterLink to="/dashboard">&#x9996;&#x9875;&#x770B;&#x677F;</RouterLink>
        <RouterLink to="/site">&#x5B98;&#x7F51;&#x914D;&#x7F6E;</RouterLink>
        <RouterLink to="/banner">&#x8F6E;&#x64AD;&#x56FE;&#x7BA1;&#x7406;</RouterLink>
        <RouterLink to="/article">&#x5BB6;&#x65CF;&#x6545;&#x4E8B;</RouterLink>
        <RouterLink to="/contact">&#x7559;&#x8A00;&#x7BA1;&#x7406;</RouterLink>
        <RouterLink to="/family">&#x5BB6;&#x65CF;&#x7BA1;&#x7406;</RouterLink>
        <RouterLink to="/member">&#x6210;&#x5458;&#x7BA1;&#x7406;</RouterLink>
        <RouterLink to="/relationship">&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;</RouterLink>
      </nav>
    </aside>

    <section class="main-area">
      <header class="header">
        <strong>{{ pageTitle }}</strong>
        <div>
          <span class="user">{{ username }}</span>
          <button class="btn secondary small" @click="logout">&#x9000;&#x51FA;</button>
        </div>
      </header>

      <main class="main">
        <RouterView />
      </main>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const pageTitle = computed(() => {
  const map: Record<string, string> = {
    '/dashboard': '\u9996\u9875\u770B\u677F',
    '/site': '\u5B98\u7F51\u914D\u7F6E',
    '/banner': '\u8F6E\u64AD\u56FE\u7BA1\u7406',
    '/article': '\u5BB6\u65CF\u6545\u4E8B',
    '/contact': '\u7559\u8A00\u7BA1\u7406',
    '/family': '\u5BB6\u65CF\u7BA1\u7406',
    '/member': '\u6210\u5458\u7BA1\u7406',
    '/relationship': '\u4EB2\u5C5E\u5173\u7CFB'
  }

  return map[route.path] || '\u540E\u53F0\u7BA1\u7406'
})

const username = computed(() => {
  const raw = localStorage.getItem('admin_user')

  if (!raw) {
    return 'admin'
  }

  try {
    const user = JSON.parse(raw)
    return user.nickname || user.username || 'admin'
  } catch {
    return 'admin'
  }
})

function logout() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_user')
  router.push('/login')
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 240px 1fr;
}

.aside {
  background: #3b1f13;
  color: #fef3e2;
  padding: 18px;
  overflow-y: auto;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 6px 24px;
}

.brand-logo {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  object-fit: contain;
}

.brand span {
  display: block;
  margin-top: 4px;
  color: #e7c9a8;
  font-size: 13px;
}

nav {
  display: grid;
  gap: 8px;
}

nav a {
  padding: 12px 14px;
  border-radius: 12px;
  color: #f8ead8;
}

nav a.router-link-active,
nav a:hover {
  background: rgba(255, 255, 255, 0.12);
}

.main-area {
  min-width: 0;
}

.header {
  height: 64px;
  padding: 0 24px;
  background: #fffaf3;
  border-bottom: 1px solid #ead8c0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header > div {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user {
  color: #7c4a2a;
}

.main {
  padding: 24px;
}

@media (max-width: 800px) {
  .layout {
    grid-template-columns: 1fr;
  }

  nav {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>