<template>
  <div class="table-cell-stack">
    <strong v-if="title || $slots.title"><slot name="title">{{ title }}</slot></strong>
    <span v-if="description || $slots.description"><slot name="description">{{ description }}</slot></span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  description?: string
}>()
</script>

<style scoped>
.table-cell-stack {
  display: grid;
  min-width: 0;
  gap: 3px;
}

.table-cell-stack strong,
.table-cell-stack span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.table-cell-stack strong {
  color: var(--color-text-primary);
  font-size: 14px;
  font-weight: 700;
}

.table-cell-stack span {
  color: var(--color-text-secondary);
  font-size: 12px;
}
</style>
