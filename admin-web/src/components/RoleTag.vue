<template>
  <el-tag :type="type" effect="plain" round>{{ role }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ role: string }>()

const type = computed(() => {
  if (props.role === 'ROOT_ADMIN') return 'danger'
  if (props.role === 'SUPER_ADMIN') return 'warning'
  if (props.role === 'FOUNDER') return 'success'
  return 'info'
})
</script>
