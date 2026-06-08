<template>
  <div class="page-stack">
    <PageHeader title="游客留言审核" description="公开内容与联系方式分区展示，公开列表不展示联系方式。" />
    <SearchPanel>
      <el-form>
        <el-form-item label="状态"><el-select v-model="status" clearable placeholder="全部" style="width: 140px"><el-option label="待审核" value="PENDING" /><el-option label="已通过" value="APPROVED" /><el-option label="已拒绝" value="REJECTED" /></el-select></el-form-item>
        <el-form-item label="关键词"><el-input v-model="keyword" placeholder="家庭 / 留言人 / 内容" /></el-form-item>
        <el-button type="primary">查询</el-button>
      </el-form>
    </SearchPanel>
    <DataTable>
      <el-table :data="filtered" stripe>
        <el-table-column prop="id" label="留言 ID" width="130" />
        <el-table-column prop="familyName" label="家庭" />
        <el-table-column prop="visitorName" label="留言人" />
        <el-table-column prop="publicContent" label="公开内容" min-width="220" />
        <el-table-column prop="contactSummary" label="联系方式摘要" min-width="220" />
        <el-table-column label="状态" width="110"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column label="操作" width="260">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button size="small" type="primary" @click="openAudit(row.id, 'approve')">通过</el-button>
              <el-button size="small" type="danger" @click="openAudit(row.id, 'reject')">拒绝</el-button>
              <el-button size="small" type="danger" plain @click="deleteVisible = true">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </DataTable>
    <AuditDialog v-model="auditVisible" :title="auditTitle" :action="auditAction" @submit="submitAudit" />
    <ConfirmDialog v-model="deleteVisible" title="删除留言" message="删除留言是危险操作，静态原型仅展示二次确认。" danger />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { computed, ref } from 'vue'

import AuditDialog from '@/components/AuditDialog.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import { visitorMessages } from '@/mock/data'

const status = ref('')
const keyword = ref('')
const auditVisible = ref(false)
const deleteVisible = ref(false)
const auditAction = ref<'approve' | 'reject'>('approve')
const auditTitle = ref('留言审核')

const filtered = computed(() => visitorMessages.filter((item) => (!status.value || item.status === status.value) && (!keyword.value || JSON.stringify(item).includes(keyword.value))))

function openAudit(id: string, action: 'approve' | 'reject') {
  auditAction.value = action
  auditTitle.value = `游客留言 ${id}`
  auditVisible.value = true
}

function submitAudit() {
  ElMessage.success('mock 留言审核已记录')
}
</script>
