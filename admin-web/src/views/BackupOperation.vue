<template>
  <div class="page-stack">
    <PageHeader title="备份与存储" description="管理每日数据库备份的保留策略与容量保护。备份每天 03:20 执行，配置在下一次任务生效。" />
    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

    <div v-loading="loading" class="status-grid">
      <section class="surface status-card"><span>备份文件</span><strong>{{ overview?.status.backupCount ?? 0 }}</strong><em>当前保留数量</em></section>
      <section class="surface status-card"><span>备份占用</span><strong>{{ formatBytes(overview?.status.storageBytes ?? 0) }}</strong><em>受目录上限保护</em></section>
      <section class="surface status-card" :class="{ warning: (overview?.status.diskUsagePercent ?? 0) >= 75 }"><span>服务器磁盘</span><strong>{{ overview?.status.diskUsagePercent ?? 0 }}%</strong><em>达到阈值将跳过备份</em></section>
      <section class="surface status-card"><span>最近备份</span><strong class="latest">{{ formatTime(overview?.status.latestBackupAt) }}</strong><em>最近一次完成时间</em></section>
    </div>

    <el-card v-loading="loading" shadow="never">
      <template #header><div class="card-head"><strong>备份策略</strong><el-button :loading="loading" @click="load">刷新状态</el-button></div></template>
      <el-alert type="info" :closable="false" show-icon title="保护规则：达到备份目录上限或服务器磁盘阈值时，本次备份会跳过并写入巡检日志；旧备份会按保留天数自动清理。" />
      <el-form v-if="form" class="config-form" label-width="180px">
        <el-form-item label="自动数据库备份"><el-switch v-model="form.enabled" active-text="启用" inactive-text="暂停" /></el-form-item>
        <el-form-item label="保留天数"><el-input-number v-model="form.retentionDays" :min="1" :max="90" /><el-text type="info">保留 1 至 90 天，超过期限自动删除。</el-text></el-form-item>
        <el-form-item label="备份目录上限"><el-input-number v-model="form.maxStorageMb" :min="64" :max="102400" :step="64" /><el-text type="info">MB。达到上限后不创建新备份，避免目录无限增长。</el-text></el-form-item>
        <el-form-item label="服务器磁盘保护阈值"><el-input-number v-model="form.maxDiskUsagePercent" :min="60" :max="95" /><el-text type="info">%。达到阈值时跳过备份，优先保护线上服务。</el-text></el-form-item>
        <el-form-item><el-button type="primary" :loading="saving" @click="save">保存备份策略</el-button></el-form-item>
      </el-form>
      <el-descriptions v-if="overview" class="storage-details" :column="1" border>
        <el-descriptions-item label="备份目录">{{ overview.status.backupDirectory }}</el-descriptions-item>
        <el-descriptions-item label="当前目录占用">{{ formatBytes(overview.status.storageBytes) }} / {{ form?.maxStorageMb ?? 0 }} MB</el-descriptions-item>
        <el-descriptions-item label="最近配置更新">{{ formatTime(overview.config.updatedAt) }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import PageHeader from '@/components/PageHeader.vue'
import { getApiErrorMessage } from '@/api/client'
import { getBackupOperationOverview, updateBackupOperationConfig, type BackupOperationConfig, type BackupOperationOverview } from '@/api/backupOperation'

const overview = ref<BackupOperationOverview | null>(null)
const form = ref<Omit<BackupOperationConfig, 'updatedAt'> | null>(null)
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
function formatBytes(value: number) { return value < 1024 * 1024 ? `${Math.max(0, value / 1024).toFixed(1)} KB` : `${(value / 1024 / 1024).toFixed(1)} MB` }
function formatTime(value?: string | null) { return value ? value.replace('T', ' ').slice(0, 19) : '暂无成功备份' }
async function load() { loading.value = true; errorMessage.value = ''; try { const value = await getBackupOperationOverview(); overview.value = value; form.value = { enabled: value.config.enabled, retentionDays: value.config.retentionDays, maxStorageMb: value.config.maxStorageMb, maxDiskUsagePercent: value.config.maxDiskUsagePercent } } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { loading.value = false } }
async function save() { if (!form.value) return; saving.value = true; errorMessage.value = ''; try { const value = await updateBackupOperationConfig(form.value); if (overview.value) overview.value.config = value; ElMessage.success('备份策略已保存，将在下一次定时备份时生效') } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { saving.value = false } }
onMounted(load)
</script>

<style scoped>
.status-grid { display:grid; grid-template-columns:repeat(4, minmax(0, 1fr)); gap:16px; }.status-card { display:flex; min-height:132px; flex-direction:column; justify-content:space-between; padding:20px; }.status-card span,.status-card em { color:var(--color-text-secondary); font-size:13px; font-style:normal; }.status-card strong { color:var(--color-primary); font-size:32px; line-height:1.2; }.status-card strong.latest { font-size:18px; }.status-card.warning strong { color:var(--color-warning); }.card-head { display:flex; align-items:center; justify-content:space-between; }.config-form { max-width:860px; margin-top:22px; }.config-form :deep(.el-text) { margin-left:12px; }.storage-details { margin-top:22px; } @media (max-width:980px) { .status-grid { grid-template-columns:repeat(2, minmax(0, 1fr)); } } @media (max-width:640px) { .status-grid { grid-template-columns:1fr; } }
</style>
