<template>
  <div class="page-stack">
    <PageHeader title="家庭管理" description="真实家庭状态、公开状态、成员数与家谱版本。" />
    <SearchPanel>
      <el-form inline>
        <el-form-item label="关键词" class="field-wide">
          <el-input v-model="keyword" placeholder="ID / 名称 / 姓氏 / 地区" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" clearable style="width: 160px">
            <el-option label="正常" value="NORMAL" />
            <el-option label="解散待审" value="DISSOLUTION_PENDING" />
            <el-option label="恢复冷静期" value="DISSOLUTION_COOLDOWN" />
            <el-option label="已解散" value="DISSOLVED" />
          </el-select>
        </el-form-item>
        <el-button type="primary" @click="search">查询</el-button>
      </el-form>
    </SearchPanel>
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <DataTable>
      <el-table v-loading="loading" :data="items" stripe>
        <el-table-column prop="id" label="ID" width="90" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column prop="familyName" label="家庭名称" min-width="220" show-overflow-tooltip />
        <el-table-column prop="familySurname" label="姓氏" width="80" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column prop="regionText" label="地区" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="120" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column label="公开状态" width="120" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><StatusTag :status="row.publicDisplayStatus" /></template></el-table-column>
        <el-table-column prop="memberCount" label="成员" width="70" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column prop="graphVersion" label="家谱版本" width="92" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column label="操作" width="150" fixed="right" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">
            <TableActionGroup
              primary-text="家庭树"
              :actions="familyRowActions"
              @primary="openFamilyTree(row.id)"
              @command="(command) => handleFamilyCommand(row.id, command)"
            />
            <el-button link type="primary" @click="editFamily(row)">编辑</el-button>
            <el-button link :type="row.status === 'DISABLED' ? 'success' : 'warning'" @click="toggleFamily(row)">{{ row.status === 'DISABLED' ? '启用' : '停用' }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="pager" layout="total, prev, pager, next" :total="total" :page-size="20" @current-change="changePage" />
    </DataTable>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getApiErrorMessage } from '@/api/client'
import { listFamilies, setFamilyStatus, updateFamily } from '@/api/management'
import { ElMessage, ElMessageBox } from 'element-plus'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import TableActionGroup from '@/components/TableActionGroup.vue'
import type { ManagedFamily } from '@/types/api'

type FamilyRowCommand = 'detail' | 'members'

const router = useRouter()
const keyword = ref('')
const status = ref('')
const items = ref<ManagedFamily[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const error = ref('')
const familyRowActions = [
  { command: 'detail', label: '详情' },
  { command: 'members', label: '成员' }
] satisfies Array<{ command: FamilyRowCommand; label: string }>

async function load() {
  loading.value = true
  error.value = ''
  try {
    const r = await listFamilies({ keyword: keyword.value, status: status.value, page: page.value, pageSize: 20 })
    items.value = r.items
    total.value = r.total
  } catch (e) {
    error.value = getApiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openFamilyTree(id: number) {
  void router.push(`/admin/families/${id}/tree`)
}

function handleFamilyCommand(id: number, command: string) {
  if (command === 'detail') {
    void router.push(`/admin/families/${id}`)
    return
  }
  if (command === 'members') {
    void router.push(`/admin/families/${id}/members`)
  }
}

async function editFamily(row: ManagedFamily) {
  try {
    const familyName = await ElMessageBox.prompt('请输入家庭名称', '编辑家庭', { inputValue: row.familyName, inputPattern: /\S+/, inputErrorMessage: '家庭名称不能为空' })
    await updateFamily(row.id, { familyName: familyName.value })
    ElMessage.success('家庭资料已更新')
    await load()
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(getApiErrorMessage(e)) }
}

async function toggleFamily(row: ManagedFamily) {
  const next = row.status === 'DISABLED' ? 'NORMAL' : 'DISABLED'
  try { await ElMessageBox.confirm(`确认${next === 'DISABLED' ? '停用' : '启用'}家庭“${row.familyName}”？`, '操作确认'); await setFamilyStatus(row.id, next); ElMessage.success('操作成功'); await load() } catch (e: any) { if (e !== 'cancel') ElMessage.error(getApiErrorMessage(e)) }
}

function search() {
  page.value = 1
  void load()
}

function changePage(v: number) {
  page.value = v
  void load()
}

onMounted(load)
</script>
<style scoped>
.pager { justify-content: flex-end; padding: 18px; }
.field-wide { min-width: 320px; }
</style>
