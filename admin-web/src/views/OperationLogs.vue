<template>
  <div class="page-stack">
    <PageHeader title="操作日志" description="真实审计日志，详情字段已由后端递归脱敏。" />
    <SearchPanel>
      <el-form inline>
        <el-form-item label="关键词" class="field-wide">
          <el-input v-model="keyword" placeholder="模块 / 动作 / 目标类型" />
        </el-form-item>
        <el-form-item label="结果">
          <el-select v-model="result" clearable style="width: 140px">
            <el-option label="成功" value="SUCCESS" />
            <el-option label="失败" value="FAILED" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作者">
          <el-select v-model="operatorType" clearable placeholder="全部" style="width: 120px">
            <el-option label="后台管理员" value="ADMIN" />
            <el-option label="平台用户" value="USER" />
            <el-option label="系统任务" value="SYSTEM" />
          </el-select>
        </el-form-item>
        <el-form-item label="管理员角色">
          <el-select v-model="role" clearable placeholder="全部" style="width: 140px">
            <el-option label="根管理员" value="ROOT_ADMIN" />
            <el-option label="超级管理员" value="SUPER_ADMIN" />
            <el-option label="平台管理员" value="PLATFORM_ADMIN" />
          </el-select>
        </el-form-item>
        <el-form-item label="家庭 ID">
          <el-input v-model="familyId" placeholder="可选" />
        </el-form-item>
        <el-form-item label="操作日期">
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" />
        </el-form-item>
        <el-button type="primary" @click="search">查询</el-button>
        <el-button @click="resetSearch">重置</el-button>
      </el-form>
    </SearchPanel>
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <DataTable>
      <el-table v-loading="loading" :data="items" stripe>
        <el-table-column prop="id" label="日志 ID" width="90" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column label="操作者" min-width="160">
          <template #default="{ row }">{{ operatorText(row.operatorType) }}</template>
        </el-table-column>
        <el-table-column prop="operatorRole" label="角色" width="120" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">
            <RoleTag v-if="row.operatorRole" :role="row.operatorRole" />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="模块" width="140" show-overflow-tooltip><template #default="{ row }">{{ codeText(row.module) }}</template></el-table-column>
        <el-table-column label="动作" min-width="220" show-overflow-tooltip><template #default="{ row }">{{ actionText(row.action) }}</template></el-table-column>
        <el-table-column prop="targetId" label="目标 ID" width="96" show-overflow-tooltip align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column label="结果" width="90" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><StatusTag :status="row.result" /></template></el-table-column>
        <el-table-column label="时间" width="160" show-overflow-tooltip align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }">{{ formatTime(row.createdAt) }}</template></el-table-column>
        <el-table-column label="操作" width="90" fixed="right" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">
            <el-button size="small" @click="open(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        class="pager"
        layout="total, prev, pager, next"
        :total="total"
        :page-size="20"
        @current-change="changePage"
      />
    </DataTable>
    <DetailDrawer v-model="drawer" title="操作日志详情">
      <template v-if="selected">
        <section class="log-summary">
          <div class="summary-heading"><div><span>审计事件 #{{ selected.id }}</span><h3>{{ actionText(selected.action) }}</h3></div><StatusTag :status="selected.result" /></div>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="发生时间">{{ formatTime(selected.createdAt) }}</el-descriptions-item>
            <el-descriptions-item label="模块">{{ codeText(selected.module) }}</el-descriptions-item>
            <el-descriptions-item label="操作者">{{ operatorText(selected.operatorType) }}</el-descriptions-item>
            <el-descriptions-item label="管理员角色">{{ selected.operatorRole || '—' }}</el-descriptions-item>
            <el-descriptions-item label="家庭 ID">{{ selected.familyId || '—' }}</el-descriptions-item>
            <el-descriptions-item label="目标">{{ targetText(selected) }}</el-descriptions-item>
            <el-descriptions-item label="成员 ID">{{ selected.memberId || '—' }}</el-descriptions-item>
            <el-descriptions-item label="用户 ID">{{ selected.userId || '—' }}</el-descriptions-item>
          </el-descriptions>
          <el-alert v-if="selected.errorMessage" class="error-message" type="error" show-icon :closable="false" :title="selected.errorMessage" />
        </section>
        <el-collapse class="raw-detail">
          <el-collapse-item title="查看原始审计数据" name="raw"><JsonViewer :value="selected.detailJson || {}" /></el-collapse-item>
        </el-collapse>
      </template>
    </DetailDrawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getApiErrorMessage } from '@/api/client'
