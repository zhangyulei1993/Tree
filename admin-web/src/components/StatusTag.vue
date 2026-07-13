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
  ACCEPTED: '已接受',
  REJECTED: '已拒绝',
  CANCELLED: '已取消',
  EXPIRED: '已过期',
  DISABLED: '已禁用',
  DELETED: '已删除',
  DISSOLVED: '已解散',
  DISSOLUTION_PENDING: '解散待审',
  DISSOLUTION_COOLDOWN: '恢复冷静期',
  TAKEN_DOWN: '已下架',
  PRIVATE: '私有',
  MERGED: '已合并',
  DRAFT: '草稿',
  PUBLISHED: '已发布',
  ARCHIVED: '已归档',
  SUCCESS: '成功',
  FAILED: '失败'
}

const label = computed(() => labelMap[props.status] || '未知状态')
const type = computed(() => {
  if (['ACTIVE', 'NORMAL', 'APPROVED', 'ACCEPTED', 'PUBLISHED', 'SUCCESS'].includes(props.status)) return 'success'
  if (['PENDING', 'DISSOLUTION_PENDING', 'DISSOLUTION_COOLDOWN', 'DRAFT'].includes(props.status)) return 'warning'
  if (['REJECTED', 'DISABLED', 'DELETED', 'DISSOLVED', 'TAKEN_DOWN', 'FAILED'].includes(props.status)) return 'danger'
  return 'info'
})
</script>
