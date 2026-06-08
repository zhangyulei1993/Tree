<template>
  <div class="page-stack">
    <PageHeader title="家庭管理" description="家庭列表、公开状态、graph_version 与高风险入口展示。" />
    <SearchPanel>
      <el-form>
        <el-form-item label="关键词"><el-input v-model="keyword" placeholder="家族名称 / 姓氏 / 地区" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable placeholder="全部" style="width: 150px">
            <el-option label="正常" value="NORMAL" />
            <el-option label="解散待审" value="DISSOLUTION_PENDING" />
            <el-option label="已解散" value="DISSOLVED" />
          </el-select>
        </el-form-item>
        <el-form-item label="公开状态">
          <el-select v-model="publicStatus" clearable placeholder="全部" style="width: 150px">
            <el-option label="已公开" value="APPROVED" />
            <el-option label="待审核" value="PENDING" />
            <el-option label="私有" value="PRIVATE" />
          </el-select>
        </el-form-item>
        <el-button type="primary">查询</el-button>
        <el-button>重置</el-button>
      </el-form>
    </SearchPanel>
    <DataTable>
      <el-table :data="filtered" stripe>
        <el-table-column prop="id" label="家庭 ID" width="120" />
        <el-table-column prop="name" label="家庭名称" />
        <el-table-column prop="surname" label="姓氏" width="80" />
        <el-table-column prop="regionText" label="地区" min-width="150" />
        <el-table-column label="状态" width="120"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column label="公开状态" width="120"><template #default="{ row }"><StatusTag :status="row.publicDisplayStatus" /></template></el-table-column>
        <el-table-column prop="founder" label="当前创始人" width="120" />
        <el-table-column prop="memberCount" label="成员数" width="90" />
        <el-table-column prop="graphVersion" label="graph_version" width="130" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button size="small" @click="$router.push(`/admin/families/${row.id}`)">详情</el-button>
              <el-button size="small" @click="$router.push(`/admin/families/${row.id}/members`)">成员</el-button>
              <el-button size="small">编辑</el-button>
              <el-button v-if="auth.admin?.role !== 'PLATFORM_ADMIN'" size="small" type="danger" @click="dangerVisible = true">恢复/删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="pager" layout="total, prev, pager, next" :total="filtered.length" />
    </DataTable>
    <ConfirmDialog v-model="dangerVisible" title="高风险家庭操作" message="PLATFORM_ADMIN 不显示恢复 / 删除等高风险入口。本弹窗仅为静态原型。" danger />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import ConfirmDialog from '@/components/ConfirmDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import { families } from '@/mock/data'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const keyword = ref('')
const status = ref('')
const publicStatus = ref('')
const dangerVisible = ref(false)

const filtered = computed(() =>
  families.filter((item) =>
    (!keyword.value || JSON.stringify(item).includes(keyword.value)) &&
    (!status.value || item.status === status.value) &&
    (!publicStatus.value || item.publicDisplayStatus === publicStatus.value)
  )
)
</script>

<style scoped>
.pager {
  justify-content: flex-end;
  padding: 14px;
}
</style>
