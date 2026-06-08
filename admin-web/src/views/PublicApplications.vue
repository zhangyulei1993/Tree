<template>
  <div class="page-stack">
    <PageHeader title="公开申请审核" description="家庭公开展示申请的列表、筛选与统一审核弹窗。" />
    <SearchPanel>
      <el-form>
        <el-form-item label="状态"><el-select v-model="status" clearable placeholder="全部" style="width: 140px"><el-option label="待审核" value="PENDING" /><el-option label="已通过" value="APPROVED" /><el-option label="已拒绝" value="REJECTED" /></el-select></el-form-item>
        <el-form-item label="家庭"><el-input v-model="keyword" placeholder="家庭名称 / 申请人" /></el-form-item>
        <el-button type="primary">查询</el-button>
      </el-form>
    </SearchPanel>
    <DataTable>
      <el-table :data="filtered" stripe>
        <el-table-column prop="id" label="申请 ID" width="140" />
        <el-table-column prop="familyName" label="家庭名称" />
        <el-table-column prop="applicant" label="申请人" />
        <el-table-column prop="reason" label="申请理由" min-width="240" />
        <el-table-column label="状态" width="110"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="190">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button size="small" type="primary" @click="openAudit(row.id, 'approve')">通过</el-button>
              <el-button size="small" type="danger" @click="openAudit(row.id, 'reject')">拒绝</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </DataTable>
    <AuditDialog v-model="auditVisible" :title="auditTitle" :action="auditAction" @submit="submitAudit" />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { computed, ref } from 'vue'

import AuditDialog from '@/components/AuditDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import { publicApplications } from '@/mock/data'

const status = ref('')
const keyword = ref('')
const auditVisible = ref(false)
const auditAction = ref<'approve' | 'reject'>('approve')
const auditTitle = ref('公开申请审核')

const filtered = computed(() => publicApplications.filter((item) => (!status.value || item.status === status.value) && (!keyword.value || JSON.stringify(item).includes(keyword.value))))

function openAudit(id: string, action: 'approve' | 'reject') {
  auditAction.value = action
  auditTitle.value = `公开申请 ${id}`
  auditVisible.value = true
}

function submitAudit() {
  ElMessage.success('mock 审核操作已记录')
}
</script>
