<template>
  <div class="page-stack">
    <PageHeader title="分享配置" description="修改小程序首页、公开家庭和邀请分享的标题与封面图。保存后新分享立即使用。" />
    <el-alert v-if="loadError" :title="`加载失败：${loadError}`" type="error" show-icon :closable="false" />
    <el-alert v-if="operationError" :title="`保存失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <div v-loading="loading" class="share-grid">
      <el-card v-for="item in forms" :key="item.configKey" class="share-card" shadow="never">
        <template #header>
          <div class="card-header">
            <div>
              <div class="card-title-row">
                <strong>{{ label(item.configKey) }}</strong>
                <el-tag size="small" effect="plain">{{ item.configKey }}</el-tag>
              </div>
              <p class="card-description">{{ item.description || '小程序分享配置' }}</p>
            </div>
            <el-tag type="success" effect="plain">已启用</el-tag>
          </div>
        </template>
        <el-form label-position="top">
          <div class="field-block">
            <div class="field-label">分享标题</div>
            <el-input v-model="item.titleTemplate" maxlength="300" show-word-limit />
            <p class="hint">可用变量：{{ item.configKey === 'HOME' ? '无' : 'familyName、targetMemberName（使用双大括号包裹）' }}</p>
          </div>
          <div class="field-block image-block">
            <div class="field-heading">
              <div>
                <div class="field-label">分享图片</div>
                <div class="field-help">选择一张作为当前分享封面，也可以继续上传备用图片。</div>
              </div>
              <span class="library-count">图片库 {{ item.imageUrls.length }}/5</span>
            </div>
            <div class="image-workspace">
              <div class="current-preview">
                <span class="preview-label">当前封面</span>
                <img v-if="item.imageUrl" :src="item.imageUrl" alt="当前分享图片" @error="failedPreview[item.configKey] = true" />
                <span v-if="failedPreview[item.configKey]" class="preview-error">图片暂时无法加载</span>
              </div>
              <div class="library-panel">
                <div class="image-library">
                  <div
                    v-for="imageUrl in item.imageUrls"
                    :key="imageUrl"
                    class="library-image"
                    :class="{ active: imageUrl === item.imageUrl }"
                    @click="selectImage(item, imageUrl)"
                  >
                    <img :src="imageUrl" alt="分享图片" @error="failedPreview[`${item.configKey}:${imageUrl}`] = true" />
                    <span v-if="failedPreview[`${item.configKey}:${imageUrl}`]">加载失败</span>
                    <i v-if="imageUrl === item.imageUrl">当前</i>
                    <el-button
                      v-if="item.imageUrls.length > 1"
                      class="remove-image"
                      circle
                      size="small"
                      @click.stop="removeImage(item, imageUrl)"
                    >
                      ×
                    </el-button>
                  </div>
                </div>
                <el-upload
                  accept="image/jpeg,image/png,image/webp"
                  :show-file-list="false"
                  :before-upload="beforeImageUpload"
                  :http-request="(options: UploadRequestOptions) => uploadImage(item, options)"
                >
                  <el-button plain :loading="Boolean(uploading[item.configKey])">上传图片</el-button>
                </el-upload>
              </div>
            </div>
            <div class="image-url-field">
              <span class="url-label">图片地址</span>
              <el-input v-model="item.imageUrl" maxlength="500" placeholder="也可以直接填写 HTTPS 地址" clearable />
            </div>
          </div>
          <div class="card-footer">
            <span class="save-tip">上传或切换图片后，请保存配置</span>
            <el-button type="primary" :loading="item.saving" @click="save(item)">保存配置</el-button>
          </div>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { UploadRequestOptions, UploadRawFile } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadContentImage } from '@/api/content'
import { getApiErrorMessage } from '@/api/client'
import { listShareConfigs, updateShareConfig } from '@/api/shareConfig'
import PageHeader from '@/components/PageHeader.vue'
import type { ShareConfig } from '@/types/api'

interface Form extends ShareConfig { saving: boolean }
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const forms = reactive<Form[]>([])
const failedPreview = reactive<Record<string, boolean>>({})
const uploading = reactive<Record<string, boolean>>({})

function label(key: string) {
  return ({ HOME: '首页分享', PUBLIC_FAMILY: '公开家庭分享', INVITATION_NODE: '节点绑定邀请', INVITATION_PENDING_MEMBER: '暂存成员邀请' } as Record<string, string>)[key] || key
}

async function load() {
  loading.value = true
  loadError.value = ''
  try { forms.splice(0, forms.length, ...(await listShareConfigs()).map((item) => ({ ...item, saving: false }))) }
  catch (error) { loadError.value = getApiErrorMessage(error) }
  finally { loading.value = false }
}

async function save(item: Form) {
  item.saving = true
  operationError.value = ''
  try {
    const updated = await updateShareConfig(item.configKey, { titleTemplate: item.titleTemplate, imageUrl: item.imageUrl, imageUrls: item.imageUrls, description: item.description })
    if (updated) Object.assign(item, updated)
  } catch (error) { operationError.value = getApiErrorMessage(error) }
  finally { item.saving = false }
}

