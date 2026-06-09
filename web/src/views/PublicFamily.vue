<template>
  <PageShell>
    <section class="container family-page">
      <div class="card profile">
        <span class="eyebrow">公开家庭主页</span>
        <h1>{{ family.name }}</h1>
        <p>{{ family.description }}</p>
        <dl>
          <div><dt>姓氏</dt><dd>{{ family.surname }}</dd></div>
          <div><dt>籍贯</dt><dd>{{ family.nativePlace }}</dd></div>
          <div><dt>地区</dt><dd>{{ family.regionText }}</dd></div>
          <div><dt>创始人</dt><dd>{{ family.founderName }}</dd></div>
        </dl>
        <p class="contact">{{ family.publicContact || '该家庭未设置公开联系方式，如需联系请通过平台协助。' }}</p>
        <div class="actions">
          <RouterLink class="button" :to="`/families/${family.id}/tree/public`">查看公开树</RouterLink>
          <RouterLink class="button secondary" :to="`/families/${family.id}/join`">申请加入</RouterLink>
        </div>
      </div>
      <div class="grid two">
        <section class="card panel">
          <h2>游客留言</h2>
          <article v-for="message in messages" :key="message.id" class="message">
            <strong>{{ message.visitorName }}</strong>
            <p>{{ message.content }}</p>
            <span>{{ message.createdAt }}</span>
          </article>
        </section>
        <section class="card panel">
          <h2>提交留言</h2>
          <form class="message-form" @submit.prevent="submitMessage">
            <input v-model.trim="visitorName" class="field" maxlength="20" placeholder="访客称呼" />
            <textarea
              v-model.trim="messageContent"
              class="field textarea"
              maxlength="300"
              placeholder="留言内容，静态原型不提交真实后端"
            />
            <div class="form-meta">
              <span>{{ messageContent.length }}/300</span>
              <button class="button" :disabled="submitting">
                {{ submitting ? '提交中...' : '提交 mock 留言' }}
              </button>
            </div>
          </form>
          <p v-if="formError" class="feedback error" role="alert">{{ formError }}</p>
          <p v-if="submitResult" class="feedback success" role="status">{{ submitResult }}</p>
        </section>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'

import PageShell from '@/components/PageShell.vue'
import { publicFamilies, visitorMessages } from '@/mock/data'

const route = useRoute()
const family = computed(() => publicFamilies.find((item) => item.id === route.params.familyId) || publicFamilies[0])
const localMessages = ref([...visitorMessages])
const visitorName = ref('')
const messageContent = ref('')
const formError = ref('')
const submitResult = ref('')
const submitting = ref(false)
const messages = computed(() => localMessages.value.filter((item) => item.familyId === family.value.id))

function submitMessage() {
  formError.value = ''
  submitResult.value = ''

  if (!visitorName.value) {
    formError.value = '请填写访客称呼。'
    return
  }
  if (!messageContent.value) {
    formError.value = '请填写留言内容。'
    return
  }

  submitting.value = true
  localMessages.value.unshift({
    id: `message_mock_${Date.now()}`,
    familyId: family.value.id,
    visitorName: visitorName.value,
    content: messageContent.value,
    createdAt: new Date().toLocaleDateString('zh-CN')
  })
  visitorName.value = ''
  messageContent.value = ''
  submitting.value = false
  submitResult.value = '留言已提交。当前为 mock 原型，正式环境中需审核后公开。'
}
</script>

<style scoped>
.family-page {
  display: grid;
  gap: 18px;
  padding: 32px 0;
}

.profile,
.panel {
  padding: 22px;
}

.eyebrow {
  color: var(--color-success);
  font-weight: 700;
}

h1 {
  margin: 10px 0;
  font-size: 40px;
}

p {
  line-height: 1.8;
}

dl,
.two {
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

.contact {
  border-radius: 10px;
  background: #f4f1e8;
  padding: 12px;
}

.actions,
.panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message-form {
  display: grid;
  gap: 12px;
}

.form-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--color-text-secondary);
  font-size: 13px;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.feedback {
  margin: 0;
  font-weight: 700;
}

.feedback.error {
  color: var(--color-danger);
}

.feedback.success {
  color: var(--color-success);
}

.message {
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 12px;
}

.two {
  grid-template-columns: 1fr 1fr;
}

@media (max-width: 820px) {
  dl,
  .two {
    grid-template-columns: 1fr;
  }
}
</style>
