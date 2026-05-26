<template>
  <div class="card">
    <h2 class="page-title">官网配置</h2>

    <div class="form-grid">
      <div class="form-row">
        <label>网站名称</label>
        <input v-model="form.site_name" class="input" />
      </div>

      <div class="form-row">
        <label>Logo</label>
        <input type="file" accept="image/*" @change="uploadLogo" />
        <img v-if="form.logo" :src="form.logo" class="preview" />
      </div>

      <div class="form-row">
        <label>电话</label>
        <input v-model="form.phone" class="input" />
      </div>

      <div class="form-row">
        <label>邮箱</label>
        <input v-model="form.email" class="input" />
      </div>

      <div class="form-row">
        <label>地址</label>
        <input v-model="form.address" class="input" />
      </div>

      <div class="form-row">
        <label>简介</label>
        <textarea v-model="form.description" class="textarea"></textarea>
      </div>

      <button class="btn" @click="save">保存配置</button>

      <p v-if="message">{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type SiteConfig } from '../api'

const message = ref('')

const form = reactive<SiteConfig>({
  site_name: '',
  logo: '',
  phone: '',
  email: '',
  address: '',
  description: ''
})

onMounted(load)

async function load() {
  const res = await api.getSiteConfig()
  Object.assign(form, res)
}

async function uploadLogo(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const res = await api.upload(file)
  form.logo = res.url
}

async function save() {
  const res = await api.updateSiteConfig(form)
  Object.assign(form, res)
  message.value = '\u4FDD\u5B58\u6210\u529F'
}
</script>