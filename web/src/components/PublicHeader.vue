<template>
  <header class="header">
    <div class="header-inner">
      <RouterLink class="brand" to="/">
        <span class="brand-mark">T</span>
        <span class="brand-copy">
          <strong>Tree</strong>
          <small>家脉亲缘</small>
        </span>
      </RouterLink>
      <nav>
        <RouterLink to="/">首页</RouterLink>
        <RouterLink to="/families">找家族</RouterLink>
        <RouterLink to="/content">阅读</RouterLink>
        <RouterLink to="/me/families">我的家庭</RouterLink>
        <RouterLink to="/me">我的</RouterLink>
      </nav>
      <div v-if="session.isLoggedIn" class="account-actions">
        <RouterLink class="account-link" to="/me">
          {{ session.user?.nickname || '个人中心' }}
        </RouterLink>
        <button class="button secondary compact" :disabled="loggingOut" @click="logout">
          {{ loggingOut ? '退出中...' : '退出登录' }}
        </button>
      </div>
      <RouterLink v-else class="button secondary compact" to="/login">登录 / 注册</RouterLink>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const router = useRouter()
const loggingOut = ref(false)

async function logout() {
  loggingOut.value = true
  try {
    await session.logout()
  } catch {
    // The session store clears local state even if the logout request fails.
  } finally {
    loggingOut.value = false
    await router.push('/login')
  }
}
</script>

<style scoped>
.header {
  position: sticky;
  top: 0;
  z-index: 10;
  border-bottom: 1px solid rgba(232, 237, 240, 0.82);
  background: rgba(251, 252, 251, 0.78);
  backdrop-filter: blur(22px);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-height: 72px;
  width: min(1180px, calc(100vw - 40px));
  margin: 0 auto;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
}

.brand-mark {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border-radius: 15px;
  background: var(--gradient-primary);
  color: #fff;
  box-shadow: 0 12px 28px rgba(31, 58, 95, 0.18);
  font-weight: 900;
}

.brand-copy {
  display: grid;
  gap: 1px;
}

.brand-copy strong {
  line-height: 1;
  letter-spacing: -0.02em;
}

.brand-copy small {
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 600;
}

nav {
  display: flex;
  gap: 6px;
  color: var(--color-text-secondary);
  border: 1px solid rgba(31, 58, 95, 0.08);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.74);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 8px 22px rgba(31, 58, 95, 0.04);
  padding: 5px;
}

nav a {
  border-radius: 999px;
  padding: 8px 13px;
  font-size: 14px;
  font-weight: 700;
  transition:
    background var(--transition-fast),
    color var(--transition-fast),
    transform var(--transition-fast);
}

nav a:hover {
  transform: translateY(-1px);
}

nav a.router-link-active {
  background: var(--color-heritage-green-light);
  color: var(--color-heritage-green);
  font-weight: 700;
}

.account-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.account-link {
  color: var(--color-primary);
  font-weight: 700;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.compact {
  min-height: 38px;
  padding-inline: 15px;
}

@media (max-width: 760px) {
  .header-inner {
    width: min(100vw - 24px, 1180px);
    align-items: stretch;
    flex-direction: column;
    padding-block: 12px;
  }

  nav {
    width: 100%;
    flex-wrap: wrap;
    border-radius: 18px;
  }

  .account-actions {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
