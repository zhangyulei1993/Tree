<template>
  <el-tag :type="type" effect="plain" round>{{ label }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ role: string }>()

const label = computed(() => ({
  ROOT_ADMIN: '根管理员',
  SUPER_ADMIN: '超级管理员',
  PLATFORM_ADMIN: '平台管理员',
  FAMILY_FOUNDER: '家庭创建人',
  FOUNDER: '家庭创建人',
  FAMILY_ADMIN: '家庭管理员',
  MEMBER: '普通成员'
} as Record<string, string>)[props.role] || props.role || '未设置')

const type = computed(() => {
  if (props.role === 'ROOT_ADMIN') return 'danger'
  if (props.role === 'SUPER_ADMIN') return 'warning'
  if (props.role === 'FOUNDER' || props.role === 'FAMILY_FOUNDER' || props.role === 'FAMILY_ADMIN') return 'success'
  return 'info'
})
</script>
