<template>
  <div class="page-stack">
    <PageHeader title="优质家庭邀请规则" description="配置家庭达到什么条件后，平台自动发出优质家庭申请资格。家庭创建者仍需自行决定是否提交申请。修改后立即生效，无需重新发布小程序。" />
    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />
    <el-card v-loading="loading" shadow="never">
      <template #header><div class="card-head"><strong>自动准入条件</strong><el-tag type="info" effect="plain">仅 ROOT_ADMIN / SUPER_ADMIN</el-tag></div></template>
      <el-alert type="info" :closable="false" show-icon title="建议先使用宽松条件收集案例，再逐步提高资料完整度和绑定要求。" />
      <el-form v-if="form" class="config-form" label-width="190px">
        <el-form-item label="开放自动邀请资格"><el-switch v-model="form.enabled" active-text="开放" inactive-text="关闭" /></el-form-item>
        <el-form-item label="最少活跃节点数"><el-input-number v-model="form.minMemberCount" :min="1" :max="100000" /><el-text type="info">家庭至少有多少个有效节点。</el-text></el-form-item>
        <el-form-item label="最少亲属关系数"><el-input-number v-model="form.minRelationshipCount" :min="0" :max="100000" /><el-text type="info">用于确认家庭树不是孤立节点集合。</el-text></el-form-item>
        <el-form-item label="最少绑定用户数"><el-input-number v-model="form.minBoundMemberCount" :min="0" :max="100000" /><el-text type="info">至少有多少节点已绑定真实账号。</el-text></el-form-item>
        <el-form-item label="节点资料完整度"><el-input-number v-model="form.minProfileCompleteRate" :min="0" :max="100" /><span class="suffix">%</span><el-text type="info">姓名、性别、出生年份完整的节点比例；0 表示不限制。</el-text></el-form-item>
        <el-form-item label="节点绑定比例"><el-input-number v-model="form.minBoundMemberRate" :min="0" :max="100" /><span class="suffix">%</span><el-text type="info">已绑定真实用户的节点比例；0 表示不限制。</el-text></el-form-item>
        <el-form-item label="要求家庭简介"><el-switch v-model="form.requireDescription" /></el-form-item>
        <el-form-item label="要求家庭所在地"><el-switch v-model="form.requireLocation" /></el-form-item>
        <el-form-item><el-button type="primary" :loading="saving" @click="save">保存准入规则</el-button><el-button @click="load">重新读取</el-button></el-form-item>
      </el-form>
      <el-descriptions v-if="form" :column="1" border class="summary">
        <el-descriptions-item label="当前初始建议">3 个节点、2 条关系、1 个绑定用户，资料和绑定比例暂不设硬性比例。</el-descriptions-item>
        <el-descriptions-item label="最近更新">{{ formatTime(form.updatedAt) }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import PageHeader from '@/components/PageHeader.vue'
import { getApiErrorMessage } from '@/api/client'
import { getQualityFamilyEligibilityConfig, updateQualityFamilyEligibilityConfig, type QualityFamilyEligibilityConfig } from '@/api/qualityFamily'

const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const form = ref<QualityFamilyEligibilityConfig | null>(null)
function formatTime(value: string) { return value ? value.replace('T', ' ').slice(0, 19) : '暂无' }
async function load() { loading.value = true; errorMessage.value = ''; try { form.value = await getQualityFamilyEligibilityConfig() } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { loading.value = false } }
async function save() { if (!form.value) return; saving.value = true; errorMessage.value = ''; try { form.value = await updateQualityFamilyEligibilityConfig({ ...form.value }); ElMessage.success('优质家庭准入规则已保存') } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { saving.value = false } }
onMounted(load)
</script>

<style scoped>
.card-head { display:flex; align-items:center; justify-content:space-between; }.config-form { max-width:900px; margin-top:22px; }.config-form :deep(.el-text) { margin-left:12px; }.suffix { margin-left:8px; }.summary { max-width:900px; margin-top:22px; }
</style>
