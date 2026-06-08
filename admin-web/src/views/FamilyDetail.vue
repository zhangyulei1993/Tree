<template>
  <div class="page-stack">
    <PageHeader title="家庭详情" description="家庭基础信息与相关业务记录的 mock tabs。">
      <el-button @click="$router.push(`/admin/families/${family.id}/members`)">成员管理</el-button>
    </PageHeader>
    <section class="surface detail-card">
      <el-descriptions :column="4" border>
        <el-descriptions-item label="家庭 ID">{{ family.id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ family.name }}</el-descriptions-item>
        <el-descriptions-item label="姓氏">{{ family.surname }}</el-descriptions-item>
        <el-descriptions-item label="状态"><StatusTag :status="family.status" /></el-descriptions-item>
        <el-descriptions-item label="籍贯">{{ family.nativePlace }}</el-descriptions-item>
        <el-descriptions-item label="地区">{{ family.regionText }}</el-descriptions-item>
        <el-descriptions-item label="公开状态"><StatusTag :status="family.publicDisplayStatus" /></el-descriptions-item>
        <el-descriptions-item label="graph_version">{{ family.graphVersion }}</el-descriptions-item>
      </el-descriptions>
    </section>
    <section class="surface tabs">
      <el-tabs model-value="members">
        <el-tab-pane label="基础信息" name="base"><JsonViewer :value="family" /></el-tab-pane>
        <el-tab-pane label="成员" name="members">
          <el-table :data="memberRows">
            <el-table-column prop="id" label="成员 ID" />
            <el-table-column prop="name" label="姓名" />
            <el-table-column prop="role" label="家庭角色"><template #default="{ row }"><RoleTag :role="row.role" /></template></el-table-column>
            <el-table-column prop="bindingStatus" label="绑定状态" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="公开申请" name="public"><el-table :data="publicApplications"><el-table-column prop="id" label="申请 ID" /><el-table-column prop="status" label="状态" /></el-table></el-tab-pane>
        <el-tab-pane label="加入申请" name="join"><StateBlock description="加入申请记录 mock empty 状态" /></el-tab-pane>
        <el-tab-pane label="邀请记录" name="invite"><StateBlock description="邀请记录 mock empty 状态" /></el-tab-pane>
        <el-tab-pane label="留言" name="messages"><el-table :data="visitorMessages"><el-table-column prop="id" label="留言 ID" /><el-table-column prop="publicContent" label="公开内容" /></el-table></el-tab-pane>
        <el-tab-pane label="日志" name="logs"><JsonViewer :value="{ familyId: family.id, logs: ['CREATE_FAMILY', 'UPDATE_PUBLIC_STATUS'] }" /></el-tab-pane>
      </el-tabs>
    </section>
  </div>
</template>

<script setup lang="ts">
import JsonViewer from '@/components/JsonViewer.vue'
import PageHeader from '@/components/PageHeader.vue'
import RoleTag from '@/components/RoleTag.vue'
import StateBlock from '@/components/StateBlock.vue'
import StatusTag from '@/components/StatusTag.vue'
import { families, memberTree, publicApplications, visitorMessages } from '@/mock/data'

const family = families[0]
const memberRows = [memberTree[0], ...(memberTree[0].children || [])]
</script>

<style scoped>
.detail-card,
.tabs {
  padding: 16px;
}
</style>
