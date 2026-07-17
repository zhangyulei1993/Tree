<template>
  <div class="page-stack">
    <PageHeader
      title="账号权益配置"
      description="按信任等级配置用户可创建家庭、家庭成员与加入家庭的数量上限。保存后立即生效，无需重新部署小程序。"
    />

    <el-alert v-if="loadError" :title="`加载失败：${loadError}`" type="error" show-icon :closable="false" />
    <el-alert v-if="operationError" :title="`操作失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <div v-loading="loading" class="quota-grid">
      <el-card v-for="item in forms" :key="item.trustTier" class="quota-card" shadow="never">
        <template #header>
          <div class="card-header">
            <strong>{{ tierLabel(item.trustTier) }}</strong>
            <span class="muted">最近更新：{{ formatTime(item.updatedAt) }}</span>
          </div>
        </template>

        <el-form label-width="180px">
          <el-form-item label="可创建家庭数">
            <el-input-number v-model="item.maxOwnedFamilies" :min="0" :max="99" />
          </el-form-item>
          <el-form-item label="每个家庭成员上限">
            <el-input-number v-model="item.maxMembersPerOwnedFamily" :min="1" :max="999" />
          </el-form-item>
          <el-form-item label="可加入家庭数">
            <el-input-number v-model="item.maxJoinedFamilies" :min="0" :max="99" />
          </el-form-item>
          <el-divider content-position="left">字辈权益</el-divider>
          <el-form-item label="支持字辈体系">
            <el-switch
              v-model="item.supportsGenerationNaming"
              active-text="开启"
              inactive-text="关闭"
            />
          </el-form-item>
        </el-form>

        <div class="impact-box" v-if="item.impact">
          <p>保存后预计影响：</p>
          <ul>
            <li>受影响用户：{{ item.impact.affectedUsers }}</li>
            <li>受影响家庭：{{ item.impact.affectedFamilies }}</li>
          </ul>
        </div>

        <div class="card-actions">
          <el-button :loading="item.previewing" @click="preview(item)">预览影响范围</el-button>
          <el-button type="primary" :loading="item.saving" @click="save(item)">保存配置</el-button>
        </div>
      </el-card>
    </div>

    <el-card class="feature-card" shadow="never">
      <template #header>
        <div class="card-header">
          <strong>字辈功能灰度体验</strong>
          <span class="muted">指定手机号可提前体验字辈体系；不改变账号等级权益配置。</span>
        </div>
      </template>

      <el-form label-width="180px">
        <el-form-item label="体验手机号">
          <el-input
            v-model="generationNamingPhonesInput"
            type="textarea"
            :rows="4"
            placeholder="输入完整手机号，多个号码可换行、逗号或空格分隔"
          />
        </el-form-item>
      </el-form>

      <div class="override-list">
        <span class="muted">当前名单：</span>
        <el-tag v-for="item in generationNamingOverrides" :key="item.phoneMask" type="info">
          {{ item.phoneMask }}
        </el-tag>
        <span v-if="generationNamingOverrides.length === 0" class="muted">暂无</span>
      </div>

      <div class="card-actions">
        <el-button :loading="overrideLoading" @click="loadGenerationNamingOverrides">刷新名单</el-button>
        <el-button type="primary" :loading="overrideSaving" @click="saveGenerationNamingOverrides">保存体验名单</el-button>
        <el-button
          type="danger"
          plain
          :disabled="generationNamingOverrides.length === 0"
          :loading="overrideSaving"
          @click="clearGenerationNamingOverrides"
        >
          清空名单
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessageBox } from 'element-plus'

import { getApiErrorMessage } from '@/api/client'
import {
  listAccountFeatureOverrides,
  listAccountQuotaConfigs,
  previewAccountQuotaImpact,
  updateAccountFeatureOverrides,
  updateAccountQuotaConfig
} from '@/api/accountQuota'
import PageHeader from '@/components/PageHeader.vue'
import type { AccountFeatureOverrideItem, AccountQuotaConfig, AccountQuotaImpactPreview } from '@/types/api'

interface QuotaForm extends AccountQuotaConfig {
  saving: boolean
  previewing: boolean
  impact: AccountQuotaImpactPreview | null
}

const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const forms = reactive<QuotaForm[]>([])
const generationNamingPhonesInput = ref('')
const generationNamingOverrides = ref<AccountFeatureOverrideItem[]>([])
const overrideLoading = ref(false)
const overrideSaving = ref(false)
const GENERATION_NAMING_FEATURE = 'GENERATION_NAMING'

function tierLabel(tier: string) {
  return tier === 'PHONE_BOUND' ? '备用登录已开启' : '仅微信登录'
}

function formatTime(value?: string) {
  if (!value) return '-'
  return value.replace('T', ' ').slice(0, 19)
}

function snapshot(config: AccountQuotaConfig): QuotaForm {
  return {
    ...config,
    saving: false,
    previewing: false,
    impact: null
  }
}