import { listOperationLogs } from '@/api/management'
import DataTable from '@/components/DataTable.vue'
import DetailDrawer from '@/components/DetailDrawer.vue'
import JsonViewer from '@/components/JsonViewer.vue'
import PageHeader from '@/components/PageHeader.vue'
import RoleTag from '@/components/RoleTag.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { AdminRole, OperationLogRecord } from '@/types/api'

const keyword = ref('')
const result = ref('')
const operatorType = ref('')
const role = ref<AdminRole | ''>('')
const familyId = ref('')
const dateRange = ref<string[]>([])
const items = ref<OperationLogRecord[]>([])
const selected = ref<OperationLogRecord | null>(null)
const drawer = ref(false)
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const response = await listOperationLogs({
      keyword: keyword.value,
      result: result.value,
      operatorType: operatorType.value,
      role: role.value,
      familyId: familyId.value || undefined,
      startDate: dateRange.value[0],
      endDate: dateRange.value[1],
      page: page.value,
      pageSize: 20
    })
    items.value = response.items
    total.value = response.total
  } catch (loadError) {
    error.value = getApiErrorMessage(loadError)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  void load()
}

function resetSearch() {
  keyword.value = ''
  result.value = ''
  operatorType.value = ''
  role.value = ''
  familyId.value = ''
  dateRange.value = []
  page.value = 1
  void load()
}

function changePage(value: number) {
  page.value = value
  void load()
}

function open(row: OperationLogRecord) {
  selected.value = row
  drawer.value = true
}

function operatorText(type: string) {
  return ({ ADMIN: '后台管理员', USER: '平台用户', SYSTEM: '系统任务' } as Record<string, string>)[type] || type || '未记录'
}

function codeText(value: string) {
  return value ? value.replace(/_/g, ' ') : '未记录'
}

function actionText(value: string) {
  const labels: Record<string, string> = {
    APPROVE_DISSOLUTION_REQUEST: '通过家庭解散申请',
    REJECT_DISSOLUTION_REQUEST: '驳回家庭解散申请',
    FINALIZE_DISSOLUTION: '完成家庭解散',
    RESTORE_FAMILY: '恢复家庭',
    APPROVE_FOUNDER_TRANSFER_REQUEST: '通过创始人转让',
    REJECT_FOUNDER_TRANSFER_REQUEST: '驳回创始人转让',
    APPROVE_PUBLIC_APPLICATION: '通过优质家庭申请',
    REJECT_PUBLIC_APPLICATION: '驳回优质家庭申请',
    TAKE_DOWN_PUBLIC_FAMILY: '移出优质家庭列表',
    CREATE_MEMBER: '创建成员节点',
    UPDATE_MEMBER: '更新成员资料',
    DELETE_MEMBER: '删除成员节点',
    RESTORE_MEMBER: '恢复成员节点',
    CREATE_RELATIONSHIP: '创建成员关系',
    UPDATE_RELATIONSHIP: '调整成员关系',
    DELETE_RELATIONSHIP: '删除成员关系',
    RESTORE_RELATIONSHIP: '恢复成员关系',
    UPDATE_SHARE_CONFIG: '更新分享配置',
    UPDATE_TOOL_CONFIG: '更新常用工具配置',
    CREATE_SHOWCASE_FAMILY: '创建帝王世家',
    UPDATE_SHOWCASE_FAMILY: '编辑帝王世家',
    MARK_SHOWCASE_UNPUBLISHED_AFTER_CHANGE: '修改后退出展示',
    PUBLISH_SHOWCASE_FAMILY: '发布帝王世家',
    UNPUBLISH_SHOWCASE_FAMILY: '下架帝王世家',
    DELETE_SHOWCASE_FAMILY: '删除帝王世家',
    ADD_SHOWCASE_RELATIVE: '在展示树添加亲属'
  }
  return labels[value] || codeText(value)
}

function targetText(row: OperationLogRecord) {
  if (!row.targetType && !row.targetId) return '—'
  return `${codeText(row.targetType || '目标')} #${row.targetId || '—'}`
}

function formatTime(value: string) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
}

onMounted(load)
</script>

<style scoped>
.pager {
  justify-content: flex-end;
  padding: 18px;
}

.field-wide {
  min-width: 280px;
}

.log-summary { display:grid;gap:16px }.summary-heading { display:flex;align-items:flex-start;justify-content:space-between;gap:16px }.summary-heading span { color:var(--color-text-secondary);font-size:12px;font-weight:700 }.summary-heading h3 { margin:4px 0 0;color:var(--color-primary);font-size:20px }.error-message { margin-top:4px }.raw-detail { margin-top:18px }
</style>
