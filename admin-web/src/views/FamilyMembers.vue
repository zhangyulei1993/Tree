<template>
  <div class="page-stack">
    <PageHeader title="家庭成员管理" description="左侧成员树、右侧成员详情和关系操作区。">
      <el-button type="primary">新增成员</el-button>
    </PageHeader>
    <section class="surface family-summary">
      <span>张氏家族</span>
      <span>姓氏：张</span>
      <span>地区：山东济南</span>
      <span>状态：正常</span>
      <span>graph_version：14</span>
    </section>
    <div class="split-grid">
      <MemberTreePanel :nodes="memberTree" @select="selected = $event" />
      <MemberDetailPanel :member="selected" />
    </div>
    <StateBlock description="存在下级成员时禁止删除；ADD_SIBLING 无父母节点时提示先创建父亲或母亲节点。" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import MemberDetailPanel from '@/components/MemberDetailPanel.vue'
import MemberTreePanel from '@/components/MemberTreePanel.vue'
import PageHeader from '@/components/PageHeader.vue'
import StateBlock from '@/components/StateBlock.vue'
import { memberTree, type MemberNode } from '@/mock/data'

const selected = ref<MemberNode>(memberTree[0])
</script>

<style scoped>
.family-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  padding: 14px 16px;
  color: var(--color-text-secondary);
}

.family-summary span:first-child {
  color: var(--color-text-primary);
  font-weight: 700;
}
</style>
