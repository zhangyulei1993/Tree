<template>
  <div class="page-header">
    <div>
      <el-breadcrumb separator="/">
        <el-breadcrumb-item>后台</el-breadcrumb-item>
        <el-breadcrumb-item>{{ title }}</el-breadcrumb-item>
      </el-breadcrumb>
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div class="extra">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  padding: 8px 2px 4px;
}

h1 {
  margin: 10px 0 0;
  color: var(--color-primary);
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

p {
  margin: 8px 0 0;
  color: var(--color-text-secondary);
  max-width: 720px;
  font-size: 14px;
  line-height: 1.7;
}

.extra {
  display: flex;
  gap: 8px;
}

@media (max-width: 860px) {
  .page-header {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
