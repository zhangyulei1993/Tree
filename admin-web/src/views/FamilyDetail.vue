<template>
  <div class="page-stack">
    <PageHeader title="家庭详情" description="家庭真实基础信息与当前状态。">
      <el-button @click="$router.push('/admin/families')">返回家庭列表</el-button>
      <el-button type="primary" plain @click="$router.push(`/admin/families/${familyId}/tree`)">查看家庭树</el-button>
      <el-button @click="$router.push(`/admin/families/${familyId}/members`)">查看成员</el-button>
    </PageHeader>
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <template v-else-if="family">
      <section class="surface family-overview">
        <div>
          <span class="eyebrow">家庭 #{{ family.id }}</span>
          <h2>{{ family.familyName }}</h2>
          <p>{{ family.familySurname }}氏 · {{ family.regionText || family.nativePlace || '地区未设置' }}</p>
        </div>
        <div class="status-tags"><StatusTag :status="family.status" /><StatusTag :status="family.publicDisplayStatus" /></div>
      </section>
      <section class="surface detail-card">
        <div class="section-heading"><div><h3>家庭资料</h3><p>用于识别与展示的基础信息。</p></div></div>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="家庭 ID">{{ family.id }}</el-descriptions-item>
          <el-descriptions-item label="姓氏">{{ family.familySurname }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(family.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="籍贯">{{ family.nativePlace || '未设置' }}</el-descriptions-item>
          <el-descriptions-item label="地区">{{ family.regionText || '未设置' }}</el-descriptions-item>
          <el-descriptions-item label="当前创始人节点">{{ family.currentFounderMemberId || '未记录' }}</el-descriptions-item>
        </el-descriptions>
        <div class="description-block"><strong>家庭简介</strong><p>{{ family.description || '未填写家庭简介' }}</p></div>
      </section>
      <section class="surface detail-card">
        <div class="section-heading"><div><h3>结构与可见性</h3><p>反映家庭树版本及公开展示状态。</p></div></div>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="成员数">{{ family.memberCount }}</el-descriptions-item>
          <el-descriptions-item label="家谱版本">{{ family.graphVersion }}</el-descriptions-item>
          <el-descriptions-item label="允许搜索">{{ family.searchable ? '是' : '否' }}</el-descriptions-item>
        </el-descriptions>
      </section>
    </template>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getApiErrorMessage } from '@/api/client'
import { getFamily } from '@/api/management'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { ManagedFamily } from '@/types/api'

const route = useRoute()
const familyId = String(route.params.familyId)
const family = ref<ManagedFamily | null>(null)
const error = ref('')

function formatTime(value: string) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
}

onMounted(async () => {
  try {
    family.value = await getFamily(familyId)
  } catch (requestError) {
    error.value = getApiErrorMessage(requestError)
  }
})
</script>
<style scoped>
.family-overview { display:flex;align-items:flex-start;justify-content:space-between;gap:24px;padding:24px;background:linear-gradient(135deg,rgba(23,47,74,.96),rgba(47,107,87,.93));color:#fff }
.eyebrow { color:rgba(248,231,194,.86);font-size:12px;font-weight:800;letter-spacing:.08em }.family-overview h2 { margin:8px 0;font-size:30px }.family-overview p { margin:0;color:rgba(255,255,255,.76) }.status-tags { display:flex;flex-wrap:wrap;justify-content:flex-end;gap:8px }.detail-card { padding:20px }.section-heading { margin-bottom:16px }.section-heading h3 { margin:0;color:var(--color-primary);font-size:18px }.section-heading p,.description-block p { margin:6px 0 0;color:var(--color-text-secondary);line-height:1.7 }.description-block { margin-top:18px;padding-top:16px;border-top:1px dashed var(--color-border) } @media (max-width:760px){.family-overview{flex-direction:column}.status-tags{justify-content:flex-start}}
</style>
