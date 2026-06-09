<template>
  <PageShell>
    <section class="container join-page">
      <form class="card join-card" @submit.prevent="submit">
        <h1>申请加入 {{ family.name }}</h1>
        <p class="muted">加入申请为 mock 流程，不提交真实后端。</p>
        <div v-if="session.state === 'guest'" class="notice">
          请先登录后申请加入。
          <RouterLink class="button" to="/login">去登录</RouterLink>
        </div>
        <div v-else-if="!session.isPhoneBound" class="notice">
          申请加入前需要绑定手机号。
          <button class="button" type="button" @click="session.mockBindPhone()">模拟绑定手机号</button>
        </div>
        <template v-else>
          <textarea v-model.trim="reason" class="field textarea" maxlength="300" placeholder="请填写申请理由" />
          <button class="button" :disabled="submitted">提交 mock 申请</button>
        </template>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
        <p v-if="submitted" class="result">申请已提交，当前状态：PENDING。</p>
      </form>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'

import PageShell from '@/components/PageShell.vue'
import { publicFamilies } from '@/mock/data'
import { useSessionStore } from '@/stores/session'

const route = useRoute()
const session = useSessionStore()
const reason = ref('')
const submitted = ref(false)
const error = ref('')
const family = computed(() => publicFamilies.find((item) => item.id === route.params.familyId) || publicFamilies[0])

function submit() {
  error.value = ''
  if (!reason.value) {
    error.value = '请填写申请理由。'
    return
  }
  submitted.value = true
}
</script>

<style scoped>
.join-page {
  display: grid;
  min-height: 560px;
  place-items: center;
  padding: 32px 0;
}

.join-card {
  display: grid;
  width: min(560px, 100%);
  gap: 14px;
  padding: 24px;
}

.notice {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.result {
  color: var(--color-success);
  font-weight: 700;
}

.error {
  color: var(--color-danger);
  font-weight: 700;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>
