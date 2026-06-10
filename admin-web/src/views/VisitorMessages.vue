<template>
  <div class="page-stack">
    <PageHeader title="游客留言审核" description="审核公开家庭收到的游客留言，联系方式仅供后台处理使用。" />

    <SearchPanel>
      <el-form>
        <el-form-item label="状态">
          <el-select v-model="status" clearable placeholder="全部" style="width: 150px">
            <el-option label="待审核" value="PENDING" />
            <el-option label="已通过" value="APPROVED" />
            <el-option label="已拒绝" value="REJECTED" />
            <el-option label="已删除" value="DELETED" />
          </el-select>
        </el-form-item>
        <el-form-item label="家庭 ID">
          <el-input-number v-model="familyId" :min="1" :controls="false" placeholder="全部" />
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="search">查询</el-button>
        <el-button :disabled="loading" @click="resetSearch">重置</el-button>
      </el-form>
    </SearchPanel>

    <el-alert v-if="loadError" :title="`留言列表加载失败：${loadError}`" type="error" show-icon :closable="false" />
    <el-alert v-if="operationError" :title="`操作失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <DataTable>
      <el-table v-loading="loading" :data="items" stripe empty-text="暂无游客留言">
        <el-table-column prop="messageId" label="留言 ID" width="105" />
        <el-table-column prop="familyId" label="家庭 ID" width="100" />
        <el-table-column prop="familyName" label="家庭" min-width="140" />
        <el-table-column prop="visitorName" label="留言人" width="120" />
        <el-table-column prop="messageContent" label="留言内容" min-width="240" show-overflow-tooltip />
        <el-table-column label="联系方式" min-width="190">
          <template #default="{ row }">
            <div class="contact">
              <span>电话：{{ row.visitorPhone || '-' }}</span>
              <span>微信：{{ row.visitorWechat || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="105">
          <template #default="{ row }"><StatusTag :status="row.status" /></template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button
                v-if="row.status === 'PENDING' || row.status === 'REJECTED'"
                size="small"
                type="primary"
                @click="openAudit(row.messageId, 'approve')"
              >
                通过
              </el-button>
              <el-button
                v-if="row.status !== 'DELETED'"
                size="small"
                type="danger"
                @click="openAudit(row.messageId, 'reject')"
              >
                拒绝
              </el-button>
              <el-button
                v-if="row.status !== 'DELETED'"
                size="small"
                type="danger"
                plain
                @click="openDelete(row.messageId)"
              >
                删除
              </el-button>
            </div>
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
    <ConfirmDialog
      v-model="deleteVisible"
      title="删除游客留言"
      message="删除后该留言不会再出现在公开列表中。"
      reason-label="填写删除原因（可选）"
      :submitting="submitting"
      :close-on-confirm="false"
      danger
      @confirm="submitDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, ref } from 'vue'

import { getApiErrorMessage } from '@/api/client'
import {
  approveVisitorMessage,
  deleteVisitorMessage,
  listVisitorMessages,
  rejectVisitorMessage
} from '@/api/visitorMessages'
import AuditDialog from '@/components/AuditDialog.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { VisitorMessage } from '@/types/api'

const status = ref('')
const familyId = ref<number | undefined>()
const items = ref<VisitorMessage[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const submitting = ref(false)
const auditVisible = ref(false)
const deleteVisible = ref(false)
const auditAction = ref<'approve' | 'reject'>('approve')
const auditTitle = ref('游客留言审核')
const selectedMessageId = ref<number | null>(null)

onMounted(loadMessages)

async function loadMessages() {
  loading.value = true
  loadError.value = ''
  try {
    const result = await listVisitorMessages({
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
  void loadMessages()
}

function resetSearch() {
  status.value = ''
  familyId.value = undefined
  page.value = 1
  void loadMessages()
}

function changePage(value: number) {
  page.value = value
  void loadMessages()
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  void loadMessages()
}

function openAudit(messageId: number, action: 'approve' | 'reject') {
  operationError.value = ''
  selectedMessageId.value = messageId
  auditAction.value = action
  auditTitle.value = `游客留言 ${messageId}`
  auditVisible.value = true
}

function openDelete(messageId: number) {
  operationError.value = ''
  selectedMessageId.value = messageId
  deleteVisible.value = true
}

async function submitAudit(reviewComment: string) {
  if (!selectedMessageId.value || submitting.value) return
  submitting.value = true
  operationError.value = ''
  try {
    const input = { reviewComment: reviewComment.trim() || undefined }
    if (auditAction.value === 'approve') {
      await approveVisitorMessage(selectedMessageId.value, input)
    } else {
      await rejectVisitorMessage(selectedMessageId.value, input)
    }
    auditVisible.value = false
    ElMessage.success(auditAction.value === 'approve' ? '留言已通过' : '留言已驳回')
    await loadMessages()
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

async function submitDelete(deleteReason: string) {
  if (!selectedMessageId.value || submitting.value) return
  submitting.value = true
  operationError.value = ''
  try {
    await deleteVisitorMessage(selectedMessageId.value, {
      deleteReason: deleteReason.trim() || undefined
    })
    deleteVisible.value = false
    ElMessage.success('留言已删除')
    await loadMessages()
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

.contact {
  display: grid;
  gap: 4px;
  font-size: 13px;
}
</style>
