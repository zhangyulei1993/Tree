<template>
  <el-tag :type="type" effect="light" round>{{ label }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const labelMap: Record<string, string> = {
  ACTIVE: '正常',
  NORMAL: '正常',
  PENDING: '待处理',
  APPROVED: '已通过',
  REJECTED: '已拒绝',
  CANCELLED: '已取消',
  DISABLED: '已禁用',
  DELETED: '已删除',
  DISSOLVED: '已解散',
  DISSOLUTION_PENDING: '解散待审',
  TAKEN_DOWN: '已下架',
  PRIVATE: '私有',
  MERGED: '已合并'
}

const label = computed(() => labelMap[props.status] || props.status)
const type = computed(() => {
  if (['ACTIVE', 'NORMAL', 'APPROVED'].includes(props.status)) return 'success'
  if (['PENDING', 'DISSOLUTION_PENDING'].includes(props.status)) return 'warning'
  if (['REJECTED', 'DISABLED', 'DELETED', 'DISSOLVED', 'TAKEN_DOWN'].includes(props.status)) return 'danger'
  return 'info'
})
</script>
