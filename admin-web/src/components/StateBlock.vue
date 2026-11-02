<template>
  <div class="surface state-block">
    <el-empty :description="description">
      <el-button v-if="actionText" type="primary">{{ actionText }}</el-button>
    </el-empty>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  description: string
  actionText?: string
}>()
</script>

<style scoped>
.state-block {
  padding: 20px;
}
</style>
