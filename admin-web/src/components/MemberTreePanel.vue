<template>
  <section class="surface member-tree">
    <div class="panel-title">成员树</div>
    <el-input v-model="keyword" placeholder="搜索成员姓名" clearable />
    <el-tree
      :data="filtered"
      node-key="id"
      default-expand-all
      highlight-current
      :props="{ label: 'name', children: 'children' }"
      @node-click="emit('select', $event)"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import type { MemberNode } from '@/mock/data'

const props = defineProps<{ nodes: MemberNode[] }>()
const emit = defineEmits<{ select: [node: MemberNode] }>()
const keyword = ref('')

const filtered = computed(() => {
  if (!keyword.value) return props.nodes
  return props.nodes.filter((node) => JSON.stringify(node).includes(keyword.value))
})
</script>

<style scoped>
.member-tree {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
}

.panel-title {
  font-weight: 700;
}
</style>
