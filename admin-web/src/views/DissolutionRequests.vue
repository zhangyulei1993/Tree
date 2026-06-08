<template>
  <div class="page-stack">
    <PageHeader title="家庭解散审核" description="仅 ROOT_ADMIN / SUPER_ADMIN 可访问，高风险操作需要二次确认。" />
    <DataTable>
      <el-table :data="dissolutionRequests" stripe>
        <el-table-column prop="id" label="申请 ID" width="160" />
        <el-table-column prop="familyName" label="家庭" />
        <el-table-column prop="requester" label="申请人" />
        <el-table-column prop="reason" label="原因" min-width="220" />
        <el-table-column label="状态"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column prop="createdAt" label="创建时间" />
        <el-table-column label="操作" width="250">
          <template #default>
            <div class="table-actions">
              <el-button size="small">详情</el-button>
              <el-button size="small" type="primary" @click="openConfirm('approve')">通过</el-button>
              <el-button size="small" type="danger" @click="openConfirm('reject')">拒绝</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </DataTable>
    <ConfirmDialog v-model="confirmVisible" title="家庭解散审核确认" :message="message" danger @confirm="submit" />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { ref } from 'vue'

import ConfirmDialog from '@/components/ConfirmDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { dissolutionRequests } from '@/mock/data'

const confirmVisible = ref(false)
const message = ref('')

function openConfirm(action: 'approve' | 'reject') {
  message.value = action === 'approve' ? '通过后家庭将进入已解散状态，公开展示将关闭。' : '拒绝后家庭状态回到正常。'
  confirmVisible.value = true
}

function submit() {
  ElMessage.success('mock 解散审核已确认')
}
</script>
