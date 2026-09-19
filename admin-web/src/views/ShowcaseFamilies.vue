<template>
  <div class="page-stack">
    <PageHeader
      title="帝王世家配置"
      description="仅维护平台运营展示内容，不关联用户创建的家庭、成员或公开申请。发布后才会进入小程序帝王世家入口。"
    >
      <el-button type="primary" @click="openCreate">新建帝王世家</el-button>
    </PageHeader>

    <div class="notice-bar">
      <strong>运营展示内容</strong>
      <span>仅填写已获得授权的示例文案和图片，不录入真实用户家庭资料。</span>
    </div>
    <section class="rotation-panel">
      <div>
        <strong>随机置顶</strong>
        <span>每个周期从已发布帝王世家中随机选一套置顶，期间保持不变。</span>
      </div>
      <div class="rotation-controls">
        <span>置顶</span>
        <el-input-number v-model="rotationDays" :min="1" :max="30" :disabled="settingsLoading" controls-position="right" />
        <span>天</span>
        <el-button size="small" type="primary" :loading="settingsSaving" @click="saveDisplaySettings">保存</el-button>
      </div>
      <span class="rotation-status">{{ rotationStatus }}</span>
    </section>
    <el-alert v-if="loadError" :title="`加载失败：${loadError}`" type="error" show-icon :closable="false" />
    <el-alert v-if="operationError" :title="`操作失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <DataTable v-loading="loading" class="showcase-card">
      <el-table :data="families" row-key="id" class="showcase-table">
        <el-table-column label="展示内容" min-width="250">
          <template #default="{ row }">
            <div class="showcase-name">
              <el-image v-if="row.coverImageUrl" :src="row.coverImageUrl" fit="cover" class="cover-preview">
                <template #error><div class="image-fallback">封面</div></template>
              </el-image>
              <TableCellStack :title="row.displayName" :description="showcaseSubtitle(row)" />
            </div>
          </template>
        </el-table-column>
        <el-table-column
          label="规模"
          width="142"
          align="center"
          header-align="center"
          class-name="showcase-center-cell"
          label-class-name="showcase-center-header"
        >
          <template #default="{ row }">
            <div class="table-status-stack table-status-stack--center">
              <span class="table-status-stack__meta">{{ row.memberCount }} 人 · {{ row.relationshipCount }} 条关系</span>
              <span v-if="row.treeStatus !== 'READY'" class="table-status-stack__hint">关系待完善</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          label="状态"
          width="96"
          align="center"
          header-align="center"
          class-name="showcase-center-cell"
          label-class-name="showcase-center-header"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 'PUBLISHED' ? 'success' : 'info'" effect="plain">
              {{ row.status === 'PUBLISHED' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="sortOrder"
          label="排序"
          width="72"
          align="center"
          header-align="center"
          class-name="showcase-center-cell"
          label-class-name="showcase-center-header"
        />
        <el-table-column
          label="最近更新"
          width="156"
          align="center"
          header-align="center"
          class-name="showcase-center-cell"
          label-class-name="showcase-center-header"
        >
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="156"
          fixed="right"
          align="center"
          header-align="center"
          class-name="showcase-center-cell"
          label-class-name="showcase-center-header"
        >
          <template #default="{ row }">
            <TableActionGroup
              :primary-text="row.treeStatus === 'READY' ? '查看家谱' : '继续编排'"
              :loading="Boolean(saving[row.id])"
              :actions="actionItems(row)"
              @primary="$router.push(`/admin/showcase-families/${row.id}/tree`)"
              @command="handleCommand($event, row)"
            />
          </template>
        </el-table-column>
      </el-table>
    </DataTable>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑帝王世家' : '新建帝王世家'" width="680px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="88px">
        <el-form-item label="展示名称" prop="displayName">
          <el-input v-model="form.displayName" maxlength="120" show-word-limit placeholder="例如 张氏家谱示例" />
        </el-form-item>
        <el-form-item label="地区说明" prop="regionText">
          <el-input v-model="form.regionText" maxlength="120" show-word-limit placeholder="例如 山东省济南市" />
        </el-form-item>
        <el-form-item label="展示简介" prop="summary">
          <el-input v-model="form.summary" type="textarea" :rows="4" maxlength="1000" show-word-limit placeholder="说明这个示例家谱展示的内容与特点" />
        </el-form-item>
        <el-form-item label="封面图片" prop="coverImageUrl">
          <el-input v-model="form.coverImageUrl" maxlength="500" placeholder="可选：图片地址" />
        </el-form-item>
        <el-form-item label="补充配图" prop="treeImageUrl">
          <el-input v-model="form.treeImageUrl" maxlength="500" placeholder="可选：用于展示介绍的补充图片，不替代可编辑家谱树" />
        </el-form-item>
        <el-form-item label="排序" prop="sortOrder">
          <el-input-number v-model="form.sortOrder" :min="0" :max="9999" controls-position="right" />
          <span class="field-tip">数值越小越靠前。</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">{{ editing ? '保存修改' : '创建草稿' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus'

import { getApiErrorMessage } from '@/api/client'
import { createShowcaseFamily, deleteShowcaseFamily, getShowcaseFamilyDisplaySettings, listShowcaseFamilies, setShowcaseFamilyPublished, updateShowcaseFamily, updateShowcaseFamilyDisplaySettings, type ShowcaseFamilyInput } from '@/api/showcaseFamily'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import TableCellStack from '@/components/TableCellStack.vue'
import TableActionGroup from '@/components/TableActionGroup.vue'
import type { ShowcaseFamily, ShowcaseFamilyDisplaySettings } from '@/types/api'

const families = ref<ShowcaseFamily[]>([])
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const loadError = ref('')
const operationError = ref('')
const dialogVisible = ref(false)
const editing = ref<ShowcaseFamily | null>(null)
const saving = reactive<Record<number, boolean>>({})
const displaySettings = ref<ShowcaseFamilyDisplaySettings | null>(null)
const settingsLoading = ref(false)
const settingsSaving = ref(false)
const rotationDays = ref(7)
const formRef = ref<FormInstance>()
const form = reactive<ShowcaseFamilyInput>({ displayName: '', regionText: '', summary: '', coverImageUrl: '', treeImageUrl: '', sortOrder: 0 })
const rotationStatus = computed(() => {
  if (!displaySettings.value?.pinnedShowcaseFamilyId) return '保存后将在下一次展示读取时随机置顶'
  const family = families.value.find((item) => item.id === displaySettings.value?.pinnedShowcaseFamilyId)
  return `${family?.displayName || '当前帝王世家'} · 至 ${formatTime(displaySettings.value.pinnedUntil)}`
})

const rules: FormRules = {
  displayName: [{ required: true, message: '请输入展示名称', trigger: 'blur' }],
  summary: [{ required: true, message: '请输入展示简介', trigger: 'blur' }]
}

function formatTime(value?: string | null) {
  return value ? value.replace('T', ' ').slice(0, 19) : '-'
}

function showcaseSubtitle(item: ShowcaseFamily) {
  const region = item.regionText || '未填写地区'
  return item.summary ? `${region} · ${item.summary}` : region
}

function resetForm() {
  Object.assign(form, { displayName: '', regionText: '', summary: '', coverImageUrl: '', treeImageUrl: '', sortOrder: 0 })
}

function openCreate() {
  editing.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(item: ShowcaseFamily) {
  editing.value = item
  Object.assign(form, {
    displayName: item.displayName,
    regionText: item.regionText,
    summary: item.summary,
    coverImageUrl: item.coverImageUrl,
    treeImageUrl: item.treeImageUrl,
    sortOrder: item.sortOrder
  })
  dialogVisible.value = true
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    families.value = await listShowcaseFamilies()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function loadDisplaySettings() {
  settingsLoading.value = true
  try {
    displaySettings.value = await getShowcaseFamilyDisplaySettings()
    rotationDays.value = displaySettings.value.rotationDays
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    settingsLoading.value = false
  }
}

async function saveDisplaySettings() {
  settingsSaving.value = true
  operationError.value = ''
  try {
    displaySettings.value = await updateShowcaseFamilyDisplaySettings(rotationDays.value)
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    settingsSaving.value = false
  }
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  operationError.value = ''
  try {
    if (editing.value) {
      const updated = await updateShowcaseFamily(editing.value.id, form)
      const index = families.value.findIndex((item) => item.id === updated.id)
      if (index >= 0) families.value.splice(index, 1, updated)
    } else {
      const created = await createShowcaseFamily(form)
      families.value.unshift(created)
      dialogVisible.value = false
      await router.push(`/admin/showcase-families/${created.id}/tree`)
      return
    }
    families.value.sort((left, right) => left.sortOrder - right.sortOrder || right.id - left.id)
    dialogVisible.value = false
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

async function togglePublished(item: ShowcaseFamily) {
  if (item.status !== 'PUBLISHED') {
    try {
      await ElMessageBox.confirm(`确认发布“${item.displayName}”？发布后会进入帝王世家入口。`, '确认发布', { type: 'warning', confirmButtonText: '发布', cancelButtonText: '取消' })
    } catch {
      return
    }
  }
  saving[item.id] = true
  operationError.value = ''
  try {
    const updated = await setShowcaseFamilyPublished(item.id, item.status !== 'PUBLISHED')
    Object.assign(item, updated)
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    saving[item.id] = false
  }
}

function handleCommand(command: string, item: ShowcaseFamily) {
  if (command === 'edit') openEdit(item)
  if (command === 'toggle') void togglePublished(item)
  if (command === 'delete') void remove(item)
}

function actionItems(item: ShowcaseFamily) {
  return [
    { command: 'edit', label: '编辑信息' },
    { command: 'toggle', label: item.status === 'PUBLISHED' ? '下架展示' : '发布展示' },
    { command: 'delete', label: '删除帝王世家', divided: true, danger: true }
  ]
}

async function remove(item: ShowcaseFamily) {
  try {
    await ElMessageBox.confirm(`将删除“${item.displayName}”及其 ${item.memberCount} 名成员、${item.relationshipCount} 条关系，该操作不可恢复。`, '确认删除', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  } catch {
    return
  }
  saving[item.id] = true
  operationError.value = ''
  try {
    await deleteShowcaseFamily(item.id)
    families.value = families.value.filter((row) => row.id !== item.id)
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    saving[item.id] = false
  }
}

onMounted(() => { void load(); void loadDisplaySettings() })
</script>

<style scoped>
.showcase-card { min-width: 0; }
.notice-bar { display: flex; align-items: center; gap: 12px; min-height: 42px; padding: 10px 16px; border: 1px solid rgba(200, 164, 93, .24); border-radius: 12px; color: var(--el-text-color-regular); background: rgba(255, 250, 239, .78); font-size: 13px; }
.notice-bar strong { flex: 0 0 auto; color: #8b6b35; }
.rotation-panel { display: grid; grid-template-columns: minmax(260px, 1fr) auto auto; align-items: center; gap: 18px; padding: 14px 16px; border: 1px solid rgba(31, 58, 95, .12); border-radius: 12px; background: rgba(255, 255, 255, .76); }
.rotation-panel strong, .rotation-panel span { display: block; }
.rotation-panel strong { color: var(--color-primary); font-size: 14px; }
.rotation-panel > div:first-child span, .rotation-status { margin-top: 4px; color: var(--el-text-color-secondary); font-size: 12px; }
.rotation-controls { display: flex; align-items: center; gap: 8px; white-space: nowrap; }
.rotation-controls :deep(.el-input-number) { width: 98px; }
.rotation-status { min-width: 188px; margin-top: 0; }
.showcase-table { width: 100%; min-width: 0; }
.showcase-name { display: flex; align-items: center; gap: 10px; min-width: 0; }
.showcase-name :deep(.table-cell-stack) { flex: 1; }
.cover-preview { width: 46px; height: 38px; flex: 0 0 auto; border-radius: 6px; border: 1px solid var(--el-border-color-lighter); }
.image-fallback { display: grid; width: 100%; height: 100%; place-items: center; color: var(--el-text-color-secondary); font-size: 11px; background: var(--el-fill-color-light); }
.showcase-table :deep(.showcase-center-header .cell),
.showcase-table :deep(.showcase-center-cell .cell) {
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
}
.showcase-table :deep(.showcase-center-cell .cell) { min-height: 100%; }
.table-status-stack__hint { color: var(--color-warning); font-size: 12px; white-space: nowrap; }
.field-tip { margin-left: 10px; color: var(--el-text-color-secondary); font-size: 12px; }
@media (max-width: 760px) { .notice-bar, .rotation-panel { align-items: flex-start; flex-direction: column; gap: 8px; } .rotation-panel { display: flex; } }
</style>
