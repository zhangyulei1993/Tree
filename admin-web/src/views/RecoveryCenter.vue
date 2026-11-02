<template>
  <div class="page-stack">
    <PageHeader title="变更恢复中心" description="展示近 30 日仍可恢复的成员与关系。恢复成员会恢复同次删除的关系，但不会重新绑定账号。" />
    <el-alert type="warning" :closable="false" show-icon title="恢复会再次更新家庭树版本，并写入操作日志；如原有位置已被新节点占用，系统会拒绝恢复。" />
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <DataTable>
      <el-table v-loading="loading" :data="items" stripe>
        <el-table-column label="类型" width="110" align="center" header-align="center"><template #default="{ row }"><el-tag :type="row.kind === 'MEMBER' ? 'warning' : 'info'">{{ row.kind === 'MEMBER' ? '成员' : '关系' }}</el-tag></template></el-table-column>
        <el-table-column prop="familyName" label="家庭" min-width="150" show-overflow-tooltip />
        <el-table-column prop="title" label="恢复对象" min-width="180" show-overflow-tooltip />
        <el-table-column prop="description" label="说明" min-width="200" show-overflow-tooltip />
        <el-table-column label="删除原因" min-width="180" show-overflow-tooltip><template #default="{ row }">{{ row.deleteReason || '未填写' }}</template></el-table-column>
        <el-table-column label="删除时间" width="180" align="center" header-align="center"><template #default="{ row }">{{ formatTime(row.deletedAt) }}</template></el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center" header-align="center"><template #default="{ row }"><el-button type="primary" plain size="small" @click="confirmRestore(row)">恢复</el-button></template></el-table-column>
      </el-table>
      <el-empty v-if="!loading && !items.length" description="近 30 日没有可恢复的结构变更" />
      <el-pagination class="pager" layout="total, prev, pager, next" :total="total" :page-size="20" @current-change="changePage" />
    </DataTable>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getApiErrorMessage } from '@/api/client'
import { listRecoveryItems, restoreRecoveryItem } from '@/api/management'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import type { RecoveryItem } from '@/types/api'

const items = ref<RecoveryItem[]>([]); const total = ref(0); const page = ref(1); const loading = ref(false); const error = ref('')
async function load() { loading.value = true; error.value = ''; try { const result = await listRecoveryItems({ page: page.value, pageSize: 20 }); items.value = result.items; total.value = result.total } catch (e) { error.value = getApiErrorMessage(e) } finally { loading.value = false } }
async function confirmRestore(item: RecoveryItem) { try { await ElMessageBox.confirm(`确定恢复“${item.title}”吗？此操作会更新家庭树。`, '确认恢复', { type: 'warning', confirmButtonText: '恢复', cancelButtonText: '取消' }); await restoreRecoveryItem(item); ElMessage.success('已恢复并记录审计日志'); await load() } catch (e) { if (e !== 'cancel' && e !== 'close') ElMessage.error(getApiErrorMessage(e)) } }
function changePage(value: number) { page.value = value; void load() }
function formatTime(value: string) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' }
onMounted(load)
</script>
<style scoped>.pager { justify-content:flex-end; padding:18px }</style>
