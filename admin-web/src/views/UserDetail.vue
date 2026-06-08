<template>
  <div class="page-stack">
    <PageHeader title="用户详情" description="基础资料、绑定关系、状态记录和操作记录 mock 展示。" />
    <section class="surface detail-card">
      <el-descriptions :column="3" border>
        <el-descriptions-item label="用户 ID">{{ user.id }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ user.maskedPhone }}</el-descriptions-item>
        <el-descriptions-item label="状态"><StatusTag :status="user.status" /></el-descriptions-item>
        <el-descriptions-item label="昵称">{{ user.nickname }}</el-descriptions-item>
        <el-descriptions-item label="真实姓名">{{ user.realName }}</el-descriptions-item>
        <el-descriptions-item label="账号来源">{{ user.origin }}</el-descriptions-item>
        <el-descriptions-item label="注册端">{{ user.clientType }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ user.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="最后登录">{{ user.lastLoginAt }}</el-descriptions-item>
      </el-descriptions>
    </section>
    <el-tabs class="surface tabs" model-value="families">
      <el-tab-pane label="加入家庭" name="families">
        <el-table :data="families.slice(0, 2)">
          <el-table-column prop="id" label="家庭 ID" />
          <el-table-column prop="name" label="家庭名称" />
          <el-table-column prop="surname" label="姓氏" />
          <el-table-column prop="status" label="状态">
            <template #default="{ row }"><StatusTag :status="row.status" /></template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="状态记录" name="status"><JsonViewer :value="{ disabled: false, phoneVerified: user.phoneVerified, notes: 'mock status history' }" /></el-tab-pane>
      <el-tab-pane label="操作记录" name="logs"><JsonViewer :value="{ recentActions: ['LOGIN_PHONE', 'CREATE_FAMILY'], sensitive: 'omitted' }" /></el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import JsonViewer from '@/components/JsonViewer.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { families, users } from '@/mock/data'

const user = users[0]
</script>

<style scoped>
.detail-card,
.tabs {
  padding: 16px;
}
</style>
