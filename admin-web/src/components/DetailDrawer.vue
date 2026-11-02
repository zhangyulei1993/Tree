<template>
  <el-drawer v-model="visible" :title="title" size="520px">
    <slot />
  </el-drawer>
</template>

<script setup lang="ts">
const visible = defineModel<boolean>({ required: true })

defineProps<{ title: string }>()
</script>
