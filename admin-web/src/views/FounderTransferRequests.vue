<template>
  <div class="page-stack">
    <PageHeader title="创始人转让审核" description="仅 ROOT_ADMIN / SUPER_ADMIN 可访问，PLATFORM_ADMIN 不显示入口。" />
    <DataTable>
      <el-table :data="founderTransfers" stripe>
        <el-table-column prop="id" label="申请 ID" width="150" />
        <el-table-column prop="familyName" label="家庭" />
        <el-table-column prop="fromFounder" label="原创始人" />
        <el-table-column prop="toMember" label="目标成员" />
        <el-table-column label="状态"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column prop="createdAt" label="创建时间" />
        <el-table-column label="操作" width="250">
          <template #default>
            <div class="table-actions">
              <el-button size="small">详情</el-button>
              <el-button size="small" type="primary" @click="openHighRisk('approve')">通过</el-button>
              <el-button size="small" type="danger" @click="openHighRisk('reject')">拒绝</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </DataTable>
    <ConfirmDialog v-model="confirmVisible" title="创始人转让审核确认" :message="message" danger @confirm="submit" />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { ref } from 'vue'

import ConfirmDialog from '@/components/ConfirmDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { founderTransfers } from '@/mock/data'

const confirmVisible = ref(false)
const message = ref('')

function openHighRisk(action: 'approve' | 'reject') {
  message.value = action === 'approve' ? '通过后家庭创始人将发生变更，必须二次确认。' : '拒绝创始人转让申请，静态原型不写入后端。'
  confirmVisible.value = true
}

function submit() {
  ElMessage.success('mock 高风险审核已确认')
}
</script>
