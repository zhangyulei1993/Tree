<template>
  <section class="search-panel surface">
    <slot />
  </section>
</template>

<style scoped>
.search-panel {
  border-radius: 16px;
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.10), transparent 180px),
    linear-gradient(135deg, rgba(255, 255, 255, 0.96) 0%, rgba(244, 250, 247, 0.92) 100%);
  padding: 16px 18px;
  box-shadow: 0 10px 28px rgba(18, 32, 51, 0.05);
}

:deep(.el-form) {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 16px;
}

:deep(.el-form-item) {
  align-items: center;
  margin-bottom: 0;
  min-height: 32px;
}

:deep(.el-form-item__label) {
  color: var(--color-text-secondary);
  font-weight: 600;
}

:deep(.el-form-item__content) {
  align-items: center;
}
</style>
