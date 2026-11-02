<template>
  <div class="page-stack">
    <PageHeader title="后台管理员" description="维护管理员账号、角色和登录状态。" />
    <div class="toolbar"><el-button type="primary" @click="openCreate">新增管理员</el-button></div>
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
        <el-table-column label="操作" width="220" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openEdit(row)">编辑</el-button><el-button v-if="row.status === 'ACTIVE'" link type="warning" @click="changeStatus(row, 'DISABLED')">禁用</el-button><el-button v-else link type="success" @click="changeStatus(row, 'ACTIVE')">启用</el-button><el-button v-if="row.lockedUntil" link @click="changeStatus(row, 'UNLOCKED')">解锁</el-button></template></el-table-column>
      </el-table>
    </DataTable>
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑管理员' : '新增管理员'" width="520px">
      <el-form :model="form" label-width="90px"><el-form-item label="用户名" required><el-input v-model="form.username" :disabled="editing" /></el-form-item><el-form-item v-if="!editing" label="密码" required><el-input v-model="form.password" type="password" show-password /></el-form-item><el-form-item label="显示名"><el-input v-model="form.displayName" /></el-form-item><el-form-item label="角色"><el-select v-model="form.role"><el-option label="超级管理员" value="SUPER_ADMIN" /><el-option label="平台管理员" value="PLATFORM_ADMIN" /></el-select></el-form-item><el-form-item label="手机号"><el-input v-model="form.phone" /></el-form-item><el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item><el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item></el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getApiErrorMessage } from '@/api/client'
import { createAdmin, listAdmins, setAdminStatus, updateAdmin } from '@/api/management'
import { ElMessage, ElMessageBox } from 'element-plus'
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
const dialogVisible = ref(false)
const saving = ref(false)
const editing = ref<ManagedAdmin | null>(null)
const form = ref<any>({ username: '', password: '', displayName: '', role: 'PLATFORM_ADMIN', phone: '', email: '', remark: '' })

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
function openCreate() { editing.value = null; form.value = { username: '', password: '', displayName: '', role: 'PLATFORM_ADMIN', phone: '', email: '', remark: '' }; dialogVisible.value = true }
function openEdit(row: ManagedAdmin) { editing.value = row; form.value = { username: row.username, displayName: row.displayName || '', role: row.role, phone: row.phone || '', email: row.email || '', remark: '' }; dialogVisible.value = true }
async function save() { saving.value = true; try { if (editing.value) await updateAdmin(editing.value.id, form.value); else await createAdmin(form.value); ElMessage.success('保存成功'); dialogVisible.value = false; await load() } catch (e) { ElMessage.error(getApiErrorMessage(e)) } finally { saving.value = false } }
async function changeStatus(row: ManagedAdmin, next: 'ACTIVE' | 'DISABLED' | 'UNLOCKED') { try { await ElMessageBox.confirm(`确认${next === 'DISABLED' ? '禁用' : next === 'UNLOCKED' ? '解锁' : '启用'}管理员 ${row.username}？`, '操作确认'); await setAdminStatus(row.id, next); ElMessage.success('操作成功'); await load() } catch (e: any) { if (e !== 'cancel') ElMessage.error(getApiErrorMessage(e)) } }
function formatTime(value?: string | null) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '从未登录' }

onMounted(load)
</script>
