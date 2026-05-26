<template>
  <div class="card">
    <div class="toolbar">
      <h2 class="page-title">家族故事</h2>
      <button class="btn" @click="openCreate">新增故事</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>封面</th>
            <th>标题</th>
            <th>摘要</th>
            <th>排序</th>
            <th>状态</th>
            <th>操作</th>
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
              <button class="btn secondary small" @click="openEdit(item)">编辑</button>
              <button class="btn danger small" @click="remove(item.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pager">
      <button class="btn secondary small" :disabled="page <= 1" @click="prev">上一页</button>
      <span>第 {{ page }} 页，共 {{ total }} 条</span>
      <button class="btn secondary small" :disabled="page * pageSize >= total" @click="next">下一页</button>
    </div>

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">编辑故事</h3>
        <h3 v-else>新增故事</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>标题</label>
            <input v-model="form.title" class="input" />
          </div>

          <div class="form-row">
            <label>封面</label>
            <input type="file" accept="image/*" @change="uploadCover" />
            <img v-if="form.cover" :src="form.cover" class="preview" />
          </div>

          <div class="form-row">
            <label>摘要</label>
            <textarea v-model="form.summary" class="textarea"></textarea>
          </div>

          <div class="form-row">
            <label>内容</label>
            <textarea v-model="form.content" class="textarea article-textarea"></textarea>
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