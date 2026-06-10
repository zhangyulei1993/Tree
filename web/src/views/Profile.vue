<template>
  <PageShell>
    <section class="container profile-page">
      <div class="page-heading">
        <div>
          <span class="eyebrow">{{ session.mode === 'real' ? '真实 API 会话' : 'Mock 会话' }}</span>
          <h1>个人中心</h1>
        </div>
        <span class="status">{{ user.status }}</span>
      </div>
      <div class="card profile-card">
        <dl>
          <div><dt>用户 ID</dt><dd>{{ user.id }}</dd></div>
          <div><dt>昵称</dt><dd>{{ user.nickname || '未设置' }}</dd></div>
          <div><dt>手机号</dt><dd>{{ maskedPhone }}</dd></div>
          <div><dt>手机验证</dt><dd>{{ user.phoneVerified ? '已验证' : '未验证' }}</dd></div>
        </dl>
        <div class="actions">
          <RouterLink class="button" to="/me/families">我的家庭</RouterLink>
          <RouterLink class="button secondary" to="/me/invitations">我的邀请</RouterLink>
          <RouterLink class="button secondary" to="/me/join-requests">我的加入申请</RouterLink>
          <button class="button secondary" :disabled="loggingOut" @click="logout">
            {{ loggingOut ? '退出中...' : '退出登录' }}
          </button>
        </div>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'
import type { UserInfo } from '@/types/api'

const session = useSessionStore()
const router = useRouter()
const loggingOut = ref(false)
const error = ref('')

const user = computed<UserInfo>(() => session.user || {
  id: '-',
  phoneVerified: false,
  status: 'UNKNOWN'
})

const maskedPhone = computed(() => {
  const phone = user.value.phone
  if (!phone) return '未绑定'
  if (phone.includes('*')) return phone
  return phone.replace(/^(\d{3})\d{4}(\d{4})$/, '$1****$2')
})

async function logout() {
  error.value = ''
  loggingOut.value = true
  try {
    await session.logout()
    await router.push('/login')
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '退出请求失败，本地会话已清理。')
    await router.push('/login')
  } finally {
    loggingOut.value = false
  }
}
</script>

<style scoped>
.profile-page {
  padding: 32px 0;
}

.page-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.page-heading h1 {
  margin: 8px 0 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.status {
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: #fff;
  padding: 6px 10px;
  color: var(--color-success);
  font-size: 13px;
  font-weight: 700;
}

.profile-card {
  padding: 22px;
}

dl {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

dt {
  color: var(--color-text-secondary);
}

dd {
  margin: 4px 0 0;
  font-weight: 700;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 22px;
}

.error {
  color: var(--color-danger);
  font-weight: 600;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 760px) {
  dl {
    grid-template-columns: 1fr;
  }
}
</style>
