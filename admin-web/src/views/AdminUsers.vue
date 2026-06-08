<template>
  <div class="page-stack">
    <PageHeader title="后台管理员管理" description="ROOT_ADMIN / SUPER_ADMIN 可访问，PLATFORM_ADMIN 菜单不可见。">
      <el-button type="primary">新增管理员</el-button>
    </PageHeader>
    <SearchPanel>
      <el-form>
        <el-form-item label="角色"><el-select v-model="role" clearable placeholder="全部" style="width: 160px"><el-option label="ROOT_ADMIN" value="ROOT_ADMIN" /><el-option label="SUPER_ADMIN" value="SUPER_ADMIN" /><el-option label="PLATFORM_ADMIN" value="PLATFORM_ADMIN" /></el-select></el-form-item>
        <el-form-item label="状态"><el-select v-model="status" clearable placeholder="全部" style="width: 140px"><el-option label="正常" value="ACTIVE" /><el-option label="禁用" value="DISABLED" /></el-select></el-form-item>
        <el-button type="primary">查询</el-button>
      </el-form>
    </SearchPanel>
    <DataTable>
      <el-table :data="filtered" stripe>
        <el-table-column prop="id" label="管理员 ID" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="displayName" label="显示名" />
        <el-table-column label="角色"><template #default="{ row }"><RoleTag :role="row.role" /></template></el-table-column>
        <el-table-column label="状态"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column prop="lastLoginAt" label="最后登录" />
        <el-table-column label="操作" width="260">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button size="small">编辑</el-button>
              <el-button size="small" type="warning" :disabled="row.role === 'ROOT_ADMIN'">锁定</el-button>
              <el-button size="small" type="danger" :disabled="row.role === 'ROOT_ADMIN'">禁用</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RoleTag from '@/components/RoleTag.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import { adminUsers } from '@/mock/data'

const role = ref('')
const status = ref('')
const filtered = computed(() => adminUsers.filter((item) => (!role.value || item.role === role.value) && (!status.value || item.status === status.value)))
</script>
