<template>
  <section class="section">
    <div class="container contact">
      <div>
        <p class="eyebrow">&#x8054;&#x7CFB;&#x6211;&#x4EEC;</p>
        <h1 class="section-title">&#x8BA9;&#x6211;&#x4EEC;&#x4E00;&#x8D77;&#x68B3;&#x7406;&#x5BB6;&#x5EAD;&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;</h1>
        <p class="section-subtitle">
          &#x5982;&#x679C;&#x4F60;&#x60F3;&#x6574;&#x7406;&#x5BB6;&#x8C31;&#x3001;&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x56FE;&#x3001;&#x5BB6;&#x65CF;&#x6545;&#x4E8B;&#x6216;&#x957F;&#x8F88;&#x8BB0;&#x5FC6;&#xFF0C;&#x53EF;&#x4EE5;&#x901A;&#x8FC7;&#x8868;&#x5355;&#x7559;&#x4E0B;&#x4FE1;&#x606F;&#x3002;
        </p>

        <div class="card contact-info">
          <p><strong>&#x7535;&#x8BDD;&#xFF1A;</strong>{{ site.phone || '-' }}</p>
          <p><strong>&#x90AE;&#x7BB1;&#xFF1A;</strong>{{ site.email || '-' }}</p>
          <p><strong>&#x5730;&#x5740;&#xFF1A;</strong>{{ site.address || '-' }}</p>
        </div>
      </div>

      <form class="card form" @submit.prevent="submit">
        <label>
          &#x59D3;&#x540D;
          <input v-model="form.name" class="input" placeholder="&#x8BF7;&#x8F93;&#x5165;&#x60A8;&#x7684;&#x59D3;&#x540D;" />
        </label>

        <label>
          &#x7535;&#x8BDD;
          <input v-model="form.phone" class="input" placeholder="&#x8BF7;&#x8F93;&#x5165;&#x8054;&#x7CFB;&#x7535;&#x8BDD;" />
        </label>

        <label>
          &#x90AE;&#x7BB1;
          <input v-model="form.email" class="input" placeholder="&#x8BF7;&#x8F93;&#x5165;&#x90AE;&#x7BB1;" />
        </label>

        <label>
          &#x7559;&#x8A00;&#x5185;&#x5BB9;
          <textarea v-model="form.content" class="textarea" placeholder="&#x8BF7;&#x7B80;&#x5355;&#x8BF4;&#x660E;&#x60A8;&#x7684;&#x9700;&#x6C42;"></textarea>
        </label>

        <button class="btn" type="submit" :disabled="submitting">
          <span v-if="submitting">&#x63D0;&#x4EA4;&#x4E2D;...</span>
          <span v-else>&#x63D0;&#x4EA4;&#x7559;&#x8A00;</span>
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