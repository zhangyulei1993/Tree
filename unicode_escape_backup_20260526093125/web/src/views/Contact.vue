<template>
  <section class="section">
    <div class="container contact">
      <div>
        <p class="eyebrow">联系我们</p>
        <h1 class="section-title">让我们一起梳理家庭亲属关系</h1>
        <p class="section-subtitle">
          如果你想整理家谱、亲属关系图、家族故事或长辈记忆，可以通过表单留下信息。
        </p>

        <div class="card contact-info">
          <p><strong>电话：</strong>{{ site.phone || '-' }}</p>
          <p><strong>邮箱：</strong>{{ site.email || '-' }}</p>
          <p><strong>地址：</strong>{{ site.address || '-' }}</p>
        </div>
      </div>

      <form class="card form" @submit.prevent="submit">
        <label>
          姓名
          <input v-model="form.name" class="input" placeholder="请输入您的姓名" />
        </label>

        <label>
          电话
          <input v-model="form.phone" class="input" placeholder="请输入联系电话" />
        </label>

        <label>
          邮箱
          <input v-model="form.email" class="input" placeholder="请输入邮箱" />
        </label>

        <label>
          留言内容
          <textarea v-model="form.content" class="textarea" placeholder="请简单说明您的需求"></textarea>
        </label>

        <button class="btn" type="submit" :disabled="submitting">
          <span v-if="submitting">提交中...</span>
          <span v-else>提交留言</span>
        </button>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type SiteConfig } from '../api'

const submitting = ref(false)

const site = reactive<SiteConfig>({
  site_name: '',
  logo: '',
  phone: '',
  email: '',
  address: '',
  description: ''
})

const form = reactive({
  name: '',
  phone: '',
  email: '',
  content: ''
})

onMounted(async () => {
  try {
    const res = await api.siteConfig()
    Object.assign(site, res)
  } catch {
    // ignore
  }
})

async function submit() {
  if (!form.name && !form.phone && !form.email) {
    alert('\u8BF7\u81F3\u5C11\u586B\u5199\u59D3\u540D\u3001\u7535\u8BDD\u6216\u90AE\u7BB1\u4E2D\u7684\u4E00\u9879')
    return
  }

  submitting.value = true

  try {
    await api.submitContact(form)
    alert('\u63D0\u4EA4\u6210\u529F\uFF0C\u6211\u4EEC\u4F1A\u5C3D\u5FEB\u4E0E\u60A8\u8054\u7CFB')
    form.name = ''
    form.phone = ''
    form.email = ''
    form.content = ''
  } catch (err) {
    alert(err instanceof Error ? err.message : '\u63D0\u4EA4\u5931\u8D25')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.contact {
  display: grid;
  grid-template-columns: 0.85fr 1.15fr;
  gap: 34px;
  align-items: start;
}

.eyebrow {
  margin: 0 0 10px;
  color: #b91c1c;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.contact-info {
  padding: 22px;
}

.contact-info p {
  line-height: 1.8;
  color: #7c4a2a;
}

.form {
  display: grid;
  gap: 18px;
  padding: 28px;
}

.form label {
  display: grid;
  gap: 8px;
  font-weight: 700;
}

.form button {
  width: fit-content;
}

@media (max-width: 768px) {
  .contact {
    grid-template-columns: 1fr;
  }

  .form button {
    width: 100%;
  }
}
</style>