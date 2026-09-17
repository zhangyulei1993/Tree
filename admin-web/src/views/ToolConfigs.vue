<template>
  <div class="page-stack">
    <PageHeader
      title="常用工具配置"
      description="控制小程序常用工具是否对用户开放。新工具默认停用，完成小程序接入后再启用。"
    >
      <el-button type="primary" @click="openCreate">新增工具</el-button>
    </PageHeader>

    <el-alert v-if="loadError" :title="`加载失败：${loadError}`" type="error" show-icon :closable="false" />
    <el-alert v-if="operationError" :title="`操作失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <el-card v-loading="loading" class="tool-config-card" shadow="never">
      <el-table :data="configs" row-key="toolKey" class="tool-table">
        <el-table-column prop="displayName" label="工具名称" min-width="110" />
        <el-table-column prop="toolKey" label="配置标识" min-width="170">
          <template #default="{ row }">
            <el-tag effect="plain">{{ row.toolKey }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="功能说明" min-width="190" show-overflow-tooltip />
        <el-table-column label="展示" width="82">
          <template #default="{ row }">
            <el-tag :type="row.visible ? 'success' : 'info'" effect="plain">
              {{ row.visible ? '已展示' : '不展示' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="功能" width="82">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'warning'" effect="plain">
              {{ row.enabled ? '已开放' : '未开放' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="排序" width="132">
          <template #default="{ row }">
            <div class="sort-controls">
              <span class="sort-position">{{ sortPosition(row) }}</span>
              <el-button
                size="small"
                text
                :disabled="!canMove(row, -1) || Boolean(saving[row.toolKey])"
                title="上移"
                aria-label="上移"
                @click="moveTool(row, -1)"
              >
                ↑
              </el-button>
              <el-button
                size="small"
                text
                :disabled="!canMove(row, 1) || Boolean(saving[row.toolKey])"
                title="下移"
                aria-label="下移"
                @click="moveTool(row, 1)"
              >
                ↓
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="最近更新" width="140">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="225">
          <template #default="{ row }">
            <div class="tool-actions">
              <el-button size="small" :type="row.visible ? 'success' : 'default'" :loading="Boolean(saving[row.toolKey])" @click="toggleField(row, 'visible')">
                {{ row.visible ? '隐藏' : '展示' }}
              </el-button>
              <el-button size="small" :type="row.enabled ? 'success' : 'warning'" :loading="Boolean(saving[row.toolKey])" @click="toggleField(row, 'enabled')">
                {{ row.enabled ? '停用' : '启用' }}
              </el-button>
              <el-button size="small" :type="row.pinned ? 'primary' : 'default'" :loading="Boolean(saving[row.toolKey])" @click="toggleField(row, 'pinned')">
                {{ row.pinned ? '取消置顶' : '置顶' }}
              </el-button>
              <el-button size="small" :type="row.highlighted ? 'danger' : 'default'" :loading="Boolean(saving[row.toolKey])" @click="toggleField(row, 'highlighted')">
                {{ row.highlighted ? '取消高亮' : '高亮' }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createVisible" title="新增常用工具" width="520px" destroy-on-close>
      <el-alert
        title="新工具创建后默认停用，不会立即出现在小程序中。"
        type="info"
        show-icon
        :closable="false"
        class="create-tip"
      />
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="96px">
        <el-form-item label="工具标识" prop="toolKey">
          <el-input
            v-model="createForm.toolKey"
            maxlength="80"
            show-word-limit
            placeholder="例如 FAMILY_CONTACTS"
          />
        </el-form-item>
        <el-form-item label="工具名称" prop="displayName">
          <el-input v-model="createForm.displayName" maxlength="120" show-word-limit placeholder="例如 家庭联系人" />
        </el-form-item>
        <el-form-item label="功能说明" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="简要说明这个工具的用途"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="submitCreate">创建并停用</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import { getApiErrorMessage } from '@/api/client'
import { createToolConfig, listToolConfigs, updateToolConfig } from '@/api/toolConfig'
import PageHeader from '@/components/PageHeader.vue'
import type { ToolConfig } from '@/types/api'

const configs = ref<ToolConfig[]>([])
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const saving = reactive<Record<string, boolean>>({})
const createVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({ toolKey: '', displayName: '', description: '' })
const createRules: FormRules = {
  toolKey: [
    { required: true, message: '请输入工具标识', trigger: 'blur' },
    { pattern: /^[A-Za-z][A-Za-z0-9_]{0,79}$/, message: '只能使用字母、数字和下划线', trigger: 'blur' }
  ],
  displayName: [{ required: true, message: '请输入工具名称', trigger: 'blur' }],
  description: [{ required: true, message: '请输入功能说明', trigger: 'blur' }]
}

function formatTime(value?: string) {
  return value ? value.replace('T', ' ').slice(0, 19) : '-'
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    configs.value = await listToolConfigs()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

type BooleanToolField = 'visible' | 'enabled' | 'pinned' | 'highlighted'

async function toggleField(item: ToolConfig, field: BooleanToolField) {
  const previous = item[field]
  const nextValue = !previous
  item[field] = nextValue
  saving[item.toolKey] = true
  operationError.value = ''
  try {
    const updated = await updateToolConfig(item.toolKey, { [field]: nextValue })
    if (updated) Object.assign(item, updated)
  } catch (error) {
    item[field] = previous
    operationError.value = getApiErrorMessage(error)
  } finally {
    saving[item.toolKey] = false
  }
}

function compareToolConfigs(left: ToolConfig, right: ToolConfig) {
  if (left.pinned !== right.pinned) return left.pinned ? -1 : 1
  if (left.sortOrder !== right.sortOrder) return left.sortOrder - right.sortOrder
  if (left.highlighted !== right.highlighted) return left.highlighted ? -1 : 1
  return left.id - right.id
}

function sortGroup(item: ToolConfig) {
  return configs.value.filter((config) => config.pinned === item.pinned).sort(compareToolConfigs)
}

function sortPosition(item: ToolConfig) {
  const index = sortGroup(item).findIndex((config) => config.toolKey === item.toolKey)
  return index >= 0 ? index + 1 : '-'
}

function canMove(item: ToolConfig, direction: -1 | 1) {
  const group = sortGroup(item)
  const index = group.findIndex((config) => config.toolKey === item.toolKey)
  return index >= 0 && index + direction >= 0 && index + direction < group.length
}

async function moveTool(item: ToolConfig, direction: -1 | 1) {
  const group = sortGroup(item)
  const currentIndex = group.findIndex((config) => config.toolKey === item.toolKey)
  const targetIndex = currentIndex + direction
  if (currentIndex < 0 || targetIndex < 0 || targetIndex >= group.length) return

  const previousOrders = new Map(group.map((config) => [config.toolKey, config.sortOrder]))
  const reordered = [...group]
  const [moved] = reordered.splice(currentIndex, 1)
  if (!moved) return
  reordered.splice(targetIndex, 0, moved)
  reordered.forEach((config, index) => {
    config.sortOrder = index * 10
  })

  const changed = reordered.filter((config) => config.sortOrder !== previousOrders.get(config.toolKey))
  if (!changed.length) return

  operationError.value = ''
  changed.forEach((config) => {
    saving[config.toolKey] = true
  })
  configs.value.sort(compareToolConfigs)

  try {
    const updatedConfigs = await Promise.all(
      changed.map((config) => updateToolConfig(config.toolKey, { sortOrder: config.sortOrder }))
    )
    updatedConfigs.forEach((updated, index) => {
      if (updated) Object.assign(changed[index], updated)
    })
    configs.value.sort(compareToolConfigs)
  } catch (error) {
    changed.forEach((config) => {
      config.sortOrder = previousOrders.get(config.toolKey) ?? config.sortOrder
    })
    configs.value.sort(compareToolConfigs)
    operationError.value = getApiErrorMessage(error)
  } finally {
    changed.forEach((config) => {
      saving[config.toolKey] = false
    })
  }
}

function openCreate() {
  Object.assign(createForm, { toolKey: '', displayName: '', description: '' })
  createVisible.value = true
}

async function submitCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  createLoading.value = true
  operationError.value = ''
  try {
    const created = await createToolConfig(createForm)
    if (created) {
      configs.value.unshift(created)
      createVisible.value = false
    }
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    createLoading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.tool-config-card {
  border-radius: 12px;
}

.tool-table {
  width: 100%;
  min-width: 1090px;
}

.tool-actions {
  display: flex;
  flex-wrap: nowrap;
  gap: 6px;
  align-items: center;
  white-space: nowrap;
}

.tool-actions :deep(.el-button) {
  padding: 5px 7px;
}

.sort-controls {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.sort-position {
  min-width: 18px;
  color: var(--el-text-color-secondary);
  text-align: center;
}

.sort-controls :deep(.el-button) {
  min-width: 24px;
  padding: 4px;
  font-size: 16px;
}


.create-tip {
  margin-bottom: 18px;
}
</style>