function beforeImageUpload(file: UploadRawFile) {
  if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
    ElMessage.warning('仅支持 JPG、PNG、WEBP 图片')
    return false
  }
  if (file.size <= 0 || file.size > 5 * 1024 * 1024) {
    ElMessage.warning('图片大小需在 5MB 以内')
    return false
  }
  return true
}

async function uploadImage(item: Form, options: UploadRequestOptions) {
  uploading[item.configKey] = true
  try {
    const result = await uploadContentImage(options.file)
    if (!item.imageUrls.includes(result.url)) {
      if (item.imageUrls.length >= 5) {
        ElMessage.warning('每组图片库最多保留 5 张图片')
        options.onError?.(new Error('图片库已满') as any)
        return
      }
      item.imageUrls.push(result.url)
    }
    item.imageUrl = result.url
    failedPreview[item.configKey] = false
    options.onSuccess?.(result)
    ElMessage.success('分享图片已上传，请保存配置')
  } catch (error) {
    options.onError?.(error as any)
    ElMessage.error(getApiErrorMessage(error))
  } finally {
    uploading[item.configKey] = false
  }
}

function selectImage(item: Form, imageUrl: string) {
  item.imageUrl = imageUrl
  failedPreview[item.configKey] = false
}

function removeImage(item: Form, imageUrl: string) {
  const index = item.imageUrls.indexOf(imageUrl)
  if (index < 0 || item.imageUrls.length <= 1) return
  item.imageUrls.splice(index, 1)
  if (item.imageUrl === imageUrl) item.imageUrl = item.imageUrls[Math.max(0, index - 1)]
}

onMounted(load)
</script>

<style scoped>
.share-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 20px; align-items: start; }
.share-card { border-radius: 12px; }
.card-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.card-title-row { display: flex; align-items: center; gap: 10px; color: var(--color-primary); font-size: 17px; }
.card-description, .muted, .hint, .field-help { color: var(--color-text-secondary); font-size: 13px; }
.card-description { margin: 7px 0 0; line-height: 1.5; }
.field-block { margin-bottom: 22px; }
.field-label { margin-bottom: 8px; color: var(--color-text-primary); font-size: 14px; font-weight: 600; }
.hint { margin: 8px 0 0; line-height: 1.5; }
.field-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.field-help { line-height: 1.5; }
.library-count { flex: none; color: var(--color-text-secondary); font-size: 12px; }
.image-workspace { display: grid; grid-template-columns: 148px minmax(0, 1fr); gap: 16px; padding: 14px; border: 1px solid var(--color-border); border-radius: 10px; background: var(--color-bg-admin); }
.current-preview { position: relative; height: 112px; overflow: hidden; border-radius: 8px; background: var(--color-bg-card); }
.preview-label { position: absolute; z-index: 1; top: 7px; left: 7px; padding: 2px 6px; color: white; font-size: 11px; background: rgba(31, 58, 95, 0.82); border-radius: 4px; }
.current-preview img { width: 100%; height: 100%; object-fit: cover; }
.preview-error { display: grid; place-items: center; height: 100%; padding: 10px; color: var(--color-text-secondary); font-size: 12px; text-align: center; }
.library-panel { min-width: 0; }
.image-library { display: flex; flex-wrap: wrap; gap: 8px; min-height: 90px; }
.library-image { position: relative; width: 82px; height: 64px; overflow: hidden; border: 1px solid var(--color-border); border-radius: 6px; background: var(--color-bg-card); cursor: pointer; }
.library-image.active { border: 2px solid var(--color-primary); }
.library-image img { width: 100%; height: 100%; object-fit: cover; }
.library-image span { position: absolute; inset: 0; display: grid; place-items: center; color: var(--color-text-secondary); background: var(--color-bg-admin); font-size: 11px; }
.library-image i { position: absolute; left: 3px; bottom: 3px; padding: 2px 4px; color: white; font-style: normal; font-size: 10px; background: var(--color-primary); border-radius: 3px; }
.remove-image { position: absolute; top: 2px; right: 2px; opacity: 0.9; transform: scale(0.72); transform-origin: top right; }
.library-panel > .el-upload { display: inline-block; margin-top: 10px; }
.image-url-field { display: flex; align-items: center; gap: 10px; margin-top: 12px; }
.image-url-field .el-input { flex: 1; }
.url-label { flex: none; color: var(--color-text-secondary); font-size: 12px; }
.card-footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding-top: 16px; border-top: 1px solid var(--color-border); }
.save-tip { color: var(--color-text-secondary); font-size: 12px; }
@media (max-width: 1100px) { .share-grid { grid-template-columns: 1fr; } }
@media (max-width: 560px) { .image-workspace { grid-template-columns: 1fr; } .current-preview { height: 150px; } .card-footer { align-items: flex-start; flex-direction: column; } }
</style>
