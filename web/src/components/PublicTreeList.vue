<template>
  <section class="tree-list">
    <article v-for="node in nodes" :key="node.memberId" class="card tree-node">
      <div>
        <h3>{{ node.name }}</h3>
        <p>{{ node.gender }} · {{ node.birthText }}<template v-if="node.deathText"> - {{ node.deathText }}</template></p>
      </div>
      <p>{{ node.description }}</p>
      <div class="relations">
        <span>父母：{{ names(node.parentIds) || '未展示' }}</span>
        <span>配偶：{{ names(node.spouseIds) || '无' }}</span>
        <span>子女：{{ names(node.childrenIds) || '无' }}</span>
      </div>
    </article>
  </section>
</template>

<script setup lang="ts">
import type { TreeNode } from '@/mock/data'

const props = defineProps<{ nodes: TreeNode[] }>()

function names(ids: string[]) {
  return ids.map((id) => props.nodes.find((node) => node.memberId === id)?.name).filter(Boolean).join('、')
}
</script>

<style scoped>
.tree-list {
  display: grid;
  gap: 14px;
}

.tree-node {
  display: grid;
  gap: 10px;
  padding: 20px;
}

h3,
p {
  margin: 0;
}

.relations {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  color: var(--color-text-secondary);
  font-size: 13px;
}

.relations span {
  border-radius: 999px;
  background: var(--color-bg-soft);
  padding: 6px 10px;
}
</style>
