<template>
  <PageShell>
    <section class="container tree-page">
      <h1>张氏家族公开树</h1>
      <p class="muted">LIST_TREE 轻量展示，不包含 userId、手机号、openid、unionid、token 或绑定用户信息。</p>
      <PublicTreeList :nodes="treeNodes" />
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import PageShell from '@/components/PageShell.vue'
import PublicTreeList from '@/components/PublicTreeList.vue'
import { treeNodes } from '@/mock/data'
</script>

<style scoped>
.tree-page {
  padding: 32px 0;
}

h1 {
  font-size: 34px;
}
</style>
