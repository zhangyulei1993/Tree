<template>
  <div class="card">
    <div class="toolbar">
      <h2 class="page-title">&#x5BB6;&#x65CF;&#x7BA1;&#x7406;</h2>
      <button class="btn" @click="openCreate">&#x65B0;&#x589E;&#x5BB6;&#x65CF;</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>&#x5C01;&#x9762;</th>
            <th>&#x5BB6;&#x65CF;&#x540D;&#x79F0;</th>
            <th>&#x59D3;&#x6C0F;</th>
            <th>&#x7956;&#x7C4D;</th>
            <th>&#x72B6;&#x6001;</th>
            <th>&#x64CD;&#x4F5C;</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td>{{ item.id }}</td>
            <td><img v-if="item.cover" :src="item.cover" class="cover" /></td>
            <td>{{ item.name }}</td>
            <td>{{ item.surname }}</td>
            <td>{{ item.origin }}</td>
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
        <h3 v-if="form.id">&#x7F16;&#x8F91;&#x5BB6;&#x65CF;</h3>
        <h3 v-else>&#x65B0;&#x589E;&#x5BB6;&#x65CF;</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>&#x5BB6;&#x65CF;&#x540D;&#x79F0;</label>
            <input v-model="form.name" class="input" placeholder="&#x5F20;&#x6C0F;&#x5BB6;&#x65CF;" />
          </div>

          <div class="form-row">
            <label>&#x59D3;&#x6C0F;</label>
            <input v-model="form.surname" class="input" placeholder="&#x5F20;" />
          </div>

          <div class="form-row">
            <label>&#x7956;&#x7C4D;</label>
            <input v-model="form.origin" class="input" placeholder="&#x6D59;&#x6C5F;&#x676D;&#x5DDE;" />
          </div>

          <div class="form-row">
            <label>&#x5C01;&#x9762;</label>
            <input type="file" accept="image/*" @change="uploadCover" />
            <img v-if="form.cover" :src="form.cover" class="preview" />
          </div>

          <div class="form-row">
            <label>&#x5BB6;&#x65CF;&#x7B80;&#x4ECB;</label>
            <textarea v-model="form.description" class="textarea"></textarea>
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
import { api, type Family } from '../api'

const list = ref<Family[]>([])
const showModal = ref(false)

function emptyForm(): Family {
  return {
    name: '',
    surname: '',
    origin: '',
    cover: '',
    description: '',
    status: 1
  }
}

const form = reactive<Family>(emptyForm())

onMounted(load)

async function load() {
  list.value = await api.familyList()
}

function statusText(status: number) {
  return status === 1 ? '\u663E\u793A' : '\u9690\u85CF'
}

function openCreate() {
  Object.assign(form, emptyForm())
  showModal.value = true
}

function openEdit(item: Family) {
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
  if (!form.name) {
    alert('\u8BF7\u8F93\u5165\u5BB6\u65CF\u540D\u79F0')
    return
  }

  if (form.id) {
    await api.familyUpdate(form)
  } else {
    await api.familyCreate(form)
  }

  showModal.value = false
  await load()
}

async function remove(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u4E2A\u5BB6\u65CF\u5417\uFF1F\u76F8\u5173\u6210\u5458\u548C\u5173\u7CFB\u4E5F\u4F1A\u4E00\u8D77\u5220\u9664\u3002')) return

  await api.familyDelete(id)
  await load()
}
</script>