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

    <el-card class="splash-card" shadow="never">
      <template #header><div class="splash-card-title">小程序开屏</div></template>
      <el-form v-if="splashConfig" :model="splashConfig" label-width="100px" inline>
        <el-form-item label="开屏状态"><el-switch v-model="splashConfig.enabled" active-text="启用" inactive-text="关闭" /></el-form-item>
        <el-form-item label="跳过按钮"><el-switch v-model="splashConfig.skipEnabled" active-text="显示" inactive-text="隐藏" /></el-form-item>
        <el-form-item label="展示时长"><el-input-number v-model="splashConfig.durationMs" :min="1500" :max="15000" :step="100" /><span class="unit-text">毫秒</span></el-form-item>
        <el-form-item label="白天图片" class="splash-wide-item">
          <div class="splash-image-field">
            <el-image v-if="splashConfig.dayImageUrl" :src="splashConfig.dayImageUrl" fit="cover" class="splash-image-preview" />
            <el-upload
              :show-file-list="false"
              accept="image/jpeg,image/png,image/webp"
              :before-upload="beforeSplashImageUpload"
              :http-request="(options: UploadRequestOptions) => uploadSplashImage('day', options)"
            >
              <el-button :loading="splashImageUploading.day || splashImageCompressing.day">{{ splashConfig.dayImageUrl ? '更换图片' : '上传图片' }}</el-button>
            </el-upload>
            <el-button v-if="splashConfig.dayImageUrl" text type="danger" @click="clearSplashImage('day')">清除</el-button>
            <span class="splash-field-help">06:00 至 17:59 使用，上传后自动压缩到尽量小。</span>
          </div>
        </el-form-item>
        <el-form-item label="夜晚图片" class="splash-wide-item">
          <div class="splash-image-field">
            <el-image v-if="splashConfig.nightImageUrl" :src="splashConfig.nightImageUrl" fit="cover" class="splash-image-preview" />
            <el-upload
              :show-file-list="false"
              accept="image/jpeg,image/png,image/webp"
              :before-upload="beforeSplashImageUpload"
              :http-request="(options: UploadRequestOptions) => uploadSplashImage('night', options)"
            >
              <el-button :loading="splashImageUploading.night || splashImageCompressing.night">{{ splashConfig.nightImageUrl ? '更换图片' : '上传图片' }}</el-button>
            </el-upload>
            <el-button v-if="splashConfig.nightImageUrl" text type="danger" @click="clearSplashImage('night')">清除</el-button>
            <span class="splash-field-help">18:00 至 05:59 使用；只上传一张时全天使用，上传后自动压缩到尽量小。</span>
          </div>
        </el-form-item>
        <el-form-item label="主标题" class="splash-wide-item"><el-input v-model="splashConfig.title" maxlength="120" /></el-form-item>
        <el-form-item label="副标题" class="splash-wide-item"><el-input v-model="splashConfig.tagline" maxlength="240" /></el-form-item>
        <el-form-item class="splash-submit-item"><el-button type="primary" :loading="splashSaving" @click="saveSplash">保存开屏配置</el-button></el-form-item>
      </el-form>
    </el-card>

    <DataTable v-loading="loading" class="tool-config-card">
      <el-table :data="configs" row-key="toolKey" class="tool-table">
        <el-table-column prop="displayName" label="工具名称" min-width="140" />
        <el-table-column prop="toolKey" label="配置标识" min-width="190">
          <template #default="{ row }">
            <el-tag effect="plain">{{ row.toolKey }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="功能说明" min-width="260" show-overflow-tooltip />
        <el-table-column label="展示" width="82" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">
            <el-tag :type="row.visible ? 'success' : 'info'" effect="plain">
              {{ row.visible ? '已展示' : '不展示' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="功能" width="82" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'warning'" effect="plain">
              {{ row.enabled ? '已开放' : '未开放' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="排序" width="118" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
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
        <el-table-column label="最近更新" width="160" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="168" fixed="right" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">
            <div class="tool-actions">
              <el-button size="small" :type="row.visible ? 'success' : 'default'" :loading="Boolean(saving[row.toolKey])" @click="toggleField(row, 'visible')">
                {{ row.visible ? '隐藏' : '展示' }}
              </el-button>
              <el-button size="small" :type="row.enabled ? 'success' : 'warning'" :loading="Boolean(saving[row.toolKey])" @click="toggleField(row, 'enabled')">
                {{ row.enabled ? '停用' : '启用' }}
              </el-button>
              <el-dropdown trigger="click" @command="handleAction(row, $event)">
                <el-button size="small" :disabled="Boolean(saving[row.toolKey])">更多</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="pinned">{{ row.pinned ? '取消置顶' : '置顶' }}</el-dropdown-item>
                    <el-dropdown-item command="highlighted">{{ row.highlighted ? '取消高亮' : '高亮' }}</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </DataTable>

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
import { ElMessage } from 'element-plus/es/components/message/index'
import type { FormInstance, FormRules, UploadRawFile, UploadRequestOptions } from 'element-plus'

import { getApiErrorMessage } from '@/api/client'
import { createToolConfig, listToolConfigs, updateToolConfig } from '@/api/toolConfig'
import { getLaunchSplashConfig, updateLaunchSplashConfig } from '@/api/launchSplash'
import { uploadContentImage } from '@/api/content'
import { compressUploadImage } from '@/utils/imageCompression'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import type { ToolConfig } from '@/types/api'
import type { LaunchSplashConfig } from '@/types/api'

const configs = ref<ToolConfig[]>([])
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const saving = reactive<Record<string, boolean>>({})
const createVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({ toolKey: '', displayName: '', description: '' })
const splashConfig = ref<LaunchSplashConfig | null>(null)
const splashSaving = ref(false)
const splashImageUploading = reactive({ day: false, night: false })
const splashImageCompressing = reactive({ day: false, night: false })
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
    const [toolRows, splash] = await Promise.all([listToolConfigs(), getLaunchSplashConfig()])
    configs.value = toolRows
    splashConfig.value = splash
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function saveSplash(notify = true): Promise<boolean> {
  if (!splashConfig.value) return false
  splashSaving.value = true
  operationError.value = ''
  try {
    splashConfig.value = await updateLaunchSplashConfig({
      enabled: splashConfig.value.enabled,
      title: splashConfig.value.title,
      tagline: splashConfig.value.tagline,
      imageUrl: splashConfig.value.imageUrl,
      dayImageUrl: splashConfig.value.dayImageUrl,
      nightImageUrl: splashConfig.value.nightImageUrl,
      durationMs: splashConfig.value.durationMs,
      skipEnabled: splashConfig.value.skipEnabled
    })
    if (notify) ElMessage.success('开屏配置已保存')
    return true
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
    return false
  } finally { splashSaving.value = false }
}

function beforeSplashImageUpload(file: UploadRawFile) {
  if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
    operationError.value = '仅支持 JPG、PNG、WEBP 图片'
    return false
  }
  if (file.size > 5 * 1024 * 1024) {
    operationError.value = '图片大小需在 5MB 以内'
    return false
  }
  operationError.value = ''
  return true
}

type SplashImageSlot = 'day' | 'night'

async function uploadSplashImage(slot: SplashImageSlot, options: UploadRequestOptions) {
  if (!splashConfig.value) return
  splashImageUploading[slot] = true
  splashImageCompressing[slot] = true
  operationError.value = ''
  try {
    const compressedFile = await compressUploadImage(options.file)
    splashImageCompressing[slot] = false
    const result = await uploadContentImage(compressedFile)
    if (slot === 'day') splashConfig.value.dayImageUrl = result.url
    else splashConfig.value.nightImageUrl = result.url
    const saved = await saveSplash(false)
    if (!saved) {
      options.onError?.(new Error('开屏配置保存失败') as Parameters<NonNullable<UploadRequestOptions['onError']>>[0])
      return
    }
    ElMessage.success(`${slot === 'day' ? '白天' : '夜晚'}开屏图片已压缩并替换`)
    options.onSuccess?.(result)
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
    options.onError?.(error as Parameters<NonNullable<UploadRequestOptions['onError']>>[0])
  } finally {
    splashImageUploading[slot] = false
    splashImageCompressing[slot] = false
  }
}

async function clearSplashImage(slot: SplashImageSlot) {
  if (!splashConfig.value) return
  if (slot === 'day') splashConfig.value.dayImageUrl = ''
  else splashConfig.value.nightImageUrl = ''
  if (await saveSplash(false)) ElMessage.success(`${slot === 'day' ? '白天' : '夜晚'}开屏图片已删除`)
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

function handleAction(item: ToolConfig, command: unknown) {
  if (command === 'pinned' || command === 'highlighted') void toggleField(item, command)
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
  min-width: 0;
}

.splash-card { margin-bottom: 16px; }
.splash-card-title { font-weight: 650; }
.splash-wide-item { min-width: 320px; }
.splash-submit-item { margin-left: 100px; }
.unit-text { margin-left: 8px; color: var(--el-text-color-secondary); }
.splash-image-field { display: flex; align-items: center; flex-wrap: wrap; gap: 10px; }
.splash-image-preview { width: 76px; height: 96px; border: 1px solid var(--el-border-color); border-radius: 6px; background: var(--el-fill-color-light); }
.splash-field-help { color: var(--el-text-color-secondary); font-size: 12px; }

.tool-table {
  width: 100%;
  min-width: 970px;
}

.tool-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 6px;
  width: 100%;
  white-space: normal;
}

.tool-actions :deep(.el-button) {
  padding: 5px 7px;
}

.tool-actions :deep(.el-button + .el-button) {
  margin-left: 0;
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
