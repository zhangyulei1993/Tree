<template>
  <div class="card">
    <div class="toolbar">
      <h2 class="page-title">留言管理</h2>
      <button class="btn secondary" @click="load">刷新</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>姓名</th>
            <th>电话</th>
            <th>邮箱</th>
            <th>留言内容</th>
            <th>提交时间</th>
            <th>操作</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td>{{ item.id }}</td>
            <td>{{ item.name }}</td>
            <td>{{ item.phone }}</td>
            <td>{{ item.email }}</td>
            <td>{{ item.content }}</td>
            <td>{{ item.created_at }}</td>
            <td>
              <button class="btn danger small" @click="remove(item.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, type ContactMessage } from '../api'

const list = ref<ContactMessage[]>([])

onMounted(load)

async function load() {
  list.value = await api.contactList()
}

async function remove(id: number) {
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u6761\u7559\u8A00\u5417\uFF1F')) return

  await api.contactDelete(id)
  load()
}
</script>