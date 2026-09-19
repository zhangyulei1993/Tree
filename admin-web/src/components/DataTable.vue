<template>
  <section class="surface data-table" :class="`data-table--${density}`">
    <slot />
  </section>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  density?: 'compact' | 'regular'
}>(), {
  density: 'regular'
})
</script>

<style scoped>
.data-table {
  overflow-x: auto;
  overflow-y: hidden;
  border-radius: 16px;
  box-shadow: 0 10px 28px rgba(18, 32, 51, 0.06);
  overscroll-behavior: contain;
}

.data-table::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.data-table::-webkit-scrollbar-thumb {
  border: 2px solid transparent;
  border-radius: 999px;
  background-clip: padding-box;
  background-color: rgba(31, 58, 95, 0.26);
}

:deep(.el-table) {
  --el-table-header-bg-color: #f3f7f5;
  --el-table-row-hover-bg-color: #f6faf8;
}

:deep(.el-table th.el-table__cell) {
  height: 48px;
  padding: 0;
  border-bottom-color: rgba(47, 107, 87, 0.12);
  color: var(--color-primary);
  font-size: 13px;
  font-weight: 800;
  vertical-align: middle;
}

:deep(.el-table__row) {
  transition: background-color 0.18s ease;
}

:deep(.el-table__body-wrapper) {
  max-height: min(62vh, 640px);
  overflow: auto;
  overscroll-behavior: contain;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar) {
  width: 8px;
  height: 8px;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar-thumb) {
  border: 2px solid transparent;
  border-radius: 999px;
  background-clip: padding-box;
  background-color: rgba(31, 58, 95, 0.26);
}

:deep(.el-table td.el-table__cell) {
  padding: 0;
  border-bottom-color: rgba(148, 163, 184, 0.12);
  vertical-align: middle;
}

:deep(.el-table .cell) {
  min-height: 56px;
  min-width: 0;
  line-height: 1.5;
}

:deep(.el-table th .cell) {
  display: flex;
  align-items: center;
  min-height: 48px;
}

:deep(.el-table .table-column--center .cell) {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100%;
  text-align: center;
}

:deep(.el-table .cell > .table-actions) {
  display: flex;
  align-items: center;
}

:deep(.el-table .cell > .el-tag),
:deep(.el-table .cell > .el-button) {
  flex: 0 0 auto;
}

.data-table--compact :deep(.el-table .cell) {
  min-height: 48px;
}
</style>
