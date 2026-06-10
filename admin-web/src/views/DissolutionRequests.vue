<template>
  <div class="page-stack">
    <PageHeader title="家庭解散审核" description="仅 ROOT_ADMIN / SUPER_ADMIN 可审核。解散和恢复均为高风险操作。" />

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
        <el-form-item label="家庭 ID">
          <el-input-number v-model="familyId" :min="1" :controls="false" placeholder="全部" />
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="search">查询</el-button>
        <el-button :disabled="loading" @click="resetSearch">重置</el-button>
      </el-form>
    </SearchPanel>

    <el-alert v-if="loadError" :title="`解散申请列表加载失败：${loadError}`" type="error" show-icon :closable="false" />
    <el-alert v-if="operationError" :title="`操作失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <DataTable>
      <el-table v-loading="loading" :data="items" stripe empty-text="暂无家庭解散申请">
        <el-table-column prop="requestId" label="申请 ID" width="100" />
        <el-table-column prop="familyId" label="家庭 ID" width="100" />
        <el-table-column prop="familyName" label="家庭名称" min-width="150" />
        <el-table-column label="申请人标识" min-width="170">
          <template #default="{ row }">
            <div class="id-pair">
              <span>成员：{{ row.requesterMemberId }}</span>
              <span>用户：{{ row.requesterUserId }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="requestReason" label="申请原因" min-width="220" show-overflow-tooltip />
        <el-table-column label="状态" width="105">
          <template #default="{ row }"><StatusTag :status="row.requestStatus" /></template>
        </el-table-column>
        <el-table-column prop="reviewResult" label="审核结果" width="105" />
        <el-table-column prop="reviewComment" label="审核意见" min-width="180" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="审核时间" width="180">
          <template #default="{ row }">{{ formatTime(row.reviewedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <div v-if="row.requestStatus === 'PENDING'" class="table-actions">
              <el-button size="small" type="danger" @click="openAudit(row.requestId, 'approve')">通过解散</el-button>
              <el-button size="small" @click="openAudit(row.requestId, 'reject')">驳回</el-button>
            </div>
            <div v-else-if="row.requestStatus === 'APPROVED'" class="table-actions">
              <el-tag v-if="restoredFamilyIds.has(row.familyId)" type="success">本次会话已恢复</el-tag>
              <el-button v-else size="small" type="primary" @click="openRestore(row.familyId)">恢复家庭</el-button>
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
      :key="auditKey"
      v-model="auditVisible"
      :title="auditTitle"
      :action="auditAction"
      :submitting="submitting"
      :close-on-submit="false"
      @submit="submitAudit"
    />
    <ConfirmDialog
      v-model="approveConfirmVisible"
      title="确认解散家庭"
      message="批准后家庭将变为已解散，停止公开展示且不可搜索。该操作不会物理删除家庭数据。"
      :submitting="submitting"
      :close-on-confirm="false"
      danger
      @confirm="confirmApprove"
    />
    <ConfirmDialog
      v-model="restoreConfirmVisible"
      title="确认恢复家庭"
      message="恢复后家庭状态变为 NORMAL、公开状态变为 PRIVATE，并默认恢复为可搜索。历史邀请和申请不会重新激活。"
      :submitting="submitting"
      :close-on-confirm="false"
      danger
      @confirm="confirmRestore"
    />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, ref } from 'vue'

import { restoreFamily } from '@/api/adminFamilies'
import { getApiErrorMessage } from '@/api/client'
import {
  approveDissolutionRequest,
  listDissolutionRequests,
  rejectDissolutionRequest
} from '@/api/dissolutions'
import AuditDialog from '@/components/AuditDialog.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { DissolutionRequest } from '@/types/api'

const status = ref('')
const familyId = ref<number | undefined>()
const items = ref<DissolutionRequest[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const auditVisible = ref(false)
const approveConfirmVisible = ref(false)
const restoreConfirmVisible = ref(false)
const auditAction = ref<'approve' | 'reject'>('approve')
const auditTitle = ref('家庭解散审核')
const auditKey = ref(0)
const selectedRequestId = ref<number | null>(null)
const selectedFamilyId = ref<number | null>(null)
const pendingReviewComment = ref<string | undefined>()
const restoredFamilyIds = ref(new Set<number>())
const submitting = ref(false)

onMounted(loadRequests)

async function loadRequests() {
  loading.value = true
  loadError.value = ''
  try {
    const result = await listDissolutionRequests({
      status: status.value || undefined,
      familyId: familyId.value || undefined,
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
  void loadRequests()
}

function resetSearch() {
  status.value = ''
  familyId.value = undefined
  page.value = 1
  void loadRequests()
}

function changePage(value: number) {
  page.value = value
  void loadRequests()
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  void loadRequests()
}

function openAudit(requestId: number, action: 'approve' | 'reject') {
  operationError.value = ''
  selectedRequestId.value = requestId
  auditAction.value = action
  auditTitle.value = `家庭解散申请 ${requestId}`
  auditKey.value += 1
  auditVisible.value = true
}

function openRestore(targetFamilyId: number) {
  operationError.value = ''
  selectedFamilyId.value = targetFamilyId
  restoreConfirmVisible.value = true
}

async function submitAudit(reviewComment: string) {
  if (!selectedRequestId.value || submitting.value) return
  const comment = reviewComment.trim() || undefined
  if (auditAction.value === 'approve') {
    pendingReviewComment.value = comment
    auditVisible.value = false
    approveConfirmVisible.value = true
    return
  }

  submitting.value = true
  operationError.value = ''
  try {
    await rejectDissolutionRequest(selectedRequestId.value, { reviewComment: comment })
    auditVisible.value = false
    ElMessage.success('家庭解散申请已驳回')
    await loadRequests()
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

async function confirmApprove() {
  if (!selectedRequestId.value || submitting.value) return
  submitting.value = true
  operationError.value = ''
  try {
    await approveDissolutionRequest(selectedRequestId.value, {
      reviewComment: pendingReviewComment.value
    })
    approveConfirmVisible.value = false
    ElMessage.success('家庭解散申请已通过')
    await loadRequests()
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

async function confirmRestore() {
  if (!selectedFamilyId.value || submitting.value) return
  submitting.value = true
  operationError.value = ''
  try {
    await restoreFamily(selectedFamilyId.value, { searchable: true })
    restoredFamilyIds.value = new Set(restoredFamilyIds.value).add(selectedFamilyId.value)
    restoreConfirmVisible.value = false
    ElMessage.success('家庭已恢复，公开状态已重置为私有')
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function formatTime(value?: string) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
}
</script>

<style scoped>
.pager {
  justify-content: flex-end;
  padding: 14px;
}

.id-pair {
  display: grid;
  gap: 4px;
  font-size: 13px;
}
</style>
