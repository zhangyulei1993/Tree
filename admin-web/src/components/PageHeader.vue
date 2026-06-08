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
}

h1 {
  margin: 10px 0 0;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0;
}

p {
  margin: 8px 0 0;
  color: var(--color-text-secondary);
  font-size: 13px;
}

.extra {
  display: flex;
  gap: 8px;
}
</style>
