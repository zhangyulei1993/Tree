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
        <el-form-item label="家庭 ID">
          <el-input v-model="familyId" placeholder="可选" />
        </el-form-item>
        <el-button type="primary" @click="search">查询</el-button>
      </el-form>
    </SearchPanel>
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <DataTable>
      <el-table v-loading="loading" :data="items" stripe>
        <el-table-column prop="id" label="日志 ID" width="100" />
        <el-table-column prop="operatorType" label="操作者类型" min-width="120" show-overflow-tooltip />
        <el-table-column prop="operatorRole" label="角色" min-width="140">
          <template #default="{ row }">
            <RoleTag v-if="row.operatorRole" :role="row.operatorRole" />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="module" label="模块" min-width="160" show-overflow-tooltip />
        <el-table-column prop="action" label="动作" min-width="220" show-overflow-tooltip />
        <el-table-column prop="targetId" label="目标 ID" min-width="100" show-overflow-tooltip />
        <el-table-column prop="result" label="结果" min-width="90" />
        <el-table-column prop="createdAt" label="时间" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="90" fixed="right">
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
      <JsonViewer :value="selected || {}" />
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
import type { OperationLogRecord } from '@/types/api'

const keyword = ref('')
const result = ref('')
const familyId = ref('')
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
      familyId: familyId.value || undefined,
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

function changePage(value: number) {
  page.value = value
  void load()
}

function open(row: OperationLogRecord) {
  selected.value = row
  drawer.value = true
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
</style>
