<template>
  <section class="surface data-table">
    <slot />
  </section>
</template>

<style scoped>
.data-table {
  overflow: hidden;
}

:deep(.el-table) {
  --el-table-header-bg-color: #f8fafc;
}

:deep(.el-table th.el-table__cell) {
  font-weight: 700;
}
</style>
