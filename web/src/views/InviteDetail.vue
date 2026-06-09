<template>
  <PageShell>
    <section class="container invite-page">
      <div class="card invite-card">
        <span class="eyebrow">邀请确认</span>
        <h1>你收到一个家庭成员绑定邀请</h1>
        <dl>
          <div><dt>家庭</dt><dd>{{ invitation.familyName }}</dd></div>
          <div><dt>成员</dt><dd>{{ invitation.memberName }}</dd></div>
          <div><dt>地区</dt><dd>{{ invitation.regionText }}</dd></div>
          <div><dt>有效期至</dt><dd>{{ invitation.expiresAt }}</dd></div>
        </dl>
        <div v-if="session.state === 'guest'" class="notice">
          请先登录后查看完整邀请。
          <RouterLink class="button" to="/login">去登录</RouterLink>
        </div>
        <div v-else-if="session.state === 'wechatLoggedInPendingPhone'" class="notice">
          当前已微信登录但未绑定手机号，请先绑定手机号。
          <button class="button" @click="session.mockBindPhone()">模拟绑定手机号</button>
        </div>
        <div v-else class="actions">
          <button class="button" @click="result = '已接受邀请，mock 绑定成功。'">接受绑定</button>
          <button class="button secondary" @click="result = '已拒绝邀请，mock 状态已更新。'">拒绝</button>
        </div>
        <p v-if="result" class="result">{{ result }}</p>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import PageShell from '@/components/PageShell.vue'
import { invitation } from '@/mock/data'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const result = ref('')
</script>

<style scoped>
.invite-page {
  display: grid;
  min-height: 560px;
  place-items: center;
  padding: 32px 0;
}

.invite-card {
  width: min(640px, 100%);
  padding: 24px;
}

.eyebrow {
  color: var(--color-success);
  font-weight: 700;
}

dl {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

dt {
  color: var(--color-text-secondary);
}

dd {
  margin: 4px 0 0;
  font-weight: 700;
}

.notice,
.actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 18px;
}

.result {
  color: var(--color-success);
  font-weight: 700;
}
</style>
