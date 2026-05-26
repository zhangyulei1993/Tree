<template>
  <div class="card">
    <div class="toolbar">
      <h2 class="page-title">&#x8F6E;&#x64AD;&#x56FE;&#x7BA1;&#x7406;</h2>
      <button class="btn" @click="openCreate">&#x65B0;&#x589E;&#x8F6E;&#x64AD;&#x56FE;</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>&#x56FE;&#x7247;</th>
            <th>&#x6807;&#x9898;</th>
            <th>&#x94FE;&#x63A5;</th>
            <th>&#x6392;&#x5E8F;</th>
            <th>&#x72B6;&#x6001;</th>
            <th>&#x64CD;&#x4F5C;</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td>{{ item.id }}</td>
            <td><img :src="item.image_url" class="cover" /></td>
            <td>{{ item.title }}</td>
            <td>{{ item.link_url }}</td>
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

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">&#x7F16;&#x8F91;&#x8F6E;&#x64AD;&#x56FE;</h3>
        <h3 v-else>&#x65B0;&#x589E;&#x8F6E;&#x64AD;&#x56FE;</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>&#x6807;&#x9898;</label>
            <input v-model="form.title" class="input" />
          </div>

          <div class="form-row">
            <label>&#x56FE;&#x7247;</label>
            <input type="file" accept="image/*" @change="uploadImage" />
            <img v-if="form.image_url" :src="form.image_url" class="preview" />
          </div>

          <div class="form-row">
            <label>&#x94FE;&#x63A5;</label>
            <input v-model="form.link_url" class="input" placeholder="/about" />
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
import { api, type Banner } from '../api'

const list = ref<Banner[]>([])
const showModal = ref(false)

function emptyForm(): Banner {
  return {
    title: '',
    image_url: '',
    link_url: '',
    sort: 0,
    status: 1
  }
}

const form = reactive<Banner>(emptyForm())

onMounted(load)

async function load() {
  list.value = await api.bannerList()
}

function statusText(status: number) {
  return status === 1 ? '\u663E\u793A' : '\u9690\u85CF'
}

function openCreate() {
  Object.assign(form, emptyForm())
  showModal.value = true
}

function openEdit(item: Banner) {
  Object.assign(form, item)
  showModal.value = true
}

async function uploadImage(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const res = await api.upload(file)
  form.image_url = res.url
}

async function save() {
  if (!form.image_url) {
    alert('\u8BF7\u5148\u4E0A\u4F20\u56FE\u7247')
    return
  }

  if (form.id) {
    await api.bannerUpdate(form)
  } else {
    await api.bannerCreate(form)
  }

  showModal.value = false
  load()
}

async function remove(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u5F20\u8F6E\u64AD\u56FE\u5417\uFF1F')) return

  await api.bannerDelete(id)
  load()
}
</script>