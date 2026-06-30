<template>
  <div class="page-stack">
    <PageHeader title="无权限" description="当前管理员角色不能访问该页面。" />
    <StateBlock description="请使用具备相应权限的管理员账号访问。" action-text="返回仪表盘" />
  </div>
</template>

<script setup lang="ts">
import PageHeader from '@/components/PageHeader.vue'
import StateBlock from '@/components/StateBlock.vue'
</script>
