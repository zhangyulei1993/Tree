<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="submit">
      <div class="logo">
        <img src="/brand/favicon.svg" alt="logo" />
      </div>

      <h1>&#x5BB6;&#x8109;&#x4EB2;&#x7F18;&#x540E;&#x53F0;&#x7BA1;&#x7406;</h1>
      <p>&#x7BA1;&#x7406;&#x5B98;&#x7F51;&#x5185;&#x5BB9;&#x3001;&#x8F6E;&#x64AD;&#x56FE;&#x3001;&#x5BB6;&#x65CF;&#x6545;&#x4E8B;&#x3001;&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x4E0E;&#x7559;&#x8A00;&#x3002;</p>

      <label>
        &#x8D26;&#x53F7;
        <input v-model="form.username" class="input" placeholder="admin" />
      </label>

      <label>
        &#x5BC6;&#x7801;
        <input v-model="form.password" class="input" type="password" placeholder="admin123" />
      </label>

      <button class="btn login-btn" type="submit" :disabled="loading">
        <span v-if="loading">&#x767B;&#x5F55;&#x4E2D;...</span>
        <span v-else>&#x767B;&#x5F55;</span>
      </button>

      <p v-if="error" class="error">{{ error }}</p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'

const router = useRouter()
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: 'admin',
  password: 'admin123'
})

async function submit() {
  error.value = ''

  if (!form.username || !form.password) {
    error.value = '\u8BF7\u8F93\u5165\u8D26\u53F7\u548C\u5BC6\u7801'
    return
  }

  loading.value = true

  try {
    const res = await api.login(form)
    localStorage.setItem('admin_token', res.token)
    localStorage.setItem('admin_user', JSON.stringify(res.user))
    router.push('/dashboard')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '\u767B\u5F55\u5931\u8D25'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(circle at 20% 20%, rgba(185, 28, 28, 0.18), transparent 32%),
    linear-gradient(135deg, #fff7ed, #f8ead8);
}

.login-card {
  width: min(420px, 100%);
  background: #fffaf3;
  border: 1px solid #ead8c0;
  border-radius: 24px;
  padding: 34px;
  box-shadow: 0 22px 80px rgba(80, 45, 20, 0.14);
  display: grid;
  gap: 16px;
}

.logo img {
  width: 62px;
  height: 62px;
  border-radius: 18px;
  object-fit: contain;
}

h1 {
  margin: 0;
  font-size: 26px;
}

p {
  margin: 0;
  color: #7c4a2a;
  line-height: 1.7;
}

label {
  display: grid;
  gap: 8px;
  font-weight: 700;
}

.login-btn {
  width: 100%;
}

.error {
  color: #b91c1c;
}
</style>