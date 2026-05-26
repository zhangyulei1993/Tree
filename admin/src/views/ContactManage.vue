<template>
  <div class="card">
    <div class="toolbar">
      <h2 class="page-title">&#x7559;&#x8A00;&#x7BA1;&#x7406;</h2>
      <button class="btn secondary" @click="load">&#x5237;&#x65B0;</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>&#x59D3;&#x540D;</th>
            <th>&#x7535;&#x8BDD;</th>
            <th>&#x90AE;&#x7BB1;</th>
            <th>&#x7559;&#x8A00;&#x5185;&#x5BB9;</th>
            <th>&#x63D0;&#x4EA4;&#x65F6;&#x95F4;</th>
            <th>&#x64CD;&#x4F5C;</th>
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
              <button class="btn danger small" @click="remove(item.id)">&#x5220;&#x9664;</button>
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