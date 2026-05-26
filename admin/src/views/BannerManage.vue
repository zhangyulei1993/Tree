<template>
  <div class="card">
    <div class="toolbar">
      <h2 class="page-title">轮播图管理</h2>
      <button class="btn" @click="openCreate">新增轮播图</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>图片</th>
            <th>标题</th>
            <th>链接</th>
            <th>排序</th>
            <th>状态</th>
            <th>操作</th>
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
              <button class="btn secondary small" @click="openEdit(item)">编辑</button>
              <button class="btn danger small" @click="remove(item.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">编辑轮播图</h3>
        <h3 v-else>新增轮播图</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>标题</label>
            <input v-model="form.title" class="input" />
          </div>

          <div class="form-row">
            <label>图片</label>
            <input type="file" accept="image/*" @change="uploadImage" />
            <img v-if="form.image_url" :src="form.image_url" class="preview" />
          </div>

          <div class="form-row">
            <label>链接</label>
            <input v-model="form.link_url" class="input" placeholder="/about" />
          </div>

          <div class="form-row">
            <label>排序</label>
            <input v-model.number="form.sort" class="input" type="number" />
          </div>

          <div class="form-row">
            <label>状态</label>
            <select v-model.number="form.status" class="select">
              <option :value="1">显示</option>
              <option :value="0">隐藏</option>
            </select>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn secondary" @click="showModal = false">取消</button>
          <button class="btn" @click="save">保存</button>
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
  return status === 1 ? '显示' : '隐藏'
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
    alert('请先上传图片')
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
  if (!confirm('确定删除这张轮播图吗？')) return

  await api.bannerDelete(id)
  load()
}
</script>