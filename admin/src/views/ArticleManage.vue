<template>
  <div class="card">
    <div class="toolbar">
      <h2 class="page-title">&#x5BB6;&#x65CF;&#x6545;&#x4E8B;</h2>
      <button class="btn" @click="openCreate">&#x65B0;&#x589E;&#x6545;&#x4E8B;</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>&#x5C01;&#x9762;</th>
            <th>&#x6807;&#x9898;</th>
            <th>&#x6458;&#x8981;</th>
            <th>&#x6392;&#x5E8F;</th>
            <th>&#x72B6;&#x6001;</th>
            <th>&#x64CD;&#x4F5C;</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td>{{ item.id }}</td>
            <td><img v-if="item.cover" :src="item.cover" class="cover" /></td>
            <td>{{ item.title }}</td>
            <td>{{ item.summary }}</td>
            <td>{{ item.sort }}</td>
            <td>{{ statusText(item.status) }}</td>
            <td>
              <button class="btn secondary small" @click="openEdit(item)">&#x7F16;&#x8F91;</button>
              <button class="btn danger small" @click="remove(item.id)">&#x5220;&#x9664;</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pager">
      <button class="btn secondary small" :disabled="page <= 1" @click="prev">&#x4E0A;&#x4E00;&#x9875;</button>
      <span>&#x7B2C; {{ page }} &#x9875;&#xFF0C;&#x5171; {{ total }} &#x6761;</span>
      <button class="btn secondary small" :disabled="page * pageSize >= total" @click="next">&#x4E0B;&#x4E00;&#x9875;</button>
    </div>

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">&#x7F16;&#x8F91;&#x6545;&#x4E8B;</h3>
        <h3 v-else>&#x65B0;&#x589E;&#x6545;&#x4E8B;</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>&#x6807;&#x9898;</label>
            <input v-model="form.title" class="input" />
          </div>

          <div class="form-row">
            <label>&#x5C01;&#x9762;</label>
            <input type="file" accept="image/*" @change="uploadCover" />
            <img v-if="form.cover" :src="form.cover" class="preview" />
          </div>

          <div class="form-row">
            <label>&#x6458;&#x8981;</label>
            <textarea v-model="form.summary" class="textarea"></textarea>
          </div>

          <div class="form-row">
            <label>&#x5185;&#x5BB9;</label>
            <textarea v-model="form.content" class="textarea article-textarea"></textarea>
          </div>

          <div class="form-row">
            <label>&#x6392;&#x5E8F;</label>
            <input v-model.number="form.sort" class="input" type="number" />
          </div>

          <div class="form-row">
            <label>&#x72B6;&#x6001;</label>
            <select v-model.number="form.status" class="select">
              <option :value="1">&#x663E;&#x793A;</option>
              <option :value="0">&#x9690;&#x85CF;</option>
            </select>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn secondary" @click="showModal = false">&#x53D6;&#x6D88;</button>
          <button class="btn" @click="save">&#x4FDD;&#x5B58;</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type Article } from '../api'

const list = ref<Article[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const showModal = ref(false)

function emptyForm(): Article {
  return {
    title: '',
    cover: '',
    summary: '',
    content: '',
    category_id: 0,
    sort: 0,
    status: 1
  }
}

const form = reactive<Article>(emptyForm())

onMounted(load)

async function load() {
  const res = await api.articleList({
    page: page.value,
    page_size: pageSize.value
  })

  list.value = res.list
  total.value = res.total
}

function statusText(status: number) {
  return status === 1 ? '\u663E\u793A' : '\u9690\u85CF'
}

function openCreate() {
  Object.assign(form, emptyForm())
  showModal.value = true
}

function openEdit(item: Article) {
  Object.assign(form, item)
  showModal.value = true
}

async function uploadCover(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const res = await api.upload(file)
  form.cover = res.url
}

async function save() {
  if (!form.title) {
    alert('\u8BF7\u8F93\u5165\u6807\u9898')
    return
  }

  if (form.id) {
    await api.articleUpdate(form)
  } else {
    await api.articleCreate(form)
  }

  showModal.value = false
  load()
}

async function remove(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u7BC7\u6545\u4E8B\u5417\uFF1F')) return

  await api.articleDelete(id)
  load()
}

function prev() {
  if (page.value <= 1) return
  page.value--
  load()
}

function next() {
  if (page.value * pageSize.value >= total.value) return
  page.value++
  load()
}
</script>

<style scoped>
.pager {
  margin-top: 18px;
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: flex-end;
}

.article-textarea {
  min-height: 220px;
}
</style>