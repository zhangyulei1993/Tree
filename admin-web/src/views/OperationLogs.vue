<template>
  <div class="page-stack">
    <PageHeader title="操作日志" description="敏感字段过滤后的 mock 审计日志。" />
    <SearchPanel>
      <el-form>
        <el-form-item label="模块"><el-input v-model="keyword" placeholder="模块 / 动作 / 目标 ID" /></el-form-item>
        <el-form-item label="结果"><el-select v-model="result" clearable placeholder="全部" style="width: 140px"><el-option label="成功" value="SUCCESS" /><el-option label="失败" value="FAILED" /></el-select></el-form-item>
        <el-button type="primary">查询</el-button>
      </el-form>
    </SearchPanel>
    <DataTable>
      <el-table :data="filtered" stripe>
        <el-table-column prop="id" label="日志 ID" width="120" />
        <el-table-column prop="actorType" label="操作者类型" width="120" />
        <el-table-column prop="actor" label="操作者" />
        <el-table-column prop="role" label="角色"><template #default="{ row }"><RoleTag :role="row.role" /></template></el-table-column>
        <el-table-column prop="module" label="模块" />
        <el-table-column prop="action" label="动作" min-width="220" />
        <el-table-column prop="targetId" label="目标 ID" />
        <el-table-column prop="result" label="结果" />
        <el-table-column prop="createdAt" label="时间" width="160" />
        <el-table-column label="操作" width="100"><template #default="{ row }"><el-button size="small" @click="openDetail(row)">详情</el-button></template></el-table-column>
      </el-table>
    </DataTable>
    <DetailDrawer v-model="drawerVisible" title="操作日志详情">
      <JsonViewer :value="selected?.detail || {}" />
    </DetailDrawer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import DataTable from '@/components/DataTable.vue'
import DetailDrawer from '@/components/DetailDrawer.vue'
import JsonViewer from '@/components/JsonViewer.vue'
import PageHeader from '@/components/PageHeader.vue'
import RoleTag from '@/components/RoleTag.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import { operationLogs, type OperationLogRow } from '@/mock/data'

const keyword = ref('')
const result = ref('')
const drawerVisible = ref(false)
const selected = ref<OperationLogRow>()

const filtered = computed(() =>
  operationLogs.filter((item) => (!result.value || item.result === result.value) && (!keyword.value || JSON.stringify(item).includes(keyword.value)))
)

function openDetail(row: OperationLogRow) {
  selected.value = row
  drawerVisible.value = true
}
</script>