function isLowering(current: AccountQuotaConfig, next: QuotaForm) {
  return (
    next.maxOwnedFamilies < current.maxOwnedFamilies ||
    next.maxMembersPerOwnedFamily < current.maxMembersPerOwnedFamily ||
    next.maxJoinedFamilies < current.maxJoinedFamilies ||
    (current.supportsGenerationNaming && !next.supportsGenerationNaming)
  )
}

async function loadConfigs() {
  loading.value = true
  loadError.value = ''
  try {
    const rows = await listAccountQuotaConfigs()
    forms.splice(0, forms.length, ...rows.map(snapshot))
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function loadGenerationNamingOverrides() {
  overrideLoading.value = true
  operationError.value = ''
  try {
    generationNamingOverrides.value = await listAccountFeatureOverrides(GENERATION_NAMING_FEATURE)
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    overrideLoading.value = false
  }
}

function parsePhonesInput(value: string) {
  return value
    .split(/[\s,，;；]+/)
    .map((item) => item.trim())
    .filter(Boolean)
}

async function saveGenerationNamingOverrides() {
  const phones = parsePhonesInput(generationNamingPhonesInput.value)
  try {
    await ElMessageBox.confirm(
      `保存后将用这 ${phones.length} 个手机号替换当前字辈体验名单。确认保存？`,
      '确认保存体验名单',
      { type: 'warning', confirmButtonText: '确认保存', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  overrideSaving.value = true
  operationError.value = ''
  try {
    generationNamingOverrides.value = await updateAccountFeatureOverrides(GENERATION_NAMING_FEATURE, phones)
    generationNamingPhonesInput.value = ''
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    overrideSaving.value = false
  }
}

async function clearGenerationNamingOverrides() {
  if (generationNamingOverrides.value.length === 0) {
    return
  }
  try {
    await ElMessageBox.confirm(
      '清空后，当前灰度手机号将全部移出字辈体验名单。确认清空？',
      '确认清空体验名单',
      { type: 'warning', confirmButtonText: '确认清空', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  overrideSaving.value = true
  operationError.value = ''
  try {
    generationNamingOverrides.value = await updateAccountFeatureOverrides(GENERATION_NAMING_FEATURE, [])
    generationNamingPhonesInput.value = ''
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    overrideSaving.value = false
  }
}

async function preview(item: QuotaForm): Promise<boolean> {
  item.previewing = true
  operationError.value = ''
  try {
    item.impact = await previewAccountQuotaImpact(item.trustTier, {
      maxOwnedFamilies: item.maxOwnedFamilies,
      maxMembersPerOwnedFamily: item.maxMembersPerOwnedFamily,
      maxJoinedFamilies: item.maxJoinedFamilies,
      supportsGenerationNaming: item.supportsGenerationNaming
    })
    return true
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
    return false
  } finally {
    item.previewing = false
  }
}

async function save(item: QuotaForm) {
  const original = await listAccountQuotaConfigs().then((rows) => rows.find((row) => row.trustTier === item.trustTier))
  if (!original) {
    operationError.value = '未找到原始配置'
    return
  }

  if (!(await preview(item))) {
    return
  }

  if (isLowering(original, item)) {
    try {
      await ElMessageBox.confirm(
        `降低 ${tierLabel(item.trustTier)} 额度后，已有家庭、成员和绑定不会被删除，但用户将无法继续新增、恢复或加入。预计影响 ${item.impact?.affectedUsers ?? 0} 位用户、${item.impact?.affectedFamilies ?? 0} 个家庭。确认保存？`,
        '确认降低额度',
        { type: 'warning', confirmButtonText: '确认保存', cancelButtonText: '取消' }
      )
    } catch {
      return
    }
  }

  item.saving = true
  operationError.value = ''
  try {
    const saved = await updateAccountQuotaConfig(item.trustTier, {
      maxOwnedFamilies: item.maxOwnedFamilies,
      maxMembersPerOwnedFamily: item.maxMembersPerOwnedFamily,
      maxJoinedFamilies: item.maxJoinedFamilies,
      supportsGenerationNaming: item.supportsGenerationNaming
    })
    Object.assign(item, snapshot(saved))
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    item.saving = false
  }
}

onMounted(() => {
  loadConfigs()
  loadGenerationNamingOverrides()
})
</script>

<style scoped>
.quota-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 20px;
}

.quota-card,
.feature-card {
  border-radius: 16px;
}

.feature-card {
  margin-top: 20px;
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.muted {
  color: #7a8797;
  font-size: 13px;
}

.impact-box {
  margin: 8px 0 16px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f6f8fb;
  color: #4f5d6b;
}

.impact-box ul {
  margin: 8px 0 0;
  padding-left: 18px;
}

.card-actions {
  display: flex;
  gap: 12px;
}

.override-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 8px 0 16px;
  align-items: center;
}
</style>
