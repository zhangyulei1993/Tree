<template>
  <div class="page-stack">
    <PageHeader title="后台管理员" description="管理员账号与登录状态只读视图；账号新增与状态变更尚未开放。" />
    <SearchPanel>
      <el-form inline>
        <el-form-item label="角色">
          <el-select v-model="role" clearable placeholder="全部" style="width: 160px">
            <el-option label="根管理员" value="ROOT_ADMIN" />
            <el-option label="超级管理员" value="SUPER_ADMIN" />
            <el-option label="平台管理员" value="PLATFORM_ADMIN" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable placeholder="全部" style="width: 130px">
            <el-option label="正常" value="ACTIVE" />
            <el-option label="禁用" value="DISABLED" />
          </el-select>
        </el-form-item>
        <el-button type="primary" @click="search">查询</el-button>
        <el-button @click="resetSearch">重置</el-button>
      </el-form>
    </SearchPanel>
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <DataTable>
      <el-table v-loading="loading" :data="items" stripe empty-text="暂无符合条件的管理员账号">
        <el-table-column prop="id" label="ID" width="90" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column prop="username" label="用户名" min-width="180" />
        <el-table-column label="显示名" min-width="160"><template #default="{ row }">{{ row.displayName || '未设置' }}</template></el-table-column>
        <el-table-column label="角色" width="132" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><RoleTag :role="row.role" /></template></el-table-column>
        <el-table-column label="状态" width="96" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column label="最后登录" width="168" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }">{{ formatTime(row.lastLoginAt) }}</template></el-table-column>
      </el-table>
    </DataTable>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getApiErrorMessage } from '@/api/client'
import { listAdmins } from '@/api/management'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RoleTag from '@/components/RoleTag.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { AdminRole, ManagedAdmin } from '@/types/api'

const role = ref<AdminRole | ''>('')
const status = ref('')
const items = ref<ManagedAdmin[]>([])
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    items.value = (await listAdmins({ role: role.value, status: status.value, page: 1, pageSize: 100 })).items
  } catch (requestError) {
    error.value = getApiErrorMessage(requestError)
  } finally {
    loading.value = false
  }
}

function search() { void load() }
function resetSearch() { role.value = ''; status.value = ''; void load() }
function formatTime(value?: string | null) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '从未登录' }

onMounted(load)
</script>
