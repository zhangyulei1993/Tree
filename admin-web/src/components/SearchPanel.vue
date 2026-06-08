<template>
  <section class="search-panel surface">
    <slot />
  </section>
</template>

<style scoped>
.search-panel {
  padding: 16px;
}

:deep(.el-form) {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 16px;
}

:deep(.el-form-item) {
  margin-bottom: 0;
}
</style>
