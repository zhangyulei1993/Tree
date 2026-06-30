<template>
  <section class="surface data-table">
    <slot />
  </section>
</template>

<style scoped>
.data-table {
  overflow: hidden;
  border-radius: 24px;
  box-shadow: 0 18px 46px rgba(18, 32, 51, 0.07);
}

:deep(.el-table) {
  --el-table-header-bg-color: #f3f7f5;
  --el-table-row-hover-bg-color: #f6faf8;
}

:deep(.el-table th.el-table__cell) {
  border-bottom-color: rgba(47, 107, 87, 0.12);
  color: var(--color-primary);
  font-size: 13px;
  font-weight: 800;
}

:deep(.el-table td.el-table__cell) {
  border-bottom-color: rgba(148, 163, 184, 0.12);
}

:deep(.el-table .cell) {
  line-height: 1.5;
}
</style>
