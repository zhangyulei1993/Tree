<template>
  <div class="page-stack">
    <PageHeader title="无权限" description="当前 mock 管理员角色不能访问该页面。" />
    <StateBlock description="请在右上角切换 ROOT_ADMIN 或 SUPER_ADMIN 查看受限页面。" action-text="返回仪表盘" />
  </div>
</template>

<script setup lang="ts">
import PageHeader from '@/components/PageHeader.vue'
import StateBlock from '@/components/StateBlock.vue'
</script>
