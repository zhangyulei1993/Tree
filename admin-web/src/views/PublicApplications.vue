<template>
  <div class="page-stack">
    <PageHeader title="公开申请审核" description="审核家庭公开展示申请。" />

    <SearchPanel>
      <el-form>
        <el-form-item label="状态">
          <el-select v-model="status" clearable placeholder="全部" style="width: 150px">
            <el-option label="待审核" value="PENDING" />
            <el-option label="已通过" value="APPROVED" />
            <el-option label="已拒绝" value="REJECTED" />
            <el-option label="已取消" value="CANCELLED" />
          </el-select>
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="search">查询</el-button>
        <el-button :disabled="loading" @click="resetSearch">重置</el-button>
      </el-form>
    </SearchPanel>

    <el-alert v-if="loadError" :title="`申请列表加载失败：${loadError}`" type="error" show-icon :closable="false" />
    <el-alert v-if="operationError" :title="`操作失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <DataTable>
      <el-table v-loading="loading" :data="items" stripe empty-text="暂无公开申请">
        <el-table-column prop="applicationId" label="申请 ID" width="110" />
        <el-table-column prop="familyId" label="家庭 ID" width="100" />
        <el-table-column prop="familyName" label="家庭名称" min-width="150" />
        <el-table-column prop="applicantUserId" label="申请用户 ID" width="120" />
        <el-table-column prop="reason" label="申请理由" min-width="240" show-overflow-tooltip />
        <el-table-column label="状态" width="110">
          <template #default="{ row }"><StatusTag :status="row.status" /></template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div v-if="row.status === 'PENDING'" class="table-actions">
              <el-button size="small" type="primary" @click="openAudit(row.applicationId, 'approve')">通过</el-button>
              <el-button size="small" type="danger" @click="openAudit(row.applicationId, 'reject')">拒绝</el-button>
            </div>
            <span v-else class="muted">已处理</span>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        class="pager"
        layout="total, sizes, prev, pager, next"
        :total="total"
        :current-page="page"
        :page-size="pageSize"
        :page-sizes="[10, 20, 50]"
        @current-change="changePage"
        @size-change="changePageSize"
      />
    </DataTable>

    <AuditDialog
      v-model="auditVisible"
      :title="auditTitle"
      :action="auditAction"
      :submitting="submitting"
      :close-on-submit="false"
      @submit="submitAudit"
    />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, ref } from 'vue'

import { getApiErrorMessage } from '@/api/client'
import {
  approvePublicApplication,
  listPublicApplications,
  rejectPublicApplication
} from '@/api/publicApplications'
import AuditDialog from '@/components/AuditDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { PublicApplication } from '@/types/api'

const status = ref('')
const items = ref<PublicApplication[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const auditVisible = ref(false)
const auditAction = ref<'approve' | 'reject'>('approve')
const auditTitle = ref('公开申请审核')
const selectedApplicationId = ref<number | null>(null)
const submitting = ref(false)

onMounted(loadApplications)

async function loadApplications() {
  loading.value = true
  loadError.value = ''
  try {
    const result = await listPublicApplications({
      status: status.value || undefined,
      page: page.value,
      pageSize: pageSize.value
    })
    items.value = result.items
    total.value = result.total
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  void loadApplications()
}

function resetSearch() {
  status.value = ''
  page.value = 1
  void loadApplications()
}

function changePage(value: number) {
  page.value = value
  void loadApplications()
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  void loadApplications()
}

function openAudit(applicationId: number, action: 'approve' | 'reject') {
  operationError.value = ''
  selectedApplicationId.value = applicationId
  auditAction.value = action
  auditTitle.value = `公开申请 ${applicationId}`
  auditVisible.value = true
}

async function submitAudit(reviewComment: string) {
  if (!selectedApplicationId.value || submitting.value) return
  submitting.value = true
  operationError.value = ''
  try {
    const input = { reviewComment: reviewComment.trim() || undefined }
    if (auditAction.value === 'approve') {
      await approvePublicApplication(selectedApplicationId.value, input)
    } else {
      await rejectPublicApplication(selectedApplicationId.value, input)
    }
    auditVisible.value = false
    ElMessage.success(auditAction.value === 'approve' ? '公开申请已通过' : '公开申请已驳回')
    await loadApplications()
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function formatTime(value: string) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
}
</script>

<style scoped>
.pager {
  justify-content: flex-end;
  padding: 14px;
}
</style>
