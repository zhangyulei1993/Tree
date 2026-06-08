<template>
  <div class="page-stack">
    <PageHeader title="用户管理" description="用户账号列表、状态筛选与预创建账号原型。">
      <el-button type="primary" @click="preCreateVisible = true">预创建账号</el-button>
    </PageHeader>
    <SearchPanel>
      <el-form>
        <el-form-item label="关键词"><el-input v-model="keyword" placeholder="昵称 / 姓名 / 用户 ID" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable placeholder="全部" style="width: 140px">
            <el-option label="正常" value="ACTIVE" />
            <el-option label="待绑定" value="PENDING" />
            <el-option label="禁用" value="DISABLED" />
          </el-select>
        </el-form-item>
        <el-button type="primary">查询</el-button>
        <el-button>重置</el-button>
      </el-form>
    </SearchPanel>
    <DataTable>
      <el-table :data="filteredUsers" stripe>
        <el-table-column prop="id" label="用户 ID" width="120" />
        <el-table-column prop="maskedPhone" label="手机号" width="140" />
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column prop="realName" label="真实姓名" />
        <el-table-column prop="origin" label="来源" width="130" />
        <el-table-column prop="clientType" label="注册端" width="120" />
        <el-table-column label="手机号验证" width="110">
          <template #default="{ row }"><el-tag :type="row.phoneVerified ? 'success' : 'warning'">{{ row.phoneVerified ? '已验证' : '未验证' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><StatusTag :status="row.status" /></template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button size="small" @click="$router.push(`/admin/users/${row.id}`)">详情</el-button>
              <el-button size="small">编辑</el-button>
              <el-button size="small" type="danger" @click="openDanger('禁用用户')">禁用</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="pager" layout="total, prev, pager, next" :total="filteredUsers.length" />
    </DataTable>
    <el-dialog v-model="preCreateVisible" title="预创建账号" width="520px">
      <el-form label-width="96px">
        <el-form-item label="姓名"><el-input placeholder="虚构姓名" /></el-form-item>
        <el-form-item label="手机号"><el-input placeholder="138****0000" /></el-form-item>
        <el-form-item label="备注"><el-input type="textarea" :rows="3" placeholder="mock 备注" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="preCreateVisible = false">取消</el-button>
        <el-button type="primary" @click="preCreateVisible = false">创建 mock 账号</el-button>
      </template>
    </el-dialog>
    <ConfirmDialog v-model="dangerVisible" title="危险操作确认" :message="dangerMessage" danger />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import ConfirmDialog from '@/components/ConfirmDialog.vue'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import { users } from '@/mock/data'

const keyword = ref('')
const status = ref('')
const preCreateVisible = ref(false)
const dangerVisible = ref(false)
const dangerMessage = ref('')

const filteredUsers = computed(() =>
  users.filter((item) => (!status.value || item.status === status.value) && (!keyword.value || JSON.stringify(item).includes(keyword.value)))
)

function openDanger(action: string) {
  dangerMessage.value = `${action} 是高风险操作，静态原型仅展示二次确认。`
  dangerVisible.value = true
}
</script>

<style scoped>
.pager {
  justify-content: flex-end;
  padding: 14px;
}
</style>
